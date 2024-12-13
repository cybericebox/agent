package service

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

func (s *Service) Test() error {
	log.Debug().Msg("Testing service normal")
	if err := s.testNormal(); err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed normal test").Err()
	}
	log.Debug().Msg("Testing service deleting lab with challenges")
	if err := s.testDeletingLabWithChallenges(); err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed deleting lab with challenges test").Err()
	}

	return nil
}

func (s *Service) testNormal() error {
	// test if the service is working properly
	ctx := context.Background()

	var errs error
	// try to create a new lab
	log.Debug().Msg("Creating test lab")
	labID, err := s.LabService.CreateLab(ctx, 26, "")
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to create test lab").Err()
	}

	log.Debug().Msg("Adding test challenge to test lab")
	// try to add a challenge to the lab
	if err = s.LabService.AddLabChallenges(ctx, labID.ID.String(), []model.ChallengeConfig{{
		ID: "test-challenge",
		Instances: []model.InstanceConfig{{
			ID:    "test-instance",
			Image: "nginx:latest",
			Resources: model.ResourcesConfig{
				Requests: model.ResourceConfig{
					CPU:    5,
					Memory: 50 * 1024 * 1024,
				},
				Limit: model.ResourceConfig{
					CPU:    5,
					Memory: 50 * 1024 * 1024,
				},
			},
			Envs: []model.EnvConfig{{
				Name:  "TEST_ENV",
				Value: "test",
			}},
			Records: []model.DNSRecordConfig{{
				Type: "A",
				Name: "test.cybericebox.local",
			}},
		}},
	}}); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to add test challenge to test lab").Err())
	}

	log.Debug().Msg("Deleting test challenge from test lab")
	// try to delete the challenge
	if err = s.LabService.DeleteLabChallenges(ctx, labID.ID.String(), []string{"test-challenge"}); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to delete test challenge from test lab").Err())
	}

	log.Debug().Msg("Deleting test lab")
	// try to delete the lab
	if err = s.LabService.DeleteLab(ctx, labID.ID.String()); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to delete test lab").Err())
	}
	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to test normal").Err()
	}
	return nil
}

func (s *Service) testDeletingLabWithChallenges() error {
	// test if the service is working properly
	ctx := context.Background()

	var errs error
	log.Debug().Msg("Creating test lab")
	// try to create a new lab
	labID, err := s.LabService.CreateLab(ctx, 26, "")
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to create test lab").Err()
	}

	log.Debug().Msg("Adding test challenge to test lab")
	// try to add a challenge to the lab
	if err = s.LabService.AddLabChallenges(ctx, labID.ID.String(), []model.ChallengeConfig{{
		ID: "test-challenge",
		Instances: []model.InstanceConfig{{
			ID:    "test-instance",
			Image: "nginx:latest",
			Resources: model.ResourcesConfig{
				Requests: model.ResourceConfig{
					CPU:    5,
					Memory: 50 * 1024 * 1024,
				},
				Limit: model.ResourceConfig{
					CPU:    5,
					Memory: 50 * 1024 * 1024,
				},
			},
			Envs: []model.EnvConfig{{
				Name:  "TEST_ENV",
				Value: "test",
			}},
			Records: []model.DNSRecordConfig{{
				Type: "A",
				Name: "test.cybericebox.local",
			}},
		}},
	}}); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to add test challenge to test lab").Err())
	}

	log.Debug().Msg("Deleting test lab")
	// try to delete the lab
	if err = s.LabService.DeleteLab(ctx, labID.ID.String()); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to delete test lab").Err())
	}
	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to test deleting lab with challenges").Err()
	}
	return nil
}
