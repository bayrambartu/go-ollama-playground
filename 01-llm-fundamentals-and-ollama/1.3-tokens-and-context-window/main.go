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
		panic(err)
	}

	ctx := context.Background()

	// messages := []api.Message{
	// 	{
	// 		Role:    "system",
	// 		Content: "you are a expert in go programming language. answer the questions in Turkish. answer in 1 sentence.",
	// 	},
	// 	{
	// 		Role:    "user",
	// 		Content: "what is Goroutines?",
	// 	},

	// 	{
	// 		Role:    "user",
	// 		Content: "what is channels?",
	// 	},
	// }

	messages := []api.Message{
		{
			Role:    "system",
			Content: "You are a Go instructor. Give short answers in Turkish.",
		},
		{
			Role:    "user",
			Content: "What is a goroutine?",
		},
		{
			Role:    "assistant",
			Content: "A goroutine is x.",
		},
		{
			Role:    "user",
			Content: "How do I start one?",
		},
	}

	req := &api.ChatRequest{
		Model:    "llama3.2:latest",
		Messages: messages,
	}

	err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)

		if resp.Done {
			fmt.Println("\n\n--- ABOUT THE TOKENS ---")
			fmt.Println("Prompt tokens:", resp.PromptEvalCount)
			fmt.Println("Generated tokens:", resp.EvalCount)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

}
