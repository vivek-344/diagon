package config

import "time"

type Config struct {
	App  AppConfig
	Auth AuthConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Version         string
	HTTPPort        int
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	GRPCAddress string
}
