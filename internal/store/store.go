package store

import (
	"context"
	"errors"
	"github.com/carinfinin/keeper/internal/store/models"
)

var ErrDouble = errors.New("login already taken")

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Repository --filename=repositorymock_test.go --inpackage
type Repository interface {
	Login(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Register(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*models.AuthResponse, error)
	Close(ctx context.Context) error
	LastSync(ctx context.Context) (*models.LastSync, error)
	SaveItems(ctx context.Context, items []*models.Item) ([]*models.Item, error)
}
