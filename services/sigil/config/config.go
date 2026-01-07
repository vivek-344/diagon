package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading .env file: %v", err)
	}

	slog.Debug("config loaded", "settings", viper.AllSettings())
	viper.AutomaticEnv()

	cfg := &Config{
		DatabaseURL: viper.GetString("DATABASE_URL"),
		Port:        viper.GetString("PORT"),
	}

	// Default port if not set
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	return nil
}
