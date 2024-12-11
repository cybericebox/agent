package useCase

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/lib/pkg/worker"
	"github.com/hashicorp/go-multierror"
	"sync"
)

type (
	IRestoreService interface {
		GetStoredLabs(ctx context.Context) ([]model.Lab, error)
		RestoreLabIfNeeded(ctx context.Context, lab model.Lab) error
	}
)

func (u *UseCase) Restore() error {
	if err := u.RestoreLabsFromState(context.Background()); err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to restore labs from state").Err()
	}
	return nil
}

func (u *UseCase) RestoreLabsFromState(ctx context.Context) error {
	// get all the labs in the state
	labs, err := u.service.GetStoredLabs(ctx)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get stored labs").Err()
	}
	var errs error
	wg := new(sync.WaitGroup)

	// check if the labs exist in the infrastructure
	for _, lab := range labs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(lab.ID.String(), "restore_lab").
			WithDo(func() error {
				if err = u.service.RestoreLabIfNeeded(ctx, lab); err != nil {
					errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to restore lab").WithContext("labID", lab.ID.String()).Err())
					return err
				}
				return nil
			}).
			WithOnDone(func(_, _ error) {
				wg.Done()
			}).Create())

	}

	wg.Wait()

	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to restore labs from state").Err()
	}

	return nil
}
