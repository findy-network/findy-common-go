// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// AgentServiceClient is the client API for AgentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AgentServiceClient interface {
	// Listen is bidirectional function to stream AgentStatus. ClientID must be
	// unique. AgentStatus includes only enough information to access the actual
	// PSM and DIDComm connection with the ProtocolService.Status function.
	// Summary: you listen your agent but work with our protocols.
	Listen(ctx context.Context, in *ClientID, opts ...grpc.CallOption) (AgentService_ListenClient, error)
	// Wait is bidirectional function to stream service agent Questions. With
	// Wait you listen your agent and if it's Issuing or Verifying VC it needs
	// more information and immetiate answers from you. For instance, if a proof
	// can be validated. Note! if your agent is only casual Holder it doesn't
	// need to answer any of these questions. Holder communicate goes with
	// ProtocolService.Resume(). Please see Give for more information.
	Wait(ctx context.Context, in *ClientID, opts ...grpc.CallOption) (AgentService_WaitClient, error)
	// Give is function to answer to Questions sent from CA and arived from Wait
	// function. Questions have ID and clientID which should be used when
	// answering the questions.
	Give(ctx context.Context, in *Answer, opts ...grpc.CallOption) (*ClientID, error)
	// CreateInvitation returns an invitation according to InvitationBase.
	CreateInvitation(ctx context.Context, in *InvitationBase, opts ...grpc.CallOption) (*Invitation, error)
	// SetImplId sets implementation ID for the clould agent. It should be "grpc".
	// TODO: REMOVE!! Check Agency implementation first. We still need something
	// for this. At least the autoaccept mode for now.
	// TODO: Rename? Rethink logic: SetSAMode(), etc.?
	SetImplId(ctx context.Context, in *SAImplementation, opts ...grpc.CallOption) (*SAImplementation, error)
	// Ping pings the cloud agent.
	Ping(ctx context.Context, in *PingMsg, opts ...grpc.CallOption) (*PingMsg, error)
	// CreateSchema creates a new schema and writes it to ledger.
	CreateSchema(ctx context.Context, in *SchemaCreate, opts ...grpc.CallOption) (*Schema, error)
	// CreateCredDef creates a new credential definition to wallet and writes it
	// to the ledger. Note! With current indysdk VC structure the running time is
	// long, like 10-20 seconds.
	CreateCredDef(ctx context.Context, in *CredDefCreate, opts ...grpc.CallOption) (*CredDef, error)
	// GetSchema returns a schema structure.
	GetSchema(ctx context.Context, in *Schema, opts ...grpc.CallOption) (*SchemaData, error)
	// GetCredDef returns a credential definition.
	GetCredDef(ctx context.Context, in *CredDef, opts ...grpc.CallOption) (*CredDefData, error)
}

type agentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentServiceClient(cc grpc.ClientConnInterface) AgentServiceClient {
	return &agentServiceClient{cc}
}

func (c *agentServiceClient) Listen(ctx context.Context, in *ClientID, opts ...grpc.CallOption) (AgentService_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &_AgentService_serviceDesc.Streams[0], "/agency.v1.AgentService/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &agentServiceListenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AgentService_ListenClient interface {
	Recv() (*AgentStatus, error)
	grpc.ClientStream
}

type agentServiceListenClient struct {
	grpc.ClientStream
}

