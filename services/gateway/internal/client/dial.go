package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial opens a gRPC client connection with insecure transport and unary retry+jitter.
func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UnaryRetryInterceptor(DefaultRetryConfig())),
	}
	return grpc.NewClient(addr, append(base, opts...)...)
}
