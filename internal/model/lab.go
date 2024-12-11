package model

import (
	"github.com/cybericebox/lib/pkg/ipam"
	"github.com/gofrs/uuid"
	"net/netip"
)

const (
	// Statuses
	StatusUnknown = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusStopped
	StatusError
)

type (
	Status int

	Lab struct {
		ID          uuid.UUID
		CIDRManager *ipam.IPAManager
		CIDR        netip.Prefix
	}
	LabStatus struct {
		ID        uuid.UUID
		CIDR      string
		DNS       *DNSStatus
		Instances []InstanceStatus
	}

	InstanceStatus struct {
		ID          uuid.UUID
		ChallengeID uuid.UUID
		Status      Status
		Resources   ResourceConfig
		Reason      string
	}

	DNSStatus struct {
		Status    Status
		Resources ResourceConfig
		Reason    string
	}

	PlatformDynamicObject struct {
		Status    Status
		Resources ResourcesConfig
		Reason    string
		Labels    map[string]string
	}
)
