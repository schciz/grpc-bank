package handler

import (
	"context"

	reportpb "grpc-bank/pkg/api/report"
	"grpc-bank/internal/report/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReportHandler struct {
	reportpb.UnimplementedReportServiceServer
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GenerateReport(ctx context.Context, req *reportpb.GenerateReportRequest) (*reportpb.GenerateReportResponse, error) {
	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}
	if req.GetStartDate() == "" || req.GetEndDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}
	rep, err := h.svc.GenerateReport(ctx, req.GetAccountId(), req.GetStartDate(), req.GetEndDate(), req.GetFormat())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "generate report failed: %v", err)
	}
	return &reportpb.GenerateReportResponse{
		AccountId:   rep.AccountID,
		StartDate:   rep.StartDate,
		EndDate:     rep.EndDate,
		DownloadUrl: rep.DownloadURL,
		Format:      rep.Format,
		Storage:     rep.Storage,
		ExpiresIn:   rep.ExpiresIn,
	}, nil
}
