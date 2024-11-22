package model

import (
	"github.com/cybericebox/lib/pkg/ipam"
	"github.com/gofrs/uuid"
)

type Lab struct {
	ID          uuid.UUID
	CIDRManager *ipam.IPAManager
}
