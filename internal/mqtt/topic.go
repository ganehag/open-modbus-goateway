package mqtt

import (
	"fmt"
	"regexp"
	"strings"
)

// placeholderRegex matches placeholders in the format `{placeholder}`
var placeholderRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// Topic represents a parsed MQTT topic
type Topic struct {
	Format string            // The topic format (e.g., "modbus/{device}/request")
	Values map[string]string // Placeholder values (e.g., {"device": "device123"})
}

// ParseTopic parses a topic string based on a format string with placeholders like `{device}`.
// It returns a Topic instance containing the parsed values.
func ParseTopic(topic, format string) (*Topic, error) {
	topicParts := strings.Split(topic, "/")
	formatParts := strings.Split(format, "/")

	if len(topicParts) != len(formatParts) {
		return nil, fmt.Errorf("topic %q does not match format %q", topic, format)
	}

	values := make(map[string]string)
	for i, part := range formatParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			values[key] = topicParts[i]
		} else if part != topicParts[i] {
			return nil, fmt.Errorf("topic %q does not match format %q at part %d", topic, format, i)
		}
	}

	return &Topic{
		Format: format,
		Values: values,
	}, nil
}

// Build reconstructs the topic string from the format and values.
func (t *Topic) Build() (string, error) {
	formatParts := strings.Split(t.Format, "/")

	for i, part := range formatParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			value, ok := t.Values[key]
			if !ok {
				return "", fmt.Errorf("missing value for placeholder %q", key)
			}
			formatParts[i] = value
		}
	}

	return strings.Join(formatParts, "/"), nil
}

// WithWildcard converts the topic format into an MQTT wildcard subscription.
// For example, "modbus/{device}/{action}" -> "modbus/+/+".
func (t *Topic) WithWildcard() string {
	return placeholderRegex.ReplaceAllString(t.Format, `+`)
}
