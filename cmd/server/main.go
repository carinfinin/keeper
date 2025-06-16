package main

import (
	"fmt"
	cfg "github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/logger"
	"github.com/carinfinin/keeper/internal/router"
	"github.com/carinfinin/keeper/internal/server"
	"github.com/carinfinin/keeper/internal/service"
	"github.com/carinfinin/keeper/internal/store/storepg"
	"log"
)

func main() {
	// config
	config := cfg.New()
	fmt.Println(config)
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
	if err = svr.Start(); err != nil {
		logger.Log.Fatal(err)
	}
}
