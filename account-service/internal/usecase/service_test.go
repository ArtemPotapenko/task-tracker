package usecase

import (
	"context"
	"errors"
	"testing"

	"task-tracker/account-service/internal/domain"
)

type stubUserRepo struct {
	getByEmailFunc            func(ctx context.Context, email string) (domain.User, error)
	createWithOutboxEventFunc func(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error)
	getByIDsFunc              func(ctx context.Context, ids []int64) ([]domain.User, error)
}

func (r *stubUserRepo) CreateWithOutboxEvent(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
	return r.createWithOutboxEventFunc(ctx, user, event)
}

func (r *stubUserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.getByEmailFunc(ctx, email)
}

func (r *stubUserRepo) GetByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	return r.getByIDsFunc(ctx, ids)
}

type stubHasher struct {
	hashFunc    func(password string) (string, error)
	compareFunc func(hash string, password string) bool
}

func (h stubHasher) Hash(password string) (string, error) {
	return h.hashFunc(password)
}

func (h stubHasher) Compare(hash string, password string) bool {
	return h.compareFunc(hash, password)
}

type stubTokens struct {
	newTokenFunc func(userID int64, email string) (string, error)
}

func (t stubTokens) NewToken(userID int64, email string) (string, error) {
	return t.newTokenFunc(userID, email)
}

func TestAuthServiceRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &stubUserRepo{
			getByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{}, domain.ErrNotFound
			},
			createWithOutboxEventFunc: func(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
				if user.PasswordHash != "hashed-secret" {
					t.Fatalf("unexpected password hash: %q", user.PasswordHash)
				}
				if event.Topic != "register" || event.Key != "user@example.com" || len(event.Payload) == 0 {
					t.Fatalf("unexpected outbox event: %+v", event)
				}
				return domain.User{ID: 42, Email: user.Email, PasswordHash: user.PasswordHash}, nil
			},
			getByIDsFunc: func(ctx context.Context, ids []int64) ([]domain.User, error) {
				return nil, nil
			},
		}

		svc := NewAuthService(
			repo,
			stubHasher{hashFunc: func(password string) (string, error) { return "hashed-secret", nil }},
			stubTokens{newTokenFunc: func(userID int64, email string) (string, error) {
				if userID != 42 || email != "user@example.com" {
					t.Fatalf("unexpected token payload: %d %s", userID, email)
				}
				return "jwt-token", nil
			}},
		)

		token, err := svc.Register(context.Background(), "user@example.com", "secret")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if token != "jwt-token" {
			t.Fatalf("Register() token = %q, want jwt-token", token)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		svc := NewAuthService(
			&stubUserRepo{
				getByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{ID: 1, Email: email}, nil
				},
				createWithOutboxEventFunc: func(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
					return domain.User{}, nil
				},
				getByIDsFunc: func(ctx context.Context, ids []int64) ([]domain.User, error) {
					return nil, nil
				},
			},
			stubHasher{hashFunc: func(password string) (string, error) { return "", nil }},
			stubTokens{newTokenFunc: func(userID int64, email string) (string, error) { return "", nil }},
		)

		_, err := svc.Register(context.Background(), "user@example.com", "secret")
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			t.Fatalf("Register() error = %v, want %v", err, domain.ErrUserAlreadyExists)
		}
	})
}

func TestAuthServiceLogin(t *testing.T) {
	repo := &stubUserRepo{
		getByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{ID: 7, Email: email, PasswordHash: "stored-hash"}, nil
		},
		createWithOutboxEventFunc: func(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
			return domain.User{}, nil
		},
		getByIDsFunc: func(ctx context.Context, ids []int64) ([]domain.User, error) {
			return nil, nil
		},
	}

	t.Run("success", func(t *testing.T) {
		svc := NewAuthService(
			repo,
			stubHasher{compareFunc: func(hash string, password string) bool {
				return hash == "stored-hash" && password == "secret"
			}},
			stubTokens{newTokenFunc: func(userID int64, email string) (string, error) {
				return "jwt-token", nil
			}},
		)

		token, err := svc.Login(context.Background(), "user@example.com", "secret")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if token != "jwt-token" {
			t.Fatalf("Login() token = %q, want jwt-token", token)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		svc := NewAuthService(
			repo,
			stubHasher{compareFunc: func(hash string, password string) bool { return false }},
			stubTokens{newTokenFunc: func(userID int64, email string) (string, error) { return "", nil }},
		)

		_, err := svc.Login(context.Background(), "user@example.com", "bad-password")
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("Login() error = %v, want %v", err, domain.ErrInvalidCredentials)
		}
	})
}

func TestAuthServiceGetUsersByIDs(t *testing.T) {
	repo := &stubUserRepo{
		getByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{}, nil
		},
		createWithOutboxEventFunc: func(ctx context.Context, user domain.User, event domain.OutboxEvent) (domain.User, error) {
			return domain.User{}, nil
		},
		getByIDsFunc: func(ctx context.Context, ids []int64) ([]domain.User, error) {
			return []domain.User{{ID: 1, Email: "one@example.com"}, {ID: 2, Email: "two@example.com"}}, nil
		},
	}

	svc := NewAuthService(
		repo,
		stubHasher{},
		stubTokens{},
	)

	users, err := svc.GetUsersByIDs(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("GetUsersByIDs() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("GetUsersByIDs() len = %d, want 2", len(users))
	}
}
