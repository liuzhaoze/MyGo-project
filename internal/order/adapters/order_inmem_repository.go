package adapters

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"strconv"
	"sync"
	"time"

	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
)

type OrderRepositoryMemory struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemoryOrderRepository() *OrderRepositoryMemory {
	s := []*domain.Order{{
		ID:          "fake-ID",
		CustomerID:  "fake-CustomerID",
		Status:      "fake-Status",
		PaymentLink: "fake-PaymentLink",
		Items:       nil,
	}}
	return &OrderRepositoryMemory{lock: &sync.RWMutex{}, store: s}
}

func (m *OrderRepositoryMemory) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	_, deferLog := logging.WhenRequest(ctx, "OrderRepositoryMemory.Create", map[string]any{"order": order})
	defer deferLog(created, &err)

	m.lock.Lock()
	defer m.lock.Unlock()

	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	m.store = append(m.store, newOrder)

	return newOrder, nil
}

func (m *OrderRepositoryMemory) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	_, deferLog := logging.WhenRequest(ctx, "OrderRepositoryMemory.Get", map[string]any{"id": id, "customerID": customerID})
	defer deferLog(got, &err)

	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			return o, nil
		}
	}
	return nil, &domain.NotFoundError{OrderID: id}
}

func (m *OrderRepositoryMemory) Update(ctx context.Context, o *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) (err error) {
	_, deferLog := logging.WhenRequest(ctx, "OrderRepositoryMemory.Update", map[string]any{"order": o})
	defer deferLog(nil, &err)

	m.lock.Lock()
	defer m.lock.Unlock()

	isFound := false

	for index, order := range m.store {
		if order.ID == o.ID && order.CustomerID == o.CustomerID {
			isFound = true
			updatedOrder, err := updateFn(ctx, o)
			if err != nil {
				return err
			}
			m.store[index] = updatedOrder
		}
	}
	if !isFound {
		return &domain.NotFoundError{OrderID: o.ID}
	}
	return nil
}
