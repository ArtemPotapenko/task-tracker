package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/gen/private/scheduler"
	"task-tracker/internal/scheduler/config"
	"task-tracker/pkg/logger"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatalf("load config: %v", err)
	}

	conn, err := grpc.NewClient(cfg.TaskGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatalf("dial task grpc: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Log.Infof("close grpc conn: %v", err)
		}
	}()

	client := schedulerpb.NewSchedulerServiceClient(conn)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CallTimeout)
		defer cancel()

		logger.Log.Infof("scheduler: process recent expired start")
		if _, err := client.ProcessRecentExpired(ctx, &emptypb.Empty{}); err != nil {
			st := status.Convert(err)
			if st != nil {
				logger.Log.Infof("process recent expired: %s", st.Message())
				return
			}
			logger.Log.Infof("process recent expired: %v", err)
			return
		}
		logger.Log.Infof("scheduler: process recent expired ok")
	}

	run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			run()
		case <-sigCh:
			return
		}
	}
}
