// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: transaction/transaction.proto

package transaction

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

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Tx:
	//
	//	*Transaction_Ethereum
	//	*Transaction_Bitcoin
	Tx isTransaction_Tx `protobuf_oneof:"tx"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_transaction_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_transaction_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_transaction_transaction_proto_rawDescGZIP(), []int{0}
}

func (m *Transaction) GetTx() isTransaction_Tx {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (x *Transaction) GetEthereum() *EthereumTransaction {
	if x, ok := x.GetTx().(*Transaction_Ethereum); ok {
		return x.Ethereum
	}
	return nil
}

func (x *Transaction) GetBitcoin() *BitcoinTransaction {
	if x, ok := x.GetTx().(*Transaction_Bitcoin); ok {
		return x.Bitcoin
	}
	return nil
}

type isTransaction_Tx interface {
	isTransaction_Tx()
}

type Transaction_Ethereum struct {
	Ethereum *EthereumTransaction `protobuf:"bytes,1,opt,name=ethereum,proto3,oneof"`
}

type Transaction_Bitcoin struct {
	Bitcoin *BitcoinTransaction `protobuf:"bytes,2,opt,name=bitcoin,proto3,oneof"`
}

func (*Transaction_Ethereum) isTransaction_Tx() {}

func (*Transaction_Bitcoin) isTransaction_Tx() {}

type BitcoinTransaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Inputs   []*UTXO   `protobuf:"bytes,1,rep,name=inputs,proto3" json:"inputs,omitempty"`
	Outputs  []*Output `protobuf:"bytes,2,rep,name=outputs,proto3" json:"outputs,omitempty"`
	Version  uint32    `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
	LockTime uint32    `protobuf:"varint,4,opt,name=lock_time,json=lockTime,proto3" json:"lock_time,omitempty"`
}

func (x *BitcoinTransaction) Reset() {
	*x = BitcoinTransaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_transaction_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BitcoinTransaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BitcoinTransaction) ProtoMessage() {}

func (x *BitcoinTransaction) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_transaction_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BitcoinTransaction.ProtoReflect.Descriptor instead.
func (*BitcoinTransaction) Descriptor() ([]byte, []int) {
	return file_transaction_transaction_proto_rawDescGZIP(), []int{1}
}

func (x *BitcoinTransaction) GetInputs() []*UTXO {
	if x != nil {
		return x.Inputs
	}
	return nil
}

func (x *BitcoinTransaction) GetOutputs() []*Output {
	if x != nil {
		return x.Outputs
	}
	return nil
}

func (x *BitcoinTransaction) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *BitcoinTransaction) GetLockTime() uint32 {
	if x != nil {
		return x.LockTime
	}
	return 0
}

type EthereumTransaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	To       string  `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	Value    []byte  `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	GasPrice []byte  `protobuf:"bytes,3,opt,name=gas_price,json=gasPrice,proto3,oneof" json:"gas_price,omitempty"`
	GasLimit *uint64 `protobuf:"varint,4,opt,name=gas_limit,json=gasLimit,proto3,oneof" json:"gas_limit,omitempty"`
	Data     []byte  `protobuf:"bytes,5,opt,name=data,proto3,oneof" json:"data,omitempty"`
	Nonce    *uint64 `protobuf:"varint,6,opt,name=nonce,proto3,oneof" json:"nonce,omitempty"`
	ChainId  *uint64 `protobuf:"varint,7,opt,name=chain_id,json=chainId,proto3,oneof" json:"chain_id,omitempty"`
}

func (x *EthereumTransaction) Reset() {
	*x = EthereumTransaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_transaction_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EthereumTransaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EthereumTransaction) ProtoMessage() {}

func (x *EthereumTransaction) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_transaction_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EthereumTransaction.ProtoReflect.Descriptor instead.
func (*EthereumTransaction) Descriptor() ([]byte, []int) {
	return file_transaction_transaction_proto_rawDescGZIP(), []int{2}
}

