package handler

import (
	"context"
	"errors"
	"time"

	"semantic-search-api/internal/model"
	"semantic-search-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	service *service.DocumentService
}

func NewDocumentHandler(
	s *service.DocumentService,
) *DocumentHandler {
	return &DocumentHandler{service: s}
}

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.Background(),
		90*time.Second,
	)
}

func (h *DocumentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateDocumentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid JSON",
		})
	}

	ctx, cancel := requestContext()
	defer cancel()

	doc, err := h.service.Create(ctx, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(201).JSON(doc)
}

func (h *DocumentHandler) List(c *fiber.Ctx) error {
	ctx, cancel := requestContext()
	defer cancel()

	docs, err := h.service.List(ctx)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(fiber.Map{
		"documents": docs,
	})
}

func (h *DocumentHandler) Search(c *fiber.Ctx) error {
	var req model.SearchRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid JSON",
		})
	}

	ctx, cancel := requestContext()
	defer cancel()

	results, err := h.service.Search(ctx, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(fiber.Map{
		"query":   req.Query,
		"results": results,
	})
}

func respondError(c *fiber.Ctx, err error) error {
	if errors.Is(err, service.ErrInvalidInput) {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return c.Status(504).JSON(fiber.Map{
			"error": "request timeout",
		})
	}

	return c.Status(500).JSON(fiber.Map{
		"error": "internal server error",
	})
}
