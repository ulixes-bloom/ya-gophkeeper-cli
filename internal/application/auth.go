package application

import (
	"context"
	"fmt"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

// AuthService provides authentication-related operations (register, login).
type AuthService struct {
	client          domain.AuthClient
	tokenRepository domain.TokenRepository
}

// NewAuthService creates a new instance of AuthService with the required dependencies.
func NewAuthService(client domain.AuthClient, tokenRepository domain.TokenRepository) *AuthService {
	return &AuthService{
		client:          client,
		tokenRepository: tokenRepository,
	}
}

// Register creates a new user account with the given credentials.
// On success, it saves the received JWT token to the token repository.
// Returns an error if registration fails or token can't be saved.
func (s *AuthService) Register(ctx context.Context, login, passsword string) error {
	jwtToken, err := s.client.Register(ctx, login, passsword)
	if err != nil {
		return fmt.Errorf("application.Register: %w", err)
	}
	if err = s.tokenRepository.SaveToken(jwtToken); err != nil {
		return fmt.Errorf("application.Register: %w", err)
	}
	return nil
}

// Login authenticates a user with the given credentials.
// On success, it saves the received JWT token to the token repository.
// Returns an error if authentication fails or token can't be saved.
func (s *AuthService) Login(ctx context.Context, login, passsword string) error {
	jwtToken, err := s.client.Login(ctx, login, passsword)
	if err != nil {
		return fmt.Errorf("application.Login: %w", err)
	}
	if err = s.tokenRepository.SaveToken(jwtToken); err != nil {
		return fmt.Errorf("application.Login: %w", err)
	}
	return nil
}
