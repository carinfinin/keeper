package storepg

import (
	"context"
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/jwtr"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Store struct {
	db     *sqlx.DB
	config *config.Config
}

func New(cfg *config.Config) (*Store, error) {
	db, err := sqlx.Open("pgx", cfg.DBPath)
	if err != nil {
		logger.Log.Errorf("store error: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Log.Errorf("failed to ping database: %v", err)
		db.Close()
		return nil, err
	}

	return &Store{
		db:     db,
		config: cfg,
	}, nil
}

//	func (s *Store) User(ctx context.Context, login string) (*models.User, error) {
//		user := models.User{
//			Login: login,
//		}
//		row := s.db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE login = $1", login)
//		row.Scan(&user.ID, &user.PassHash)
//		if err := row.Err(); err != nil {
//			return nil, err
//		}
//		return &user, nil
//	}

func (s *Store) Login(ctx context.Context, u *models.User) (*models.AuthResponse, error) {

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var passHash []byte

	row := tx.QueryRowContext(ctx, "SELECT id, password_hash, salt FROM users WHERE login = $1", u.Login)
	row.Scan(&u.ID, &passHash, &u.Salt)
	if err = row.Err(); err != nil {
		logger.Log.Error("store login get user error: ", err)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(passHash, []byte(u.PassHash))
	if err != nil {
		logger.Log.Error("store login CompareHashAndPassword error: ", err)
		return nil, err
	}

	resp, err := s.genToken(ctx, u)
	if err != nil {
		logger.Log.Error("Register gen token error: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`, u.ID, resp.Refresh, time.Now().Add(time.Hour*24*7))
	if err != nil {
		logger.Log.Error("store login save token error: ", err)
		return nil, err
	}
	tx.Commit()

	resp.Salt = u.Salt

	return resp, nil
}

//func (s *Store) SaveUser(ctx context.Context, u *models.User) (int64, error) {
//	var id int64
//	err := s.db.QueryRowContext(ctx, "INSERT INTO users (login, password_hash, salt) VALUES ($1, $2, $2) RETURNING id", u.Login, u.PassHash, u.Salt).Scan(&id)
//	if err != nil {
//		var errPG *pgconn.PgError
//		if errors.As(err, &errPG) && pgerrcode.IsIntegrityConstraintViolation(errPG.Code) {
//			return 0, store.ErrDouble
//		}
//		return 0, err
//	}
//	return id, nil
//}

func (s *Store) Register(ctx context.Context, u *models.User) (*models.AuthResponse, error) {

	tx, err := s.db.Begin()
	if err != nil {
		logger.Log.Error("service register begin error: ", err)
		return nil, err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx, "INSERT INTO users (login, password_hash, salt) VALUES ($1, $2, $3) RETURNING id", u.Login, u.PassHash, u.Salt).Scan(&id)
	if err != nil {
		logger.Log.Error("service register QueryRowContext error: ", err)
		var errPG *pgconn.PgError
		if errors.As(err, &errPG) && pgerrcode.IsIntegrityConstraintViolation(errPG.Code) {
			return nil, store.ErrDouble
		}
		return nil, err
	}
	u.ID = id

	fmt.Println(u)

	resp, err := s.genToken(ctx, u)
	if err != nil {
		logger.Log.Error("Register gen token error: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`, id, resp.Refresh, time.Now().Add(time.Hour*24*7))
	if err != nil {
		logger.Log.Error("service register save token error: ", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		logger.Log.Error("service register commit error: ", err)
		return nil, err
	}

	return resp, nil
}

//func (s *Store) SaveToken(ctx context.Context, userID int64, token string) error {
//	_, err := s.db.ExecContext(ctx, `
//        INSERT INTO refresh_tokens (user_id, token, expires_at)
//        VALUES ($1, $2, $3)`,
//		userID,
//		token,
//		time.Now().Add(time.Hour*24*7),
//	)
//	return err
//}

func (s *Store) genToken(ctx context.Context, u *models.User) (*models.AuthResponse, error) {
	if u == nil {
		logger.Log.Error("nil user in genToken")
		return nil, fmt.Errorf("user is nil")
	}

	if s.config == nil {
		logger.Log.Error("nil config in genToken")
		return nil, fmt.Errorf("store config is nil")
	}

	accessToken, err := jwtr.Generate(u, "access", s.config)
	if err != nil {
		logger.Log.Error("generate access token error: ", err)
		return nil, fmt.Errorf("access token generation failed: %w", err)
	}
	refreshToken, err := jwtr.Generate(u, "refresh", s.config)
	if err != nil {
		logger.Log.Error("generate refresh token error: ", err)
		return nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	return &models.AuthResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *Store) Close(ctx context.Context) error {

	return nil
}
