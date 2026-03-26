package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"task-tracker/gateway/internal/config"
	accountpb "task-tracker/proto-lib/gen/public/account"
	taskpb "task-tracker/proto-lib/gen/public/task"
	"task-tracker/shared-libs/pkg/logger"
	"task-tracker/shared-libs/pkg/openapiui"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gatewayMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	accountConn, err := grpc.NewClient(cfg.AccountGRPCAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("dial account grpc: %w", err)
	}
	defer func() {
		if err := accountConn.Close(); err != nil {
			logger.Log.Infof("close account grpc: %v", err)
		}
	}()

	taskConn, err := grpc.NewClient(cfg.TaskGRPCAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("dial task grpc: %w", err)
	}
	defer func() {
		if err := taskConn.Close(); err != nil {
			logger.Log.Infof("close task grpc: %v", err)
		}
	}()

	if err := accountpb.RegisterAuthServiceHandler(ctx, gatewayMux, accountConn); err != nil {
		return fmt.Errorf("register auth handler: %w", err)
	}
	if err := taskpb.RegisterTaskServiceHandler(ctx, gatewayMux, taskConn); err != nil {
		return fmt.Errorf("register task handler: %w", err)
	}

	handler := http.NewServeMux()
	openapiui.Mount(handler, cfg.OpenAPIDir)
	handler.Handle("/", gatewayMux)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
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
			return fmt.Errorf("gateway serve: %w", err)
		}
	case <-sigCh:
		logger.Log.Infof("gateway shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Infof("gateway shutdown: %v", err)
	}

	return nil
}
