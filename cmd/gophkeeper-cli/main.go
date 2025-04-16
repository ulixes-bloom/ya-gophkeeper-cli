package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/application"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/config"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/persistence"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/interfaces/grpc"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/presentation/cli"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	file, err := createLogFile()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}

	// Initialize logging with pretty console output
	output := zerolog.ConsoleWriter{Out: file, TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	// Load configuration
	conf, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	// Set up context with graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	// Configure logging level
	logLvl, err := zerolog.ParseLevel(conf.LogLvl)
	if err != nil {
		log.Warn().Err(err).Str("level", conf.LogLvl).Msg("Invalid log level, defaulting to info")
		logLvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLvl)
	log.Info().Str("level", logLvl.String()).Msg("Logging level configured")

	// Initialize token repository
	tokenRepo, err := persistence.NewTokenRepo()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize token repo")
	}

	// Create context with timeout for gRPC requests
	grpcCtx, grpcCancel := context.WithTimeout(ctx, conf.GRPCTimeout)
	defer grpcCancel()

	// Create credentials for grpc connection
	creds := insecure.NewCredentials()
	if conf.TLSCertPath != "" {
		pemData, err := os.ReadFile(conf.TLSCertPath)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed read tls cert file '%s'", conf.TLSCertPath)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pemData) {
			log.Fatal().Msg("Failed to add tls cert to application cert pool")
		}

		creds = credentials.NewTLS(&tls.Config{
			RootCAs: certPool,
		})
	}

	// Initialize gRPC clients
	authClient, err := grpc.NewAuthClient(conf.GRPCRunAddr, creds)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth client")
	}
	defer authClient.Close()

	secretClient, err := grpc.NewSecretClient(conf.GRPCRunAddr, tokenRepo, creds)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize secret client")
	}
	defer secretClient.Close()

	// Initialize services
	secretService := application.NewSecretService(secretClient)
	authService := application.NewAuthService(authClient, tokenRepo)

	// Initialize and run CLI
	rootCmd := cli.NewCLI(grpcCtx, secretService, authService)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Fatal cli error")
	}
}

func createLogFile() (*os.File, error) {
	path, err := getLogFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

func getLogFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %v", err)
	}
	return filepath.Join(homeDir, ".gophkeeper-cli", "gophkeeper-cli.log"), nil
}
