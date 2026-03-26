package app

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	schedulerpb "task-tracker/proto-lib/gen/private/scheduler"
	taskpb "task-tracker/proto-lib/gen/public/task"
	"task-tracker/task-service/internal/domain"
	transportgrpc "task-tracker/task-service/internal/transport/grpc"
	"task-tracker/task-service/internal/usecase"
)

const taskBufSize = 1024 * 1024

type memoryTaskRepo struct {
	nextID int64
	tasks  map[int64]domain.Task
}

func newMemoryTaskRepo() *memoryTaskRepo {
	return &memoryTaskRepo{
		nextID: 1,
		tasks:  make(map[int64]domain.Task),
	}
}

func (r *memoryTaskRepo) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = task
	return task, nil
}

func (r *memoryTaskRepo) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	task, ok := r.tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}
	return task, nil
}

func (r *memoryTaskRepo) GetByIDAndUserID(ctx context.Context, id, userID int64) (domain.Task, error) {
	task, ok := r.tasks[id]
	if !ok || task.UserID != userID {
		return domain.Task{}, domain.ErrNotFound
	}
	return task, nil
}

func (r *memoryTaskRepo) GetByUserIDAndDueDateBetween(ctx context.Context, userID int64, from, to time.Time) ([]domain.Task, error) {
	var result []domain.Task
	for _, task := range r.tasks {
		if task.UserID == userID && !task.DueDate.Before(from) && task.DueDate.Before(to) {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *memoryTaskRepo) GetByDueDateBetween(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
	var result []domain.Task
	for _, task := range r.tasks {
		if !task.DueDate.Before(from) && task.DueDate.Before(to) {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *memoryTaskRepo) GetByDueDateBetweenAndStatusNot(ctx context.Context, from, to time.Time, status domain.TaskStatus) ([]domain.Task, error) {
	var result []domain.Task
	for _, task := range r.tasks {
		if !task.DueDate.Before(from) && task.DueDate.Before(to) && task.Status != status {
			result = append(result, task)
		}
	}
	return result, nil
}

func (r *memoryTaskRepo) UpdateStatusByIDAndUserID(ctx context.Context, id, userID int64, status domain.TaskStatus) (domain.Task, error) {
	task, ok := r.tasks[id]
	if !ok || task.UserID != userID {
		return domain.Task{}, domain.ErrNotFound
	}
	task.Status = status
	r.tasks[id] = task
	return task, nil
}

func (r *memoryTaskRepo) UpdateStatusByIDs(ctx context.Context, ids []int64, status domain.TaskStatus) error {
	for _, id := range ids {
		task, ok := r.tasks[id]
		if !ok {
			continue
		}
		task.Status = status
		r.tasks[id] = task
	}
	return nil
}

type taskTestTokenParser struct{}

func (taskTestTokenParser) ParseUserID(token string) (int64, error) {
	if token == "jwt-user-1" {
		return 1, nil
	}
	if token == "jwt-user-2" {
		return 2, nil
	}
	return 0, errors.New("invalid token")
}

type taskTestPublisher struct{}

func (taskTestPublisher) PublishExpiredSummary(ctx context.Context, summary domain.ExpiredSummary) error {
	return nil
}

func taskAuthContext(token string) context.Context {
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func TestTaskServiceE2E(t *testing.T) {
	repo := newMemoryTaskRepo()
	svc := usecase.NewTaskService(repo, taskTestTokenParser{}, taskTestPublisher{})

	lis := bufconn.Listen(taskBufSize)
	server := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(server, transportgrpc.NewTaskHandler(svc))
	schedulerpb.RegisterSchedulerServiceServer(server, transportgrpc.NewSchedulerHandler(svc))

	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("Serve() error = %v", err)
		}
	}()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	taskClient := taskpb.NewTaskServiceClient(conn)
	schedulerClient := schedulerpb.NewSchedulerServiceClient(conn)

	createResp, err := taskClient.CreateTask(taskAuthContext("jwt-user-1"), &taskpb.CreateTaskRequest{
		Description: "ship feature",
		DueDate:     time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if createResp.GetTask().GetId() == 0 {
		t.Fatalf("CreateTask() id = 0, want non-zero")
	}

	getResp, err := taskClient.GetTask(taskAuthContext("jwt-user-1"), &taskpb.GetTaskRequest{
		Id: createResp.GetTask().GetId(),
	})
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if getResp.GetTask().GetDescription() != "ship feature" {
		t.Fatalf("GetTask() description = %q", getResp.GetTask().GetDescription())
	}

	todayResp, err := taskClient.GetTodayTasks(taskAuthContext("jwt-user-1"), &taskpb.GetTasksRequest{})
	if err != nil {
		t.Fatalf("GetTodayTasks() error = %v", err)
	}
	if len(todayResp.GetTasks()) == 0 {
		t.Fatalf("GetTodayTasks() returned no tasks")
	}

	updateResp, err := taskClient.UpdateTaskStatus(taskAuthContext("jwt-user-1"), &taskpb.UpdateTaskStatusRequest{
		Id:     createResp.GetTask().GetId(),
		Status: taskpb.TaskStatus_TASK_STATUS_COMPLETED,
	})
	if err != nil {
		t.Fatalf("UpdateTaskStatus() error = %v", err)
	}
	if updateResp.GetTask().GetStatus() != taskpb.TaskStatus_TASK_STATUS_COMPLETED {
		t.Fatalf("UpdateTaskStatus() status = %v", updateResp.GetTask().GetStatus())
	}

	if _, err := taskClient.UpdateTaskStatus(taskAuthContext("jwt-user-2"), &taskpb.UpdateTaskStatusRequest{
		Id:     createResp.GetTask().GetId(),
		Status: taskpb.TaskStatus_TASK_STATUS_AT_WORK,
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("UpdateTaskStatus() foreign status = %v, want %v", status.Code(err), codes.PermissionDenied)
	}

	expiredSeed, err := repo.Create(context.Background(), domain.Task{
		UserID:      1,
		Description: "stale task",
		Status:      domain.CREATED,
		CreatedAt:   time.Now().Add(-15 * time.Minute),
		DueDate:     time.Now().Add(-5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("seed Create() error = %v", err)
	}

	if _, err := schedulerClient.ProcessRecentExpired(context.Background(), &emptypb.Empty{}); err != nil {
		t.Fatalf("ProcessRecentExpired() error = %v", err)
	}

	expiredTask, err := repo.GetByID(context.Background(), expiredSeed.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if expiredTask.Status != domain.EXPIRED {
		t.Fatalf("expired task status = %v, want %v", expiredTask.Status, domain.EXPIRED)
	}
}
