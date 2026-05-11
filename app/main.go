package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"

	// "log"
	"os"

	// "github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
)

func main() {
	// err := godotenv.Load() // Loads .env by default
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}
	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: messages,
				Tools: []openai.ChatCompletionToolUnionParam{
					{
						OfFunction: &openai.ChatCompletionFunctionToolParam{
							Type: constant.Function("function"),
							Function: shared.FunctionDefinitionParam{
								Name:        "Read",
								Description: openai.String("Read and return the contents of a file"),
								Strict:      openai.Bool(true),
								Parameters: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"file_path": map[string]interface{}{
											"type":        "string",
											"description": "The path to the file to read",
										},
									},
									"required": []string{"file_path"},
								},
							},
						},
					},
					{
						OfFunction: &openai.ChatCompletionFunctionToolParam{
							Type: constant.Function("function"),
							Function: shared.FunctionDefinitionParam{
								Name:        "Write",
								Description: openai.String("Write content to a file"),
								Strict:      openai.Bool(true),
								Parameters: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"file_path": map[string]interface{}{
											"type":        "string",
											"description": "The path of the file to write to",
										},
										"content": map[string]interface{}{
											"type":        "string",
											"description": "The content to write to the file",
										},
									},
									"required": []string{"file_path", "content"},
								},
							},
						},
					},
					{
						OfFunction: &openai.ChatCompletionFunctionToolParam{
							Type: constant.Function("function"),
							Function: shared.FunctionDefinitionParam{
								Name:        "Bash",
								Description: openai.String("Execute a shell command"),
								Strict:      openai.Bool(true),
								Parameters: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"command": map[string]interface{}{
											"type":        "string",
											"description": "The command to execute",
										},
									},
									"required": []string{"command"},
								},
							},
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		// You can use print statements as follows for debugging, they'll be visible when running tests.
		fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

		messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			Role: constant.Assistant("assistant"),
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(resp.Choices[0].Message.Content),
			},
			ToolCalls: resp.Choices[0].Message.ToParam().GetToolCalls(),
		}})

		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			fmt.Print(resp.Choices[0].Message.Content)
			return
		}

		// we have tool calls
		toolCalls := resp.Choices[0].Message.ToolCalls
		for _, toolCall := range toolCalls {
			if toolCall.Function.Name == "Read" {
				// we need to read using golang filesystem apis
				var args struct {
					FilePath string `json:"file_path"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				content, err := os.ReadFile(args.FilePath)

				if err != nil {
					_ = fmt.Errorf("error: %v", err)

				}
				messages = append(messages, openai.ChatCompletionMessageParamUnion{OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(string(content)),
					},
					ToolCallID: toolCall.ID,
					Role:       constant.Tool("tool"),
				}})
			}
			if toolCall.Function.Name == "Write" {
				// we need to read using golang filesystem apis
				var args struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

				err := os.WriteFile(args.FilePath, []byte(args.Content), 0644)

				if err != nil {
					_ = fmt.Errorf("error while writing to file: %v", err)
				}

				messages = append(messages, openai.ChatCompletionMessageParamUnion{OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String("Successfully wrote to file"),
					},
					ToolCallID: toolCall.ID,
					Role:       constant.Tool("tool"),
				}})
			}
			if toolCall.Function.Name == "Bash" {
				// we need to read using golang filesystem apis
				var args struct {
					Command string `json:"command"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				cmd := exec.Command(args.Command)

				out, err := cmd.Output()

				if err != nil {
					_ = fmt.Errorf("Errored command is :%v", args.Command)
					_ = fmt.Errorf("error while executing the bash command: %v", err)
				}

				messages = append(messages, openai.ChatCompletionMessageParamUnion{OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(string(out)),
					},
					ToolCallID: toolCall.ID,
					Role:       constant.Tool("tool"),
				}})
			}
		}

	}

	// When the LLM requests a Read tool call, the output matches the exact file contents
	// When the LLM does not request a tool call, the output is the LLM's text response
	// Your program exits with code 0

}
