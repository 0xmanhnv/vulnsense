package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// RSSSource represents a single RSS feed source
type RSSSource struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Type    string `yaml:"type"`   // blog, advisory, vulnerability, etc.
	Source  string `yaml:"source"` // source name
	Enabled bool   `yaml:"enabled"`
}

// DatabaseConfig holds configuration for any database type (new multi-database support)
type DatabaseConfig struct {
	// Database type
	Type string `mapstructure:"type" yaml:"type"`

	// Connection details
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name" yaml:"name"`

	// PostgreSQL specific
	SSLMode string `mapstructure:"sslmode" yaml:"sslmode"`

	// Connection pool settings
	MaxConnections int           `mapstructure:"max_connections" yaml:"max_connections"`
	MinConnections int           `mapstructure:"min_connections" yaml:"min_connections"`
	MaxLifetime    time.Duration `mapstructure:"max_lifetime" yaml:"max_lifetime"`
	MaxIdleTime    time.Duration `mapstructure:"max_idle_time" yaml:"max_idle_time"`

	// Additional options
	Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// Config holds all configuration for the application.
type Config struct {
	AppEnv     string `mapstructure:"APP_ENV"`
	Log        LogConfig
	DB         DBConfig
	Cache      CacheConfig
	Feeds      FeedsConfig
	Splunk     SplunkConfig
	Slack      SlackConfig
	Telegram   TelegramConfig
	Matcher    MatcherConfig
	Redis      RedisConfig
	RSSSources []RSSSource `mapstructure:"rss_sources"`

	// New multi-database support
	Database  DatabaseConfig            `mapstructure:"database" yaml:"database"`
	Databases map[string]DatabaseConfig `mapstructure:"databases" yaml:"databases"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// DBConfig holds database configuration (existing, for backward compatibility).
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// CacheConfig holds Redis cache configuration.
type CacheConfig struct {
	Host string `mapstructure:"CACHE_HOST"`
	Port int    `mapstructure:"CACHE_PORT"`
}

// FeedsConfig holds feed fetching configuration.
type FeedsConfig struct {
	Schedule string `mapstructure:"schedule"`
}

// RedisConfig holds all the configuration for the Redis connection.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SplunkConfig holds Splunk client configuration.
type SplunkConfig struct {
	Host     string `mapstructure:"host"`
	Token    string `mapstructure:"token"`
	Insecure bool   `mapstructure:"insecure"`
}

// SlackConfig holds Slack webhook configuration.
type SlackConfig struct {
	WebhookURL string `mapstructure:"SLACK_WEBHOOK_URL"`
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	BotToken string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	ChatID   int64  `mapstructure:"TELEGRAM_CHAT_ID"`
}

// MatcherConfig holds vulnerability matching configuration.
type MatcherConfig struct {
	FuzzyScoreThreshold float64 `mapstructure:"fuzzy_score_threshold"`
}

// NewConfig initializes and returns a new configuration object.
// It reads from a YAML file and overrides with environment variables.
func NewConfig() (*Config, error) {
	var cfg Config

	// 1. Set up viper to read from YAML file
	viper.AddConfigPath("./configs")
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	// 2. Set up viper to read from environment variables
	viper.AutomaticEnv()
	// This allows viper to read DB_HOST as db.host
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Don't fail if the config file is not found, env vars can still be used
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal the configuration into the struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Load RSS sources from sources.yaml
	if err := cfg.loadRSSSources(); err != nil {
		return nil, fmt.Errorf("failed to load RSS sources: %w", err)
	}

	return &cfg, nil
}

// loadRSSSources loads RSS sources from the sources.yaml file
func (c *Config) loadRSSSources() error {
	sourcesFile := "./configs/sources.yaml"

	// Check if file exists
	if _, err := os.Stat(sourcesFile); os.IsNotExist(err) {
		// File doesn't exist, return empty sources
		c.RSSSources = []RSSSource{}
		return nil
	}

	// Read the file
	data, err := os.ReadFile(sourcesFile)
	if err != nil {
		return fmt.Errorf("failed to read sources file: %w", err)
	}

	// Parse YAML
	var sources []RSSSource
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return fmt.Errorf("failed to parse sources YAML: %w", err)
	}

	c.RSSSources = sources
	return nil
}
