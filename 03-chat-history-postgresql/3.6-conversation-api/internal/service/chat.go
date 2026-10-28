package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ollama/ollama/api"

	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/llm"
	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/repository"
)

var ErrEmptyMessage = errors.New("message cannot be empty")
var ErrEmptyTitle = errors.New("title cannot be empty")

type ChatService struct {
	repo *repository.ChatRepository
	llm  *llm.Client
}

type ChatResult struct {
	Response       string `json:"response"`
	PromptToken    int    `json:"prompt_token"`
	GeneratedToken int    `json:"generated_token"`
	ConversationID int64  `json:"conversation_id"`
}

func NewChatService(
	repo *repository.ChatRepository,
	llmClient *llm.Client,
) *ChatService {
	return &ChatService{
		repo: repo,
		llm:  llmClient,
	}
}

func (s *ChatService) CreateConversation(
	ctx context.Context,
	title string,
) (repository.Conversation, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return repository.Conversation{}, ErrEmptyTitle
	}

	return s.repo.CreateConversation(ctx, title)
}

func (s *ChatService) GetConversations(
	ctx context.Context,
) ([]repository.Conversation, error) {
	return s.repo.GetConversations(ctx)
}

func (s *ChatService) GetMessages(
	ctx context.Context,
	id int64,
) ([]repository.Message, error) {
	return s.repo.GetMessages(ctx, id)
}

func (s *ChatService) SendMessage(
	ctx context.Context,
	conversationID int64,
	message string,
) (ChatResult, error) {
	message = strings.TrimSpace(message)

	if message == "" {
		return ChatResult{}, ErrEmptyMessage
	}

	history, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		return ChatResult{}, err
	}

	messages := []api.Message{
		{
			Role: "system",
			Content: "You are a Go backend instructor. " +
				"Give clear answers in Turkish.",
		},
	}

	for _, msg := range history {
		messages = append(messages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, api.Message{
		Role:    "user",
		Content: message,
	})

	// Save the user message.
	_, err = s.repo.SaveMessage(
		ctx,
		conversationID,
		"user",
		message,
	)
	if err != nil {
		return ChatResult{}, err
	}

	// Get a response from Ollama.
	result, err := s.llm.Generate(ctx, messages)
	if err != nil {
		return ChatResult{}, err
	}

	// Save the assistant response.
	_, err = s.repo.SaveMessage(
		ctx,
		conversationID,
		"assistant",
		result.Text,
	)
	if err != nil {
		return ChatResult{}, err
	}

	return ChatResult{
		Response:       result.Text,
		PromptToken:    result.PromptTokens,
		GeneratedToken: result.GeneratedTokens,
		ConversationID: conversationID,
	}, nil
}
