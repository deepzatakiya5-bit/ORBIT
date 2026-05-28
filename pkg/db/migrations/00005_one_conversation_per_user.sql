-- +goose Up
-- Keep each user's most recently updated conversation; merge messages from older threads.
WITH keeper AS (
    SELECT DISTINCT ON (user_id) id AS keep_id, user_id
    FROM conversations
    ORDER BY user_id, updated_at DESC, created_at DESC
)
UPDATE messages m
SET conversation_id = k.keep_id
FROM conversations c
JOIN keeper k ON k.user_id = c.user_id
WHERE m.conversation_id = c.id
  AND c.id <> k.keep_id;

DELETE FROM conversations c
WHERE NOT EXISTS (
    SELECT 1
    FROM (
        SELECT DISTINCT ON (user_id) id
        FROM conversations
        ORDER BY user_id, updated_at DESC, created_at DESC
    ) k
    WHERE k.id = c.id
);

CREATE UNIQUE INDEX idx_conversations_one_per_user ON conversations (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_conversations_one_per_user;
