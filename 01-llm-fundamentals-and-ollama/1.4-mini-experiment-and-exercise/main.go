package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	messages := []api.Message{
		{
			Role: "system",
			Content: "You are an experienced Go backend instructor. " +
				"Speak only in Turkish and give simple examples.",
		},
		{
			Role:    "user",
			Content: "What is a goroutine?",
		},
		{
			Role: "assistant",
			Content: "A goroutine is a lightweight task that lets " +
				"functions run concurrently in Go.",
		},
		{
			Role:    "user",
			Content: "How do I start one? Show a short example.",
		},
	}

	req := &api.ChatRequest{
		Model:    "llama3.2:latest",
		Messages: messages,
		Options: map[string]interface{}{
			"num_predict": 100,
			"temperature": 0.2,
		},
	}

	fmt.Println("--- ANSWER MODEL ---")

	err = client.Chat(
		context.Background(),
		req,
		func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)

			if resp.Done {
				fmt.Println("\n\n--- STATISTICS ---")
				fmt.Println("Prompt tokens:", resp.PromptEvalCount)
				fmt.Println("Generated tokens:", resp.EvalCount)
			}

			return nil
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
