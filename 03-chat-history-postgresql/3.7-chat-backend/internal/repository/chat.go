package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConversationNotFound = errors.New("conversation not found")

type Conversation struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type Message struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	Role           string `json:"role"`
	Content        string `json:"content"`
}

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateConversation(
	ctx context.Context,
	title string,
) (Conversation, error) {
	var c Conversation

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO conversations (title)
		 VALUES ($1)
		 RETURNING id, title`,
		title,
	).Scan(&c.ID, &c.Title)

	return c, err
}

func (r *ChatRepository) GetConversations(
	ctx context.Context,
) ([]Conversation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title
		 FROM conversations
		 ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)

	for rows.Next() {
		var c Conversation

		if err := rows.Scan(&c.ID, &c.Title); err != nil {
			return nil, err
		}

		conversations = append(conversations, c)
	}

	return conversations, rows.Err()
}

func (r *ChatRepository) GetMessages(
	ctx context.Context,
	conversationID int64,
) ([]Message, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM conversations WHERE id = $1
		)`,
		conversationID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrConversationNotFound
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, conversation_id, role, content
		 FROM messages
		 WHERE conversation_id = $1
		 ORDER BY id ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var m Message

		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.Role,
			&m.Content,
		); err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}

	return messages, rows.Err()
}

func (r *ChatRepository) SaveMessage(
	ctx context.Context,
	conversationID int64,
	role, content string,
) (Message, error) {
	var m Message

	err := r.db.QueryRow(ctx,
		`INSERT INTO messages
		 (conversation_id, role, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, conversation_id, role, content`,
		conversationID,
		role,
		content,
	).Scan(
		&m.ID,
		&m.ConversationID,
		&m.Role,
		&m.Content,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, fmt.Errorf("save message: %w", err)
	}

	return m, err
}
