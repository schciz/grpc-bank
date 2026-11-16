package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"grpc-bank/internal/payment/domain"
	auditservice "grpc-bank/internal/audit/service"
	accountpb "grpc-bank/pkg/api/account"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentRepository interface {
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
	CreatePending(ctx context.Context, tx *domain.Transaction) error
	UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus, errorMsg string) error
}

// TransferEventPublisher is an integration point with external systems (Kafka/ABS).
// For local/dev and tests use MockTransferEventPublisher.
type TransferEventPublisher interface {
	PublishTransferCompleted(ctx context.Context, tx *domain.Transaction) error
}

type MockTransferEventPublisher struct {
	Published []*domain.Transaction
	Err       error
}

func (m *MockTransferEventPublisher) PublishTransferCompleted(_ context.Context, tx *domain.Transaction) error {
	if m.Err != nil {
		return m.Err
	}
	if m != nil {
		m.Published = append(m.Published, tx)
	}
	return nil
}

type PaymentService struct {
	repo           PaymentRepository
	accountClient  accountpb.AccountServiceClient
	publisher      TransferEventPublisher
	auditPublisher auditservice.Publisher
}

func NewPaymentService(repo PaymentRepository, accountClient accountpb.AccountServiceClient) *PaymentService {
	return &PaymentService{
		repo:           repo,
		accountClient:  accountClient,
		publisher:      &MockTransferEventPublisher{},
		auditPublisher: auditservice.NewAsyncPublisher(&auditservice.MockPublisher{}),
	}
}

func NewPaymentServiceWithPublisher(repo PaymentRepository, accountClient accountpb.AccountServiceClient, publisher TransferEventPublisher) *PaymentService {
	if publisher == nil {
		publisher = &MockTransferEventPublisher{}
	}
	return &PaymentService{repo: repo, accountClient: accountClient, publisher: publisher, auditPublisher: auditservice.NewAsyncPublisher(&auditservice.MockPublisher{})}
}

func (s *PaymentService) ProcessPayment(
	ctx context.Context,
	idempotencyKey string,
	fromAccountID string,
	toAccountID string,
	amount decimal.Decimal,
) (*domain.Transaction, error) {
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotencyKey is required")
	}
	if fromAccountID == "" || toAccountID == "" {
		return nil, fmt.Errorf("sender and recipient accounts are required")
	}
	if fromAccountID == toAccountID {
		return nil, fmt.Errorf("sender account can't equal recipient account")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount must be strictly positive")
	}

	existingTx, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		switch existingTx.Status {
		case domain.StatusSuccess:
			return existingTx, nil
		case domain.StatusPending:
			return nil, fmt.Errorf("a payment with this key is already being processed")
		case domain.StatusFailed:
			return nil, fmt.Errorf("the previous payment attempt was rejected: %s", existingTx.ErrorMessage)
		}
	} else if !errors.Is(err, domain.ErrTransactionNotFound) {
		return nil, fmt.Errorf("idempotency check error: %w", err)
	}

	tx := &domain.Transaction{
		ID:             idempotencyKey,
		IdempotencyKey: idempotencyKey,
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		Amount:         amount,
		Status:         domain.StatusPending,
	}

	if err := s.repo.CreatePending(ctx, tx); err != nil {
		if errors.Is(err, domain.ErrDuplicateTransaction) {
			return nil, fmt.Errorf("a request with this key has already been recorded")
		}
		return nil, fmt.Errorf("failed to register transaction: %w", err)
	}
	s.publishAudit(ctx, "", "TRANSFER_INITIATED", "", map[string]any{"idempotency_key": idempotencyKey, "from_account_id": fromAccountID, "to_account_id": toAccountID, "amount": amount.String()})

	transferReq := &accountpb.ExecuteTransferRequest{
		FromAccountId: fromAccountID,
		ToAccountId:   toAccountID,
		Amount:        amount.String(),
	}

	_, err = s.accountClient.ExecuteTransfer(ctx, transferReq)
	if err != nil {
		st, ok := status.FromError(err)
		var userErrMsg string

		if ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				userErrMsg = "Insufficient funds in the account"
			case codes.NotFound:
				userErrMsg = "One of the accounts was not found."
			default:
				userErrMsg = fmt.Sprintf("Banking core error: %s", st.Message())
			}
		} else {
			userErrMsg = fmt.Sprintf("Network communication failure: %v", err)
		}

		_ = s.repo.UpdateStatus(ctx, tx.ID, domain.StatusFailed, userErrMsg)
		tx.Status = domain.StatusFailed
		tx.ErrorMessage = userErrMsg
		s.publishAudit(ctx, "", "TRANSFER_FAILED", "", map[string]any{"idempotency_key": idempotencyKey, "from_account_id": fromAccountID, "to_account_id": toAccountID, "amount": amount.String(), "error": userErrMsg})

		return tx, errors.New(userErrMsg)
	}

	if err := s.repo.UpdateStatus(ctx, tx.ID, domain.StatusSuccess, ""); err != nil {
		return nil, fmt.Errorf("платеж проведен, но возникла ошибка обновления статуса: %w", err)
	}

	tx.Status = domain.StatusSuccess
	// Integration with external message bus/ABS is mocked.
	_ = s.publisher.PublishTransferCompleted(ctx, tx)
	s.publishAudit(ctx, "", "TRANSFER_SUCCESS", "", map[string]any{"idempotency_key": idempotencyKey, "from_account_id": fromAccountID, "to_account_id": toAccountID, "amount": amount.String(), "payment_id": tx.ID})
	return tx, nil
}

func (s *PaymentService) publishAudit(ctx context.Context, userID, action, ip string, payload any) {
	if s == nil || s.auditPublisher == nil {
		return
	}
	var body string
	switch v := payload.(type) {
	case string:
		body = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			body = "{}"
		} else {
			body = string(b)
		}
	}
	// Async delivery to Audit Service transport is mocked.
	s.auditPublisher.Publish(ctx, userID, action, ip, body)
}

