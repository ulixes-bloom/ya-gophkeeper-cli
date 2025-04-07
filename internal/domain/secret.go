package domain

import (
	"context"
	"fmt"
	"io"
	"time"
)

// SecretType represents the classification of different secret kinds.
type SecretType string

const (
	CredentialsSecretType SecretType = "credentials"
	PaymentCardSecretType SecretType = "payment_card"
	FileSecretType        SecretType = "file"
	TextSecretType        SecretType = "text"
)

// Secret represents a protected piece of information with its metadata.
// The Data field contains the actual secret content in string form.
type Secret struct {
	Info SecretInfo
	Data string
}

// SecretInfo contains metadata about a secret.
type SecretInfo struct {
	Name      string
	Type      SecretType
	Metadata  string
	Version   int32
	CreatedAt time.Time
}

// CredentialsSecret represents login/password credentials.
type CredentialsSecret struct {
	Login    string
	Password string
}

// PaymentCardSecret represents payment card information.
type PaymentCardSecret struct {
	Number string
}

// SecretClient defines the interface for client-server operations.
// Implementations should handle communication with the server backend.
type SecretClient interface {
	// CreateSecret stores a new secret (for non-streamable types).
	// Returns an error if the operation fails.
	CreateSecret(ctx context.Context, secret Secret) error

	// CreateSecretStream stores a new secret by streaming its content (for large data).
	// Returns an error if the operation fails.
	CreateSecretStream(ctx context.Context, SecretInfo Secret, reader io.Reader) error

	// ListSecrets retrieves all available secret names.
	// Returns a list of secret names or an error if the operation fails.
	ListSecrets(ctx context.Context) ([]string, error)

	// GetLatestSecret retrieves the most recent version of a secret.
	// Returns the secret or an error if the operation fails.
	GetLatestSecret(ctx context.Context, secretName string) (*Secret, error)

	// GetLatestSecretStream retrieves the most recent version of a streamable secret.
	// Returns a reader for the content, secret info, or an error if the operation fails.
	GetLatestSecretStream(ctx context.Context, secretName string) (io.Reader, *SecretInfo, error)

	// GetSecretByVersion retrieves a specific version of a secret.
	// Returns the secret or an error if the operation fails.
	GetSecretByVersion(ctx context.Context, secretName string, version int32) (*Secret, error)

	// GetSecretStreamByVersion retrieves a specific version of a streamable secret.
	// Returns a reader for the content, secret info, or an error if the operation fails.
	GetSecretStreamByVersion(ctx context.Context, secretName string, version int32) (io.Reader, *SecretInfo, error)

	// DeleteSecret removes a secret and all its versions.
	// Returns an error if the operation fails.
	DeleteSecret(ctx context.Context, secretName string) error
}

// SecretService defines the business logic operations for secret management.
type SecretService interface {
	// CreateSecret creates a new secret, handling both direct and streamed content
	// based on the secret type. Returns an error if the operation fails.
	CreateSecret(ctx context.Context, secret Secret, contentReader io.Reader) error

	// ListSecrets retrieves all available secret names.
	// Returns a list of secret names or an error if the operation fails.
	ListSecrets(ctx context.Context) ([]string, error)

	// GetLatestSecret retrieves the most recent version of a secret.
	// Returns the secret or an error if the operation fails.
	GetLatestSecret(ctx context.Context, secretName string) (*Secret, error)

	// GetLatestSecretStream retrieves the most recent version of a streamable secret.
	// Returns a reader for the content, secret info, or an error if the operation fails.
	GetLatestSecretStream(ctx context.Context, secretName string) (io.Reader, *SecretInfo, error)

	// GetSecretByVersion retrieves a specific version of a secret.
	// Returns the secret or an error if the operation fails.
	GetSecretByVersion(ctx context.Context, secretName string, version int32) (*Secret, error)

	// GetSecretStreamByVersion retrieves a specific version of a streamable secret.
	// Returns a reader for the content, secret info, or an error if the operation fails.
	GetSecretStreamByVersion(ctx context.Context, secretName string, version int32) (io.Reader, *SecretInfo, error)

	// DeleteSecret removes a secret and all its versions.
	// Returns an error if the operation fails.
	DeleteSecret(ctx context.Context, secretName string) error
}

var (
	ErrUnknownSecretType = fmt.Errorf("unknown secret type")
)
