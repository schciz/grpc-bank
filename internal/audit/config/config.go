package config

import (
	"os"

	"grpc-bank/pkg/env"
)

const (
	grpcPortKey = "AUDIT_SERVICE_GRPC_PORT"
	dbURLKey    = "AUDIT_SERVICE_DB_URL"
	databaseKey = "AUDIT_DB_NAME"
)

type Config struct {
	GRPCPort string
	DBURL    string
}

func Load() (*Config, error) {
	if err := env.Required(grpcPortKey); err != nil {
		return nil, err
	}
	dbURL, err := env.DatabaseURL(dbURLKey, databaseKey)
	if err != nil {
		return nil, err
	}
	return &Config{GRPCPort: os.Getenv(grpcPortKey), DBURL: dbURL}, nil
}
