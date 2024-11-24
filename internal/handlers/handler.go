package handlers

// Handler is an interface for processing MQTT messages
type Handler interface {
	Handle(device string, payload string) string
}
