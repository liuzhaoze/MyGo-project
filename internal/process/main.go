package main

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/broker"
	grpcClient "github.com/liuzhaoze/MyGo-project/common/client"
	"github.com/liuzhaoze/MyGo-project/common/tracing"
	"github.com/liuzhaoze/MyGo-project/process/adapters"
	"github.com/liuzhaoze/MyGo-project/process/infrastructure/consumer"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/common/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("process.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	orderClient, closeFunc, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer closeFunc()

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

	orderGRPC := adapters.NewOrderGRPC(orderClient)
	go consumer.NewConsumer(orderGRPC).Listen(ch)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		logrus.Infoln("receive signal, exiting...")
		os.Exit(0)
	}()

	logrus.Println("Press Ctrl+C to terminate")

	select {}
}
