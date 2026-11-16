package service

import (
	"context"
	"fmt"
	"strings"

	"grpc-bank/internal/audit/domain"
)

type AuditRepository interface {
	AppendEvent(ctx context.Context, event *domain.AuditEvent) error
}

type AuditService struct {
	repo AuditRepository
}

func NewAuditService(repo AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) LogEvent(ctx context.Context, userID, action, ipAddress, payloadJSON string) (*domain.AuditEvent, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if payloadJSON == "" {
		payloadJSON = "{}"
	}
	event := &domain.AuditEvent{
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		Payload:   payloadJSON,
	}
	if err := s.repo.AppendEvent(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

// Publisher is used by other services to emit audit events asynchronously.
// MockPublisher is used for tests and when external/async transport is mocked.
type Publisher interface {
	Publish(ctx context.Context, userID, action, ipAddress, payloadJSON string)
}

type MockPublisher struct {
	Events []domain.AuditEvent
}

func (m *MockPublisher) Publish(_ context.Context, userID, action, ipAddress, payloadJSON string) {
	m.Events = append(m.Events, domain.AuditEvent{UserID: userID, Action: action, IPAddress: ipAddress, Payload: payloadJSON})
}

type AsyncPublisher struct {
	inner Publisher
}

func NewAsyncPublisher(inner Publisher) *AsyncPublisher {
	if inner == nil {
		inner = &MockPublisher{}
	}
	return &AsyncPublisher{inner: inner}
}

func (p *AsyncPublisher) Publish(ctx context.Context, userID, action, ipAddress, payloadJSON string) {
	if p == nil || p.inner == nil {
		return
	}
	go p.inner.Publish(context.WithoutCancel(ctx), userID, action, ipAddress, payloadJSON)
}
