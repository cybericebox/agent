package useCase

import "github.com/cybericebox/lib/pkg/worker"

type (
	IService interface {
		IRestoreService
		IChallengeService
		ILabService
	}

	Dependencies struct {
		Service IService
		Worker  worker.Worker
	}

	UseCase struct {
		service IService
		worker  worker.Worker
	}
)

func NewUseCase(deps Dependencies) *UseCase {
	return &UseCase{
		service: deps.Service,
		worker:  deps.Worker,
	}
}
