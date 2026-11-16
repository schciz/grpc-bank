package storage

import (
	"context"
	"fmt"
	"time"
)

// ObjectStorage is an integration point with external S3-compatible storage.
// MockStorage is used by default.
type ObjectStorage interface {
	StoreStatement(ctx context.Context, accountID, startDate, endDate, format string, content []byte) (url string, err error)
}

type MockStorage struct {
	BaseURL string
	TTL     time.Duration
	Objects map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{BaseURL: "https://mock-storage.local", TTL: 24 * time.Hour, Objects: make(map[string][]byte)}
}

func (m *MockStorage) StoreStatement(_ context.Context, accountID, startDate, endDate, format string, content []byte) (string, error) {
	if format == "" {
		format = "xlsx"
	}
	key := fmt.Sprintf("statements/%s/%s_%s.%s", accountID, startDate, endDate, format)
	if m.Objects != nil {
		m.Objects[key] = append([]byte(nil), content...)
	}
	return m.BaseURL + "/" + key, nil
}
