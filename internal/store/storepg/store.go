package storepg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/jwtr"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/router"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// Store представляет основную структуру хранилища.
type Store struct {
	db     *sqlx.DB
	config *config.Config
}

// NotFoundRows - ошибка, возвращаемая когда запрашиваемые данные не найдены.
var NotFoundRows error = errors.New("not found rows")

// New создает новое подключение к хранилищу PostgreSQL.
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

// Refresh обновляет пару access/refresh токенов по валидному refresh токену.
func (s *Store) Refresh(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID int64
	var deviceID int64
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, device_id, expires_at 
		FROM refresh_tokens 
		WHERE token = $1 AND expires_at > NOW()`,
		refreshToken,
	).Scan(&userID, &deviceID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn("Invalid or expired refresh token")
			return nil, errors.New("invalid or expired refresh token")
		}
		logger.Log.Error("store refresh token lookup error: ", err)
		return nil, err
	}

	u := &models.User{ID: userID}
	row := tx.QueryRowContext(ctx, "SELECT login, salt FROM users WHERE id = $1", userID)
	if err = row.Scan(&u.Login, &u.Salt); err != nil {
		logger.Log.Error("store refresh get user error: ", err)
		return nil, err
	}

	resp, err := s.genToken(ctx, u)
	if err != nil {
		logger.Log.Error("Refresh gen token error: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE refresh_tokens SET token = $1, expires_at = $2 WHERE user_id = $3 AND device_id = $4`, resp.Refresh, time.Now().Add(time.Duration(s.config.RefreshTokenDuration)*time.Hour), userID, deviceID)
	if err != nil {
		logger.Log.Error("store refresh update token error: ", err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Error("store refresh commit error: ", err)
		return nil, err
	}

	resp.Salt = u.Salt
	return resp, nil
}

// Login выполняет аутентификацию пользователя по логину и паролю.
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

	var id int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM devices WHERE user_id = $1 AND device_name = $2", u.ID, u.DeviceName).Scan(&id)
	if err != nil {
		err := tx.QueryRowContext(ctx, "INSERT INTO devices (device_name, user_id) VALUES ($1, $2) RETURNING id", u.DeviceName, u.ID).Scan(&id)
		if err != nil {
			logger.Log.Error("service register add device error: ", err)
			return nil, err
		}
	}
	u.DeviceID = id

	resp, err := s.genToken(ctx, u)
	if err != nil {
		logger.Log.Error("Register gen token error: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, device_id, expires_at) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`, u.ID, resp.Refresh, u.DeviceID, time.Now().Add(time.Duration(s.config.RefreshTokenDuration)*time.Hour))
	if err != nil {
		logger.Log.Error("store login save token error: ", err)
		return nil, err
	}
	tx.Commit()

	resp.Salt = u.Salt

	return resp, nil
}

// Register регистрирует нового пользователя в системе.
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

	err = tx.QueryRowContext(ctx, "INSERT INTO devices (device_name, user_id, last_sync) VALUES ($1, $2, $3) RETURNING id", u.DeviceName, id, time.Now()).Scan(&u.DeviceID)
	if err != nil {
		logger.Log.Error("service register add device error: ", err)
		return nil, err
	}

	resp, err := s.genToken(ctx, u)
	if err != nil {
		logger.Log.Error("Register gen token error: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token, device_id, expires_at) VALUES ($1, $2, $3, $4)`, id, resp.Refresh, u.DeviceID, time.Now().Add(time.Duration(s.config.RefreshTokenDuration)*time.Hour))
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

// genToken генерирует новую пару JWT токенов для пользователя.
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

// Close закрывает соединение с базой данных.
func (s *Store) Close(ctx context.Context) error {
	return s.db.Close()
}

// SaveItems сохраняет или обновляет массив элементов в хранилище.
func (s *Store) SaveItems(ctx context.Context, items []*models.Item) ([]*models.Item, error) {

	userData, ok := ctx.Value(router.UserData).(*jwtr.JwtData)
	if !ok {
		return nil, fmt.Errorf("неверный тип данных пользователя в контексте")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	uids := make([]string, 0)

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO secrets (
            uid, user_id, type, data, description, created_at, updated_at, is_deleted
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (uid) DO UPDATE SET
            type = EXCLUDED.type,
            user_id = EXCLUDED.user_id,
            data = EXCLUDED.data,
            description = EXCLUDED.description,
            created_at = EXCLUDED.created_at,
            updated_at = EXCLUDED.updated_at,
            is_deleted = EXCLUDED.is_deleted
        WHERE secrets.updated_at < EXCLUDED.updated_at
        RETURNING uid`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {

		var uid string

		err = stmt.QueryRowContext(
			ctx,
			item.UID,
			userData.UserID,
			item.Type,
			item.Data,
			item.Description,
			item.Created,
			item.Updated,
			item.IsDeleted,
		).Scan(
			&uid,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("failed to save item %s: %w", item.UID, err)
		}

		uids = append(uids, uid)
	}

	// get

	var ls models.LastSync
	err = tx.QueryRowContext(ctx, `SELECT last_sync FROM devices WHERE id = $1`, userData.DeviceID).Scan(&ls.Update)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("client not found")
			return nil, NotFoundRows
		}
		return nil, err
	}

	res := make([]*models.Item, 0)
	params := []interface{}{ls.Update, userData.UserID}

	query := `SELECT uid, type, data, description, created_at, updated_at, is_deleted 
			  FROM secrets 
			  WHERE updated_at > $1 AND user_id = $2`

	if len(uids) > 0 {
		query += " AND NOT (uid = ANY($3))"
		params = append(params, pq.Array(uids))
	}

	fmt.Println(query)
	fmt.Println(params)

	rows, err := tx.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Log.Error("ошибка получения изменённых записей с сервера")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := models.Item{}
		err = rows.Scan(&tmp.UID, &tmp.Type, &tmp.Data, &tmp.Description, &tmp.Created, &tmp.Updated, &tmp.IsDeleted)
		if err != nil {
			logger.Log.Error("ошибка при сканировании получения изменённых с сервера err: ", err)
			return nil, err
		}
		res = append(res, &tmp)
	}
	err = rows.Err()
	if err != nil {
		logger.Log.Error("ошибка получения изменённых с сервера err: ", err)
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE devices SET last_sync = $1 WHERE id = $2`, time.Now(), userData.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to update last sync device: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return res, nil
}

// LastSync возвращает время последней синхронизации для текущего устройства.
func (s *Store) LastSync(ctx context.Context) (*models.LastSync, error) {
	userData, ok := ctx.Value(router.UserData).(*jwtr.JwtData)
	if !ok {
		return nil, fmt.Errorf("неверный тип данных пользователя в контексте")
	}

	var ls models.LastSync
	err := s.db.QueryRowContext(ctx, `
		SELECT last_sync 
		FROM devices 
		WHERE id = $1`,
		userData.DeviceID,
	).Scan(&ls.Update)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFoundRows
		}
		return nil, err
	}
	return &ls, nil
}
