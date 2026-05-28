-- +goose Up
ALTER TABLE users
    ADD COLUMN email TEXT,
    ADD COLUMN phone TEXT,
    ADD COLUMN device_id TEXT;

CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX idx_users_phone ON users (phone) WHERE phone IS NOT NULL;
CREATE UNIQUE INDEX idx_users_device_id ON users (device_id) WHERE device_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_device_id;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users
    DROP COLUMN IF EXISTS device_id,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS email;
