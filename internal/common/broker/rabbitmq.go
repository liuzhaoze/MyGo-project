package broker

import (
	"context"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"time"

	_ "github.com/liuzhaoze/MyGo-project/common/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	DLX                = "dlx"
	DLQ                = "dlq"
	amqpRetryHeaderKey = "x-retry-count"
)

var (
	maxRetryCount = viper.GetInt64("rabbitmq.max-retry")
)

func Connect(user, password, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	conn, err := amqp.Dial(address)
	if err != nil {
		logrus.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatal(err)
	}

	err = ch.ExchangeDeclare(EventOrderCreated, "direct", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	err = ch.ExchangeDeclare(EventOrderPaid, "fanout", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := createDLX(ch); err != nil {
		logrus.Fatal(err)
	}
	return ch, conn.Close
}

func createDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare("share_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DLX, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, "", DLX, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DLQ, true, false, false, false, nil)
	return err
}

func HandleRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) (err error) {
	fields, deferLog := logging.WhenRequest(ctx, "HandleRetry", map[string]any{
		"delivery":        d,
		"max_retry_count": maxRetryCount,
	})
	defer deferLog(nil, &err)

	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}
	retryCount, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCount = 0
	}
	retryCount++
	d.Headers[amqpRetryHeaderKey] = retryCount
	fields["retry_count"] = retryCount

	if retryCount > maxRetryCount {
		logrus.WithContext(ctx).Infof("moving message %s to dlq", d.MessageId)
		return doPublish(ctx, ch, "", DLQ, false, false, amqp.Publishing{
			Headers:      d.Headers,
			ContentType:  "application/json",
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
	}

	logrus.WithContext(ctx).Debugf("retrying message %s, count=%d", d.MessageId, retryCount)
	time.Sleep(time.Second * time.Duration(retryCount))
	return doPublish(ctx, ch, d.Exchange, d.RoutingKey, false, false, amqp.Publishing{
		Headers:      d.Headers,
		ContentType:  "application/json",
		Body:         d.Body,
		DeliveryMode: amqp.Persistent,
	})
}

type RabbitMQHeaderCarrier map[string]interface{}

func (r RabbitMQHeaderCarrier) Get(key string) string {
	value, ok := r[key]
	if !ok {
		return ""
	}
	return value.(string)
}

func (r RabbitMQHeaderCarrier) Set(key string, value string) {
	r[key] = value
}

func (r RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, len(r))
	i := 0
	for k := range r {
		keys[i] = k
		i++
	}
	return keys
}

func InjectRabbitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headers))
}
