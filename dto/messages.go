package dto

import (
	"encoding/json"
	"time"
)

// NotificationMessage represents the base structure for Kafka messages
type NotificationMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "email", "sms", "feedback", "inapp", etc.
	Payload   json.RawMessage        `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}

// NewNotificationMessage creates a new NotificationMessage with marshaled payload
func NewNotificationMessage(id, msgType string, payload interface{}) (*NotificationMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &NotificationMessage{
		ID:        id,
		Type:      msgType,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
		Headers:   make(map[string]interface{}),
	}, nil
}

// UnmarshalPayload unmarshals the payload into the provided struct
func (n *NotificationMessage) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(n.Payload, v)
}
