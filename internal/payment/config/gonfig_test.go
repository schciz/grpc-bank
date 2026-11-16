package config

import (
	"strings"
	"testing"
)

func setPostgres(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_USER", "bank")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_HOST", "db.example")
	t.Setenv("POSTGRES_PORT", "6543")
	t.Setenv("POSTGRES_SSLMODE", "require")
	t.Setenv("PAYMENTS_DB_NAME", "payments_db")
}

func TestLoad(t *testing.T) {
	t.Run("builds the database url", func(t *testing.T) {
		setPostgres(t)
		t.Setenv("PAYMENT_SERVICE_GRPC_PORT", "60052")
		t.Setenv("PAYMENT_SERVICE_DB_URL", "")
		t.Setenv("ACCOUNT_SERVICE_ADDR", "account:60051")
		t.Setenv("AUDIT_SERVICE_ADDR", "audit:60054")
		t.Setenv("KAFKA_BROKERS", "kafka:9092")
		t.Setenv("KAFKA_TRANSFER_TOPIC", "banking.transfers.v1")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.GRPCPort != "60052" {
			t.Fatalf("expected port 60052, got %q", cfg.GRPCPort)
		}
		want := "postgres://bank:secret@db.example:6543/payments_db?sslmode=require"
		if cfg.DBURL != want {
			t.Fatalf("expected %q, got %q", want, cfg.DBURL)
		}
		if cfg.AccountServiceAddr != "account:60051" {
			t.Fatalf("unexpected account service address %q", cfg.AccountServiceAddr)
		}
		if cfg.AuditServiceAddr != "audit:60054" {
			t.Fatalf("unexpected audit service address %q", cfg.AuditServiceAddr)
		}
		if cfg.KafkaBrokers != "kafka:9092" || cfg.KafkaTransferTopic != "banking.transfers.v1" {
			t.Fatalf("unexpected kafka config: %#v", cfg)
		}
	})

	t.Run("explicit url wins", func(t *testing.T) {
		setPostgres(t)
		t.Setenv("PAYMENT_SERVICE_GRPC_PORT", "60052")
		t.Setenv("PAYMENT_SERVICE_DB_URL", "postgres://user:pass@db:5432/custom")
		t.Setenv("ACCOUNT_SERVICE_ADDR", "account:60051")
		t.Setenv("AUDIT_SERVICE_ADDR", "audit:60054")
		t.Setenv("KAFKA_BROKERS", "kafka:9092")
		t.Setenv("KAFKA_TRANSFER_TOPIC", "banking.transfers.v1")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DBURL != "postgres://user:pass@db:5432/custom" {
			t.Fatalf("unexpected db url %q", cfg.DBURL)
		}
	})

	t.Run("reports missing variables", func(t *testing.T) {
		setPostgres(t)
		t.Setenv("PAYMENT_SERVICE_GRPC_PORT", "60052")
		t.Setenv("PAYMENT_SERVICE_DB_URL", "")
		t.Setenv("ACCOUNT_SERVICE_ADDR", "account:60051")
		t.Setenv("AUDIT_SERVICE_ADDR", "audit:60054")
		t.Setenv("KAFKA_BROKERS", "")
		t.Setenv("KAFKA_TRANSFER_TOPIC", "banking.transfers.v1")

		_, err := Load()
		if err == nil {
			t.Fatal("expected an error")
		}
		if !strings.Contains(err.Error(), "KAFKA_BROKERS") {
			t.Fatalf("error %q does not name the missing variable", err)
		}
	})
}
