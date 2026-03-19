package app

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/gen/private/scheduler"
)

type testSchedulerServer struct {
	schedulerpb.UnimplementedSchedulerServiceServer
	calls atomic.Int32
}

func (s *testSchedulerServer) ProcessRecentExpired(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	s.calls.Add(1)
	return &emptypb.Empty{}, nil
}

func TestRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer lis.Close()

	serverImpl := &testSchedulerServer{}
	server := grpc.NewServer()
	schedulerpb.RegisterSchedulerServiceServer(server, serverImpl)
	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	t.Setenv("TASK_GRPC_ADDR", lis.Addr().String())
	t.Setenv("SCHEDULER_INTERVAL", "20ms")
	t.Setenv("SCHEDULER_CALL_TIMEOUT", "200ms")

	done := make(chan error, 1)
	go func() {
		done <- Run()
	}()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if serverImpl.calls.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if serverImpl.calls.Load() < 2 {
		t.Fatalf("expected at least 2 scheduler calls, got %d", serverImpl.calls.Load())
	}

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("Kill() error = %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("Run() did not exit after SIGTERM")
	}
}
