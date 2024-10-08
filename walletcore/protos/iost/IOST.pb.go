// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.26.1
// source: IOST.proto

package iost

import (
	common "github.com/vultisig/vultisigner/walletcore/protos/common"
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

// The enumeration defines the signature algorithm.
type Algorithm int32

const (
	// unknown
	Algorithm_UNKNOWN Algorithm = 0
	// secp256k1
	Algorithm_SECP256K1 Algorithm = 1
	// ed25519
	Algorithm_ED25519 Algorithm = 2
)

// Enum value maps for Algorithm.
var (
	Algorithm_name = map[int32]string{
		0: "UNKNOWN",
		1: "SECP256K1",
		2: "ED25519",
	}
	Algorithm_value = map[string]int32{
		"UNKNOWN":   0,
		"SECP256K1": 1,
		"ED25519":   2,
	}
)

func (x Algorithm) Enum() *Algorithm {
	p := new(Algorithm)
	*p = x
	return p
}

func (x Algorithm) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Algorithm) Descriptor() protoreflect.EnumDescriptor {
	return file_IOST_proto_enumTypes[0].Descriptor()
}

func (Algorithm) Type() protoreflect.EnumType {
	return &file_IOST_proto_enumTypes[0]
}

func (x Algorithm) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Algorithm.Descriptor instead.
func (Algorithm) EnumDescriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{0}
}

// The message defines transaction action struct.
type Action struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// contract name
	Contract string `protobuf:"bytes,1,opt,name=contract,proto3" json:"contract,omitempty"`
	// action name
	ActionName string `protobuf:"bytes,2,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
	// data
	Data string `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Action) Reset() {
	*x = Action{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Action) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Action) ProtoMessage() {}

func (x *Action) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Action.ProtoReflect.Descriptor instead.
func (*Action) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{0}
}

func (x *Action) GetContract() string {
	if x != nil {
		return x.Contract
	}
	return ""
}

func (x *Action) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

func (x *Action) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

// The message defines transaction amount limit struct.
type AmountLimit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// token name
	Token string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	// limit value
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *AmountLimit) Reset() {
	*x = AmountLimit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AmountLimit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AmountLimit) ProtoMessage() {}

func (x *AmountLimit) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AmountLimit.ProtoReflect.Descriptor instead.
func (*AmountLimit) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{1}
}

func (x *AmountLimit) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *AmountLimit) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// The message defines signature struct.
type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// signature algorithm
	Algorithm Algorithm `protobuf:"varint,1,opt,name=algorithm,proto3,enum=TW.IOST.Proto.Algorithm" json:"algorithm,omitempty"`
	// signature bytes
	Signature []byte `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	// public key
	PublicKey []byte `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{2}
}

func (x *Signature) GetAlgorithm() Algorithm {
	if x != nil {
		return x.Algorithm
	}
	return Algorithm_UNKNOWN
}

func (x *Signature) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Signature) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

// The message defines the transaction request.
type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// transaction timestamp
	Time int64 `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	// expiration timestamp
	Expiration int64 `protobuf:"varint,2,opt,name=expiration,proto3" json:"expiration,omitempty"`
	// gas price
	GasRatio float64 `protobuf:"fixed64,3,opt,name=gas_ratio,json=gasRatio,proto3" json:"gas_ratio,omitempty"`
	// gas limit
	GasLimit float64 `protobuf:"fixed64,4,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	// delay nanoseconds
	Delay int64 `protobuf:"varint,5,opt,name=delay,proto3" json:"delay,omitempty"`
	// chain id
	ChainId uint32 `protobuf:"varint,6,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	// action list
	Actions []*Action `protobuf:"bytes,7,rep,name=actions,proto3" json:"actions,omitempty"`
	// amount limit
	AmountLimit []*AmountLimit `protobuf:"bytes,8,rep,name=amount_limit,json=amountLimit,proto3" json:"amount_limit,omitempty"`
	// signer list
	Signers []string `protobuf:"bytes,9,rep,name=signers,proto3" json:"signers,omitempty"`
	// signatures of signers
	Signatures []*Signature `protobuf:"bytes,10,rep,name=signatures,proto3" json:"signatures,omitempty"`
	// publisher
	Publisher string `protobuf:"bytes,11,opt,name=publisher,proto3" json:"publisher,omitempty"`
	// signatures of publisher
	PublisherSigs []*Signature `protobuf:"bytes,12,rep,name=publisher_sigs,json=publisherSigs,proto3" json:"publisher_sigs,omitempty"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[3]
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
	return file_IOST_proto_rawDescGZIP(), []int{3}
}

