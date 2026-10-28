package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateWithFirstMessage(
	ctx context.Context,
	title string,
	firstMessage string,
) (int64, error) {
	if firstMessage == "" {
		return 0, fmt.Errorf("first message cannot be empty")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	var conversationID int64

	err = tx.QueryRow(
		ctx,
		`INSERT INTO conversations (title)
		 VALUES ($1)
		 RETURNING id`,
		title,
	).Scan(&conversationID)

	if err != nil {
		return 0, fmt.Errorf("insert conversation: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO messages
		 (conversation_id, role, content)
		 VALUES ($1, $2, $3)`,
		conversationID,
		"user",
		firstMessage,
	)

	if err != nil {
		return 0, fmt.Errorf("insert message: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return conversationID, nil
}
