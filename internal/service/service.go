package service

import (
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/service/challenge"
	"github.com/cybericebox/agent/internal/service/dns"
	"github.com/cybericebox/agent/internal/service/lab"
	"github.com/cybericebox/wireguard/pkg/ipam"
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
		CIDR:           deps.Config.LabCIDR,
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
