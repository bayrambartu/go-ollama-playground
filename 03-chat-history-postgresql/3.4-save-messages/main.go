package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-ollama-playground/03-chat-history-postgresql/3.4-save-messages/repository"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
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

	fmt.Println("PostgreSQL connected!")

	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// 1. Create a new conversation
	conversation, err := conversationRepo.Create(
		ctx,
		"Go Backend and LLM",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Conversation created: ID=%d Title=%s\n",
		conversation.ID,
		conversation.Title,
	)

	// 2. Save the user message
	userMessage, err := messageRepo.Create(
		ctx,
		conversation.ID,
		"user",
		"What is an interface in Go?",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"User message saved: ID=%d\n",
		userMessage.ID,
	)

	// 3. Save the assistant message
	assistantMessage, err := messageRepo.Create(
		ctx,
		conversation.ID,
		"assistant",
		"An interface is a type that defines method behavior.",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Assistant message saved: ID=%d\n",
		assistantMessage.ID,
	)

	// chat
	chatRepo := repository.NewChatRepository(db)

	conversationID, err := chatRepo.CreateWithFirstMessage(
		ctx,
		"Test",
		"",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction completed. Conversation ID:", conversationID)
}
