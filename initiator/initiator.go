// Package initiator provides initialization and cleanup logic
// for notification-related services such as Kafka producer and configuration loading.
package initiator

import (
	"github.com/dawit-go/notification-kafka-lib/config"
	producer "github.com/dawit-go/notification-kafka-lib/kafka"
	"gitlab.com/bersufekadgetachew/cbe-super-app-shared/shared/utils"
)

// NotificationServices holds initialized notification-related services and configuration.
type NotificationServices struct {
	Producer *producer.NotificationProducer // Kafka producer instance for publishing messages
	Config   *config.ConfigParsed           // Loaded configuration including email and Kafka settings
}

// InitializeNotificationServices loads configuration from Vault and
// creates a Kafka NotificationProducer instance with the given logger.
//
// Returns initialized NotificationServices or an error if loading config or
// producer creation fails.
func InitializeNotificationServices(logger utils.Logger) (*NotificationServices, error) {
	cfg, err := config.Load()
	if err != nil {
		logger.Errorf("Failed to load config: %v", err)
		return nil, err
	}

	prod, err := producer.NewNotificationProducer(cfg.Kafka, logger)
	if err != nil {
		logger.Errorf("Failed to initialize Kafka producer: %v", err)
		return nil, err
	}

	logger.Infof("Notification services initialized successfully")

	return &NotificationServices{
		Producer: prod,
		Config:   cfg,
	}, nil
}

// Cleanup gracefully closes any active connections or resources,
// such as the Kafka producer, to ensure clean shutdown.
func (ns *NotificationServices) Cleanup() {
	if ns.Producer != nil {
		ns.Producer.Close()
	}
}
