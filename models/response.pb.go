// Code generated by protoc-gen-go. DO NOT EDIT.
// source: response.proto

package models

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type IndexResponse struct {
	// lensIdentifier is the identifier of the indexed object according to the lens system
	LensIdentifier string `protobuf:"bytes,1,opt,name=lensIdentifier,proto3" json:"lensIdentifier,omitempty"`
	// data is miscellaneous data associated with the response
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IndexResponse) Reset()         { *m = IndexResponse{} }
func (m *IndexResponse) String() string { return proto.CompactTextString(m) }
func (*IndexResponse) ProtoMessage()    {}
func (*IndexResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0fbc901015fa5021, []int{0}
}

func (m *IndexResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IndexResponse.Unmarshal(m, b)
}
func (m *IndexResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IndexResponse.Marshal(b, m, deterministic)
}
func (m *IndexResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexResponse.Merge(m, src)
}
func (m *IndexResponse) XXX_Size() int {
	return xxx_messageInfo_IndexResponse.Size(m)
}
func (m *IndexResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexResponse.DiscardUnknown(m)
}

var xxx_messageInfo_IndexResponse proto.InternalMessageInfo

func (m *IndexResponse) GetLensIdentifier() string {
	if m != nil {
		return m.LensIdentifier
	}
	return ""
}

func (m *IndexResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type SearchResponse struct {
	// name is the "name" of the object, such as an IPFS content hash
	Names []string `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
	// objectType is the type of the object, such as IPLD
	ObjectType           string   `protobuf:"bytes,2,opt,name=objectType,proto3" json:"objectType,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SearchResponse) Reset()         { *m = SearchResponse{} }
func (m *SearchResponse) String() string { return proto.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()    {}
func (*SearchResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0fbc901015fa5021, []int{1}
}

func (m *SearchResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchResponse.Unmarshal(m, b)
}
func (m *SearchResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchResponse.Marshal(b, m, deterministic)
}
func (m *SearchResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchResponse.Merge(m, src)
}
func (m *SearchResponse) XXX_Size() int {
	return xxx_messageInfo_SearchResponse.Size(m)
}
func (m *SearchResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SearchResponse proto.InternalMessageInfo

func (m *SearchResponse) GetNames() []string {
	if m != nil {
		return m.Names
	}
	return nil
}

func (m *SearchResponse) GetObjectType() string {
	if m != nil {
		return m.ObjectType
	}
	return ""
}

func init() {
	proto.RegisterType((*IndexResponse)(nil), "response.IndexResponse")
	proto.RegisterType((*SearchResponse)(nil), "response.SearchResponse")
}

func init() { proto.RegisterFile("response.proto", fileDescriptor_0fbc901015fa5021) }

var fileDescriptor_0fbc901015fa5021 = []byte{
	// 154 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2b, 0x4a, 0x2d, 0x2e,
	0xc8, 0xcf, 0x2b, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0x95, 0xbc,
	0xb9, 0x78, 0x3d, 0xf3, 0x52, 0x52, 0x2b, 0x82, 0xa0, 0x02, 0x42, 0x6a, 0x5c, 0x7c, 0x39, 0xa9,
	0x79, 0xc5, 0x9e, 0x29, 0xa9, 0x79, 0x25, 0x99, 0x69, 0x99, 0xa9, 0x45, 0x12, 0x8c, 0x0a, 0x8c,
	0x1a, 0x9c, 0x41, 0x68, 0xa2, 0x42, 0x42, 0x5c, 0x2c, 0x29, 0x89, 0x25, 0x89, 0x12, 0x4c, 0x0a,
	0x8c, 0x1a, 0x3c, 0x41, 0x60, 0xb6, 0x92, 0x1b, 0x17, 0x5f, 0x70, 0x6a, 0x62, 0x51, 0x72, 0x06,
	0xdc, 0x34, 0x11, 0x2e, 0xd6, 0xbc, 0xc4, 0xdc, 0xd4, 0x62, 0x09, 0x46, 0x05, 0x66, 0x0d, 0xce,
	0x20, 0x08, 0x47, 0x48, 0x8e, 0x8b, 0x2b, 0x3f, 0x29, 0x2b, 0x35, 0xb9, 0x24, 0xa4, 0xb2, 0x20,
	0x15, 0x6c, 0x02, 0x67, 0x10, 0x92, 0x48, 0x12, 0x1b, 0xd8, 0x95, 0xc6, 0x80, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xcd, 0xbb, 0xe8, 0xd9, 0xb7, 0x00, 0x00, 0x00,
}
