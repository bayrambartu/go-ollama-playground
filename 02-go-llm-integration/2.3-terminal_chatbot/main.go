package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

func main() {

	client, err := api.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	reader := bufio.NewScanner(os.Stdin)

	messages := []api.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that answers questions about Go programming language. Answer in 1 sentence in Turkish.",
		},
	}

	fmt.Println("=== GO AI CHATBOT ===")
	fmt.Println(" write 'exit' to quit.")
	fmt.Println(" write 'clear' to reset the conversation.")

	for {
		fmt.Print("\nYou: ")
		if !reader.Scan() {
			break
		}
		input := strings.TrimSpace(reader.Text())

		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("Bye!")
			break
		}

		if input == "clear" {
			messages = messages[:1]
			fmt.Println("conversation history cleared.")
			continue
		}

		if input == "history" {
			for i, msg := range messages {
				fmt.Printf(
					"%d - %s: %s\n",
					i,
					msg.Role,
					msg.Content,
				)
			}
			continue

		}

		messages = append(messages, api.Message{
			Role:    "user",
			Content: input,
		})

		stream := true

		req := &api.ChatRequest{
			Model:    "llama3.2:latest",
			Messages: messages,
			Stream:   &stream,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		var answer strings.Builder

		fmt.Print("\nAI: ")

		err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)
			answer.WriteString(resp.Message.Content)
			return nil
		})

		cancel()

		if err != nil {
			fmt.Println("\nError:", err)

			messages = messages[:len(messages)-1]
			continue
		}

		// fmt.Println(answer.String())
		fmt.Println()

		messages = append(messages, api.Message{
			Role:    "assistant",
			Content: answer.String(),
		})
	}

	if err := reader.Err(); err != nil {
		log.Println("Error reading terminal:", err)
	}

}
