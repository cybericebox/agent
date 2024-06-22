package dns

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/internal/service/helper"
	"github.com/hashicorp/go-multierror"
	"slices"
	"text/template"
)

const (
	dnsName       = "dns-server"
	dnsConfigName = "dns-config"

	image = "coredns/coredns:1.10.0"

	coreFile = "Corefile"
	zoneFile = "zonefile"

	coreFileContent = `. {
    file zonefile
    prometheus     # enable metrics
    errors         # show errors
    log            # enable query logs
}
`
	zonePrefixContent = `$ORIGIN .
@   3600 IN SOA sns.dns.icann.org. noc.dns.icann.org. (
                2017042745 ; serial
                7200       ; refresh (2 hours)
                3600       ; retry (1 hour)
                1209600    ; expire (2 weeks)
                3600       ; minimum (1 hour)
                )

{{range .}}{{.Name}} IN {{.Type}} {{.Data}}
{{end}}
`
)

type (
	Infrastructure interface {
		ApplyDeployment(ctx context.Context, config model.ApplyDeploymentConfig) error
		ResetDeployment(ctx context.Context, name, namespace string) error

		ApplyConfigMap(ctx context.Context, name, namespace string, data map[string]string) error
		GetConfigMapData(ctx context.Context, name, namespace string) (map[string]string, error)
	}
	DNSService struct {
		infrastructure Infrastructure
	}

	DNSServer struct {
		labID          string
		records        []model.DNSRecordConfig
		infrastructure Infrastructure
	}
)

func NewDNSService(infrastructure Infrastructure) *DNSService {
	return &DNSService{
		infrastructure: infrastructure,
	}
}

// CreateDNSServer creates a new DNS server for the lab
func (dns *DNSService) CreateDNSServer(ctx context.Context, labID, ip string) error {

	newDNS := newDNSServer(dns.infrastructure, labID)

	if err := newDNS.setConfig(ctx); err != nil {
		return err
	}

	return dns.infrastructure.ApplyDeployment(ctx, model.ApplyDeploymentConfig{
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
				Memory: "50Mi",
				CPU:    "300m",
			},
			Limit: model.ResourceConfig{
				Memory: "50Mi",
				CPU:    "300m",
			},
		},
		Args: []string{"-conf", fmt.Sprintf("/%s", coreFile)},
		Volumes: []model.Volume{{
			Name:          dnsName,
			ConfigMapName: dnsConfigName,
			Mounts: []model.Mount{
				{
					MountPath: fmt.Sprintf("/%s", coreFile),
					SubPath:   coreFile,
				},
				{
					MountPath: fmt.Sprintf("/%s", zoneFile),
					SubPath:   zoneFile,
				}},
		}},
	})
}

func (dns *DNSService) RefreshDNSRecords(ctx context.Context, labId string, records []model.DNSRecordConfig, isAddingRecords bool) error {
	newDNS := newDNSServer(dns.infrastructure, labId)

	if err := newDNS.getRecords(ctx); err != nil {
		return err
	}

	if isAddingRecords {
		if err := newDNS.addRecords(records); err != nil {
			return err
		}
	} else {
		if err := newDNS.deleteRecords(records); err != nil {
			return err
		}
	}

	if err := newDNS.setConfig(ctx); err != nil {
		return err
	}

	return newDNS.reset(ctx)
}

// newDNSServer creates a new DNS server instance

func newDNSServer(infrastructure Infrastructure, labId string) *DNSServer {
	return &DNSServer{
		labID:          labId,
		infrastructure: infrastructure,
		records:        make([]model.DNSRecordConfig, 0),
	}
}

func (dns *DNSServer) generateZoneConfig() (string, error) {
	var tpl bytes.Buffer

	t, err := template.New("config").Parse(zonePrefixContent)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&tpl, dns.records)
	if err != nil {
		panic(err)
	}
	return tpl.String(), nil
}

func (dns *DNSServer) setConfig(ctx context.Context) error {

	cfg, err := dns.generateZoneConfig()

	if err != nil {
		return err
	}

	return dns.infrastructure.ApplyConfigMap(ctx, dnsConfigName, dns.labID, map[string]string{
		coreFile:                coreFileContent,
		zoneFile:                cfg,
		config.RecordsListLabel: helper.RecordsToStr(dns.records),
	})
}

func (dns *DNSServer) getRecords(ctx context.Context) error {
	data, err := dns.infrastructure.GetConfigMapData(ctx, dnsConfigName, dns.labID)
	if err != nil {
		return err
	}

	if len(data[config.RecordsListLabel]) == 0 {
		return nil
	}

	dns.records = helper.RecordsFromStr(data[config.RecordsListLabel])

	return nil
}

func (dns *DNSServer) reset(ctx context.Context) error {
	return dns.infrastructure.ResetDeployment(ctx, dnsName, dns.labID)
}

func (dns *DNSServer) addRecords(records []model.DNSRecordConfig) error {
	var errs error

	for _, r := range records {
		if slices.Contains(dns.records, r) {
			errs = multierror.Append(errs, fmt.Errorf("record for %s IN %s %s already exists", r.Name, r.Type, r.Data))
		} else {
			dns.records = append(dns.records, r)
		}
	}

	return errs
}

func (dns *DNSServer) deleteRecords(records []model.DNSRecordConfig) error {
	var errs error

	for _, r := range records {
		if slices.Contains(dns.records, r) {
			errs = multierror.Append(errs, fmt.Errorf("record for %s IN %s %s already exists", r.Name, r.Type, r.Data))
		} else {
			dns.records = append(dns.records, r)
		}
	}

	return errs
}
