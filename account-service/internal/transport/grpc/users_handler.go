package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"task-tracker/account-service/internal/domain"
	"task-tracker/account-service/internal/usecase"
	accountpb "task-tracker/proto-lib/gen/private/account"
)

type UsersHandler struct {
	accountpb.UnimplementedUsersServiceServer
	svc *usecase.AuthService
}

func NewUsersHandler(svc *usecase.AuthService) UsersHandler {
	return UsersHandler{svc: svc}
}

func (h UsersHandler) GetUsersByIDs(ctx context.Context, req *accountpb.GetUsersByIDsRequest) (*accountpb.UsersResponse, error) {
	ids := req.GetIds()

	users, err := h.svc.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, mapUsersError(err)
	}

	resp := &accountpb.UsersResponse{Users: make([]*accountpb.User, 0, len(users))}
	for _, user := range users {
		resp.Users = append(resp.Users, &accountpb.User{Id: user.ID, Email: user.Email})
	}
	return resp, nil
}

func mapUsersError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
