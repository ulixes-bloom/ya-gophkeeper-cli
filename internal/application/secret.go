package application

import (
	"context"
	"fmt"
	"io"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

// SecretService provides operations for managing secrets of different types.
// It handles both regular secrets and streaming secrets (like files).
type SecretService struct {
	client domain.SecretClient
}

// NewSecretService creates a new instance of SecretService with the required client.

func NewSecretService(client domain.SecretClient) *SecretService {
	return &SecretService{client: client}
}

// CreateSecret creates a new secret based on its type.
// For credential and payment card types, it reads all data from the reader first.
// For file and text types, it streams the content directly.
// Returns an error if the operation fails.
func (s *SecretService) CreateSecret(ctx context.Context, secret domain.Secret, contentReader io.Reader) error {
	switch secret.Info.Type {
	case domain.CredentialsSecretType, domain.PaymentCardSecretType:
		secretData, err := io.ReadAll(contentReader)
		if err != nil {
			return fmt.Errorf("application.CreateSecret: failed to read secret content: %w", err)
		}
		secret.Data = string(secretData)

		err = s.client.CreateSecret(ctx, secret)
		if err != nil {
			return fmt.Errorf("application.CreateSecret: failed to create secret: %w", err)
		}
	case domain.FileSecretType, domain.TextSecretType:
		err := s.client.CreateSecretStream(ctx, secret, contentReader)
		if err != nil {
			return fmt.Errorf("application.CreateSecret: failed to create stream secret: %w", err)
		}
	}

	return nil
}

// ListSecrets retrieves a list of all secret names available to the user.
// Returns an error if the operation fails.
func (s *SecretService) ListSecrets(ctx context.Context) ([]string, error) {
	secretList, err := s.client.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("application.ListSecrets: %w", err)
	}
	return secretList, nil
}

// GetLatestSecret retrieves the most recent version of a secret by name.
// Returns the secret or an error if the operation fails.
func (s *SecretService) GetLatestSecret(ctx context.Context, secretName string) (*domain.Secret, error) {
	secret, err := s.client.GetLatestSecret(ctx, secretName)
	if err != nil {
		return nil, fmt.Errorf("application.GetLatestSecret: %w", err)
	}
	return secret, nil
}

// GetLatestSecretStream retrieves the most recent version of a streamable secret by name.
// Returns a reader for the content, secret info, or an error if the operation fails.
func (s *SecretService) GetLatestSecretStream(ctx context.Context, secretName string) (io.Reader, *domain.SecretInfo, error) {
	stream, secretInfo, err := s.client.GetLatestSecretStream(ctx, secretName)
	if err != nil {
		return nil, nil, fmt.Errorf("application.GetLatestSecretStream: %w", err)
	}
	return stream, secretInfo, nil
}

// GetSecretByVersion retrieves a specific version of a secret by name and version number.
// Returns the secret or an error if the operation fails.
func (s *SecretService) GetSecretByVersion(ctx context.Context, secretName string, version int32) (*domain.Secret, error) {
	secret, err := s.client.GetSecretByVersion(ctx, secretName, version)
	if err != nil {
		return nil, fmt.Errorf("application.GetSecretByVersion: %w", err)
	}
	return secret, nil
}

// GetSecretStreamByVersion retrieves a specific version of a streamable secret by name and version.
// Returns a reader for the content, secret info, or an error if the operation fails.
func (s *SecretService) GetSecretStreamByVersion(ctx context.Context, secretName string, version int32) (io.Reader, *domain.SecretInfo, error) {
	stream, secretInfo, err := s.client.GetSecretStreamByVersion(ctx, secretName, version)
	if err != nil {
		return nil, nil, fmt.Errorf("application.GetSecretStreamByVersion: %w", err)
	}
	return stream, secretInfo, nil
}

// DeleteSecret removes a secret and all its versions by name.
// Returns an error if the operation fails.
func (s *SecretService) DeleteSecret(ctx context.Context, secretName string) error {
	err := s.client.DeleteSecret(ctx, secretName)
	if err != nil {
		return fmt.Errorf("application.DeleteSecret: %w", err)
	}
	return nil
}
