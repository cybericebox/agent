package infrastructure

import k8s "github.com/cybericebox/agent/internal/delivery/infrastructure/kubernetes"

type (
	Infrastructure struct {
		*k8s.Kubernetes
	}
)

func NewInfrastructure() *Infrastructure {
	return &Infrastructure{
		k8s.NewKubernetes(),
	}
}
