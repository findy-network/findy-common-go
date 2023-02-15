// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: authn.proto

package v1

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

// AuthnServiceClient is the client API for AuthnService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthnServiceClient interface {
	// Enter enters authn command.
	Enter(ctx context.Context, in *Cmd, opts ...grpc.CallOption) (AuthnService_EnterClient, error)
	// EnterSecret enters needed secrets after specific CmdStatus has received.
	EnterSecret(ctx context.Context, in *SecretMsg, opts ...grpc.CallOption) (*SecretResult, error)
}

type authnServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthnServiceClient(cc grpc.ClientConnInterface) AuthnServiceClient {
	return &authnServiceClient{cc}
}

func (c *authnServiceClient) Enter(ctx context.Context, in *Cmd, opts ...grpc.CallOption) (AuthnService_EnterClient, error) {
	stream, err := c.cc.NewStream(ctx, &AuthnService_ServiceDesc.Streams[0], "/authn.v1.AuthnService/Enter", opts...)
	if err != nil {
		return nil, err
	}
	x := &authnServiceEnterClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AuthnService_EnterClient interface {
	Recv() (*CmdStatus, error)
	grpc.ClientStream
}

type authnServiceEnterClient struct {
	grpc.ClientStream
}

func (x *authnServiceEnterClient) Recv() (*CmdStatus, error) {
	m := new(CmdStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *authnServiceClient) EnterSecret(ctx context.Context, in *SecretMsg, opts ...grpc.CallOption) (*SecretResult, error) {
	out := new(SecretResult)
	err := c.cc.Invoke(ctx, "/authn.v1.AuthnService/EnterSecret", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthnServiceServer is the server API for AuthnService service.
// All implementations must embed UnimplementedAuthnServiceServer
// for forward compatibility
type AuthnServiceServer interface {
	// Enter enters authn command.
	Enter(*Cmd, AuthnService_EnterServer) error
	// EnterSecret enters needed secrets after specific CmdStatus has received.
	EnterSecret(context.Context, *SecretMsg) (*SecretResult, error)
	mustEmbedUnimplementedAuthnServiceServer()
}

// UnimplementedAuthnServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthnServiceServer struct {
}

func (UnimplementedAuthnServiceServer) Enter(*Cmd, AuthnService_EnterServer) error {
	return status.Errorf(codes.Unimplemented, "method Enter not implemented")
}
func (UnimplementedAuthnServiceServer) EnterSecret(context.Context, *SecretMsg) (*SecretResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnterSecret not implemented")
}
func (UnimplementedAuthnServiceServer) mustEmbedUnimplementedAuthnServiceServer() {}

// UnsafeAuthnServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthnServiceServer will
// result in compilation errors.
type UnsafeAuthnServiceServer interface {
	mustEmbedUnimplementedAuthnServiceServer()
}

func RegisterAuthnServiceServer(s grpc.ServiceRegistrar, srv AuthnServiceServer) {
	s.RegisterService(&AuthnService_ServiceDesc, srv)
}

func _AuthnService_Enter_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Cmd)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AuthnServiceServer).Enter(m, &authnServiceEnterServer{stream})
}

type AuthnService_EnterServer interface {
	Send(*CmdStatus) error
	grpc.ServerStream
}

type authnServiceEnterServer struct {
	grpc.ServerStream
}

func (x *authnServiceEnterServer) Send(m *CmdStatus) error {
	return x.ServerStream.SendMsg(m)
}

func _AuthnService_EnterSecret_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SecretMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthnServiceServer).EnterSecret(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/authn.v1.AuthnService/EnterSecret",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthnServiceServer).EnterSecret(ctx, req.(*SecretMsg))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthnService_ServiceDesc is the grpc.ServiceDesc for AuthnService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthnService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "authn.v1.AuthnService",
	HandlerType: (*AuthnServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EnterSecret",
			Handler:    _AuthnService_EnterSecret_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Enter",
			Handler:       _AuthnService_Enter_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "authn.proto",
}
