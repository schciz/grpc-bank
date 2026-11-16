package config

import (
	"os"
	"strconv"
	"time"

	"grpc-bank/pkg/env"
)

const (
	httpPortKey     = "GATEWAY_HTTP_PORT"
	accountAddrKey  = "ACCOUNT_SERVICE_ADDR"
	paymentAddrKey  = "PAYMENT_SERVICE_ADDR"
	reportAddrKey   = "REPORT_SERVICE_ADDR"
	auditAddrKey    = "AUDIT_SERVICE_ADDR"
	dbURLKey        = "GATEWAY_DB_URL"
	databaseKey     = "ACCOUNTS_DB_NAME"
	jwtSecretKey    = "JWT_SECRET"
	jwtTTLSecondsKey = "JWT_TTL_SECONDS"
)

type Config struct {
	HTTPPort           string
	AccountServiceAddr string
	PaymentServiceAddr string
	ReportServiceAddr  string
	AuditServiceAddr   string
	DBURL              string
	JWTSecret          string
	JWTTTL             time.Duration
}

func Load() (*Config, error) {
	if err := env.Required(httpPortKey, accountAddrKey, paymentAddrKey, reportAddrKey, auditAddrKey, jwtSecretKey); err != nil {
		return nil, err
	}
	dbURL, err := env.DatabaseURL(dbURLKey, databaseKey)
	if err != nil {
		return nil, err
	}
	ttl := time.Hour
	if v := os.Getenv(jwtTTLSecondsKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Second
		}
	}
	return &Config{
		HTTPPort:           os.Getenv(httpPortKey),
		AccountServiceAddr: os.Getenv(accountAddrKey),
		PaymentServiceAddr: os.Getenv(paymentAddrKey),
		ReportServiceAddr:  os.Getenv(reportAddrKey),
		AuditServiceAddr:   os.Getenv(auditAddrKey),
		DBURL:              dbURL,
		JWTSecret:          os.Getenv(jwtSecretKey),
		JWTTTL:             ttl,
	}, nil
}
