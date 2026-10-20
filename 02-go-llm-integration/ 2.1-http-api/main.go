package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

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
	Message         Message `json:"message"`
	Done            bool    `json:"done"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
}

func main() {
	request := ChatRequest{
		Model: "llama3.2:latest",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a Go instructor. Give short answers in Turkish.",
			},
			{
				Role:    "user",
				Content: "What is the difference between pointer and value receivers in Go?",
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// http post

	resp, err := client.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Ollama error (%d): %s", resp.StatusCode, body)
	}

	var result ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- MODEL ANSWER ---")
	fmt.Println(result.Message.Content)

	fmt.Println("\n--- TOKEN STATISTICS ---")
	fmt.Println("Prompt tokens:", result.PromptEvalCount)
	fmt.Println("Generated tokens:", result.EvalCount)
}
