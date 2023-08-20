package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string, devMode bool) (config Config, err error) {
	var name string

	if devMode {
		name = ".dev.env"
	} else {
		name = ".prod.env"
	}

	viper.AddConfigPath(path)
	viper.SetConfigFile(name)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
