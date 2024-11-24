# Open Modbus Goateway

Open Modbus Goateway is a lightweight, efficient gateway that bridges MQTT and Modbus protocols. It provides seamless communication between devices, supporting modern TLS-secured MQTT brokers and handling Modbus TCP communication for robust and secure automation workflows.

## Features

- **MQTT to Modbus Gateway**: Routes MQTT messages to Modbus devices and vice versa.
- **TLS Support**: Secured communication with modern MQTT brokers using CA certificates or system certificates.
- **Flexible Configuration**: Customizable through a `config.yaml` file.
- **Minimalist Container**: Runs on a lightweight Distroless base image for secure and production-ready deployments.

## Getting Started

### Prerequisites

- Docker (for containerized deployment)
- A Modbus-compatible device or simulator
- An MQTT broker (e.g., [Mosquitto](https://mosquitto.org))

### Configuration

Create a `config.yaml` file for your application with the following structure:

```yaml
mqtt:
  broker: "ssl://test.mosquitto.org:8886"  # MQTT broker URL
  client_id: "open-modbus-goateway"
  username: "your-username"
  password: "your-password"
  request_topic: "modbus/{device}/request"
  response_topic: "modbus/{device}/response"  
  ca_cert_path: ""  # Path to CA certificate file (optional)
  cert_path: ""     # Path to client certificate (optional)
  key_path: ""      # Path to client key (optional)
```

### Building the Project

To build the application, use the following commands:

```bash
# Clone the repository
git clone https://github.com/ganehag/open-modbus-goateway.git
cd open-modbus-goateway

# Build the Docker container
docker build -t open-modbus-goateway -f Containerfile .
```

### Running the Application

Run the containerized application:

```bash
docker run --rm -v $(pwd)/config/config.yaml:/config/config.yaml open-modbus-goateway
```

---

## Development

### Building Locally

If you prefer to build and run the application without Docker:

```bash
# Install Go (if not already installed)
go mod tidy
go build -o open-modbus-goateway ./cmd/open-modbus-goateway
./open-modbus-goateway
```

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

---

## Acknowledgments

- Built with [Go](https://golang.org)
- Inspired by the Open Modbus Gateway software
- Leverages [Distroless](https://github.com/GoogleContainerTools/distroless) for secure container images
