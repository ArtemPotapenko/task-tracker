package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/gen/scheduler"
	"task-tracker/internal/task/domain"
	"task-tracker/internal/task/usecase"
)

type SchedulerHandler struct {
	schedulerpb.UnimplementedSchedulerServiceServer
	svc usecase.TaskService
}

func NewSchedulerHandler(svc usecase.TaskService) SchedulerHandler {
	return SchedulerHandler{svc: svc}
}

func (h SchedulerHandler) ProcessRecentExpired(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := h.svc.ProcessRecentExpired(ctx); err != nil {
		return nil, mapSchedulerError(err)
	}
	return &emptypb.Empty{}, nil
}

func mapSchedulerError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
