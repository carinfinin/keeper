package store

import (
	"context"
	"errors"
	"github.com/carinfinin/keeper/internal/store/models"
)

var ErrDouble = errors.New("login already taken")
var ErrNotAuth = errors.New("invalid login password pair")
var ErrRowDouble = errors.New("rows double")
var ErrBusy = errors.New("uploaded by another user")
var ErrBalanceLow = errors.New("there are insufficient funds in the account")

var ErrUserNotFound = errors.New("user if not found")

type Repository interface {
	Login(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Register(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*models.AuthResponse, error)
	Close(ctx context.Context) error
}
