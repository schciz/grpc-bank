package config

import (
	"os"

	"grpc-bank/pkg/env"
)

const (
	grpcPortKey        = "PAYMENT_SERVICE_GRPC_PORT"
	dbURLKey           = "PAYMENT_SERVICE_DB_URL"
	databaseKey        = "PAYMENTS_DB_NAME"
	accountAddrKey     = "ACCOUNT_SERVICE_ADDR"
	auditAddrKey       = "AUDIT_SERVICE_ADDR"
	kafkaBrokersKey    = "KAFKA_BROKERS"
	kafkaTopicKey      = "KAFKA_TRANSFER_TOPIC"
)

type Config struct {
	GRPCPort           string
	DBURL              string
	AccountServiceAddr string
	AuditServiceAddr   string
	KafkaBrokers       string
	KafkaTransferTopic string
}

func Load() (*Config, error) {
	if err := env.Required(grpcPortKey, accountAddrKey, auditAddrKey, kafkaBrokersKey, kafkaTopicKey); err != nil {
		return nil, err
	}

	dbURL, err := env.DatabaseURL(dbURLKey, databaseKey)
	if err != nil {
		return nil, err
	}

	return &Config{
		GRPCPort:           os.Getenv(grpcPortKey),
		DBURL:              dbURL,
		AccountServiceAddr: os.Getenv(accountAddrKey),
		AuditServiceAddr:   os.Getenv(auditAddrKey),
		KafkaBrokers:       os.Getenv(kafkaBrokersKey),
		KafkaTransferTopic: os.Getenv(kafkaTopicKey),
	}, nil
}
