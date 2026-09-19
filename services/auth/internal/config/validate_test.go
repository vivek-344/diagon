package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		App: AppConfig{
			Name:            "auth-service",
			Env:             EnvDevelopment,
			GRPCPort:        50051,
			ShutdownTimeout: 10 * time.Second,
		},
		DB: DBConfig{
			Host:       "localhost",
			Port:       5432,
			User:       "postgres",
			Password:   "password",
			Name:       "auth",
			SearchPath: "auth",
			SSLMode:    SSLDisable,
		},
		JWT: JWTConfig{
			SigningKey:      "12345678901234567890123456789012",
			AccessDuration:  15 * time.Minute,
			RefreshDuration: 7 * 24 * time.Hour,
		},
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantValid bool
		wantErrs  []string
	}{
		{
			name:      "valid config",
			config:    validConfig(),
			wantValid: true,
		},
		{
			name: "missing app name",
			config: func() Config {
				cfg := validConfig()
				cfg.App.Name = ""
				return cfg
			}(),
			wantErrs: []string{
				"config: APP_NAME is required",
			},
		},
		{
			name: "invalid app environment",
			config: func() Config {
				cfg := validConfig()
				cfg.App.Env = Environment("invalid")
				return cfg
			}(),
			wantErrs: []string{
				`config: invalid APP_ENV: "invalid"`,
			},
		},
		{
			name: "invalid grpc port",
			config: func() Config {
				cfg := validConfig()
				cfg.App.GRPCPort = 0
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid GRPC_PORT: 0",
			},
		},
		{
			name: "invalid shutdown timeout",
			config: func() Config {
				cfg := validConfig()
				cfg.App.ShutdownTimeout = 0
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid SHUTDOWN_TIMEOUT: 0s",
			},
		},
		{
			name: "missing database fields",
			config: func() Config {
				cfg := validConfig()
				cfg.DB.Host = ""
				cfg.DB.User = ""
				cfg.DB.Password = ""
				cfg.DB.Name = ""
				cfg.DB.SearchPath = ""
				return cfg
			}(),
			wantErrs: []string{
				"config: DB_HOST is required",
				"config: DB_USER is required",
				"config: DB_PASSWORD is required",
				"config: DB_NAME is required",
				"config: DB_SEARCH_PATH is required",
			},
		},
		{
			name: "invalid database port",
			config: func() Config {
				cfg := validConfig()
				cfg.DB.Port = 65536
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid DB_PORT: 65536",
			},
		},
		{
			name: "invalid database ssl mode",
			config: func() Config {
				cfg := validConfig()
				cfg.DB.SSLMode = SSLMode("invalid")
				return cfg
			}(),
			wantErrs: []string{
				`config: invalid DB_SSLMODE: "invalid"`,
			},
		},
		{
			name: "missing jwt secret",
			config: func() Config {
				cfg := validConfig()
				cfg.JWT.SigningKey = ""
				return cfg
			}(),
			wantErrs: []string{
				"config: missing required environment variable: JWT_SECRET",
			},
		},
		{
			name: "short jwt secret",
			config: func() Config {
				cfg := validConfig()
				cfg.JWT.SigningKey = "short"
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid JWT_SECRET: must be at least 32 characters long",
			},
		},
		{
			name: "invalid access duration",
			config: func() Config {
				cfg := validConfig()
				cfg.JWT.AccessDuration = 0
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid JWT_ACCESS_DURATION: 0s",
			},
		},
		{
			name: "invalid refresh duration",
			config: func() Config {
				cfg := validConfig()
				cfg.JWT.RefreshDuration = 0
				return cfg
			}(),
			wantErrs: []string{
				"config: invalid JWT_REFRESH_DURATION: 0s",
			},
		},
		{
			name: "refresh duration not greater than access duration",
			config: func() Config {
				cfg := validConfig()
				cfg.JWT.AccessDuration = time.Hour
				cfg.JWT.RefreshDuration = time.Hour
				return cfg
			}(),
			wantErrs: []string{
				"config: JWT_REFRESH_DURATION must be greater than JWT_ACCESS_DURATION",
			},
		},
		{
			name: "multiple validation errors",
			config: Config{
				App: AppConfig{
					Name:            "",
					Env:             Environment("invalid"),
					GRPCPort:        0,
					ShutdownTimeout: 0,
				},
				DB: DBConfig{
					Host:       "",
					Port:       0,
					User:       "",
					Password:   "",
					Name:       "",
					SearchPath: "",
					SSLMode:    SSLMode("invalid"),
				},
				JWT: JWTConfig{
					SigningKey:      "short",
					AccessDuration:  0,
					RefreshDuration: 0,
				},
			},
			wantErrs: []string{
				"config: APP_NAME is required",
				`config: invalid APP_ENV: "invalid"`,
				"config: invalid GRPC_PORT: 0",
				"config: invalid SHUTDOWN_TIMEOUT: 0s",
				"config: DB_HOST is required",
				"config: DB_USER is required",
				"config: DB_PASSWORD is required",
				"config: DB_NAME is required",
				"config: DB_SEARCH_PATH is required",
				"config: invalid DB_PORT: 0",
				`config: invalid DB_SSLMODE: "invalid"`,
				"config: invalid JWT_SECRET: must be at least 32 characters long",
				"config: invalid JWT_ACCESS_DURATION: 0s",
				"config: invalid JWT_REFRESH_DURATION: 0s",
				"config: JWT_REFRESH_DURATION must be greater than JWT_ACCESS_DURATION",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantValid {
				if err != nil {
					t.Fatalf("expected valid config, got error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			errMessage := err.Error()

			for _, wantErr := range tt.wantErrs {
				if !strings.Contains(errMessage, wantErr) {
					t.Errorf(
						"expected error to contain %q, got %q",
						wantErr,
						errMessage,
					)
				}
			}
		})
	}
}

func TestIsPortValid(t *testing.T) {
	tests := []struct {
		name string
		port int
		want bool
	}{
		{
			name: "minimum valid port",
			port: 1,
			want: true,
		},
		{
			name: "maximum valid port",
			port: 65535,
			want: true,
		},
		{
			name: "zero",
			port: 0,
			want: false,
		},
		{
			name: "negative",
			port: -1,
			want: false,
		},
		{
			name: "above maximum",
			port: 65536,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPortValid(tt.port); got != tt.want {
				t.Fatalf(
					"isPortValid(%d) = %v, want %v",
					tt.port,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEnvironment_Valid(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		want bool
	}{
		{
			name: "development",
			env:  EnvDevelopment,
			want: true,
		},
		{
			name: "staging",
			env:  EnvStaging,
			want: true,
		},
		{
			name: "production",
			env:  EnvProduction,
			want: true,
		},
		{
			name: "invalid",
			env:  Environment("invalid"),
			want: false,
		},
		{
			name: "empty",
			env:  Environment(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.Valid(); got != tt.want {
				t.Fatalf(
					"Environment.Valid() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestSSLMode_Valid(t *testing.T) {
	tests := []struct {
		name string
		mode SSLMode
		want bool
	}{
		{
			name: "disable",
			mode: SSLDisable,
			want: true,
		},
		{
			name: "require",
			mode: SSLRequire,
			want: true,
		},
		{
			name: "verify ca",
			mode: SSLVerifyCA,
			want: true,
		},
		{
			name: "verify full",
			mode: SSLVerifyFull,
			want: true,
		},
		{
			name: "invalid",
			mode: SSLMode("invalid"),
			want: false,
		},
		{
			name: "empty",
			mode: SSLMode(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.Valid(); got != tt.want {
				t.Fatalf(
					"SSLMode.Valid() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestRequire(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			name:      "non-empty",
			value:     "value",
			wantError: false,
		},
		{
			name:      "empty",
			value:     "",
			wantError: true,
		},
		{
			name:      "whitespace",
			value:     "   ",
			wantError: true,
		},
		{
			name:      "value with surrounding whitespace",
			value:     " value ",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Validator

			require(&v, "TEST_VALUE", tt.value)

			err := v.Err()

			if tt.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if tt.wantError {
				want := "config: TEST_VALUE is required"

				if err.Error() != want {
					t.Fatalf(
						"expected %q, got %q",
						want,
						err.Error(),
					)
				}
			}
		})
	}
}

func TestValidator_Add(t *testing.T) {
	var v Validator

	if err := v.Err(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	v.Add(nil)

	if err := v.Err(); err != nil {
		t.Fatalf("expected nil error after Add(nil), got %v", err)
	}

	v.Add(errors.New("first error"))
	v.Add(errors.New("second error"))

	err := v.Err()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "first error") {
		t.Errorf("expected first error, got %v", err)
	}

	if !strings.Contains(err.Error(), "second error") {
		t.Errorf("expected second error, got %v", err)
	}
}