func (x *agentServiceListenClient) Recv() (*AgentStatus, error) {
	m := new(AgentStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *agentServiceClient) Wait(ctx context.Context, in *ClientID, opts ...grpc.CallOption) (AgentService_WaitClient, error) {
	stream, err := c.cc.NewStream(ctx, &_AgentService_serviceDesc.Streams[1], "/agency.v1.AgentService/Wait", opts...)
	if err != nil {
		return nil, err
	}
	x := &agentServiceWaitClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AgentService_WaitClient interface {
	Recv() (*Question, error)
	grpc.ClientStream
}

type agentServiceWaitClient struct {
	grpc.ClientStream
}

func (x *agentServiceWaitClient) Recv() (*Question, error) {
	m := new(Question)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *agentServiceClient) Give(ctx context.Context, in *Answer, opts ...grpc.CallOption) (*ClientID, error) {
	out := new(ClientID)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/Give", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) CreateInvitation(ctx context.Context, in *InvitationBase, opts ...grpc.CallOption) (*Invitation, error) {
	out := new(Invitation)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/CreateInvitation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) SetImplId(ctx context.Context, in *SAImplementation, opts ...grpc.CallOption) (*SAImplementation, error) {
	out := new(SAImplementation)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/SetImplId", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) Ping(ctx context.Context, in *PingMsg, opts ...grpc.CallOption) (*PingMsg, error) {
	out := new(PingMsg)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) CreateSchema(ctx context.Context, in *SchemaCreate, opts ...grpc.CallOption) (*Schema, error) {
	out := new(Schema)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/CreateSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) CreateCredDef(ctx context.Context, in *CredDefCreate, opts ...grpc.CallOption) (*CredDef, error) {
	out := new(CredDef)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/CreateCredDef", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) GetSchema(ctx context.Context, in *Schema, opts ...grpc.CallOption) (*SchemaData, error) {
	out := new(SchemaData)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/GetSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentServiceClient) GetCredDef(ctx context.Context, in *CredDef, opts ...grpc.CallOption) (*CredDefData, error) {
	out := new(CredDefData)
	err := c.cc.Invoke(ctx, "/agency.v1.AgentService/GetCredDef", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServiceServer is the server API for AgentService service.
// All implementations must embed UnimplementedAgentServiceServer
// for forward compatibility
type AgentServiceServer interface {
	// Listen is bidirectional function to stream AgentStatus. ClientID must be
	// unique. AgentStatus includes only enough information to access the actual
	// PSM and DIDComm connection with the ProtocolService.Status function.
	// Summary: you listen your agent but work with our protocols.
	Listen(*ClientID, AgentService_ListenServer) error
	// Wait is bidirectional function to stream service agent Questions. With
	// Wait you listen your agent and if it's Issuing or Verifying VC it needs
	// more information and immetiate answers from you. For instance, if a proof
	// can be validated. Note! if your agent is only casual Holder it doesn't
	// need to answer any of these questions. Holder communicate goes with
	// ProtocolService.Resume(). Please see Give for more information.
	Wait(*ClientID, AgentService_WaitServer) error
	// Give is function to answer to Questions sent from CA and arived from Wait
	// function. Questions have ID and clientID which should be used when
	// answering the questions.
	Give(context.Context, *Answer) (*ClientID, error)
	// CreateInvitation returns an invitation according to InvitationBase.
	CreateInvitation(context.Context, *InvitationBase) (*Invitation, error)
	// SetImplId sets implementation ID for the clould agent. It should be "grpc".
	// TODO: REMOVE!! Check Agency implementation first. We still need something
	// for this. At least the autoaccept mode for now.
	// TODO: Rename? Rethink logic: SetSAMode(), etc.?
	SetImplId(context.Context, *SAImplementation) (*SAImplementation, error)
	// Ping pings the cloud agent.
	Ping(context.Context, *PingMsg) (*PingMsg, error)
	// CreateSchema creates a new schema and writes it to ledger.
	CreateSchema(context.Context, *SchemaCreate) (*Schema, error)
	// CreateCredDef creates a new credential definition to wallet and writes it
	// to the ledger. Note! With current indysdk VC structure the running time is
	// long, like 10-20 seconds.
	CreateCredDef(context.Context, *CredDefCreate) (*CredDef, error)
	// GetSchema returns a schema structure.
	GetSchema(context.Context, *Schema) (*SchemaData, error)
	// GetCredDef returns a credential definition.
	GetCredDef(context.Context, *CredDef) (*CredDefData, error)
	mustEmbedUnimplementedAgentServiceServer()
}

// UnimplementedAgentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAgentServiceServer struct {
}

func (UnimplementedAgentServiceServer) Listen(*ClientID, AgentService_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedAgentServiceServer) Wait(*ClientID, AgentService_WaitServer) error {
	return status.Errorf(codes.Unimplemented, "method Wait not implemented")
}
func (UnimplementedAgentServiceServer) Give(context.Context, *Answer) (*ClientID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Give not implemented")
}
func (UnimplementedAgentServiceServer) CreateInvitation(context.Context, *InvitationBase) (*Invitation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInvitation not implemented")
}
func (UnimplementedAgentServiceServer) SetImplId(context.Context, *SAImplementation) (*SAImplementation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetImplId not implemented")
}
func (UnimplementedAgentServiceServer) Ping(context.Context, *PingMsg) (*PingMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedAgentServiceServer) CreateSchema(context.Context, *SchemaCreate) (*Schema, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSchema not implemented")
}
func (UnimplementedAgentServiceServer) CreateCredDef(context.Context, *CredDefCreate) (*CredDef, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCredDef not implemented")
}
func (UnimplementedAgentServiceServer) GetSchema(context.Context, *Schema) (*SchemaData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSchema not implemented")
}
func (UnimplementedAgentServiceServer) GetCredDef(context.Context, *CredDef) (*CredDefData, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCredDef not implemented")
}
func (UnimplementedAgentServiceServer) mustEmbedUnimplementedAgentServiceServer() {}

// UnsafeAgentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AgentServiceServer will
// result in compilation errors.
type UnsafeAgentServiceServer interface {
	mustEmbedUnimplementedAgentServiceServer()
}

func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	s.RegisterService(&_AgentService_serviceDesc, srv)
}

func _AgentService_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ClientID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AgentServiceServer).Listen(m, &agentServiceListenServer{stream})
}

