package client

import (
	"context"
	"errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"net"
	"time"

	"github.com/liuzhaoze/MyGo-project/common/discovery"
	"github.com/liuzhaoze/MyGo-project/common/genproto/orderpb"
	"github.com/liuzhaoze/MyGo-project/common/genproto/stockpb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewStockGRPCClient(ctx context.Context) (client stockpb.StockServiceClient, close func() error, err error) {
	if !waitForStockGRPCClient(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("stock grpc not available")
	}
	grpcAddr, err := discovery.GetServiceAddress(ctx, viper.GetString("stock.service-name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	if grpcAddr == "" {
		logrus.Warn("no stock grpc service address found")
	}
	opts := grpcDialOption(grpcAddr)
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return stockpb.NewStockServiceClient(conn), conn.Close, nil
}

func NewOrderGRPCClient(ctx context.Context) (client orderpb.OrderServiceClient, close func() error, err error) {
	if !waitForOrderGRPCClient(viper.GetDuration("dial-grpc-timeout") * time.Second) {
		return nil, nil, errors.New("order grpc not available")
	}
	grpcAddr, err := discovery.GetServiceAddress(ctx, viper.GetString("order.service-name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	if grpcAddr == "" {
		logrus.Warn("no order grpc service address found")
	}
	opts := grpcDialOption(grpcAddr)
	conn, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return orderpb.NewOrderServiceClient(conn), conn.Close, nil
}

func grpcDialOption(_ string) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

func waitForOrderGRPCClient(timeout time.Duration) bool {
	logrus.Infof("waiting for order grpc client connection, timeout=%vs", timeout.Seconds())
	return waitFor(viper.GetString("order.grpc-addr"), timeout)
}

func waitForStockGRPCClient(timeout time.Duration) bool {
	logrus.Infof("waiting for stock grpc client connection, timeout=%vs", timeout.Seconds())
	return waitFor(viper.GetString("stock.grpc-addr"), timeout)
}

func waitFor(addr string, timeout time.Duration) bool {
	portAvailable := make(chan struct{})
	timeoutCh := time.After(timeout)

	go func() {
		for {
			select {
			case <-timeoutCh:
				return
			default:
			}
			_, err := net.Dial("tcp", addr)
			if err == nil {
				close(portAvailable)
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case <-portAvailable:
		return true
	case <-timeoutCh:
		return false
	}
}
