package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	v := viper.New()

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = ".env"
	}

	v.SetConfigFile(configFile)
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) && !os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"config: read config: %w",
				err,
			)
		}
	}

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_VERSION", "dev")
	v.SetDefault("HTTP_PORT", 8080)
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")
	v.SetDefault("AUTH_GRPC_ADDRESS", "localhost:50051")

	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_VERSION",
		"HTTP_PORT",
		"SHUTDOWN_TIMEOUT",
		"AUTH_GRPC_ADDRESS",
	} {
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf(
				"config: bind %s: %w",
				key,
				err,
			)
		}
	}

	port := v.GetInt("HTTP_PORT")
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf(
			"config: invalid HTTP_PORT: %d",
			port,
		)
	}

	if v.GetString("APP_NAME") == "" {
		return nil, fmt.Errorf(
			"config: APP_NAME is required",
		)
	}

	if v.GetString("AUTH_GRPC_ADDRESS") == "" {
		return nil, fmt.Errorf(
			"config: AUTH_GRPC_ADDRESS is required",
		)
	}

	shutdownTimeout, err := time.ParseDuration(
		v.GetString("SHUTDOWN_TIMEOUT"),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"config: parse SHUTDOWN_TIMEOUT: %w",
			err,
		)
	}

	if shutdownTimeout <= 0 {
		return nil, fmt.Errorf(
			"config: invalid SHUTDOWN_TIMEOUT: %v",
			shutdownTimeout,
		)
	}

	return &Config{
		App: AppConfig{
			Name:            v.GetString("APP_NAME"),
			Env:             v.GetString("APP_ENV"),
			Version:         v.GetString("APP_VERSION"),
			HTTPPort:        port,
			ShutdownTimeout: shutdownTimeout,
		},
		Auth: AuthConfig{
			GRPCAddress: v.GetString("AUTH_GRPC_ADDRESS"),
		},
	}, nil
}
