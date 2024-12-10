package grpc

import (
	"context"
	"github.com/cybericebox/agent/internal/model"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
	"github.com/cybericebox/lib/pkg/worker"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
)

type (
	ILabService interface {
		CreateLab(ctx context.Context, subnetMask uint32) (string, error)
		GetLab(ctx context.Context, labID string) (*model.Lab, error)
		DeleteLab(ctx context.Context, labID string) error
		StartLab(ctx context.Context, labID string) error
		StopLab(ctx context.Context, labID string) error
	}
)

func (a *Agent) GetLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.GetLabsResponse, error) {
	var errs error

	labs := make([]*protobuf.Lab, 0, len(request.GetIDs()))

	wg := new(sync.WaitGroup)

	for _, id := range request.GetIDs() {
		wg.Add(1)
		a.worker.AddTask(worker.NewTask().
			WithKey(id, "get_lab").
			WithDo(func() error {
				lab, err := a.service.GetLab(ctx, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				labs = append(labs, &protobuf.Lab{
					ID:   lab.ID.String(),
					CIDR: lab.CIDRManager.GetCIDR(),
				})
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to get labs")
		return nil, errs
	}

	return &protobuf.GetLabsResponse{
		Labs: labs,
	}, nil
}

func (a *Agent) CreateLabs(ctx context.Context, request *protobuf.CreateLabsRequest) (*protobuf.CreateLabsResponse, error) {
	labIDs := make([]string, 0, request.GetCount())
	var errs error
	wg := new(sync.WaitGroup)

	for i := uint32(0); i < request.GetCount(); i++ {
		wg.Add(1)
		a.worker.AddTask(worker.NewTask().
			WithKey(strconv.Itoa(int(i)), "create_lab").
			WithDo(func() error {
				labID, err := a.service.CreateLab(ctx, request.GetCIDRMask())
				if err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				labIDs = append(labIDs, labID)
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to create labs")
		return nil, errs
	}

	return &protobuf.CreateLabsResponse{
		IDs: labIDs,
	}, nil
}

func (a *Agent) DeleteLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	var errs error
	wg := new(sync.WaitGroup)

	for _, id := range request.GetIDs() {
		wg.Add(1)
		a.worker.AddTask(worker.NewTask().
			WithKey(id, "delete_lab").
			WithDo(func() error {
				if err := a.service.DeleteLab(ctx, id); err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).
			WithOnDone(func(_, _ error) {
				wg.Done()
			}).
			Create())
	}

	wg.Wait()

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to delete labs")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StartLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	var errs error
	wg := new(sync.WaitGroup)

	for _, id := range request.GetIDs() {
		wg.Add(1)
		a.worker.AddTask(worker.NewTask().
			WithKey(id, "start_lab").
			WithDo(func() error {
				if err := a.service.StartLab(ctx, id); err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).
			WithOnDone(func(_, _ error) {
				wg.Done()
			}).Create())
	}

	wg.Wait()

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to start labs")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}

func (a *Agent) StopLabs(ctx context.Context, request *protobuf.LabsRequest) (*protobuf.EmptyResponse, error) {
	var errs error
	wg := new(sync.WaitGroup)

	for _, id := range request.GetIDs() {
		wg.Add(1)
		a.worker.AddTask(worker.NewTask().
			WithKey(id, "stop_lab").
			WithDo(func() error {
				if err := a.service.StopLab(ctx, id); err != nil {
					errs = multierror.Append(errs, err)
					return err
				}
				return nil
			}).WithOnDone(func(_, _ error) {
			wg.Done()
		}).Create())
	}

	wg.Wait()

	if errs != nil {
		log.Error().Err(errs).Msg("Failed to stop labs")
		return nil, errs
	}

	return &protobuf.EmptyResponse{}, nil
}
