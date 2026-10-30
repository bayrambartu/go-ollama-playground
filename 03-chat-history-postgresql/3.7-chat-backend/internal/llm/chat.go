package llm

import (
	"context"
	"strings"

	"github.com/ollama/ollama/api"
)

type Client struct {
	client *api.Client
	model  string
}

type Result struct {
	Text            string
	PromptTokens    int
	GeneratedTokens int
}

func NewClient(model string) (*Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

func (c *Client) Generate(
	ctx context.Context,
	messages []api.Message,
) (Result, error) {
	stream := false

	req := &api.ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   &stream,
	}

	var answer strings.Builder
	var result Result

	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		answer.WriteString(resp.Message.Content)

		if resp.Done {
			result.PromptTokens = resp.PromptEvalCount
			result.GeneratedTokens = resp.EvalCount
		}

		return nil
	})
	if err != nil {
		return Result{}, err
	}

	result.Text = answer.String()
	return result, nil
}
