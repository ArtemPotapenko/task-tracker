package app

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"

	schedulerpb "task-tracker/gen/scheduler"
	taskpb "task-tracker/gen/task"
	"task-tracker/internal/task/config"
	taskkafka "task-tracker/internal/task/kafka"
	"task-tracker/internal/task/repo"
	transportgrpc "task-tracker/internal/task/transport/grpc"
	"task-tracker/internal/task/usecase"
	"task-tracker/pkg/db"
	pkgjwt "task-tracker/pkg/jwt"
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

	taskRepo := repo.NewTaskRepository(dbConn)
	parser := pkgjwt.Parser{Secret: []byte(cfg.JWTSecret)}

	writer, err := pkgkafka.NewWriter(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("init kafka writer: %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("close kafka writer: %v", err)
		}
	}()

	publisher := taskkafka.NewPublisher(writer)
	taskSvc := usecase.NewTaskService(&taskRepo, parser, publisher)
	taskHandler := transportgrpc.NewTaskHandler(taskSvc)
	schedulerHandler := transportgrpc.NewSchedulerHandler(taskSvc)

	server := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(server, taskHandler)
	schedulerpb.RegisterSchedulerServiceServer(server, schedulerHandler)

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
