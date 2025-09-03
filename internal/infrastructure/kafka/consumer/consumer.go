package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"order_service/config"
	"order_service/internal/domain"
	"order_service/internal/logger"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer создает новый Kafka consumer с конфигурацией
func NewConsumer(cfg *config.Config) *Consumer {
	logger.DebugLogger.Println("Initializing Kafka Consumer")
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
	}
}

// ReadMessage читает сообщение из Kafka, декодирует в Order и коммитит
func (c *Consumer) ReadMessage(ctx context.Context) (*domain.Order, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("reading from Kafka cancelled: %w", ctx.Err())
	}
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	logger.InfoLogger.Printf(
		"Message at topic/partition/offset %v/%v/%v",
		msg.Topic,
		msg.Partition,
		msg.Offset,
	)

	data := msg.Value
	order := domain.Order{}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&order); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to commit messages: %w", err)
	}

	return &order, nil
}

// Close закрывает Kafka reader
func (c *Consumer) Close() error {
	return c.reader.Close()
}
