package repository

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"semantic-search-api/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func vectorToString(v []float64) (string, error) {
	if len(v) != 1024 {
		return "", fmt.Errorf("expected 1024 dimensions")
	}

	values := make([]string, len(v))

	for i, x := range v {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return "", fmt.Errorf("invalid vector value")
		}
		values[i] = strconv.FormatFloat(x, 'g', -1, 64)
	}

	return "[" + strings.Join(values, ",") + "]", nil
}

func (r *DocumentRepository) Create(
	ctx context.Context,
	doc model.Document,
	embedding []float64,
) (model.Document, error) {
	vector, err := vectorToString(embedding)
	if err != nil {
		return doc, err
	}

	err = r.db.QueryRow(ctx, `
		INSERT INTO documents (title, content, embedding)
		VALUES ($1, $2, $3::vector)
		ON CONFLICT (title)
		DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding
		RETURNING id
	`, doc.Title, doc.Content, vector).Scan(&doc.ID)

	return doc, err
}

func (r *DocumentRepository) List(
	ctx context.Context,
) ([]model.Document, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, content
		FROM documents
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := make([]model.Document, 0)

	for rows.Next() {
		var doc model.Document
		if err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content,
		); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

func (r *DocumentRepository) Search(
	ctx context.Context,
	embedding []float64,
	topK int,
) ([]model.SearchResult, error) {
	vector, err := vectorToString(embedding)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id, title, content,
			1 - (embedding <=> $1::vector) AS similarity
		FROM documents
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vector, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.SearchResult, 0)

	for rows.Next() {
		var result model.SearchResult

		if err := rows.Scan(
			&result.ID,
			&result.Title,
			&result.Content,
			&result.Similarity,
		); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
