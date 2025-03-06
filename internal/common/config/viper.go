package config

import (
	"github.com/spf13/viper"
	"strings"
)

func NewViperConfig() error {
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../common/config") // 相对于使用该函数的包的路径
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	return viper.ReadInConfig()
}
