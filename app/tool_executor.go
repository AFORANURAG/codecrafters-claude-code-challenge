package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/openai/openai-go/v3"
)

type ToolExecutor struct{}

func NewToolExecutor() ToolExecutor {
	return ToolExecutor{}
}

func (ToolExecutor) Execute(ctx context.Context, toolCall openai.ChatCompletionMessageToolCallUnion) string {
	switch toolCall.Function.Name {
	case ToolNameRead:
		return executeRead(toolCall.Function.Arguments)
	case ToolNameWrite:
		return executeWrite(toolCall.Function.Arguments)
	case ToolNameBash:
		return executeBash(ctx, toolCall.Function.Arguments)
	default:
		return fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)
	}
}

func executeRead(rawArgs string) string {
	var args struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return fmt.Sprintf("Invalid Read arguments: %v", err)
	}

	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}

	return string(content)
}

func executeWrite(rawArgs string) string {
	var args struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return fmt.Sprintf("Invalid Write arguments: %v", err)
	}

	if err := os.WriteFile(args.FilePath, []byte(args.Content), 0644); err != nil {
		return fmt.Sprintf("error while writing to file: %v", err)
	}

	return "Successfully wrote to file"
}

func executeBash(ctx context.Context, rawArgs string) string {
	var args struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return fmt.Sprintf("Invalid Bash arguments: %v", err)
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", args.Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Errored command is: %s\nerror while executing the bash command: %v\n%s", args.Command, err, string(out))
	}

	return string(out)
}
