CREATE TABLE metadata (
      secret_id INTEGER NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
      key VARCHAR(255) NOT NULL,
      value TEXT NOT NULL,
      PRIMARY KEY (secret_id, key)
);