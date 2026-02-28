package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/gen/scheduler"
	"task-tracker/internal/scheduler/config"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	conn, err := grpc.Dial(cfg.TaskGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial task grpc: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("close grpc conn: %v", err)
		}
	}()

	client := schedulerpb.NewSchedulerServiceClient(conn)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CallTimeout)
		defer cancel()

		if _, err := client.ProcessRecentExpired(ctx, &emptypb.Empty{}); err != nil {
			st := status.Convert(err)
			if st != nil {
				log.Printf("process recent expired: %s", st.Message())
				return
			}
			log.Printf("process recent expired: %v", err)
		}
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
