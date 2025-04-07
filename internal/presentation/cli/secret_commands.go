package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

// newCreateCredentialsSecretCmd creates a command for storing credential secret.
func newCreateCredentialsSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var name, login, password, metadata string

	cmd := &cobra.Command{
		Use:   "create-credentials",
		Short: "Store login/password credentials",
		Long:  "Securely stores username/password combinations with optional metadata",

		RunE: func(cmd *cobra.Command, args []string) error {
			secret := domain.Secret{
				Info: domain.SecretInfo{
					Name:     name,
					Metadata: metadata,
					Type:     domain.CredentialsSecretType,
				},
			}
			credentials := domain.CredentialsSecret{
				Login:    login,
				Password: password,
			}
			marshaled, err := json.Marshal(credentials)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal credentials")
				return fmt.Errorf("failed to marshal credentials")
			}

			if err = secretService.CreateSecret(ctx, secret, bytes.NewReader(marshaled)); err != nil {
				log.Error().Err(err).Msg("failed to create secret")
				return fmt.Errorf("failed to create secret: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully stored credentials for '%s'\n", name)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name for the credentials (required)")
	cmd.Flags().StringVarP(&login, "login", "l", "", "Username/login (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// newCreatePaymentCardSecretCmd creates a command for storing payment card information.
func newCreatePaymentCardSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var name, number, metadata string

	cmd := &cobra.Command{
		Use:   "create-paymentcard",
		Short: "Store payment card information",
		Long:  `Securely stores payment card details with optional metadata`,

		RunE: func(cmd *cobra.Command, args []string) error {
			secret := domain.Secret{
				Info: domain.SecretInfo{
					Name:     name,
					Metadata: metadata,
					Type:     domain.PaymentCardSecretType,
				},
			}
			credentials := domain.PaymentCardSecret{
				Number: number,
			}
			marshaled, err := json.Marshal(credentials)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal card data")
				return fmt.Errorf("failed to marshal card data")
			}

			reader := bytes.NewReader(marshaled)

			if err = secretService.CreateSecret(ctx, secret, reader); err != nil {
				log.Error().Err(err).Msg("failed to create secret")
				return fmt.Errorf("failed to marshal card data")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name for the card (required)")
	cmd.Flags().StringVarP(&number, "number", "c", "", "Card number (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("number")

	return cmd
}

// newCreateTextSecretCmd creates a command for storing text secrets with interactive input.
func newCreateTextSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var name, metadata string

	cmd := &cobra.Command{
		Use:   "create-text",
		Short: "Store text content interactively",
		Long:  `Stream text content to be stored securely. Type 'end' on a new line to finish input.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			secret := domain.Secret{
				Info: domain.SecretInfo{
					Name:     name,
					Metadata: metadata,
					Type:     domain.TextSecretType,
				},
			}

			pr, pw := io.Pipe()

			go func() {
				defer pw.Close()
				scanner := bufio.NewScanner(os.Stdin)

				for scanner.Scan() {
					line := scanner.Text()
					if strings.TrimSpace(line) == "end" {
						break
					}
					if _, err := pw.Write([]byte(line + "\n")); err != nil {
						log.Error().Err(err).Msg("Error writing text")
						return
					}
				}
			}()

			if err := secretService.CreateSecret(ctx, secret, pr); err != nil {
				log.Error().Err(err).Msg("Failed to store text")
				return fmt.Errorf("failed to store text")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the text content (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newCreateFileSecretCmd creates a command for storing file secrets.
func newCreateFileSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var name, metadata, filePath string

	cmd := &cobra.Command{
		Use:   "create-file",
		Short: "Store a file securely",
		Long:  `Uploads and securely stores a file from the local filesystem`,

		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(filePath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to open file '%s'", filePath)
				return fmt.Errorf("failed to open file '%s'", filePath)
			}
			defer file.Close()

			secret := domain.Secret{
				Info: domain.SecretInfo{
					Name:     name,
					Metadata: metadata,
					Type:     domain.FileSecretType,
				},
			}

			if err = secretService.CreateSecret(ctx, secret, file); err != nil {
				log.Error().Err(err).Msgf("Failed to store file '%s' in secret storage", filePath)
				return fmt.Errorf("failed to store file '%s' in secret storage", filePath)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully stored file '%s' as '%s'\n", filePath, name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique name for the file (required)")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to file (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// newListSecretsCmd creates a command to list all stored secrets
func newListSecretsCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stored secrets",
		Long:  `Displays names and types of all stored secrets`,

		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := secretService.ListSecrets(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to list secrets")
				return fmt.Errorf("failed to list secrets")
			}

			if len(secrets) == 0 {
				fmt.Fprintln(os.Stdout, "No secrets found")
				return nil
			}

			fmt.Fprintln(os.Stdout, "Stored secrets:")
			for _, name := range secrets {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
			}
			return nil
		},
	}
}

// newGetCredentialsSecretCmd creates a command to retrieve credentials
func newGetCredentialsSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var (
		name    string
		version int32
	)

	cmd := &cobra.Command{
		Use:   "get-credentials",
		Short: "Retrieve stored credentials",

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				secret *domain.Secret
				err    error
			)

			if version == 0 {
				secret, err = secretService.GetLatestSecret(ctx, name)
			} else {
				secret, err = secretService.GetSecretByVersion(ctx, name, version)
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to retrieve credentials")
				return fmt.Errorf("failed to retrieve credentials")
			}

			var creds domain.CredentialsSecret
			err = json.Unmarshal([]byte(secret.Data), &creds)
			if err != nil {
				log.Error().Err(err).Msg("Failed to decode credentials")
				return fmt.Errorf("failed to decode credentials")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", secret.Info.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %d\n", secret.Info.Version)
			if secret.Info.Metadata != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Metadata: %s\n", secret.Info.Metadata)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Login: %s\n", creds.Login)
			fmt.Fprintf(cmd.OutOrStdout(), "Password: %s\n", creds.Password)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of credentials to retrieve (required)")
	cmd.Flags().Int32VarP(&version, "version", "v", 0, "Specific version to retrieve (default: latest)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newGetPaymentCardSecretCmd creates a command to retrieve payment card details
func newGetPaymentCardSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var (
		name    string
		version int32
	)

	cmd := &cobra.Command{
		Use:   "get-paymentcard",
		Short: "Retrieve stored payment card",

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				secret *domain.Secret
				err    error
			)

			if version == 0 {
				secret, err = secretService.GetLatestSecret(ctx, name)
			} else {
				secret, err = secretService.GetSecretByVersion(ctx, name, version)
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to retrieve card")
				return fmt.Errorf("failed to retrieve card")
			}

			var card domain.PaymentCardSecret
			err = json.Unmarshal([]byte(secret.Data), &card)
			if err != nil {
				log.Error().Err(err).Msg("Failed to decode card data")
				return fmt.Errorf("failed to decode card data")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", secret.Info.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %d\n", secret.Info.Version)
			if secret.Info.Metadata != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Metadata: %s\n", secret.Info.Metadata)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Card Number: %s\n", card.Number)

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of card to retrieve (required)")
	cmd.Flags().Int32VarP(&version, "version", "v", 0, "Specific version to retrieve (default: latest)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newGetTextSecretCmd retrieves and displays a text secret
func newGetTextSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var (
		name    string
		version int32
	)

	cmd := &cobra.Command{
		Use:   "get-text",
		Short: "Retrieve and display a text secret",

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				secretInfo *domain.SecretInfo
				reader     io.Reader
				err        error
			)

			if version == 0 {
				reader, secretInfo, err = secretService.GetLatestSecretStream(ctx, name)
			} else {
				reader, secretInfo, err = secretService.GetSecretStreamByVersion(ctx, name, version)
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to retrieve text")
				return fmt.Errorf("failed to retrieve text")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", secretInfo.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %d\n", secretInfo.Version)
			if secretInfo.Metadata != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Metadata: %s\n", secretInfo.Metadata)
			}
			fmt.Fprintln(os.Stdout, "\nContent:")

			if _, err = io.Copy(os.Stdout, reader); err != nil {
				log.Error().Err(err).Msg("Failed to read secret data")
				return fmt.Errorf("failed to read secret data")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the text to retrieve (required)")
	cmd.Flags().Int32VarP(&version, "version", "v", 0, "Specific version to retrieve (default: latest)")

	_ = cmd.MarkFlagRequired(name)

	return cmd
}

// newGetFileSecretCmd creates a command to download stored files
func newGetFileSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var (
		name    string
		version int32
	)

	cmd := &cobra.Command{
		Use:   "get-file-secret",
		Short: "Download a stored file",

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				reader     io.Reader
				secretInfo *domain.SecretInfo
				err        error
			)

			if version == 0 {
				reader, secretInfo, err = secretService.GetLatestSecretStream(ctx, name)
			} else {
				reader, secretInfo, err = secretService.GetSecretStreamByVersion(ctx, name, version)
			}
			if err != nil {
				log.Error().Err(err).Msg("Failed to retrieve file")
				return fmt.Errorf("failed to retrieve file")
			}

			outputFile, err := os.Create(secretInfo.Name)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create output file")
				return fmt.Errorf("failed to create output file")
			}
			defer outputFile.Close()

			if _, err = io.Copy(outputFile, reader); err != nil {
				log.Error().Err(err).Msgf("Failed to write secret data into output file '%s'", outputFile.Name())
				return fmt.Errorf("failed to write secret data into output file '%s'", outputFile.Name())
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully downloaded '%s' version %d\n",
				secretInfo.Name, secretInfo.Version)
			if secretInfo.Metadata != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Metadata: %s\n", secretInfo.Metadata)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of file to retrieve (required)")
	cmd.Flags().Int32VarP(&version, "version", "v", 0, "Specific version to retrieve (default: latest)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newDeleteSecretCmd creates a command to delete secrets
func newDeleteSecretCmd(ctx context.Context, secretService domain.SecretService) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a secret with all its versions",

		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete '%s'? (y/n): ", name)

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if strings.ToLower(scanner.Text()) != "y" {
				fmt.Fprintln(os.Stdout, "Deletion cancelled")
				return nil
			}

			err := secretService.DeleteSecret(ctx, name)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to delete secret '%s'", name)
				return fmt.Errorf("failed to delete secret '%s'", name)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of secret to delete (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
