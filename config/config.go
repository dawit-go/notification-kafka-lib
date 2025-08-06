// Package config provides configuration management for the notification service,
// supporting retrieval of sensitive configuration values from HashiCorp Vault or
// falling back to environment variables or default values.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/vault/api"
)

// ConfigParsed holds the full parsed configuration for the notification service,
// including email and Kafka related settings.
type ConfigParsed struct {
	Email EmailConfig `json:"email"`
	Kafka KafkaConfig `json:"kafka"`
}

// EmailConfig contains configuration fields related to email sending service,
// such as Mailjet API keys and sender details.
type EmailConfig struct {
	MailjetAPIKey      string `json:"mailjet_api_key"`
	MailjetSecret      string `json:"mailjet_secret"`
	MailjetSenderEmail string `json:"mailjet_sender_email"`
	MailjetSenderName  string `json:"mailjet_sender_name"`
}

// KafkaConfig holds Kafka connection and topic configuration details.
type KafkaConfig struct {
	Brokers          string `json:"brokers"`           // Comma separated list of Kafka brokers (e.g. "broker1:9092,broker2:9092")
	EmailTopic       string `json:"email_topic"`       // Kafka topic name for email notifications
	SmsTopic         string `json:"sms_topic"`         // Kafka topic name for SMS notifications
	FeedbackTopic    string `json:"feedback_topic"`    // Kafka topic name for feedback events
	ConsumerGroup    string `json:"consumer_group"`    // Kafka consumer group ID
	AutoOffsetReset  string `json:"auto_offset_reset"` // Kafka consumer auto offset reset policy ("earliest" or "latest")
	EnableAutoCommit bool   `json:"enable_auto_commit"`// Enable Kafka consumer auto commit of offsets
	SessionTimeoutMs int    `json:"session_timeout_ms"`// Kafka consumer session timeout in milliseconds
	SASLEnabled      bool   `json:"sasl_enabled"`      // Enable SASL authentication
	SASLUsername     string `json:"sasl_username"`     // SASL authentication username
	SASLPassword     string `json:"sasl_password"`     // SASL authentication password
	SASLMechanism    string `json:"sasl_mechanism"`    // SASL mechanism (e.g. "PLAIN")
}

// VaultClient provides a client wrapper for interacting with HashiCorp Vault
// to securely fetch secrets at a configured Vault path.
type VaultClient struct {
	client     *api.Client               // Vault API client instance
	path       string                    // Vault secret path to read from
	secretData map[string]interface{}   // Cached secrets data read from Vault
}

// NewVaultClient initializes a new VaultClient using environment variables:
// VAULT_ADDR, VAULT_TOKEN, VAULT_PATH. It authenticates with Vault and fetches
// secrets from the given Vault path.
//
// Returns an error if any required environment variable is missing or
// Vault client initialization/fetching fails.
func NewVaultClient() (*VaultClient, error) {
	vaultAddr := getEnv("VAULT_ADDR")
	vaultToken := getEnv("VAULT_TOKEN")
	vaultPath := getEnv("VAULT_PATH")

	if vaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable is not set")
	}
	if vaultToken == "" {
		return nil, fmt.Errorf("VAULT_TOKEN environment variable is not set")
	}
	if vaultPath == "" {
		return nil, fmt.Errorf("VAULT_PATH environment variable is not set")
	}

	config := &api.Config{Address: vaultAddr}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Vault client: %w", err)
	}
	client.SetToken(vaultToken)

	vaultClient := &VaultClient{
		client: client,
		path:   vaultPath,
	}

	if err := vaultClient.fetchSecrets(); err != nil {
		return nil, fmt.Errorf("failed to fetch secrets from Vault: %w", err)
	}

	return vaultClient, nil
}

// fetchSecrets reads secrets from Vault at the configured path and caches them.
// Returns an error if no secrets are found or if Vault read fails.
func (v *VaultClient) fetchSecrets() error {
	secret, err := v.client.Logical().Read(v.path)
	if err != nil {
		return err
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("no secrets found at path: %s", v.path)
	}
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		v.secretData = data
	}
	return nil
}

// GetSecret retrieves a secret value by key from the cached Vault secrets.
// Returns empty string and no error if key not found.
func (v *VaultClient) GetSecret(key string) (string, error) {
	if value, ok := v.secretData[key].(string); ok {
		return value, nil
	}
	return "", nil
}

// Load loads the full ConfigParsed by fetching secrets from Vault and
// falling back to environment variables or default values if Vault secrets
// are missing.
//
// This function returns the fully parsed configuration ready for use.
func Load() (*ConfigParsed, error) {
	vaultClient, err := NewVaultClient()
	if err != nil {
		return nil, err
	}

	getConfigValue := func(vaultKey, defaultValue string) string {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			return vaultValue
		}
		return defaultValue
	}

	getConfigBool := func(vaultKey string, defaultValue bool) bool {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			if boolValue, err := strconv.ParseBool(vaultValue); err == nil {
				return boolValue
			}
		}
		return defaultValue
	}

	getConfigInt := func(vaultKey string, defaultValue int) int {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			if intValue, err := strconv.Atoi(vaultValue); err == nil {
				return intValue
			}
		}
		return defaultValue
	}

	cfg := &ConfigParsed{
		Email: EmailConfig{
			MailjetAPIKey:      getConfigValue("MAILJET_API_KEY", "8ffe2ad16061f06aa2be7e98e94647d8"),
			MailjetSecret:      getConfigValue("MAILJET_SECRET_KEY", "b507c6e61ab193f98910589767770fd9"),
			MailjetSenderEmail: getConfigValue("MAILJET_SENDER_EMAIL", "noreply@dubeale.com"),
			MailjetSenderName:  getConfigValue("MAILJET_SENDER_NAME", "CBE"),
		},
		Kafka: KafkaConfig{
			Brokers:          getConfigValue("KAFKA_BROKERS", ""),
			EmailTopic:       getConfigValue("KAFKA_EMAIL_TOPIC", "email-notifications"),
			SmsTopic:         getConfigValue("KAFKA_SMS_TOPIC", "sms-notifications"),
			FeedbackTopic:    getConfigValue("KAFKA_FEEDBACK_TOPIC", "feedback-events"),
			ConsumerGroup:    getConfigValue("KAFKA_CONSUMER_GROUP", "notification-service"),
			AutoOffsetReset:  getConfigValue("KAFKA_AUTO_OFFSET_RESET", "earliest"),
			EnableAutoCommit: getConfigBool("KAFKA_ENABLE_AUTO_COMMIT", true),
			SessionTimeoutMs: getConfigInt("KAFKA_SESSION_TIMEOUT_MS", 10000),
			SASLEnabled:      getConfigBool("KAFKA_SASL_ENABLED", false),
			SASLUsername:     getConfigValue("KAFKA_SASL_USERNAME", ""),
			SASLPassword:     getConfigValue("KAFKA_SASL_PASSWORD", ""),
			SASLMechanism:    getConfigValue("KAFKA_SASL_MECHANISM", "PLAIN"),
		},
	}

	return cfg, nil
}

// getEnv reads the value of an environment variable.
// Returns an empty string if the variable is not set.
func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return ""
}
