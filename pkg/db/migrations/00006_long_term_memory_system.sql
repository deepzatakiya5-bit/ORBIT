-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'memory_type_v2') THEN
        CREATE TYPE memory_type_v2 AS ENUM (
            'FACT',
            'PREFERENCE',
            'GOAL',
            'RELATIONSHIP',
            'EVENT',
            'EMOTIONAL_PATTERN',
            'IDENTITY',
            'ROUTINE'
        );
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'memory_validity') THEN
        CREATE TYPE memory_validity AS ENUM ('ACTIVE', 'SUPERSEDED', 'INACTIVE', 'CONFLICTED');
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'timeline_event_type') THEN
        CREATE TYPE timeline_event_type AS ENUM (
            'TRAVEL',
            'HOMETOWN_VISIT',
            'CAREER_CHANGE',
            'RELATIONSHIP_MILESTONE',
            'ACHIEVEMENT',
            'HEALTH',
            'ROUTINE_CHANGE',
            'CUSTOM'
        );
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type memory_type_v2 NOT NULL,
    content TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    salience REAL NOT NULL DEFAULT 0.5 CHECK (salience >= 0 AND salience <= 1),
    emotional_weight REAL NOT NULL DEFAULT 0 CHECK (emotional_weight >= 0 AND emotional_weight <= 1),
    recency_score REAL NOT NULL DEFAULT 0.5 CHECK (recency_score >= 0 AND recency_score <= 1),
    decay_score REAL NOT NULL DEFAULT 1 CHECK (decay_score >= 0 AND decay_score <= 1),
    frequency INT NOT NULL DEFAULT 1,
    explicit_importance REAL NOT NULL DEFAULT 0 CHECK (explicit_importance >= 0 AND explicit_importance <= 1),
    goal_relevance REAL NOT NULL DEFAULT 0 CHECK (goal_relevance >= 0 AND goal_relevance <= 1),
    user_emphasis REAL NOT NULL DEFAULT 0 CHECK (user_emphasis >= 0 AND user_emphasis <= 1),
    validity memory_validity NOT NULL DEFAULT 'ACTIVE',
    contradiction_group TEXT,
    source_conversation_id UUID REFERENCES conversations (id) ON DELETE SET NULL,
    source_message_id UUID REFERENCES messages (id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_referenced_at TIMESTAMPTZ,
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_memory_user_type_validity ON memory (user_id, type, validity);
CREATE INDEX IF NOT EXISTS idx_memory_user_salience ON memory (user_id, salience DESC);
CREATE INDEX IF NOT EXISTS idx_memory_contradiction_group ON memory (contradiction_group);

CREATE TABLE IF NOT EXISTS memory_embedding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL UNIQUE REFERENCES memory (id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT 'gemini',
    dimensions INT NOT NULL DEFAULT 1536,
    embedding vector(1536) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_memory_embedding_ivfflat
    ON memory_embedding USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE TABLE IF NOT EXISTS timeline_event (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations (id) ON DELETE SET NULL,
    memory_id UUID REFERENCES memory (id) ON DELETE SET NULL,
    event_type timeline_event_type NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    entities JSONB NOT NULL DEFAULT '[]'::jsonb,
    location TEXT,
    inferred BOOLEAN NOT NULL DEFAULT false,
    confidence REAL NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_timeline_event_user_start ON timeline_event (user_id, start_date DESC);
CREATE INDEX IF NOT EXISTS idx_timeline_event_type_start ON timeline_event (event_type, start_date DESC);

CREATE TABLE IF NOT EXISTS daily_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day_date DATE NOT NULL,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, day_date)
);

CREATE TABLE IF NOT EXISTS weekly_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, week_start)
);

CREATE TABLE IF NOT EXISTS monthly_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    month_date DATE NOT NULL,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, month_date)
);

CREATE TABLE IF NOT EXISTS yearly_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    year_date DATE NOT NULL,
    summary TEXT NOT NULL,
    highlights JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, year_date)
);

CREATE TABLE IF NOT EXISTS memory_relationship (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_memory_id UUID NOT NULL REFERENCES memory (id) ON DELETE CASCADE,
    target_memory_id UUID NOT NULL REFERENCES memory (id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_memory_id, target_memory_id, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_memory_relationship_target ON memory_relationship (target_memory_id);

CREATE TABLE IF NOT EXISTS memory_access_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    memory_id UUID NOT NULL REFERENCES memory (id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages (id) ON DELETE SET NULL,
    query_text TEXT,
    retrieval_score REAL,
    used_in_prompt BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_memory_access_log_user_created ON memory_access_log (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_access_log_memory_created ON memory_access_log (memory_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS memory_access_log;
DROP TABLE IF EXISTS memory_relationship;
DROP TABLE IF EXISTS yearly_summary;
DROP TABLE IF EXISTS monthly_summary;
DROP TABLE IF EXISTS weekly_summary;
DROP TABLE IF EXISTS daily_summary;
DROP TABLE IF EXISTS timeline_event;
DROP INDEX IF EXISTS idx_memory_embedding_ivfflat;
DROP TABLE IF EXISTS memory_embedding;
DROP TABLE IF EXISTS memory;
DROP TYPE IF EXISTS timeline_event_type;
DROP TYPE IF EXISTS memory_validity;
DROP TYPE IF EXISTS memory_type_v2;
