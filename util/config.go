package util

import (
	"time"

	"github.com/spf13/viper"
)

// All the app configurations readed by viper from env variables
type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:TOKEN_SYMMETRIC_KEY`
	AccessTokenDuration time.Duration `mapstructure:ACCESS_TOKEN_DURATION`
}

func LoadConfig(path string, devMode bool) (config Config, err error) {
	var name string

	if devMode {
		name = ".dev.env"
	} else {
		name = ".prod.env"
	}

	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
