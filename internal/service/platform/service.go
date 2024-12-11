package platform

import (
	"context"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/delivery/repository/postgres"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/gofrs/uuid"
)

type (
	IInfrastructure interface {
		GetPodsMetrics(ctx context.Context, namespace string, selectors ...string) ([]model.PodMetrics, error)
		GetDeploymentsInNamespaceBySelector(ctx context.Context, namespace string, selector ...string) ([]model.DeploymentStatus, error)
	}

	IRepository interface {
		GetLaboratories(ctx context.Context) ([]postgres.Laboratory, error)
	}

	Dependencies struct {
		Infrastructure IInfrastructure
		Repository     IRepository
	}

	PlatformService struct {
		infrastructure IInfrastructure
		repository     IRepository
	}
)

func NewPlatformService(deps Dependencies) *PlatformService {
	return &PlatformService{
		infrastructure: deps.Infrastructure,
		repository:     deps.Repository,
	}
}

func (s *PlatformService) GetLabsStatus(ctx context.Context) ([]*model.LabStatus, error) {
	labs, err := s.repository.GetLaboratories(ctx)
	if err != nil {
		return nil, appError.ErrPlatform.WithError(err).WithMessage("Failed to get all laboratories").Err()
	}

	labsMap := make(map[string]*model.LabStatus)
	for _, lab := range labs {
		labsMap[lab.ID.String()] = &model.LabStatus{
			ID:        lab.ID,
			CIDR:      lab.Cidr.String(),
			DNS:       &model.DNSStatus{},
			Instances: make([]model.InstanceStatus, 0),
		}
	}

	deps, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, "", config.PlatformLabel)
	if err != nil {
		return nil, appError.ErrPlatform.WithError(err).WithMessage("Failed to get all platform deployments").Err()
	}
	pods, err := s.infrastructure.GetPodsMetrics(ctx, "", config.PlatformLabel)
	if err != nil {
		return nil, appError.ErrPlatform.WithError(err).WithMessage("Failed to get all platform pods").Err()
	}

	podsMap := make(map[string]model.PodMetrics)
	for _, pod := range pods {
		t := pod.Labels[config.PlatformLabel]
		if t == config.Challenge {
			podsMap[pod.Labels[config.InstanceIDLabel]] = pod
		}
		if t == config.LabDNSServer {
			labsMap[pod.Labels[config.LabIDLabel]].DNS.Resources = pod.Resources
		}
	}

	for _, dep := range deps {
		t := dep.Labels[config.PlatformLabel]
		if t == config.Challenge {
			labsMap[dep.Labels[config.LabIDLabel]].Instances = append(labsMap[dep.Labels[config.LabIDLabel]].Instances, model.InstanceStatus{
				ID:          uuid.FromStringOrNil(dep.Labels[config.InstanceIDLabel]),
				ChallengeID: uuid.FromStringOrNil(dep.Labels[config.ChallengeIDLabel]),
				Status:      dep.Status,
				Resources:   podsMap[dep.Labels[config.InstanceIDLabel]].Resources,
				Reason:      dep.Reason,
			})
		}
		if t == config.LabDNSServer {
			labsMap[dep.Labels[config.LabIDLabel]].DNS.Status = dep.Status
			labsMap[dep.Labels[config.LabIDLabel]].DNS.Reason = dep.Reason
		}
	}

	labsStatus := make([]*model.LabStatus, 0)
	for _, lab := range labsMap {
		labsStatus = append(labsStatus, lab)
	}

	return labsStatus, nil
}
