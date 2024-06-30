package app

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/delivery/controller"
	"github.com/cybericebox/agent/internal/delivery/infrastructure"
	"github.com/cybericebox/agent/internal/service"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg := config.GetConfig()

	infra := infrastructure.NewInfrastructure()

	serv := service.NewService(
		service.Dependencies{
			Config:         &cfg.Service,
			Infrastructure: infra,
		},
	)

	// test if the service is working properly
	if err := serv.Test(); err != nil {
		log.Fatal().Err(err).Msg("Service initial test failed")
	}

	ctrl := controller.NewController(controller.Dependencies{
		Config:  &cfg.Controller,
		Service: serv,
	})

	ctrl.Start()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	ctrl.Stop()
}
