package domain

import (
	"context"
	"errors"
)

// AuthReq represents authentication request data containing login credentials.
type AuthReq struct {
	Login    string
	Password string
}

// AuthService defines the interface for authentication operations.
// Implementations should handle user registration and login processes.
type AuthService interface {
	// Register creates a new user account with the provided credentials.
	// Returns an error if registration fails (e.g., user already exists).
	Register(ctx context.Context, login, passsword string) error

	// Login authenticates a user with the provided credentials.
	// Returns an error if authentication fails (e.g., invalid credentials).
	Login(ctx context.Context, login, passsword string) error
}

// AuthClient defines the interface for authentication client operations.
// Implementations should communicate with the authentication server.
type AuthClient interface {
	// Register sends a registration request to the authentication server.
	// Returns a JWT token on success or an error if registration fails.
	Register(ctx context.Context, login, passsword string) (string, error)

	// Login sends an authentication request to the authentication server.
	// Returns a JWT token on success or an error if authentication fails.
	Login(ctx context.Context, login, passsword string) (string, error)
}

// TokenRepository defines the interface for token storage operations.
// Implementations should handle secure token persistence and retrieval.
type TokenRepository interface {
	// GetToken retrieves the stored authentication token.
	// Returns the token or an error if retrieval fails (e.g., token not found).
	GetToken() (string, error)

	// SaveToken stores the authentication token securely.
	// Returns an error if storage fails.
	SaveToken(token string) error
}

var (
	ErrTokenNotFound      = errors.New("token not found in storage")
	ErrLoginAlreayExists  = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)
