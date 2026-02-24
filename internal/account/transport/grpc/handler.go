package grpc

import (
	"context"
	"errors"
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountpb "task-tracker/gen"
	"task-tracker/internal/account/domain"
	"task-tracker/internal/account/usecase"
)

const minPasswordLength = 8

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type AuthHandler struct {
	accountpb.UnimplementedAuthServiceServer
	svc usecase.AuthService
}

func NewAuthHandler(svc usecase.AuthService) AuthHandler {
	return AuthHandler{svc: svc}
}

func (h AuthHandler) Register(ctx context.Context, req *accountpb.RegisterRequest) (*accountpb.AuthResponse, error) {
	if err := validateEmailPassword(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(req.GetRepeatPassword()) < minPasswordLength {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	jwt, err := h.svc.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &accountpb.AuthResponse{Jwt: jwt}, nil
}

func (h AuthHandler) Login(ctx context.Context, req *accountpb.LoginRequest) (*accountpb.AuthResponse, error) {
	if err := validateEmailPassword(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	jwt, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &accountpb.AuthResponse{Jwt: jwt}, nil
}

func validateEmailPassword(email string, password string) error {
	if !emailPattern.MatchString(email) {
		return errors.New("invalid email format")
	}
	if len(password) < minPasswordLength {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
