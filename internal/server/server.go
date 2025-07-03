package server

import (
	"context"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/router"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"time"
)

// Server.
type Server struct {
	http.Server
	Config *config.Config
}

// New конструктор .
func New(cfg *config.Config, router *router.Router) *Server {

	var s Server
	s.Config = cfg
	s.Addr = cfg.Addr
	s.Handler = router.Handler
	s.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	s.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	return &s
}

// Stop останавливает server
func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

// Start запускает сервер.
func (s *Server) Start() error {
	if s.Config.IsTLS {
		return s.Serve(autocert.NewListener(s.Config.Addr))
	}
	logger.Log.Info("Sever start on addr: ", s.Addr)
	return s.ListenAndServe()
}
