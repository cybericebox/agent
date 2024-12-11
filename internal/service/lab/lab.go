package lab

import (
	"context"
	"github.com/cybericebox/agent/internal/delivery/repository/postgres"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/lib/pkg/ipam"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-multierror"
	"net/netip"
)

type (
	IInfrastructure interface {
		ApplyNetwork(ctx context.Context, name, cidr string, blockSize int) error
		GetNetworkCIDR(ctx context.Context, name string) (string, error)
		DeleteNetwork(ctx context.Context, name string) error

		ApplyNamespace(ctx context.Context, name string, ipPoolName *string) error
		NamespaceExists(ctx context.Context, name string) (bool, error)
		DeleteNamespace(ctx context.Context, name string) error

		ApplyNetworkPolicy(ctx context.Context, labID string) error

		ScaleDeployment(ctx context.Context, name, namespace string, scale int32) error
		GetDeploymentsInNamespaceBySelector(ctx context.Context, namespace string, selector ...string) ([]model.DeploymentStatus, error)
	}

	IRepository interface {
		GetLaboratories(ctx context.Context) ([]postgres.Laboratory, error)
		CreateLaboratory(ctx context.Context, laboratory postgres.CreateLaboratoryParams) error
		DeleteLaboratory(ctx context.Context, id uuid.UUID) (int64, error)
	}

	iIPAManager interface {
		AcquireSingleIP(ctx context.Context, specificIP ...string) (string, error)
		ReleaseSingleIP(ctx context.Context, ip string) error
		AcquireChildCIDR(ctx context.Context, blockSize uint32) (*ipam.IPAManager, error)
		ReleaseChildCIDR(ctx context.Context, childCIDR string) error
		GetChildCIDR(ctx context.Context, cidr string) (*ipam.IPAManager, error)
	}

	iDNSService interface {
		CreateDNSServer(ctx context.Context, labID, ip string) error
		RefreshDNSRecords(ctx context.Context, labID string, records []model.DNSRecordConfig, isAddRecords bool) error
	}

	iChallengeService interface {
		CreateChallenge(ctx context.Context, lab *model.Lab, challengeConfig model.ChallengeConfig) ([]model.DNSRecordConfig, error)
		DeleteChallenge(ctx context.Context, lab *model.Lab, challengeId string) ([]model.DNSRecordConfig, error)
		StartChallenge(ctx context.Context, labID, challengeID string) (errs error)
		StopChallenge(ctx context.Context, labID, challengeID string) (errs error)
		ResetChallenge(ctx context.Context, labID, challengeID string) (errs error)
	}

	iLabService interface {
		iDNSService
		iChallengeService
	}

	LabService struct {
		infrastructure IInfrastructure
		ipaManager     iIPAManager
		service        iLabService
		repository     IRepository
	}

	Dependencies struct {
		Infrastructure IInfrastructure
		IPAManager     iIPAManager
		Service        iLabService
		Repository     IRepository
	}
)

func NewLabService(deps Dependencies) *LabService {
	return &LabService{infrastructure: deps.Infrastructure, ipaManager: deps.IPAManager, service: deps.Service, repository: deps.Repository}
}

func (s *LabService) RestoreLabIfNeeded(ctx context.Context, lab model.Lab) error {
	exists, err := s.infrastructure.NamespaceExists(ctx, lab.ID.String())
	if err != nil {
		return appError.ErrPlatform.WithError(err).WithMessage("Failed to check if namespace exists").WithContext("labID", lab.ID.String()).Err()
	}
	if !exists {
		// create the lab in the infrastructure
		if err = s.createSpecificLab(ctx, uint32(lab.CIDR.Bits()), lab.ID, lab.CIDR); err != nil {
			return appError.ErrPlatform.WithError(err).WithMessage("Failed to create lab in infrastructure").WithContext("labID", lab.ID.String()).Err()
		}
	}

	return nil
}

