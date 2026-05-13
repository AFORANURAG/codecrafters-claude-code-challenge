package main

import (
	"errors"
	"flag"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
	defaultModel   = "anthropic/claude-haiku-4.5"
)

type Config struct {
	Prompt  string
	APIKey  string
	BaseURL string
	Model   string
}

func LoadConfig(args []string, getenv func(string) string) (Config, error) {
	flags := flag.NewFlagSet("claude-code", flag.ContinueOnError)

	var cfg Config
	flags.StringVar(&cfg.Prompt, "p", "", "Prompt to send to LLM")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}

	if cfg.Prompt == "" {
		return Config{}, errors.New("prompt must not be empty")
	}

	cfg.APIKey = getenv("OPENROUTER_API_KEY")
	if cfg.APIKey == "" {
		return Config{}, errors.New("env variable OPENROUTER_API_KEY not found")
	}

	cfg.BaseURL = getenv("OPENROUTER_BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	cfg.Model = getenv("OPENROUTER_MODEL")
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}

	return cfg, nil
}
