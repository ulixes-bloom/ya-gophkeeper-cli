package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
)

// newRegisterCmd creates a cobra command for user registration
// ctx: Context for request cancellation and timeouts
// authService: Authentication service interface
// Returns: Configured cobra.Command for registration
func newRegisterCmd(ctx context.Context, authService domain.AuthService) *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user account",
		Long:  "Register a new user account with the provided credentials",

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := authService.Register(ctx, login, password); err != nil {
				log.Error().Err(err).Msg("failed to register")

				switch {
				case errors.Is(err, domain.ErrLoginAlreayExists):
					return domain.ErrLoginAlreayExists
				default:
					return fmt.Errorf("failed to register")
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Successfully registered user")
			return nil
		},
	}

	cmd.Flags().StringVarP(&login, "login", "l", "", "User login (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "User password (required)")

	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// newLoginCmd creates a cobra command for user authentication
// ctx: Context for request cancellation and timeouts
// authService: Authentication service interface
// Returns: Configured cobra.Command for login
func newLoginCmd(ctx context.Context, authService domain.AuthService) *cobra.Command {
	var login, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your account",
		Long:  `Authenticate with your existing account credentials`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := authService.Login(ctx, login, password); err != nil {
				log.Error().Err(err).Msg("failed to log in")

				switch {
				case errors.Is(err, domain.ErrInvalidCredentials):
					return domain.ErrInvalidCredentials
				case errors.Is(err, domain.ErrUserNotFound):
					return domain.ErrUserNotFound
				default:
					return fmt.Errorf("failed to log in")
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Successfully logged in")
			return nil
		},
	}

	cmd.Flags().StringVarP(&login, "login", "l", "", "Login")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")

	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