func (s *LabService) GetStoredLabs(ctx context.Context) ([]model.Lab, error) {
	labs, err := s.repository.GetLaboratories(ctx)
	if err != nil {
		return nil, appError.ErrPostgres.WithError(err).WithMessage("Failed to get laboratories from state").Err()
	}

	var errs error

	storedLabs := make([]model.Lab, 0, len(labs))
	for _, lab := range labs {
		CIDRManager, err := s.ipaManager.GetChildCIDR(ctx, lab.Cidr.String())
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to get child cidr").WithContext("labID", lab.ID.String()).Err())
			continue
		}
		storedLabs = append(storedLabs, model.Lab{
			ID:          lab.ID,
			CIDR:        lab.Cidr,
			CIDRManager: CIDRManager,
		})
	}

	if errs != nil {
		return nil, appError.ErrLab.WithError(errs).WithMessage("Failed to get stored labs").Err()
	}

	return storedLabs, nil
}

func (s *LabService) GetLab(ctx context.Context, labID string) (*model.Lab, error) {
	parsedID, err := uuid.FromString(labID)
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to parse lab id").WithContext("labID", labID).Err()
	}

	cidr, err := s.infrastructure.GetNetworkCIDR(ctx, labID)
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to get lab cidr").WithContext("labID", labID).Err()
	}

	parsedCIDR, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to parse cidr").WithContext("labID", labID).Err()
	}

	lab := &model.Lab{
		ID:   parsedID,
		CIDR: parsedCIDR,
	}

	lab.CIDRManager, err = s.ipaManager.GetChildCIDR(ctx, cidr)
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to get child cidr").WithContext("labID", labID).Err()
	}

	return lab, nil
}

func (s *LabService) CreateLab(ctx context.Context, subnetMask uint32) (*model.Lab, error) {
	var err error

	lab := &model.Lab{
		ID: uuid.Must(uuid.NewV7()),
	}

	lab.CIDRManager, err = s.ipaManager.AcquireChildCIDR(ctx, subnetMask)
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to acquire child cidr").WithContext("subnetMask", subnetMask).Err()
	}

	// create network
	if err = s.infrastructure.ApplyNetwork(ctx, lab.ID.String(), lab.CIDRManager.GetCIDR(), int(subnetMask)); err != nil {
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply network").Err()
		}
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to apply network").WithContext("labID", lab.ID.String()).Err()
	}

	labPool := lab.ID.String()
	// create namespace
	if err = s.infrastructure.ApplyNamespace(ctx, lab.ID.String(), &labPool); err != nil {
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in apply namespace").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply namespace").WithContext("labID", lab.ID.String()).Err()
		}
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to apply namespace").WithContext("labID", lab.ID.String()).Err()
	}

	// set network policy
	if err = s.infrastructure.ApplyNetworkPolicy(ctx, lab.ID.String()); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to apply network policy").WithContext("labID", lab.ID.String()).Err()
	}

	singleIP, err := lab.CIDRManager.AcquireSingleIP(ctx)
	if err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to acquire single ip").WithContext("labID", lab.ID.String()).Err()
	}

	if err = s.service.CreateDNSServer(ctx, lab.ID.String(), singleIP); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to create dns server").WithContext("labID", lab.ID.String()).Err()
	}

	// save lab to db

	cidr, err := netip.ParsePrefix(lab.CIDRManager.GetCIDR())
	if err != nil {
		return nil, appError.ErrLab.WithError(err).WithMessage("Failed to parse cidr").WithContext("labID", lab.ID.String()).Err()
	}

	if err = s.repository.CreateLaboratory(ctx, postgres.CreateLaboratoryParams{
		ID:   lab.ID,
		Cidr: cidr,
	}); err != nil {
		if err1 := s.DeleteLab(ctx, lab.ID.String()); err1 != nil {
			return nil, appError.ErrLab.WithError(err1).WithMessage("Failed to delete lab in create lab").WithContext("labID", lab.ID.String()).Err()
		}
		return nil, appError.ErrLab.WithWrappedError(appError.ErrPostgres.WithError(err)).WithMessage("Failed to create lab in db").WithContext("labID", lab.ID.String()).Err()
	}

	return lab, nil
}

