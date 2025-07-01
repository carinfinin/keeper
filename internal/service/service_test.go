package service

import (
	"context"
	"errors"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/store"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestService_Register(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		mockSetup   func(*store.MockRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful registration",
			user: &models.User{
				Login:    "testuser",
				PassHash: "password123",
			},
			mockSetup: func(m *store.MockRepository) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(&models.AuthResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate user",
			user: &models.User{
				Login:    "existinguser",
				PassHash: "password123",
			},
			mockSetup: func(m *store.MockRepository) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(nil, store.ErrDouble)
			},
			wantErr:     true,
			expectedErr: store.ErrDouble,
		},
		{
			name: "password hashing error",
			user: &models.User{
				Login:    "testuser",
				PassHash: string(make([]byte, 100)), // слишком длинный пароль для bcrypt
			},
			mockSetup:   func(m *store.MockRepository) {},
			wantErr:     true,
			expectedErr: bcrypt.ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &store.MockRepository{}
			tt.mockSetup(mockRepo)

			svc := New(mockRepo, &config.Config{})

			resp, err := svc.Register(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.True(t, errors.Is(err, tt.expectedErr), "expected error: %v, got: %v", tt.expectedErr, err)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEqual(t, tt.user.PassHash, "password123", "password should be hashed")
				assert.NotEmpty(t, tt.user.Salt, "salt should be generated")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		mockSetup   func(*store.MockRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful login",
			user: &models.User{
				Login:    "testuser",
				PassHash: "password123",
			},
			mockSetup: func(m *store.MockRepository) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(&models.AuthResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid credentials",
			user: &models.User{
				Login:    "testuser",
				PassHash: "wrongpassword",
			},
			mockSetup: func(m *store.MockRepository) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(nil, errors.New("invalid credentials"))
			},
			wantErr:     true,
			expectedErr: errors.New("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &store.MockRepository{}
			tt.mockSetup(mockRepo)

			svc := New(mockRepo, &config.Config{})

			resp, err := svc.Login(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Refresh(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		mockSetup   func(*store.MockRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "successful refresh",
			token: "valid_refresh_token",
			mockSetup: func(m *store.MockRepository) {
				m.On("Refresh", mock.Anything, "valid_refresh_token").
					Return(&models.AuthResponse{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "invalid token",
			token: "invalid_token",
			mockSetup: func(m *store.MockRepository) {
				m.On("Refresh", mock.Anything, "invalid_token").
					Return(nil, errors.New("invalid token"))
			},
			wantErr:     true,
			expectedErr: errors.New("invalid token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &store.MockRepository{}
			tt.mockSetup(mockRepo)

			svc := New(mockRepo, &config.Config{})

			resp, err := svc.Refresh(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_LastSync(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*store.MockRepository)
		want        *models.LastSync
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful last sync",
			mockSetup: func(m *store.MockRepository) {
				m.On("LastSync", mock.Anything).
					Return(&models.LastSync{Update: someTime}, nil)
			},
			want:    &models.LastSync{Update: someTime},
			wantErr: false,
		},
		{
			name: "no sync data",
			mockSetup: func(m *store.MockRepository) {
				m.On("LastSync", mock.Anything).
					Return(nil, store.NotFoundRows)
			},
			wantErr:     true,
			expectedErr: store.NotFoundRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &store.MockRepository{}
			tt.mockSetup(mockRepo)

			svc := New(mockRepo, &config.Config{})

			got, err := svc.LastSync(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.True(t, errors.Is(err, tt.expectedErr), "expected error: %v, got: %v", tt.expectedErr, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_SaveItems(t *testing.T) {
	testItems := []*models.Item{
		{UID: "item1"},
		{UID: "item2"},
	}

	tests := []struct {
		name        string
		items       []*models.Item
		mockSetup   func(*store.MockRepository)
		want        []*models.Item
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "successful save",
			items: testItems,
			mockSetup: func(m *store.MockRepository) {
				m.On("SaveItems", mock.Anything, testItems).
					Return(testItems, nil)
			},
			want:    testItems,
			wantErr: false,
		},
		{
			name:  "save error",
			items: testItems,
			mockSetup: func(m *store.MockRepository) {
				m.On("SaveItems", mock.Anything, testItems).
					Return(nil, errors.New("save failed"))
			},
			wantErr:     true,
			expectedErr: errors.New("save failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &store.MockRepository{}
			tt.mockSetup(mockRepo)

			svc := New(mockRepo, &config.Config{})

			got, err := svc.SaveItems(context.Background(), tt.items)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

var someTime = time.Now() // глобальная переменная для тестов времени
