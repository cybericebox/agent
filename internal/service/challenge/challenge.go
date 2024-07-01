package challenge

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/helper"
	"github.com/hashicorp/go-multierror"
)

type (
	Infrastructure interface {
		DeploymentExists(ctx context.Context, name, namespace string) (bool, error)
		ApplyDeployment(ctx context.Context, config model.ApplyDeploymentConfig) error
		GetDeploymentsInNamespaceBySelector(ctx context.Context, namespace string, selector ...string) ([]model.DeploymentStatus, error)
		ResetDeployment(ctx context.Context, name, namespace string) error
		ScaleDeployment(ctx context.Context, name, namespace string, scale int32) error
		DeleteDeployment(ctx context.Context, name, namespace string) error
	}

	ChallengeService struct {
		infrastructure Infrastructure
	}

	Dependencies struct {
		Infrastructure Infrastructure
	}
)

func NewChallengeService(deps Dependencies) *ChallengeService {
	return &ChallengeService{
		infrastructure: deps.Infrastructure,
	}
}

func (s *ChallengeService) CreateChallenge(ctx context.Context, lab *model.Lab, challengeConfig model.ChallengeConfig) (records []model.DNSRecordConfig, errs error) {
	for _, inst := range challengeConfig.Instances {
		// check if the instance is already deployed
		ex, err := s.infrastructure.DeploymentExists(ctx, inst.Id, lab.ID.String())
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to check if deployment exists: [%w]", err))
			continue
		}

		if ex {
			errs = multierror.Append(errs, fmt.Errorf("instance already exists: [%s]", inst.Id))
			continue
		}

		ip, err := lab.CIDRManager.AcquireSingleIP(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to acquire ip for instance: [%w]", err))
			continue
		}

		if err = s.infrastructure.ApplyDeployment(ctx, model.ApplyDeploymentConfig{
			Name:  inst.Id,
			LabID: lab.ID.String(),
			Labels: map[string]string{
				config.PlatformLabel:    config.Challenge,
				config.LabIDLabel:       lab.ID.String(),
				config.ChallengeIDLabel: challengeConfig.Id,
				config.InstanceIDLabel:  inst.Id,
				config.RecordsListLabel: helper.RecordsToStr(inst.Records),
			},
			Image:     inst.Image,
			Resources: inst.Resources,
			Envs:      inst.Envs,
			IP:        ip,
		}); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to apply deployment for instance: [%w]", err))
			if err = lab.CIDRManager.ReleaseSingleIP(ctx, ip); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("failed to release ip for instance in apply: [%w]", err))
			}
			continue
		}

		for _, r := range inst.Records {
			if r.Type == "A" {
				r.Data = ip
			}
			records = append(records, r)
		}
	}
	return
}

func (s *ChallengeService) DeleteChallenge(ctx context.Context, lab *model.Lab, challengeID string) (records []model.DNSRecordConfig, errs error) {
	dps, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, lab.ID.String(),
		fmt.Sprintf("%s=%s", config.PlatformLabel, config.Challenge),
		fmt.Sprintf("%s=%s", config.LabIDLabel, lab.ID.String()),
		fmt.Sprintf("%s=%s", config.ChallengeIDLabel, challengeID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances in namespace by selector: [%w]", err)
	}

	records = make([]model.DNSRecordConfig, 0)

	for _, dp := range dps {
		if err = s.infrastructure.DeleteDeployment(ctx, dp.Name, lab.ID.String()); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to delete deployment: [%w]", err))
		}

		if err = lab.CIDRManager.ReleaseSingleIP(ctx, dp.IP); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to release ip for instance in delete: [%w]", err))
		}

		records = append(records, helper.RecordsFromStr(dp.Labels[config.RecordsListLabel])...)
	}

	return
}

func (s *ChallengeService) StartChallenge(ctx context.Context, labID, challengeID string) (errs error) {
	dps, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, labID,
		fmt.Sprintf("%s=%s", config.PlatformLabel, config.Challenge),
		fmt.Sprintf("%s=%s", config.LabIDLabel, labID),
		fmt.Sprintf("%s=%s", config.ChallengeIDLabel, challengeID),
	)
	if err != nil {
		return fmt.Errorf("failed to get instances in namespace by selector: [%w]", err)
	}

	for _, dp := range dps {
		if err = s.infrastructure.ScaleDeployment(ctx, dp.Name, labID, 1); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to upscale deployment: [%w]", err))
		}
	}
	return
}

func (s *ChallengeService) StopChallenge(ctx context.Context, labID, challengeID string) (errs error) {
	dps, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, labID,
		fmt.Sprintf("%s=%s", config.PlatformLabel, config.Challenge),
		fmt.Sprintf("%s=%s", config.LabIDLabel, labID),
		fmt.Sprintf("%s=%s", config.ChallengeIDLabel, challengeID),
	)
	if err != nil {
		return fmt.Errorf("failed to get instances in namespace by selector: [%w]", err)
	}

	for _, dp := range dps {
		if err = s.infrastructure.ScaleDeployment(ctx, dp.Name, labID, 0); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to downscale deployment: [%w]", err))
		}
	}
	return
}

func (s *ChallengeService) ResetChallenge(ctx context.Context, labID, challengeID string) (errs error) {
	dps, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, labID,
		fmt.Sprintf("%s=%s", config.PlatformLabel, config.Challenge),
		fmt.Sprintf("%s=%s", config.LabIDLabel, labID),
		fmt.Sprintf("%s=%s", config.ChallengeIDLabel, challengeID),
	)
	if err != nil {
		return fmt.Errorf("failed to get instances in namespace by selector: [%w]", err)
	}

	for _, dp := range dps {
		if err = s.infrastructure.ResetDeployment(ctx, dp.Name, labID); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to reset deployment: [%w]", err))
		}
	}
	return
}
