package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vivek-344/diagon/services/auth/internal/server"
)

func (a *App) Run(ctx context.Context) error {
	healthCtx, cancel := context.WithTimeout(
		ctx,
		2*time.Second,
	)
	defer cancel()

	var schema string

	if err := a.db.QueryRowContext(
		healthCtx,
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

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info(
			"auth grpc server listening",
			"grpc_address", listener.Addr().String(),
		)

		serverErr <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf(
			"app: grpc serve: %w",
			err,
		)

	case <-ctx.Done():
		a.logger.Info(
			"shutting down auth service",
		)

		grpcServer.GracefulStop()

		return nil
	}
}
