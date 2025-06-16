package service

import (
	"context"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Store  store.Repository
	Config *config.Config
}

func New(s store.Repository, cfg *config.Config) *Service {
	return &Service{
		Store:  s,
		Config: cfg,
	}
}

func (s *Service) Register(ctx context.Context, u *models.User) (*models.AuthResponse, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.PassHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("generate password error: ", err)
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}
	u.PassHash = string(passHash)

	return s.Store.Register(ctx, u)
}

func (s *Service) Login(ctx context.Context, u *models.User) (*models.AuthResponse, error) {
	return s.Store.Login(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, token string) (*models.AuthResponse, error) {

	return &models.AuthResponse{}, nil
}
