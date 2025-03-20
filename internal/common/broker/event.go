package broker

import (
	"context"
	"encoding/json"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
)

type RoutingType string

const (
	FanOut RoutingType = "fan-out"
	Direct RoutingType = "direct"
)

type PublishEventRequest struct {
	Channel  *amqp.Channel
	Routing  RoutingType
	Queue    string
	Exchange string
	Body     any
}

func PublishEvent(ctx context.Context, p PublishEventRequest) (err error) {
	_, deferLog := logging.WhenEventPublish(ctx, p)
	defer deferLog(nil, &err)

	if err = checkParam(p); err != nil {
		return err
	}

	switch p.Routing {
	default:
		logrus.WithContext(ctx).Panicf("unsupported routing type: %s", string(p.Routing))
	case FanOut:
		return fanOut(ctx, p)
	case Direct:
		return directQueue(ctx, p)
	}
	return nil
}

func checkParam(p PublishEventRequest) error {
	if p.Channel == nil {
		return errors.New("channel is nil")
	}
	return nil
}

func directQueue(ctx context.Context, p PublishEventRequest) (err error) {
	_, err = p.Channel.QueueDeclare(p.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}

	return doPublish(ctx, p.Channel, p.Exchange, p.Queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func fanOut(ctx context.Context, p PublishEventRequest) (err error) {
	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}

	return doPublish(ctx, p.Channel, p.Exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func doPublish(ctx context.Context, ch *amqp.Channel, exchange string, key string, mandatory bool, immediate bool, message amqp.Publishing) error {
	if err := ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, message); err != nil {
		logging.Warnf(ctx, nil, "_publish_event_failed || exchange=%s, key=%s, msg=%v", exchange, key, message)
		return errors.Wrap(err, "publish event error")
	}
	return nil
}
