package handler

type CreateAccountRequest struct {
	UserId   string `json:"user_id"`
	Currency string `json:"currency"`
}

type DepositRequest struct {
	AccountId string `json:"account_id"`
	Amount    string `json:"amount"`
}

type TransferRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	FromAccountId  string `json:"from_account_id"`
	ToAccountId    string `json:"to_account_id"`
	Amount         string `json:"amount"`
}

type AccountResponse struct {
	AccountId     string `json:"account_id"`
	UserId        string `json:"user_id"`
	AccountNumber string `json:"account_number"`
	Balance       string `json:"balance"`
	Currency      string `json:"currency"`
}

type CreateAccountResponse struct {
	AccountId     string `json:"account_id"`
	AccountNumber string `json:"account_number"`
}

type DepositResponse struct {
	Status       string `json:"status"`
	BalanceAfter string `json:"balance_after"`
}

type TransferResponse struct {
	PaymentId string `json:"payment_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CardResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	PanMask   string `json:"pan_mask"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CardsResponse struct {
	Cards []CardResponse `json:"cards"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

type StatementRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type StatementResponse struct {
	AccountID   string `json:"account_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	DownloadURL string `json:"download_url"`
	Format      string `json:"format"`
	Storage     string `json:"storage"`
}