package service

import (
	"context"
	"errors"
	"strings"

	"semantic-search-api/internal/embedding"
	"semantic-search-api/internal/model"
	"semantic-search-api/internal/repository"
)

var ErrInvalidInput = errors.New("invalid input")

type DocumentService struct {
	repo   *repository.DocumentRepository
	ollama *embedding.Client
}

func NewDocumentService(
	repo *repository.DocumentRepository,
	ollama *embedding.Client,
) *DocumentService {
	return &DocumentService{
		repo:   repo,
		ollama: ollama,
	}
}

func (s *DocumentService) Create(
	ctx context.Context,
	req model.CreateDocumentRequest,
) (model.Document, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)

	if req.Title == "" || req.Content == "" {
		return model.Document{}, ErrInvalidInput
	}

	vector, err := s.ollama.Generate(
		ctx,
		req.Title+". "+req.Content,
	)
	if err != nil {
		return model.Document{}, err
	}

	return s.repo.Create(ctx, model.Document{
		Title:   req.Title,
		Content: req.Content,
	}, vector)
}

func (s *DocumentService) List(
	ctx context.Context,
) ([]model.Document, error) {
	return s.repo.List(ctx)
}

func (s *DocumentService) Search(
	ctx context.Context,
	req model.SearchRequest,
) ([]model.SearchResult, error) {
	req.Query = strings.TrimSpace(req.Query)

	if req.Query == "" {
		return nil, ErrInvalidInput
	}

	if req.TopK == 0 {
		req.TopK = 3
	}
	if req.TopK < 1 || req.TopK > 100 {
		return nil, ErrInvalidInput
	}

	vector, err := s.ollama.Generate(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	return s.repo.Search(ctx, vector, req.TopK)
}
