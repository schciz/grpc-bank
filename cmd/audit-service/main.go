package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"grpc-bank/internal/audit/config"
	"grpc-bank/internal/audit/handler"
	"grpc-bank/internal/audit/repository"
	"grpc-bank/internal/audit/service"
	auditpb "grpc-bank/pkg/api/audit"
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

	repo := repository.NewAuditRepository(pool)
	svc := service.NewAuditService(repo)
	h := handler.NewAuditHandler(svc)

	grpcServer := grpc.NewServer()
	auditpb.RegisterAuditServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen port %s: %v", cfg.GRPCPort, err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down audit-service...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("Audit-service successfully started on port :%s\n", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("gRPC server crashed: %v", err)
	}
	log.Println("Audit-service stopped correctly.")
}
