package main

import (
	"context"
	"log"
	"os"
	"time"

	"semantic-search-api/internal/embedding"
	"semantic-search-api/internal/handler"
	"semantic-search-api/internal/repository"
	"semantic-search-api/internal/service"

	"semantic-search-api/internal/llm"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file:", err)
	}

	// Read environment variable
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewDocumentRepository(db)
	ollama := embedding.NewClient()
	svc := service.NewDocumentService(repo, ollama)
	h := handler.NewDocumentHandler(svc)

	llmClient := llm.NewClient()

	ragService := service.NewRAGService(svc, llmClient)
	ragHandler := handler.NewRAGHandler(ragService)

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Post("/documents", h.Create)
	app.Get("/documents", h.List)
	app.Post("/search", h.Search)

	app.Post("/ask", ragHandler.Ask)

	log.Println("Semantic Search API running on :8080")

	log.Fatal(app.Listen(":8080"))
}
