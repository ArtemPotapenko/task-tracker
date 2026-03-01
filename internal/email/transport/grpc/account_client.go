package grpc

import (
	"context"
	"task-tracker/internal/email/transport/kafka"

	accountpb "task-tracker/gen/account"
)

type AccountClientAdapter struct {
	client accountpb.UsersServiceClient
}

func NewAccountClientAdapter(client accountpb.UsersServiceClient) AccountClientAdapter {
	return AccountClientAdapter{client: client}
}

func (a AccountClientAdapter) GetUsersByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	resp, err := a.client.GetUsersByIDs(ctx, &accountpb.GetUsersByIDsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	result := make(map[int64]string, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		result[user.GetId()] = user.GetEmail()
	}
	return result, nil
}

var _ kafka.UsersClient = AccountClientAdapter{}
