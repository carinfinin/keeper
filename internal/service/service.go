package service

import (
	"context"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/crypted"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"golang.org/x/crypto/bcrypt"
)

// Service
type Service struct {
	Store  store.Repository
	Config *config.Config
}

// New создает новый экземпляр Service.
func New(s store.Repository, cfg *config.Config) *Service {
	return &Service{
		Store:  s,
		Config: cfg,
	}
}

// Register регистрирует нового пользователя в системе.
func (s *Service) Register(ctx context.Context, u *models.User) (*models.AuthResponse, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.PassHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("generate password error: ", err)
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}
	salt, err := crypted.GenerateSalt()
	if err != nil {
		logger.Log.Error("generate salt error: ", err)
		return nil, fmt.Errorf("salt failed: %w", err)
	}
	u.Salt = salt

	u.PassHash = string(passHash)

	return s.Store.Register(ctx, u)
}

// Login выполняет аутентификацию пользователя.
func (s *Service) Login(ctx context.Context, u *models.User) (*models.AuthResponse, error) {
	return s.Store.Login(ctx, u)
}

// Refresh обновляет пару токенов доступа.
func (s *Service) Refresh(ctx context.Context, token string) (*models.AuthResponse, error) {
	return s.Store.Refresh(ctx, token)
}

// LastSync возвращает информацию о последней синхронизации данных.
func (s *Service) LastSync(ctx context.Context) (*models.LastSync, error) {
	return s.Store.LastSync(ctx)
}

// SaveItems сохраняет список элементов данных.
func (s *Service) SaveItems(ctx context.Context, items []*models.Item) ([]*models.Item, error) {
	return s.Store.SaveItems(ctx, items)
}
