package dto

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// NotificationMessage represents the base structure for Kafka messages
type NotificationMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "email", "sms", "feedback", "inapp", etc.
	Payload   json.RawMessage        `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}

// SendEmailRequest represents the request to send an email
type SendEmailRequest struct {
	Recipients         []EmailContact         `json:"recipients" validate:"required"`
	Subject            string                 `json:"subject" validate:"required"`
	OTPCode            string                 `json:"otp_code,omitempty"`
	Type               string                 `json:"type" validate:"required"` // otp, message, transaction, etc.
	Receiver           string                 `json:"receiver,omitempty"`
	MessageBody        string                 `json:"message_body,omitempty"`
	Link               string                 `json:"link,omitempty"`
	CustomerName       string                 `json:"customer_name,omitempty"`
	TransactionDetails map[string]interface{} `json:"transaction_details,omitempty"`
	CC                 []EmailContact         `json:"cc,omitempty" bson:"cc,omitempty"`
}

// Validate validates the SendEmailRequest fields
func (s SendEmailRequest) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Recipients, validation.Required.Error("recipients are required")),
		validation.Field(&s.Subject, validation.Required.Error("subject is required")),
		validation.Field(&s.Type, validation.Required.Error("type is required")),
	)
}

// EmailKafkaMessage represents an email message consumed from Kafka
type EmailKafkaMessage struct {
	Recipients         []EmailContact         `json:"recipients"`
	CC                 []EmailContact         `json:"cc,omitempty"`
	Subject            string                 `json:"subject"`
	Type               string                 `json:"type"` // "otp", "message", "transaction", etc.
	OTPCode            string                 `json:"otp_code,omitempty"`
	Receiver           string                 `json:"receiver,omitempty"`
	MessageBody        string                 `json:"message_body,omitempty"`
	Link               string                 `json:"link,omitempty"`
	CustomerName       string                 `json:"customer_name,omitempty"`
	TransactionDetails map[string]interface{} `json:"transaction_details,omitempty"`
	Priority           int                    `json:"priority,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// ToSendEmailRequest converts EmailKafkaMessage to SendEmailRequest
func (e *EmailKafkaMessage) ToSendEmailRequest() SendEmailRequest {
	return SendEmailRequest{
		Recipients:         e.Recipients,
		CC:                 e.CC,
		Subject:            e.Subject,
		Type:               e.Type,
		OTPCode:            e.OTPCode,
		Receiver:           e.Receiver,
		MessageBody:        e.MessageBody,
		Link:               e.Link,
		CustomerName:       e.CustomerName,
		TransactionDetails: e.TransactionDetails,
	}
}

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

// FeedbackKafkaMessage represents feedback messages from Kafka
type FeedbackKafkaMessage struct {
	UserID      string                 `json:"user_id"`
	FullName    string                 `json:"full_name"`
	Feedback    string                 `json:"feedback"`
	Rating      int                    `json:"rating"`
	Module      string                 `json:"module"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	SubmittedAt *time.Time             `json:"submitted_at,omitempty"`
}

// EmailContact represents an email contact
type EmailContact struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email" bson:"email"`
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
