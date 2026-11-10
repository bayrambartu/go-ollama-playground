package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"time"
)

const (
	ollamaURL = "http://localhost:11434/api/embed"
	model     = "bge-m3:latest"
)

type Document struct {
	ID      int
	Title   string
	Content string
}

type SearchResult struct {
	Document Document
	Score    float64
}

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func getEmbeddings(
	ctx context.Context,
	texts []string,
) ([][]float64, error) {
	body, err := json.Marshal(EmbedRequest{
		Model: model,
		Input: texts,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(
			io.LimitReader(resp.Body, 4096),
		)
		return nil, fmt.Errorf(
			"ollama error %d: %s",
			resp.StatusCode,
			string(b),
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
		return 0, fmt.Errorf("invalid vector dimensions")
	}

	var dot, normA, normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero vector")
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

func semanticSearch(
	queryEmbedding []float64,
	documents []Document,
	documentEmbeddings [][]float64,
	topK int,
) ([]SearchResult, error) {
	if len(documents) != len(documentEmbeddings) {
		return nil, fmt.Errorf("document and embedding count mismatch")
	}

	results := make([]SearchResult, 0, len(documents))

	for i, doc := range documents {
		score, err := cosineSimilarity(
			queryEmbedding,
			documentEmbeddings[i],
		)
		if err != nil {
			return nil, err
		}

		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK < 0 {
		topK = 0
	}
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

func main() {
	documents := []Document{
		{
			ID:    1,
			Title: "Go Concurrency",
			Content: "In the Go programming language, a goroutine " +
				"lets functions run concurrently. " +
				"The go keyword is used to start a goroutine.",
		},
		{
			ID:    2,
			Title: "Go Channels",
			Content: "In Go, channels provide safe data communication " +
				"between goroutines.",
		},
		{
			ID:    3,
			Title: "PostgreSQL Index",
			Content: "In a PostgreSQL database, using an index " +
				"helps queries run faster.",
		},
		{
			ID:    4,
			Title: "Docker",
			Content: "Docker lets applications run in a container " +
				"together with their dependencies.",
		},
		{
			ID:    5,
			Title: "RabbitMQ",
			Content: "RabbitMQ is a message queue system. " +
				"It provides asynchronous communication between publishers and consumers.",
		},
		{
			ID:    6,
			Title: "Go Mutex",
			Content: "In Go, a mutex synchronizes access to shared data " +
				"from multiple goroutines.",
		},
	}

	query := "What is a Postgres index?"

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	// 1. Prepare the document texts.
	texts := make([]string, 0, len(documents))
	for _, doc := range documents {
		texts = append(texts, doc.Title+". "+doc.Content)
	}

	// 2. Generate the document embeddings.
	docEmbeddings, err := getEmbeddings(ctx, texts)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Generate the query embedding.
	queryVectors, err := getEmbeddings(ctx, []string{query})
	if err != nil {
		log.Fatal(err)
	}

	// 4. Find the 3 closest documents.
	results, err := semanticSearch(
		queryVectors[0],
		documents,
		docEmbeddings,
		5,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Query:", query)
	fmt.Println("\nTop 3 most relevant documents:")

	for i, result := range results {
		fmt.Printf(
			"\n%d. %s\nScore: %.4f\nContent: %s\n",
			i+1,
			result.Document.Title,
			result.Score,
			result.Document.Content,
		)
	}
}
