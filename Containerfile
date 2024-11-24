# Stage 1: Build the Go application
FROM docker.io/golang:1.23 AS build

# Set the working directory
WORKDIR /go/src/app

# Copy the source code and Go modules
COPY . .

# Download dependencies
RUN go mod download

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/open-modbus-goateway ./cmd/open-modbus-goateway

# Stage 2: Use Distroless as the runtime base image
FROM gcr.io/distroless/static-debian12

# Copy the built binary from the builder stage
COPY --from=build /go/bin/open-modbus-goateway /

# Set the command to run the application
CMD ["/open-modbus-goateway"]
