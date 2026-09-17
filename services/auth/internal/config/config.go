package config

import "time"

type SSLMode string

const (
	SSLDisable    SSLMode = "disable"
	SSLRequire    SSLMode = "require"
	SSLVerifyCA   SSLMode = "verify-ca"
	SSLVerifyFull SSLMode = "verify-full"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

type Config struct {
	App AppConfig `mapstructure:",squash"`
	DB  DBConfig  `mapstructure:",squash"`
	JWT JWTConfig `mapstructure:",squash"`
}

type AppConfig struct {
	Name            string        `mapstructure:"APP_NAME"`
	Env             Environment   `mapstructure:"APP_ENV"`
	Version         string        `mapstructure:"APP_VERSION"`
	GRPCPort        int           `mapstructure:"GRPC_PORT"`
	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT"`
}

type DBConfig struct {
	Host       string  `mapstructure:"DB_HOST"`
	Port       int     `mapstructure:"DB_PORT"`
	User       string  `mapstructure:"DB_USER"`
	Password   string  `mapstructure:"DB_PASSWORD"`
	Name       string  `mapstructure:"DB_NAME"`
	SearchPath string  `mapstructure:"DB_SEARCH_PATH"`
	SSLMode    SSLMode `mapstructure:"DB_SSLMODE"`
}

type JWTConfig struct {
	SigningKey      string        `mapstructure:"JWT_SECRET"`
	AccessDuration  time.Duration `mapstructure:"JWT_ACCESS_DURATION"`
	RefreshDuration time.Duration `mapstructure:"JWT_REFRESH_DURATION"`
}
