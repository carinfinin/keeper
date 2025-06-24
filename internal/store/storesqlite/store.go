package storesqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/carinfinin/keeper/internal/store/models"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func InitDB(path string) (*sql.DB, error) {
	createTable := `CREATE TABLE IF NOT EXISTS secrets (
			uid TEXT PRIMARY KEY NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('login', 'text', 'binary', 'card')),
			data BLOB NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS metadata (
			  secret_id TEXT NOT NULL REFERENCES secrets(uid) ON DELETE CASCADE,
			  key TEXT NOT NULL,
			  value TEXT NOT NULL,
			  PRIMARY KEY (secret_id, key)
		);
		CREATE TABLE IF NOT EXISTS tokens (
			  id INTEGER PRIMARY KEY AUTOINCREMENT,
			  access TEXT NOT NULL,
			  refresh TEXT NOT NULL,
			  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`

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

	_, err = tx.ExecContext(ctx, "INSERT INTO secrets (uid, type, data, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", item.UID, item.Type, item.Data, item.Description, item.Created, item.Updated)
	if err != nil {
		return err
	}

	//cond := make([]string, 0)
	//values := make([]interface{}, 0)
	//
	//for k, v := range item.Meta {
	//	cond = append(cond, "(?, ?, ?)")
	//	values = append(values, item.UID, k, v)
	//}
	//if len(cond) > 0 {
	//	_, err = tx.ExecContext(ctx, "INSERT INTO metadata (secret_id, key, Value) VALUES "+strings.Join(cond, ", "), values...)
	//	if err != nil {
	//		return err
	//	}
	//}

	return tx.Commit()
}

func GetItem(ctx context.Context, db *sql.DB, uid string) (*models.Item, error) {
	var item models.Item

	err := db.QueryRowContext(ctx,
		`SELECT type, data, description, created_at, updated_at FROM secrets WHERE uid = ?`, uid).Scan(&item.Type, &item.Data, &item.Description, &item.Created, &item.Updated)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no rows found")
		}
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	return &item, nil
}

func SaveTokens(ctx context.Context, db *sql.DB, item *models.AuthResponse) error {
	_, err := db.ExecContext(ctx, "INSERT INTO tokens (access, refresh) VALUES (?, ?)", item.Access, item.Refresh)
	return err
}

// GetTokens - получает последние сохраненные токены
func GetTokens(ctx context.Context, db *sql.DB) (*models.AuthResponse, error) {
	var tokens models.AuthResponse
	err := db.QueryRowContext(ctx,
		`SELECT access, refresh FROM tokens ORDER BY id DESC LIMIT 1`,
	).Scan(&tokens.Access, &tokens.Refresh)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no tokens found")
		}
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	return &tokens, nil
}

// UpdateTokens - обновляет существующие токены
func UpdateTokens(ctx context.Context, db *sql.DB, item *models.AuthResponse) error {
	_, err := db.ExecContext(ctx,
		`UPDATE tokens SET 
			access = ?,
			refresh = ?,
			updated_at = ?
		WHERE id = (
			SELECT id FROM tokens ORDER BY id DESC LIMIT 1
		)`,
		item.Access,
		item.Refresh,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	return nil
}

// UpsertTokens - создает или обновляет токены
func UpsertTokens(ctx context.Context, db *sql.DB, item *models.AuthResponse) error {
	// Пробуем обновить последнюю запись
	res, err := db.ExecContext(ctx,
		`UPDATE tokens SET 
			access = ?,
			refresh = ?,
			updated_at = ?
		WHERE id = (
			SELECT id FROM tokens ORDER BY id DESC LIMIT 1
		)`,
		item.Access,
		item.Refresh,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Если ни одна запись не была обновлена, создаем новую
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return SaveTokens(ctx, db, item)
	}

	return nil
}
