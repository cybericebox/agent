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
	// ILabService interface
	ILabService interface {
		CreateLab(ctx context.Context, subnetMask uint32, labsGroupID string) (*model.Lab, error)
		GetLab(ctx context.Context, labID string) (*model.Lab, error)
		StartLab(ctx context.Context, labID string) error
		StopLab(ctx context.Context, labID string) error
		DeleteLab(ctx context.Context, labID string) error

		GetLabsStatus(ctx context.Context) ([]*model.LabStatus, error)
	}
)

func (u *UseCase) GetLabs(ctx context.Context, labsGroupID string, labIDs []string) ([]*model.Lab, error) {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return nil, appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	labs := make([]*model.Lab, 0, len(labIDs))

	wg := new(sync.WaitGroup)

	for _, id := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(id, "get_lab").
			WithDo(func() error {
				lab, err := u.service.GetLab(ctx, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				labs = append(labs, lab)
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		return nil, appError.ErrPlatform.WithError(errs).WithMessage("Failed to get labs").Err()
	}

	return labs, nil

}

func (u *UseCase) GetLabsStatus(ctx context.Context) ([]*model.LabStatus, error) {
	labs, err := u.service.GetLabsStatus(ctx)
	if err != nil {
		return nil, err
	}

	return labs, nil
}

func (u *UseCase) CreateLabs(ctx context.Context, labsGroupID string, subnetMask uint32, count int) ([]*model.Lab, error) {
	var errs error

	labs := make([]*model.Lab, 0, count)

	wg := new(sync.WaitGroup)

	for i := 0; i < count; i++ {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithDo(func() error {
				lab, err := u.service.CreateLab(ctx, subnetMask, labsGroupID)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				labs = append(labs, lab)
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		return nil, appError.ErrPlatform.WithError(errs).WithMessage("Failed to create labs").Err()
	}

	return labs, nil
}

func (u *UseCase) StartLabs(ctx context.Context, labsGroupID string, labIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, id := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(id, "start_lab").
			WithDo(func() error {
				err := u.service.StartLab(ctx, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to start labs").Err()
	}

	return nil
}

func (u *UseCase) StopLabs(ctx context.Context, labsGroupID string, labIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, id := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(id, "stop_lab").
			WithDo(func() error {
				err := u.service.StopLab(ctx, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to stop labs").Err()
	}

	return nil
}

func (u *UseCase) DeleteLabs(ctx context.Context, labsGroupID string, labIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, id := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(id, "delete_lab").
			WithDo(func() error {
				err := u.service.DeleteLab(ctx, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to delete labs").Err()
	}

	return nil
}
