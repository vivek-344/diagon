package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vivek-344/diagon/services/auth/internal/server"
)

func (a *App) Run() error {
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
		return err
	}

	a.logger.Info(
		"database health check passed",
		"database_address", fmt.Sprintf(
			"%s:%d/%s",
			a.cfg.DB.Host,
			a.cfg.DB.Port,
			a.cfg.DB.Name,
		),
		"schema", schema,
	)

	grpcServer, listener, err := server.NewGRPCServer(
		a.cfg.App,
		a.authHandler,
	)
	if err != nil {
		return fmt.Errorf("app: create grpc server: %w", err)
	}

	a.logger.Info(
		"auth service initialized",
		"grpc_address", listener.Addr().String(),
	)

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("app: grpc serve: %w", err)
	}

	return nil
}
