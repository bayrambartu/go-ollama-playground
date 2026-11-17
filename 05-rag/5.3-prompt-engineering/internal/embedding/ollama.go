package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	url        string
	model      string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		url:   "http://localhost:11434/api/embed",
		model: "bge-m3:latest",
	}
}

func (c *Client) Generate(
	ctx context.Context,
	text string,
) ([]float64, error) {
	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"input": text,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf(
			"ollama returned %d: %s",
			resp.StatusCode,
			string(b),
		)
	}

	var result struct {
		Embeddings [][]float64 `json:"embeddings"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Embeddings) != 1 {
		return nil, fmt.Errorf("expected one embedding")
	}

	if len(result.Embeddings[0]) != 1024 {
		return nil, fmt.Errorf("invalid embedding dimensions")
	}

	return result.Embeddings[0], nil
}
