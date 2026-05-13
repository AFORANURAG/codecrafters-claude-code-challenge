package main

import "testing"

func TestLoadConfig(t *testing.T) {
	getenv := func(key string) string {
		values := map[string]string{
			"OPENROUTER_API_KEY": "test-key",
		}
		return values[key]
	}

	cfg, err := LoadConfig([]string{"-p", "hello"}, getenv)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Prompt != "hello" {
		t.Fatalf("Prompt = %q, want %q", cfg.Prompt, "hello")
	}
	if cfg.APIKey != "test-key" {
		t.Fatalf("APIKey = %q, want %q", cfg.APIKey, "test-key")
	}
	if cfg.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, defaultBaseURL)
	}
	if cfg.Model != defaultModel {
		t.Fatalf("Model = %q, want %q", cfg.Model, defaultModel)
	}
}

func TestLoadConfigRequiresPrompt(t *testing.T) {
	_, err := LoadConfig(nil, func(string) string {
		return "test-key"
	})
	if err == nil {
		t.Fatal("LoadConfig returned nil error, want prompt error")
	}
}

func TestLoadConfigRequiresAPIKey(t *testing.T) {
	_, err := LoadConfig([]string{"-p", "hello"}, func(string) string {
		return ""
	})
	if err == nil {
		t.Fatal("LoadConfig returned nil error, want API key error")
	}
}

func TestLoadConfigAllowsOverrides(t *testing.T) {
	getenv := func(key string) string {
		values := map[string]string{
			"OPENROUTER_API_KEY":  "test-key",
			"OPENROUTER_BASE_URL": "https://example.test/api",
			"OPENROUTER_MODEL":    "test-model",
		}
		return values[key]
	}

	cfg, err := LoadConfig([]string{"-p", "hello"}, getenv)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.BaseURL != "https://example.test/api" {
		t.Fatalf("BaseURL = %q, want override", cfg.BaseURL)
	}
	if cfg.Model != "test-model" {
		t.Fatalf("Model = %q, want override", cfg.Model)
	}
}
