package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config represents the structure of the configuration file
type Config struct {
	MQTT MQTTConfig `yaml:"mqtt"`
}

// MQTTConfig holds MQTT-related settings
type MQTTConfig struct {
	Broker        string `yaml:"broker"`         // MQTT broker address
	ClientID      string `yaml:"client_id"`      // MQTT client ID
	Username      string `yaml:"username"`       // MQTT username
	Password      string `yaml:"password"`       // MQTT password
	RequestTopic  string `yaml:"request_topic"`  // Action placeholder for request topics
	ResponseTopic string `yaml:"response_topic"` // Action placeholder for response topics
	CACertPath    string `yaml:"ca_cert_path"`   // Path to CA certificate
	CertPath      string `yaml:"cert_path"`      // Path to client certificate
	KeyPath       string `yaml:"key_path"`       // Path to client key
}

// Load loads the configuration from the given YAML file
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %w", err)
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validate checks for required fields and logical consistency in the configuration
func (c *Config) validate() error {
	if c.MQTT.Broker == "" {
		return fmt.Errorf("mqtt.broker must be specified")
	}
	if c.MQTT.ClientID == "" {
		return fmt.Errorf("mqtt.client_id must be specified")
	}
	if c.MQTT.RequestTopic == "" {
		return fmt.Errorf("mqtt.request_action must be specified")
	}
	if c.MQTT.ResponseTopic == "" {
		return fmt.Errorf("mqtt.response_action must be specified")
	}
	return nil
}
