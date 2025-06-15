CREATE TABLE IF NOT EXISTS tokens (
        id SERIAL PRIMARY KEY,
        access VARCHAR(255) NOT NULL,
        refresh VARCHAR(255) NOT NULL,
        user_id INT REFERENCES users(id) ON DELETE SET NULL,
        created_at TIMESTAMPTZ DEFAULT NOW()
    );
