package service

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/service/challenge"
	"github.com/cybericebox/agent/internal/service/dns"
	"github.com/cybericebox/agent/internal/service/lab"
	"github.com/cybericebox/agent/internal/service/platform"
	"github.com/cybericebox/lib/pkg/ipam"
	"github.com/rs/zerolog/log"
)

type (
	Service struct {
		*lab.LabService
		*challenge.ChallengeService
		*platform.PlatformService
	}

	IInfrastructure interface {
		lab.IInfrastructure
		challenge.IInfrastructure
		dns.IInfrastructure
		platform.IInfrastructure
	}

	labService struct {
		*challenge.ChallengeService
		*dns.DNSService
	}

	IRepository interface {
		lab.IRepository
		platform.IRepository
	}

	Dependencies struct {
		Config         *config.Config
		Infrastructure IInfrastructure
		Repository     IRepository
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
		PlatformService: platform.NewPlatformService(platform.Dependencies{
			Infrastructure: deps.Infrastructure,
			Repository:     deps.Repository,
		}),
	}
}
