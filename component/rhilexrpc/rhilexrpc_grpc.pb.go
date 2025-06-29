// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: rhilexrpc.proto

package rhilexrpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	RhilexRpc_Request_FullMethodName = "/rhilexrpc.RhilexRpc/Request"
)

// RhilexRpcClient is the client API for RhilexRpc service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RhilexRpcClient interface {
	Request(ctx context.Context, in *RpcRequest, opts ...grpc.CallOption) (*RpcResponse, error)
}

type rhilexRpcClient struct {
	cc grpc.ClientConnInterface
}

func NewRhilexRpcClient(cc grpc.ClientConnInterface) RhilexRpcClient {
	return &rhilexRpcClient{cc}
}

func (c *rhilexRpcClient) Request(ctx context.Context, in *RpcRequest, opts ...grpc.CallOption) (*RpcResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RpcResponse)
	err := c.cc.Invoke(ctx, RhilexRpc_Request_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RhilexRpcServer is the server API for RhilexRpc service.
// All implementations must embed UnimplementedRhilexRpcServer
// for forward compatibility.
type RhilexRpcServer interface {
	Request(context.Context, *RpcRequest) (*RpcResponse, error)
	mustEmbedUnimplementedRhilexRpcServer()
}

// UnimplementedRhilexRpcServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRhilexRpcServer struct{}

func (UnimplementedRhilexRpcServer) Request(context.Context, *RpcRequest) (*RpcResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Request not implemented")
}
func (UnimplementedRhilexRpcServer) mustEmbedUnimplementedRhilexRpcServer() {}
func (UnimplementedRhilexRpcServer) testEmbeddedByValue()                   {}

// UnsafeRhilexRpcServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RhilexRpcServer will
// result in compilation errors.
type UnsafeRhilexRpcServer interface {
	mustEmbedUnimplementedRhilexRpcServer()
}

func RegisterRhilexRpcServer(s grpc.ServiceRegistrar, srv RhilexRpcServer) {
	// If the following call pancis, it indicates UnimplementedRhilexRpcServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RhilexRpc_ServiceDesc, srv)
}

func _RhilexRpc_Request_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RpcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RhilexRpcServer).Request(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RhilexRpc_Request_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RhilexRpcServer).Request(ctx, req.(*RpcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RhilexRpc_ServiceDesc is the grpc.ServiceDesc for RhilexRpc service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RhilexRpc_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rhilexrpc.RhilexRpc",
	HandlerType: (*RhilexRpcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Request",
			Handler:    _RhilexRpc_Request_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rhilexrpc.proto",
}
