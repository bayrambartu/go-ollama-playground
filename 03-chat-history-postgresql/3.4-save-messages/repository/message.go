package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
}

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(
	ctx context.Context,
	conversationID int64,
	role string,
	content string,
) (Message, error) {

	query := `
		INSERT INTO messages (conversation_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, role, content
	`

	var message Message

	err := r.db.QueryRow(
		ctx,
		query,
		conversationID,
		role,
		content,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.Role,
		&message.Content,
	)

	if err != nil {
		return Message{}, fmt.Errorf(
			"create message: %w", err,
		)
	}

	return message, nil
}
