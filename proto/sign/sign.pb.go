// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: sign/sign.proto

package sign

import (
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

type SignMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Msg:
	//
	//	*SignMessage_SignRequestTo1Output
	//	*SignMessage_SignRound1To2Output
	//	*SignMessage_SignRound2To3Output
	//	*SignMessage_SignRound3To4Output
	//	*SignMessage_SignRound4ToResponseOutput
	Msg isSignMessage_Msg `protobuf_oneof:"msg"`
}

func (x *SignMessage) Reset() {
	*x = SignMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignMessage) ProtoMessage() {}

func (x *SignMessage) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignMessage.ProtoReflect.Descriptor instead.
func (*SignMessage) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{0}
}

func (m *SignMessage) GetMsg() isSignMessage_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *SignMessage) GetSignRequestTo1Output() *SignRequestTo1Output {
	if x, ok := x.GetMsg().(*SignMessage_SignRequestTo1Output); ok {
		return x.SignRequestTo1Output
	}
	return nil
}

func (x *SignMessage) GetSignRound1To2Output() *SignRound1To2Output {
	if x, ok := x.GetMsg().(*SignMessage_SignRound1To2Output); ok {
		return x.SignRound1To2Output
	}
	return nil
}

func (x *SignMessage) GetSignRound2To3Output() *SignRound2To3Output {
	if x, ok := x.GetMsg().(*SignMessage_SignRound2To3Output); ok {
		return x.SignRound2To3Output
	}
	return nil
}

func (x *SignMessage) GetSignRound3To4Output() *SignRound3To4Output {
	if x, ok := x.GetMsg().(*SignMessage_SignRound3To4Output); ok {
		return x.SignRound3To4Output
	}
	return nil
}

func (x *SignMessage) GetSignRound4ToResponseOutput() *SignRound4ToResponseOutput {
	if x, ok := x.GetMsg().(*SignMessage_SignRound4ToResponseOutput); ok {
		return x.SignRound4ToResponseOutput
	}
	return nil
}

type isSignMessage_Msg interface {
	isSignMessage_Msg()
}

type SignMessage_SignRequestTo1Output struct {
	SignRequestTo1Output *SignRequestTo1Output `protobuf:"bytes,1,opt,name=signRequestTo1Output,proto3,oneof"`
}

type SignMessage_SignRound1To2Output struct {
	SignRound1To2Output *SignRound1To2Output `protobuf:"bytes,2,opt,name=signRound1To2Output,proto3,oneof"`
}

type SignMessage_SignRound2To3Output struct {
	SignRound2To3Output *SignRound2To3Output `protobuf:"bytes,3,opt,name=signRound2To3Output,proto3,oneof"`
}

type SignMessage_SignRound3To4Output struct {
	SignRound3To4Output *SignRound3To4Output `protobuf:"bytes,4,opt,name=signRound3To4Output,proto3,oneof"`
}

type SignMessage_SignRound4ToResponseOutput struct {
	SignRound4ToResponseOutput *SignRound4ToResponseOutput `protobuf:"bytes,5,opt,name=signRound4ToResponseOutput,proto3,oneof"`
}

func (*SignMessage_SignRequestTo1Output) isSignMessage_Msg() {}

func (*SignMessage_SignRound1To2Output) isSignMessage_Msg() {}

func (*SignMessage_SignRound2To3Output) isSignMessage_Msg() {}

func (*SignMessage_SignRound3To4Output) isSignMessage_Msg() {}

func (*SignMessage_SignRound4ToResponseOutput) isSignMessage_Msg() {}

// 요청 메시지 DTO
type SignRequestMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	SecretKey string `protobuf:"bytes,2,opt,name=secret_key,json=secretKey,proto3" json:"secret_key,omitempty"` // encoded base64
	TxOrigin  string `protobuf:"bytes,3,opt,name=tx_origin,json=txOrigin,proto3" json:"tx_origin,omitempty"`    // encoded base64
}

func (x *SignRequestMessage) Reset() {
	*x = SignRequestMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRequestMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRequestMessage) ProtoMessage() {}

func (x *SignRequestMessage) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRequestMessage.ProtoReflect.Descriptor instead.
func (*SignRequestMessage) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{1}
}

func (x *SignRequestMessage) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *SignRequestMessage) GetSecretKey() string {
	if x != nil {
		return x.SecretKey
	}
	return ""
}

