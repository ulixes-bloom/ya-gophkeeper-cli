package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// Config holds application configuration parameters.
type Config struct {
	GRPCRunAddr string        `env:"GRPC_RUN_ADDRESS"` // gRPC server address
	LogLvl      string        `env:"LOGLVL"`           // Logging level (Debug, Info, Warn, Error)
	GRPCTimeout time.Duration `env:"GRPC_TIMEOUT"`
	TLSCertPath string        `env:"TLS_CERT_PATH"`
}

// Parse loads configuration from environment variables with fallback to defaults.
// Returns the configuration or an error if environment loading fails.
func Parse() (*Config, error) {
	// Try to load .env file if it exists (optional)
	_ = godotenv.Load()

	conf := GetDefault()

	if err := env.Parse(conf); err != nil {
		return nil, fmt.Errorf("config.Parse: %w", err)
	}

	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("config.Parse: %w", err)
	}

	return conf, nil
}

// GetDefault returns the default configuration values.
func GetDefault() (conf *Config) {
	return &Config{
		GRPCRunAddr: ":8097",
		LogLvl:      "Info",
		GRPCTimeout: 30 * time.Second,
		TLSCertPath: "",
	}
}

// normalizeLogLevel standardizes the log level string format.
func normalizeLogLevel(level string) string {
	return strings.ToLower(strings.TrimSpace(level))
}

// validateConfig checks if the configuration values are valid.
func validateConfig(c *Config) error {
	if c.GRPCRunAddr == "" {
		return errors.New("GRPC_RUN_ADDRESS cannot be empty")
	}

	switch normalizeLogLevel(c.LogLvl) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLvl)
	}
}
