package server

import (
	"fmt"
	"net"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"github.com/vivek-344/diagon/services/auth/internal/config"
	"github.com/vivek-344/diagon/services/auth/internal/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(
	cfg config.AppConfig,
	authHandler *handler.AuthHandler,
) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%d", cfg.GRPCPort),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"server: listen: %w",
			err,
		)
	}

	grpcServer := grpc.NewServer()

	authv1.RegisterAuthServiceServer(
		grpcServer,
		authHandler,
	)

	reflection.Register(grpcServer)

	return grpcServer, listener, nil
}
