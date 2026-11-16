package handler

import (
	"context"

	auditpb "grpc-bank/pkg/api/audit"
	"grpc-bank/internal/audit/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuditHandler struct {
	auditpb.UnimplementedAuditServiceServer
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) LogEvent(ctx context.Context, req *auditpb.LogEventRequest) (*auditpb.LogEventResponse, error) {
	if req.GetAction() == "" {
		return nil, status.Error(codes.InvalidArgument, "action is required")
	}
	event, err := h.svc.LogEvent(ctx, req.GetUserId(), req.GetAction(), req.GetIpAddress(), req.GetPayloadJson())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "log event failed: %v", err)
	}
	return &auditpb.LogEventResponse{Id: event.ID, Status: "RECORDED"}, nil
}
