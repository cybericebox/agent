package grpc

import (
	"context"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/rs/zerolog/log"
)

type IChallengeService interface {
	StartChallenge(ctx context.Context, labID, challengeID string) (errs error)
	StopChallenge(ctx context.Context, labID, challengeID string) (errs error)
	ResetChallenge(ctx context.Context, labID, challengeID string) (errs error)
}

func (a *Agent) StartChallenge(ctx context.Context, request *protobuf.ChallengeRequest) (*protobuf.EmptyResponse, error) {
	if err := a.service.StartChallenge(ctx, request.GetLabID(), request.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to start challenge")
		return &protobuf.EmptyResponse{}, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StopChallenge(ctx context.Context, request *protobuf.ChallengeRequest) (*protobuf.EmptyResponse, error) {
	if err := a.service.StopChallenge(ctx, request.GetLabID(), request.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to stop challenge")
		return &protobuf.EmptyResponse{}, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) ResetChallenge(ctx context.Context, request *protobuf.ChallengeRequest) (*protobuf.EmptyResponse, error) {
	if err := a.service.ResetChallenge(ctx, request.GetLabID(), request.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to reset challenge")
		return &protobuf.EmptyResponse{}, err
	}

	return &protobuf.EmptyResponse{}, nil
}
