-- +goose Up
ALTER TABLE users
    ADD COLUMN name TEXT,
    ADD COLUMN gender TEXT
        CHECK (
            gender IS NULL
            OR gender IN (
                'male',
                'female',
                'non_binary',
                'other',
                'prefer_not_to_say'
            )
        ),
    ADD COLUMN birthdate DATE,
    ADD COLUMN occupation TEXT,
    ADD COLUMN preferred_language TEXT,
    ADD COLUMN onboarding_completed_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS birthdate,
    DROP COLUMN IF EXISTS occupation,
    DROP COLUMN IF EXISTS preferred_language,
    DROP COLUMN IF EXISTS onboarding_completed_at;
