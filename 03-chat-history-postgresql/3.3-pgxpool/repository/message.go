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
type Conversation struct {
	ID    int64
	Title string
}

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) GetByConversationID(
	ctx context.Context,
	conversationID int64) ([]Message, error) {
	query := `
		SELECT id, conversation_id, role, content
		FROM messages
		WHERE conversation_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var msg Message

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content)

		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

func (r *MessageRepository) GetAllConversations(ctx context.Context) ([]Conversation, error) {
	query := `
		SELECT id, title
		FROM conversations
		ORDER BY id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)

	for rows.Next() {
		var cnv Conversation

		err := rows.Scan(&cnv.ID, &cnv.Title)

		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		conversations = append(conversations, cnv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return conversations, nil

}
