CREATE TABLE secrets (
     uid TEXT PRIMARY KEY NOT NULL,
     user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     type TEXT NOT NULL CHECK (type IN ('login', 'text', 'binary', 'card')),
     data BYTEA NOT NULL,
     description TEXT,
     created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
     is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);