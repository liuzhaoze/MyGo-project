package main

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/common/genproto/stockpb"
	"github.com/liuzhaoze/MyGo-project/common/server"
	"github.com/liuzhaoze/MyGo-project/stock/ports"
	"github.com/liuzhaoze/MyGo-project/stock/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.service-protocol")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application := service.NewApplication(ctx)
	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(s *grpc.Server) {
			stockpb.RegisterStockServiceServer(s, ports.NewGRPCServer(application))
		})
	case "http":
		// 暂时不用
	default:
		panic("Unexpected server protocol")
	}
}
