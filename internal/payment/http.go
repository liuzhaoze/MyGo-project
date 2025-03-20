package main

import (
	"encoding/json"
	"fmt"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/payment/domain"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: ch}
}

func (h *PaymentHandler) RegisterRoutes(e *gin.Engine) {
	e.POST("/api/webhook", h.handleWebhook)
}

func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	logrus.WithContext(c.Request.Context()).Info("Got webhook from stripe")
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(c.Request.Context(), nil, "handleWebhook err=%v", err)
		} else {
			logging.Infof(c.Request.Context(), nil, "%s", "handleWebhook success")
		}
	}()
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		err = errors.Wrap(err, "error reading request body")
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("stripe-endpoint-secret"))

	if err != nil {
		err = errors.Wrap(err, "error verifying webhook signature")
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err = json.Unmarshal(event.Data.Raw, &session); err != nil {
			err = errors.Wrap(err, "error unmarshalling event.data.raw into session")
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			var items []*orderpb.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			t := otel.Tracer("RabbitMQ")
			ctx, span := t.Start(c.Request.Context(), fmt.Sprintf("RabbitMQ.%s.publish", broker.EventOrderPaid))
			defer span.End()

			_ = broker.PublishEvent(ctx, broker.PublishEventRequest{
				Channel:  h.channel,
				Routing:  broker.FanOut,
				Queue:    "",
				Exchange: broker.EventOrderPaid,
				Body: &domain.Order{
					ID:          session.Metadata["orderID"],
					CustomerID:  session.Metadata["customerID"],
					Status:      string(stripe.CheckoutSessionPaymentStatusPaid),
					PaymentLink: session.Metadata["paymentLink"],
					Items:       items,
				},
			})
		}
	}

	c.JSON(http.StatusOK, nil)

}
