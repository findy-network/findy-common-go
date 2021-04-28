// Copyright 2020 Harri @ OP Techlab.
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: agency.proto

// Package ops.v1 is the first version of findy gRPC API. As long as we'll not
// have changes that aren't backward compatible, we can just update the API.
// The gRPC itself will take care off that, like adding a new fields to
// messages. We just need to follow the gRPC practises and rules.
//
// As said, as long as we can maintain backward compatibility, we are working
// with version 1.0.  The version 2.0 will be introduced when we cannot solve
// something only with the version 1.0. The 2.0 will include all the current
// APIs of 1.0 and we support them both together until the decision shall be
// made to depracate 1.0 totally. The deprecation rules will be specified
// later.

package v1

import (
	v1 "github.com/findy-network/findy-common-go/grpc/agency/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Cmd_Type int32

const (
	Cmd_PING    Cmd_Type = 0
	Cmd_LOGGING Cmd_Type = 1
	Cmd_COUNT   Cmd_Type = 2
)

// Enum value maps for Cmd_Type.
var (
	Cmd_Type_name = map[int32]string{
		0: "PING",
		1: "LOGGING",
		2: "COUNT",
	}
	Cmd_Type_value = map[string]int32{
		"PING":    0,
		"LOGGING": 1,
		"COUNT":   2,
	}
)

func (x Cmd_Type) Enum() *Cmd_Type {
	p := new(Cmd_Type)
	*p = x
	return p
}

func (x Cmd_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Cmd_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_agency_proto_enumTypes[0].Descriptor()
}

func (Cmd_Type) Type() protoreflect.EnumType {
	return &file_agency_proto_enumTypes[0]
}

func (x Cmd_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Cmd_Type.Descriptor instead.
func (Cmd_Type) EnumDescriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{4, 0}
}

// Onboarding is structure for cloud agent (CA) onboarding.
type Onboarding struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Email string `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"` // email is then name or handle used for pointing the CA.
}

func (x *Onboarding) Reset() {
	*x = Onboarding{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Onboarding) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Onboarding) ProtoMessage() {}

func (x *Onboarding) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Onboarding.ProtoReflect.Descriptor instead.
func (*Onboarding) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{0}
}

func (x *Onboarding) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

// OnboardResult is structure to transport Onboarding result.
type OnboardResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok     bool                    `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`        // result if Onboarding was successful.
	Result *OnboardResult_OKResult `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"` // Instance of the OK result.
}

func (x *OnboardResult) Reset() {
	*x = OnboardResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnboardResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnboardResult) ProtoMessage() {}

func (x *OnboardResult) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnboardResult.ProtoReflect.Descriptor instead.
func (*OnboardResult) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{1}
}

func (x *OnboardResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *OnboardResult) GetResult() *OnboardResult_OKResult {
	if x != nil {
		return x.Result
	}
	return nil
}

//
//AgencyStatus is message returned by PSMHook. These status messages encapsulates
//protocol state machine information. The message has its own id. The protocol
//specific information comes in ProtocolStatus. Outside of the actual protocol but
//relevant information are current agent DID and Connection ID which arrives in
//own fields.
type AgencyStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID             string             `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	ProtocolStatus *v1.ProtocolStatus `protobuf:"bytes,2,opt,name=protocol_status,json=protocolStatus,proto3" json:"protocol_status,omitempty"` // Detailed protocol information
	DID            string             `protobuf:"bytes,3,opt,name=DID,proto3" json:"DID,omitempty"`                                             // Agent DID if available
	ConnectionID   string             `protobuf:"bytes,4,opt,name=connectionID,proto3" json:"connectionID,omitempty"`                           // Connection (pairwise) ID if available
}

func (x *AgencyStatus) Reset() {
	*x = AgencyStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AgencyStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgencyStatus) ProtoMessage() {}

func (x *AgencyStatus) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AgencyStatus.ProtoReflect.Descriptor instead.
func (*AgencyStatus) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{2}
}

func (x *AgencyStatus) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *AgencyStatus) GetProtocolStatus() *v1.ProtocolStatus {
	if x != nil {
		return x.ProtocolStatus
	}
	return nil
}

func (x *AgencyStatus) GetDID() string {
	if x != nil {
		return x.DID
	}
	return ""
}

func (x *AgencyStatus) GetConnectionID() string {
	if x != nil {
		return x.ConnectionID
	}
	return ""
}

// DataHook is structure identify data hook.
type DataHook struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID string `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"` // UUID to uniquely identify the hook
}

