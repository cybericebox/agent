package grpc

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/rs/zerolog/log"
)

type (
	ILabUseCase interface {
		CreateLabs(ctx context.Context, subnetMask uint32, count int) ([]*model.Lab, error)
		GetLabs(ctx context.Context, labIDs []string) ([]*model.Lab, error)
		DeleteLabs(ctx context.Context, labIDs []string) error
		StartLabs(ctx context.Context, labIDs []string) error
		StopLabs(ctx context.Context, labIDs []string) error
	}
)

func (a *Agent) GetLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.GetLabsResponse, error) {
	labs, err := a.useCase.GetLabs(ctx, request.GetIDs())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get labs")
		return nil, err
	}

	convLabs := make([]*protobuf.Lab, 0, len(labs))
	for _, lab := range labs {
		convLabs = append(convLabs, &protobuf.Lab{
			ID:   lab.ID.String(),
			CIDR: lab.CIDR.String(),
		})
	}

	return &protobuf.GetLabsResponse{
		Labs: convLabs,
	}, nil
}

func (a *Agent) CreateLabs(ctx context.Context, request *protobuf.CreateLabsRequest) (*protobuf.CreateLabsResponse, error) {
	labs, err := a.useCase.CreateLabs(ctx, request.GetCIDRMask(), int(request.GetCount()))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create labs")
		return nil, err
	}

	convLabs := make([]*protobuf.Lab, 0, len(labs))
	for _, lab := range labs {
		convLabs = append(convLabs, &protobuf.Lab{
			ID:   lab.ID.String(),
			CIDR: lab.CIDR.String(),
		})
	}

	return &protobuf.CreateLabsResponse{
		Labs: convLabs,
	}, nil
}

func (a *Agent) StartLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.StartLabs(ctx, request.GetIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to start labs")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StopLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.StopLabs(ctx, request.GetIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to stop labs")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) DeleteLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.DeleteLabs(ctx, request.GetIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to delete labs")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}
