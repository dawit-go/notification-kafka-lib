package dto

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// EmailContact represents an email contact
type EmailContact struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email" bson:"email"`
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
