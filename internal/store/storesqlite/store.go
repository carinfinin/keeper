package storesqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/carinfinin/keeper/internal/store/models"
	_ "github.com/mattn/go-sqlite3"
	"strings"
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

func SaveItem(ctx context.Context, db *sql.DB, item *models.Item) error {

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO secrets (uid, type, data, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", item.UID, item.Type, item.Data, item.Created, item.Updated)
	if err != nil {
		return err
	}

	cond := make([]string, 0)
	values := make([]interface{}, 0)

	for k, v := range item.Meta {
		cond = append(cond, "(?, ?, ?)")
		values = append(values, item.UID, k, v)
	}
	if len(cond) > 0 {
		_, err = tx.ExecContext(ctx, "INSERT INTO metadata (secret_id, key, Value) VALUES "+strings.Join(cond, ", "), values...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
