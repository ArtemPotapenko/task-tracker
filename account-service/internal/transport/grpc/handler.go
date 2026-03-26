package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"task-tracker/account-service/internal/domain"
	"task-tracker/account-service/internal/usecase"
	accountpb "task-tracker/proto-lib/gen/public/account"
)

type AuthHandler struct {
	accountpb.UnimplementedAuthServiceServer
	svc *usecase.AuthService
}

func NewAuthHandler(svc *usecase.AuthService) AuthHandler {
	return AuthHandler{svc: svc}
}

func (h AuthHandler) Register(ctx context.Context, req *accountpb.RegisterRequest) (*accountpb.AuthResponse, error) {
	if req.GetPassword() != req.GetRepeatPassword() {
		return nil, status.Error(codes.InvalidArgument, "passwords do not match")
	}

	jwt, err := h.svc.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &accountpb.AuthResponse{Jwt: jwt}, nil
}

func (h AuthHandler) Login(ctx context.Context, req *accountpb.LoginRequest) (*accountpb.AuthResponse, error) {
	jwt, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &accountpb.AuthResponse{Jwt: jwt}, nil
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
