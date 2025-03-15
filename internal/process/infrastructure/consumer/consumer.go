package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	var err error
	logrus.Infof("process receive a message from %s, msg=%v", q.Name, string(msg.Body))

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("RabbitMQ")
	mqCtx, span := t.Start(ctx, fmt.Sprintf("RabbitMQ.%s.consume", q.Name))

	defer func() {
		span.End()
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Infof("fail to unmarshal order, err=%v", err)
		return
	}

	if o.Status != "paid" {
		err = errors.New("order status is not paid, cannot process order")
	}
	process(o)

	if err := c.orderGRPC.UpdateOrder(mqCtx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      "ready",
		Items:       o.Items,
		PaymentLink: o.PaymentLink,
	}); err != nil {
		if err = broker.HandleRetry(mqCtx, ch, &msg); err != nil {
			logrus.Warnf("process || fail to process order, err=%v", err)
		}
		return
	}

	span.AddEvent(fmt.Sprintf("process processed: %v", o))
	logrus.Info("consume successfully")
}

func process(o *Order) {
	logrus.Printf("processing order: %s\n", o.ID)
	time.Sleep(5 * time.Second)
	logrus.Printf("order %s processed\n", o.ID)
}
