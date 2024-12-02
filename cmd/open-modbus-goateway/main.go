package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ganehag/open-modbus-goateway/internal/config"
	"github.com/ganehag/open-modbus-goateway/internal/handlers"
	"github.com/ganehag/open-modbus-goateway/internal/mqtt"
)

func main() {
	log.Println("Starting Open Modbus Goateway...")

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create the Modbus handler
	handler := &handlers.ModbusHandler{}

	// Create the Dummy handler
	// handler := &handlers.DummyHandler{}

	// Define the number of workers
	workerCount := 4 // Adjust this based on expected load and available resources

	// Initialize the MQTT client with the handler and worker count
	client, err := mqtt.NewClient(cfg.MQTT, handler, workerCount)
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}

	// Create a context to manage shutdown signals
	ctx, cancel := context.WithCancel(context.Background())

	// Start workers
	client.StartWorkers(ctx)

	log.Println("Open Modbus Goateway is running. Waiting for messages...")

	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	<-signalChan
	log.Println("Received termination signal. Shutting down...")

	// Cancel the context to stop workers
	cancel()

	// Stop the client
	client.Stop()

	log.Println("Open Modbus Goateway stopped gracefully.")
}
