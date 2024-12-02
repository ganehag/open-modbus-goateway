package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ModbusRequest represents a parsed Modbus query request
type ModbusRequest struct {
	Cookie          uint64
	IPAddress       string
	Port            uint16
	Timeout         time.Duration
	SlaveID         uint8
	FunctionCode    uint8
	RegisterAddress uint16
	RegisterCount   uint16
	Data            []uint16
}

// parseRequest parses the Modbus request payload into a ModbusRequest struct
func parseRequest(payload string) (*ModbusRequest, error) {
	parts := strings.Fields(payload)
	if len(parts) < 9 {
		return nil, fmt.Errorf("incomplete request payload")
	}

	// Parse fixed parts of the request
	cookie, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid COOKIE value: %v", err)
	}

	ip := parts[3]
	port, err := strconv.ParseUint(parts[4], 10, 16)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT value: %v", err)
	}

	timeout, err := strconv.Atoi(parts[5])
	if err != nil || timeout < 1 || timeout > 999 {
		return nil, fmt.Errorf("invalid TIMEOUT value: %v", err)
	}

	slaveID, err := strconv.ParseUint(parts[6], 10, 8)
	if err != nil || slaveID < 1 || slaveID > 255 {
		return nil, fmt.Errorf("invalid SLAVE_ID value: %v", err)
	}

	functionCode, err := strconv.ParseUint(parts[7], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid MODBUS_FUNCTION value: %v", err)
	}

	registerAddress, err := strconv.ParseUint(parts[8], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid REGISTER_NUMBER value: %v", err)
	}
	registerAddress -= 1 // Requests uses RegisterNumbers

	registerCount := uint16(0)
	data := []uint16{}

	// Parse function-specific values
	switch functionCode {
	case 1, 2, 3, 4: // Reading functions
		if len(parts) < 10 {
			return nil, fmt.Errorf("missing REGISTER_COUNT for function %d", functionCode)
		}
		count, err := strconv.ParseUint(parts[9], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid REGISTER_COUNT value: %v", err)
		}
		registerCount = uint16(count)
	case 15, 16: // Writing multiple registers/coils
		if len(parts) < 11 {
			return nil, fmt.Errorf("missing REGISTER_COUNT or DATA for function %d", functionCode)
		}
		count, err := strconv.ParseUint(parts[9], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid REGISTER_COUNT value: %v", err)
		}
		registerCount = uint16(count)
		rawData := strings.Split(parts[10], ",")
		for _, v := range rawData {
			value, err := strconv.ParseUint(v, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid DATA value: %v", err)
			}
			data = append(data, uint16(value))
		}
		if len(data) != int(registerCount) {
			return nil, fmt.Errorf("mismatch between REGISTER_COUNT and DATA length")
		}
	}

	return &ModbusRequest{
		Cookie:          cookie,
		IPAddress:       ip,
		Port:            uint16(port),
		Timeout:         time.Duration(timeout) * time.Second,
		SlaveID:         uint8(slaveID),
		FunctionCode:    uint8(functionCode),
		RegisterAddress: uint16(registerAddress),
		RegisterCount:   registerCount,
		Data:            data,
	}, nil
}
