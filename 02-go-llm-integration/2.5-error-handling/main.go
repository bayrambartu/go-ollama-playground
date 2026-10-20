package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

func askOllama(
	ctx context.Context,
	model string,
	prompt string,
) (string, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "", fmt.Errorf("client cannot be created: %w", err)
	}

	stream := false

	req := &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: &stream,
	}

	var answer strings.Builder

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		answer.WriteString(resp.Message.Content)
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("request of ollama failed: %w", err)
	}

	return answer.String(), nil
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	//cancel() --> request cancelled by user
	defer cancel()
	answer, err := askOllama(
		ctx,
		// "llama3.233:latest",
		"llama3.2:latest",
		"what is context in Go? explain it to just 1 sentence.",
	)

	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			log.Println("ERROR: request timed out. Please try again later.")

		case errors.Is(err, context.Canceled):
			log.Println("Error: request canceled by user.")

		default:
			log.Println("ERROR:", err)
		}
		return
	}

	fmt.Println("--- MODEL ANSWER ---")
	fmt.Println(answer)
}
