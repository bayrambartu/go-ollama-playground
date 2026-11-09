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
)

const (
	ollamaURL = "http://localhost:11434/api/embed"
	model     = "bge-m3:latest"
)

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func getEmbedding(
	ctx context.Context,
	texts []string,
) ([][]float64, error) {

	requestBody := EmbedRequest{
		Model: model,
		Input: texts,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(
			io.LimitReader(resp.Body, 4096),
		)
		return nil, fmt.Errorf(
			"ollama error %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var result EmbedResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf(
			"expected %d embeddings, got %d",
			len(texts),
			len(result.Embeddings),
		)
	}

	return result.Embeddings, nil
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	texts := []string{
		"What is a goroutine?",
		"What is a goroutine?",
	}

	embeddings, err := getEmbedding(ctx, texts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Model:", model)
	fmt.Println("Total embeddings:", len(embeddings))

	for i, embedding := range embeddings {
		fmt.Println("\nText:", texts[i])
		fmt.Println("Vector size:", len(embedding))

		fmt.Println("First 10 values:")

		limit := 10
		if len(embedding) < limit {
			limit = len(embedding)
		}

		for j := 0; j < limit; j++ {
			fmt.Printf(
				"embedding[%d] = %.6f\n",
				j,
				embedding[j],
			)
		}
	}
}
