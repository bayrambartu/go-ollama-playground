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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ollamaURL = "http://localhost:11434/api/embed"
	model     = "bge-m3:latest"

	databaseURL = "postgres://vectoruser:vectorpassword@localhost:5434/vector_db"
)

type Document struct {
	Title   string
	Content string
}

type SearchResult struct {
	ID      int64
	Title   string
	Content string
	Score   float64
}

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

var documents = []Document{
	{
		Title: "Go Concurrency",
		Content: "In the Go programming language, a goroutine " +
			"lets functions run concurrently. " +
			"The go keyword is used to start a goroutine.",
	},
	{
		Title: "Go Channels",
		Content: "In Go, channels provide safe data communication " +
			"between goroutines.",
	},
	{
		Title: "PostgreSQL Index",
		Content: "In a PostgreSQL database, using an index " +
			"helps queries run faster.",
	},
	{
		Title: "Docker",
		Content: "Docker lets applications run in a container " +
			"together with their dependencies.",
	},
	{
		Title: "RabbitMQ",
		Content: "RabbitMQ is a message queue system. " +
			"It provides asynchronous communication between publishers and consumers.",
	},
	{
		Title: "Go Mutex",
		Content: "In Go, a mutex synchronizes access to shared data " +
			"from multiple goroutines.",
	},
}

func getEmbeddings(
	ctx context.Context,
	texts []string,
) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

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
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
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

	for _, vector := range result.Embeddings {
		if len(vector) != 1024 {
			return nil, fmt.Errorf(
				"expected 1024 dimensions, got %d",
				len(vector),
			)
		}
	}

	return result.Embeddings, nil
}

// []float64 -> "[0.1,0.2,0.3,...]"
// pgvector can read this text representation.
func vectorToString(vector []float64) (string, error) {
	if len(vector) != 1024 {
		return "", fmt.Errorf("invalid vector dimension: %d", len(vector))
	}

	values := make([]string, len(vector))

	for i, value := range vector {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return "", fmt.Errorf("invalid vector value at index %d", i)
		}
		values[i] = strconv.FormatFloat(value, 'g', -1, 64)
	}

	return "[" + strings.Join(values, ",") + "]", nil
}

func seedDocuments(
	ctx context.Context,
	db *pgxpool.Pool,
) error {
	texts := make([]string, 0, len(documents))

	for _, doc := range documents {
		texts = append(texts, doc.Title+". "+doc.Content)
	}

	fmt.Println("Generating embeddings...")

	vectors, err := getEmbeddings(ctx, texts)
	if err != nil {
		return err
	}

	// We run the operations in a transaction so that
	// running seed again does not duplicate the same examples.
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, doc := range documents {
		vectorText, err := vectorToString(vectors[i])
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO documents (title, content, embedding)
			VALUES ($1, $2, $3::vector)
			ON CONFLICT (title)
			DO UPDATE SET
				content = EXCLUDED.content,
				embedding = EXCLUDED.embedding
		`, doc.Title, doc.Content, vectorText)
		if err != nil {
			return fmt.Errorf("insert %s: %w", doc.Title, err)
		}

		fmt.Printf("Saved: %s\n", doc.Title)
	}

	return tx.Commit(ctx)
}

func searchDocuments(
	ctx context.Context,
	db *pgxpool.Pool,
	query string,
	topK int,
) ([]SearchResult, error) {
	vectors, err := getEmbeddings(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	vectorText, err := vectorToString(vectors[0])
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `
		SELECT
			id,
			title,
			content,
			1 - (embedding <=> $1::vector) AS similarity
		FROM documents
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vectorText, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]SearchResult, 0)

	for rows.Next() {
		var result SearchResult

		err := rows.Scan(
			&result.ID,
			&result.Title,
			&result.Content,
			&result.Score,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, rows.Err()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./04-embeddings-vector-db/4.5-go-pgvector [seed|search]")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Minute,
	)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("database connection failed: ", err)
	}

	switch os.Args[1] {
	case "seed":
		if err := seedDocuments(ctx, db); err != nil {
			log.Fatal(err)
		}
		fmt.Println("All documents saved.")

	case "search":
		if len(os.Args) < 3 {
			log.Fatal("usage: search \"your question\"")
		}

		query := strings.Join(os.Args[2:], " ")

		results, err := searchDocuments(ctx, db, query, 3)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Query:", query)

		for i, result := range results {
			fmt.Printf(
				"\n%d. %s\nSimilarity: %.4f\nContent: %s\n",
				i+1,
				result.Title,
				result.Score,
				result.Content,
			)
		}

	default:
		log.Fatal("unknown command: ", os.Args[1])
	}
}
