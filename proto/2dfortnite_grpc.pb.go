// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: proto/2dfortnite.proto

package fortnite

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FortniteServiceClient is the client API for FortniteService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FortniteServiceClient interface {
	// transmit the player's info and receive other player info
	RegisterPlayer(ctx context.Context, in *RegisterPlayerRequest, opts ...grpc.CallOption) (*RegisterPlayerResponse, error)
	WorldState(ctx context.Context, in *PlayerId, opts ...grpc.CallOption) (FortniteService_WorldStateClient, error)
	PlayerStream(ctx context.Context, opts ...grpc.CallOption) (FortniteService_PlayerStreamClient, error)
	ProjectileInfo(ctx context.Context, in *PlayerId, opts ...grpc.CallOption) (FortniteService_ProjectileInfoClient, error)
	DoAction(ctx context.Context, in *DoActionRequest, opts ...grpc.CallOption) (*DoActionResponse, error)
}

type fortniteServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFortniteServiceClient(cc grpc.ClientConnInterface) FortniteServiceClient {
	return &fortniteServiceClient{cc}
}

func (c *fortniteServiceClient) RegisterPlayer(ctx context.Context, in *RegisterPlayerRequest, opts ...grpc.CallOption) (*RegisterPlayerResponse, error) {
	out := new(RegisterPlayerResponse)
	err := c.cc.Invoke(ctx, "/fortniteservice.FortniteService/RegisterPlayer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fortniteServiceClient) WorldState(ctx context.Context, in *PlayerId, opts ...grpc.CallOption) (FortniteService_WorldStateClient, error) {
	stream, err := c.cc.NewStream(ctx, &FortniteService_ServiceDesc.Streams[0], "/fortniteservice.FortniteService/WorldState", opts...)
	if err != nil {
		return nil, err
	}
	x := &fortniteServiceWorldStateClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FortniteService_WorldStateClient interface {
	Recv() (*WorldStateResponse, error)
	grpc.ClientStream
}

type fortniteServiceWorldStateClient struct {
	grpc.ClientStream
}

func (x *fortniteServiceWorldStateClient) Recv() (*WorldStateResponse, error) {
	m := new(WorldStateResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fortniteServiceClient) PlayerStream(ctx context.Context, opts ...grpc.CallOption) (FortniteService_PlayerStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &FortniteService_ServiceDesc.Streams[1], "/fortniteservice.FortniteService/PlayerStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &fortniteServicePlayerStreamClient{stream}
	return x, nil
}

type FortniteService_PlayerStreamClient interface {
	Send(*Player) error
	Recv() (*Player, error)
	grpc.ClientStream
}

type fortniteServicePlayerStreamClient struct {
	grpc.ClientStream
}

func (x *fortniteServicePlayerStreamClient) Send(m *Player) error {
	return x.ClientStream.SendMsg(m)
}

func (x *fortniteServicePlayerStreamClient) Recv() (*Player, error) {
	m := new(Player)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fortniteServiceClient) ProjectileInfo(ctx context.Context, in *PlayerId, opts ...grpc.CallOption) (FortniteService_ProjectileInfoClient, error) {
	stream, err := c.cc.NewStream(ctx, &FortniteService_ServiceDesc.Streams[2], "/fortniteservice.FortniteService/ProjectileInfo", opts...)
	if err != nil {
		return nil, err
	}
	x := &fortniteServiceProjectileInfoClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FortniteService_ProjectileInfoClient interface {
	Recv() (*Projectile, error)
	grpc.ClientStream
}

type fortniteServiceProjectileInfoClient struct {
	grpc.ClientStream
}

func (x *fortniteServiceProjectileInfoClient) Recv() (*Projectile, error) {
	m := new(Projectile)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fortniteServiceClient) DoAction(ctx context.Context, in *DoActionRequest, opts ...grpc.CallOption) (*DoActionResponse, error) {
	out := new(DoActionResponse)
	err := c.cc.Invoke(ctx, "/fortniteservice.FortniteService/DoAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FortniteServiceServer is the server API for FortniteService service.
// All implementations must embed UnimplementedFortniteServiceServer
// for forward compatibility
type FortniteServiceServer interface {
	// transmit the player's info and receive other player info
	RegisterPlayer(context.Context, *RegisterPlayerRequest) (*RegisterPlayerResponse, error)
	WorldState(*PlayerId, FortniteService_WorldStateServer) error
	PlayerStream(FortniteService_PlayerStreamServer) error
	ProjectileInfo(*PlayerId, FortniteService_ProjectileInfoServer) error
	DoAction(context.Context, *DoActionRequest) (*DoActionResponse, error)
	mustEmbedUnimplementedFortniteServiceServer()
}

// UnimplementedFortniteServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFortniteServiceServer struct {
}

func (UnimplementedFortniteServiceServer) RegisterPlayer(context.Context, *RegisterPlayerRequest) (*RegisterPlayerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterPlayer not implemented")
}
func (UnimplementedFortniteServiceServer) WorldState(*PlayerId, FortniteService_WorldStateServer) error {
	return status.Errorf(codes.Unimplemented, "method WorldState not implemented")
}
func (UnimplementedFortniteServiceServer) PlayerStream(FortniteService_PlayerStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method PlayerStream not implemented")
}
func (UnimplementedFortniteServiceServer) ProjectileInfo(*PlayerId, FortniteService_ProjectileInfoServer) error {
	return status.Errorf(codes.Unimplemented, "method ProjectileInfo not implemented")
}
func (UnimplementedFortniteServiceServer) DoAction(context.Context, *DoActionRequest) (*DoActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DoAction not implemented")
}
func (UnimplementedFortniteServiceServer) mustEmbedUnimplementedFortniteServiceServer() {}

// UnsafeFortniteServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FortniteServiceServer will
// result in compilation errors.
type UnsafeFortniteServiceServer interface {
	mustEmbedUnimplementedFortniteServiceServer()
}

func RegisterFortniteServiceServer(s grpc.ServiceRegistrar, srv FortniteServiceServer) {
	s.RegisterService(&FortniteService_ServiceDesc, srv)
}

func _FortniteService_RegisterPlayer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterPlayerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FortniteServiceServer).RegisterPlayer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fortniteservice.FortniteService/RegisterPlayer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FortniteServiceServer).RegisterPlayer(ctx, req.(*RegisterPlayerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FortniteService_WorldState_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PlayerId)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FortniteServiceServer).WorldState(m, &fortniteServiceWorldStateServer{stream})
}

type FortniteService_WorldStateServer interface {
	Send(*WorldStateResponse) error
	grpc.ServerStream
}

type fortniteServiceWorldStateServer struct {
	grpc.ServerStream
}

func (x *fortniteServiceWorldStateServer) Send(m *WorldStateResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _FortniteService_PlayerStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FortniteServiceServer).PlayerStream(&fortniteServicePlayerStreamServer{stream})
}

type FortniteService_PlayerStreamServer interface {
	Send(*Player) error
	Recv() (*Player, error)
	grpc.ServerStream
}

type fortniteServicePlayerStreamServer struct {
	grpc.ServerStream
}

func (x *fortniteServicePlayerStreamServer) Send(m *Player) error {
	return x.ServerStream.SendMsg(m)
}

func (x *fortniteServicePlayerStreamServer) Recv() (*Player, error) {
	m := new(Player)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _FortniteService_ProjectileInfo_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PlayerId)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FortniteServiceServer).ProjectileInfo(m, &fortniteServiceProjectileInfoServer{stream})
}

type FortniteService_ProjectileInfoServer interface {
	Send(*Projectile) error
	grpc.ServerStream
}

type fortniteServiceProjectileInfoServer struct {
	grpc.ServerStream
}

func (x *fortniteServiceProjectileInfoServer) Send(m *Projectile) error {
	return x.ServerStream.SendMsg(m)
}

func _FortniteService_DoAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DoActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FortniteServiceServer).DoAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fortniteservice.FortniteService/DoAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FortniteServiceServer).DoAction(ctx, req.(*DoActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FortniteService_ServiceDesc is the grpc.ServiceDesc for FortniteService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FortniteService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fortniteservice.FortniteService",
	HandlerType: (*FortniteServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterPlayer",
			Handler:    _FortniteService_RegisterPlayer_Handler,
		},
		{
			MethodName: "DoAction",
			Handler:    _FortniteService_DoAction_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WorldState",
			Handler:       _FortniteService_WorldState_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PlayerStream",
			Handler:       _FortniteService_PlayerStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ProjectileInfo",
			Handler:       _FortniteService_ProjectileInfo_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/2dfortnite.proto",
}
