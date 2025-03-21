package command

import (
	"context"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/common/decorator"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/liuzhaoze/MyGo-project/order/app/query"
	"github.com/liuzhaoze/MyGo-project/order/converter"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	"github.com/liuzhaoze/MyGo-project/order/domain/service"
	"github.com/liuzhaoze/MyGo-project/order/entity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo      domain.Repository
	stockGRPC      query.StockService
	eventPublisher domain.EventPublisher
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockService,
	eventPublisher domain.EventPublisher,
	logger *logrus.Logger,
	metricClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}
	if stockGRPC == nil {
		panic("stockGRPC is nil")
	}
	if eventPublisher == nil {
		panic("eventPublisher is nil")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{orderRepo: orderRepo, stockGRPC: stockGRPC, eventPublisher: eventPublisher},
		logger,
		metricClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreateOrderHandler", cmd, err)

	t := otel.Tracer("RabbitMQ")
	ctx, span := t.Start(ctx, fmt.Sprintf("RabbitMQ.%s.publish", broker.EventOrderCreated))
	defer span.End()

	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}

	pendingOrder, err := domain.NewPendingOrder(cmd.CustomerID, validItems)
	if err != nil {
		return nil, err
	}

	o, err := service.NewOrderDomainService(c.orderRepo, c.eventPublisher).CreateOrder(ctx, *pendingOrder)
	if err != nil {
		return nil, err
	}

	return &CreateOrderResult{OrderID: o.ID}, nil
}

func (c createOrderHandler) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least 1 item")
	}
	items = packItems(items)
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, converter.NewItemWithQuantityConverter().EntitiesToProtos(items))
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return converter.NewItemConverter().ProtosToEntities(resp.Items), nil
}

func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}
	var res []*entity.ItemWithQuantity
	for id, quantity := range merged {
		res = append(res, &entity.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}
	return res
}