func (s *LabService) createSpecificLab(ctx context.Context, subnetMask uint32, labID uuid.UUID, cidr netip.Prefix) error {
	var err error

	lab := &model.Lab{
		ID: labID,
	}

	lab.CIDRManager, err = s.ipaManager.GetChildCIDR(ctx, cidr.String())
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get child cidr").WithContext("labID", labID.String()).Err()
	}

	// create network
	if err = s.infrastructure.ApplyNetwork(ctx, lab.ID.String(), lab.CIDRManager.GetCIDR(), int(subnetMask)); err != nil {
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply network").WithContext("labID", lab.ID.String()).Err()
		}
		return appError.ErrLab.WithError(err).WithMessage("Failed to apply network").WithContext("labID", lab.ID.String()).Err()
	}

	labPool := lab.ID.String()
	// create namespace
	if err = s.infrastructure.ApplyNamespace(ctx, lab.ID.String(), &labPool); err != nil {
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in apply namespace").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply namespace").WithContext("labID", lab.ID.String()).Err()
		}
		return appError.ErrLab.WithError(err).WithMessage("Failed to apply namespace").WithContext("labID", lab.ID.String()).Err()
	}

	// set network policy
	if err = s.infrastructure.ApplyNetworkPolicy(ctx, lab.ID.String()); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in apply network policy").WithContext("labID", lab.ID.String()).Err()
		}
		return appError.ErrLab.WithError(err).WithMessage("Failed to apply network policy").WithContext("labID", lab.ID.String()).Err()
	}

	singleIP, err := lab.CIDRManager.GetFirstIP()
	if err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		return appError.ErrLab.WithError(err).WithMessage("Failed to acquire single ip").WithContext("labID", lab.ID.String()).Err()
	}

	if err = s.service.CreateDNSServer(ctx, lab.ID.String(), singleIP); err != nil {
		if err1 := s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete namespace in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to delete network in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		if err1 := s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err1 != nil {
			return appError.ErrLab.WithError(err1).WithMessage("Failed to release child cidr in acquire single ip").WithContext("labID", lab.ID.String()).Err()
		}
		return appError.ErrLab.WithError(err).WithMessage("Failed to create dns server").WithContext("labID", lab.ID.String()).Err()
	}

	return nil
}

func (s *LabService) DeleteLab(ctx context.Context, labID string) error {
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get lab").WithContext("labID", labID).Err()
	}

	if err = s.infrastructure.DeleteNamespace(ctx, lab.ID.String()); err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to delete namespace").WithContext("labID", labID).Err()
	}

	if err = s.infrastructure.DeleteNetwork(ctx, lab.ID.String()); err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to delete network").WithContext("labID", labID).Err()
	}

	if err = s.ipaManager.ReleaseChildCIDR(ctx, lab.CIDRManager.GetCIDR()); err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to release child cidr").WithContext("labID", labID).Err()
	}

	// delete lab from db
	if _, err = s.repository.DeleteLaboratory(ctx, lab.ID); err != nil {
		return appError.ErrLab.WithWrappedError(appError.ErrPostgres.WithError(err)).WithMessage("Failed to delete lab from db").WithContext("labID", labID).Err()
	}

	return nil
}

func (s *LabService) StartLab(ctx context.Context, labID string) error {
	// get all deployments in the lab
	deployments, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, labID)
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get deployments in namespace by selector").WithContext("labID", labID).Err()
	}

	var errs error

	// scale all deployments to 1
	for _, deployment := range deployments {
		if err = s.infrastructure.ScaleDeployment(ctx, deployment.Name, labID, 1); err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to scale deployment").WithContext("labID", labID).WithContext("deploymentName", deployment.Name).Err())
		}
	}

	if errs != nil {
		return appError.ErrLab.WithError(errs).WithMessage("Failed to start lab").WithContext("labID", labID).Err()
	}

	return nil
}

