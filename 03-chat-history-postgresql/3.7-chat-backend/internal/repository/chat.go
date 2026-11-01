package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
type ChatRequest struct {
	ID                 int64
	ConversationID     int64
	IdempotencyKey     string
	UserMessageID      int64
	AssistantMessageID *int64
	Status             string
	ErrorMessage       *string
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

func (r *ChatRepository) FindRequest(
	ctx context.Context,
	conversationID int64,
	key string,
) (ChatRequest, bool, error) {
	var req ChatRequest

	err := r.db.QueryRow(ctx, `
		SELECT id, conversation_id, idempotency_key,
		       user_message_id, assistant_message_id,
		       status, error_message
		FROM chat_requests
		WHERE conversation_id = $1
		  AND idempotency_key = $2
	`, conversationID, key).Scan(
		&req.ID,
		&req.ConversationID,
		&req.IdempotencyKey,
		&req.UserMessageID,
		&req.AssistantMessageID,
		&req.Status,
		&req.ErrorMessage,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ChatRequest{}, false, nil
	}
	if err != nil {
		return ChatRequest{}, false, err
	}

	return req, true, nil
}

// Save the message and request in one transaction.
func (r *ChatRepository) CreatePendingRequest(
	ctx context.Context,
	conversationID int64,
	key, content string,
) (ChatRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ChatRequest{}, err
	}
	defer tx.Rollback(ctx)

	var messageID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, role, content)
		VALUES ($1, 'user', $2)
		RETURNING id
	`, conversationID, content).Scan(&messageID)
	if err != nil {
		return ChatRequest{}, err
	}

	var req ChatRequest

	err = tx.QueryRow(ctx, `
		INSERT INTO chat_requests (
			conversation_id, idempotency_key,
			user_message_id, status
		)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id
	`, conversationID, key, messageID).Scan(&req.ID)
	if err != nil {
		return ChatRequest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ChatRequest{}, err
	}

	req.ConversationID = conversationID
	req.IdempotencyKey = key
	req.UserMessageID = messageID
	req.Status = "pending"

	return req, nil
}

// Update the assistant message and request status in one transaction.
func (r *ChatRepository) CompleteRequest(
	ctx context.Context,
	requestID, conversationID int64,
	answer string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var assistantID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, role, content)
		VALUES ($1, 'assistant', $2)
		RETURNING id
	`, conversationID, answer).Scan(&assistantID)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE chat_requests
		SET status = 'completed',
		    assistant_message_id = $1,
		    error_message = NULL,
		    updated_at = NOW()
		WHERE id = $2 AND status = 'pending'
	`, assistantID, requestID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("request is not pending")
	}

	return tx.Commit(ctx)
}

func (r *ChatRepository) FailRequest(
	ctx context.Context,
	requestID int64,
	reason string,
) error {
	_, err := r.db.Exec(ctx, `
		UPDATE chat_requests
		SET status = 'failed',
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, requestID, reason)

	return err
}

func (r *ChatRepository) GetMessageContent(
	ctx context.Context,
	messageID int64,
) (string, error) {
	var content string

	err := r.db.QueryRow(ctx, `
		SELECT content FROM messages WHERE id = $1
	`, messageID).Scan(&content)

	return content, err
}

// PostgreSQL advisory lock.
// Only one operation can run for each conversation.
func (r *ChatRepository) TryLockConversation(
	ctx context.Context,
	conversationID int64,
) (func(), bool, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}

	var locked bool

	err = conn.QueryRow(ctx,
		`SELECT pg_try_advisory_lock($1)`,
		conversationID,
	).Scan(&locked)

	if err != nil || !locked {
		conn.Release()
		return nil, false, err
	}

	unlock := func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_, err := conn.Exec(
			cleanupCtx,
			`SELECT pg_advisory_unlock($1)`,
			conversationID,
		)

		if err != nil {
			// It is not safe to return a locked connection to the pool.
			conn.Hijack().Close(context.Background())
			return
		}

		conn.Release()
	}

	return unlock, true, nil
}
