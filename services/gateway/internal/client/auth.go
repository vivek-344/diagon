package client

import (
	"context"
	"fmt"

	authv1 "github.com/vivek-344/diagon/gen/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(
	address string,
) (*AuthClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"client: create auth connection: %w",
			err,
		)
	}

	return &AuthClient{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) Register(
	ctx context.Context,
	email string,
	password string,
) (*authv1.RegisterResponse, error) {
	return c.client.Register(
		ctx,
		&authv1.RegisterRequest{
			Email:    email,
			Password: password,
		},
	)
}
