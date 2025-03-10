package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel"

	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/payment/app"
	"github.com/liuzhaoze/MyGo-project/payment/app/command"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(application app.Application) *Consumer {
	return &Consumer{
		app: application,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s, err=%v", q.Name, err)
	}

	forever := make(chan struct{})
	go func() {
		for m := range msgs {
			c.handleMessage(m, q)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue) {
	logrus.Infof("payment receive a message from %s, msg=%v", q.Name, string(msg.Body))

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("RabbitMQ")
	_, span := t.Start(ctx, fmt.Sprintf("RabbitMQ.%s.consume", q.Name))
	defer span.End()

	o := &orderpb.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Infof("fail to unmarshal order, err=%v", err)
		_ = msg.Nack(false, false)
		return
	}

	if _, err := c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: o}); err != nil {
		// TODO: retry
		logrus.Infof("fail to create payment, err=%v", err)
		_ = msg.Nack(false, false)
	}

	span.AddEvent("payment.created")

	_ = msg.Ack(false)
	logrus.Info("consume successfully")
}
