package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
	pb "github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/proto/gen"
)

// AuthClient provides gRPC client methods for authentication operations.
type AuthClient struct {
	client pb.AuthClient
	conn   *grpc.ClientConn
}

// NewAuthClient creates a new gRPC authentication client.
// Returns error if connection fails.
func NewAuthClient(serverAddr string, creds credentials.TransportCredentials) (*AuthClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc chanel: %w", err)
	}

	client := pb.NewAuthClient(conn)
	return &AuthClient{client: client, conn: conn}, nil
}

// Close terminates the gRPC connection.
// Returns error if connection closing fails.
func (c *AuthClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Register creates a new user account with the given credentials.
// Returns JWT token on success or error if registration fails.
func (c *AuthClient) Register(ctx context.Context, login, passsword string) (string, error) {
	resp, err := c.client.Register(ctx, &pb.AuthRequest{
		Login:    login,
		Password: passsword,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return "", domain.ErrLoginAlreayExists
			default:
				return "", fmt.Errorf("registration failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("registration failed: can't parse gRPC response code: %w", err)
		}
	}

	token := resp.GetToken()
	if token == "" {
		return "", fmt.Errorf("received empty token")
	}

	return token, nil
}

// Login authenticates a user with the given credentials.
// Returns JWT token on success or error if authentication fails.
func (c *AuthClient) Login(ctx context.Context, login, passsword string) (string, error) {
	resp, err := c.client.Login(ctx, &pb.AuthRequest{
		Login:    login,
		Password: passsword,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return "", domain.ErrLoginAlreayExists
			case codes.InvalidArgument:
				return "", domain.ErrInvalidCredentials
			default:
				return "", fmt.Errorf("registration failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("registration failed: can't parse gRPC response code: %w", err)
		}
	}

	token := resp.GetToken()
	if token == "" {
		return "", fmt.Errorf("received empty token")
	}

	return token, nil
}
