package grpc

import (
	"context"
	"github.com/cybericebox/agent/pkg/controller/grpc/protobuf"
)

func (a *Agent) Ping(_ context.Context, _ *protobuf.EmptyRequest) (*protobuf.EmptyResponse, error) {
	return &protobuf.EmptyResponse{}, nil
}
