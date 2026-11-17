package handler

import (
	"semantic-search-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

type RAGHandler struct {
	service *service.RAGService
}

func NewRAGHandler(s *service.RAGService) *RAGHandler {
	return &RAGHandler{service: s}
}

func (h *RAGHandler) Ask(c *fiber.Ctx) error {
	var req service.AskRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid JSON",
		})
	}

	ctx, cancel := requestContext()
	defer cancel()

	response, err := h.service.Ask(ctx, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(response)
}
