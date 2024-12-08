package client

import (
	"context"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	NoTokenErrMsg      = "token contains an invalid number of segments"
	UnauthorizedErrMsg = "unauthorized"
	AuthKey            = "authKey"
)

type (
	Credentials struct {
		Token    string
		Insecure bool
	}

	Config struct {
		Endpoint string
		Auth     Auth
		TLS      TLS
	}

	Auth struct {
		AuthKey string
		SignKey string
	}

	TLS struct {
		Enabled  bool
		CertFile string
		CertKey  string
		CaFile   string
	}

	AgentClient interface {
		Close() error
		protobuf.AgentClient
	}

	agentClient struct {
		protobuf.AgentClient
		connection *grpc.ClientConn
	}
)

func (c Credentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"token": c.Token,
	}, nil
}

func (c Credentials) RequireTransportSecurity() bool {
	return !c.Insecure
}

func getCredentials(conf TLS) (credentials.TransportCredentials, error) {
	log.Debug().Msg("Preparing credentials for RPC")
	if conf.Enabled {
		creds, err := credentials.NewServerTLSFromFile(conf.CertFile, conf.CertKey)
		if err != nil {
			return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to create server credentials").Err()
		}
		return creds, nil
	} else {
		return insecure.NewCredentials(), nil
	}
}

func translateRPCErr(err error) error {
	st, ok := status.FromError(err)
	if ok {
		msg := st.Message()
		switch {
		case UnauthorizedErrMsg == msg:
			return appError.ErrGRPCUnauthenticated.WithError(err).Err()

		case NoTokenErrMsg == msg:
			return appError.ErrGRPCUnauthenticated.WithError(err).Err()

		}

		return appError.ErrGRPC.WithError(err).WithMessage("Failed to perform RPC").Err()
	}

	return appError.ErrGRPC.WithError(err).WithMessage("Failed to perform RPC").Err()
}

func constructAuthCredentials(authKey, signKey string) (Credentials, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		AuthKey: authKey,
	})
	tokenString, err := token.SignedString([]byte(signKey))
	if err != nil {
		return Credentials{}, translateRPCErr(err)
	}
	authCreds := Credentials{Token: tokenString}
	return authCreds, nil
}

// NewAgentConnection creates a new connection to the agent service and returns the client and a function to close the connection
func NewAgentConnection(config Config) (AgentClient, error) {
	log.Debug().Str("url", config.Endpoint).Msg("Connecting to agent")

	authCreds, err := constructAuthCredentials(config.Auth.AuthKey, config.Auth.SignKey)
	if err != nil {
		return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to construct auth credentials").Err()
	}
	creds, err := getCredentials(config.TLS)
	if err != nil {
		return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to get credentials").Err()
	}
	var dialOpts []grpc.DialOption
	if config.TLS.Enabled {
		log.Debug().Bool("TLS", true).Msg("TLS for agent enabled, creating secure connection...")
		dialOpts = []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(authCreds),
		}
	} else {
		authCreds.Insecure = true
		dialOpts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithPerRPCCredentials(authCreds),
		}
	}

	conn, err := grpc.NewClient(config.Endpoint, dialOpts...)
	if err != nil {
		return nil, translateRPCErr(err)
	}

	c := protobuf.NewAgentClient(conn)

	client := &agentClient{
		AgentClient: c,
		connection:  conn,
	}

	return client, nil
}

func (c *agentClient) Close() error {
	if err := c.connection.Close(); err != nil {
		return appError.ErrGRPC.WithError(err).WithMessage("Failed to close agent client").Err()
	}

	return nil
}
