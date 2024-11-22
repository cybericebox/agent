package service

import (
	"context"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/challenge"
	"github.com/cybericebox/agent/internal/service/dns"
	"github.com/cybericebox/agent/internal/service/lab"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/lib/pkg/ipam"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

type (
	Service struct {
		*lab.LabService
		*challenge.ChallengeService
	}

	Infrastructure interface {
		lab.Infrastructure
		challenge.Infrastructure
		dns.Infrastructure
	}

	labService struct {
		*challenge.ChallengeService
		*dns.DNSService
	}

	Repository interface {
		lab.Repository
	}

	Dependencies struct {
		Config         *config.Config
		Infrastructure Infrastructure
		Repository     Repository
	}
)

func NewService(deps Dependencies) *Service {
	IPAManager, err := ipam.NewIPAManager(ipam.Dependencies{
		PostgresConfig: ipam.PostgresConfig(deps.Config.Repository.Postgres),
		CIDR:           deps.Config.Service.LabsCIDR,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize IPAManager")
	}

	challengeService := challenge.NewChallengeService(challenge.Dependencies{
		Infrastructure: deps.Infrastructure,
	})

	return &Service{
		LabService: lab.NewLabService(lab.Dependencies{
			Infrastructure: deps.Infrastructure,
			IPAManager:     IPAManager,
			Repository:     deps.Repository,
			Service: labService{
				ChallengeService: challengeService,
				DNSService:       dns.NewDNSService(deps.Infrastructure),
			},
		}),
		ChallengeService: challengeService,
	}
}

func (s *Service) Restore() error {
	if err := s.LabService.RestoreLabsFromState(context.Background()); err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to restore labs from state").Err()
	}
	return nil
}

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
	labID, err := s.LabService.CreateLab(ctx, 26)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to create test lab").Err()
	}

	log.Debug().Msg("Adding test challenge to test lab")
	// try to add a challenge to the lab
	if err = s.LabService.AddLabChallenges(ctx, labID, []model.ChallengeConfig{{
		ID: "test-challenge",
		Instances: []model.InstanceConfig{{
			ID:    "test-instance",
			Image: "nginx:latest",
			Resources: model.ResourcesConfig{
				Requests: model.ResourceConfig{
					CPU:    "5m",
					Memory: "50Mi",
				},
				Limit: model.ResourceConfig{
					CPU:    "5m",
					Memory: "50Mi",
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
	if err = s.LabService.DeleteLabChallenges(ctx, labID, []string{"test-challenge"}); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to delete test challenge from test lab").Err())
	}

	log.Debug().Msg("Deleting test lab")
	// try to delete the lab
	if err = s.LabService.DeleteLab(ctx, labID); err != nil {
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
	labID, err := s.LabService.CreateLab(ctx, 26)
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to create test lab").Err()
	}

	log.Debug().Msg("Adding test challenge to test lab")
	// try to add a challenge to the lab
	if err = s.LabService.AddLabChallenges(ctx, labID, []model.ChallengeConfig{{
		ID: "test-challenge",
		Instances: []model.InstanceConfig{{
			ID:    "test-instance",
			Image: "nginx:latest",
			Resources: model.ResourcesConfig{
				Requests: model.ResourceConfig{
					CPU:    "5m",
					Memory: "50Mi",
				},
				Limit: model.ResourceConfig{
					CPU:    "5m",
					Memory: "50Mi",
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
	if err = s.LabService.DeleteLab(ctx, labID); err != nil {
		errs = multierror.Append(errs, appError.ErrPlatform.WithError(err).WithMessage("Failed to delete test lab").Err())
	}
	if errs != nil {
		return appError.ErrPlatform.WithError(errs).WithMessage("Failed to test deleting lab with challenges").Err()
	}
	return nil
}
