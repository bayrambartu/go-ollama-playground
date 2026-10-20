package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	stream := true

	req := &api.ChatRequest{
		Model: "llama3.2:latest",
		Messages: []api.Message{
			{
				Role: "system",
				Content: "You are a Go instructor. " +
					"Give clear answers in Turkish.",
			},
			{
				Role:    "user",
				Content: "What is Go? Explain it in one sentence.",
			},
		},
		Stream: &stream,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	var fullAnswer strings.Builder
	chunkCount := 0

	fmt.Println("--- STREAMING ANSWER ---")

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		// Print the new chunk.
		fmt.Print(resp.Message.Content)

		// Save the chunk to build the full answer.
		fullAnswer.WriteString(resp.Message.Content)

		if resp.Message.Content != "" {
			chunkCount++
		}

		if resp.Done {
			fmt.Println("\n\n--- STATISTICS ---")
			fmt.Println("Chunks received:", chunkCount)
			fmt.Println("Prompt tokens:", resp.PromptEvalCount)
			fmt.Println("Generated tokens:", resp.EvalCount)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nFull answer length:", fullAnswer.Len(), "bytes")
}
