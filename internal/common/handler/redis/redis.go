package redis

import (
	"github.com/liuzhaoze/MyGo-project/common/handler/factory"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"time"
)

const (
	configName    = "redis"
	localSupplier = "local"
)

var (
	singleton = factory.NewSingleton(supplier)
)

func Init() {
	config := viper.GetStringMap(configName)
	for supplyName := range config {
		Client(supplyName)
	}
}

func LocalClient() *redis.Client {
	return Client(localSupplier)
}

func Client(name string) *redis.Client {
	return singleton.Get(name).(*redis.Client)
}

func supplier(key string) any {
	configKey := configName + "." + key
	type Section struct {
		IP           string        `mapstructure:"ip"`
		Port         string        `mapstructure:"port"`
		PoolSize     int           `mapstructure:"pool_size"`
		MaxConn      int           `mapstructure:"max_conn"`
		ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}
	var s Section
	if err := viper.UnmarshalKey(configKey, &s); err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            s.IP + ":" + s.Port,
		PoolSize:        s.PoolSize,
		MaxActiveConns:  s.MaxConn,
		ConnMaxLifetime: s.ConnTimeout * time.Millisecond,
		ReadTimeout:     s.ReadTimeout * time.Millisecond,
		WriteTimeout:    s.WriteTimeout * time.Millisecond,
	})
}
