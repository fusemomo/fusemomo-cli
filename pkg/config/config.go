// Package config loads and validates the Fusemomo CLI configuration.
// Priority: CLI flag > environment variable > config file > default.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ConfigDir is the directory for the fusemomo config file.
	ConfigDir = ".fusemomo"
	// ConfigFile is the config filename (without extension).
	ConfigFile = "config"
	// ConfigType is the config file format.
	ConfigType = "yaml"
)

// Config holds all resolved CLI configuration values.
type Config struct {
	APIKey  string
	APIURL  string
	Timeout int
	Output  string
	Debug   bool
}

// Load initialises Viper from environment variables, the config file (if present),
// and binds the provided Cobra flags so that flag > env > file > default resolution
// works automatically.
func Load(rootCmd *cobra.Command) (*Config, error) {
	// Bind env vars.
	viper.SetEnvPrefix("FUSEMOMO")
	viper.AutomaticEnv()

	// Explicit env bindings (needed because key names differ from env names).
	_ = viper.BindEnv("api_key", "FUSEMOMO_API_KEY")
	_ = viper.BindEnv("api_url", "FUSEMOMO_API_URL")
	_ = viper.BindEnv("timeout", "FUSEMOMO_TIMEOUT")
	_ = viper.BindEnv("output", "FUSEMOMO_OUTPUT")
	_ = viper.BindEnv("debug", "FUSEMOMO_DEBUG")

	// Bind Cobra flags (overrides env when explicitly set).
	if f := rootCmd.Flags().Lookup("api-key"); f != nil {
		_ = viper.BindPFlag("api_key", f)
	}
	if f := rootCmd.Flags().Lookup("api-url"); f != nil {
		_ = viper.BindPFlag("api_url", f)
	}
	if f := rootCmd.Flags().Lookup("timeout"); f != nil {
		_ = viper.BindPFlag("timeout", f)
	}
	if f := rootCmd.Flags().Lookup("output"); f != nil {
		_ = viper.BindPFlag("output", f)
	}
	if f := rootCmd.Flags().Lookup("debug"); f != nil {
		_ = viper.BindPFlag("debug", f)
	}

	// Default values.
	viper.SetDefault("api_url", "https://api.fusemomo.com")
	viper.SetDefault("timeout", 30)
	viper.SetDefault("output", "json")
	viper.SetDefault("debug", false)

	// Load config file if it exists — do NOT create it automatically.
	home, err := os.UserHomeDir()
	if err == nil {
		cfgPath := filepath.Join(home, ConfigDir)
		viper.AddConfigPath(cfgPath)
		viper.SetConfigName(ConfigFile)
		viper.SetConfigType(ConfigType)
		// Ignore file-not-found; any other error propagates.
		if err := viper.ReadInConfig(); err != nil {
			if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	cfg := &Config{
		APIKey:  viper.GetString("api_key"),
		APIURL:  viper.GetString("api_url"),
		Timeout: viper.GetInt("timeout"),
		Output:  viper.GetString("output"),
		Debug:   viper.GetBool("debug"),
	}

	return cfg, nil
}

// Validate checks that the config is valid for making API calls.
// Returns an error with the correct exit-code hint embedded.
// Skipped by: setup, version, completion.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return &ValidationError{
			ExitCode: 3,
			Message:  "API key is not set. Run `fusemomo setup` or set FUSEMOMO_API_KEY.",
		}
	}
	if !strings.HasPrefix(c.APIKey, "fm_live_") && !strings.HasPrefix(c.APIKey, "fm_test_") {
		return &ValidationError{
			ExitCode: 3,
			Message:  "API key has an invalid format. It must start with fm_live_ or fm_test_.",
		}
	}
	if c.APIURL != "" {
		if _, err := url.ParseRequestURI(c.APIURL); err != nil {
			return &ValidationError{
				ExitCode: 3,
				Message:  fmt.Sprintf("api_url is not a valid URL: %s", c.APIURL),
			}
		}
	}
	if c.Timeout <= 0 {
		return &ValidationError{
			ExitCode: 3,
			Message:  "timeout must be a positive integer.",
		}
	}
	return nil
}

// ConfigFilePath returns the expected path to the config file.
func ConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ConfigDir, ConfigFile+"."+ConfigType), nil
}

// ConfigDirPath returns the expected path to the config directory.
func ConfigDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ConfigDir), nil
}

// ValidationError is returned when config validation fails.
// ExitCode encodes the CLI exit code (3 for config/auth errors).
type ValidationError struct {
	ExitCode int
	Message  string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsHTTP returns true if the URL scheme is http (non-TLS).
func (c *Config) IsHTTP() bool {
	return strings.HasPrefix(c.APIURL, "http://")
}
