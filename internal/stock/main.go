package main

import (
	"github.com/liuzhaoze/MyGo-project/common/genproto/stockpb"
	"github.com/liuzhaoze/MyGo-project/common/server"
	"github.com/liuzhaoze/MyGo-project/stock/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.service-protocol")

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(s *grpc.Server) {
			stockpb.RegisterStockServiceServer(s, ports.NewGRPCServer())
		})
	case "http":
		// 暂时不用
	default:
		panic("Unexpected server protocol")
	}
}
