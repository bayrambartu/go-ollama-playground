package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-ollama-playground/02-go-llm-integration/2.7-mini-project/internal/service"
)

type ChatHandler struct {
	service *service.ChatService
}

type ChatRequest struct {
	Message string `json:"message"`
}

func NewChatHandler(s *service.ChatService) *ChatHandler {
	return &ChatHandler{service: s}
}

func (h *ChatHandler) Chat(c *fiber.Ctx) error {
	var req ChatRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON request",
		})
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	result, err := h.service.Ask(ctx, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyMessage):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message cannot be empty",
			})

		case errors.Is(err, context.DeadlineExceeded):
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error": "LLM request timed out",
			})

		default:
			log.Println("Chat error:", err)

			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "LLM service unavailable",
			})
		}
	}

	return c.JSON(fiber.Map{
		"response":        result.Text,
		"prompt_token":    result.PromptTokens,
		"generated_token": result.GeneratedTokens,
	})
}
