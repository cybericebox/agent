package grpc

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

type (
	ILabService interface {
		CreateLab(ctx context.Context, subnetMask uint32) (string, error)
		GetLab(ctx context.Context, id string) (*model.Lab, error)
		DeleteLab(ctx context.Context, id string) error

		AddLabChallenges(ctx context.Context, labID string, configs []model.ChallengeConfig) error
		DeleteLabChallenges(ctx context.Context, labID string, challengeIds []string) error
	}
)

func (a *Agent) GetLabs(ctx context.Context, request *protobuf.GetLabsRequest) (*protobuf.GetLabsResponse, error) {
	var errs error

	labs := make([]*protobuf.Lab, 0, len(request.GetLabIDs()))

	for _, id := range request.GetLabIDs() {
		lab, err := a.service.GetLab(ctx, id)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		labs = append(labs, &protobuf.Lab{
			Id:   lab.ID.String(),
			Cidr: lab.CIDRManager.GetCIDR(),
		})
	}

	if errs != nil {
		return &protobuf.GetLabsResponse{}, errs
	}

	return &protobuf.GetLabsResponse{
		Labs: labs,
	}, nil
}

func (a *Agent) CreateLab(ctx context.Context, request *protobuf.CreateLabRequest) (*protobuf.CreateLabResponse, error) {

	labID, err := a.service.CreateLab(ctx, request.GetCidrMask())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create lab")
		return &protobuf.CreateLabResponse{}, err
	}

	return &protobuf.CreateLabResponse{
		Id: labID,
	}, nil
}

func (a *Agent) DeleteLabs(ctx context.Context, request *protobuf.DeleteLabsRequest) (*protobuf.EmptyResponse, error) {
	var errs error
	for _, id := range request.GetLabIDs() {
		if err := a.service.DeleteLab(ctx, id); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs != nil {
		return &protobuf.EmptyResponse{}, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) AddLabsChallenges(ctx context.Context, request *protobuf.AddLabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {

		challengesConfigs := make([]model.ChallengeConfig, 0)

		for _, chConfig := range request.GetChallenges() {
			instances := make([]model.InstanceConfig, 0)

			for _, inst := range chConfig.GetInstances() {
				envs := make([]model.EnvConfig, 0)
				for _, env := range inst.GetEnvs() {
					envs = append(envs, model.EnvConfig{
						Name:  env.GetName(),
						Value: env.GetValue(),
					})
				}

				records := make([]model.DNSRecordConfig, 0)
				for _, record := range inst.GetRecords() {
					records = append(records, model.DNSRecordConfig{
						Type: record.GetType(),
						Name: record.GetName(),
						Data: record.GetData(),
					})
				}

				instances = append(instances, model.InstanceConfig{
					Id:    inst.GetId(),
					Image: inst.GetImage(),
					Resources: model.ResourcesConfig{
						Requests: model.ResourceConfig{
							Memory: inst.GetResources().GetMemory(),
							CPU:    inst.GetResources().GetCpu(),
						},
						Limit: model.ResourceConfig{
							Memory: inst.GetResources().GetMemory(),
							CPU:    inst.GetResources().GetCpu(),
						},
					},
					Envs:    envs,
					Records: records,
				})
			}

			challengesConfigs = append(challengesConfigs, model.ChallengeConfig{Id: chConfig.GetId(), Instances: instances})
		}

		if err := a.service.AddLabChallenges(ctx, labID, challengesConfigs); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		return &protobuf.EmptyResponse{}, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) DeleteLabsChallenges(ctx context.Context, request *protobuf.DeleteLabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {
		if err := a.service.DeleteLabChallenges(ctx, labID, request.GetChallengeIDs()); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		return &protobuf.EmptyResponse{}, errs
	}

	return &protobuf.EmptyResponse{}, nil
}
