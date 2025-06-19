package storesqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(path string) (*sql.DB, error) {
	createTable := `CREATE TABLE IF NOT EXISTS secrets (
			uid TEXT PRIMARY KEY NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('login', 'text', 'binary', 'card')),
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS metadata (
			  secret_id TEXT NOT NULL REFERENCES secrets(uid) ON DELETE CASCADE,
			  key TEXT NOT NULL,
			  value TEXT NOT NULL,
			  PRIMARY KEY (secret_id, key)
		);`

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(createTable); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицу: %w", err)
	}
	return db, nil
}
