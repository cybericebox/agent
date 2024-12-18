package infrastructure

import (
	"github.com/cybericebox/agent/internal/config"
	k8s "github.com/cybericebox/agent/internal/delivery/infrastructure/kubernetes"
)

type (
	Infrastructure struct {
		*k8s.Kubernetes
	}

	Dependencies struct {
		Config *config.InfrastructureConfig
	}
)

func NewInfrastructure(deps Dependencies) *Infrastructure {
	return &Infrastructure{
		k8s.NewKubernetes(k8s.Dependencies{
			Config: &deps.Config.Kubernetes,
		}),
	}
}
