package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Feeds            []string `yaml:"feeds"`
	GmailAddress     string
	GmailAppPassword string
	RecipientEmail   string
}

func Load(configPath string) (*Config, error) {
	_ = godotenv.Load()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.GmailAddress = os.Getenv("GMAIL_ADDRESS")
	cfg.GmailAppPassword = os.Getenv("GMAIL_APP_PASSWORD")
	cfg.RecipientEmail = os.Getenv("RECIPIENT_EMAIL")

	if cfg.GmailAddress == "" {
		return nil, fmt.Errorf("GMAIL_ADDRESS environment variable is required")
	}
	if cfg.GmailAppPassword == "" {
		return nil, fmt.Errorf("GMAIL_APP_PASSWORD environment variable is required")
	}
	if cfg.RecipientEmail == "" {
		return nil, fmt.Errorf("RECIPIENT_EMAIL environment variable is required")
	}

	if len(cfg.Feeds) == 0 {
		return nil, fmt.Errorf("no feeds configured in config.yaml")
	}

	return &cfg, nil
}