type AgentService_ListenServer interface {
	Send(*AgentStatus) error
	grpc.ServerStream
}

type agentServiceListenServer struct {
	grpc.ServerStream
}

func (x *agentServiceListenServer) Send(m *AgentStatus) error {
	return x.ServerStream.SendMsg(m)
}

func _AgentService_Wait_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ClientID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AgentServiceServer).Wait(m, &agentServiceWaitServer{stream})
}

type AgentService_WaitServer interface {
	Send(*Question) error
	grpc.ServerStream
}

type agentServiceWaitServer struct {
	grpc.ServerStream
}

func (x *agentServiceWaitServer) Send(m *Question) error {
	return x.ServerStream.SendMsg(m)
}

func _AgentService_Give_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Answer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).Give(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/Give",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).Give(ctx, req.(*Answer))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_CreateInvitation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvitationBase)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).CreateInvitation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/CreateInvitation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).CreateInvitation(ctx, req.(*InvitationBase))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_SetImplId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SAImplementation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).SetImplId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/SetImplId",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).SetImplId(ctx, req.(*SAImplementation))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).Ping(ctx, req.(*PingMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_CreateSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SchemaCreate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).CreateSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/CreateSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).CreateSchema(ctx, req.(*SchemaCreate))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_CreateCredDef_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CredDefCreate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).CreateCredDef(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/CreateCredDef",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).CreateCredDef(ctx, req.(*CredDefCreate))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_GetSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Schema)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).GetSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/GetSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).GetSchema(ctx, req.(*Schema))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_GetCredDef_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CredDef)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).GetCredDef(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/agency.v1.AgentService/GetCredDef",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).GetCredDef(ctx, req.(*CredDef))
	}
	return interceptor(ctx, in, info, handler)
}

var _AgentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "agency.v1.AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Give",
			Handler:    _AgentService_Give_Handler,
		},
		{
			MethodName: "CreateInvitation",
			Handler:    _AgentService_CreateInvitation_Handler,
		},
		{
			MethodName: "SetImplId",
			Handler:    _AgentService_SetImplId_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _AgentService_Ping_Handler,
		},
		{
			MethodName: "CreateSchema",
			Handler:    _AgentService_CreateSchema_Handler,
		},
		{
			MethodName: "CreateCredDef",
			Handler:    _AgentService_CreateCredDef_Handler,
		},
		{
			MethodName: "GetSchema",
			Handler:    _AgentService_GetSchema_Handler,
		},
		{
			MethodName: "GetCredDef",
			Handler:    _AgentService_GetCredDef_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _AgentService_Listen_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Wait",
			Handler:       _AgentService_Wait_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "agent.proto",
}