func (x *Transaction) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *Transaction) GetExpiration() int64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

func (x *Transaction) GetGasRatio() float64 {
	if x != nil {
		return x.GasRatio
	}
	return 0
}

func (x *Transaction) GetGasLimit() float64 {
	if x != nil {
		return x.GasLimit
	}
	return 0
}

func (x *Transaction) GetDelay() int64 {
	if x != nil {
		return x.Delay
	}
	return 0
}

func (x *Transaction) GetChainId() uint32 {
	if x != nil {
		return x.ChainId
	}
	return 0
}

func (x *Transaction) GetActions() []*Action {
	if x != nil {
		return x.Actions
	}
	return nil
}

func (x *Transaction) GetAmountLimit() []*AmountLimit {
	if x != nil {
		return x.AmountLimit
	}
	return nil
}

func (x *Transaction) GetSigners() []string {
	if x != nil {
		return x.Signers
	}
	return nil
}

func (x *Transaction) GetSignatures() []*Signature {
	if x != nil {
		return x.Signatures
	}
	return nil
}

func (x *Transaction) GetPublisher() string {
	if x != nil {
		return x.Publisher
	}
	return ""
}

func (x *Transaction) GetPublisherSigs() []*Signature {
	if x != nil {
		return x.PublisherSigs
	}
	return nil
}

type AccountInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	ActiveKey []byte `protobuf:"bytes,2,opt,name=active_key,json=activeKey,proto3" json:"active_key,omitempty"`
	OwnerKey  []byte `protobuf:"bytes,3,opt,name=owner_key,json=ownerKey,proto3" json:"owner_key,omitempty"`
}

func (x *AccountInfo) Reset() {
	*x = AccountInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountInfo) ProtoMessage() {}

func (x *AccountInfo) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountInfo.ProtoReflect.Descriptor instead.
func (*AccountInfo) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{4}
}

func (x *AccountInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AccountInfo) GetActiveKey() []byte {
	if x != nil {
		return x.ActiveKey
	}
	return nil
}

func (x *AccountInfo) GetOwnerKey() []byte {
	if x != nil {
		return x.OwnerKey
	}
	return nil
}

