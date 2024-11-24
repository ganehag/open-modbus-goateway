package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// NewTLSConfig creates a TLS configuration for MQTT connections with optional client certificates and CA file
func NewTLSConfig(caCertPath, certPath, keyPath, serverName string) (*tls.Config, error) {
	var certPool *x509.CertPool
	var err error

	if caCertPath != "" {
		// Load the CA certificate from the provided path
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		// Create a new certificate pool and append the CA certificate
		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate to pool")
		}
	} else {
		// Use the system's root CA pool if no CA file is specified
		certPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system root CA pool: %w", err)
		}
	}

	// Create the TLS configuration
	tlsConfig := &tls.Config{
		RootCAs:    certPool,   // Use the appropriate CA pool
		ServerName: serverName, // Explicit server name for hostname verification
	}

	// If client certificate and key are provided, load them
	if certPath != "" && keyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}
