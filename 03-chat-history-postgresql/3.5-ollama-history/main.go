package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ollama/ollama/api"

	"go-ollama-playground/03-chat-history-postgresql/3.5-ollama-history/repository"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		90*time.Second,
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

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	messageRepo := repository.NewMessageRepository(db)

	// The Go Concurrency conversation we created earlier.
	conversationID := int64(4)

	history, err := messageRepo.GetByConversationID(
		ctx,
		conversationID,
	)
	if err != nil {
		log.Fatal(err)
	}

	if len(history) == 0 {
		log.Fatal("Conversation history is empty")
	}

	// Build the message format Ollama expects.
	messages := []api.Message{
		{
			Role: "system",
			Content: "You are a Go instructor. " +
				"Give short and clear answers in Turkish.",
		},
	}

	for _, msg := range history {
		messages = append(messages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})

		fmt.Printf("History message count: %d\n", len(history))
	}

	//A new question that continues the conversation.
	newQuestion := "How can I start it?"

	messages = append(messages, api.Message{
		Role:    "user",
		Content: newQuestion,
	})

	fmt.Println("--- SENDING MESSAGES TO OLLAMA ---")

	for _, msg := range messages {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}

	stream := false

	req := &api.ChatRequest{
		Model:    "llama3.2:latest",
		Messages: messages,
		Stream:   &stream,
	}

	var answer strings.Builder

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		answer.WriteString(resp.Message.Content)

		if resp.Done {
			fmt.Println("\n--- TOKEN STATISTICS ---")
			fmt.Printf("History message count: %d\n", len(history))
			fmt.Printf("Total messages sent to Ollama: %d\n", len(messages))
			fmt.Printf("Prompt token count: %d\n", resp.PromptEvalCount)
			fmt.Printf("Generated token count: %d\n", resp.EvalCount)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- ASSISTANT ANSWER ---")
	fmt.Println(answer.String())
}
