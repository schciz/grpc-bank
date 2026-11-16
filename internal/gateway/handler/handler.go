package handler

import (
	"encoding/json"
	accountpb "grpc-bank/pkg/api/account"
	paymentpb "grpc-bank/pkg/api/payment"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayHandler struct {
	accountClient accountpb.AccountServiceClient
	paymentClient paymentpb.PaymentServiceClient
}

func NewGatewayHandler(accountClient accountpb.AccountServiceClient, paymentClient paymentpb.PaymentServiceClient) *GatewayHandler {
	return &GatewayHandler{
		accountClient: accountClient,
		paymentClient: paymentClient,
	}
}

func (h *GatewayHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Login) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}
	// Mocked authentication/integration: no external IdP, issue deterministic demo JWT-like token.
	token := "mock-jwt-" + strings.ReplaceAll(req.Login, " ", "_") + "-" + time.Now().UTC().Format("20060102150405")
	_ = json.NewEncoder(w).Encode(LoginResponse{Token: token, TokenType: "Bearer", ExpiresIn: 3600})
}

func (h *GatewayHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.CreateAccount(r.Context(), &accountpb.CreateAccountRequest{
		UserId:   req.UserId,
		Currency: req.Currency,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := CreateAccountResponse{
		AccountId:     grpcResp.GetAccountId(),
		AccountNumber: grpcResp.GetAccountNumber(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.GetAccount(r.Context(), &accountpb.GetAccountRequest{AccountId: accountID})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := AccountResponse{
		AccountId:     grpcResp.GetAccountId(),
		UserId:        grpcResp.GetUserId(),
		AccountNumber: grpcResp.GetAccountNumber(),
		Balance:       grpcResp.GetBalance(),
		Currency:      grpcResp.GetCurrency(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	grpcResp, err := h.accountClient.Deposit(r.Context(), &accountpb.DepositRequest{
		AccountId: req.AccountId,
		Amount:    req.Amount,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	resp := DepositResponse{Status: grpcResp.Status, BalanceAfter: grpcResp.BalanceAfter}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) IssueCard(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}
	grpcResp, err := h.accountClient.IssueCard(r.Context(), &accountpb.IssueCardRequest{AccountId: accountID})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toCardResponse(grpcResp.GetCard()))
}

func (h *GatewayHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}
	grpcResp, err := h.accountClient.GetCards(r.Context(), &accountpb.GetCardsRequest{AccountId: accountID})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}
	cards := make([]CardResponse, 0, len(grpcResp.GetCards()))
	for _, c := range grpcResp.GetCards() {
		cards = append(cards, toCardResponse(c))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CardsResponse{Cards: cards})
}

func (h *GatewayHandler) BlockCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "id")
	if cardID == "" {
		http.Error(w, "card id is required", http.StatusBadRequest)
		return
	}
	grpcResp, err := h.accountClient.BlockCard(r.Context(), &accountpb.BlockCardRequest{CardId: cardID})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toCardResponse(grpcResp.GetCard()))
}

func (h *GatewayHandler) GetStatement(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		http.Error(w, "account id is required", http.StatusBadRequest)
		return
	}
	start := r.URL.Query().Get("start_date")
	end := r.URL.Query().Get("end_date")
	if start == "" || end == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}
	// Mocked report generation/storage integration (no external Report Service/S3): return a deterministic mock URL.
	resp := StatementResponse{
		AccountID:   accountID,
		StartDate:   start,
		EndDate:     end,
		Format:      "xlsx",
		Storage:     "mock-s3",
		DownloadURL: "https://mock-storage.local/statements/" + accountID + "/" + start + "_" + end + ".xlsx",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GatewayHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if key := strings.TrimSpace(r.Header.Get("X-Idempotency-Key")); key != "" {
		req.IdempotencyKey = key
	}

	grpcResp, err := h.paymentClient.CreatePayment(r.Context(), &paymentpb.CreatePaymentRequest{
		IdempotencyKey: req.IdempotencyKey,
		FromAccountId:  req.FromAccountId,
		ToAccountId:    req.ToAccountId,
		Amount:         req.Amount,
	})
	if err != nil {
		handleGRPCErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TransferResponse{
		PaymentId: grpcResp.GetPaymentId(),
		Status:    grpcResp.GetStatus(),
		CreatedAt: grpcResp.GetCreatedAt(),
	})
}

func toCardResponse(c *accountpb.Card) CardResponse {
	if c == nil {
		return CardResponse{}
	}
	return CardResponse{
		ID:        c.GetId(),
		AccountID: c.GetAccountId(),
		PanMask:   c.GetPanMask(),
		Status:    c.GetStatus(),
		CreatedAt: c.GetCreatedAt(),
	}
}

func handleGRPCErr(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.FailedPrecondition:
		http.Error(w, st.Message(), http.StatusUnprocessableEntity)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

