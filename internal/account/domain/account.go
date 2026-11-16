package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidAmount        = errors.New("the transfer amount must be strictly positive")
	ErrSameAccountTransfer  = errors.New("you cannot transfer funds to the same account")
	ErrAccountAlreadyExists = errors.New("the user already has an account in this currency")
	ErrCardNotFound         = errors.New("card not found")
)

const (
	CardStatusActive  = "ACTIVE"
	CardStatusBlocked = "BLOCKED"
)

type Account struct {
	ID            string
	UserID        string
	AccountNumber string
	Balance       decimal.Decimal
	Currency      string
	CreatedAt     time.Time
}

type Card struct {
	ID        string
	AccountID string
	PanMask   string
	Status    string
	CreatedAt time.Time
}
