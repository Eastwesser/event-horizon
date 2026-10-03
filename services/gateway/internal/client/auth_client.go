package client

import (
	"fmt"
	"log"

	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
	"google.golang.org/grpc"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	log.Printf("Connected to Auth gRPC server at %s", addr)

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) GetClient() pb.AuthServiceClient {
	return c.client
}
