package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/ollama/ollama/api"
)

var (
	ErrModelUnavailable   = errors.New("ollama model unavailable")
	ErrServiceUnavailable = errors.New("ollama service unavailable")
)

func classifyError(err error) error {
	if err == nil {
		return nil
	}

	// 1. Timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("ollama timeout: %w", err)
	}

	// 2. Canceled request
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("ollama canceled: %w", err)
	}

	// 3. Ollama HTTP error
	var statusErr api.StatusError

	if errors.As(err, &statusErr) {
		if statusErr.StatusCode == 404 {
			return fmt.Errorf(
				"%w: %v",
				ErrModelUnavailable,
				err,
			)
		}
	}

	// 4. Network error
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return fmt.Errorf(
			"%w: %v",
			ErrServiceUnavailable,
			err,
		)
	}

	// 5. URL / connection error
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf(
			"%w: %v",
			ErrServiceUnavailable,
			err,
		)
	}

	return err
}

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
		return Result{}, classifyError(err)
	}

	result.Text = answer.String()
	return result, nil
}
