package grpc

import (
	"context"
	"errors"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/rs/zerolog/log"
	"io"
)

type (
	IMonitoringUseCase interface {
		GetLabsStatus(ctx context.Context) ([]*model.LabStatus, error)
	}
)

func (a *Agent) Ping(_ context.Context, _ *protobuf.EmptyRequest) (*protobuf.EmptyResponse, error) {
	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) Monitoring(stream protobuf.Agent_MonitoringServer) error {
	log.Debug().Msg("Client connected to monitoring")
	for {
		select {
		case <-stream.Context().Done():
			log.Debug().Msg("Client disconnected from monitoring")
			return nil
		default:
			break
		}
		_, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug().Msg("Client disconnected from monitoring")
				return nil
			}
			log.Error().Err(err).Msg("Failed to receive monitoring request")
			continue
		}

		labs, err := a.useCase.GetLabsStatus(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Failed to get labs")
			continue
		}

		convLabs := make([]*protobuf.LabStatus, 0, len(labs))
		for _, lab := range labs {

			instance := make([]*protobuf.InstanceStatus, 0, len(lab.Instances))
			for _, inst := range lab.Instances {
				instance = append(instance, &protobuf.InstanceStatus{
					ID:     inst.ID.String(),
					Status: int32(inst.Status),
					Reason: inst.Reason,
					Resources: &protobuf.Resources{
						Memory: inst.Resources.Memory,
						CPU:    inst.Resources.CPU,
					},
				})
			}

			convLabs = append(convLabs, &protobuf.LabStatus{
				ID: lab.ID.String(),
				DNS: &protobuf.DNSStatus{
					Status: int32(lab.DNS.Status),
					Reason: lab.DNS.Reason,
					Resources: &protobuf.Resources{
						Memory: lab.DNS.Resources.Memory,
						CPU:    lab.DNS.Resources.CPU,
					},
				},
				Instances: instance,
			})
		}

		if err = stream.Send(&protobuf.MonitoringResponse{
			Labs: convLabs,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to send monitoring response")
		}
	}
}