func (x *EthereumTransaction) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *EthereumTransaction) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *EthereumTransaction) GetGasPrice() []byte {
	if x != nil {
		return x.GasPrice
	}
	return nil
}

func (x *EthereumTransaction) GetGasLimit() uint64 {
	if x != nil && x.GasLimit != nil {
		return *x.GasLimit
	}
	return 0
}

func (x *EthereumTransaction) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *EthereumTransaction) GetNonce() uint64 {
	if x != nil && x.Nonce != nil {
		return *x.Nonce
	}
	return 0
}

func (x *EthereumTransaction) GetChainId() uint64 {
	if x != nil && x.ChainId != nil {
		return *x.ChainId
	}
	return 0
}

type UTXO struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Txid         string `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	Vout         uint32 `protobuf:"varint,2,opt,name=vout,proto3" json:"vout,omitempty"`
	ScriptPubKey string `protobuf:"bytes,3,opt,name=script_pub_key,json=scriptPubKey,proto3" json:"script_pub_key,omitempty"`
	Amount       uint64 `protobuf:"varint,4,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *UTXO) Reset() {
	*x = UTXO{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_transaction_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UTXO) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UTXO) ProtoMessage() {}

func (x *UTXO) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_transaction_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UTXO.ProtoReflect.Descriptor instead.
func (*UTXO) Descriptor() ([]byte, []int) {
	return file_transaction_transaction_proto_rawDescGZIP(), []int{3}
}

func (x *UTXO) GetTxid() string {
	if x != nil {
		return x.Txid
	}
	return ""
}

func (x *UTXO) GetVout() uint32 {
	if x != nil {
		return x.Vout
	}
	return 0
}

func (x *UTXO) GetScriptPubKey() string {
	if x != nil {
		return x.ScriptPubKey
	}
	return ""
}

func (x *UTXO) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

type Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Amount       uint64 `protobuf:"varint,1,opt,name=amount,proto3" json:"amount,omitempty"`
	ScriptPubKey string `protobuf:"bytes,2,opt,name=script_pub_key,json=scriptPubKey,proto3" json:"script_pub_key,omitempty"`
}

func (x *Output) Reset() {
	*x = Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transaction_transaction_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Output) ProtoMessage() {}

func (x *Output) ProtoReflect() protoreflect.Message {
	mi := &file_transaction_transaction_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Output.ProtoReflect.Descriptor instead.
func (*Output) Descriptor() ([]byte, []int) {
	return file_transaction_transaction_proto_rawDescGZIP(), []int{4}
}

func (x *Output) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Output) GetScriptPubKey() string {
	if x != nil {
		return x.ScriptPubKey
	}
	return ""
}

var File_transaction_transaction_proto protoreflect.FileDescriptor

