package service

import (
	"context"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestService_Register(t *testing.T) {
	type tt struct {
		name     string
		user     models.User
		response models.AuthResponse
		error    bool
		err      error
	}
	tests := []tt{
		{
			name:     "positive",
			user:     models.User{Login: "234298", PassHash: "text"},
			response: models.AuthResponse{Refresh: "234298", Access: "234298", Salt: "234298"},
			err:      nil,
		},
		{
			name:  "positive",
			user:  models.User{Login: "2342899", PassHash: "text"},
			error: true,
			err:   store.ErrDouble,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := store.MockRepository{}
			r.On("Register", mock.Anything, &test.user).Return(&test.response, test.err)

			cfg := config.Config{}
			service := New(&r, &cfg)
			response, err := service.Register(context.Background(), &test.user)

			if test.error {
				assert.Error(t, err)
				assert.ErrorIs(t, err, test.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &test.response, response)
			}

		})
	}
}
