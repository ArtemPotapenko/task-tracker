package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/proto-lib/gen/private/scheduler"
	"task-tracker/shared-libs/pkg/logger"
	"task-tracker/task-service/internal/domain"
	"task-tracker/task-service/internal/usecase"
)

type SchedulerHandler struct {
	schedulerpb.UnimplementedSchedulerServiceServer
	svc *usecase.TaskService
}

func NewSchedulerHandler(svc *usecase.TaskService) *SchedulerHandler {
	return &SchedulerHandler{svc: svc}
}

func (h *SchedulerHandler) ProcessRecentExpired(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := h.svc.ProcessRecentExpired(ctx); err != nil {
		logger.Log.Infof("grpc process recent expired: err=%v", err)
		return nil, mapSchedulerError(err)
	}
	logger.Log.Infof("grpc process recent expired: ok")
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
