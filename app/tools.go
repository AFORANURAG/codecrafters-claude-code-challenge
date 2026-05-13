package main

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
)

const (
	ToolNameRead  = "Read"
	ToolNameWrite = "Write"
	ToolNameBash  = "Bash"
)

func ToolDefinitions() []openai.ChatCompletionToolUnionParam {
	return []openai.ChatCompletionToolUnionParam{
		functionTool(
			ToolNameRead,
			"Read and return the contents of a file",
			map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to read",
				},
			},
			[]string{"file_path"},
		),
		functionTool(
			ToolNameWrite,
			"Write content to a file",
			map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "The path of the file to write to",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to write to the file",
				},
			},
			[]string{"file_path", "content"},
		),
		functionTool(
			ToolNameBash,
			"Execute a shell command",
			map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The command to execute",
				},
			},
			[]string{"command"},
		),
	}
}

func functionTool(name string, description string, properties map[string]interface{}, required []string) openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Type: constant.Function("function"),
			Function: shared.FunctionDefinitionParam{
				Name:        name,
				Description: openai.String(description),
				Strict:      openai.Bool(true),
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		},
	}
}
