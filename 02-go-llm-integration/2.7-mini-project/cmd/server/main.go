package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"go-ollama-playground/02-go-llm-integration/2.7-mini-project/internal/handler"
	"go-ollama-playground/02-go-llm-integration/2.7-mini-project/internal/llm"
	"go-ollama-playground/02-go-llm-integration/2.7-mini-project/internal/service"
)

func main() {
	app := fiber.New()

	// 1. LLM Client
	ollamaClient, err := llm.NewOllamaClient("llama3.2:latest")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Service
	chatService := service.NewChatService(ollamaClient)

	// 3. Handler
	chatHandler := handler.NewChatHandler(chatService)

	// 4. Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/chat", chatHandler.Chat)

	log.Println("LLM Backend running on :8080")

	log.Fatal(app.Listen(":8080"))
}
