package app

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/delivery/controller"
	"github.com/cybericebox/agent/internal/delivery/infrastructure"
	"github.com/cybericebox/agent/internal/delivery/repository"
	"github.com/cybericebox/agent/internal/service"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg := config.GetConfig()

	infra := infrastructure.NewInfrastructure(cfg.Service.LabsCIDR)

	repo := repository.NewRepository(repository.Dependencies{
		Config: &cfg.Repository,
	})

	serv := service.NewService(
		service.Dependencies{
			Config:         cfg,
			Infrastructure: infra,
			Repository:     repo,
		},
	)

	// test if the service is working properly
	log.Debug().Msg("Testing service")
	if err := serv.Test(); err != nil {
		log.Fatal().Err(err).Msg("Service initial test failed")
	}
	log.Info().Msg("Service test passed")

	// restore the state of the service
	log.Debug().Msg("Restoring service state")
	if err := serv.Restore(); err != nil {
		log.Fatal().Err(err).Msg("Service state restore failed")
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
