package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/vivek-344/diagon/services/auth/internal/config"
)

const (
	DefaultGRPCPort = 50051
	DefaultDBPort   = 5432
)

func LoadConfig() (*config.Config, error) {
	v := viper.New()

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = ".env"
	}

	v.SetConfigFile(configFile)
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("bootstrap: read config: %w", err)
		}
	}

	setDefaults(v)

	if err := bindEnv(v); err != nil {
		return nil, err
	}

	var cfg config.Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("bootstrap: unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("bootstrap: validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", config.EnvDevelopment)
	v.SetDefault("APP_VERSION", "dev")
	v.SetDefault("GRPC_PORT", DefaultGRPCPort)
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	v.SetDefault("DB_PORT", DefaultDBPort)
	v.SetDefault("DB_SSLMODE", config.SSLDisable)

	v.SetDefault("JWT_ACCESS_DURATION", "15m")
	v.SetDefault("JWT_REFRESH_DURATION", "168h")
}

func bindEnv(v *viper.Viper) error {
	keys := []string{
		"APP_NAME",
		"APP_ENV",
		"APP_VERSION",
		"GRPC_PORT",
		"SHUTDOWN_TIMEOUT",

		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SEARCH_PATH",
		"DB_SSLMODE",

		"JWT_SECRET",
		"JWT_ACCESS_DURATION",
		"JWT_REFRESH_DURATION",
	}

	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return fmt.Errorf("bootstrap: bind env %s: %w", key, err)
		}
	}

	return nil
}
