-- +goose Up
ALTER TABLE users
    ADD COLUMN nickname TEXT,
    ADD COLUMN timezone TEXT,
    ADD COLUMN communication_style TEXT
        CHECK (
            communication_style IS NULL
            OR communication_style IN (
                'warm_short',
                'warm_long',
                'direct_short',
                'playful_short'
            )
        ),
    ADD COLUMN why_here TEXT[];

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS nickname,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS communication_style,
    DROP COLUMN IF EXISTS why_here;
