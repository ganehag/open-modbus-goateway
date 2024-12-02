package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/simonvetter/modbus"
)

// ModbusHandler implements the Handler interface for Modbus devices
type ModbusHandler struct{}

// Handle processes the incoming payload, performs Modbus operations, and returns a response
func (h *ModbusHandler) Handle(topic string, payload string) string {
	// Parse and validate the request payload
	request, err := parseRequest(payload)
	if err != nil {
		log.Printf("Invalid request: %v", err)
		return fmt.Sprintf("%d ERROR: %v", 0, err) // If cookie is invalid, default to 0
	}

	// Perform Modbus query
	response, err := h.executeModbusQuery(request)
	if err != nil {
		log.Printf("Modbus query failed: %v", err)
		return fmt.Sprintf("%d ERROR: %v", request.Cookie, err)
	}

	// Construct the response
	if len(response) > 0 {
		return fmt.Sprintf("%d OK %s", request.Cookie, strings.Join(response, " "))
	}

	return fmt.Sprintf("%d OK", request.Cookie)
}

func (h *ModbusHandler) executeModbusQuery(req *ModbusRequest) ([]string, error) {
	// Create the Modbus client
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     fmt.Sprintf("tcp://%s:%d", req.IPAddress, req.Port),
		Timeout: req.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Modbus client: %v", err)
	}
	defer client.Close()

	// Open the connection to the Modbus device
	err = client.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Modbus server: %v", err)
	}

	// Set the Slave ID (Unit ID)
	client.SetUnitId(req.SlaveID)

	// Variable to store the results
	var results []uint16

	// Handle each supported function code
	switch req.FunctionCode {
	case 1: // Read Coils (0x01)
		// Read coils and convert to uint16 values (1 or 0)
		bits, err := client.ReadCoils(req.RegisterAddress, req.RegisterCount)
		if err != nil {
			return nil, fmt.Errorf("failed to read coils: %v", err)
		}
		for _, bit := range bits {
			if bit {
				results = append(results, 1)
			} else {
				results = append(results, 0)
			}
		}
	case 2: // Read Discrete Inputs (0x02)
		// Read discrete inputs and convert to uint16 values (1 or 0)
		bits, err := client.ReadDiscreteInputs(req.RegisterAddress, req.RegisterCount)
		if err != nil {
			return nil, fmt.Errorf("failed to read discrete inputs: %v", err)
		}
		for _, bit := range bits {
			if bit {
				results = append(results, 1)
			} else {
				results = append(results, 0)
			}
		}
	case 3: // Read Holding Registers (0x03)
		results, err = client.ReadRegisters(req.RegisterAddress, req.RegisterCount, modbus.HOLDING_REGISTER)
		if err != nil {
			return nil, fmt.Errorf("failed to read holding registers: %v", err)
		}
	case 4: // Read Input Registers (0x04)
		results, err = client.ReadRegisters(req.RegisterAddress, req.RegisterCount, modbus.INPUT_REGISTER)
		if err != nil {
			return nil, fmt.Errorf("failed to read input registers: %v", err)
		}
	case 5: // Write Single Coil (0x05)
		// Convert uint16 to bool for writing a single coil
		value := req.Data[0] != 0
		err = client.WriteCoil(req.RegisterAddress, value)
		if err != nil {
			return nil, fmt.Errorf("failed to write single coil: %v", err)
		}
	case 6: // Write Single Register (0x06)
		err = client.WriteRegister(req.RegisterAddress, req.Data[0])
		if err != nil {
			return nil, fmt.Errorf("failed to write single register: %v", err)
		}
	case 15: // Write Multiple Coils (0x0F)
		// Convert []uint16 to []bool for writing multiple coils
		bitValues := make([]bool, len(req.Data))
		for i, v := range req.Data {
			bitValues[i] = v != 0
		}
		err = client.WriteCoils(req.RegisterAddress, bitValues)
		if err != nil {
			return nil, fmt.Errorf("failed to write multiple coils: %v", err)
		}
	case 16: // Write Multiple Registers (0x10)
		err = client.WriteRegisters(req.RegisterAddress, req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to write multiple registers: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported function code: %d", req.FunctionCode)
	}

	// Format results into strings
	response := make([]string, len(results))
	for i, val := range results {
		response[i] = strconv.Itoa(int(val))
	}

	return response, nil
}
