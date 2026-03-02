package app

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	taskkafka "task-tracker/internal/task/transport/kafka"
	"task-tracker/pkg/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	taskpb "task-tracker/gen/external/task"
	schedulerpb "task-tracker/gen/internal/scheduler"
	"task-tracker/internal/task/config"
	"task-tracker/internal/task/repo"
	transportgrpc "task-tracker/internal/task/transport/grpc"
	"task-tracker/internal/task/usecase"
	"task-tracker/pkg/db"
	pkgjwt "task-tracker/pkg/jwt"
	"task-tracker/pkg/kafka"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatalf("load config: %v", err)
	}

	dbConn, err := db.Open(context.Background(), cfg.DBDriver, cfg.DBDSN, 5*time.Second)
	if err != nil {
		logger.Log.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Log.Infof("close db: %v", err)
		}
	}()

	taskRepo := repo.NewTaskRepository(dbConn)
	parser := pkgjwt.Parser{Secret: []byte(cfg.JWTSecret)}

	writer, err := kafka.NewWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	if err != nil {
		logger.Log.Fatalf("init kafka writer: %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Log.Infof("close kafka writer: %v", err)
		}
	}()

	publisher := taskkafka.NewPublisher(writer)
	taskSvc := usecase.NewTaskService(&taskRepo, parser, publisher)
	taskHandler := transportgrpc.NewTaskHandler(taskSvc)
	schedulerHandler := transportgrpc.NewSchedulerHandler(taskSvc)

	server := grpc.NewServer(grpc.UnaryInterceptor(loggingUnaryServerInterceptor))
	taskpb.RegisterTaskServiceServer(server, taskHandler)
	schedulerpb.RegisterSchedulerServiceServer(server, schedulerHandler)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		logger.Log.Fatalf("listen grpc: %v", err)
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
			logger.Log.Fatalf("grpc serve: %v", err)
		}
	case <-sigCh:
		logger.Log.Infof("shutting down")
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
