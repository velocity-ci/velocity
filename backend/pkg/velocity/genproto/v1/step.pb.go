// Code generated by protoc-gen-go. DO NOT EDIT.
// source: step.proto

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

type Step struct {
	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	// Types that are valid to be assigned to Impl:
	//	*Step_DockerRun
	//	*Step_DockerBuild
	//	*Step_DockerCompose
	//	*Step_DockerPush
	Impl                 isStep_Impl `protobuf_oneof:"impl"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Step) Reset()         { *m = Step{} }
func (m *Step) String() string { return proto.CompactTextString(m) }
func (*Step) ProtoMessage()    {}
func (*Step) Descriptor() ([]byte, []int) {
	return fileDescriptor_d3716d53a7e1b752, []int{0}
}

func (m *Step) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Step.Unmarshal(m, b)
}
func (m *Step) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Step.Marshal(b, m, deterministic)
}
func (m *Step) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Step.Merge(m, src)
}
func (m *Step) XXX_Size() int {
	return xxx_messageInfo_Step.Size(m)
}
func (m *Step) XXX_DiscardUnknown() {
	xxx_messageInfo_Step.DiscardUnknown(m)
}

var xxx_messageInfo_Step proto.InternalMessageInfo

func (m *Step) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type isStep_Impl interface {
	isStep_Impl()
}

type Step_DockerRun struct {
	DockerRun *DockerRun `protobuf:"bytes,101,opt,name=docker_run,json=dockerRun,proto3,oneof"`
}

type Step_DockerBuild struct {
	DockerBuild *DockerBuild `protobuf:"bytes,102,opt,name=docker_build,json=dockerBuild,proto3,oneof"`
}

type Step_DockerCompose struct {
	DockerCompose *DockerCompose `protobuf:"bytes,103,opt,name=docker_compose,json=dockerCompose,proto3,oneof"`
}

type Step_DockerPush struct {
	DockerPush *DockerPush `protobuf:"bytes,104,opt,name=docker_push,json=dockerPush,proto3,oneof"`
}

func (*Step_DockerRun) isStep_Impl() {}

func (*Step_DockerBuild) isStep_Impl() {}

func (*Step_DockerCompose) isStep_Impl() {}

func (*Step_DockerPush) isStep_Impl() {}

func (m *Step) GetImpl() isStep_Impl {
	if m != nil {
		return m.Impl
	}
	return nil
}

func (m *Step) GetDockerRun() *DockerRun {
	if x, ok := m.GetImpl().(*Step_DockerRun); ok {
		return x.DockerRun
	}
	return nil
}

func (m *Step) GetDockerBuild() *DockerBuild {
	if x, ok := m.GetImpl().(*Step_DockerBuild); ok {
		return x.DockerBuild
	}
	return nil
}

func (m *Step) GetDockerCompose() *DockerCompose {
	if x, ok := m.GetImpl().(*Step_DockerCompose); ok {
		return x.DockerCompose
	}
	return nil
}

func (m *Step) GetDockerPush() *DockerPush {
	if x, ok := m.GetImpl().(*Step_DockerPush); ok {
		return x.DockerPush
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Step) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Step_DockerRun)(nil),
		(*Step_DockerBuild)(nil),
		(*Step_DockerCompose)(nil),
		(*Step_DockerPush)(nil),
	}
}

type DockerRun struct {
	Image                string   `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
	Command              string   `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
	Environment          string   `protobuf:"bytes,3,opt,name=environment,proto3" json:"environment,omitempty"`
	WorkingDir           string   `protobuf:"bytes,4,opt,name=working_dir,json=workingDir,proto3" json:"working_dir,omitempty"`
	MountPoint           string   `protobuf:"bytes,5,opt,name=mount_point,json=mountPoint,proto3" json:"mount_point,omitempty"`
	IgnoreExitCode       string   `protobuf:"bytes,6,opt,name=ignore_exit_code,json=ignoreExitCode,proto3" json:"ignore_exit_code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DockerRun) Reset()         { *m = DockerRun{} }
func (m *DockerRun) String() string { return proto.CompactTextString(m) }
func (*DockerRun) ProtoMessage()    {}
func (*DockerRun) Descriptor() ([]byte, []int) {
	return fileDescriptor_d3716d53a7e1b752, []int{1}
}

func (m *DockerRun) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DockerRun.Unmarshal(m, b)
}
func (m *DockerRun) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DockerRun.Marshal(b, m, deterministic)
}
func (m *DockerRun) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DockerRun.Merge(m, src)
}
func (m *DockerRun) XXX_Size() int {
	return xxx_messageInfo_DockerRun.Size(m)
}
func (m *DockerRun) XXX_DiscardUnknown() {
	xxx_messageInfo_DockerRun.DiscardUnknown(m)
}

var xxx_messageInfo_DockerRun proto.InternalMessageInfo

func (m *DockerRun) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *DockerRun) GetCommand() string {
	if m != nil {
		return m.Command
	}
	return ""
}

func (m *DockerRun) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

func (m *DockerRun) GetWorkingDir() string {
	if m != nil {
		return m.WorkingDir
	}
	return ""
}

