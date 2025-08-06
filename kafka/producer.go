// Package producer provides a Kafka notification producer for publishing
// messages to Kafka topics with support for synchronous delivery confirmation.
package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CBE-Super-App/notification-kafka-lib/config"
	"github.com/CBE-Super-App/notification-kafka-lib/dto"
	"github.com/IBM/sarama"
	"gitlab.com/bersufekadgetachew/cbe-super-app-shared/shared/utils"
)

// NotificationProducer wraps a Sarama SyncProducer to publish notification messages
// to Kafka topics. It supports configuration, graceful close, and synchronous delivery confirmation.
type NotificationProducer struct {
	producer sarama.SyncProducer
	logger   utils.Logger
	config   config.KafkaConfig
	mu       sync.Mutex
	closed   bool
}

// NewNotificationProducer creates a new NotificationProducer instance using the
// provided KafkaConfig and logger. It configures the Sarama producer with
// specified brokers, SASL auth, and producer options.
//
// Returns an error if the brokers list is empty or if the producer fails to initialize.
func NewNotificationProducer(cfg config.KafkaConfig, logger utils.Logger) (*NotificationProducer, error) {
	if cfg.Brokers == "" {
		return nil, fmt.Errorf("Kafka brokers not configured")
	}

	brokers := strings.Split(cfg.Brokers, ",")
	for i, broker := range brokers {
		brokers[i] = strings.TrimSpace(broker)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
	kafkaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	kafkaConfig.Version = sarama.V2_6_0_0

	if cfg.SASLEnabled {
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.SASL.User = cfg.SASLUsername
		kafkaConfig.Net.SASL.Password = cfg.SASLPassword
		kafkaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(cfg.SASLMechanism)
	}

	producer, err := sarama.NewSyncProducer(brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &NotificationProducer{
		producer: producer,
		logger:   logger,
		config:   cfg,
	}, nil
}

// Close gracefully closes the Kafka producer, releasing all resources.
// It is safe to call multiple times; subsequent calls have no effect.
func (np *NotificationProducer) Close() {
	np.mu.Lock()
	defer np.mu.Unlock()

	if np.closed {
		return
	}

	np.closed = true
	if err := np.producer.Close(); err != nil {
		np.logger.Errorf("Failed to close Kafka producer: %v", err)
	} else {
		np.logger.Infof("Kafka producer closed successfully")
	}
}

// PublishMessage publishes a notification message with the specified msgType and payload
// to the given Kafka topic. The message is marshaled from a NotificationMessage DTO
// and sent synchronously with delivery confirmation.
//
// Returns an error if message creation, marshaling, or sending fails.
func (np *NotificationProducer) PublishMessage(ctx context.Context, msgType, topic string, payload interface{}) error {
	notificationMsg, err := dto.NewNotificationMessage(fmt.Sprintf("%s-%d", msgType, time.Now().UnixNano()), msgType, payload)
	if err != nil {
		return fmt.Errorf("failed to create notification message: %w", err)
	}

	messageBytes, err := json.Marshal(notificationMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(messageBytes),
		Headers: []sarama.RecordHeader{
			{Key: []byte("message_id"), Value: []byte(notificationMsg.ID)},
			{Key: []byte("type"), Value: []byte(msgType)},
			{Key: []byte("timestamp"), Value: []byte(notificationMsg.CreatedAt.Format(time.RFC3339))},
		},
	}

	return np.produceAndWait(ctx, kafkaMsg, notificationMsg.ID, topic)
}

// produceAndWait sends the Kafka message asynchronously but waits for delivery confirmation,
// respecting context cancellation or a timeout of 30 seconds.
//
// Returns an error if the message fails to send or if the context is cancelled or times out.
func (np *NotificationProducer) produceAndWait(ctx context.Context, kafkaMsg *sarama.ProducerMessage, messageID, topic string) error {
	done := make(chan error, 1)

	go func() {
		partition, offset, err := np.safeSendMessage(kafkaMsg)
		if err != nil {
			done <- fmt.Errorf("failed to send Kafka message: %w", err)
			return
		}
		np.logger.Infof("Kafka message sent successfully | ID: %s | Topic: %s | Partition: %d | Offset: %d", messageID, topic, partition, offset)
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout while waiting for message delivery")
	}
}

// safeSendMessage sends the given Kafka message under mutex protection to ensure
// the producer is not closed while sending. It returns partition and offset on success.
//
// Returns an error if the producer is closed.
func (np *NotificationProducer) safeSendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	np.mu.Lock()
	defer np.mu.Unlock()

	if np.closed {
		return 0, 0, fmt.Errorf("producer is closed")
	}

	return np.producer.SendMessage(msg)
}
