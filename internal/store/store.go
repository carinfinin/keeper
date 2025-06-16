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
	User(ctx context.Context, login string) (*models.User, error)
	SaveUser(ctx context.Context, login string, passHash []byte, salt string) (int64, error)
	SaveToken(ctx context.Context, userID int64, token string) error
	Close(ctx context.Context) error
}
