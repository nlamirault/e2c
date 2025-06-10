package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"github.com/nlamirault/e2c/internal/featureflags"
)

// Config represents the application configuration
type Config struct {
	AWS         AWSConfig                   `mapstructure:"aws"`
	UI          UIConfig                    `mapstructure:"ui"`
	FeatureFlags featureflags.FeatureFlagsConfig `mapstructure:"feature_flags"`
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	DefaultRegion   string        `mapstructure:"default_region"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
	Profile         string        `mapstructure:"profile"`
}

// UIConfig holds UI-specific configuration
type UIConfig struct {
	Compact bool `mapstructure:"compact"`
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig(log *slog.Logger) (*Config, error) {
	// Set defaults
	viper.SetDefault("aws.default_region", "us-west-1")
	viper.SetDefault("aws.refresh_interval", "30s")
	viper.SetDefault("aws.profile", "")
	viper.SetDefault("ui.compact", false)
	viper.SetDefault("feature_flags.enabled", false)
	viper.SetDefault("feature_flags.provider", "configcat")
	viper.SetDefault("feature_flags.configcat.sdk_key", "")
	viper.SetDefault("feature_flags.configcat.environment", "")
	viper.SetDefault("feature_flags.configcat.base_url", "")
	viper.SetDefault("feature_flags.configcat.cache_ttl_seconds", 60)
	viper.SetDefault("feature_flags.configcat.polling_interval_seconds", 60)
	viper.SetDefault("feature_flags.env.prefix", "E2C_FEATURE_")
	viper.SetDefault("feature_flags.env.case_sensitive", false)


	// Config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config search paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warn("Could not determine user home directory", "error", err)
	} else {
		configDir := filepath.Join(homeDir, ".config", "e2c")
		viper.AddConfigPath(configDir)
	}

	// Also look in current directory
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("E2C")
	viper.AutomaticEnv()

	// Try to read config file
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("No config file found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	} else {
		log.Info("Using config file", "file", viper.ConfigFileUsed())
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}
	
	// Initialize feature flags if enabled
	if config.FeatureFlags.Enabled {
		if err := featureflags.InitializeClient(log, config.FeatureFlags); err != nil {
			log.Warn("Failed to initialize feature flags client", "error", err)
		}
	}

	return &config, nil
}

// Override allows command-line flags to override config
func (c *Config) Override(profile, region string) {
	if profile != "" {
		c.AWS.Profile = profile
	}
	if region != "" {
		c.AWS.DefaultRegion = region
	}
}

// OverrideFeatureFlags allows command-line flags to override feature flags config
func (c *Config) OverrideFeatureFlags(provider string) {
	if provider != "" {
		c.FeatureFlags.Provider = featureflags.ProviderType(provider)
		// When switching to environment provider, set default prefix if not already set
		if c.FeatureFlags.Provider == featureflags.EnvProvider && c.FeatureFlags.Env.Prefix == "" {
			c.FeatureFlags.Env.Prefix = "E2C_FEATURE_"
		}
	}
}