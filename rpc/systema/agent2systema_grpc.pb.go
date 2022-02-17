// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package systema

import (
	context "context"
	manager "github.com/TinderBackend/telepresence/rpc/v2/manager"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SystemAAgentClient is the client API for SystemAAgent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SystemAAgentClient interface {
	// ReviewIntercept gives SystemA an opportunity to review an
	// intercept before the Ambassador Telepresence agent activates it.
	//
	// There is no "remove" call; SystemA should call
	// telepresence.manager.Manager/WatchIntercepts to be informed of
	// such things.
	ReviewIntercept(ctx context.Context, in *manager.InterceptInfo, opts ...grpc.CallOption) (*ReviewInterceptResponse, error)
}

type systemAAgentClient struct {
	cc grpc.ClientConnInterface
}

func NewSystemAAgentClient(cc grpc.ClientConnInterface) SystemAAgentClient {
	return &systemAAgentClient{cc}
}

func (c *systemAAgentClient) ReviewIntercept(ctx context.Context, in *manager.InterceptInfo, opts ...grpc.CallOption) (*ReviewInterceptResponse, error) {
	out := new(ReviewInterceptResponse)
	err := c.cc.Invoke(ctx, "/telepresence.systema.SystemAAgent/ReviewIntercept", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SystemAAgentServer is the server API for SystemAAgent service.
// All implementations must embed UnimplementedSystemAAgentServer
// for forward compatibility
type SystemAAgentServer interface {
	// ReviewIntercept gives SystemA an opportunity to review an
	// intercept before the Ambassador Telepresence agent activates it.
	//
	// There is no "remove" call; SystemA should call
	// telepresence.manager.Manager/WatchIntercepts to be informed of
	// such things.
	ReviewIntercept(context.Context, *manager.InterceptInfo) (*ReviewInterceptResponse, error)
	mustEmbedUnimplementedSystemAAgentServer()
}

// UnimplementedSystemAAgentServer must be embedded to have forward compatible implementations.
type UnimplementedSystemAAgentServer struct {
}

func (UnimplementedSystemAAgentServer) ReviewIntercept(context.Context, *manager.InterceptInfo) (*ReviewInterceptResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReviewIntercept not implemented")
}
func (UnimplementedSystemAAgentServer) mustEmbedUnimplementedSystemAAgentServer() {}

// UnsafeSystemAAgentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SystemAAgentServer will
// result in compilation errors.
type UnsafeSystemAAgentServer interface {
	mustEmbedUnimplementedSystemAAgentServer()
}

func RegisterSystemAAgentServer(s grpc.ServiceRegistrar, srv SystemAAgentServer) {
	s.RegisterService(&SystemAAgent_ServiceDesc, srv)
}

func _SystemAAgent_ReviewIntercept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(manager.InterceptInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemAAgentServer).ReviewIntercept(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/telepresence.systema.SystemAAgent/ReviewIntercept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemAAgentServer).ReviewIntercept(ctx, req.(*manager.InterceptInfo))
	}
	return interceptor(ctx, in, info, handler)
}

// SystemAAgent_ServiceDesc is the grpc.ServiceDesc for SystemAAgent service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SystemAAgent_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "telepresence.systema.SystemAAgent",
	HandlerType: (*SystemAAgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReviewIntercept",
			Handler:    _SystemAAgent_ReviewIntercept_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/systema/agent2systema.proto",
}
