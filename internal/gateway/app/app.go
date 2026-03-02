package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accountpb "task-tracker/gen/external/account"
	taskpb "task-tracker/gen/external/task"
	"task-tracker/internal/gateway/config"
	"task-tracker/pkg/logger"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	accountConn, err := grpc.NewClient(cfg.AccountGRPCAddr, dialOpts...)
	if err != nil {
		logger.Log.Fatalf("dial account grpc: %v", err)
	}
	defer func() {
		if err := accountConn.Close(); err != nil {
			logger.Log.Infof("close account grpc: %v", err)
		}
	}()

	taskConn, err := grpc.NewClient(cfg.TaskGRPCAddr, dialOpts...)
	if err != nil {
		logger.Log.Fatalf("dial task grpc: %v", err)
	}
	defer func() {
		if err := taskConn.Close(); err != nil {
			logger.Log.Infof("close task grpc: %v", err)
		}
	}()

	if err := accountpb.RegisterAuthServiceHandler(ctx, mux, accountConn); err != nil {
		logger.Log.Fatalf("register auth handler: %v", err)
	}
	if err := taskpb.RegisterTaskServiceHandler(ctx, mux, taskConn); err != nil {
		logger.Log.Fatalf("register task handler: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Log.Infof("gateway http listen: %s", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("gateway serve: %v", err)
		}
	case <-sigCh:
		logger.Log.Infof("gateway shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Infof("gateway shutdown: %v", err)
	}
}
