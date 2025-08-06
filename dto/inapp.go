package dto

import "time"

// InAppKafkaMessage represents an in-app notification message from Kafka
type InAppKafkaMessage struct {
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	ImageURL  string                 `json:"image_url,omitempty"`
	ActionURL string                 `json:"action_url,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Priority  int                    `json:"priority,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
