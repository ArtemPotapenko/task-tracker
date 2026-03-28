package app

import (
	"context"
	"errors"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"task-tracker/account-service/internal/domain"
	transportgrpc "task-tracker/account-service/internal/transport/grpc"
	"task-tracker/account-service/internal/usecase"
	accountprivatepb "task-tracker/proto-lib/gen/private/account"
	accountpb "task-tracker/proto-lib/gen/public/account"
)

const accountBufSize = 1024 * 1024

type memoryUserRepo struct {
	nextID int64
	users  map[string]domain.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{
		nextID: 1,
		users:  make(map[string]domain.User),
	}
}

func (r *memoryUserRepo) CreateWithOutboxEvent(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
	if _, ok := r.users[user.Email]; ok {
		return domain.User{}, domain.ErrUserAlreadyExists
	}
	user.ID = r.nextID
	r.nextID++
	r.users[user.Email] = user
	return user, nil
}

func (r *memoryUserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, ok := r.users[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (r *memoryUserRepo) GetByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	idSet := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	result := make([]domain.User, 0, len(ids))
	for _, user := range r.users {
		if _, ok := idSet[user.ID]; ok {
			result = append(result, user)
		}
	}
	return result, nil
}

type accountTestHasher struct{}

func (accountTestHasher) Hash(password string) (string, error) { return "hash:" + password, nil }
func (accountTestHasher) Compare(hash string, password string) bool {
	return hash == "hash:"+password
}

type accountTestTokens struct{}

func (accountTestTokens) NewToken(userID int64, email string) (string, error) {
	return email, nil
}

func TestAccountServiceE2E(t *testing.T) {
	repo := newMemoryUserRepo()
	svc := usecase.NewAuthService(repo, accountTestHasher{}, accountTestTokens{})

	lis := bufconn.Listen(accountBufSize)
	server := grpc.NewServer()
	accountpb.RegisterAuthServiceServer(server, transportgrpc.NewAuthHandler(svc))
	accountprivatepb.RegisterUsersServiceServer(server, transportgrpc.NewUsersHandler(svc))

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

	authClient := accountpb.NewAuthServiceClient(conn)
	usersClient := accountprivatepb.NewUsersServiceClient(conn)

	registerResp, err := authClient.Register(context.Background(), &accountpb.RegisterRequest{
		Email:          "user@example.com",
		Password:       "password123",
		RepeatPassword: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registerResp.GetJwt() != "user@example.com" {
		t.Fatalf("Register() jwt = %q", registerResp.GetJwt())
	}

	loginResp, err := authClient.Login(context.Background(), &accountpb.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loginResp.GetJwt() != "user@example.com" {
		t.Fatalf("Login() jwt = %q", loginResp.GetJwt())
	}

	usersResp, err := usersClient.GetUsersByIDs(context.Background(), &accountprivatepb.GetUsersByIDsRequest{
		Ids: []int64{1},
	})
	if err != nil {
		t.Fatalf("GetUsersByIDs() error = %v", err)
	}
	if len(usersResp.GetUsers()) != 1 || usersResp.GetUsers()[0].GetEmail() != "user@example.com" {
		t.Fatalf("GetUsersByIDs() = %+v", usersResp.GetUsers())
	}
}
