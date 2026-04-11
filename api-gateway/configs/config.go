package configs

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	JWTSecret          string
	AuthServiceURL     string
	PatientServiceURL  string
	LabServiceURL      string
	PharmacyServiceURL string
	BillingServiceURL  string
	ClinicalServiceURL string
	ServiceName        string
	Version            string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
