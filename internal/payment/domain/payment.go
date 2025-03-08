package domain

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
)

type Processor interface {
	CreatePaymentLink(context.Context, *orderpb.Order) (string, error)
}
