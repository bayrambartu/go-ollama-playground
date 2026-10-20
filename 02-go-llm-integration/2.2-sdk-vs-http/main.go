package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ollama/ollama/api"
)

const model = "llama3.2:latest"

const prompt = "what is interface in Go? explain it to just 1 sentence."

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message   Message `json:"message"`
	EvalCount int     `json:"eval_count"`
}

func main() {
	fmt.Println("=== HTTP API ===")
	if err := callHTTP(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== OLLAMA GO SDK ===")
	if err := callSDK(); err != nil {
		log.Fatal(err)
	}
}

// Opt. 1:  HTTP
func callHTTP() error {
	reqBody := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Post(
		"http://localhost:11434/api/chat",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Println(result.Message.Content)
	fmt.Println("Generated tokens:", result.EvalCount)

	return nil
}

// Opt. 2: Ollama SDK
func callSDK() error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	stream := false

	req := &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{Role: "user", Content: prompt},
		},
		Stream: &stream,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	return client.Chat(ctx, req, func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)

		if resp.Done {
			fmt.Println()
			fmt.Println("Generated tokens:", resp.EvalCount)
		}
		return nil
	})
}
