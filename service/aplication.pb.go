// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/aplication.proto

package service

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Ficheros Mensajes
type AddFileRequest struct {
	FileName             string   `protobuf:"bytes,1,opt,name=fileName,proto3" json:"fileName,omitempty"`
	FileExtension        string   `protobuf:"bytes,2,opt,name=fileExtension,proto3" json:"fileExtension,omitempty"`
	InfoToStorage        []byte   `protobuf:"bytes,3,opt,name=infoToStorage,proto3" json:"infoToStorage,omitempty"`
	BytesSize            int32    `protobuf:"varint,4,opt,name=bytesSize,proto3" json:"bytesSize,omitempty"`
	Tags                 []string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddFileRequest) Reset()         { *m = AddFileRequest{} }
func (m *AddFileRequest) String() string { return proto.CompactTextString(m) }
func (*AddFileRequest) ProtoMessage()    {}
func (*AddFileRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{0}
}

func (m *AddFileRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddFileRequest.Unmarshal(m, b)
}
func (m *AddFileRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddFileRequest.Marshal(b, m, deterministic)
}
func (m *AddFileRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddFileRequest.Merge(m, src)
}
func (m *AddFileRequest) XXX_Size() int {
	return xxx_messageInfo_AddFileRequest.Size(m)
}
func (m *AddFileRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AddFileRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AddFileRequest proto.InternalMessageInfo

func (m *AddFileRequest) GetFileName() string {
	if m != nil {
		return m.FileName
	}
	return ""
}

func (m *AddFileRequest) GetFileExtension() string {
	if m != nil {
		return m.FileExtension
	}
	return ""
}

func (m *AddFileRequest) GetInfoToStorage() []byte {
	if m != nil {
		return m.InfoToStorage
	}
	return nil
}

func (m *AddFileRequest) GetBytesSize() int32 {
	if m != nil {
		return m.BytesSize
	}
	return 0
}

func (m *AddFileRequest) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

type AddFileResponse struct {
	Response             string   `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddFileResponse) Reset()         { *m = AddFileResponse{} }
func (m *AddFileResponse) String() string { return proto.CompactTextString(m) }
func (*AddFileResponse) ProtoMessage()    {}
func (*AddFileResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{1}
}

func (m *AddFileResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddFileResponse.Unmarshal(m, b)
}
func (m *AddFileResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddFileResponse.Marshal(b, m, deterministic)
}
func (m *AddFileResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddFileResponse.Merge(m, src)
}
func (m *AddFileResponse) XXX_Size() int {
	return xxx_messageInfo_AddFileResponse.Size(m)
}
func (m *AddFileResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AddFileResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AddFileResponse proto.InternalMessageInfo

func (m *AddFileResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

type DeleteFileRequest struct {
	DeleteTags           []string `protobuf:"bytes,1,rep,name=deleteTags,proto3" json:"deleteTags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteFileRequest) Reset()         { *m = DeleteFileRequest{} }
func (m *DeleteFileRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteFileRequest) ProtoMessage()    {}
func (*DeleteFileRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{2}
}

func (m *DeleteFileRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteFileRequest.Unmarshal(m, b)
}
func (m *DeleteFileRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteFileRequest.Marshal(b, m, deterministic)
}
func (m *DeleteFileRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteFileRequest.Merge(m, src)
}
func (m *DeleteFileRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteFileRequest.Size(m)
}
func (m *DeleteFileRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteFileRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteFileRequest proto.InternalMessageInfo

func (m *DeleteFileRequest) GetDeleteTags() []string {
	if m != nil {
		return m.DeleteTags
	}
	return nil
}

type DeleteFileResponse struct {
	Response             string   `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteFileResponse) Reset()         { *m = DeleteFileResponse{} }
func (m *DeleteFileResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteFileResponse) ProtoMessage()    {}
func (*DeleteFileResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{3}
}

func (m *DeleteFileResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteFileResponse.Unmarshal(m, b)
}
func (m *DeleteFileResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteFileResponse.Marshal(b, m, deterministic)
}
func (m *DeleteFileResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteFileResponse.Merge(m, src)
}
func (m *DeleteFileResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteFileResponse.Size(m)
}
func (m *DeleteFileResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteFileResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteFileResponse proto.InternalMessageInfo

func (m *DeleteFileResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

type ListFileRequest struct {
	QueryTags            []string `protobuf:"bytes,1,rep,name=queryTags,proto3" json:"queryTags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListFileRequest) Reset()         { *m = ListFileRequest{} }
func (m *ListFileRequest) String() string { return proto.CompactTextString(m) }
func (*ListFileRequest) ProtoMessage()    {}
func (*ListFileRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{4}
}

func (m *ListFileRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListFileRequest.Unmarshal(m, b)
}
func (m *ListFileRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListFileRequest.Marshal(b, m, deterministic)
}
func (m *ListFileRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListFileRequest.Merge(m, src)
}
func (m *ListFileRequest) XXX_Size() int {
	return xxx_messageInfo_ListFileRequest.Size(m)
}
func (m *ListFileRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListFileRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListFileRequest proto.InternalMessageInfo

func (m *ListFileRequest) GetQueryTags() []string {
	if m != nil {
		return m.QueryTags
	}
	return nil
}

type ListFileResponse struct {
	Response             map[string][]byte `protobuf:"bytes,1,rep,name=Response,proto3" json:"Response,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ListFileResponse) Reset()         { *m = ListFileResponse{} }
func (m *ListFileResponse) String() string { return proto.CompactTextString(m) }
func (*ListFileResponse) ProtoMessage()    {}
func (*ListFileResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{5}
}

func (m *ListFileResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListFileResponse.Unmarshal(m, b)
}
func (m *ListFileResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListFileResponse.Marshal(b, m, deterministic)
}
func (m *ListFileResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListFileResponse.Merge(m, src)
}
func (m *ListFileResponse) XXX_Size() int {
	return xxx_messageInfo_ListFileResponse.Size(m)
}
func (m *ListFileResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListFileResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListFileResponse proto.InternalMessageInfo

func (m *ListFileResponse) GetResponse() map[string][]byte {
	if m != nil {
		return m.Response
	}
	return nil
}

type AddTagsRequest struct {
	QueryTags            []string `protobuf:"bytes,1,rep,name=QueryTags,proto3" json:"QueryTags,omitempty"`
	AddTags              []string `protobuf:"bytes,2,rep,name=AddTags,proto3" json:"AddTags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddTagsRequest) Reset()         { *m = AddTagsRequest{} }
func (m *AddTagsRequest) String() string { return proto.CompactTextString(m) }
func (*AddTagsRequest) ProtoMessage()    {}
func (*AddTagsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{6}
}

func (m *AddTagsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddTagsRequest.Unmarshal(m, b)
}
func (m *AddTagsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddTagsRequest.Marshal(b, m, deterministic)
}
func (m *AddTagsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddTagsRequest.Merge(m, src)
}
func (m *AddTagsRequest) XXX_Size() int {
	return xxx_messageInfo_AddTagsRequest.Size(m)
}
func (m *AddTagsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AddTagsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AddTagsRequest proto.InternalMessageInfo

func (m *AddTagsRequest) GetQueryTags() []string {
	if m != nil {
		return m.QueryTags
	}
	return nil
}

func (m *AddTagsRequest) GetAddTags() []string {
	if m != nil {
		return m.AddTags
	}
	return nil
}

type AddTagsResponse struct {
	Response             string   `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddTagsResponse) Reset()         { *m = AddTagsResponse{} }
func (m *AddTagsResponse) String() string { return proto.CompactTextString(m) }
func (*AddTagsResponse) ProtoMessage()    {}
func (*AddTagsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{7}
}

func (m *AddTagsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddTagsResponse.Unmarshal(m, b)
}
func (m *AddTagsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddTagsResponse.Marshal(b, m, deterministic)
}
func (m *AddTagsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddTagsResponse.Merge(m, src)
}
func (m *AddTagsResponse) XXX_Size() int {
	return xxx_messageInfo_AddTagsResponse.Size(m)
}
func (m *AddTagsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AddTagsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AddTagsResponse proto.InternalMessageInfo

func (m *AddTagsResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

type DeleteTagsRequest struct {
	QueryTags            []string `protobuf:"bytes,1,rep,name=QueryTags,proto3" json:"QueryTags,omitempty"`
	DeleteTags           []string `protobuf:"bytes,2,rep,name=DeleteTags,proto3" json:"DeleteTags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteTagsRequest) Reset()         { *m = DeleteTagsRequest{} }
func (m *DeleteTagsRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteTagsRequest) ProtoMessage()    {}
func (*DeleteTagsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{8}
}

func (m *DeleteTagsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteTagsRequest.Unmarshal(m, b)
}
func (m *DeleteTagsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteTagsRequest.Marshal(b, m, deterministic)
}
func (m *DeleteTagsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteTagsRequest.Merge(m, src)
}
func (m *DeleteTagsRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteTagsRequest.Size(m)
}
func (m *DeleteTagsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteTagsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteTagsRequest proto.InternalMessageInfo

func (m *DeleteTagsRequest) GetQueryTags() []string {
	if m != nil {
		return m.QueryTags
	}
	return nil
}

func (m *DeleteTagsRequest) GetDeleteTags() []string {
	if m != nil {
		return m.DeleteTags
	}
	return nil
}

type DeleteTagsResponse struct {
	Response             string   `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteTagsResponse) Reset()         { *m = DeleteTagsResponse{} }
func (m *DeleteTagsResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteTagsResponse) ProtoMessage()    {}
func (*DeleteTagsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e58dece1a8e61e2, []int{9}
}

func (m *DeleteTagsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteTagsResponse.Unmarshal(m, b)
}
func (m *DeleteTagsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteTagsResponse.Marshal(b, m, deterministic)
}
func (m *DeleteTagsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteTagsResponse.Merge(m, src)
}
func (m *DeleteTagsResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteTagsResponse.Size(m)
}
func (m *DeleteTagsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteTagsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteTagsResponse proto.InternalMessageInfo

func (m *DeleteTagsResponse) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

func init() {
	proto.RegisterType((*AddFileRequest)(nil), "service.AddFileRequest")
	proto.RegisterType((*AddFileResponse)(nil), "service.AddFileResponse")
	proto.RegisterType((*DeleteFileRequest)(nil), "service.DeleteFileRequest")
	proto.RegisterType((*DeleteFileResponse)(nil), "service.DeleteFileResponse")
	proto.RegisterType((*ListFileRequest)(nil), "service.ListFileRequest")
	proto.RegisterType((*ListFileResponse)(nil), "service.ListFileResponse")
	proto.RegisterMapType((map[string][]byte)(nil), "service.ListFileResponse.ResponseEntry")
	proto.RegisterType((*AddTagsRequest)(nil), "service.AddTagsRequest")
	proto.RegisterType((*AddTagsResponse)(nil), "service.AddTagsResponse")
	proto.RegisterType((*DeleteTagsRequest)(nil), "service.DeleteTagsRequest")
	proto.RegisterType((*DeleteTagsResponse)(nil), "service.DeleteTagsResponse")
}

func init() {
	proto.RegisterFile("proto/aplication.proto", fileDescriptor_7e58dece1a8e61e2)
}

var fileDescriptor_7e58dece1a8e61e2 = []byte{
	// 465 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xc1, 0x6e, 0xd4, 0x30,
	0x10, 0x95, 0x77, 0xbb, 0x74, 0x33, 0x6d, 0x69, 0xb1, 0x10, 0x98, 0xb4, 0xaa, 0xa2, 0x08, 0x89,
	0x5c, 0xc8, 0xa2, 0xf6, 0x82, 0x00, 0x09, 0x15, 0x08, 0xe2, 0x80, 0x90, 0xea, 0xf6, 0xc4, 0x2d,
	0xed, 0x4e, 0x2b, 0x8b, 0x10, 0x6f, 0x63, 0x6f, 0x45, 0xf8, 0x0e, 0xbe, 0x82, 0x1f, 0xe2, 0x77,
	0x50, 0x1c, 0x27, 0x71, 0xba, 0x5b, 0x29, 0xb7, 0xf1, 0x9b, 0xf1, 0x7b, 0xf3, 0x66, 0x9c, 0xc0,
	0x93, 0x45, 0x21, 0xb5, 0x9c, 0xa5, 0x8b, 0x4c, 0x5c, 0xa6, 0x5a, 0xc8, 0x3c, 0x36, 0x00, 0xdd,
	0x54, 0x58, 0xdc, 0x8a, 0x4b, 0x0c, 0xff, 0x12, 0x78, 0x78, 0x32, 0x9f, 0x7f, 0x16, 0x19, 0x72,
	0xbc, 0x59, 0xa2, 0xd2, 0xd4, 0x87, 0xe9, 0x95, 0xc8, 0xf0, 0x5b, 0xfa, 0x13, 0x19, 0x09, 0x48,
	0xe4, 0xf1, 0xf6, 0x4c, 0x9f, 0xc3, 0x4e, 0x15, 0x27, 0xbf, 0x34, 0xe6, 0x4a, 0xc8, 0x9c, 0x8d,
	0x4c, 0x41, 0x1f, 0xac, 0xaa, 0x44, 0x7e, 0x25, 0xcf, 0xe5, 0x99, 0x96, 0x45, 0x7a, 0x8d, 0x6c,
	0x1c, 0x90, 0x68, 0x9b, 0xf7, 0x41, 0x7a, 0x00, 0xde, 0x45, 0xa9, 0x51, 0x9d, 0x89, 0xdf, 0xc8,
	0x36, 0x02, 0x12, 0x4d, 0x78, 0x07, 0x50, 0x0a, 0x1b, 0x3a, 0xbd, 0x56, 0x6c, 0x12, 0x8c, 0x23,
	0x8f, 0x9b, 0x38, 0x7c, 0x09, 0xbb, 0x6d, 0xaf, 0x6a, 0x21, 0x73, 0x85, 0x55, 0xb3, 0x85, 0x8d,
	0x9b, 0x66, 0x9b, 0x73, 0x78, 0x0c, 0x8f, 0x3e, 0x61, 0x86, 0x1a, 0x5d, 0x77, 0x87, 0x00, 0x73,
	0x03, 0x9e, 0x57, 0xec, 0xc4, 0xb0, 0x3b, 0x48, 0xf8, 0x0a, 0xa8, 0x7b, 0x69, 0x80, 0xcc, 0x0c,
	0x76, 0xbf, 0x0a, 0xa5, 0x5d, 0x91, 0x03, 0xf0, 0x6e, 0x96, 0x58, 0x94, 0x8e, 0x46, 0x07, 0x84,
	0x7f, 0x08, 0xec, 0x75, 0x37, 0xac, 0xc2, 0x47, 0x98, 0xf2, 0x4e, 0x61, 0x1c, 0x6d, 0x1d, 0xbd,
	0x88, 0xed, 0x92, 0xe2, 0xbb, 0xc5, 0x71, 0x13, 0x24, 0xb9, 0x2e, 0x4a, 0xde, 0x5e, 0xf4, 0xdf,
	0xc2, 0x4e, 0x2f, 0x45, 0xf7, 0x60, 0xfc, 0x03, 0x4b, 0xdb, 0x72, 0x15, 0xd2, 0xc7, 0x30, 0xb9,
	0x4d, 0xb3, 0x25, 0x9a, 0xcd, 0x6d, 0xf3, 0xfa, 0xf0, 0x66, 0xf4, 0x9a, 0x84, 0x5f, 0xcc, 0x4b,
	0xa8, 0x3a, 0x74, 0x6c, 0x9c, 0xde, 0xb5, 0xd1, 0x02, 0x94, 0xc1, 0xa6, 0xad, 0x67, 0x23, 0x93,
	0x6b, 0x8e, 0x76, 0x4f, 0x35, 0xd3, 0x80, 0x01, 0x9e, 0x36, 0x7b, 0x1a, 0xae, 0x7d, 0x08, 0xd0,
	0x5d, 0xb1, 0xf2, 0x0e, 0xd2, 0x6d, 0x71, 0x68, 0x13, 0x47, 0xff, 0x46, 0x00, 0x27, 0xed, 0x67,
	0x42, 0xdf, 0x19, 0x73, 0xd5, 0xd0, 0xe9, 0xd3, 0x76, 0x0f, 0xfd, 0x0f, 0xc5, 0x67, 0xab, 0x09,
	0x2b, 0x94, 0x34, 0xed, 0x19, 0x02, 0xbf, 0xad, 0x5b, 0x79, 0x8e, 0xfe, 0xfe, 0xda, 0x9c, 0xa5,
	0x79, 0x0f, 0xd3, 0x66, 0xf5, 0x94, 0xad, 0x79, 0x0d, 0x35, 0xc5, 0xb3, 0x7b, 0xdf, 0x89, 0x75,
	0x61, 0x26, 0xd6, 0x73, 0xe1, 0x0c, 0xba, 0xef, 0xa2, 0x37, 0xae, 0xc4, 0x1d, 0xf2, 0x8a, 0x0b,
	0x97, 0x63, 0x7f, 0x6d, 0xae, 0xa6, 0xf9, 0xb0, 0xf5, 0xdd, 0x8b, 0x67, 0x36, 0x7f, 0xf1, 0xc0,
	0xfc, 0x7f, 0x8e, 0xff, 0x07, 0x00, 0x00, 0xff, 0xff, 0x6e, 0xed, 0xd4, 0xf3, 0x99, 0x04, 0x00,
	0x00,
}
