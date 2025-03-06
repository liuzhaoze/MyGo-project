package main

import (
	"github.com/liuzhaoze/MyGo-project/common/broker"
	"github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/liuzhaoze/MyGo-project/common/server"
	"github.com/liuzhaoze/MyGo-project/payment/infrastructure/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("payment.service-name")
	serverType := viper.GetString("payment.service-protocol")

	paymentHandler := NewPaymentHandler()

	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()

	go consumer.NewConsumer().Listen(ch)

	switch serverType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported type")
	default:
		logrus.Panic("unsupported type")
	}
}
