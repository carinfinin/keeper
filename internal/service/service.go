package service

import (
	"context"
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/jwtr"
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
	// Хеширование пароля
	passHash, err := bcrypt.GenerateFromPassword([]byte(u.PassHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("generate password error: ", err)
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// Сохранение пользователя
	id, err := s.Store.SaveUser(ctx, u.Login, passHash, u.Salt)
	if err != nil {
		logger.Log.Error("save user error: ", err)
		return nil, fmt.Errorf("user registration failed: %w", err)
	}
	u.ID = id

	// Генерация access токена
	accessToken, err := jwtr.Generate(u, "access", s.Config)
	if err != nil {
		logger.Log.Error("generate access token error: ", err)
		return nil, fmt.Errorf("access token generation failed: %w", err)
	}

	// Генерация refresh токена
	refreshToken, err := jwtr.Generate(u, "refresh", s.Config)
	if err != nil {
		logger.Log.Error("generate refresh token error: ", err)
		return nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	// Сохранение refresh токена в БД
	err = s.Store.SaveToken(ctx, u.ID, refreshToken)
	if err != nil {
		logger.Log.Error("save refresh token error: ", err)
		return nil, fmt.Errorf("refresh token storage failed: %w", err)
	}

	return &models.AuthResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *Service) Login(ctx context.Context, u *models.User) (*models.AuthResponse, error) {

	//_, err := jwtr.Decode("eyJhbGciOiJSUzI1NiIsImtpZCI6InYxIiwidHlwIjoiSldUIn0.eyJhdWQiOiJleGFtcGxlLWFwcCIsImV4cCI6MTc0OTgxMDk2NSwiaWF0IjoxNzQ5ODEwMDY1LCJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJuYmYiOjE3NDk4MTAwNjUsInByZWZlcnJlZF91c2VybmFtZSI6IiIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJzdWIiOjAsInR5cGUiOiJhY2Nlc3MifQ.dI2D5UV_q-Hb-ZTAs5hKWA_sdOA5EbHGg8YJ2kW0hiUJJCdLkMk3nSIK_tIWBzFiH_apz8AjEG7wBbwP5-5yBrGtOosKnDpwezmVUSB8H7I_ZaoUNCGZGPza-siOVhJgOJmSLFCsXmK7pXZIdN5_eyhNGoLL2Ib0G_uv3gTSV3zuZm7y8m9lT7vfoNa97NDIzQfePTpvzzkM6I8kJ2LxAlHATHJxVyFQwT2jA6FFEgLFFs2dgqoVfjefMzZacgfSXoc4biVwzFs0I1U9LsD7J9WWu064SRQp1TA0ce68K9SyAfEKWZiVnC2FwiULmFeaFdjuPdI_AX8qxVaa5Gia_G7GRRhVIYwZL8AVO7hJGbyo3K2W0EKNnwVH_CORoM9PYWiNT7j7IIonqH-lqKtqcEMrHHdusalFAZgiuDuWfMKp0Rdt5iR0A4hY6YrTVLVgPPKdc5q7JrF3mgREcDEtvT2XW6mXHWhxvv69qoIVfObhN6J6CzdiHfKDabvasegV8OWpyhC4OV9nNoOvTCZmZ-LLkhawRoIIQYSS08eqJTl0U8gYYZewllESiBhvmGDOVs5cR02gziBFFhz5Pgnu3sWyEDZmuB6SKhi4_DPhtFKmuQ52q5lOh0IFv_qoXs1ExJh4tc3Jht8iCADgi4NrtX3lu2FefyQx3I2F1cuwHEQ",
	//	r.Config)
	//if err != nil {
	//	fmt.Println(err)
	//}

	return &models.AuthResponse{}, nil
}

func (s *Service) Refresh(ctx context.Context, token string) (*models.AuthResponse, error) {

	return &models.AuthResponse{}, nil
}
