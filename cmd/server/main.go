package main

import (
	"context"
	"fmt"
	cfg "github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/router"
	"github.com/carinfinin/keeper/internal/server"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/storepg"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	// config
	config := cfg.New()
	// logger
	err := logger.Configure(config.LogLevel)
	if err != nil {
		fmt.Errorf("error configure logger : %v\n", err)
	}
	// store
	bd, err := storepg.New(config)
	if err != nil {
		log.Fatal("error store : ", err)
	}
	//service
	s := service.New(bd, config)
	r := router.New(config, s)
	r.Configure()
	// server
	svr := server.New(config, r)
	go func() {
		if err = svr.Start(); err != nil {
			logger.Log.Error("server failed error: ", err)
		}
	}()

	logger.Log.Info("server started")

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err = svr.Stop(shutdownCtx); err != nil {
		logger.Log.Error("error stop server: ", err)
	}

	if err = bd.Close(shutdownCtx); err != nil {
		logger.Log.Error("error stop store: ", err)
	}
	logger.Log.Info("stopping server")
}
