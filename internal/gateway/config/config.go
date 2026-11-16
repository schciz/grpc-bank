package config

import (
	"os"

	"grpc-bank/pkg/env"
)

const (
	httpPortKey    = "GATEWAY_HTTP_PORT"
	accountAddrKey = "ACCOUNT_SERVICE_ADDR"
	paymentAddrKey = "PAYMENT_SERVICE_ADDR"
	reportAddrKey  = "REPORT_SERVICE_ADDR"
)

type Config struct {
	HTTPPort           string
	AccountServiceAddr string
	PaymentServiceAddr string
	ReportServiceAddr  string
}

func Load() (*Config, error) {
	if err := env.Required(httpPortKey, accountAddrKey, paymentAddrKey, reportAddrKey); err != nil {
		return nil, err
	}

	return &Config{
		HTTPPort:           os.Getenv(httpPortKey),
		AccountServiceAddr: os.Getenv(accountAddrKey),
		PaymentServiceAddr: os.Getenv(paymentAddrKey),
		ReportServiceAddr:  os.Getenv(reportAddrKey),
	}, nil
}
