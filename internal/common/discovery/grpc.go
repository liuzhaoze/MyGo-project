package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/liuzhaoze/MyGo-project/common/discovery/consul"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func RegisterToConsul(ctx context.Context, serviceName string) (func() error, error) {
	registry, err := consul.New(viper.GetString("consul.address"))
	if err != nil {
		return func() error { return nil }, err
	}

	instanceID := GenerateInstanceID(serviceName)
	grpcAddr := viper.Sub(serviceName).GetString("grpc-addr")
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		return func() error { return nil }, err
	}
	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logrus.Panicf("no heartbeat from %s to registry, err=%v", serviceName, err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
		"addr":        grpcAddr,
	}).Info("register to consul")

	return func() error {
		return registry.Deregister(ctx, instanceID, serviceName)
	}, nil
}

func GetServiceAddress(ctx context.Context, serviceName string) (string, error) {
	registry, err := consul.New(viper.GetString("consul.address"))
	if err != nil {
		return "", err
	}

	addresses, err := registry.Discover(ctx, serviceName)
	if err != nil {
		return "", err
	}

	if len(addresses) == 0 {
		return "", fmt.Errorf("got empty %s address from consul", serviceName)
	}
	i := rand.Intn(len(addresses))
	logrus.Infof("discovered %d instances of %s, addrs=%v", len(addresses), serviceName, addresses)

	return addresses[i], nil
}
