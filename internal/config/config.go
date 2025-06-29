package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Splunk SplunkConfig `yaml:"splunk"`
	Feeds  FeedConfig   `yaml:"feeds"`
	Alert  AlertConfig  `yaml:"alert"`
}

type SplunkConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type FeedConfig struct {
	Schedule string        `yaml:"schedule"` // e.g. "30m"
	Interval time.Duration `yaml:"-"`        // parsed value
}

type AlertConfig struct {
	SlackWebhook string `yaml:"slack_webhook"`
	EmailTo      string `yaml:"email_to,omitempty"`
}

// LoadConfig reads and parses a YAML config file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Parse string interval to time.Duration
	if dur, err := time.ParseDuration(cfg.Feeds.Schedule); err == nil {
		cfg.Feeds.Interval = dur
	}
	return &cfg, nil
}
