package main

import (
	"fmt"
	cfg "github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/router"
	"github.com/carinfinin/keeper/internal/server"
	"github.com/carinfinin/keeper/internal/service"
)

func main() {

	config := cfg.New()
	fmt.Println(config)

	err := logger.Configure(config.LogLevel)
	if err != nil {
		fmt.Errorf("error configure logger : %v\n", err)
	}
	s := service.New()
	r := router.New(config, s)
	r.Configure()

	svr := server.New(config, r)
	if err = svr.Start(); err != nil {
		logger.Log.Fatal(err)
	}

	/*
		config
		logger
		bd
		service
		server
	*/
}
