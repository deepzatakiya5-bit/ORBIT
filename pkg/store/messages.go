package store

import (
	"context"

	"github.com/google/uuid"

	"orbit/pkg/models"
)

func (s *Store) CreateMessage(ctx context.Context, conversationID uuid.UUID, role, content string) (models.Message, error) {
	var msg models.Message
	err := s.db.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, role, content, created_at
	`, conversationID, role, content).Scan(
		&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt,
	)
	return msg, err
}

func (s *Store) ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, conversation_id, role, content, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
