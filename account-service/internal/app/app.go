package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	accountkafka "task-tracker/account-service/internal/transport/kafka"
	"task-tracker/shared-libs/pkg/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	"task-tracker/account-service/internal/config"
	"task-tracker/account-service/internal/repo"
	transportgrpc "task-tracker/account-service/internal/transport/grpc"
	"task-tracker/account-service/internal/usecase"
	accountinternalpb "task-tracker/proto-lib/gen/private/account"
	accountpb "task-tracker/proto-lib/gen/public/account"
	"task-tracker/shared-libs/pkg/db"
	"task-tracker/shared-libs/pkg/jwt"
	pkgkafka "task-tracker/shared-libs/pkg/kafka"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbConn, err := db.Open(context.Background(), cfg.DBDriver, cfg.DBDSN, 5*time.Second)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Log.Infof("close db: %v", err)
		}
	}()

	if err := db.UpSQL(dbConn, "account-service/migrations/00001_create_users.sql"); err != nil {
		return fmt.Errorf("run account migrations: %w", err)
	}

	userRepo := repo.NewUserRepository(dbConn)
	hasher := &usecase.BcryptHasher{Cost: cfg.BcryptCost}
	tokens := jwt.Manager{
		Secret: []byte(cfg.JWTSecret),
		TTL:    cfg.JWTTTL,
	}

	writer, err := pkgkafka.NewWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("init kafka writer: %w", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Log.Infof("close kafka writer: %v", err)
		}
	}()

	publisher := accountkafka.NewPublisher(writer)
	authSvc := usecase.NewAuthService(&userRepo, hasher, tokens, publisher)
	handler := transportgrpc.NewAuthHandler(authSvc)

	server := grpc.NewServer(grpc.UnaryInterceptor(loggingUnaryServerInterceptor))
	accountpb.RegisterAuthServiceServer(server, handler)
	usersHandler := transportgrpc.NewUsersHandler(authSvc)
	accountinternalpb.RegisterUsersServiceServer(server, usersHandler)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
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
			return fmt.Errorf("grpc serve: %w", err)
		}
	case <-sigCh:
		logger.Log.Infof("shutting down")
		gracefulStop(server, 5*time.Second)
	}

	return nil
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

func loggingUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	if err != nil {
		logger.Log.Infof("grpc request: method=%s duration=%s err=%v", info.FullMethod, time.Since(start), err)
		return resp, err
	}
	logger.Log.Infof("grpc request: method=%s duration=%s ok", info.FullMethod, time.Since(start))
	return resp, nil
}
