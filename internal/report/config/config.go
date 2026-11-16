package config

import (
	"os"

	"grpc-bank/pkg/env"
)

const (
	grpcPortKey     = "REPORT_SERVICE_GRPC_PORT"
	dbURLKey        = "REPORT_SERVICE_DB_URL"
	databaseKey     = "PAYMENTS_DB_NAME"
	s3EndpointKey   = "S3_ENDPOINT"
	s3BucketKey     = "S3_BUCKET"
	s3AccessKey     = "S3_ACCESS_KEY"
	s3SecretKey     = "S3_SECRET_KEY"
	s3UseSSLKey     = "S3_USE_SSL"
	s3RegionKey     = "S3_REGION"
	s3PublicBaseKey = "S3_PUBLIC_BASE_URL"
)

type Config struct {
	GRPCPort       string
	DBURL          string
	S3Endpoint     string
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
	S3UseSSL       bool
	S3Region       string
	S3PublicBaseURL string
}

func Load() (*Config, error) {
	if err := env.Required(grpcPortKey, s3EndpointKey, s3BucketKey, s3AccessKey, s3SecretKey); err != nil {
		return nil, err
	}
	dbURL, err := env.DatabaseURL(dbURLKey, databaseKey)
	if err != nil {
		return nil, err
	}
	useSSL := os.Getenv(s3UseSSLKey) == "true" || os.Getenv(s3UseSSLKey) == "1"
	return &Config{
		GRPCPort:        os.Getenv(grpcPortKey),
		DBURL:           dbURL,
		S3Endpoint:      os.Getenv(s3EndpointKey),
		S3Bucket:        os.Getenv(s3BucketKey),
		S3AccessKey:     os.Getenv(s3AccessKey),
		S3SecretKey:     os.Getenv(s3SecretKey),
		S3UseSSL:        useSSL,
		S3Region:        os.Getenv(s3RegionKey),
		S3PublicBaseURL: os.Getenv(s3PublicBaseKey),
	}, nil
}
