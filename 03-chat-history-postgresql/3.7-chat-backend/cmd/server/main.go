package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/config"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/handler"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/llm"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/repository"
	"go-ollama-playground/03-chat-history-postgresql/3.7-chat-backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error: ", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("Database error: ", err)
	}

	// The Ollama SDK uses the environment variable.
	if err := os.Setenv("OLLAMA_HOST", cfg.OllamaHost); err != nil {
		log.Fatal(err)
	}

	ollamaClient, err := llm.NewClient(cfg.OllamaModel)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewChatRepository(db)
	chatService := service.NewChatService(repo, ollamaClient)
	chatHandler := handler.NewChatHandler(chatService)

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/conversations", chatHandler.CreateConversation)
	app.Get("/conversations", chatHandler.GetConversations)
	app.Get("/conversations/:id/messages", chatHandler.GetMessages)
	app.Post("/conversations/:id/messages", chatHandler.SendMessage)

	log.Println("Model:", cfg.OllamaModel)
	log.Println("Server listening on port:", cfg.AppPort)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
