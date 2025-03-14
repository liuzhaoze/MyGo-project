package query

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/stock/entity"
	"github.com/liuzhaoze/MyGo-project/stock/infrastructure/integration"

	"github.com/liuzhaoze/MyGo-project/common/decorator"
	domain "github.com/liuzhaoze/MyGo-project/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*entity.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*entity.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
	stripeAPI *integration.StripeAPI
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	stripeAPI *integration.StripeAPI,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	if stripeAPI == nil {
		panic("stripeAPI is nil")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*entity.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricClient,
	)
}

// Deprecated
var stub = map[string]string{
	"1": "price_1QzWgnAe8D0pztRYOGHS1igj",
	"2": "price_1R0J9KAe8D0pztRYHqE5sbPn",
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, q CheckIfItemsInStock) ([]*entity.Item, error) {
	var res []*entity.Item
	for _, item := range q.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, item.ID)
		if err != nil || priceID == "" {
			return nil, err
		}
		res = append(res, &entity.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}

func getStubPriceID(id string) string {
	priceID, ok := stub[id]
	if !ok {
		priceID = stub["1"]
	}
	return priceID
}
