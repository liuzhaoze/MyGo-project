package service

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/metrics"
	"github.com/liuzhaoze/MyGo-project/stock/adapters"
	"github.com/liuzhaoze/MyGo-project/stock/app"
	"github.com/liuzhaoze/MyGo-project/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	stockRepo := adapters.NewMemoryStockRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
