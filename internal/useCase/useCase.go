package useCase

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/lib/pkg/worker"
	"github.com/gofrs/uuid"
	"slices"
)

type (
	IService interface {
		IRestoreService
		IChallengeService
		ILabService

		GetStoredLabs(ctx context.Context, labsGroupID string) ([]model.Lab, error)
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

func (u *UseCase) getLabIDs(ctx context.Context, labsGroupID string, labIDs []string) ([]string, error) {
	parsedGroupID := uuid.FromStringOrNil(labsGroupID)
	if parsedGroupID.IsNil() {
		return labIDs, nil
	}

	labs, err := u.service.GetStoredLabs(ctx, labsGroupID)
	if err != nil {
		return nil, appError.ErrPlatform.WithError(err).WithMessage("Failed to get stored labs").Err()
	}

	ids := make([]string, 0, len(labs))
	for _, lab := range labs {
		if slices.Contains(labIDs, lab.ID.String()) || len(labIDs) == 0 {
			ids = append(ids, lab.ID.String())
		}
	}
	return ids, nil
}
