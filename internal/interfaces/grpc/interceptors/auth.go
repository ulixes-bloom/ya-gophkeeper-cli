// Package interceptors provides gRPC interceptors for handling cross-cutting concerns such as authentication.
package interceptors

import (
	"context"
	"fmt"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor is responsible for injecting authentication tokens into gRPC requests.
type AuthInterceptor struct {
	repo domain.TokenRepository
}

// NewAuthInterceptor creates a new instance of AuthInterceptor with the given TokenRepository.
func NewAuthInterceptor(repo domain.TokenRepository) *AuthInterceptor {
	return &AuthInterceptor{
		repo: repo,
	}
}

// UnaryInterceptor intercepts unary gRPC calls to add authentication metadata.
func (i *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctx, err := i.injectAuthMetadata(ctx)
	if err != nil {
		return err
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamInterceptor intercepts streaming gRPC calls to add authentication metadata.
func (i *AuthInterceptor) StreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	ctx, err := i.injectAuthMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// injectAuthMetadata retrieves the authentication token from the repository and injects it into the context metadata.
func (i *AuthInterceptor) injectAuthMetadata(ctx context.Context) (context.Context, error) {
	token, err := i.repo.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	// Add the token to metadata as an authorization header.
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md), nil
}
