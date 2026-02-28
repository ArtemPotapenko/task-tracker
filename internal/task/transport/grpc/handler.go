package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	taskpb "task-tracker/gen/task"
	"task-tracker/internal/task/domain"
	"task-tracker/internal/task/usecase"
)

type TaskHandler struct {
	taskpb.UnimplementedTaskServiceServer
	svc usecase.TaskService
}

func NewTaskHandler(svc usecase.TaskService) TaskHandler {
	return TaskHandler{svc: svc}
}

func (h TaskHandler) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.TaskResponse, error) {
	if req.GetJwt() == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	task, err := h.svc.GetByID(ctx, req.GetJwt(), req.GetId())
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func (h TaskHandler) GetTodayTasks(ctx context.Context, req *taskpb.GetTasksRequest) (*taskpb.TasksResponse, error) {
	if req.GetJwt() == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	tasks, err := h.svc.GetToday(ctx, req.GetJwt())
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TasksResponse{Tasks: toProtoTasks(tasks)}, nil
}

func (h TaskHandler) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.TaskResponse, error) {
	if req.GetJwt() == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}
	if req.GetDueDate() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid due date")
	}

	dueDate := time.Unix(req.GetDueDate(), 0)
	task, err := h.svc.Create(ctx, req.GetJwt(), req.GetDescription(), dueDate)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func (h TaskHandler) UpdateTaskStatus(ctx context.Context, req *taskpb.UpdateTaskStatusRequest) (*taskpb.TaskResponse, error) {
	if req.GetJwt() == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	statusValue, err := toDomainStatus(req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	task, err := h.svc.UpdateStatus(ctx, req.GetJwt(), req.GetId(), statusValue)
	if err != nil {
		return nil, mapTaskError(err)
	}
	return &taskpb.TaskResponse{Task: toProtoTask(task)}, nil
}

func toDomainStatus(status taskpb.TaskStatus) (domain.TaskStatus, error) {
	switch status {
	case taskpb.TaskStatus_TASK_STATUS_CREATED:
		return domain.CREATED, nil
	case taskpb.TaskStatus_TASK_STATUS_AT_WORK:
		return domain.AT_WORK, nil
	case taskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return domain.COMPLETED, nil
	case taskpb.TaskStatus_TASK_STATUS_EXPIRED:
		return domain.EXPIRED, nil
	default:
		return domain.CREATED, errors.New("unknown status")
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
