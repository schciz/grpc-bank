package repository

import (
	"context"
	"fmt"

	"grpc-bank/internal/audit/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) AppendEvent(ctx context.Context, event *domain.AuditEvent) error {
	query := `
		INSERT INTO audit_events (user_id, action, ip_address, payload)
		VALUES (NULLIF($1, '')::uuid, $2, NULLIF($3, '')::inet, $4::jsonb)
		RETURNING id, created_at;
	`
	payload := event.Payload
	if payload == "" {
		payload = "{}"
	}
	if err := r.pool.QueryRow(ctx, query, event.UserID, event.Action, event.IPAddress, payload).Scan(&event.ID, &event.CreatedAt); err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}
