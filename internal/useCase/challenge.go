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
	IChallengeService interface {
		AddLabChallenges(ctx context.Context, labID string, configs []model.ChallengeConfig) error
		DeleteLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
		StartLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
		StopLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
		ResetLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
	}
)

func (u *UseCase) AddLabsChallenges(ctx context.Context, labsGroupID string, labIDs []string, challengesConfigs []model.ChallengeConfig, flagEnvVariables map[string]map[string]map[string]model.EnvConfig) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, labID := range labIDs {
		labChallengesConfigs := make([]model.ChallengeConfig, 0, len(challengesConfigs))

		for _, chConfig := range challengesConfigs {
			instances := make([]model.InstanceConfig, 0, len(chConfig.Instances))

			for _, inst := range chConfig.Instances {
				flagEnv, ok := flagEnvVariables[labID][chConfig.ID][inst.ID]
				if ok {
					inst.Envs = append(inst.Envs, flagEnv)
				}

				instances = append(instances, inst)
			}

			labChallengesConfigs = append(labChallengesConfigs, model.ChallengeConfig{ID: chConfig.ID, Instances: instances})
		}
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(labID, "add_lab_challenges").
			WithDo(func() error {
				if err := u.service.AddLabChallenges(ctx, labID, labChallengesConfigs); err != nil {
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
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to add challenges").Err()
	}

	return nil
}

func (u *UseCase) StartLabsChallenges(ctx context.Context, labsGroupID string, labIDs, challengeIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, labID := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(labID, "start_lab_challenges").
			WithDo(func() error {
				if err := u.service.StartLabChallenges(ctx, labID, challengeIDs); err != nil {
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
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to start challenges").Err()
	}

	return nil
}

func (u *UseCase) StopLabsChallenges(ctx context.Context, labsGroupID string, labIDs, challengeIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, labID := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(labID, "stop_lab_challenges").
			WithDo(func() error {
				if err := u.service.StopLabChallenges(ctx, labID, challengeIDs); err != nil {
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
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to stop challenges").Err()
	}

	return nil
}

func (u *UseCase) ResetLabsChallenges(ctx context.Context, labsGroupID string, labIDs, challengeIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, labID := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(labID, "reset_lab_challenges").
			WithDo(func() error {
				if err := u.service.ResetLabChallenges(ctx, labID, challengeIDs); err != nil {
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
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to reset challenges").Err()
	}

	return nil
}

func (u *UseCase) DeleteLabsChallenges(ctx context.Context, labsGroupID string, labIDs, challengeIDs []string) error {
	var errs error

	labIDs, err := u.getLabIDs(ctx, labsGroupID, labIDs)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to get lab IDs").Err()
	}

	wg := new(sync.WaitGroup)

	for _, labID := range labIDs {
		wg.Add(1)
		u.worker.AddTask(worker.NewTask().
			WithKey(labID, "delete_lab_challenges").
			WithDo(func() error {
				if err := u.service.DeleteLabChallenges(ctx, labID, challengeIDs); err != nil {
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
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to delete challenges").Err()
	}

	return nil
}
