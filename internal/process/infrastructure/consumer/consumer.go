package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/consts"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"time"

	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, request *orderpb.Order) error
}

type Consumer struct {
	orderGRPC OrderService
}

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{
		orderGRPC: orderGRPC,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
		logrus.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s, err=%v", q.Name, err)
	}

	forever := make(chan struct{})
	go func() {
		for m := range msgs {
			c.handleMessage(ch, m, q)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	t := otel.Tracer("RabbitMQ")
	ctx, span := t.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("RabbitMQ.%s.consume", q.Name))
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume message failed || from=%s || msg=%+v || err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume message success")
			_ = msg.Ack(false)
		}
	}()

	o := &Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "fail to unmarshal order")
		return
	}

	if o.Status != "paid" {
		err = errors.New("order status is not paid, cannot process order")
	}
	process(ctx, o)

	if err = c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      consts.OrderStatusReady,
		Items:       o.Items,
		PaymentLink: o.PaymentLink,
	}); err != nil {
		logging.Errorf(ctx, nil, "fail to update order, orderID=%s, err=%v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrap(err, "process || fail to process order")
		}
		return
	}

	span.AddEvent(fmt.Sprintf("process processed: %v", o))
	logrus.Info("consume successfully")
}

func process(ctx context.Context, o *Order) {
	logrus.WithContext(ctx).Printf("processing order: %s\n", o.ID)
	time.Sleep(5 * time.Second)
	logrus.WithContext(ctx).Printf("order %s processed\n", o.ID)
}
