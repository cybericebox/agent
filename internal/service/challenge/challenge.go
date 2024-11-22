package challenge

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/tools"
	"github.com/cybericebox/agent/pkg/appError"
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
		ex, err := s.infrastructure.DeploymentExists(ctx, inst.ID, lab.ID.String())
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to check if deployment exists").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			continue
		}

		if ex {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithMessage("Deployment already exists").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			continue
		}

		ip, err := lab.CIDRManager.AcquireSingleIP(ctx)
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to acquire ip for instance").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			continue
		}

		dns, err := lab.CIDRManager.GetFirstIP()
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to get dns ip for instance").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			if err = lab.CIDRManager.ReleaseSingleIP(ctx, ip); err != nil {
				errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to release ip for instance in get dns: [%w]").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			}
			continue
		}

		if err = s.infrastructure.ApplyDeployment(ctx, model.ApplyDeploymentConfig{
			Name:  inst.ID,
			LabID: lab.ID.String(),
			Labels: map[string]string{
				config.PlatformLabel:    config.Challenge,
				config.LabIDLabel:       lab.ID.String(),
				config.ChallengeIDLabel: challengeConfig.ID,
				config.InstanceIDLabel:  inst.ID,
				config.RecordsListLabel: tools.RecordsToStr(inst.Records),
			},
			Image:        inst.Image,
			IP:           ip,
			DNS:          dns,
			ReplicaCount: 1,
			UsePublicDNS: true,
			Resources:    inst.Resources,
			Envs:         inst.Envs,
		}); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to apply deployment").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
			if err = lab.CIDRManager.ReleaseSingleIP(ctx, ip); err != nil {
				errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to release ip for instance in apply deployment: [%w]").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeConfig.ID).WithContext("instanceID", inst.ID).Err())
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
		return nil, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to get instances in namespace by selector").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeID).Err()
	}

	records = make([]model.DNSRecordConfig, 0)

	for _, dp := range dps {
		if err = s.infrastructure.DeleteDeployment(ctx, dp.Name, lab.ID.String()); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to delete deployment").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeID).Err())
		}

		if err = lab.CIDRManager.ReleaseSingleIP(ctx, dp.IP); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to release ip for instance").WithContext("labID", lab.ID.String()).WithContext("challengeID", challengeID).Err())
		}

		records = append(records, tools.RecordsFromStr(dp.Labels[config.RecordsListLabel])...)
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
		return appError.ErrLabChallenge.WithError(err).WithMessage("Failed to get instances in namespace by selector").WithContext("labID", labID).WithContext("challengeID", challengeID).Err()
	}

	for _, dp := range dps {
		if err = s.infrastructure.ScaleDeployment(ctx, dp.Name, labID, 1); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to upscale deployment").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
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
		return appError.ErrLabChallenge.WithError(err).WithMessage("Failed to get instances in namespace by selector").WithContext("labID", labID).WithContext("challengeID", challengeID).Err()
	}

	for _, dp := range dps {
		if err = s.infrastructure.ScaleDeployment(ctx, dp.Name, labID, 0); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to downscale deployment").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
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
		return appError.ErrLabChallenge.WithError(err).WithMessage("Failed to get instances in namespace by selector").WithContext("labID", labID).WithContext("challengeID", challengeID).Err()
	}

	for _, dp := range dps {
		if err = s.infrastructure.ResetDeployment(ctx, dp.Name, labID); err != nil {
			errs = multierror.Append(errs, appError.ErrLabChallenge.WithError(err).WithMessage("Failed to reset deployment").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
		}
	}
	return
}
