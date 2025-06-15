package storepg

import (
	"context"
	"errors"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sqlx.DB
}

func New(cfg *config.Config) (*Store, error) {
	db, err := sqlx.Open("pgx", cfg.DBPath)
	if err != nil {
		logger.Log.Errorf("store error: %v", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Log.Errorf("failed to ping database: %v", err)
		db.Close()
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) User(ctx context.Context, login string) (*models.User, error) {
	user := models.User{
		Login: login,
	}
	row := s.db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE login = $1", login)
	row.Scan(&user.ID, &user.PassHash)
	if err := row.Err(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) SaveUser(ctx context.Context, login string, passHash []byte) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id", login, passHash).Scan(&id)
	if err != nil {
		var errPG *pgconn.PgError
		if errors.As(err, &errPG) && pgerrcode.IsIntegrityConstraintViolation(errPG.Code) {
			return 0, store.ErrDouble
		}

		return 0, err
	}

	return id, nil
}
