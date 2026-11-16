package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"grpc-bank/internal/report/domain"
	"grpc-bank/internal/report/storage"
)

type TransactionReader interface {
	ListTransactions(ctx context.Context, accountID string, from, to time.Time) ([]domain.StatementTransaction, error)
}

type ReportService struct {
	reader  TransactionReader
	storage storage.ObjectStorage
}

func NewReportService(reader TransactionReader, st storage.ObjectStorage) *ReportService {
	if st == nil {
		st = storage.NewMockStorage()
	}
	return &ReportService{reader: reader, storage: st}
}

func (s *ReportService) GenerateReport(ctx context.Context, accountID, startDate, endDate, format string) (*domain.StatementReport, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account_id is required")
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("start_date must be <= end_date")
	}
	if end.Sub(start) > 365*24*time.Hour {
		return nil, fmt.Errorf("date range must not exceed 365 days")
	}
	if format == "" {
		format = "xlsx"
	}
	format = strings.ToLower(format)
	if format != "xlsx" && format != "pdf" {
		return nil, fmt.Errorf("unsupported format")
	}

	// Inclusive end-of-day for end date.
	from := start.UTC()
	to := end.UTC().Add(24*time.Hour - time.Nanosecond)
	txs, err := s.reader.ListTransactions(ctx, accountID, from, to)
	if err != nil {
		return nil, err
	}

	// Mocked export generation: build simple CSV-like bytes and store via mocked S3 storage.
	content, err := buildStatementContent(accountID, startDate, endDate, txs)
	if err != nil {
		return nil, err
	}
	url, err := s.storage.StoreStatement(ctx, accountID, startDate, endDate, format, content)
	if err != nil {
		return nil, fmt.Errorf("store statement: %w", err)
	}

	return &domain.StatementReport{
		AccountID:   accountID,
		StartDate:   startDate,
		EndDate:     endDate,
		Format:      format,
		DownloadURL: url,
		Storage:     "mock-s3",
		ExpiresIn:   "24h",
	}, nil
}

func buildStatementContent(accountID, startDate, endDate string, txs []domain.StatementTransaction) ([]byte, error) {
	var b strings.Builder
	b.WriteString("statement for account ")
	b.WriteString(accountID)
	b.WriteString(" ")
	b.WriteString(startDate)
	b.WriteString("..")
	b.WriteString(endDate)
	b.WriteString("\n")
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"created_at", "id", "from_account_id", "to_account_id", "amount", "status"})
	for _, t := range txs {
		_ = w.Write([]string{t.CreatedAt.UTC().Format(time.RFC3339), t.ID, t.FromAccountID, t.ToAccountID, t.Amount, t.Status})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}
