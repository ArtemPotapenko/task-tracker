package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accountpb "task-tracker/gen/account"
	"task-tracker/internal/account/domain"
	"task-tracker/internal/account/usecase"
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
	if len(ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ids are required")
	}

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
