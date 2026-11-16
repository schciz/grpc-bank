package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"grpc-bank/internal/report/config"
	"grpc-bank/internal/report/handler"
	"grpc-bank/internal/report/repository"
	"grpc-bank/internal/report/service"
	"grpc-bank/internal/report/storage"
	reportpb "grpc-bank/pkg/api/report"
	"grpc-bank/pkg/env"
	"grpc-bank/pkg/postgres"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := env.Load(env.File()); err != nil {
		log.Fatalf("env load error: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.NewPGXPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("postgres connect error: %v", err)
	}
	defer pool.Close()

	repo := repository.NewReportRepository(pool)
	st, err := storage.NewS3Storage(ctx, storage.S3Config{
		Endpoint:  cfg.S3Endpoint,
		Bucket:    cfg.S3Bucket,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		UseSSL:    cfg.S3UseSSL,
		Region:    cfg.S3Region,
		PublicURL: cfg.S3PublicBaseURL,
	})
	if err != nil {
		log.Fatalf("s3 storage init error: %v", err)
	}
	svc := service.NewReportService(repo, st)
	h := handler.NewReportHandler(svc)

	grpcServer := grpc.NewServer()
	reportpb.RegisterReportServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen port %s: %v", cfg.GRPCPort, err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down report-service...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("Report-service successfully started on port :%s\n", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("gRPC server crashed: %v", err)
	}
	log.Println("Report-service stopped correctly.")
}