func (m *DockerRun) GetMountPoint() string {
	if m != nil {
		return m.MountPoint
	}
	return ""
}

func (m *DockerRun) GetIgnoreExitCode() string {
	if m != nil {
		return m.IgnoreExitCode
	}
	return ""
}

type DockerBuild struct {
	Dockerfile           string   `protobuf:"bytes,1,opt,name=dockerfile,proto3" json:"dockerfile,omitempty"`
	Context              string   `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	Tags                 []string `protobuf:"bytes,3,rep,name=tags,proto3" json:"tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DockerBuild) Reset()         { *m = DockerBuild{} }
func (m *DockerBuild) String() string { return proto.CompactTextString(m) }
func (*DockerBuild) ProtoMessage()    {}
func (*DockerBuild) Descriptor() ([]byte, []int) {
	return fileDescriptor_d3716d53a7e1b752, []int{2}
}

func (m *DockerBuild) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DockerBuild.Unmarshal(m, b)
}
func (m *DockerBuild) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DockerBuild.Marshal(b, m, deterministic)
}
func (m *DockerBuild) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DockerBuild.Merge(m, src)
}
func (m *DockerBuild) XXX_Size() int {
	return xxx_messageInfo_DockerBuild.Size(m)
}
func (m *DockerBuild) XXX_DiscardUnknown() {
	xxx_messageInfo_DockerBuild.DiscardUnknown(m)
}

var xxx_messageInfo_DockerBuild proto.InternalMessageInfo

func (m *DockerBuild) GetDockerfile() string {
	if m != nil {
		return m.Dockerfile
	}
	return ""
}

func (m *DockerBuild) GetContext() string {
	if m != nil {
		return m.Context
	}
	return ""
}

func (m *DockerBuild) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

type DockerCompose struct {
	ComposeFile          string   `protobuf:"bytes,1,opt,name=compose_file,json=composeFile,proto3" json:"compose_file,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DockerCompose) Reset()         { *m = DockerCompose{} }
func (m *DockerCompose) String() string { return proto.CompactTextString(m) }
func (*DockerCompose) ProtoMessage()    {}
func (*DockerCompose) Descriptor() ([]byte, []int) {
	return fileDescriptor_d3716d53a7e1b752, []int{3}
}

func (m *DockerCompose) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DockerCompose.Unmarshal(m, b)
}
func (m *DockerCompose) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DockerCompose.Marshal(b, m, deterministic)
}
func (m *DockerCompose) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DockerCompose.Merge(m, src)
}
func (m *DockerCompose) XXX_Size() int {
	return xxx_messageInfo_DockerCompose.Size(m)
}
func (m *DockerCompose) XXX_DiscardUnknown() {
	xxx_messageInfo_DockerCompose.DiscardUnknown(m)
}

var xxx_messageInfo_DockerCompose proto.InternalMessageInfo

func (m *DockerCompose) GetComposeFile() string {
	if m != nil {
		return m.ComposeFile
	}
	return ""
}

type DockerPush struct {
	Tags                 []string `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DockerPush) Reset()         { *m = DockerPush{} }
func (m *DockerPush) String() string { return proto.CompactTextString(m) }
func (*DockerPush) ProtoMessage()    {}
func (*DockerPush) Descriptor() ([]byte, []int) {
	return fileDescriptor_d3716d53a7e1b752, []int{4}
}

func (m *DockerPush) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DockerPush.Unmarshal(m, b)
}
func (m *DockerPush) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DockerPush.Marshal(b, m, deterministic)
}
func (m *DockerPush) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DockerPush.Merge(m, src)
}
func (m *DockerPush) XXX_Size() int {
	return xxx_messageInfo_DockerPush.Size(m)
}
func (m *DockerPush) XXX_DiscardUnknown() {
	xxx_messageInfo_DockerPush.DiscardUnknown(m)
}

var xxx_messageInfo_DockerPush proto.InternalMessageInfo

func (m *DockerPush) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func init() {
	proto.RegisterType((*Step)(nil), "velocity.v1.Step")
	proto.RegisterType((*DockerRun)(nil), "velocity.v1.DockerRun")
	proto.RegisterType((*DockerBuild)(nil), "velocity.v1.DockerBuild")
	proto.RegisterType((*DockerCompose)(nil), "velocity.v1.DockerCompose")
	proto.RegisterType((*DockerPush)(nil), "velocity.v1.DockerPush")
}

func init() { proto.RegisterFile("step.proto", fileDescriptor_d3716d53a7e1b752) }

