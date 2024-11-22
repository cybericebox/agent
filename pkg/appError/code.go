package appError

import "github.com/cybericebox/lib/pkg/err"

// Object codes
const (
	platformObjectCode = iota
	postgresObjectCode
	gRPCObjectCode
	kubernetesObjectCode
	labObjectCode
	labChallengeObjectCode
	labDNSObjectCode
)

// base object errors
var (
	ErrPlatform     = err.ErrInternal.WithObjectCode(platformObjectCode)
	ErrPostgres     = err.ErrInternal.WithObjectCode(postgresObjectCode)
	ErrKubernetes   = err.ErrInternal.WithObjectCode(kubernetesObjectCode)
	ErrLab          = err.ErrInternal.WithObjectCode(labObjectCode)
	ErrLabChallenge = err.ErrInternal.WithObjectCode(labChallengeObjectCode)
	ErrLabDNS       = err.ErrInternal.WithObjectCode(labDNSObjectCode)
)
