package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"task-tracker/shared-libs/pkg/logger"
	taskkafka "task-tracker/task-service/internal/transport/kafka"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	schedulerpb "task-tracker/proto-lib/gen/private/scheduler"
	taskpb "task-tracker/proto-lib/gen/public/task"
	"task-tracker/shared-libs/pkg/db"
	pkgjwt "task-tracker/shared-libs/pkg/jwt"
	"task-tracker/shared-libs/pkg/kafka"
	"task-tracker/task-service/internal/config"
	"task-tracker/task-service/internal/repo"
	transportgrpc "task-tracker/task-service/internal/transport/grpc"
	"task-tracker/task-service/internal/usecase"
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

	if err := db.UpSQL(dbConn, "task-service/migrations/00001_create_tasks.sql"); err != nil {
		return fmt.Errorf("run task migrations: %w", err)
	}

	taskRepo := repo.NewTaskRepository(dbConn)
	outboxRepo := repo.NewOutboxRepository(dbConn)
	parser := pkgjwt.Parser{Secret: []byte(cfg.JWTSecret)}

	writer, err := kafka.NewWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("init kafka writer: %w", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Log.Infof("close kafka writer: %v", err)
		}
	}()

	outboxPublisher := taskkafka.NewOutboxPublisher(writer, &outboxRepo, time.Second, 100)
	taskSvc := usecase.NewTaskService(&taskRepo, parser)
	taskHandler := transportgrpc.NewTaskHandler(taskSvc)
	schedulerHandler := transportgrpc.NewSchedulerHandler(taskSvc)

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(validationUnaryServerInterceptor, loggingUnaryServerInterceptor))
	taskpb.RegisterTaskServiceServer(server, taskHandler)
	schedulerpb.RegisterSchedulerServiceServer(server, schedulerHandler)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	errCh := make(chan error, 1)
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		errCh <- server.Serve(lis)
	}()
	go func() {
		if err := outboxPublisher.Run(appCtx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- fmt.Errorf("outbox publisher: %w", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		cancel()
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("grpc serve: %w", err)
		}
	case <-sigCh:
		cancel()
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

type validatingRequest interface {
	ValidateAll() error
}

func validationUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if message, ok := req.(validatingRequest); ok {
		if err := message.ValidateAll(); err != nil {
			logger.Log.Infof("grpc request: method=%s validation err=%v", info.FullMethod, err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}