// Input data necessary to create a signed transaction.
type SigningInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account             *AccountInfo `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	TransactionTemplate *Transaction `protobuf:"bytes,2,opt,name=transaction_template,json=transactionTemplate,proto3" json:"transaction_template,omitempty"`
	TransferDestination string       `protobuf:"bytes,3,opt,name=transfer_destination,json=transferDestination,proto3" json:"transfer_destination,omitempty"`
	TransferAmount      string       `protobuf:"bytes,4,opt,name=transfer_amount,json=transferAmount,proto3" json:"transfer_amount,omitempty"`
	TransferMemo        string       `protobuf:"bytes,5,opt,name=transfer_memo,json=transferMemo,proto3" json:"transfer_memo,omitempty"`
}

func (x *SigningInput) Reset() {
	*x = SigningInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigningInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigningInput) ProtoMessage() {}

func (x *SigningInput) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigningInput.ProtoReflect.Descriptor instead.
func (*SigningInput) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{5}
}

func (x *SigningInput) GetAccount() *AccountInfo {
	if x != nil {
		return x.Account
	}
	return nil
}

func (x *SigningInput) GetTransactionTemplate() *Transaction {
	if x != nil {
		return x.TransactionTemplate
	}
	return nil
}

func (x *SigningInput) GetTransferDestination() string {
	if x != nil {
		return x.TransferDestination
	}
	return ""
}

func (x *SigningInput) GetTransferAmount() string {
	if x != nil {
		return x.TransferAmount
	}
	return ""
}

func (x *SigningInput) GetTransferMemo() string {
	if x != nil {
		return x.TransferMemo
	}
	return ""
}

// Transaction signing output.
type SigningOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Signed transaction
	Transaction *Transaction `protobuf:"bytes,1,opt,name=transaction,proto3" json:"transaction,omitempty"`
	// Signed and encoded transaction bytes.
	Encoded []byte `protobuf:"bytes,2,opt,name=encoded,proto3" json:"encoded,omitempty"`
	// error code, 0 is ok, other codes will be treated as errors
	Error common.SigningError `protobuf:"varint,3,opt,name=error,proto3,enum=TW.Common.Proto.SigningError" json:"error,omitempty"`
	// error code description
	ErrorMessage string `protobuf:"bytes,4,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *SigningOutput) Reset() {
	*x = SigningOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_IOST_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigningOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigningOutput) ProtoMessage() {}

func (x *SigningOutput) ProtoReflect() protoreflect.Message {
	mi := &file_IOST_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigningOutput.ProtoReflect.Descriptor instead.
func (*SigningOutput) Descriptor() ([]byte, []int) {
	return file_IOST_proto_rawDescGZIP(), []int{6}
}

func (x *SigningOutput) GetTransaction() *Transaction {
	if x != nil {
		return x.Transaction
	}
	return nil
}

func (x *SigningOutput) GetEncoded() []byte {
	if x != nil {
		return x.Encoded
	}
	return nil
}

func (x *SigningOutput) GetError() common.SigningError {
	if x != nil {
		return x.Error
	}
	return common.SigningError(0)
}

func (x *SigningOutput) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

var File_IOST_proto protoreflect.FileDescriptor

var file_IOST_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x54, 0x57,
	0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x43, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x59, 0x0a, 0x06, 0x41, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x12,
	0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x39, 0x0a, 0x0b, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x4c, 0x69,
	0x6d, 0x69, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0x80, 0x01, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x36, 0x0a,
	0x09, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x18, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x52, 0x09, 0x61, 0x6c, 0x67, 0x6f,
	0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65,
	0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x22, 0xcf, 0x03, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x69,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x61, 0x73, 0x5f, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x67, 0x61, 0x73, 0x52, 0x61,
	0x74, 0x69, 0x6f, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x61, 0x73, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x67, 0x61, 0x73, 0x4c, 0x69, 0x6d, 0x69, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x05, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x49,
	0x64, 0x12, 0x2f, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x07, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x3d, 0x0a, 0x0c, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f,
	0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x52, 0x0b, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x4c, 0x69, 0x6d, 0x69,
	0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x09, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x07, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x38, 0x0a, 0x0a, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68,
	0x65, 0x72, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73,
	0x68, 0x65, 0x72, 0x12, 0x3f, 0x0a, 0x0e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72,
	0x5f, 0x73, 0x69, 0x67, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x54, 0x57,
	0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x0d, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x72,
	0x53, 0x69, 0x67, 0x73, 0x22, 0x5d, 0x0a, 0x0b, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x4b, 0x65, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f,
	0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6f, 0x77, 0x6e, 0x65, 0x72,
	0x4b, 0x65, 0x79, 0x22, 0x94, 0x02, 0x0a, 0x0c, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x12, 0x34, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x4d, 0x0a, 0x14, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f,
	0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x13, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x31, 0x0a, 0x14, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x65, 0x72, 0x5f, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x5f, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x41,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x5f, 0x6d, 0x65, 0x6d, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x6f, 0x22, 0xc1, 0x01, 0x0a, 0x0d, 0x53,
	0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x3c, 0x0a, 0x0b,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x4f, 0x53, 0x54, 0x2e, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e,
	0x63, 0x6f, 0x64, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x65, 0x6e, 0x63,
	0x6f, 0x64, 0x65, 0x64, 0x12, 0x33, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x54, 0x57, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a, 0x34,
	0x0a, 0x09, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x0b, 0x0a, 0x07, 0x55,
	0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x45, 0x43, 0x50,
	0x32, 0x35, 0x36, 0x4b, 0x31, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x45, 0x44, 0x32, 0x35, 0x35,
	0x31, 0x39, 0x10, 0x02, 0x42, 0x17, 0x0a, 0x15, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x2e, 0x6a, 0x6e, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_IOST_proto_rawDescOnce sync.Once
	file_IOST_proto_rawDescData = file_IOST_proto_rawDesc
)

