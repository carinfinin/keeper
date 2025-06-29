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
	LastSync(ctx context.Context) (*models.LastSync, error)
	SaveItems(ctx context.Context, items []*models.Item) ([]*models.Item, error)
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

		cr.With(r.AuthMiddleware).Get("/last_sync", r.LastSync)
		cr.With(r.AuthMiddleware).Post("/item", r.SaveItems)
		//cr.With(r.AuthMiddleware).Get("/item", r.Test)
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

	u.DeviceName = request.Header.Get("User-Agent")

	logger.Log.Info(u)

	response, err := r.service.Register(request.Context(), &u)
	if err != nil {
		logger.Log.Error("service Register error: ", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)

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

	u.DeviceName = request.Header.Get("User-Agent")

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

	var rt struct {
		RefreshToken string `json:"token"`
	}

	if err := json.NewDecoder(request.Body).Decode(&rt); err != nil {
		http.Error(writer, "Invalid request", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	response, err := r.service.Refresh(request.Context(), rt.RefreshToken)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		http.Error(writer, "BadRequest", http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(response)
	if err != nil {
		logger.Log.Error("Login error: ", err)
		http.Error(writer, "InternalServerError", http.StatusInternalServerError)
		return
	}
}

func (r *Router) LastSync(writer http.ResponseWriter, request *http.Request) {

	ls, err := r.service.LastSync(request.Context())
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}

	writer.Header().Set("Content-Type", "application/json")
	jw := json.NewEncoder(writer)
	err = jw.Encode(ls)
	if err != nil {
		http.Error(writer, "error get last sync", http.StatusInternalServerError)
	}
}

func (r *Router) SaveItems(writer http.ResponseWriter, request *http.Request) {

	items := make([]*models.Item, 0)
	b, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Log.Error("bad request SaveItems err: ", err)
		http.Error(writer, "bad request", http.StatusBadRequest)
	}
	defer request.Body.Close()
	err = json.Unmarshal(b, &items)
	if err != nil {
		logger.Log.Error("Unmarshal SaveItems err: ", err)
		http.Error(writer, "bad request", http.StatusBadRequest)
	}

	serverItemsLast, err := r.service.SaveItems(request.Context(), items)
	if err != nil {
		logger.Log.Error("service SaveItems err: ", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
	}

	writer.Header().Set("Content-Type", "application/json")
	jw := json.NewEncoder(writer)
	err = jw.Encode(serverItemsLast)
	if err != nil {
		http.Error(writer, "error saveItems", http.StatusInternalServerError)
	}
}
