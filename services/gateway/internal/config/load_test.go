package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		env     map[string]string
		want    *Config
		wantErr string
	}{
		{
			name: "loads config from file",
			config: `
APP_NAME=api-gateway
APP_ENV=production
APP_VERSION=1.2.3
HTTP_PORT=8080
SHUTDOWN_TIMEOUT=15s
AUTH_GRPC_ADDRESS=localhost:50051
`,
			want: &Config{
				App: AppConfig{
					Name:            "api-gateway",
					Env:             "production",
					Version:         "1.2.3",
					HTTPPort:        8080,
					ShutdownTimeout: 15 * time.Second,
				},
				Auth: AuthConfig{
					GRPCAddress: "localhost:50051",
				},
			},
		},
		{
			name: "uses defaults",
			config: `
APP_NAME=api-gateway
`,
			want: &Config{
				App: AppConfig{
					Name:            "api-gateway",
					Env:             "development",
					Version:         "dev",
					HTTPPort:        8080,
					ShutdownTimeout: 10 * time.Second,
				},
				Auth: AuthConfig{
					GRPCAddress: "localhost:50051",
				},
			},
		},
		{
			name: "uses environment variables",
			config: `
APP_NAME=api-gateway
`,
			env: map[string]string{
				"APP_ENV":           "staging",
				"APP_VERSION":       "2.0.0",
				"HTTP_PORT":         "9090",
				"SHUTDOWN_TIMEOUT":  "30s",
				"AUTH_GRPC_ADDRESS": "auth:50051",
			},
			want: &Config{
				App: AppConfig{
					Name:            "api-gateway",
					Env:             "staging",
					Version:         "2.0.0",
					HTTPPort:        9090,
					ShutdownTimeout: 30 * time.Second,
				},
				Auth: AuthConfig{
					GRPCAddress: "auth:50051",
				},
			},
		},
		{
			name: "missing config file uses environment and defaults",
			env: map[string]string{
				"APP_NAME": "api-gateway",
			},
			want: &Config{
				App: AppConfig{
					Name:            "api-gateway",
					Env:             "development",
					Version:         "dev",
					HTTPPort:        8080,
					ShutdownTimeout: 10 * time.Second,
				},
				Auth: AuthConfig{
					GRPCAddress: "localhost:50051",
				},
			},
		},
		{
			name: "missing app name",
			config: `
APP_ENV=development
`,
			wantErr: "config: APP_NAME is required",
		},
		{
			name: "missing auth grpc address",
			config: `
APP_NAME=api-gateway
AUTH_GRPC_ADDRESS=
`,
			wantErr: "config: AUTH_GRPC_ADDRESS is required",
		},
		{
			name: "invalid http port above maximum",
			config: `
APP_NAME=api-gateway
HTTP_PORT=70000
`,
			wantErr: "config: invalid HTTP_PORT: 70000",
		},
		{
			name: "invalid http port below minimum",
			config: `
APP_NAME=api-gateway
HTTP_PORT=0
`,
			wantErr: "config: invalid HTTP_PORT: 0",
		},
		{
			name: "invalid shutdown timeout",
			config: `
APP_NAME=api-gateway
SHUTDOWN_TIMEOUT=invalid
`,
			wantErr: "config: parse SHUTDOWN_TIMEOUT:",
		},
		{
			name: "zero shutdown timeout",
			config: `
APP_NAME=api-gateway
SHUTDOWN_TIMEOUT=0s
`,
			wantErr: "config: invalid SHUTDOWN_TIMEOUT: 0s",
		},
		{
			name: "negative shutdown timeout",
			config: `
APP_NAME=api-gateway
SHUTDOWN_TIMEOUT=-5s
`,
			wantErr: "config: invalid SHUTDOWN_TIMEOUT: -5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure tests do not accidentally use the project's .env.
			t.Setenv(
				"CONFIG_FILE",
				filepath.Join(
					t.TempDir(),
					"does-not-exist.env",
				),
			)

			// Clear all environment variables read by Load.
			for _, key := range []string{
				"APP_NAME",
				"APP_ENV",
				"APP_VERSION",
				"HTTP_PORT",
				"SHUTDOWN_TIMEOUT",
				"AUTH_GRPC_ADDRESS",
			} {
				t.Setenv(key, "")
			}

			// Set test-specific environment variables.
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			// If this test provides a config file, create it.
			if tt.config != "" {
				configFile := filepath.Join(
					t.TempDir(),
					"config.env",
				)

				err := os.WriteFile(
					configFile,
					[]byte(strings.TrimSpace(tt.config)),
					0600,
				)
				if err != nil {
					t.Fatalf(
						"write config file: %v",
						err,
					)
				}

				t.Setenv("CONFIG_FILE", configFile)
			}

			got, err := Load()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf(
						"expected error containing %q, got nil",
						tt.wantErr,
					)
				}

				if !strings.Contains(
					err.Error(),
					tt.wantErr,
				) {
					t.Fatalf(
						"expected error containing %q, got %q",
						tt.wantErr,
						err.Error(),
					)
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"unexpected error: %v",
					err,
				)
			}

			if got == nil {
				t.Fatal("expected config, got nil")
			}

			if got.App.Name != tt.want.App.Name {
				t.Errorf(
					"App.Name = %q, want %q",
					got.App.Name,
					tt.want.App.Name,
				)
			}

			if got.App.Env != tt.want.App.Env {
				t.Errorf(
					"App.Env = %q, want %q",
					got.App.Env,
					tt.want.App.Env,
				)
			}

			if got.App.Version != tt.want.App.Version {
				t.Errorf(
					"App.Version = %q, want %q",
					got.App.Version,
					tt.want.App.Version,
				)
			}

			if got.App.HTTPPort != tt.want.App.HTTPPort {
				t.Errorf(
					"App.HTTPPort = %d, want %d",
					got.App.HTTPPort,
					tt.want.App.HTTPPort,
				)
			}

			if got.App.ShutdownTimeout != tt.want.App.ShutdownTimeout {
				t.Errorf(
					"App.ShutdownTimeout = %v, want %v",
					got.App.ShutdownTimeout,
					tt.want.App.ShutdownTimeout,
				)
			}

			if got.Auth.GRPCAddress != tt.want.Auth.GRPCAddress {
				t.Errorf(
					"Auth.GRPCAddress = %q, want %q",
					got.Auth.GRPCAddress,
					tt.want.Auth.GRPCAddress,
				)
			}
		})
	}
}

func TestLoad_ConfigFileReadError(t *testing.T) {
	// A directory cannot be read as a config file.
	configDir := t.TempDir()

	t.Setenv("CONFIG_FILE", configDir)

	_, err := Load()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "config: read config:") {
		t.Fatalf(
			"expected read config error, got %q",
			err.Error(),
		)
	}
}
