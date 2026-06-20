package client

import (
    "fmt"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/Eastwesser/event-horizon/services/auth/proto"  // внимательнее с путём
)

type AuthClient struct {
    conn   *grpc.ClientConn
    client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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