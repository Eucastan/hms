package configs

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DSN       string
	JWTSecret string
	HTTPPort  string
	GRPCPort  string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWTSecret too short; must be atleast 32 chars HS256")
	}

	return &cfg, nil
}
