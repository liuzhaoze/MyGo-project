package main

import (
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/common/server"
	"github.com/liuzhaoze/MyGo-project/order/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("order.service-name")

	go server.RunGRPCServer(serviceName, func(s *grpc.Server) {
		orderpb.RegisterOrderServiceServer(s, ports.NewGRPCServer())
	})

	print("HTTP server is running")
	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		ports.RegisterHandlersWithOptions(router, HTTPServer{}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
