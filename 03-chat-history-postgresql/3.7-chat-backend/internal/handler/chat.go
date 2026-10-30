package handler

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/repository"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/service"
)

type ChatHandler struct {
	service *service.ChatService
}

func NewChatHandler(s *service.ChatService) *ChatHandler {
	return &ChatHandler{service: s}
}

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 90*time.Second)
}

func parseID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid conversation id")
	}
	return id, nil
}

func handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrEmptyMessage),
		errors.Is(err, service.ErrEmptyTitle):
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, repository.ErrConversationNotFound):
		return c.Status(404).JSON(fiber.Map{
			"error": "Conversation not found",
		})

	case errors.Is(err, context.DeadlineExceeded):
		return c.Status(504).JSON(fiber.Map{
			"error": "Request timed out",
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

	ctx, cancel := requestContext()
	defer cancel()

	result, err := h.service.CreateConversation(ctx, body.Title)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(201).JSON(result)
}

func (h *ChatHandler) GetConversations(c *fiber.Ctx) error {
	ctx, cancel := requestContext()
	defer cancel()

	result, err := h.service.GetConversations(ctx)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := requestContext()
	defer cancel()

	result, err := h.service.GetMessages(ctx, id)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		Message string `json:"message"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	ctx, cancel := requestContext()
	defer cancel()

	result, err := h.service.SendMessage(ctx, id, body.Message)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(result)
}
