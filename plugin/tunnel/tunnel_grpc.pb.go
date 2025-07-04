// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: tunnel.proto

package tunnel

import (
	context "context"
	empty "google.golang.org/protobuf/types/known/emptypb"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TunnelService_Authenticate_FullMethodName      = "/TunnelService/Authenticate"
	TunnelService_ClientEventNotify_FullMethodName = "/TunnelService/ClientEventNotify"
	TunnelService_GetClientStatus_FullMethodName   = "/TunnelService/GetClientStatus"
	TunnelService_AddPortMapping_FullMethodName    = "/TunnelService/AddPortMapping"
	TunnelService_TunnelData_FullMethodName        = "/TunnelService/TunnelData"
)

// TunnelServiceClient is the client API for TunnelService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// 服务端定义
type TunnelServiceClient interface {
	// 客户端认证
	Authenticate(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	// 客户端连接事件通知
	ClientEventNotify(ctx context.Context, in *ClientEvent, opts ...grpc.CallOption) (*empty.Empty, error)
	// 客户端状态查询
	GetClientStatus(ctx context.Context, in *ClientStatusQuery, opts ...grpc.CallOption) (*ClientStatusResponse, error)
	// 端口映射
	AddPortMapping(ctx context.Context, in *PortMappingRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// 透传数据（双向流式）
	TunnelData(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[TunnelDataRequest, TunnelDataResponse], error)
}

type tunnelServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTunnelServiceClient(cc grpc.ClientConnInterface) TunnelServiceClient {
	return &tunnelServiceClient{cc}
}

func (c *tunnelServiceClient) Authenticate(ctx context.Context, in *AuthRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, TunnelService_Authenticate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelServiceClient) ClientEventNotify(ctx context.Context, in *ClientEvent, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, TunnelService_ClientEventNotify_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelServiceClient) GetClientStatus(ctx context.Context, in *ClientStatusQuery, opts ...grpc.CallOption) (*ClientStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ClientStatusResponse)
	err := c.cc.Invoke(ctx, TunnelService_GetClientStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelServiceClient) AddPortMapping(ctx context.Context, in *PortMappingRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, TunnelService_AddPortMapping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelServiceClient) TunnelData(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[TunnelDataRequest, TunnelDataResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &TunnelService_ServiceDesc.Streams[0], TunnelService_TunnelData_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[TunnelDataRequest, TunnelDataResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TunnelService_TunnelDataClient = grpc.BidiStreamingClient[TunnelDataRequest, TunnelDataResponse]

// TunnelServiceServer is the server API for TunnelService service.
// All implementations must embed UnimplementedTunnelServiceServer
// for forward compatibility.
//
// 服务端定义
type TunnelServiceServer interface {
	// 客户端认证
	Authenticate(context.Context, *AuthRequest) (*AuthResponse, error)
	// 客户端连接事件通知
	ClientEventNotify(context.Context, *ClientEvent) (*empty.Empty, error)
	// 客户端状态查询
	GetClientStatus(context.Context, *ClientStatusQuery) (*ClientStatusResponse, error)
	// 端口映射
	AddPortMapping(context.Context, *PortMappingRequest) (*empty.Empty, error)
	// 透传数据（双向流式）
	TunnelData(grpc.BidiStreamingServer[TunnelDataRequest, TunnelDataResponse]) error
	mustEmbedUnimplementedTunnelServiceServer()
}

// UnimplementedTunnelServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTunnelServiceServer struct{}

func (UnimplementedTunnelServiceServer) Authenticate(context.Context, *AuthRequest) (*AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}
func (UnimplementedTunnelServiceServer) ClientEventNotify(context.Context, *ClientEvent) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientEventNotify not implemented")
}
func (UnimplementedTunnelServiceServer) GetClientStatus(context.Context, *ClientStatusQuery) (*ClientStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClientStatus not implemented")
}
func (UnimplementedTunnelServiceServer) AddPortMapping(context.Context, *PortMappingRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPortMapping not implemented")
}
func (UnimplementedTunnelServiceServer) TunnelData(grpc.BidiStreamingServer[TunnelDataRequest, TunnelDataResponse]) error {
	return status.Errorf(codes.Unimplemented, "method TunnelData not implemented")
}
func (UnimplementedTunnelServiceServer) mustEmbedUnimplementedTunnelServiceServer() {}
func (UnimplementedTunnelServiceServer) testEmbeddedByValue()                       {}

// UnsafeTunnelServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TunnelServiceServer will
// result in compilation errors.
type UnsafeTunnelServiceServer interface {
	mustEmbedUnimplementedTunnelServiceServer()
}

func RegisterTunnelServiceServer(s grpc.ServiceRegistrar, srv TunnelServiceServer) {
	// If the following call pancis, it indicates UnimplementedTunnelServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TunnelService_ServiceDesc, srv)
}

func _TunnelService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TunnelService_Authenticate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServiceServer).Authenticate(ctx, req.(*AuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TunnelService_ClientEventNotify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientEvent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServiceServer).ClientEventNotify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TunnelService_ClientEventNotify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServiceServer).ClientEventNotify(ctx, req.(*ClientEvent))
	}
	return interceptor(ctx, in, info, handler)
}

func _TunnelService_GetClientStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientStatusQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServiceServer).GetClientStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TunnelService_GetClientStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServiceServer).GetClientStatus(ctx, req.(*ClientStatusQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _TunnelService_AddPortMapping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PortMappingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServiceServer).AddPortMapping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TunnelService_AddPortMapping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServiceServer).AddPortMapping(ctx, req.(*PortMappingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TunnelService_TunnelData_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TunnelServiceServer).TunnelData(&grpc.GenericServerStream[TunnelDataRequest, TunnelDataResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type TunnelService_TunnelDataServer = grpc.BidiStreamingServer[TunnelDataRequest, TunnelDataResponse]

// TunnelService_ServiceDesc is the grpc.ServiceDesc for TunnelService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TunnelService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "TunnelService",
	HandlerType: (*TunnelServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _TunnelService_Authenticate_Handler,
		},
		{
			MethodName: "ClientEventNotify",
			Handler:    _TunnelService_ClientEventNotify_Handler,
		},
		{
			MethodName: "GetClientStatus",
			Handler:    _TunnelService_GetClientStatus_Handler,
		},
		{
			MethodName: "AddPortMapping",
			Handler:    _TunnelService_AddPortMapping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TunnelData",
			Handler:       _TunnelService_TunnelData_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "tunnel.proto",
}
