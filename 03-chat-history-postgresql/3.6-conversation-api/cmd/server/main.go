package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/handler"
	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/llm"
	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/repository"
	"go-ollama-playground/03-chat-history-postgresql/3.6-conversation-api/internal/service"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	databaseURL := "postgres://chatuser:chatpassword@localhost:5434/llm_chat"

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	ollamaClient, err := llm.NewClient("llama3.2:latest")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewChatRepository(db)
	chatService := service.NewChatService(repo, ollamaClient)
	chatHandler := handler.NewChatHandler(chatService)

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/conversations", chatHandler.CreateConversation)
	app.Get("/conversations", chatHandler.GetConversations)
	app.Get("/conversations/:id/messages", chatHandler.GetMessages)
	app.Post("/conversations/:id/messages", chatHandler.SendMessage)

	log.Println("Conversation API running on :8080")
	log.Fatal(app.Listen(":8080"))
}
