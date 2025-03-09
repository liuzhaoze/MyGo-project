package query

import (
	"context"

	"github.com/liuzhaoze/MyGo-project/common/decorator"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	domain "github.com/liuzhaoze/MyGo-project/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
}

func NewCheckIfItemsInStockHandler(stockRepo domain.Repository, logger *logrus.Entry, metricClient decorator.MetricsClient) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo},
		logger,
		metricClient,
	)
}

var stub = map[string]string{
	"1": "price_1QzWgnAe8D0pztRYOGHS1igj",
	"2": "price_1R0J9KAe8D0pztRYHqE5sbPn",
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, q CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, item := range q.Items {
		// TODO: priceID 从 stripe 或数据库获取
		priceID, ok := stub[item.ID]
		if !ok {
			priceID = stub["1"]
		}
		res = append(res, &orderpb.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}
