package repository

import (
	"context"
	"fmt"
	"time"

	"grpc-bank/internal/report/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

func (r *ReportRepository) ListTransactions(ctx context.Context, accountID string, from, to time.Time) ([]domain.StatementTransaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount::text, status, created_at
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
		  AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at ASC, id ASC;
	`
	rows, err := r.pool.Query(ctx, query, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.StatementTransaction, 0)
	for rows.Next() {
		var t domain.StatementTransaction
		if err := rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Status, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows transactions: %w", err)
	}
	return out, nil
}
