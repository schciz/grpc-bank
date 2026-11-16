package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"grpc-bank/internal/account/domain"
	"math/big"

	"github.com/shopspring/decimal"
)

type AccountRepository interface {
	GetbyID(ctx context.Context, id string) (*domain.Account, error)
	Create(ctx context.Context, acc *domain.Account) error
	Transfer(ctx context.Context, fromID, toID string, amount decimal.Decimal) (decimal.Decimal, error)
	Deposit(ctx context.Context, accountID string, amount decimal.Decimal) (decimal.Decimal, error)
	CreateCard(ctx context.Context, card *domain.Card) error
	GetCardsByAccountID(ctx context.Context, accountID string) ([]*domain.Card, error)
	BlockCard(ctx context.Context, cardID string) (*domain.Card, error)
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	if id == "" {
		return nil, fmt.Errorf("account_id can not be empty")
	}

	return s.repo.GetbyID(ctx, id)
}

func (s *AccountService) CreateAccount(ctx context.Context, userID, currency string) (*domain.Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id can not be empty")
	}

	if currency == "" {
		currency = "RUB"
	}

	accNumber, err := generateAccountNumber()
	if err != nil {
		return nil, fmt.Errorf("generation account number error: %w", err)
	}

	acc := &domain.Account{
		UserID:        userID,
		AccountNumber: accNumber,
		Balance:       decimal.NewFromInt(0),
		Currency:      currency,
	}

	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *AccountService) Transfer(ctx context.Context, fromID, toID string, amount decimal.Decimal) (decimal.Decimal, error) {
	if fromID == toID {
		return decimal.Zero, domain.ErrSameAccountTransfer
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, domain.ErrInvalidAmount
	}

	return s.repo.Transfer(ctx, fromID, toID, amount)
}

func (s *AccountService) Deposit(ctx context.Context, accountID string, amount decimal.Decimal) (decimal.Decimal, error) {
	if accountID == "" {
		return decimal.Zero, fmt.Errorf("account_id can not be empty")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, domain.ErrInvalidAmount
	}

	return s.repo.Deposit(ctx, accountID, amount)
}

func (s *AccountService) IssueCard(ctx context.Context, accountID string) (*domain.Card, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account_id can not be empty")
	}
	if _, err := s.repo.GetbyID(ctx, accountID); err != nil {
		return nil, err
	}
	pan, err := generateCardPANMask()
	if err != nil {
		return nil, fmt.Errorf("generate card pan mask: %w", err)
	}
	card := &domain.Card{
		AccountID: accountID,
		PanMask:   pan,
		Status:    domain.CardStatusActive,
	}
	if err := s.repo.CreateCard(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *AccountService) GetCards(ctx context.Context, accountID string) ([]*domain.Card, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account_id can not be empty")
	}
	if _, err := s.repo.GetbyID(ctx, accountID); err != nil {
		return nil, err
	}
	return s.repo.GetCardsByAccountID(ctx, accountID)
}

func (s *AccountService) BlockCard(ctx context.Context, cardID string) (*domain.Card, error) {
	if cardID == "" {
		return nil, fmt.Errorf("card_id can not be empty")
	}
	return s.repo.BlockCard(ctx, cardID)
}

func generateAccountNumber() (string, error) {
	prefix := "40817810"
	maxLimit := new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)
	randomNum, err := rand.Int(rand.Reader, maxLimit)
	if err != nil {
		return "", err
	}
	suffix := fmt.Sprintf("%012d", randomNum)
	return prefix + suffix, nil
}

func generateCardPANMask() (string, error) {
	maxLimit := new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil)
	n, err := rand.Int(rand.Reader, maxLimit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("4081 78** **** %04d", n.Int64()), nil
}
