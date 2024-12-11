package grpc

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/rs/zerolog/log"
)

type IChallengeUseCase interface {
	AddLabsChallenges(ctx context.Context, labIDs []string, configs []model.ChallengeConfig, flagsEnvVars map[string]map[string]map[string]model.EnvConfig) error
	StartLabsChallenges(ctx context.Context, labIDs, challengeIDs []string) error
	StopLabsChallenges(ctx context.Context, labIDs, challengeIDs []string) error
	ResetLabsChallenges(ctx context.Context, labIDs, challengeIDs []string) error
	DeleteLabsChallenges(ctx context.Context, labIDs, challengeIDs []string) error
}

func (a *Agent) AddLabsChallenges(ctx context.Context, request *protobuf.AddLabsChallengesRequest) (*protobuf.EmptyResponse, error) {
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
				ID:    inst.GetID(),
				Image: inst.GetImage(),
				Resources: model.ResourcesConfig{
					Requests: model.ResourceConfig{
						Memory: inst.GetResources().GetMemory(),
						CPU:    inst.GetResources().GetCPU(),
					},
					Limit: model.ResourceConfig{
						Memory: inst.GetResources().GetMemory(),
						CPU:    inst.GetResources().GetCPU(),
					},
				},
				Envs:    envs,
				Records: records,
			})
		}

		challengesConfigs = append(challengesConfigs, model.ChallengeConfig{ID: chConfig.GetID(), Instances: instances})
	}
	// map[labID]map[challengeID]map[instanceID]model.EnvConfig
	flagEnvVariables := make(map[string]map[string]map[string]model.EnvConfig)

	for _, flagEnv := range request.GetFlagEnvVariables() {
		if _, ok := flagEnvVariables[flagEnv.GetLabID()]; !ok {
			flagEnvVariables[flagEnv.GetLabID()] = make(map[string]map[string]model.EnvConfig)
		}
		if _, ok := flagEnvVariables[flagEnv.GetLabID()][flagEnv.GetChallengeID()]; !ok {
			flagEnvVariables[flagEnv.GetLabID()][flagEnv.GetChallengeID()] = make(map[string]model.EnvConfig)
		}
		flagEnvVariables[flagEnv.GetLabID()][flagEnv.GetChallengeID()][flagEnv.GetInstanceID()] = model.EnvConfig{
			Name:  flagEnv.GetVariable(),
			Value: flagEnv.GetFlag(),
		}
	}

	if err := a.useCase.AddLabsChallenges(ctx, request.GetLabIDs(), challengesConfigs, flagEnvVariables); err != nil {
		log.Error().Err(err).Msg("Failed to add lab challenges")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) DeleteLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.DeleteLabsChallenges(ctx, request.GetLabIDs(), request.GetChallengeIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to delete lab challenges")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StartLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.StartLabsChallenges(ctx, request.GetLabIDs(), request.GetChallengeIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to start lab challenges")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StopLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.StopLabsChallenges(ctx, request.GetLabIDs(), request.GetChallengeIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to stop lab challenges")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) ResetLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	if err := a.useCase.ResetLabsChallenges(ctx, request.GetLabIDs(), request.GetChallengeIDs()); err != nil {
		log.Error().Err(err).Msg("Failed to reset lab challenges")
		return nil, err
	}

	return &protobuf.EmptyResponse{}, nil
}
