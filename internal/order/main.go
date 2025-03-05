package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/config"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/common/server"
	"github.com/liuzhaoze/MyGo-project/order/ports"
	"github.com/liuzhaoze/MyGo-project/order/service"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	go server.RunGRPCServer(serviceName, func(s *grpc.Server) {
		orderpb.RegisterOrderServiceServer(s, ports.NewGRPCServer(application))
	})

	print("HTTP server is running")
	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		ports.RegisterHandlersWithOptions(router, HTTPServer{app: application}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
