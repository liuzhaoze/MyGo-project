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
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	orderClient, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}

	orderGRPC := adapters.NewOrderGRPC(orderClient)
	memProcessor := processor.NewInMemProcessor()
	return newApplication(ctx, orderGRPC, memProcessor), func() {
		_ = closeOrderClient()
	}
}

func newApplication(ctx context.Context, orderGRPC command.OrderService, processor domain.Processor) app.Application {
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, orderGRPC, logger, metricClient),
		},
	}
}
