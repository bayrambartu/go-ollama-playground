package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	client *api.Client
	model  string
}

type Result struct {
	Text            string
	PromptTokens    int
	GeneratedTokens int
}

func NewOllamaClient(model string) (*OllamaClient, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

func (o *OllamaClient) Generate(
	ctx context.Context,
	prompt string,
) (Result, error) {
	stream := false

	req := &api.ChatRequest{
		Model: o.model,
		Messages: []api.Message{
			{
				Role: "system",
				Content: "You are a Go backend instructor. " +
					"Provide clear answers in Turkish.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: &stream,
	}

	var answer strings.Builder
	var result Result

	err := o.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		answer.WriteString(resp.Message.Content)

		if resp.Done {
			result.PromptTokens = resp.PromptEvalCount
			result.GeneratedTokens = resp.EvalCount
		}

		return nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("ollama chat failed: %w", err)
	}

	result.Text = answer.String()
	return result, nil
}
