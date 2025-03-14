package converter

import "sync"

var (
	orderConverter *OrderConverter
	orderOnce      sync.Once
)

var (
	itemConverter *ItemConverter
	itemOnce      sync.Once
)

var (
	itemWithQuantityConverter *ItemWithQuantityConverter
	itemWithQuantityOnce      sync.Once
)

func NewOrderConverter() *OrderConverter {
	orderOnce.Do(func() {
		orderConverter = new(OrderConverter)
	})
	return orderConverter
}

func NewItemConverter() *ItemConverter {
	itemOnce.Do(func() {
		itemConverter = new(ItemConverter)
	})
	return itemConverter
}

func NewItemWithQuantityConverter() *ItemWithQuantityConverter {
	itemWithQuantityOnce.Do(func() {
		itemWithQuantityConverter = new(ItemWithQuantityConverter)
	})
	return itemWithQuantityConverter
}
