package mq

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/broker"
	domain "github.com/liuzhaoze/MyGo-project/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQEventPublisher struct {
	Channel *amqp.Channel
}

func NewRabbitMQEventPublisher(channel *amqp.Channel) *RabbitMQEventPublisher {
	return &RabbitMQEventPublisher{Channel: channel}
}

func (r *RabbitMQEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	return broker.PublishEvent(ctx, broker.PublishEventRequest{
		Channel:  r.Channel,
		Routing:  broker.Direct,
		Queue:    event.Dest,
		Exchange: "",
		Body:     event.Data,
	})
}

func (r *RabbitMQEventPublisher) Broadcast(ctx context.Context, event domain.DomainEvent) error {
	return broker.PublishEvent(ctx, broker.PublishEventRequest{
		Channel:  r.Channel,
		Routing:  broker.FanOut,
		Queue:    event.Dest,
		Exchange: "",
		Body:     event.Data,
	})
}
