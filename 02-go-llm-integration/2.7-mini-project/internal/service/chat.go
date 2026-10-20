package service

import (
	"context"
	"errors"
	"strings"

	"go-ollama-playground/02-go-llm-integration/2.7-mini-project/internal/llm"
)

var ErrEmptyMessage = errors.New("message cannot be empty")

type ChatService struct {
	llmClient *llm.OllamaClient
}

func NewChatService(client *llm.OllamaClient) *ChatService {
	return &ChatService{
		llmClient: client,
	}
}

func (s *ChatService) Ask(
	ctx context.Context,
	message string,
) (llm.Result, error) {
	message = strings.TrimSpace(message)

	if message == "" {
		return llm.Result{}, ErrEmptyMessage
	}

	return s.llmClient.Generate(ctx, message)
}
