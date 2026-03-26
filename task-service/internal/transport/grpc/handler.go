package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	taskpb "task-tracker/proto-lib/gen/public/task"
	"task-tracker/shared-libs/pkg/logger"
	"task-tracker/task-service/internal/domain"
	"task-tracker/task-service/internal/usecase"
)

type TaskHandler struct {
	taskpb.UnimplementedTaskServiceServer
	svc *usecase.TaskService
}

func NewTaskHandler(svc *usecase.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	token, err := extractTaskToken(ctx)
	if err != nil {
		logger.Log.Infof("grpc get task: missing token")
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	task, err := h.svc.GetByID(ctx, token, req.GetId())
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func (h *TaskHandler) GetTodayTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.TasksResponse, error) {
	token, err := extractTaskToken(ctx)
	if err != nil {
		logger.Log.Infof("grpc get today: missing token")
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	tasks, err := h.svc.GetToday(ctx, token)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TasksResponse{Tasks: toProtoTasks(tasks)}, nil
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	token, err := extractTaskToken(ctx)
	if err != nil {
		logger.Log.Infof("grpc create task: missing token")
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	dueDate := time.Unix(req.GetDueDate(), 0)
	task, err := h.svc.Create(ctx, token, req.GetDescription(), dueDate)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func (h *TaskHandler) UpdateTaskStatus(ctx context.Context, req *taskpb.UpdateTaskStatusRequest) (*taskpb.TaskResponse, error) {
	token, err := extractTaskToken(ctx)
	if err != nil {
		logger.Log.Infof("grpc update task status: missing token")
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	task, err := h.svc.UpdateStatus(ctx, token, req.GetId(), toDomainStatus(req.GetStatus()))
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func extractTaskToken(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("authorization")
		for _, value := range values {
			if token, ok := parseBearerToken(value); ok {
				return token, nil
			}
		}
	}
	return "", errors.New("missing token")
}

func parseBearerToken(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len(value) < len("Bearer ")+1 {
		return "", false
	}
	if !strings.EqualFold(value[:len("Bearer ")], "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(value[len("Bearer "):])
	if token == "" {
		return "", false
	}
	return token, true
}

func toDomainStatus(status taskpb.TaskStatus) domain.TaskStatus {
	switch status {
	case taskpb.TaskStatus_TASK_STATUS_CREATED:
		return domain.CREATED
	case taskpb.TaskStatus_TASK_STATUS_AT_WORK:
		return domain.AT_WORK
	case taskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return domain.COMPLETED
	case taskpb.TaskStatus_TASK_STATUS_EXPIRED:
		return domain.EXPIRED
	default:
		return domain.CREATED
	}
}

func toProtoTask(task domain.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:          task.ID,
		UserId:      task.UserID,
		Description: task.Description,
		Status:      toProtoStatus(task.Status),
		CreatedAt:   task.CreatedAt.Unix(),
		DueDate:     task.DueDate.Unix(),
	}
}

func toProtoTasks(tasks []domain.Task) []*taskpb.Task {
	if len(tasks) == 0 {
		return nil
	}

	result := make([]*taskpb.Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, toProtoTask(task))
	}
	return result
}

func toProtoStatus(status domain.TaskStatus) taskpb.TaskStatus {
	switch status {
	case domain.CREATED:
		return taskpb.TaskStatus_TASK_STATUS_CREATED
	case domain.AT_WORK:
		return taskpb.TaskStatus_TASK_STATUS_AT_WORK
	case domain.COMPLETED:
		return taskpb.TaskStatus_TASK_STATUS_COMPLETED
	case domain.EXPIRED:
		return taskpb.TaskStatus_TASK_STATUS_EXPIRED
	default:
		return taskpb.TaskStatus_TASK_STATUS_CREATED
	}
}

func mapTaskError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, usecase.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
