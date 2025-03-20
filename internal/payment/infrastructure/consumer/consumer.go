package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/pkg/errors"
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
			c.handleMessage(ch, m, q)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	t := otel.Tracer("RabbitMQ")
	ctx, span := t.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("RabbitMQ.%s.consume", q.Name))
	defer span.End()

	logging.Infof(ctx, nil, "payment receive a message from %s, msg=%v", q.Name, string(msg.Body))
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

	o := &orderpb.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "fail to unmarshal order")
		return
	}

	if _, err = c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: o}); err != nil {
		err = errors.Wrap(err, "fail to create payment")
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error || error handle retry, messageID=%s, err=%v", msg.MessageId, err)
		}
	}

	span.AddEvent("payment.created")
}
