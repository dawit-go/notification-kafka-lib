package dto

// SMSKafkaMessage represents an SMS message received from Kafka
type SMSKafkaMessage struct {
	Recipient   string                 `json:"recipient"`
	MessageBody string                 `json:"message_body"`
	Priority    int                    `json:"priority,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PushKafkaMessage represents a push notification message received from Kafka
type PushKafkaMessage struct {
	UserID       string                   `json:"user_id"`
	DeviceTokens []string                 `json:"device_tokens"`
	Title        string                   `json:"title"`
	Body         string                   `json:"body"`
	ImageURL     string                   `json:"image_url,omitempty"`
	Priority     PushNotificationPriority `json:"priority,omitempty"`
	Data         map[string]string        `json:"data,omitempty"`
	Badge        int                      `json:"badge,omitempty"`
	Sound        string                   `json:"sound,omitempty"`
	ClickAction  string                   `json:"click_action,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
}

// PushNotificationPriority represents the priority of a push notification
type PushNotificationPriority string

const (
	// PushPriorityHigh represents high priority push notification
	PushPriorityHigh PushNotificationPriority = "high"
	// PushPriorityNormal represents normal priority push notification
	PushPriorityNormal PushNotificationPriority = "normal"
	// PushPriorityLow represents low priority push notification
	PushPriorityLow PushNotificationPriority = "low"
)

// IsValid checks if the push notification priority is valid
func (p PushNotificationPriority) IsValid() bool {
	switch p {
	case PushPriorityHigh, PushPriorityNormal, PushPriorityLow:
		return true
	default:
		return false
	}
}
