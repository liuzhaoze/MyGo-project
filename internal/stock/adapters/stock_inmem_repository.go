package adapters

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	domain "github.com/liuzhaoze/MyGo-project/stock/domain/stock"
	"sync"
)

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*orderpb.Item
}

var stub = map[string]*orderpb.Item{
	"item_id": {
		ID:       "foo_item",
		Name:     "stub_item",
		Quantity: 1000,
		PriceID:  "stub_item_price_id",
	},
}

func NewMemoryStockRepository() *MemoryStockRepository {
	return &MemoryStockRepository{lock: &sync.RWMutex{}, store: stub}
}

func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var res []*orderpb.Item
	var missing []string
	for _, id := range ids {
		if item, exist := m.store[id]; exist {
			res = append(res, item)
		} else {
			missing = append(missing, id)
		}
	}
	if len(res) > 0 {
		return res, nil
	}
	return res, &domain.NotFoundError{Missing: missing}
}
