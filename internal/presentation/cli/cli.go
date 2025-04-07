package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "HEAD"
)

func NewCLI(ctx context.Context, secretService domain.SecretService, authService domain.AuthService) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper-cli",
		Short: "GophKeeper - Secure secret management CLI",
		Long: `GophKeeper is a command-line tool for securely storing and managing secrets.
		
Features:
- Store credentials, payment cards, text notes and files
- Encrypted storage with versioning
- Cross-platform secret synchronization

Use 'gophkeeper <command> --help' for detailed usage of each command.`,
		Version:      Version,
		SilenceUsage: true,
	}

	// Add all secret management commands
	rootCmd.AddCommand(newCreateCredentialsSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newCreatePaymentCardSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newCreateTextSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newCreateFileSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newListSecretsCmd(ctx, secretService))
	rootCmd.AddCommand(newGetCredentialsSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newGetPaymentCardSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newGetTextSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newGetFileSecretCmd(ctx, secretService))
	rootCmd.AddCommand(newDeleteSecretCmd(ctx, secretService))

	// Add authentication commands
	rootCmd.AddCommand(newRegisterCmd(ctx, authService))
	rootCmd.AddCommand(newLoginCmd(ctx, authService))

	return rootCmd
}
