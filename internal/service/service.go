package service

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/challenge"
	"github.com/cybericebox/agent/internal/service/dns"
	"github.com/cybericebox/agent/internal/service/lab"
	"github.com/cybericebox/wireguard/pkg/ipam"
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

	Dependencies struct {
		Config         *config.ServiceConfig
		Infrastructure Infrastructure
	}
)

func NewService(deps Dependencies) *Service {
	IPAManager, err := ipam.NewIPAManager(ipam.Dependencies{
		PostgresConfig: ipam.PostgresConfig(deps.Config.Postgres),
		CIDR:           deps.Config.LabsCIDR,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize IPAManager")
	}

	challengeService := challenge.NewChallengeService(challenge.Dependencies{
		Infrastructure: deps.Infrastructure,
	})

	return &Service{
		LabService: lab.NewLabService(lab.Dependencies{
			Infrastructure: deps.Infrastructure,
			IPAManager:     IPAManager,
			Service: labService{
				ChallengeService: challengeService,
				DNSService:       dns.NewDNSService(deps.Infrastructure),
			},
		}),
		ChallengeService: challengeService,
	}
}

func (s *Service) Test() error {
	// test if the service is working properly
	ctx := context.Background()

	var errs error
	// try to create a new lab
	labID, err := s.LabService.CreateLab(ctx, 26)
	if err != nil {
		return fmt.Errorf("failed to create test lab: [%w]", err)
	}

	// try to add a challenge to the lab
	if err = s.LabService.AddLabChallenges(ctx, labID, []model.ChallengeConfig{{
		Id: "test-challenge",
		Instances: []model.InstanceConfig{{
			Id:    "test-instance",
			Image: "nginx:latest",
			Resources: model.ResourcesConfig{
				Requests: model.ResourceConfig{
					CPU:    "100m",
					Memory: "128Mi",
				},
				Limit: model.ResourceConfig{
					CPU:    "100m",
					Memory: "128Mi",
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
		errs = multierror.Append(errs, fmt.Errorf("failed to add test challenge to test lab: [%w]", err))
	}

	// try to delete the challenge
	if err = s.LabService.DeleteLabChallenges(ctx, labID, []string{"test-challenge"}); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to delete test challenge from test lab: [%w]", err))
	}
	//
	// try to delete the lab
	if err = s.LabService.DeleteLab(ctx, labID); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to delete test lab: [%w]", err))
	}
	if errs != nil {
		return errs
	}
	return nil
}
