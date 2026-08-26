package app

import (
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

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("bootstrap: read config: %w", err)
		}
	}

	v.AutomaticEnv()

	// Defaults
	v.SetDefault("APP_ENV", config.EnvDevelopment)
	v.SetDefault("APP_VERSION", "dev")
	v.SetDefault("GRPC_PORT", DefaultGRPCPort)

	v.SetDefault("DB_PORT", DefaultDBPort)
	v.SetDefault("DB_SSLMODE", config.SSLDisable)

	v.SetDefault("JWT_ACCESS_DURATION", "15m")
	v.SetDefault("JWT_REFRESH_DURATION", "168h")

	var cfg config.Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("bootstrap: failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