func (x *SignRequestMessage) GetTxOrigin() string {
	if x != nil {
		return x.TxOrigin
	}
	return ""
}

// 요청 -> 라운드 1
type SignRequestTo1Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	SecretKey []byte `protobuf:"bytes,2,opt,name=secret_key,json=secretKey,proto3" json:"secret_key,omitempty"` // decode base64
	TxOrigin  []byte `protobuf:"bytes,3,opt,name=tx_origin,json=txOrigin,proto3" json:"tx_origin,omitempty"`    //  decode base64
}

func (x *SignRequestTo1Output) Reset() {
	*x = SignRequestTo1Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRequestTo1Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRequestTo1Output) ProtoMessage() {}

func (x *SignRequestTo1Output) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRequestTo1Output.ProtoReflect.Descriptor instead.
func (*SignRequestTo1Output) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{2}
}

func (x *SignRequestTo1Output) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *SignRequestTo1Output) GetSecretKey() []byte {
	if x != nil {
		return x.SecretKey
	}
	return nil
}

func (x *SignRequestTo1Output) GetTxOrigin() []byte {
	if x != nil {
		return x.TxOrigin
	}
	return nil
}

// 라운드 1 -> 라운드 2
type SignRound1To2Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address   string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	SecretKey []byte `protobuf:"bytes,2,opt,name=secret_key,json=secretKey,proto3" json:"secret_key,omitempty"`
	TxOrigin  []byte `protobuf:"bytes,3,opt,name=tx_origin,json=txOrigin,proto3" json:"tx_origin,omitempty"`
	Payload   []byte `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *SignRound1To2Output) Reset() {
	*x = SignRound1To2Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRound1To2Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRound1To2Output) ProtoMessage() {}

func (x *SignRound1To2Output) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRound1To2Output.ProtoReflect.Descriptor instead.
func (*SignRound1To2Output) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{3}
}

func (x *SignRound1To2Output) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *SignRound1To2Output) GetSecretKey() []byte {
	if x != nil {
		return x.SecretKey
	}
	return nil
}

func (x *SignRound1To2Output) GetTxOrigin() []byte {
	if x != nil {
		return x.TxOrigin
	}
	return nil
}

func (x *SignRound1To2Output) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

// 라운드 2 -> 라운드 3
type SignRound2To3Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *SignRound2To3Output) Reset() {
	*x = SignRound2To3Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRound2To3Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRound2To3Output) ProtoMessage() {}

func (x *SignRound2To3Output) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRound2To3Output.ProtoReflect.Descriptor instead.
func (*SignRound2To3Output) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{4}
}

func (x *SignRound2To3Output) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

// 라운드 3 -> 라운드 4
type SignRound3To4Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *SignRound3To4Output) Reset() {
	*x = SignRound3To4Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRound3To4Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRound3To4Output) ProtoMessage() {}

func (x *SignRound3To4Output) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRound3To4Output.ProtoReflect.Descriptor instead.
func (*SignRound3To4Output) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{5}
}

func (x *SignRound3To4Output) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

// 라운드 4 -> 게이트웨이
type SignRound4ToResponseOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	V uint64 `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	R []byte `protobuf:"bytes,2,opt,name=r,proto3" json:"r,omitempty"`
	S []byte `protobuf:"bytes,3,opt,name=s,proto3" json:"s,omitempty"`
}

func (x *SignRound4ToResponseOutput) Reset() {
	*x = SignRound4ToResponseOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRound4ToResponseOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRound4ToResponseOutput) ProtoMessage() {}

func (x *SignRound4ToResponseOutput) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRound4ToResponseOutput.ProtoReflect.Descriptor instead.
func (*SignRound4ToResponseOutput) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{6}
}

func (x *SignRound4ToResponseOutput) GetV() uint64 {
	if x != nil {
		return x.V
	}
	return 0
}

func (x *SignRound4ToResponseOutput) GetR() []byte {
	if x != nil {
		return x.R
	}
	return nil
}

func (x *SignRound4ToResponseOutput) GetS() []byte {
	if x != nil {
		return x.S
	}
	return nil
}

