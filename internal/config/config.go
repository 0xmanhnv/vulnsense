package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	AppEnv   string `mapstructure:"APP_ENV"`
	Log      LogConfig
	DB       DBConfig
	Cache    CacheConfig
	Feeds    FeedsConfig
	Splunk   SplunkConfig
	Slack    SlackConfig
	Telegram TelegramConfig
	Matcher  MatcherConfig
	Redis    RedisConfig
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// DBConfig holds database configuration.
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

	return &cfg, nil
}