var file_transaction_transaction_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x90, 0x01, 0x0a,
	0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3e, 0x0a, 0x08,
	0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x45, 0x74, 0x68,
	0x65, 0x72, 0x65, 0x75, 0x6d, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x48, 0x00, 0x52, 0x08, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x12, 0x3b, 0x0a, 0x07,
	0x62, 0x69, 0x74, 0x63, 0x6f, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x42, 0x69, 0x74, 0x63,
	0x6f, 0x69, 0x6e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00,
	0x52, 0x07, 0x62, 0x69, 0x74, 0x63, 0x6f, 0x69, 0x6e, 0x42, 0x04, 0x0a, 0x02, 0x74, 0x78, 0x22,
	0xa5, 0x01, 0x0a, 0x12, 0x42, 0x69, 0x74, 0x63, 0x6f, 0x69, 0x6e, 0x54, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x55, 0x54, 0x58, 0x4f, 0x52, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74,
	0x73, 0x12, 0x2d, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x13, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73,
	0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x6c,
	0x6f, 0x63, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x8f, 0x02, 0x0a, 0x13, 0x45, 0x74, 0x68, 0x65,
	0x72, 0x65, 0x75, 0x6d, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a, 0x09, 0x67, 0x61, 0x73, 0x5f, 0x70, 0x72, 0x69,
	0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x08, 0x67, 0x61, 0x73, 0x50,
	0x72, 0x69, 0x63, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20, 0x0a, 0x09, 0x67, 0x61, 0x73, 0x5f, 0x6c,
	0x69, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x48, 0x01, 0x52, 0x08, 0x67, 0x61,
	0x73, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x88, 0x01, 0x01, 0x12, 0x17, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x02, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x88,
	0x01, 0x01, 0x12, 0x19, 0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x04, 0x48, 0x03, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x88, 0x01, 0x01, 0x12, 0x1e, 0x0a,
	0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04, 0x48,
	0x04, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0c, 0x0a,
	0x0a, 0x5f, 0x67, 0x61, 0x73, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x42, 0x0c, 0x0a, 0x0a, 0x5f,
	0x67, 0x61, 0x73, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x64, 0x61,
	0x74, 0x61, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x42, 0x0b, 0x0a, 0x09,
	0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x22, 0x6c, 0x0a, 0x04, 0x55, 0x54, 0x58,
	0x4f, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x74, 0x78, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x6f, 0x75, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x04, 0x76, 0x6f, 0x75, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x5f, 0x70, 0x75, 0x62, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12,
	0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x46, 0x0a, 0x06, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x5f, 0x70, 0x75, 0x62, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x42,
	0x1a, 0x5a, 0x18, 0x74, 0x65, 0x63, 0x64, 0x73, 0x61, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_transaction_transaction_proto_rawDescOnce sync.Once
	file_transaction_transaction_proto_rawDescData = file_transaction_transaction_proto_rawDesc
)

func file_transaction_transaction_proto_rawDescGZIP() []byte {
	file_transaction_transaction_proto_rawDescOnce.Do(func() {
		file_transaction_transaction_proto_rawDescData = protoimpl.X.CompressGZIP(file_transaction_transaction_proto_rawDescData)
	})
	return file_transaction_transaction_proto_rawDescData
}

var file_transaction_transaction_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_transaction_transaction_proto_goTypes = []any{
	(*Transaction)(nil),         // 0: transaction.Transaction
	(*BitcoinTransaction)(nil),  // 1: transaction.BitcoinTransaction
	(*EthereumTransaction)(nil), // 2: transaction.EthereumTransaction
	(*UTXO)(nil),                // 3: transaction.UTXO
	(*Output)(nil),              // 4: transaction.Output
}
var file_transaction_transaction_proto_depIdxs = []int32{
	2, // 0: transaction.Transaction.ethereum:type_name -> transaction.EthereumTransaction
	1, // 1: transaction.Transaction.bitcoin:type_name -> transaction.BitcoinTransaction
	3, // 2: transaction.BitcoinTransaction.inputs:type_name -> transaction.UTXO
	4, // 3: transaction.BitcoinTransaction.outputs:type_name -> transaction.Output
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_transaction_transaction_proto_init() }
func file_transaction_transaction_proto_init() {
	if File_transaction_transaction_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transaction_transaction_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Transaction); i {
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
		file_transaction_transaction_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*BitcoinTransaction); i {
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
		file_transaction_transaction_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*EthereumTransaction); i {
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
		file_transaction_transaction_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*UTXO); i {
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
		file_transaction_transaction_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*Output); i {
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
	file_transaction_transaction_proto_msgTypes[0].OneofWrappers = []any{
		(*Transaction_Ethereum)(nil),
		(*Transaction_Bitcoin)(nil),
	}
	file_transaction_transaction_proto_msgTypes[2].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transaction_transaction_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transaction_transaction_proto_goTypes,
		DependencyIndexes: file_transaction_transaction_proto_depIdxs,
		MessageInfos:      file_transaction_transaction_proto_msgTypes,
	}.Build()
	File_transaction_transaction_proto = out.File
	file_transaction_transaction_proto_rawDesc = nil
	file_transaction_transaction_proto_goTypes = nil
	file_transaction_transaction_proto_depIdxs = nil
}