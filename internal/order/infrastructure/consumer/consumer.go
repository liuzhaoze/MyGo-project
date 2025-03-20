package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/order/app"
	"github.com/liuzhaoze/MyGo-project/order/app/command"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
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
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
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

	o := &domain.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "error unmarshal msg.body into domain.Order")
		return
	}
	_, err = c.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := order.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		},
	})
	if err != nil {
		logging.Errorf(ctx, nil, "error updating order || orderID=%s, err=%v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error || error handle retry, messageID=%s, err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.updated")
}
