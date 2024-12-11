package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/cybericebox/agent/internal/config"
	"github.com/cybericebox/agent/pkg/appError"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"os"
)

type (
	Agent struct {
		auth    Authenticator
		config  *config.GRPCConfig
		useCase IUseCase
		protobuf.UnsafeAgentServer
	}

	IUseCase interface {
		IChallengeUseCase
		ILabUseCase
		IMonitoringUseCase
	}

	Dependencies struct {
		Config  *config.GRPCConfig
		UseCase IUseCase
	}
)

func getCredentials(conf *config.TLSConfig) (credentials.TransportCredentials, error) {
	log.Debug().Msg("Preparing credentials for RPC")

	certificate, err := tls.LoadX509KeyPair(conf.CertFile, conf.CertKey)
	if err != nil {
		return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to load certificates").Err()
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(conf.CAFile)
	if err != nil {
		return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to read CA file").Err()
	}
	// CA file for let's encrypt is located under domain conf as `chain.pem`
	// pass chain.pem location
	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, appError.ErrGRPC.WithMessage("Failed to append client certificates").Err()
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	return creds, nil
}

// SecureConn enables communication over secure channel
func secureConn(conf *config.TLSConfig) ([]grpc.ServerOption, error) {
	if conf.Enabled {
		log.Debug().Msgf("Conf cert-file: %s, cert-key: %s ca: %s", conf.CertFile, conf.CertKey, conf.CAFile)
		creds, err := getCredentials(conf)
		if err != nil {
			return []grpc.ServerOption{}, appError.ErrGRPC.WithError(err).WithMessage("Failed to get credentials").Err()
		}

		log.Debug().Msg("Server is running in secure mode!")
		return []grpc.ServerOption{grpc.Creds(creds)}, nil
	}
	return []grpc.ServerOption{}, nil
}

func New(deps Dependencies) (*grpc.Server, error) {

	gRPCServer := &Agent{
		auth:    NewAuthenticator(deps.Config.Auth.SignKey, deps.Config.Auth.AuthKey),
		config:  deps.Config,
		useCase: deps.UseCase,
	}
	opts, err := secureConn(&deps.Config.TLS)
	if err != nil {
		return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to get secure connection").Err()
	}

	gRPCEndpoint := gRPCServer.addAuth(opts...)

	reflection.Register(gRPCEndpoint)
	protobuf.RegisterAgentServer(gRPCEndpoint, gRPCServer)

	return gRPCEndpoint, nil
}

// AddAuth adds authentication to gRPC server
func (a *Agent) addAuth(opts ...grpc.ServerOption) *grpc.Server {
	streamInterceptor := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := a.auth.AuthenticateContext(stream.Context()); err != nil {
			return appError.ErrGRPC.WithError(err).WithMessage("Failed to authenticate context").Err()
		}
		return handler(srv, stream)
	}

	unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := a.auth.AuthenticateContext(ctx); err != nil {
			return nil, appError.ErrGRPC.WithError(err).WithMessage("Failed to authenticate context").Err()
		}
		return handler(ctx, req)
	}

	opts = append([]grpc.ServerOption{
		grpc.StreamInterceptor(streamInterceptor),
		grpc.UnaryInterceptor(unaryInterceptor),
	}, opts...)
	return grpc.NewServer(opts...)

}
