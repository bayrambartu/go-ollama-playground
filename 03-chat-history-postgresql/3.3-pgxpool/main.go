package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-ollama-playground/03-chat-history-postgresql/3.3-pgxpool/repository"
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
		log.Fatal("Pool creation failed:", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	fmt.Println("PostgreSQL connected successfully!")

	messageRepo := repository.NewMessageRepository(db)

	conversationIDs := []int64{1, 2, 100}

	for _, conversationID := range conversationIDs {
		messages, err := messageRepo.GetByConversationID(
			ctx,
			conversationID,
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\n--- CONVERSATION %d ---\n", conversationID)

		for _, msg := range messages {
			fmt.Printf(
				"%s: %s\n",
				msg.Role,
				msg.Content,
			)
		}
	}

	conversations, err := messageRepo.GetAllConversations(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- ALL CONVERSATIONS ---")

	for _, cnv := range conversations {
		fmt.Printf(
			"Conversation %d - %s\n",
			cnv.ID,
			cnv.Title,
		)
	}
}
