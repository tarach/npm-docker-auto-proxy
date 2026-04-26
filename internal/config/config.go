package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	NPMBaseURL string
	NPMEmail   string
	NPMPassword string
	LogLevel   string
}

func Load() (Config, error) {
	cfg := Config{
		NPMBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("NPM_BASE_URL")), "/"),
		NPMEmail: strings.TrimSpace(os.Getenv("NPM_EMAIL")),
		NPMPassword: os.Getenv("NPM_PASSWORD"),
		LogLevel: strings.TrimSpace(os.Getenv("LOG_LEVEL")),
	}

	if cfg.NPMBaseURL == "" {
		return Config{}, errors.New("NPM_BASE_URL is required")
	}

	if cfg.NPMEmail == "" {
		return Config{}, errors.New("NPM_EMAIL is required")
	}

	if cfg.NPMPassword == "" {
		return Config{}, errors.New("NPM_PASSWORD is required")
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return cfg, nil
}
