package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	ID    int64
	Title string
}

type ConversationRepository struct {
	db *pgxpool.Pool
}

func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(
	ctx context.Context,
	title string,
) (Conversation, error) {

	query := `
		INSERT INTO conversations (title)
		VALUES ($1)
		RETURNING id, title
	`

	var conversation Conversation

	err := r.db.QueryRow(ctx, query, title).Scan(
		&conversation.ID,
		&conversation.Title,
	)

	if err != nil {
		return Conversation{}, fmt.Errorf(
			"create conversation: %w", err,
		)
	}

	return conversation, nil
}
