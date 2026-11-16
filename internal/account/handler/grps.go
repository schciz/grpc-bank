package handler

import (
	"context"
	"errors"
	"time"

	"grpc-bank/internal/account/domain"
	"grpc-bank/internal/account/service"
	accountpb "grpc-bank/pkg/api/account"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountHandler struct {
	accountpb.UnimplementedAccountServiceServer
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) GetAccount(ctx context.Context, req *accountpb.GetAccountRequest) (*accountpb.GetAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	acc, err := h.svc.GetAccount(ctx, req.AccountId)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "Internal Server Errror: %v", err)
	}

	return &accountpb.GetAccountResponse{
		AccountId:     acc.ID,
		UserId:        acc.UserID,
		AccountNumber: acc.AccountNumber,
		Balance:       acc.Balance.String(),
		Currency:      acc.Currency,
	}, nil
}

func (h *AccountHandler) CreateAccount(ctx context.Context, req *accountpb.CreateAccountRequest) (*accountpb.CreateAccountResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	acc, err := h.svc.CreateAccount(ctx, req.GetUserId(), req.GetCurrency())
	if err != nil {
		if errors.Is(err, domain.ErrAccountAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "the user already has an account in this currency")
		}
		return nil, status.Errorf(codes.Internal, "failed to create an account: %v", err)
	}

	return &accountpb.CreateAccountResponse{
		AccountId:     acc.ID,
		AccountNumber: acc.AccountNumber,
	}, nil
}

func (h *AccountHandler) ExecuteTransfer(ctx context.Context, req *accountpb.ExecuteTransferRequest) (*accountpb.ExecuteTransferResponse, error) {
	if req.GetFromAccountId() == "" || req.GetToAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "incorrect amount")
	}

	newBalance, err := h.svc.Transfer(ctx, req.GetFromAccountId(), req.GetToAccountId(), amount)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountNotFound):
			return nil, status.Error(codes.NotFound, "one of the accounts was not found")
		case errors.Is(err, domain.ErrInsufficientFunds):
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds in the account")
		case errors.Is(err, domain.ErrSameAccountTransfer):
			return nil, status.Error(codes.InvalidArgument, "cannot be transferred to the same account")
		default:
			return nil, status.Errorf(codes.Internal, "transaction processing error: %v", err)
		}
	}

	return &accountpb.ExecuteTransferResponse{
		Status:           "SUCCESS",
		FromBalanceAfter: newBalance.String(),
	}, nil
}

func (h *AccountHandler) Deposit(ctx context.Context, req *accountpb.DepositRequest) (*accountpb.DepositResponse, error) {
	amount, err := decimal.NewFromString(req.GetAmount())
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "Incorrect top-up amount")
	}

	newBalance, err := h.svc.Deposit(ctx, req.GetAccountId(), amount)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "deposit error: %v", err)
	}

	return &accountpb.DepositResponse{
		Status:       "SUCCESS",
		BalanceAfter: newBalance.String(),
	}, nil
}

func (h *AccountHandler) IssueCard(ctx context.Context, req *accountpb.IssueCardRequest) (*accountpb.IssueCardResponse, error) {
	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}
	card, err := h.svc.IssueCard(ctx, req.GetAccountId())
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "issue card error: %v", err)
	}
	return &accountpb.IssueCardResponse{Card: toProtoCard(card)}, nil
}

func (h *AccountHandler) GetCards(ctx context.Context, req *accountpb.GetCardsRequest) (*accountpb.GetCardsResponse, error) {
	if req.GetAccountId() == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}
	cards, err := h.svc.GetCards(ctx, req.GetAccountId())
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "get cards error: %v", err)
	}
	out := make([]*accountpb.Card, 0, len(cards))
	for _, c := range cards {
		out = append(out, toProtoCard(c))
	}
	return &accountpb.GetCardsResponse{Cards: out}, nil
}

func (h *AccountHandler) BlockCard(ctx context.Context, req *accountpb.BlockCardRequest) (*accountpb.BlockCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	card, err := h.svc.BlockCard(ctx, req.GetCardId())
	if err != nil {
		if errors.Is(err, domain.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Errorf(codes.Internal, "block card error: %v", err)
	}
	return &accountpb.BlockCardResponse{Card: toProtoCard(card)}, nil
}

func toProtoCard(c *domain.Card) *accountpb.Card {
	if c == nil {
		return nil
	}
	return &accountpb.Card{
		Id:        c.ID,
		AccountId: c.AccountID,
		PanMask:   c.PanMask,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}
