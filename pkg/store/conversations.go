package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"orbit/pkg/models"
)

var ErrNotFound = errors.New("not found")

func (s *Store) CreateConversation(ctx context.Context, userID uuid.UUID, title *string) (models.Conversation, error) {
	var conv models.Conversation
	err := s.db.QueryRow(ctx, `
		INSERT INTO conversations (user_id, title)
		VALUES ($1, $2)
		RETURNING id, user_id, title, created_at, updated_at
	`, userID, title).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt,
	)
	return conv, err
}

func (s *Store) GetConversationByUser(ctx context.Context, userID uuid.UUID) (models.Conversation, error) {
	var conv models.Conversation
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE user_id = $1
	`, userID).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Conversation{}, ErrNotFound
	}
	return conv, err
}

func (s *Store) GetConversation(ctx context.Context, id uuid.UUID) (models.Conversation, error) {
	var conv models.Conversation
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`, id).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Conversation{}, ErrNotFound
	}
	return conv, err
}

func (s *Store) TouchConversation(ctx context.Context, conversationID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE conversations SET updated_at = now() WHERE id = $1
	`, conversationID)
	return err
}
