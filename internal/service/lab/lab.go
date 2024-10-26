package lab

import (
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/wireguard/pkg/ipam"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-multierror"
)

type (
	Infrastructure interface {
		ApplyNetwork(ctx context.Context, name, cidr string, blockSize int) error
		GetNetworkCIDR(ctx context.Context, name string) (string, error)
		DeleteNetwork(ctx context.Context, name string) error

		ApplyNamespace(ctx context.Context, name string, ipPoolName *string) error
		NamespaceExists(ctx context.Context, name string) (bool, error)
		DeleteNamespace(ctx context.Context, name string) error

		ApplyNetworkPolicy(ctx context.Context, labID string) error
	}

	ipaManager interface {
		AcquireSingleIP(ctx context.Context, specificIP ...string) (string, error)
		ReleaseSingleIP(ctx context.Context, ip string) error
		AcquireChildCIDR(ctx context.Context, blockSize uint32) (*ipam.IPAManager, error)
		ReleaseChildCIDR(ctx context.Context, childCIDR string) error
		GetChildCIDR(ctx context.Context, cidr string) (*ipam.IPAManager, error)
	}

	dnsService interface {
		CreateDNSServer(ctx context.Context, labId, ip string) error
		RefreshDNSRecords(ctx context.Context, labId string, records []model.DNSRecordConfig, isAddRecords bool) error
	}

	challengeService interface {
		CreateChallenge(ctx context.Context, lab *model.Lab, challengeConfig model.ChallengeConfig) ([]model.DNSRecordConfig, error)
		DeleteChallenge(ctx context.Context, lab *model.Lab, challengeId string) ([]model.DNSRecordConfig, error)
	}

	service interface {
		dnsService
		challengeService
	}

	LabService struct {
		infrastructure Infrastructure
		ipaManager     ipaManager
		service        service
	}

	Dependencies struct {
		Infrastructure Infrastructure
		IPAManager     ipaManager
		Service        service
	}
)

func NewLabService(deps Dependencies) *LabService {
	return &LabService{infrastructure: deps.Infrastructure, ipaManager: deps.IPAManager, service: deps.Service}
}

func (s *LabService) GetLab(ctx context.Context, labID string) (*model.Lab, error) {
	parsedID, err := uuid.FromString(labID)
	if err != nil {
		return nil, fmt.Errorf("invalid lab id: [%w]", err)
	}

	cidr, err := s.infrastructure.GetNetworkCIDR(ctx, labID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lab cidr: [%w]", err)
	}

	lab := &model.Lab{
		ID: parsedID,
	}

	lab.CIDRManager, err = s.ipaManager.GetChildCIDR(ctx, cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to get lab cidr manager: [%w]", err)
	}

	return lab, nil
}

func (s *LabService) CreateLab(ctx context.Context, subnetMask uint32) (string, error) {
	var err error

	lab := &model.Lab{
		ID: uuid.Must(uuid.NewV7()),
	}

	lab.CIDRManager, err = s.ipaManager.AcquireChildCIDR(ctx, subnetMask)
	if err != nil {
		return "", fmt.Errorf("failed to acquire child cidr: [%w]", err)
	}

	// create network
	if err = s.infrastructure.ApplyNetwork(ctx, lab.ID.String(), lab.CIDRManager.GetCIDR(), int(subnetMask)); err != nil {
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return "", fmt.Errorf("failed to release child cidr in apply network: [%w]", err1)
		}
		return "", fmt.Errorf("failed to apply network: [%w]", err)
	}

	labPool := lab.ID.String()
	// create namespace
	if err = s.infrastructure.ApplyNamespace(ctx, lab.ID.String(), &labPool); err != nil {
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete network in apply namespace: [%w]", err1)
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return "", fmt.Errorf("failed to release child cidr in apply namespace: [%w]", err1)
		}
		return "", fmt.Errorf("failed to apply namespace: [%w]", err)
	}

	// set network policy
	if err = s.infrastructure.ApplyNetworkPolicy(ctx, lab.ID.String()); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete namespace in apply network policy: [%w]", err1)
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete network in apply network policy: [%w]", err1)
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return "", fmt.Errorf("failed to release child cidr in apply network policy: [%w]", err1)
		}
		return "", fmt.Errorf("failed to apply network policy: [%w]", err)
	}

	singleIP, err := lab.CIDRManager.AcquireSingleIP(ctx)
	if err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete namespace in acquire single ip: [%w]", err1)
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete network in acquire single ip: [%w]", err1)
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return "", fmt.Errorf("failed to release child cidr in acquire single ip: [%w]", err1)
		}
		return "", fmt.Errorf("failed to acquire single ip: [%w]", err)
	}

	if err = s.service.CreateDNSServer(ctx, lab.ID.String(), singleIP); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete namespace in acquire single ip: [%w]", err1)
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return "", fmt.Errorf("failed to delete network in acquire single ip: [%w]", err1)
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return "", fmt.Errorf("failed to release child cidr in acquire single ip: [%w]", err1)
		}
	}

	return lab.ID.String(), nil
}

func (s *LabService) DeleteLab(ctx context.Context, labID string) error {
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return fmt.Errorf("failed to get lab: [%w]", err)
	}

	if err = s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err != nil {
		return fmt.Errorf("failed to delete namespace: [%w]", err)
	}

	if err = s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err != nil {
		return fmt.Errorf("failed to delete network: [%w]", err)
	}

	if err = s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err != nil {
		return fmt.Errorf("failed to release child cidr: [%w]", err)
	}
	return nil
}

func (s *LabService) AddLabChallenges(ctx context.Context, labID string, challengeConfigs []model.ChallengeConfig) (errs error) {
	// get lab by cidr
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return fmt.Errorf("failed to get lab: [%w]", err)
	}

	labRecords := make([]model.DNSRecordConfig, 0)
	for _, challengeConfig := range challengeConfigs {
		records, err := s.service.CreateChallenge(ctx, lab, challengeConfig)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to create challenge: [%w]", err))
			continue
		}

		labRecords = append(labRecords, records...)
	}

	if err = s.service.RefreshDNSRecords(ctx, lab.ID.String(), labRecords, true); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to refresh DNS records: [%w]", err))
	}
	return errs
}

func (s *LabService) DeleteLabChallenges(ctx context.Context, labID string, challengeIds []string) (errs error) {
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return fmt.Errorf("failed to get lab: [%w]", err)
	}

	labRecords := make([]model.DNSRecordConfig, 0)
	for _, challengeId := range challengeIds {
		records, err := s.service.DeleteChallenge(ctx, lab, challengeId)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to delete challenge: [%w]", err))
		}

		labRecords = append(labRecords, records...)
	}

	if err = s.service.RefreshDNSRecords(ctx, lab.ID.String(), labRecords, false); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to refresh DNS records: [%w]", err))
	}

	return errs
}
