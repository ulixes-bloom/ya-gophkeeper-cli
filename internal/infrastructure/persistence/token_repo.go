package persistence

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

// TokenRepo implements token storage and retrieval using filesystem.
// It stores tokens in the user's home directory under .gophkeeper-cli/token.txt
type TokenRepo struct {
	tokenPath string
}

// NewTokenRepo creates a new TokenRepo instance.
// It verifies that the token storage directory is accessible.
func NewTokenRepo() (*TokenRepo, error) {
	repo := &TokenRepo{}
	// Verify we can resolve the token path during initialization
	path, err := getTokenFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token repository: %w", err)
	}
	repo.tokenPath = path

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("failed to create token directory: %w", err)
	}

	return repo, nil
}

// GetToken retrieves the token from secure storage.
// Returns ErrTokenNotFound if no token exists.
func (r *TokenRepo) GetToken() (string, error) {
	data, err := os.ReadFile(r.tokenPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", domain.ErrTokenNotFound
		}
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	if len(data) == 0 {
		return "", domain.ErrTokenNotFound
	}

	return string(data), nil
}

// SaveToken saves the token to storage (file)
func (r *TokenRepo) SaveToken(token string) error {
	if token == "" {
		return errors.New("cannot save empty token")
	}

	// Ensure the directory exists (in case it was deleted after initialization)
	if err := os.MkdirAll(filepath.Dir(r.tokenPath), 0700); err != nil {
		return fmt.Errorf("failed to ensure token directory exists: %w", err)
	}

	if err := os.WriteFile(r.tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token: %w", err)
	}

	return nil
}

// getTokenFilePath returns the standardized path for token storage.
// Uses the user's home directory with a hidden application subdirectory.
func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %v", err)
	}

	return filepath.Join(homeDir, ".gophkeeper-cli", "token.txt"), nil
}
