package service

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/integration"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/persistent"

	"github.com/liuzhaoze/MyGo-project/common/metrics"
	"github.com/liuzhaoze/MyGo-project/stock/adapters"
	"github.com/liuzhaoze/MyGo-project/stock/app"
	"github.com/liuzhaoze/MyGo-project/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	//stockRepo := adapters.NewMemoryStockRepository()
	db := persistent.NewMySQL()
	stockRepo := adapters.NewMySQLStockRepository(db)
	logger := logrus.NewEntry(logrus.StandardLogger())
	stripeAPI := integration.NewStripeAPI()
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
