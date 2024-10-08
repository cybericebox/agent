// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v5.27.0
// source: agent.proto

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Agent_Ping_FullMethodName                 = "/agent.Agent/Ping"
	Agent_CreateLabs_FullMethodName           = "/agent.Agent/CreateLabs"
	Agent_DeleteLabs_FullMethodName           = "/agent.Agent/DeleteLabs"
	Agent_AddLabChallenges_FullMethodName     = "/agent.Agent/AddLabChallenges"
	Agent_DeleteLabsChallenges_FullMethodName = "/agent.Agent/DeleteLabsChallenges"
	Agent_GetLabs_FullMethodName              = "/agent.Agent/GetLabs"
	Agent_StartChallenge_FullMethodName       = "/agent.Agent/StartChallenge"
	Agent_StopChallenge_FullMethodName        = "/agent.Agent/StopChallenge"
	Agent_ResetChallenge_FullMethodName       = "/agent.Agent/ResetChallenge"
)

// AgentClient is the client API for Agent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AgentClient interface {
	// metrics
	Ping(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	// laboratory
	CreateLabs(ctx context.Context, in *CreateLabsRequest, opts ...grpc.CallOption) (*CreateLabsResponse, error)
	DeleteLabs(ctx context.Context, in *DeleteLabsRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	AddLabChallenges(ctx context.Context, in *AddLabChallengesRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	DeleteLabsChallenges(ctx context.Context, in *DeleteLabsChallengesRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	GetLabs(ctx context.Context, in *GetLabsRequest, opts ...grpc.CallOption) (*GetLabsResponse, error)
	// challenge
	StartChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	StopChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
	ResetChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type agentClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentClient(cc grpc.ClientConnInterface) AgentClient {
	return &agentClient{cc}
}

func (c *agentClient) Ping(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) CreateLabs(ctx context.Context, in *CreateLabsRequest, opts ...grpc.CallOption) (*CreateLabsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateLabsResponse)
	err := c.cc.Invoke(ctx, Agent_CreateLabs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) DeleteLabs(ctx context.Context, in *DeleteLabsRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_DeleteLabs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) AddLabChallenges(ctx context.Context, in *AddLabChallengesRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_AddLabChallenges_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) DeleteLabsChallenges(ctx context.Context, in *DeleteLabsChallengesRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_DeleteLabsChallenges_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) GetLabs(ctx context.Context, in *GetLabsRequest, opts ...grpc.CallOption) (*GetLabsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetLabsResponse)
	err := c.cc.Invoke(ctx, Agent_GetLabs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) StartChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_StartChallenge_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) StopChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_StopChallenge_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) ResetChallenge(ctx context.Context, in *ChallengeRequest, opts ...grpc.CallOption) (*EmptyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, Agent_ResetChallenge_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServer is the server API for Agent service.
// All implementations must embed UnimplementedAgentServer
// for forward compatibility
type AgentServer interface {
	// metrics
	Ping(context.Context, *EmptyRequest) (*EmptyResponse, error)
	// laboratory
	CreateLabs(context.Context, *CreateLabsRequest) (*CreateLabsResponse, error)
	DeleteLabs(context.Context, *DeleteLabsRequest) (*EmptyResponse, error)
	AddLabChallenges(context.Context, *AddLabChallengesRequest) (*EmptyResponse, error)
	DeleteLabsChallenges(context.Context, *DeleteLabsChallengesRequest) (*EmptyResponse, error)
	GetLabs(context.Context, *GetLabsRequest) (*GetLabsResponse, error)
	// challenge
	StartChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error)
	StopChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error)
	ResetChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error)
	mustEmbedUnimplementedAgentServer()
}

// UnimplementedAgentServer must be embedded to have forward compatible implementations.
type UnimplementedAgentServer struct {
}

func (UnimplementedAgentServer) Ping(context.Context, *EmptyRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedAgentServer) CreateLabs(context.Context, *CreateLabsRequest) (*CreateLabsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateLabs not implemented")
}
func (UnimplementedAgentServer) DeleteLabs(context.Context, *DeleteLabsRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLabs not implemented")
}
func (UnimplementedAgentServer) AddLabChallenges(context.Context, *AddLabChallengesRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddLabChallenges not implemented")
}
func (UnimplementedAgentServer) DeleteLabsChallenges(context.Context, *DeleteLabsChallengesRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLabsChallenges not implemented")
}
func (UnimplementedAgentServer) GetLabs(context.Context, *GetLabsRequest) (*GetLabsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLabs not implemented")
}
func (UnimplementedAgentServer) StartChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartChallenge not implemented")
}
func (UnimplementedAgentServer) StopChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopChallenge not implemented")
}
func (UnimplementedAgentServer) ResetChallenge(context.Context, *ChallengeRequest) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetChallenge not implemented")
}
func (UnimplementedAgentServer) mustEmbedUnimplementedAgentServer() {}

// UnsafeAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AgentServer will
// result in compilation errors.
type UnsafeAgentServer interface {
	mustEmbedUnimplementedAgentServer()
}

func RegisterAgentServer(s grpc.ServiceRegistrar, srv AgentServer) {
	s.RegisterService(&Agent_ServiceDesc, srv)
}

func _Agent_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).Ping(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_CreateLabs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateLabsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).CreateLabs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_CreateLabs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).CreateLabs(ctx, req.(*CreateLabsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_DeleteLabs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteLabsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).DeleteLabs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_DeleteLabs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).DeleteLabs(ctx, req.(*DeleteLabsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_AddLabChallenges_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddLabChallengesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).AddLabChallenges(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_AddLabChallenges_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).AddLabChallenges(ctx, req.(*AddLabChallengesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_DeleteLabsChallenges_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteLabsChallengesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).DeleteLabsChallenges(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_DeleteLabsChallenges_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).DeleteLabsChallenges(ctx, req.(*DeleteLabsChallengesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_GetLabs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLabsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).GetLabs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_GetLabs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).GetLabs(ctx, req.(*GetLabsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_StartChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChallengeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).StartChallenge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_StartChallenge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).StartChallenge(ctx, req.(*ChallengeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_StopChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChallengeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).StopChallenge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_StopChallenge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).StopChallenge(ctx, req.(*ChallengeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_ResetChallenge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChallengeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).ResetChallenge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Agent_ResetChallenge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).ResetChallenge(ctx, req.(*ChallengeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Agent_ServiceDesc is the grpc.ServiceDesc for Agent service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Agent_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agent.Agent",
	HandlerType: (*AgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Agent_Ping_Handler,
		},
		{
			MethodName: "CreateLabs",
			Handler:    _Agent_CreateLabs_Handler,
		},
		{
			MethodName: "DeleteLabs",
			Handler:    _Agent_DeleteLabs_Handler,
		},
		{
			MethodName: "AddLabChallenges",
			Handler:    _Agent_AddLabChallenges_Handler,
		},
		{
			MethodName: "DeleteLabsChallenges",
			Handler:    _Agent_DeleteLabsChallenges_Handler,
		},
		{
			MethodName: "GetLabs",
			Handler:    _Agent_GetLabs_Handler,
		},
		{
			MethodName: "StartChallenge",
			Handler:    _Agent_StartChallenge_Handler,
		},
		{
			MethodName: "StopChallenge",
			Handler:    _Agent_StopChallenge_Handler,
		},
		{
			MethodName: "ResetChallenge",
			Handler:    _Agent_ResetChallenge_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "agent.proto",
}
