package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// DummyHandler implements the Handler interface for Modbus devices
type DummyHandler struct{}

// Handle processes the incoming payload, performs Modbus operations, and returns a response
func (h *DummyHandler) Handle(topic string, payload string) string {
	// Parse and validate the request payload
	request, err := parseRequest(payload)
	if err != nil {
		log.Printf("Invalid request: %v", err)
		return fmt.Sprintf("%d ERROR: %v", 0, err) // If cookie is invalid, default to 0
	}

	// Perform Modbus query
	// response, err := h.executeModbusQuery(request)
	response, err := h.executeDummyQuery(request)
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

func (h *DummyHandler) executeDummyQuery(req *ModbusRequest) ([]string, error) {
	var dummyValue uint16 = 1
	var results []uint16
	switch req.FunctionCode {
	case 1:
		fallthrough
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		var i uint16
		for i = 0; i < req.RegisterCount; i++ {
			results = append(results, dummyValue)
		}
	}

	// Format results into strings
	response := make([]string, len(results))
	for i, val := range results {
		response[i] = strconv.Itoa(int(val))
	}

	return response, nil
}
