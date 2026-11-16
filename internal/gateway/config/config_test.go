package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	setPostgres := func(t *testing.T) {
		t.Helper()
		t.Setenv("POSTGRES_USER", "bank")
		t.Setenv("POSTGRES_PASSWORD", "secret")
		t.Setenv("POSTGRES_HOST", "db.example")
		t.Setenv("POSTGRES_PORT", "6543")
		t.Setenv("POSTGRES_SSLMODE", "require")
		t.Setenv("ACCOUNTS_DB_NAME", "accounts_db")
	}
	t.Run("reads environment", func(t *testing.T) {
		setPostgres(t)
		t.Setenv("GATEWAY_HTTP_PORT", "9080")
		t.Setenv("ACCOUNT_SERVICE_ADDR", "account:60051")
		t.Setenv("PAYMENT_SERVICE_ADDR", "payment:60052")
		t.Setenv("REPORT_SERVICE_ADDR", "report:60053")
		t.Setenv("AUDIT_SERVICE_ADDR", "audit:60054")
		t.Setenv("GATEWAY_DB_URL", "")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("JWT_TTL_SECONDS", "3600")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.HTTPPort != "9080" {
			t.Fatalf("expected port 9080, got %q", cfg.HTTPPort)
		}
		if cfg.AccountServiceAddr != "account:60051" {
			t.Fatalf("unexpected account service address %q", cfg.AccountServiceAddr)
		}
		if cfg.PaymentServiceAddr != "payment:60052" {
			t.Fatalf("unexpected payment service address %q", cfg.PaymentServiceAddr)
		}
		if cfg.ReportServiceAddr != "report:60053" {
			t.Fatalf("unexpected report service address %q", cfg.ReportServiceAddr)
		}
		if cfg.AuditServiceAddr != "audit:60054" {
			t.Fatalf("unexpected audit service address %q", cfg.AuditServiceAddr)
		}
		if cfg.JWTSecret != "test-secret" || cfg.JWTTTL.Seconds() != 3600 {
			t.Fatalf("unexpected jwt config: secret=%q ttl=%v", cfg.JWTSecret, cfg.JWTTTL)
		}
		wantDB := "postgres://bank:secret@db.example:6543/accounts_db?sslmode=require"
		if cfg.DBURL != wantDB {
			t.Fatalf("expected %q, got %q", wantDB, cfg.DBURL)
		}
	})

	t.Run("reports missing variables", func(t *testing.T) {
		setPostgres(t)
		t.Setenv("GATEWAY_HTTP_PORT", "9080")
		t.Setenv("ACCOUNT_SERVICE_ADDR", "account:60051")
		t.Setenv("PAYMENT_SERVICE_ADDR", "payment:60052")
		t.Setenv("REPORT_SERVICE_ADDR", "report:60053")
		t.Setenv("AUDIT_SERVICE_ADDR", "audit:60054")
		t.Setenv("GATEWAY_DB_URL", "")
		t.Setenv("JWT_SECRET", "")

		_, err := Load()
		if err == nil {
			t.Fatal("expected an error")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Fatalf("error %q does not name the missing variable", err)
		}
	})
}
