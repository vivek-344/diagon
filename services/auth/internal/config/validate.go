package config

import (
	"errors"
	"fmt"
	"strings"
)

type Validator struct {
	errs []error
}

func (c *Config) Validate() error {
	var v Validator

	v.Add(c.App.validate())
	v.Add(c.DB.validate())
	v.Add(c.JWT.validate())

	return v.Err()
}

func (a AppConfig) validate() error {
	var v Validator
	required := map[string]string{
		"APP_NAME": a.Name,
	}

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			v.Add(fmt.Errorf("config: %s is required", key))
		}
	}

	if !a.Env.Valid() {
		v.Add(fmt.Errorf("config: invalid APP_ENV: %q", a.Env))
	}

	if validatePort(a.GRPCPort) != nil {
		v.Add(fmt.Errorf("config: invalid GRPC_PORT"))
	}

	if a.ShutdownTimeout <= 0 {
		v.Add(fmt.Errorf("config: invalid SHUTDOWN_TIMEOUT"))
	}

	return v.Err()
}

func (d DBConfig) validate() error {
	var v Validator

	required := map[string]string{
		"DB_HOST":        d.Host,
		"DB_USER":        d.User,
		"DB_PASSWORD":    d.Password,
		"DB_NAME":        d.Name,
		"DB_SEARCH_PATH": d.SearchPath,
	}

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			v.Add(fmt.Errorf("config: missing required environment variable: %s", key))
		}
	}

	if validatePort(d.Port) != nil {
		v.Add(fmt.Errorf("config: invalid DB_PORT"))
	}

	if !d.SSLMode.Valid() {
		v.Add(fmt.Errorf("config: invalid DB_SSLMODE: %q", d.SSLMode))
	}

	return v.Err()
}

func (j JWTConfig) validate() error {
	var v Validator

	if len(j.SigningKey) == 0 {
		v.Add(fmt.Errorf("config: missing required environment variable: JWT_SECRET"))
	}

	if len(j.SigningKey) < 32 {
		v.Add(fmt.Errorf("config: JWT_SECRET must be at least 32 characters long"))
	}

	if j.AccessDuration <= 0 {
		v.Add(fmt.Errorf("config: invalid JWT_ACCESS_DURATION"))
	}

	if j.RefreshDuration <= 0 {
		v.Add(fmt.Errorf("config: invalid JWT_REFRESH_DURATION"))
	}

	if j.RefreshDuration <= j.AccessDuration {
		v.Add(fmt.Errorf("config: JWT_REFRESH_DURATION must be greater than JWT_ACCESS_DURATION"))
	}

	return v.Err()
}

func (v *Validator) Add(err error) {
	if err != nil {
		v.errs = append(v.errs, err)
	}
}

func (v *Validator) Err() error {
	return errors.Join(v.errs...)
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("config: invalid port: %d", port)
	}
	return nil
}

func (e Environment) Valid() bool {
	switch e {
	case EnvDevelopment, EnvStaging, EnvProduction:
		return true
	default:
		return false
	}
}

func (s SSLMode) Valid() bool {
	switch s {
	case SSLDisable, SSLRequire:
		return true
	default:
		return false
	}
}
