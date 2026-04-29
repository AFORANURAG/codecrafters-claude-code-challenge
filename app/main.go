package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"
)

func main() {
	err := godotenv.Load() // Loads .env by default
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: "anthropic/claude-haiku-4.5",
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
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

	// TODO: Uncomment the line below to pass the first stage
	fmt.Print(resp.Choices[0])
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
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
				fmt.Println(string(content))

			}
		}
	}

	// When the LLM requests a Read tool call, the output matches the exact file contents
	// When the LLM does not request a tool call, the output is the LLM's text response
	// Your program exits with code 0

}
