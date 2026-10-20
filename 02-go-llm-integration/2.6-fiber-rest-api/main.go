package main

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ollama/ollama/api"
)

const model = "llama3.2:latest"

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Response       string `json:"response"`
	Model          string `json:"model"`
	PromptToken    int    `json:"prompt_token"`
	GeneratedToken int    `json:"generated_token"`
}

func main() {
	app := fiber.New()

	ollamaClient, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/chat", func(c *fiber.Ctx) error {
		var body ChatRequest

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON request",
			})
		}

		if strings.TrimSpace(body.Message) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message cannot be empty",
			})
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			60*time.Second,
		)
		defer cancel()

		stream := false

		req := &api.ChatRequest{
			Model: model,
			Messages: []api.Message{
				{
					Role:    "system",
					Content: "you are a Go programming expert. Answers should be concise and in Turkish",
				},
				{
					Role:    "user",
					Content: body.Message,
				},
			},
			Stream: &stream,
		}

		var answer strings.Builder

		generatedTokens := 0
		promptTokens := 0
		err := ollamaClient.Chat(ctx, req, func(resp api.ChatResponse) error {
			answer.WriteString(resp.Message.Content)

			if resp.Done {
				log.Printf(
					"LLM request completed. Prompt tokens: %d, Generated tokens: %d",
					resp.PromptEvalCount,
					resp.EvalCount,
				)
			}
			promptTokens += resp.PromptEvalCount
			generatedTokens += resp.EvalCount
			return nil
		})

		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
					"error": "LLM request timed out",
				})

			case errors.Is(err, context.Canceled):
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": "LLM request canceled",
				})

			default:
				log.Println("Ollama error:", err)

				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
					"error": "LLM service unavailable",
				})
			}
		}

		return c.JSON(ChatResponse{
			Response:       answer.String(),
			Model:          model,
			PromptToken:    promptTokens,
			GeneratedToken: generatedTokens,
		})
	})

	log.Fatal(app.Listen(":8080"))
}
