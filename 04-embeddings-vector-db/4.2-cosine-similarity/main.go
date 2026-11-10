package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) == 0 || len(a) != len(b) {
		return 0, errors.New("vectors must have equal non-zero dimensions")
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, errors.New("zero vector is not allowed")
	}

	similarity := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))

	return similarity, nil
}

func main() {
	texts := []string{
		"What is a goroutine?",
		"How do you run concurrent operations in Go?",
		"How do you create a Docker container?",
		"What is an index used for in a PostgreSQL database?",
		"What is a lightweight thread in Go?",
		"What is a goroutine?",
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	// Generate embeddings for all texts.
	embeddings, err := getEmbedding(ctx, texts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Query:", texts[0])
	fmt.Println()

	// Compare the first text's embedding with the others.
	for i := 1; i < len(texts); i++ {
		score, err := cosineSimilarity(
			embeddings[0],
			embeddings[i],
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(
			"Text: %s\nCosine Similarity: %.4f\n\n",
			texts[i],
			score,
		)
	}
}