func (x *DataHook) Reset() {
	*x = DataHook{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataHook) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataHook) ProtoMessage() {}

func (x *DataHook) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataHook.ProtoReflect.Descriptor instead.
func (*DataHook) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{3}
}

func (x *DataHook) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

// Cmd is structure to transport agency cmds.
type Cmd struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type Cmd_Type `protobuf:"varint,1,opt,name=type,proto3,enum=ops.v1.Cmd_Type" json:"type,omitempty"`
	// Request is the structure to gather cmd specific arguments to type fields.
	//
	// Types that are assignable to Request:
	//	*Cmd_Logging
	Request isCmd_Request `protobuf_oneof:"Request"`
}

func (x *Cmd) Reset() {
	*x = Cmd{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cmd) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cmd) ProtoMessage() {}

func (x *Cmd) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cmd.ProtoReflect.Descriptor instead.
func (*Cmd) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{4}
}

func (x *Cmd) GetType() Cmd_Type {
	if x != nil {
		return x.Type
	}
	return Cmd_PING
}

func (m *Cmd) GetRequest() isCmd_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (x *Cmd) GetLogging() string {
	if x, ok := x.GetRequest().(*Cmd_Logging); ok {
		return x.Logging
	}
	return ""
}

type isCmd_Request interface {
	isCmd_Request()
}

type Cmd_Logging struct {
	Logging string `protobuf:"bytes,2,opt,name=logging,proto3,oneof"` // Type is LOGGING includes argument string.
}

func (*Cmd_Logging) isCmd_Request() {}

// CmdReturn is structure to return cmd results.
type CmdReturn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type Cmd_Type `protobuf:"varint,1,opt,name=type,proto3,enum=ops.v1.Cmd_Type" json:"type,omitempty"`
	// Types that are assignable to Response:
	//	*CmdReturn_Ping
	//	*CmdReturn_Count
	Response isCmdReturn_Response `protobuf_oneof:"Response"`
}

func (x *CmdReturn) Reset() {
	*x = CmdReturn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CmdReturn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CmdReturn) ProtoMessage() {}

func (x *CmdReturn) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CmdReturn.ProtoReflect.Descriptor instead.
func (*CmdReturn) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{5}
}

func (x *CmdReturn) GetType() Cmd_Type {
	if x != nil {
		return x.Type
	}
	return Cmd_PING
}

func (m *CmdReturn) GetResponse() isCmdReturn_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *CmdReturn) GetPing() string {
	if x, ok := x.GetResponse().(*CmdReturn_Ping); ok {
		return x.Ping
	}
	return ""
}

func (x *CmdReturn) GetCount() string {
	if x, ok := x.GetResponse().(*CmdReturn_Count); ok {
		return x.Count
	}
	return ""
}

type isCmdReturn_Response interface {
	isCmdReturn_Response()
}

type CmdReturn_Ping struct {
	Ping string `protobuf:"bytes,2,opt,name=ping,proto3,oneof"` // Ping cmd's result.
}

type CmdReturn_Count struct {
	Count string `protobuf:"bytes,3,opt,name=count,proto3,oneof"` // Count cmd's result.
}

func (*CmdReturn_Ping) isCmdReturn_Response() {}

func (*CmdReturn_Count) isCmdReturn_Response() {}

type OnboardResult_OKResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JWT            string `protobuf:"bytes,1,opt,name=JWT,proto3" json:"JWT,omitempty"`                       // pregenerated JWT token, mostly usefull for development.
	CADID          string `protobuf:"bytes,2,opt,name=CADID,proto3" json:"CADID,omitempty"`                   // Cloud Agent DID. The UID for CA.
	InvitationJSON string `protobuf:"bytes,3,opt,name=invitationJSON,proto3" json:"invitationJSON,omitempty"` // pregenerated Invitation, mostly in dev use.
}

