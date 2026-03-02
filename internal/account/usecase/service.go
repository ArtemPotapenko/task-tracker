package usecase

import (
	"context"
	"errors"

	"task-tracker/internal/account/domain"
	"task-tracker/pkg/logger"
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
		logger.Log.Infof("auth register: user already exists email=%s", email)
		return "", domain.ErrUserAlreadyExists
	case errors.Is(err, domain.ErrNotFound):
		// continue
	case err != nil:
		logger.Log.Infof("auth register: get by email error email=%s err=%v", email, err)
		return "", err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		logger.Log.Infof("auth register: hash error email=%s err=%v", email, err)
		return "", err
	}

	user, err := s.repo.Create(ctx, domain.User{Email: email, PasswordHash: hash})
	if err != nil {
		logger.Log.Infof("auth register: create user error email=%s err=%v", email, err)
		return "", err
	}
	logger.Log.Infof("auth register: user created id=%d email=%s", user.ID, user.Email)

	if s.publisher != nil {
		if err := s.publisher.PublishRegistered(ctx, user.Email); err != nil {
			logger.Log.Infof("auth register: publish registered error email=%s err=%v", user.Email, err)
			return "", err
		}
	}

	token, err := s.tokens.NewToken(user.ID, user.Email)
	if err != nil {
		logger.Log.Infof("auth register: new token error id=%d email=%s err=%v", user.ID, user.Email, err)
		return "", err
	}
	logger.Log.Infof("auth register: success id=%d email=%s", user.ID, user.Email)
	return token, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Log.Infof("auth login: user not found email=%s", email)
			return "", domain.ErrInvalidCredentials
		}
		logger.Log.Infof("auth login: get by email error email=%s err=%v", email, err)
		return "", err
	}

	if !s.hasher.Compare(user.PasswordHash, password) {
		logger.Log.Infof("auth login: invalid password email=%s", email)
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.NewToken(user.ID, user.Email)
	if err != nil {
		logger.Log.Infof("auth login: new token error id=%d email=%s err=%v", user.ID, user.Email, err)
		return "", err
	}
	logger.Log.Infof("auth login: success id=%d email=%s", user.ID, user.Email)
	return token, nil
}

func (s *AuthService) GetUsersByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	if len(ids) == 0 {
		logger.Log.Infof("auth get users: empty ids")
		return []domain.User{}, nil
	}
	users, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		logger.Log.Infof("auth get users: repo error err=%v", err)
		return nil, err
	}
	logger.Log.Infof("auth get users: result count=%d", len(users))
	return users, nil
}
