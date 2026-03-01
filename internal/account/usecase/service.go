package usecase

import (
	"context"
	"errors"

	"task-tracker/internal/account/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) bool
}

type TokenManager interface {
	NewToken(userID int64, email string) (string, error)
}

type AuthService struct {
	repo      domain.UserRepository
	hasher    PasswordHasher
	tokens    TokenManager
	publisher RegistrationPublisher
}

type RegistrationPublisher interface {
	PublishRegistered(ctx context.Context, email string) error
}

func NewAuthService(repo domain.UserRepository, hasher PasswordHasher, tokens TokenManager, publisher RegistrationPublisher) *AuthService {
	return &AuthService{repo: repo, hasher: hasher, tokens: tokens, publisher: publisher}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (string, error) {
	_, err := s.repo.GetByEmail(ctx, email)
	switch {
	case err == nil:
		return "", domain.ErrUserAlreadyExists
	case errors.Is(err, domain.ErrNotFound):
		// continue
	case err != nil:
		return "", err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return "", err
	}

	user, err := s.repo.Create(ctx, domain.User{Email: email, PasswordHash: hash})
	if err != nil {
		return "", err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishRegistered(ctx, user.Email); err != nil {
			return "", err
		}
	}

	return s.tokens.NewToken(user.ID, user.Email)
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if !s.hasher.Compare(user.PasswordHash, password) {
		return "", domain.ErrInvalidCredentials
	}

	return s.tokens.NewToken(user.ID, user.Email)
}

func (s *AuthService) GetUsersByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}
	return s.repo.GetByIDs(ctx, ids)
}
