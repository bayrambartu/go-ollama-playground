package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ollama/ollama/api"

	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/llm"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/repository"
)

var ErrEmptyMessage = errors.New("message cannot be empty")
var ErrEmptyTitle = errors.New("title cannot be empty")

var (
	ErrMissingIdempotencyKey = errors.New("idempotency key is required")
	ErrIdempotencyConflict   = errors.New("idempotency key used with different message")
	ErrConversationBusy      = errors.New("conversation is processing another message")
	ErrPreviousRequestFailed = errors.New("previous request failed")
)

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
	key string,
) (ChatResult, error) {
	message = strings.TrimSpace(message)
	key = strings.TrimSpace(key)

	if message == "" {
		return ChatResult{}, ErrEmptyMessage
	}
	if key == "" {
		return ChatResult{}, ErrMissingIdempotencyKey
	}

	// Allow one active operation per conversation.
	unlock, locked, err := s.repo.TryLockConversation(
		ctx, conversationID,
	)
	if err != nil {
		return ChatResult{}, err
	}
	if !locked {
		return ChatResult{}, ErrConversationBusy
	}
	defer unlock()

	// Check idempotency after acquiring the lock.
	existing, found, err := s.repo.FindRequest(
		ctx, conversationID, key,
	)
	if err != nil {
		return ChatResult{}, err
	}

	if found {
		originalMessage, err := s.repo.GetMessageContent(
			ctx, existing.UserMessageID,
		)
		if err != nil {
			return ChatResult{}, err
		}

		if originalMessage != message {
			return ChatResult{}, ErrIdempotencyConflict
		}

		switch existing.Status {
		case "completed":
			if existing.AssistantMessageID == nil {
				return ChatResult{}, fmt.Errorf(
					"completed request missing assistant message",
				)
			}

			answer, err := s.repo.GetMessageContent(
				ctx, *existing.AssistantMessageID,
			)
			if err != nil {
				return ChatResult{}, err
			}

			return ChatResult{
				Response:       answer,
				ConversationID: conversationID,
			}, nil

		case "failed":
			return ChatResult{}, ErrPreviousRequestFailed

		default:
			return ChatResult{}, ErrConversationBusy
		}
	}

	// Read the history first to avoid adding the new message twice.
	history, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		return ChatResult{}, err
	}

	llmMessages := BuildLLMMessages(history, message)

	// Save the user message and pending request atomically.
	req, err := s.repo.CreatePendingRequest(
		ctx, conversationID, key, message,
	)
	if err != nil {
		return ChatResult{}, err
	}

	// Keep the database transaction closed during the Ollama call.
	result, err := s.llm.Generate(ctx, llmMessages)
	if err != nil {
		s.markFailed(req.ID, err.Error())
		return ChatResult{}, err
	}

	if strings.TrimSpace(result.Text) == "" {
		err := fmt.Errorf("ollama returned empty response")
		s.markFailed(req.ID, err.Error())
		return ChatResult{}, err
	}

	// Save the assistant message and completed status in one transaction.
	err = s.repo.CompleteRequest(
		ctx, req.ID, conversationID, result.Text,
	)
	if err != nil {
		s.markFailed(req.ID, err.Error())
		return ChatResult{}, err
	}

	return ChatResult{
		Response:       result.Text,
		PromptToken:    result.PromptTokens,
		GeneratedToken: result.GeneratedTokens,
		ConversationID: conversationID,
	}, nil
}

func BuildLLMMessages(
	history []repository.Message,
	newMessage string,
) []api.Message {
	messages := []api.Message{
		{
			Role: "system",
			Content: "You are a Go backend instructor. " +
				"Give short and technically correct answers only in Turkish.",
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
		Content: newMessage,
	})

	return messages
}

func (s *ChatService) markFailed(
	requestID int64,
	reason string,
) {
	// The request may have timed out.
	// Save the failure status with a separate, short context.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := s.repo.FailRequest(
		ctx, requestID, reason,
	); err != nil {
		log.Printf(
			"failed to update request %d: %v",
			requestID, err,
		)
	}
}
