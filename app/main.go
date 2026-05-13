package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	cfg, err := LoadConfig(os.Args[1:], os.Getenv)
	if err != nil {
		panic(err)
	}

	client := openai.NewClient(option.WithAPIKey(cfg.APIKey), option.WithBaseURL(cfg.BaseURL))
	agent := NewAgent(client, cfg.Model, NewToolExecutor())

	output, err := agent.Run(context.Background(), cfg.Prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
