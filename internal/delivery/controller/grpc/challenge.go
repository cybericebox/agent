package grpc

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

type IChallengeService interface {
	StartLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
	StopLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
	ResetLabChallenges(ctx context.Context, labID string, challengeIDs []string) error

	AddLabChallenges(ctx context.Context, labID string, configs []model.ChallengeConfig) error
	DeleteLabChallenges(ctx context.Context, labID string, challengeIDs []string) error
}

func (a *Agent) AddLabsChallenges(ctx context.Context, request *protobuf.AddLabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

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

	for _, labID := range request.GetLabIDs() {
		labChallengesConfigs := make([]model.ChallengeConfig, 0, len(challengesConfigs))

		for _, chConfig := range challengesConfigs {
			instances := make([]model.InstanceConfig, 0, len(chConfig.Instances))

			for _, inst := range chConfig.Instances {
				if flagEnv, ok := flagEnvVariables[labID][chConfig.ID][inst.ID]; ok {
					inst.Envs = append(inst.Envs, flagEnv)
				}

				instances = append(instances, inst)
			}

			labChallengesConfigs = append(labChallengesConfigs, model.ChallengeConfig{ID: chConfig.ID, Instances: instances})
		}

		if err := a.service.AddLabChallenges(ctx, labID, labChallengesConfigs); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to add lab challenges")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) DeleteLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {
		if err := a.service.DeleteLabChallenges(ctx, labID, request.GetChallengeIDs()); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to delete lab challenges")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StartLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {
		if err := a.service.StartLabChallenges(ctx, labID, request.GetChallengeIDs()); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to start lab challenges")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StopLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {
		if err := a.service.StopLabChallenges(ctx, labID, request.GetChallengeIDs()); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to stop lab challenges")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) ResetLabsChallenges(ctx context.Context, request *protobuf.LabsChallengesRequest) (*protobuf.EmptyResponse, error) {
	var errs error

	for _, labID := range request.GetLabIDs() {
		if err := a.service.ResetLabChallenges(ctx, labID, request.GetChallengeIDs()); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to reset lab challenges")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}
