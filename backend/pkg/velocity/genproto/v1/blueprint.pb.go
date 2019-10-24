// Code generated by protoc-gen-go. DO NOT EDIT.
// source: blueprint.proto

package v1

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

type Blueprint struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Blueprint) Reset()         { *m = Blueprint{} }
func (m *Blueprint) String() string { return proto.CompactTextString(m) }
func (*Blueprint) ProtoMessage()    {}
func (*Blueprint) Descriptor() ([]byte, []int) {
	return fileDescriptor_d334b799e628a382, []int{0}
}

func (m *Blueprint) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Blueprint.Unmarshal(m, b)
}
func (m *Blueprint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Blueprint.Marshal(b, m, deterministic)
}
func (m *Blueprint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Blueprint.Merge(m, src)
}
func (m *Blueprint) XXX_Size() int {
	return xxx_messageInfo_Blueprint.Size(m)
}
func (m *Blueprint) XXX_DiscardUnknown() {
	xxx_messageInfo_Blueprint.DiscardUnknown(m)
}

var xxx_messageInfo_Blueprint proto.InternalMessageInfo

func (m *Blueprint) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Blueprint) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func init() {
	proto.RegisterType((*Blueprint)(nil), "velocity.v1.Blueprint")
}

func init() { proto.RegisterFile("blueprint.proto", fileDescriptor_d334b799e628a382) }

var fileDescriptor_d334b799e628a382 = []byte{
	// 110 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0xca, 0x29, 0x4d,
	0x2d, 0x28, 0xca, 0xcc, 0x2b, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2e, 0x4b, 0xcd,
	0xc9, 0x4f, 0xce, 0x2c, 0xa9, 0xd4, 0x2b, 0x33, 0x54, 0x72, 0xe4, 0xe2, 0x74, 0x82, 0xc9, 0x0b,
	0x09, 0x71, 0xb1, 0xe4, 0x25, 0xe6, 0xa6, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x81, 0xd9,
	0x42, 0x0a, 0x5c, 0xdc, 0x29, 0xa9, 0xc5, 0xc9, 0x45, 0x99, 0x05, 0x25, 0x99, 0xf9, 0x79, 0x12,
	0x4c, 0x60, 0x29, 0x64, 0x21, 0x27, 0x96, 0x28, 0xa6, 0x32, 0xc3, 0x24, 0x36, 0xb0, 0xe1, 0xc6,
	0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xae, 0x4a, 0x11, 0xe2, 0x6f, 0x00, 0x00, 0x00,
}