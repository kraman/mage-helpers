package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("protoc_version", "3.8.0")
	viper.SetDefault("cluster_name", "test-cluster")
	viper.SetDefault("nats_cluster_size", "1")
	viper.SetDefault("stan_cluster_size", "1")
	
	viper.SetConfigType("yaml")
	viper.SetConfigFile(".build.yml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()
	viper.SafeWriteConfig()
}