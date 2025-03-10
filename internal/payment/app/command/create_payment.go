package command

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/decorator"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/payment/domain"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type CreatePayment struct {
	Order *orderpb.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	// NOTE: 这里作者删了，但是我的正常
	t := otel.Tracer("payment")
	ctx, span := t.Start(ctx, "create_payment")
	defer span.End()

	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}

	logrus.Infof("create payment link for order: %s success, payment link: %s", cmd.Order.ID, link)
	newOrder := &orderpb.Order{
		ID:          cmd.Order.ID,
		CustomerID:  cmd.Order.CustomerID,
		Status:      "waiting_for_payment",
		Items:       cmd.Order.Items,
		PaymentLink: link,
	}

	err = c.orderGRPC.UpdateOrder(ctx, newOrder)
	return link, err
}

func NewCreatePaymentHandler(processor domain.Processor, orderGRPC OrderService, logger *logrus.Entry, metricClient decorator.MetricsClient) CreatePaymentHandler {
	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{
			processor: processor,
			orderGRPC: orderGRPC,
		}, logger, metricClient)
}