// 게이트웨이 -> 응답
type SignResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	V        uint64 `protobuf:"varint,1,opt,name=v,proto3" json:"v,omitempty"`
	R        string `protobuf:"bytes,2,opt,name=r,proto3" json:"r,omitempty"` // encoded base64
	S        string `protobuf:"bytes,3,opt,name=s,proto3" json:"s,omitempty"` // encoded base64
	Duration int32  `protobuf:"varint,4,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *SignResponse) Reset() {
	*x = SignResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sign_sign_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignResponse) ProtoMessage() {}

func (x *SignResponse) ProtoReflect() protoreflect.Message {
	mi := &file_sign_sign_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignResponse.ProtoReflect.Descriptor instead.
func (*SignResponse) Descriptor() ([]byte, []int) {
	return file_sign_sign_proto_rawDescGZIP(), []int{7}
}

func (x *SignResponse) GetV() uint64 {
	if x != nil {
		return x.V
	}
	return 0
}

func (x *SignResponse) GetR() string {
	if x != nil {
		return x.R
	}
	return ""
}

func (x *SignResponse) GetS() string {
	if x != nil {
		return x.S
	}
	return ""
}

func (x *SignResponse) GetDuration() int32 {
	if x != nil {
		return x.Duration
	}
	return 0
}

var File_sign_sign_proto protoreflect.FileDescriptor

var file_sign_sign_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x69, 0x67, 0x6e, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x73, 0x69, 0x67, 0x6e, 0x22, 0xb7, 0x03, 0x0a, 0x0b, 0x53, 0x69, 0x67, 0x6e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x50, 0x0a, 0x14, 0x73, 0x69, 0x67, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x6f, 0x31, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x2e, 0x53, 0x69, 0x67,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x6f, 0x31, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x48, 0x00, 0x52, 0x14, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x54, 0x6f, 0x31, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x4d, 0x0a, 0x13, 0x73, 0x69, 0x67,
	0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x31, 0x54, 0x6f, 0x32, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x2e, 0x53, 0x69,
	0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x31, 0x54, 0x6f, 0x32, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x48, 0x00, 0x52, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x31, 0x54,
	0x6f, 0x32, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x4d, 0x0a, 0x13, 0x73, 0x69, 0x67, 0x6e,
	0x52, 0x6f, 0x75, 0x6e, 0x64, 0x32, 0x54, 0x6f, 0x33, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x2e, 0x53, 0x69, 0x67,
	0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x32, 0x54, 0x6f, 0x33, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x48, 0x00, 0x52, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x32, 0x54, 0x6f,
	0x33, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x4d, 0x0a, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x52,
	0x6f, 0x75, 0x6e, 0x64, 0x33, 0x54, 0x6f, 0x34, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x2e, 0x53, 0x69, 0x67, 0x6e,
	0x52, 0x6f, 0x75, 0x6e, 0x64, 0x33, 0x54, 0x6f, 0x34, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x48,
	0x00, 0x52, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x33, 0x54, 0x6f, 0x34,
	0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x62, 0x0a, 0x1a, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f,
	0x75, 0x6e, 0x64, 0x34, 0x54, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x73, 0x69, 0x67,
	0x6e, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x34, 0x54, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x48, 0x00, 0x52, 0x1a,
	0x73, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x34, 0x54, 0x6f, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x42, 0x05, 0x0a, 0x03, 0x6d, 0x73,
	0x67, 0x22, 0x6a, 0x0a, 0x12, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x4b, 0x65, 0x79,
	0x12, 0x1b, 0x0a, 0x09, 0x74, 0x78, 0x5f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x78, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x22, 0x6c, 0x0a,
	0x14, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x6f, 0x31, 0x4f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x1b,
	0x0a, 0x09, 0x74, 0x78, 0x5f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x08, 0x74, 0x78, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x22, 0x85, 0x01, 0x0a, 0x13,
	0x53, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x31, 0x54, 0x6f, 0x32, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1d, 0x0a,
	0x0a, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x09, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x1b, 0x0a, 0x09,
	0x74, 0x78, 0x5f, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x08, 0x74, 0x78, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x2f, 0x0a, 0x13, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e, 0x64,
	0x32, 0x54, 0x6f, 0x33, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x22, 0x2f, 0x0a, 0x13, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75, 0x6e,
	0x64, 0x33, 0x54, 0x6f, 0x34, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x46, 0x0a, 0x1a, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x6f, 0x75,
	0x6e, 0x64, 0x34, 0x54, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x12, 0x0c, 0x0a, 0x01, 0x76, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x01,
	0x76, 0x12, 0x0c, 0x0a, 0x01, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x72, 0x12,
	0x0c, 0x0a, 0x01, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x73, 0x22, 0x54, 0x0a,
	0x0c, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0c, 0x0a,
	0x01, 0x76, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x01, 0x76, 0x12, 0x0c, 0x0a, 0x01, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x72, 0x12, 0x0c, 0x0a, 0x01, 0x73, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x32, 0x3f, 0x0a, 0x0b, 0x53, 0x69, 0x67, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x30, 0x0a, 0x04, 0x53, 0x69, 0x67, 0x6e, 0x12, 0x11, 0x2e, 0x73, 0x69, 0x67,
	0x6e, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x11, 0x2e,
	0x73, 0x69, 0x67, 0x6e, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x28, 0x01, 0x30, 0x01, 0x42, 0x13, 0x5a, 0x11, 0x74, 0x65, 0x63, 0x64, 0x73, 0x61, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_sign_sign_proto_rawDescOnce sync.Once
	file_sign_sign_proto_rawDescData = file_sign_sign_proto_rawDesc
)

func file_sign_sign_proto_rawDescGZIP() []byte {
	file_sign_sign_proto_rawDescOnce.Do(func() {
		file_sign_sign_proto_rawDescData = protoimpl.X.CompressGZIP(file_sign_sign_proto_rawDescData)
	})
	return file_sign_sign_proto_rawDescData
}

var file_sign_sign_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_sign_sign_proto_goTypes = []any{
	(*SignMessage)(nil),                // 0: sign.SignMessage
	(*SignRequestMessage)(nil),         // 1: sign.SignRequestMessage
	(*SignRequestTo1Output)(nil),       // 2: sign.SignRequestTo1Output
	(*SignRound1To2Output)(nil),        // 3: sign.SignRound1To2Output
	(*SignRound2To3Output)(nil),        // 4: sign.SignRound2To3Output
	(*SignRound3To4Output)(nil),        // 5: sign.SignRound3To4Output
	(*SignRound4ToResponseOutput)(nil), // 6: sign.SignRound4ToResponseOutput
	(*SignResponse)(nil),               // 7: sign.SignResponse
}
var file_sign_sign_proto_depIdxs = []int32{
	2, // 0: sign.SignMessage.signRequestTo1Output:type_name -> sign.SignRequestTo1Output
	3, // 1: sign.SignMessage.signRound1To2Output:type_name -> sign.SignRound1To2Output
	4, // 2: sign.SignMessage.signRound2To3Output:type_name -> sign.SignRound2To3Output
	5, // 3: sign.SignMessage.signRound3To4Output:type_name -> sign.SignRound3To4Output
	6, // 4: sign.SignMessage.signRound4ToResponseOutput:type_name -> sign.SignRound4ToResponseOutput
	0, // 5: sign.SignService.Sign:input_type -> sign.SignMessage
	0, // 6: sign.SignService.Sign:output_type -> sign.SignMessage
	6, // [6:7] is the sub-list for method output_type
	5, // [5:6] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_sign_sign_proto_init() }
func file_sign_sign_proto_init() {
	if File_sign_sign_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sign_sign_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*SignMessage); i {
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
		file_sign_sign_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*SignRequestMessage); i {
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
		file_sign_sign_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*SignRequestTo1Output); i {
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
		file_sign_sign_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*SignRound1To2Output); i {
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
		file_sign_sign_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*SignRound2To3Output); i {
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
		file_sign_sign_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*SignRound3To4Output); i {
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
		file_sign_sign_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*SignRound4ToResponseOutput); i {
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
		file_sign_sign_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*SignResponse); i {
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
	file_sign_sign_proto_msgTypes[0].OneofWrappers = []any{
		(*SignMessage_SignRequestTo1Output)(nil),
		(*SignMessage_SignRound1To2Output)(nil),
		(*SignMessage_SignRound2To3Output)(nil),
		(*SignMessage_SignRound3To4Output)(nil),
		(*SignMessage_SignRound4ToResponseOutput)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sign_sign_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_sign_sign_proto_goTypes,
		DependencyIndexes: file_sign_sign_proto_depIdxs,
		MessageInfos:      file_sign_sign_proto_msgTypes,
	}.Build()
	File_sign_sign_proto = out.File
	file_sign_sign_proto_rawDesc = nil
	file_sign_sign_proto_goTypes = nil
	file_sign_sign_proto_depIdxs = nil
}