func file_IOST_proto_rawDescGZIP() []byte {
	file_IOST_proto_rawDescOnce.Do(func() {
		file_IOST_proto_rawDescData = protoimpl.X.CompressGZIP(file_IOST_proto_rawDescData)
	})
	return file_IOST_proto_rawDescData
}

var file_IOST_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_IOST_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_IOST_proto_goTypes = []interface{}{
	(Algorithm)(0),           // 0: TW.IOST.Proto.Algorithm
	(*Action)(nil),           // 1: TW.IOST.Proto.Action
	(*AmountLimit)(nil),      // 2: TW.IOST.Proto.AmountLimit
	(*Signature)(nil),        // 3: TW.IOST.Proto.Signature
	(*Transaction)(nil),      // 4: TW.IOST.Proto.Transaction
	(*AccountInfo)(nil),      // 5: TW.IOST.Proto.AccountInfo
	(*SigningInput)(nil),     // 6: TW.IOST.Proto.SigningInput
	(*SigningOutput)(nil),    // 7: TW.IOST.Proto.SigningOutput
	(common.SigningError)(0), // 8: TW.Common.Proto.SigningError
}
var file_IOST_proto_depIdxs = []int32{
	0, // 0: TW.IOST.Proto.Signature.algorithm:type_name -> TW.IOST.Proto.Algorithm
	1, // 1: TW.IOST.Proto.Transaction.actions:type_name -> TW.IOST.Proto.Action
	2, // 2: TW.IOST.Proto.Transaction.amount_limit:type_name -> TW.IOST.Proto.AmountLimit
	3, // 3: TW.IOST.Proto.Transaction.signatures:type_name -> TW.IOST.Proto.Signature
	3, // 4: TW.IOST.Proto.Transaction.publisher_sigs:type_name -> TW.IOST.Proto.Signature
	5, // 5: TW.IOST.Proto.SigningInput.account:type_name -> TW.IOST.Proto.AccountInfo
	4, // 6: TW.IOST.Proto.SigningInput.transaction_template:type_name -> TW.IOST.Proto.Transaction
	4, // 7: TW.IOST.Proto.SigningOutput.transaction:type_name -> TW.IOST.Proto.Transaction
	8, // 8: TW.IOST.Proto.SigningOutput.error:type_name -> TW.Common.Proto.SigningError
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_IOST_proto_init() }
func file_IOST_proto_init() {
	if File_IOST_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_IOST_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Action); i {
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
		file_IOST_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AmountLimit); i {
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
		file_IOST_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
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
		file_IOST_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_IOST_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountInfo); i {
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
		file_IOST_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigningInput); i {
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
		file_IOST_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigningOutput); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_IOST_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_IOST_proto_goTypes,
		DependencyIndexes: file_IOST_proto_depIdxs,
		EnumInfos:         file_IOST_proto_enumTypes,
		MessageInfos:      file_IOST_proto_msgTypes,
	}.Build()
	File_IOST_proto = out.File
	file_IOST_proto_rawDesc = nil
	file_IOST_proto_goTypes = nil
	file_IOST_proto_depIdxs = nil
}
