package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
	PublicURL string
	TTL       time.Duration
}

type S3Storage struct {
	client    *minio.Client
	bucket    string
	publicURL string
	ttl       time.Duration
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 endpoint and bucket are required")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 24 * time.Hour
	}
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}
	exists, err := cli.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := cli.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
			// race-safe: ignore if created by another process
			exists2, err2 := cli.BucketExists(ctx, cfg.Bucket)
			if err2 != nil || !exists2 {
				return nil, fmt.Errorf("create bucket: %w", err)
			}
		}
	}
	return &S3Storage{client: cli, bucket: cfg.Bucket, publicURL: strings.TrimRight(cfg.PublicURL, "/"), ttl: cfg.TTL}, nil
}

func (s *S3Storage) StoreStatement(ctx context.Context, accountID, startDate, endDate, format string, content []byte) (string, string, error) {
	if format == "" {
		format = "xlsx"
	}
	key := fmt.Sprintf("statements/%s/%s_%s.%s", accountID, startDate, endDate, format)
	contentType := contentTypeForFormat(format)
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", "", fmt.Errorf("put object: %w", err)
	}
	// Prefer presigned URL for private buckets; fall back to public URL if configured.
	if s.publicURL != "" {
		u, err := url.Parse(s.publicURL)
		if err == nil {
			u.Path = path.Join(u.Path, s.bucket, key)
			return u.String(), "s3", nil
		}
	}
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, s.ttl, nil)
	if err != nil {
		return "", "", fmt.Errorf("presign url: %w", err)
	}
	return presigned.String(), "s3", nil
}

func contentTypeForFormat(format string) string {
	switch strings.ToLower(format) {
	case "pdf":
		return "application/pdf"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