func (s *LabService) StopLab(ctx context.Context, labID string) error {
	// get all deployments in the lab
	deployments, err := s.infrastructure.GetDeploymentsInNamespaceBySelector(ctx, labID)
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get deployments in namespace by selector").WithContext("labID", labID).Err()
	}

	var errs error

	// scale all deployments to 0
	for _, deployment := range deployments {
		if err = s.infrastructure.ScaleDeployment(ctx, deployment.Name, labID, 0); err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to scale deployment").WithContext("labID", labID).WithContext("deploymentName", deployment.Name).Err())
		}
	}

	if errs != nil {
		return appError.ErrLab.WithError(errs).WithMessage("Failed to stop lab").WithContext("labID", labID).Err()
	}

	return nil
}

// challenge methods

func (s *LabService) AddLabChallenges(ctx context.Context, labID string, challengeConfigs []model.ChallengeConfig) (errs error) {
	// get lab by cidr
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get lab").WithContext("labID", labID).Err()
	}

	labRecords := make([]model.DNSRecordConfig, 0)
	for _, challengeConfig := range challengeConfigs {
		records, err := s.service.CreateChallenge(ctx, lab, challengeConfig)
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to create challenge").WithContext("labID", labID).WithContext("challengeID", challengeConfig.ID).Err())
			continue
		}

		labRecords = append(labRecords, records...)
	}

	if err = s.service.RefreshDNSRecords(ctx, lab.ID.String(), labRecords, true); err != nil {
		errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to refresh DNS records").WithContext("labID", labID).Err())
	}
	return errs
}

func (s *LabService) DeleteLabChallenges(ctx context.Context, labID string, challengeIDs []string) (errs error) {
	lab, err := s.GetLab(ctx, labID)
	if err != nil {
		return appError.ErrLab.WithError(err).WithMessage("Failed to get lab").WithContext("labID", labID).Err()
	}

	labRecords := make([]model.DNSRecordConfig, 0)
	for _, challengeId := range challengeIDs {
		records, err := s.service.DeleteChallenge(ctx, lab, challengeId)
		if err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to delete challenge").WithContext("labID", labID).WithContext("challengeID", challengeId).Err())
		}

		labRecords = append(labRecords, records...)
	}

	if err = s.service.RefreshDNSRecords(ctx, lab.ID.String(), labRecords, false); err != nil {
		errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to refresh DNS records").WithContext("labID", labID).Err())
	}

	return errs
}

func (s *LabService) StartLabChallenges(ctx context.Context, labID string, challengeIDs []string) (errs error) {
	for _, challengeID := range challengeIDs {
		if err := s.service.StartChallenge(ctx, labID, challengeID); err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to start challenge").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
		}
	}

	if errs != nil {
		return appError.ErrLab.WithError(errs).WithMessage("Failed to start lab challenges").WithContext("labID", labID).Err()
	}

	return nil
}

func (s *LabService) StopLabChallenges(ctx context.Context, labID string, challengeIDs []string) (errs error) {
	for _, challengeID := range challengeIDs {
		if err := s.service.StopChallenge(ctx, labID, challengeID); err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to stop challenge").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
		}
	}

	if errs != nil {
		return appError.ErrLab.WithError(errs).WithMessage("Failed to stop lab challenges").WithContext("labID", labID).Err()
	}

	return nil
}

func (s *LabService) ResetLabChallenges(ctx context.Context, labID string, challengeIDs []string) (errs error) {
	for _, challengeID := range challengeIDs {
		if err := s.service.ResetChallenge(ctx, labID, challengeID); err != nil {
			errs = multierror.Append(errs, appError.ErrLab.WithError(err).WithMessage("Failed to reset challenge").WithContext("labID", labID).WithContext("challengeID", challengeID).Err())
		}
	}

	if errs != nil {
		return appError.ErrLab.WithError(errs).WithMessage("Failed to reset lab challenges").WithContext("labID", labID).Err()
	}

	return nil
}
