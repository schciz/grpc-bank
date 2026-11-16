package publisher

import (
	"context"
	"log"
	"time"

	auditpb "grpc-bank/pkg/api/audit"
)

type GRPCPublisher struct {
	client  auditpb.AuditServiceClient
	timeout time.Duration
}

func NewGRPCPublisher(client auditpb.AuditServiceClient) *GRPCPublisher {
	return &GRPCPublisher{client: client, timeout: 3 * time.Second}
}

func (p *GRPCPublisher) Publish(ctx context.Context, userID, action, ipAddress, payloadJSON string) {
	if p == nil || p.client == nil {
		return
	}
	callCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.timeout)
	defer cancel()
	_, err := p.client.LogEvent(callCtx, &auditpb.LogEventRequest{
		UserId:      userID,
		Action:      action,
		IpAddress:   ipAddress,
		PayloadJson: payloadJSON,
	})
	if err != nil {
		log.Printf("audit publish failed action=%s: %v", action, err)
	}
}
