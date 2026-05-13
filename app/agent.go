package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared/constant"
)

type Agent struct {
	client       openai.Client
	model        string
	toolExecutor ToolExecutor
}

func NewAgent(client openai.Client, model string, toolExecutor ToolExecutor) Agent {
	return Agent{
		client:       client,
		model:        model,
		toolExecutor: toolExecutor,
	}
}

func (a Agent) Run(ctx context.Context, prompt string) (string, error) {
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
		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: messages,
			Tools:    ToolDefinitions(),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", errors.New("no choices in response")
		}

		fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

		message := resp.Choices[0].Message
		messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			Role: constant.Assistant("assistant"),
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(message.Content),
			},
			ToolCalls: message.ToParam().GetToolCalls(),
		}})

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		for _, toolCall := range message.ToolCalls {
			output := a.toolExecutor.Execute(ctx, toolCall)
			messages = append(messages, openai.ChatCompletionMessageParamUnion{OfTool: &openai.ChatCompletionToolMessageParam{
				Content: openai.ChatCompletionToolMessageParamContentUnion{
					OfString: openai.String(output),
				},
				ToolCallID: toolCall.ID,
				Role:       constant.Tool("tool"),
			}})
		}
	}
}
