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

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	var res []*entity.Item
	for _, item := range query.Items {
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
	// TODO: 扣库存
	return res, nil
}

func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, query []*entity.ItemWithQuantity) error {
	var ids []string
	for _, i := range query {
		ids = append(ids, i.ID)
	}

	records, err := h.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}

	var idQuantityMap = make(map[string]int32)
	for _, r := range records {
		idQuantityMap[r.ID] += r.Quantity
	}
	var (
		ok       = true
		failedOn []struct {
			ID   string
			Want int32
			Have int32
		}
	)
	for _, i := range query {
		if i.Quantity > idQuantityMap[i.ID] {
			ok = false
			failedOn = append(failedOn, struct {
				ID   string
				Want int32
				Have int32
			}{ID: i.ID, Want: i.Quantity, Have: idQuantityMap[i.ID]})
		}
	}
	if ok {
		return nil
	}
	return domain.ExceedStockError{FailedOn: failedOn}
}

func getStubPriceID(id string) string {
	priceID, ok := stub[id]
	if !ok {
		priceID = stub["1"]
	}
	return priceID
}
