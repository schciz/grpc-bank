package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"grpc-bank/internal/payment/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *KafkaPublisher) PublishTransferCompleted(ctx context.Context, tx *domain.Transaction) error {
	if p == nil || p.writer == nil || tx == nil {
		return nil
	}
	payload := map[string]any{
		"payment_id":      tx.ID,
		"idempotency_key": tx.IdempotencyKey,
		"from_account_id": tx.FromAccountID,
		"to_account_id":   tx.ToAccountID,
		"amount":          tx.Amount.String(),
		"status":          string(tx.Status),
		"created_at":      tx.CreatedAt.Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal transfer event: %w", err)
	}
	key := tx.IdempotencyKey
	if key == "" {
		key = tx.ID
	}
	msg := kafka.Message{Key: []byte(key), Value: b, Time: time.Now()}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("publish transfer event: %w", err)
	}
	return nil
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
