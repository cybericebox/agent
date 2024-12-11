package dns

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/tools"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/hashicorp/go-multierror"
	"slices"
	"strconv"
	"text/template"
	"time"
)

const (
	dnsName       = "dns-server"
	dnsConfigName = "dns-config"
	dnsConfigPath = "/dns"

	image = "cybericebox/coredns:1.12.0"

	coreFile = "Corefile"
	zoneFile = "zonefile"

	coreFileContent = `. {
errors         # show errors
file /dns/zonefile {
	reload 5s	
}
alternate original NXDOMAIN . 8.8.8.8 1.1.1.1
}
`
	zonePrefixContent = `$ORIGIN .
@   3600 IN SOA sns.dns.icann.org. noc.dns.icann.org. (
                {{.Serial}} ; serial
                7200       ; refresh (2 hours)
                3600       ; retry (1 hour)
                1209600    ; expire (2 weeks)
                3600       ; minimum (1 hour)
                )

{{range .Records}}{{.Name}} IN {{.Type}} {{.Data}}
{{end}}
`
)

type (
	zoneConfig struct {
		Serial  string
		Records []model.DNSRecordConfig
	}

	IInfrastructure interface {
		ApplyDeployment(ctx context.Context, config model.ApplyDeploymentConfig) error
		ResetDeployment(ctx context.Context, name, namespace string) error

		ApplyConfigMap(ctx context.Context, name, namespace string, data map[string]string) error
		GetConfigMapData(ctx context.Context, name, namespace string) (map[string]string, error)
	}
	DNSService struct {
		infrastructure IInfrastructure
	}

	DNSServer struct {
		labID          string
		records        []model.DNSRecordConfig
		infrastructure IInfrastructure
	}
)

func NewDNSService(infrastructure IInfrastructure) *DNSService {
	return &DNSService{
		infrastructure: infrastructure,
	}
}

// CreateDNSServer creates a new DNS server for the lab
func (dns *DNSService) CreateDNSServer(ctx context.Context, labID, ip string) error {
	labDNSErr := appError.ErrLabDNS.WithContext("labID", labID)
	newDNS := newDNSServer(dns.infrastructure, labID)

	if err := newDNS.setConfig(ctx); err != nil {
		return labDNSErr.WithError(err).WithMessage("Failed to set config").Err()
	}

	if err := dns.infrastructure.ApplyDeployment(ctx, model.ApplyDeploymentConfig{
		Name:  dnsName,
		LabID: labID,
		Image: image,
		IP:    ip,
		Labels: map[string]string{
			config.PlatformLabel: config.LabDNSServer,
			config.LabIDLabel:    labID,
		},
		Resources: model.ResourcesConfig{
			Requests: model.ResourceConfig{
				Memory: 50 * 1024 * 1024,
				CPU:    10,
			},
			Limit: model.ResourceConfig{
				Memory: 50 * 1024 * 1024,
				CPU:    10,
			},
		},
		ReplicaCount: 1,
		Args:         []string{"-conf", fmt.Sprintf("%s/%s", dnsConfigPath, coreFile)},
		Volumes: []model.Volume{{
			Name:          dnsName,
			ConfigMapName: dnsConfigName,
			MountPath:     dnsConfigPath,
		}},
	}); err != nil {
		return labDNSErr.WithError(err).WithMessage("Failed to apply deployment").Err()
	}

	return nil
}

func (dns *DNSService) RefreshDNSRecords(ctx context.Context, labID string, records []model.DNSRecordConfig, isAddingRecords bool) error {
	labDNSErr := appError.ErrLabDNS.WithContext("labID", labID)

	newDNS := newDNSServer(dns.infrastructure, labID)

	if err := newDNS.getRecords(ctx); err != nil {
		return labDNSErr.WithError(err).WithMessage("Failed to get records").Err()
	}

	if isAddingRecords {
		if err := newDNS.addRecords(records); err != nil {
			return labDNSErr.WithError(err).WithMessage("Failed to add records").Err()
		}
	} else {
		if err := newDNS.deleteRecords(records); err != nil {
			return labDNSErr.WithError(err).WithMessage("Failed to delete records").Err()
		}
	}

	if err := newDNS.setConfig(ctx); err != nil {
		return labDNSErr.WithError(err).WithMessage("Failed to set config").Err()
	}

	return nil
}

