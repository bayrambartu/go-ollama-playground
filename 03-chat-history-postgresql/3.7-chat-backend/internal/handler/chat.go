package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/llm"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/repository"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/service"
)

type ChatHandler struct {
	service *service.ChatService
	timeout time.Duration
}

func NewChatHandler(
	s *service.ChatService,
	timeout time.Duration,
) *ChatHandler {
	return &ChatHandler{
		service: s,
		timeout: timeout,
	}
}

func (h *ChatHandler) requestContext() (
	context.Context,
	context.CancelFunc,
) {
	return context.WithTimeout(
		context.Background(),
		h.timeout,
	)
}

func handleError(c *fiber.Ctx, err error) error {
	log.Printf("HANDLER ERROR: %T | %v", err, err)

	switch {
	case errors.Is(err, service.ErrEmptyMessage),
		errors.Is(err, service.ErrEmptyTitle):
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})

	case errors.Is(err, repository.ErrConversationNotFound):
		return c.Status(404).JSON(fiber.Map{
			"error": "Conversation not found",
		})

	case errors.Is(err, llm.ErrModelUnavailable):
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Ollama model is unavailable",
		})

	case errors.Is(err, llm.ErrServiceUnavailable):
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Ollama service is unavailable",
		})

	case errors.Is(err, context.DeadlineExceeded):
		return c.Status(504).JSON(fiber.Map{
			"error": "LLM request timed out",
		})

	case errors.Is(err, service.ErrMissingIdempotencyKey):
		return c.Status(400).JSON(fiber.Map{
			"error": "Idempotency-Key header is required",
		})

	case errors.Is(err, service.ErrIdempotencyConflict):
		return c.Status(409).JSON(fiber.Map{
			"error": "Idempotency key belongs to another message",
		})

	case errors.Is(err, service.ErrConversationBusy):
		return c.Status(409).JSON(fiber.Map{
			"error": "Conversation is busy",
		})

	case errors.Is(err, service.ErrPreviousRequestFailed):
		return c.Status(409).JSON(fiber.Map{
			"error": "Previous request failed; use a new key to retry",
		})

	default:
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
}

func (h *ChatHandler) CreateConversation(c *fiber.Ctx) error {
	var body struct {
		Title string `json:"title"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	ctx, cancel := h.requestContext()
	defer cancel()

	result, err := h.service.CreateConversation(ctx, body.Title)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(201).JSON(result)
}

func (h *ChatHandler) GetConversations(c *fiber.Ctx) error {
	ctx, cancel := h.requestContext()
	defer cancel()

	result, err := h.service.GetConversations(ctx)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	id, err := h.parseID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := h.requestContext()
	defer cancel()

	result, err := h.service.GetMessages(ctx, id)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	id, err := h.parseID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var body struct {
		Message string `json:"message"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	key := c.Get("Idempotency-Key")

	ctx, cancel := h.requestContext()
	defer cancel()

	result, err := h.service.SendMessage(
		ctx,
		id,
		body.Message,
		key,
	)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}

func (h *ChatHandler) parseID(c *fiber.Ctx) (int64, error) {
	id, err := c.ParamsInt("id")
	if err != nil {
		return 0, errors.New("Invalid ID")
	}
	return int64(id), nil
}
