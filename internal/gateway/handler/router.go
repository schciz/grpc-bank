package handler

import (
	"grpc-bank/internal/gateway/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *GatewayHandler) *chi.Mux {
	return NewRouterWithAuth(h, nil)
}

func NewRouterWithAuth(h *GatewayHandler, authSvc *auth.Service) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if authSvc != nil {
		r.Use(auth.Middleware(authSvc))
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", h.Login)
		r.Group(func(r chi.Router) {
			if authSvc != nil {
				r.Use(auth.RequireAuth)
			}
			r.Post("/accounts", h.CreateAccount)
			r.Get("/accounts/{id}", h.GetAccount)
			r.Post("/accounts/deposit", h.Deposit)
			r.Post("/accounts/{accountId}/cards", h.IssueCard)
			r.Get("/accounts/{accountId}/cards", h.GetCards)
			r.Get("/accounts/{id}/statement", h.GetStatement)
			r.Post("/cards/{id}/block", h.BlockCard)
			r.Post("/transfers", h.CreateTransfer)
		})
	})

	return r
}