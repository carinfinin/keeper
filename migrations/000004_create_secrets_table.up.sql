CREATE TABLE secrets (
         id SERIAL PRIMARY KEY,
         user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
         type VARCHAR(20) NOT NULL CHECK (type IN ('login', 'text', 'binary', 'card')),
         encrypted_data BYTEA NOT NULL,
         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
         version INTEGER NOT NULL DEFAULT 1
);