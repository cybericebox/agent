package app

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/delivery/controller"
	"github.com/cybericebox/agent/internal/delivery/infrastructure"
	"github.com/cybericebox/agent/internal/delivery/repository"
	"github.com/cybericebox/agent/internal/service"
	"github.com/cybericebox/agent/internal/useCase"
	"github.com/cybericebox/lib/pkg/worker"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	cfg := config.MustGetConfig()

	w := worker.NewWorker(cfg.Worker.MaxWorkers, cfg.Worker.Throttle)

	infra := infrastructure.NewInfrastructure(infrastructure.Dependencies{
		Config: &cfg.Infrastructure,
	})

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
		log.Fatal().Err(err).Msg("IUseCase initial test failed")
	}
	log.Info().Msg("Service test passed")

	u := useCase.NewUseCase(useCase.Dependencies{
		Service: serv,
		Worker:  w,
	})

	// restore the state of the service
	log.Debug().Msg("Restoring service state")
	if err := u.Restore(); err != nil {
		log.Fatal().Err(err).Msg("IUseCase state restore failed")
	}

	ctrl := controller.NewController(controller.Dependencies{
		Config:  &cfg.Controller,
		UseCase: u,
	})

	ctrl.Start()

	log.Info().Msg("Application started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-quit
	// Stop the controller
	ctrl.Stop()
	log.Info().Msg("Controller stopped")
	// Stop repository
	repo.Close()
	log.Info().Msg("IRepository stopped")

	log.Info().Msg("Application stopped")
}
