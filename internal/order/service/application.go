package service

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/metrics"
	"github.com/liuzhaoze/MyGo-project/order/adapters"
	"github.com/liuzhaoze/MyGo-project/order/app"
	"github.com/liuzhaoze/MyGo-project/order/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderRepo, logger, metricsClient),
		},
	}
}