func (x *OnboardResult_OKResult) Reset() {
	*x = OnboardResult_OKResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agency_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnboardResult_OKResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnboardResult_OKResult) ProtoMessage() {}

func (x *OnboardResult_OKResult) ProtoReflect() protoreflect.Message {
	mi := &file_agency_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnboardResult_OKResult.ProtoReflect.Descriptor instead.
func (*OnboardResult_OKResult) Descriptor() ([]byte, []int) {
	return file_agency_proto_rawDescGZIP(), []int{1, 0}
}

func (x *OnboardResult_OKResult) GetJWT() string {
	if x != nil {
		return x.JWT
	}
	return ""
}

func (x *OnboardResult_OKResult) GetCADID() string {
	if x != nil {
		return x.CADID
	}
	return ""
}

func (x *OnboardResult_OKResult) GetInvitationJSON() string {
	if x != nil {
		return x.InvitationJSON
	}
	return ""
}

var File_agency_proto protoreflect.FileDescriptor

var file_agency_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x61, 0x67, 0x65, 0x6e, 0x63, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6f, 0x70, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x0e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x22, 0x0a, 0x0a, 0x4f, 0x6e, 0x62, 0x6f, 0x61, 0x72,
	0x64, 0x69, 0x6e, 0x67, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x22, 0xb3, 0x01, 0x0a, 0x0d, 0x4f,
	0x6e, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b, 0x12, 0x36, 0x0a, 0x06,
	0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x6f,
	0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x6e, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x2e, 0x4f, 0x4b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x1a, 0x5a, 0x0a, 0x08, 0x4f, 0x4b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x10, 0x0a, 0x03, 0x4a, 0x57, 0x54, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x4a,
	0x57, 0x54, 0x12, 0x14, 0x0a, 0x05, 0x43, 0x41, 0x44, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x43, 0x41, 0x44, 0x49, 0x44, 0x12, 0x26, 0x0a, 0x0e, 0x69, 0x6e, 0x76, 0x69,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4a, 0x53, 0x4f, 0x4e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4a, 0x53, 0x4f, 0x4e,
	0x22, 0x98, 0x01, 0x0a, 0x0c, 0x41, 0x67, 0x65, 0x6e, 0x63, 0x79, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49,
	0x44, 0x12, 0x42, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x61, 0x67, 0x65,
	0x6e, 0x63, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x44, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x44, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x22, 0x1a, 0x0a, 0x08, 0x44,
	0x61, 0x74, 0x61, 0x48, 0x6f, 0x6f, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x44, 0x22, 0x7c, 0x0a, 0x03, 0x43, 0x6d, 0x64, 0x12, 0x24,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x6f,
	0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6d, 0x64, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x07, 0x6c, 0x6f, 0x67, 0x67, 0x69, 0x6e, 0x67, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x67, 0x69, 0x6e, 0x67,
	0x22, 0x28, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x49, 0x4e, 0x47,
	0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x4c, 0x4f, 0x47, 0x47, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12,
	0x09, 0x0a, 0x05, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x10, 0x02, 0x42, 0x09, 0x0a, 0x07, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x6b, 0x0a, 0x09, 0x43, 0x6d, 0x64, 0x52, 0x65, 0x74, 0x75,
	0x72, 0x6e, 0x12, 0x24, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x10, 0x2e, 0x6f, 0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6d, 0x64, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x04, 0x70, 0x69, 0x6e, 0x67,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x12, 0x16,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x42, 0x0a, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x32, 0x7e, 0x0a, 0x0d, 0x41, 0x67, 0x65, 0x6e, 0x63, 0x79, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x07, 0x50, 0x53, 0x4d, 0x48, 0x6f, 0x6f, 0x6b, 0x12, 0x10,
	0x2e, 0x6f, 0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x48, 0x6f, 0x6f, 0x6b,
	0x1a, 0x14, 0x2e, 0x6f, 0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x63, 0x79,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x30, 0x01, 0x12, 0x36, 0x0a, 0x07, 0x4f, 0x6e,
	0x62, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x12, 0x2e, 0x6f, 0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4f,
	0x6e, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x1a, 0x15, 0x2e, 0x6f, 0x70, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x4f, 0x6e, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x22, 0x00, 0x32, 0x3a, 0x0a, 0x0d, 0x44, 0x65, 0x76, 0x4f, 0x70, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x65, 0x72, 0x12, 0x0b, 0x2e, 0x6f,
	0x70, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6d, 0x64, 0x1a, 0x11, 0x2e, 0x6f, 0x70, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x6d, 0x64, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x22, 0x00, 0x42, 0x36,
	0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6e,
	0x64, 0x79, 0x2d, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x66, 0x69, 0x6e, 0x64, 0x79,
	0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2d, 0x67, 0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f,
	0x6f, 0x70, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_agency_proto_rawDescOnce sync.Once
	file_agency_proto_rawDescData = file_agency_proto_rawDesc
)