// newDNSServer creates a new DNS server instance

func newDNSServer(infrastructure IInfrastructure, labID string) *DNSServer {
	return &DNSServer{
		labID:          labID,
		infrastructure: infrastructure,
		records:        make([]model.DNSRecordConfig, 0),
	}
}

func (dns *DNSServer) generateZoneConfig() (string, error) {
	var tpl bytes.Buffer

	t, err := template.New("config").Parse(zonePrefixContent)
	if err != nil {
		return "", appError.ErrLabDNS.WithError(err).WithContext("labID", dns.labID).WithMessage("Failed to parse template").Err()
	}
	err = t.Execute(&tpl, zoneConfig{
		// serial as time now
		Serial:  strconv.Itoa(int(time.Now().Unix())),
		Records: dns.records,
	})
	if err != nil {
		return "", appError.ErrLabDNS.WithError(err).WithContext("labID", dns.labID).WithMessage("Failed to execute template").Err()
	}
	return tpl.String(), nil
}

func (dns *DNSServer) setConfig(ctx context.Context) error {

	cfg, err := dns.generateZoneConfig()

	if err != nil {
		return appError.ErrLabDNS.WithError(err).WithContext("labID", dns.labID).WithMessage("Failed to generate zone config").Err()
	}

	if err = dns.infrastructure.ApplyConfigMap(ctx, dnsConfigName, dns.labID, map[string]string{
		coreFile:                coreFileContent,
		zoneFile:                cfg,
		config.RecordsListLabel: tools.RecordsToStr(dns.records),
	}); err != nil {
		return appError.ErrLabDNS.WithError(err).WithContext("labID", dns.labID).WithMessage("Failed to apply config map").Err()
	}

	return nil
}

func (dns *DNSServer) getRecords(ctx context.Context) error {
	data, err := dns.infrastructure.GetConfigMapData(ctx, dnsConfigName, dns.labID)
	if err != nil {
		return appError.ErrLabDNS.WithError(err).WithContext("labID", dns.labID).WithMessage("Failed to get config map data").Err()
	}

	if len(data[config.RecordsListLabel]) == 0 {
		return nil
	}

	dns.records = tools.RecordsFromStr(data[config.RecordsListLabel])

	return nil
}

func (dns *DNSServer) addRecords(records []model.DNSRecordConfig) error {
	var errs error

	for _, r := range records {
		if slices.Contains(dns.records, r) {
			errs = multierror.Append(errs, appError.ErrLabDNS.WithError(errs).WithContext("labID", dns.labID).WithMessage(fmt.Sprintf("record for %s IN %s %s already exists", r.Name, r.Type, r.Data)).Err())
		} else {
			dns.records = append(dns.records, r)
		}
	}

	if errs != nil {
		return appError.ErrLabDNS.WithError(errs).WithContext("labID", dns.labID).WithMessage("Failed to add records").Err()
	}

	return nil
}

func (dns *DNSServer) deleteRecords(records []model.DNSRecordConfig) error {
	var errs error

	newRecords := make([]model.DNSRecordConfig, 0)

	for _, r := range dns.records {
		if !recordContains(records, r) {
			newRecords = append(newRecords, r)
		}
	}

	// Check if all records to delete are present
	for _, r := range records {
		if !recordContains(records, r) {
			errs = multierror.Append(errs, appError.ErrLabDNS.WithError(errs).WithContext("labID", dns.labID).WithMessage(fmt.Sprintf("record for %s IN %s %s does not exist", r.Name, r.Type, r.Data)).Err())
		}
	}

	if errs != nil {
		return appError.ErrLabDNS.WithError(errs).WithContext("labID", dns.labID).WithMessage("Failed to delete records").Err()
	}

	dns.records = newRecords

	return nil
}

func recordContains(records []model.DNSRecordConfig, record model.DNSRecordConfig) bool {
	for _, r := range records {
		if r.Name == record.Name && r.Type == record.Type {
			return true
		}
	}
	return false
}
