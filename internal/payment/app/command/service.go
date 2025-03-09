package command

import (
	"context"

	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, order *orderpb.Order) error
}
