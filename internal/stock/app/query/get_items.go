package query

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/stock/entity"

	"github.com/liuzhaoze/MyGo-project/common/decorator"
	domain "github.com/liuzhaoze/MyGo-project/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type GetItems struct {
	ItemIDs []string
}

type GetItemsHandler decorator.QueryHandler[GetItems, []*entity.Item]

type getItemsHandler struct {
	stockRepo domain.Repository
}

func NewGetItemsHandler(stockRepo domain.Repository, logger *logrus.Logger, metricClient decorator.MetricsClient) GetItemsHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	return decorator.ApplyQueryDecorators[GetItems, []*entity.Item](
		getItemsHandler{stockRepo: stockRepo},
		logger,
		metricClient,
	)
}

func (g getItemsHandler) Handle(ctx context.Context, query GetItems) ([]*entity.Item, error) {
	items, err := g.stockRepo.GetItems(ctx, query.ItemIDs)
	if err != nil {
		return nil, err
	}
	return items, nil
}
