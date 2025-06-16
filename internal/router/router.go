package router

import (
	"context"
	"encoding/json"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/go-chi/chi"
	"io"
	"net/http"
)

type ServiceInterface interface {
	Register(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Login(ctx context.Context, u *models.User) (*models.AuthResponse, error)
	Refresh(ctx context.Context, token string) (*models.AuthResponse, error)
}

type Router struct {
	Handler *chi.Mux
	service ServiceInterface
	Config  *config.Config
}

func New(cfg *config.Config, service ServiceInterface) *Router {
	return &Router{
		Handler: chi.NewRouter(),
		service: service,
		Config:  cfg,
	}
}

func (r *Router) Configure() {

	r.Handler.Route("/api", func(cr chi.Router) {
		cr.Post("/register", r.Register)
		cr.Post("/login", r.Login)
		cr.Post("/refresh", r.Refresh)

		//cr.With(r.AuthMiddleware).Post("/orders", r.OrderSave)
	})
}

func (r *Router) Register(writer http.ResponseWriter, request *http.Request) {

	var u models.User
	reader := json.NewDecoder(request.Body)
	err := reader.Decode(&u)
	if err != nil {
		logger.Log.Error("Register decoder error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	response, err := r.service.Register(request.Context(), &u)
	if err != nil {
		logger.Log.Error("service Register error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err = encoder.Encode(response)
	if err != nil {
		logger.Log.Error("Register Encoder error: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (r *Router) Login(writer http.ResponseWriter, request *http.Request) {

	var u models.User
	reader := json.NewDecoder(request.Body)
	err := reader.Decode(&u)
	if err != nil {
		logger.Log.Error("Register error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	response, err := r.service.Login(request.Context(), &u)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err = encoder.Encode(response)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (r *Router) Refresh(writer http.ResponseWriter, request *http.Request) {

	b, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	response, err := r.service.Refresh(request.Context(), string(b))
	if err != nil {
		logger.Log.Error("Login error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(writer)
	err = encoder.Encode(response)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}
