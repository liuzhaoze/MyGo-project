package service

import (
	"context"

	grpcClient "github.com/liuzhaoze/MyGo-project/common/client"
	"github.com/liuzhaoze/MyGo-project/common/metrics"
	"github.com/liuzhaoze/MyGo-project/payment/adapters"
	"github.com/liuzhaoze/MyGo-project/payment/app"
	"github.com/liuzhaoze/MyGo-project/payment/app/command"
	"github.com/liuzhaoze/MyGo-project/payment/domain"
	"github.com/liuzhaoze/MyGo-project/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	orderClient, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}

	orderGRPC := adapters.NewOrderGRPC(orderClient)
	//memProcessor := processor.NewInMemProcessor()
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))
	return newApplication(ctx, orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

func newApplication(_ context.Context, orderGRPC command.OrderService, processor domain.Processor) app.Application {
	metricClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, orderGRPC, logrus.StandardLogger(), metricClient),
		},
	}
}
