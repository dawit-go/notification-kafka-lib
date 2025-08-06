// Package config provides configuration management for Kafka settings in the notification service,
// supporting environment variables and HashiCorp Vault for secret storage.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/vault/api"
)

// ConfigParsed holds the parsed Kafka configuration for the notification service.
type ConfigParsed struct {
	Kafka KafkaConfig `json:"kafka"`
}

// KafkaConfig holds configuration settings for Kafka integration.
type KafkaConfig struct {
	Brokers          string `json:"brokers"`            // Comma-separated list of Kafka broker addresses
	SMSTopic         string `json:"sms_topic"`         // Topic for SMS notifications
	EmailTopic       string `json:"email_topic"`       // Topic for email notifications
	InAppTopic       string `json:"inapp_topic"`       // Topic for in-app notifications
	PushTopic        string `json:"push_topic"`        // Topic for push notifications
	ConsumerGroup    string `json:"consumer_group"`    // Kafka consumer group ID
	SASLEnabled      bool   `json:"sasl_enabled"`      // Whether SASL authentication is enabled
	SASLUsername     string `json:"sasl_username"`     // SASL username for authentication
	SASLPassword     string `json:"sasl_password"`     // SASL password for authentication
	SASLMechanism    string `json:"sasl_mechanism"`    // SASL mechanism (e.g., PLAIN, SCRAM-SHA-256)
	AutoOffsetReset  string `json:"auto_offset_reset"` // Offset reset policy (e.g., earliest, latest)
	EnableAutoCommit bool   `json:"enable_auto_commit"` // Whether to enable auto-commit for consumer offsets
	SessionTimeoutMs int    `json:"session_timeout_ms"` // Consumer group session timeout in milliseconds
}

// VaultClient wraps the HashiCorp Vault client with caching capabilities for secrets.
type VaultClient struct {
	client     *api.Client
	path       string
	secretData map[string]interface{}
}

// NewVaultClient creates a new Vault client using environment variables for configuration.
// It initializes the client with the Vault address and token, and fetches secrets from the specified path.
//
// Returns a new VaultClient or an error if initialization or secret fetching fails.
func NewVaultClient() (*VaultClient, error) {
	config := &api.Config{
		Address: getEnv("VAULT_ADDR"),
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(getEnv("VAULT_TOKEN"))

	vault := &VaultClient{
		client: client,
		path:   getEnv("VAULT_PATH"),
	}

	if err := vault.fetchSecrets(); err != nil {
		return nil, fmt.Errorf("failed to fetch secrets: %w", err)
	}

	return vault, nil
}

// fetchSecrets retrieves secrets from Vault and caches them in the VaultClient.
// It reads secrets from the configured path and stores them in secretData.
//
// Returns an error if the read operation fails or no secrets are found.
func (v *VaultClient) fetchSecrets() error {
	secret, err := v.client.Logical().Read(v.path)
	if err != nil {
		return fmt.Errorf("failed to read secrets from Vault path %s: %w", v.path, err)
	}

	if secret == nil || secret.Data == nil {
		return fmt.Errorf("no secrets found at path: %s", v.path)
	}

	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		v.secretData = data
		return nil
	}

	return fmt.Errorf("invalid secret data format at path: %s", v.path)
}

// GetSecret retrieves a secret value from the cached Vault secrets by key.
//
// Returns the secret value as a string or an empty string if not found, along with any error.
func (v *VaultClient) GetSecret(key string) (string, error) {
	if value, ok := v.secretData[key].(string); ok {
		return value, nil
	}
	return "", nil
}

// Load loads Kafka configuration from environment variables with Vault fallback.
// It constructs a ConfigParsed instance with Kafka settings, using defaults if necessary.
//
// Returns the parsed configuration or an error if Vault initialization fails.
func Load() (*ConfigParsed, error) {
	vaultClient, err := NewVaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// Helper function to get config values with Vault and default fallbacks
	getConfigValue := func(vaultKey, defaultValue string) string {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			return vaultValue
		}
		return defaultValue
	}

	// Helper for boolean values
	getConfigBool := func(vaultKey string, defaultValue bool) bool {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			if boolValue, err := strconv.ParseBool(vaultValue); err == nil {
				return boolValue
			}
		}
		return defaultValue
	}

	// Helper for integer values
	getConfigInt := func(vaultKey string, defaultValue int) int {
		if vaultValue, err := vaultClient.GetSecret(vaultKey); err == nil && vaultValue != "" {
			if intValue, err := strconv.Atoi(vaultValue); err == nil {
				return intValue
			}
		}
		return defaultValue
	}

	// Build Kafka configuration
	cfg := &ConfigParsed{
		Kafka: KafkaConfig{
			Brokers:          getConfigValue("KAFKA_BROKERS", ""),
			SMSTopic:         getConfigValue("KAFKA_SMS_TOPIC", "sms-notifications"),
			EmailTopic:       getConfigValue("KAFKA_EMAIL_TOPIC", "email-notifications"),
			InAppTopic:       getConfigValue("KAFKA_INAPP_TOPIC", "inapp-notifications"),
			PushTopic:        getConfigValue("KAFKA_PUSH_TOPIC", "push-notifications"),
			ConsumerGroup:    getConfigValue("KAFKA_CONSUMER_GROUP", "notification-service"),
			SASLEnabled:      getConfigBool("KAFKA_SASL_ENABLED", false),
			SASLUsername:     getConfigValue("KAFKA_SASL_USERNAME", ""),
			SASLPassword:     getConfigValue("KAFKA_SASL_PASSWORD", ""),
			SASLMechanism:    getConfigValue("KAFKA_SASL_MECHANISM", "PLAIN"),
			AutoOffsetReset:  getConfigValue("KAFKA_AUTO_OFFSET_RESET", "earliest"),
			EnableAutoCommit: getConfigBool("KAFKA_ENABLE_AUTO_COMMIT", true),
			SessionTimeoutMs: getConfigInt("KAFKA_SESSION_TIMEOUT_MS", 10000),
		},
	}

	return cfg, nil
}

// getEnv retrieves an environment variable by key, returning an empty string if not set.
func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return ""
}