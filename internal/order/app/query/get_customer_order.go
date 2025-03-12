package query

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/tracing"

	"github.com/liuzhaoze/MyGo-project/common/decorator"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"github.com/sirupsen/logrus"
)

type GetCustomerOrder struct {
	CustomerID string
	OrderID    string
}

type GetCustomerOrderHandler decorator.QueryHandler[GetCustomerOrder, *domain.Order]

type getCustomerOrderHandler struct {
	orderRepo domain.Repository
}

func NewGetCustomerOrderHandler(orderRepo domain.Repository, logger *logrus.Entry, metricsClient decorator.MetricsClient) GetCustomerOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}
	return decorator.ApplyQueryDecorators[GetCustomerOrder, *domain.Order](
		getCustomerOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (g getCustomerOrderHandler) Handle(ctx context.Context, q GetCustomerOrder) (*domain.Order, error) {
	_, span := tracing.Start(ctx, "getCustomerOrderHandler.Handle")
	o, err := g.orderRepo.Get(ctx, q.OrderID, q.CustomerID)
	if err != nil {
		return nil, err
	}
	span.AddEvent("get_success")
	span.End()
	return o, nil
}
