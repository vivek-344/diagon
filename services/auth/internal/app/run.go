package app

import (
	"context"
	"fmt"
	"time"
)

func (a *App) Run() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	var schema string

	if err := a.db.QueryRowContext(
		ctx,
		"SELECT current_schema()",
	).Scan(&schema); err != nil {
		a.logger.Error(
			"database health check failed",
			"error", err,
		)
		return
	}

	a.logger.Info(
		"auth service initialized",
		"grpc_port", a.cfg.App.GRPCPort,
		"database_address", fmt.Sprintf(
			"%s:%d/%s",
			a.cfg.DB.Host,
			a.cfg.DB.Port,
			a.cfg.DB.Name,
		),
		"schema", schema,
	)
}
