package server

import (
	"github.com/gin-gonic/gin"
	"github.com/liuzhaoze/MyGo-project/common/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func RunHTTPServer(serviceName string, wrapper func(router *gin.Engine)) {
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		panic("empty http address")
	}
	RunHTTPServerOnAddr(addr, wrapper)
}

func RunHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) {
	apiRouter := gin.New()
	setMiddlewares(apiRouter)
	wrapper(apiRouter)
	apiRouter.Group("/api")
	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}

func setMiddlewares(e *gin.Engine) {
	e.Use(middleware.StructuredLog(logrus.NewEntry(logrus.StandardLogger())))
	e.Use(gin.Recovery())
	e.Use(middleware.RequestLog(logrus.NewEntry(logrus.StandardLogger())))
	e.Use(otelgin.Middleware("default_server"))
}
