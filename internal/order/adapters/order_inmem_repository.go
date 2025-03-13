package adapters

import (
	"context"
	"strconv"
	"sync"
	"time"

	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"github.com/sirupsen/logrus"
)

type OrderRepositoryMemory struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemoryOrderRepository() *OrderRepositoryMemory {
	s := make([]*domain.Order, 0)
	s = append(s, &domain.Order{
		ID:          "fake-ID",
		CustomerID:  "fake-CustomerID",
		Status:      "fake-Status",
		PaymentLink: "fake-PaymentLink",
		Items:       nil,
	})
	return &OrderRepositoryMemory{lock: &sync.RWMutex{}, store: s}
}

func (m *OrderRepositoryMemory) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
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

	logrus.WithFields(logrus.Fields{
		"input_order":        order,
		"store_after_create": m.store,
	}).Debug("memory_order_repo_create")

	return newOrder, nil
}

func (m *OrderRepositoryMemory) Get(ctx context.Context, id, customerID string) (*domain.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			logrus.Debugf("memory_order_repo_get || found || id=%s || customerID=%s || res=%+v", id, customerID, *o)
			return o, nil
		}
	}
	return nil, &domain.NotFoundError{OrderID: id}
}

func (m *OrderRepositoryMemory) Update(ctx context.Context, o *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) error {
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