var fileDescriptor_d3716d53a7e1b752 = []byte{
	// 408 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0x41, 0x6f, 0xd3, 0x40,
	0x10, 0x85, 0xe3, 0xc4, 0x0d, 0xca, 0xb8, 0xad, 0xd0, 0x0a, 0xc1, 0x8a, 0x03, 0x18, 0x9f, 0x72,
	0x8a, 0xd4, 0x72, 0x40, 0x42, 0xe2, 0x92, 0x14, 0x94, 0x63, 0x65, 0x6e, 0x70, 0xb0, 0x52, 0xef,
	0xd4, 0x19, 0xd5, 0xde, 0x5d, 0xad, 0xd7, 0x26, 0xfc, 0x24, 0x7e, 0x0a, 0xff, 0x0a, 0x79, 0xbd,
	0xae, 0x8d, 0x94, 0x5b, 0xde, 0xb7, 0x6f, 0x5e, 0xf4, 0x66, 0x0c, 0x50, 0x5b, 0xd4, 0x1b, 0x6d,
	0x94, 0x55, 0x2c, 0x6a, 0xb1, 0x54, 0x39, 0xd9, 0xdf, 0x9b, 0xf6, 0x26, 0xf9, 0x33, 0x87, 0xf0,
	0xbb, 0x45, 0xcd, 0x62, 0x88, 0x04, 0xd6, 0xb9, 0x21, 0x6d, 0x49, 0x49, 0x1e, 0xc4, 0xc1, 0x7a,
	0x95, 0x4e, 0x11, 0xfb, 0x04, 0x20, 0x54, 0xfe, 0x84, 0x26, 0x33, 0x8d, 0xe4, 0x18, 0x07, 0xeb,
	0xe8, 0xf6, 0xf5, 0x66, 0x12, 0xb6, 0xb9, 0x73, 0xcf, 0x69, 0x23, 0xf7, 0xb3, 0x74, 0x25, 0x06,
	0xc1, 0xbe, 0xc0, 0xa5, 0x1f, 0x7c, 0x68, 0xa8, 0x14, 0xfc, 0xd1, 0x8d, 0xf2, 0x33, 0xa3, 0xdb,
	0xee, 0x7d, 0x3f, 0x4b, 0x23, 0x31, 0x4a, 0xb6, 0x83, 0x6b, 0x3f, 0x9e, 0xab, 0x4a, 0xab, 0x1a,
	0x79, 0xe1, 0x02, 0xde, 0x9e, 0x09, 0xd8, 0xf5, 0x8e, 0xfd, 0x2c, 0xbd, 0x12, 0x53, 0xc0, 0x3e,
	0x83, 0xcf, 0xcc, 0x74, 0x53, 0x1f, 0xf9, 0xd1, 0x25, 0xbc, 0x39, 0x93, 0x70, 0xdf, 0xd4, 0xc7,
	0xfd, 0x2c, 0xf5, 0x55, 0x3b, 0xb5, 0x5d, 0x42, 0x48, 0x95, 0x2e, 0x93, 0xbf, 0x01, 0xac, 0x9e,
	0x2b, 0xb2, 0x57, 0x70, 0x41, 0xd5, 0xa1, 0x40, 0xbf, 0xaa, 0x5e, 0x30, 0x0e, 0x2f, 0x72, 0x55,
	0x55, 0x07, 0x29, 0xf8, 0xdc, 0xf1, 0x41, 0x76, 0x0b, 0x46, 0xd9, 0x92, 0x51, 0xb2, 0x42, 0x69,
	0xf9, 0xa2, 0x5f, 0xf0, 0x04, 0xb1, 0xf7, 0x10, 0xfd, 0x52, 0xe6, 0x89, 0x64, 0x91, 0x09, 0x32,
	0x3c, 0x74, 0x0e, 0xf0, 0xe8, 0x8e, 0x4c, 0x67, 0xa8, 0x54, 0x23, 0x6d, 0xa6, 0x15, 0x49, 0xcb,
	0x2f, 0x7a, 0x83, 0x43, 0xf7, 0x1d, 0x61, 0x6b, 0x78, 0x49, 0x85, 0x54, 0x06, 0x33, 0x3c, 0x91,
	0xcd, 0x72, 0x25, 0x90, 0x2f, 0x9d, 0xeb, 0xba, 0xe7, 0x5f, 0x4f, 0x64, 0x77, 0x4a, 0x60, 0xf2,
	0x13, 0xa2, 0xc9, 0xca, 0xd9, 0xbb, 0xe1, 0xb6, 0x8f, 0x54, 0x0e, 0x8d, 0x26, 0xa4, 0xaf, 0x25,
	0x2d, 0x9e, 0xec, 0x58, 0xcb, 0x49, 0xc6, 0x20, 0xb4, 0x87, 0xa2, 0xe6, 0x8b, 0x78, 0xb1, 0x5e,
	0xa5, 0xee, 0x77, 0x72, 0x0b, 0x57, 0xff, 0x9d, 0x83, 0x7d, 0x80, 0x4b, 0x7f, 0xbb, 0x6c, 0xf2,
	0x07, 0x91, 0x67, 0xdf, 0xa8, 0xc4, 0x24, 0x06, 0x18, 0x0f, 0xf0, 0x9c, 0x1a, 0x8c, 0xa9, 0xdb,
	0xf0, 0xc7, 0xbc, 0xbd, 0x79, 0x58, 0xba, 0x8f, 0xf8, 0xe3, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xeb, 0x10, 0x6b, 0x84, 0xd2, 0x02, 0x00, 0x00,
}