func file_agency_proto_rawDescGZIP() []byte {
	file_agency_proto_rawDescOnce.Do(func() {
		file_agency_proto_rawDescData = protoimpl.X.CompressGZIP(file_agency_proto_rawDescData)
	})
	return file_agency_proto_rawDescData
}

var file_agency_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_agency_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_agency_proto_goTypes = []interface{}{
	(Cmd_Type)(0),                  // 0: ops.v1.Cmd.Type
	(*Onboarding)(nil),             // 1: ops.v1.Onboarding
	(*OnboardResult)(nil),          // 2: ops.v1.OnboardResult
	(*AgencyStatus)(nil),           // 3: ops.v1.AgencyStatus
	(*DataHook)(nil),               // 4: ops.v1.DataHook
	(*Cmd)(nil),                    // 5: ops.v1.Cmd
	(*CmdReturn)(nil),              // 6: ops.v1.CmdReturn
	(*OnboardResult_OKResult)(nil), // 7: ops.v1.OnboardResult.OKResult
	(*v1.ProtocolStatus)(nil),      // 8: agency.v1.ProtocolStatus
}
var file_agency_proto_depIdxs = []int32{
	7, // 0: ops.v1.OnboardResult.result:type_name -> ops.v1.OnboardResult.OKResult
	8, // 1: ops.v1.AgencyStatus.protocol_status:type_name -> agency.v1.ProtocolStatus
	0, // 2: ops.v1.Cmd.type:type_name -> ops.v1.Cmd.Type
	0, // 3: ops.v1.CmdReturn.type:type_name -> ops.v1.Cmd.Type
	4, // 4: ops.v1.AgencyService.PSMHook:input_type -> ops.v1.DataHook
	1, // 5: ops.v1.AgencyService.Onboard:input_type -> ops.v1.Onboarding
	5, // 6: ops.v1.DevOpsService.Enter:input_type -> ops.v1.Cmd
	3, // 7: ops.v1.AgencyService.PSMHook:output_type -> ops.v1.AgencyStatus
	2, // 8: ops.v1.AgencyService.Onboard:output_type -> ops.v1.OnboardResult
	6, // 9: ops.v1.DevOpsService.Enter:output_type -> ops.v1.CmdReturn
	7, // [7:10] is the sub-list for method output_type
	4, // [4:7] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_agency_proto_init() }
func file_agency_proto_init() {
	if File_agency_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_agency_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Onboarding); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnboardResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AgencyStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataHook); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cmd); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CmdReturn); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_agency_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnboardResult_OKResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_agency_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*Cmd_Logging)(nil),
	}
	file_agency_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*CmdReturn_Ping)(nil),
		(*CmdReturn_Count)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_agency_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_agency_proto_goTypes,
		DependencyIndexes: file_agency_proto_depIdxs,
		EnumInfos:         file_agency_proto_enumTypes,
		MessageInfos:      file_agency_proto_msgTypes,
	}.Build()
	File_agency_proto = out.File
	file_agency_proto_rawDesc = nil
	file_agency_proto_goTypes = nil
	file_agency_proto_depIdxs = nil
}
