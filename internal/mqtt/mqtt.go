package mqtt

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ganehag/open-modbus-goateway/internal/config"
	"github.com/ganehag/open-modbus-goateway/internal/handlers"
	"github.com/ganehag/open-modbus-goateway/internal/tlsutil"
)

// convertToWildcard replaces placeholders like {device} with MQTT wildcards (+)
func convertToWildcard(topic string) string {
	return strings.ReplaceAll(topic, "{device}", "+")
}

// Client wraps the MQTT client, configuration, and worker pool
type Client struct {
	mqttClient     mqtt.Client
	cfg            config.MQTTConfig
	handler        handlers.Handler
	workers        int
	messageCh      chan mqtt.Message
	wg             sync.WaitGroup
	requestCounter int32
	ctx            context.Context    // Context for managing client lifecycle
	cancelFunc     context.CancelFunc // Cancel function to signal termination
}

// NewClient initializes and connects an MQTT client based on the provided configuration
// and sets up concurrent message handling.
func NewClient(cfg config.MQTTConfig, handler handlers.Handler, workers int) (*Client, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be greater than zero")
	}

	// Parse the broker URL to check if TLS is required
	u, err := url.Parse(cfg.Broker)
	if err != nil {
		return nil, fmt.Errorf("failed to parse broker URL: %w", err)
	}

	// Initialize message channel
	messageCh := make(chan mqtt.Message, workers*10) // Buffered channel for better throughput

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Printf("Connected to MQTT broker: %v", cfg.Broker)

			// Create a Topic struct for request_topic
			requestTopic := &Topic{Format: cfg.RequestTopic}
			subscriptionTopic := requestTopic.WithWildcard()

			// Subscribe to the topic on connect/reconnect
			token := client.Subscribe(subscriptionTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
				messageCh <- msg // Send message to the channel
			})
			token.Wait()
			if token.Error() != nil {
				log.Printf("Failed to subscribe to topic %s: %v", subscriptionTopic, token.Error())
			} else {
				log.Printf("Subscribed to topic: %s", subscriptionTopic)
			}
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("Connection lost: %v", err)
		})

	if u.Scheme == "ssl" {
		// Parse the broker URL to extract the hostname
		u, err := url.Parse(cfg.Broker)
		if err != nil {
			return nil, fmt.Errorf("failed to parse broker URL: %w", err)
		}

		// Extract the hostname
		hostname := u.Hostname()

		// Create the TLS configuration
		tlsConfig, err := tlsutil.NewTLSConfig(cfg.CACertPath, cfg.CertPath, cfg.KeyPath, hostname)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	// Create a cancellable context
	ctx, cancelFunc := context.WithCancel(context.Background())

	c := &Client{
		mqttClient: client,
		cfg:        cfg,
		handler:    handler,
		workers:    workers,
		messageCh:  messageCh,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	// Start the background routine for request counting
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.startRequestCounterLogger()
	}()

	return c, nil
}

// startWorkers starts a pool of goroutines to process messages concurrently
func (c *Client) StartWorkers() {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for msg := range c.messageCh {
				c.processMessage(msg)
			}
		}()
	}
}

func (c *Client) Stop() {
	log.Println("Stopping MQTT client and workers...")

	// Cancel the context to stop background routines
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// Disconnect the MQTT client
	c.mqttClient.Disconnect(250)

	// Close the message channel to stop workers
	close(c.messageCh)

	// Wait for all workers and routines to finish
	c.wg.Wait()

	log.Println("MQTT client and workers stopped.")
}

func (c *Client) startRequestCounterLogger() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done(): // Context canceled
			log.Println("Request counter logger stopped")
			return
		case <-ticker.C:
			// Safely read and reset the counter
			count := atomic.LoadInt32(&c.requestCounter)
			atomic.StoreInt32(&c.requestCounter, 0)

			log.Printf("Requests handled in the last minute: %d", count)
		}
	}
}

func (c *Client) processMessage(msg mqtt.Message) {
	// Increment the counter atomically
	atomic.AddInt32(&c.requestCounter, 1)

	// Parse the incoming topic
	requestTopic, err := ParseTopic(msg.Topic(), c.cfg.RequestTopic)
	if err != nil {
		log.Printf("Failed to parse topic %q: %v", msg.Topic(), err)
		return
	}

	// Pass the full topic and payload to the handler
	responsePayload := c.handler.Handle(msg.Topic(), string(msg.Payload()))

	// Rebuild the response topic dynamically
	responseTopic := &Topic{
		Format: c.cfg.ResponseTopic,
		Values: requestTopic.Values, // Reuse extracted values
	}
	responseTopicString, err := responseTopic.Build()
	if err != nil {
		log.Printf("Failed to build response topic: %v", err)
		return
	}

	// Publish the response
	token := c.mqttClient.Publish(responseTopicString, 1, false, responsePayload)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to publish response to topic %s: %v", responseTopicString, token.Error())
	}
}
