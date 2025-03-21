package service

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/broker"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"github.com/pkg/errors"
)

type OrderDomainService struct {
	Repo           domain.Repository
	EventPublisher domain.EventPublisher
}

func NewOrderDomainService(repo domain.Repository, eventPublisher domain.EventPublisher) *OrderDomainService {
	return &OrderDomainService{Repo: repo, EventPublisher: eventPublisher}
}

func (s *OrderDomainService) CreateOrder(ctx context.Context, order domain.Order) (*domain.Order, error) {
	root := domain.NewAggregateRoot(
		domain.Identity{CustomerID: order.CustomerID, OrderID: order.ID},
		&order,
	)
	o, err := s.Repo.Create(ctx, root.OrderEntity)
	if err != nil {
		return nil, err
	}
	if err = s.EventPublisher.Publish(ctx, domain.DomainEvent{
		Dest: broker.EventOrderCreated,
		Data: o,
	}); err != nil {
		return nil, errors.Wrapf(err, "publish event error, q.Name=%s", broker.EventOrderCreated)
	}

	return o, nil
}
