package app

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	accountkafka "task-tracker/internal/account/transport/kafka"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	accountpb "task-tracker/gen/account"
	"task-tracker/internal/account/config"
	"task-tracker/internal/account/repo"
	transportgrpc "task-tracker/internal/account/transport/grpc"
	"task-tracker/internal/account/usecase"
	"task-tracker/pkg/db"
	"task-tracker/pkg/jwt"
	pkgkafka "task-tracker/pkg/kafka"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbConn, err := db.Open(context.Background(), cfg.DBDriver, cfg.DBDSN, 5*time.Second)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	userRepo := repo.NewUserRepository(dbConn)
	hasher := usecase.BcryptHasher{Cost: cfg.BcryptCost}
	tokens := jwt.Manager{
		Secret: []byte(cfg.JWTSecret),
		TTL:    cfg.JWTTTL,
	}

	writer, err := pkgkafka.NewWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("init kafka writer: %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("close kafka writer: %v", err)
		}
	}()

	publisher := accountkafka.NewPublisher(writer)
	authSvc := usecase.NewAuthService(&userRepo, hasher, tokens, publisher)
	handler := transportgrpc.NewAuthHandler(authSvc)

	server := grpc.NewServer()
	accountpb.RegisterAuthServiceServer(server, handler)
	usersHandler := transportgrpc.NewUsersHandler(authSvc)
	accountpb.RegisterUsersServiceServer(server, usersHandler)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(lis)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("grpc serve: %v", err)
		}
	case <-sigCh:
		log.Printf("shutting down")
		gracefulStop(server, 5*time.Second)
	}
}

func gracefulStop(server *grpc.Server, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
	case <-timer.C:
		server.Stop()
	}
}
