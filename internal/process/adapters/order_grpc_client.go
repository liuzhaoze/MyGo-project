package adapters

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
)

type OrderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOrderGRPC(client orderpb.OrderServiceClient) *OrderGRPC {
	return &OrderGRPC{client: client}
}

func (o *OrderGRPC) UpdateOrder(ctx context.Context, request *orderpb.Order) error {
	_, err := o.client.UpdateOrder(ctx, request)
	return err
}
