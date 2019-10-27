// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pipeline.proto

package v1

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type Pipeline struct {
	// PIPELINE_ID = commit_sha+pipeline_name
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ProjectId            string   `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	CommitId             string   `protobuf:"bytes,3,opt,name=commit_id,json=commitId,proto3" json:"commit_id,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Stages               []*Stage `protobuf:"bytes,5,rep,name=stages,proto3" json:"stages,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Pipeline) Reset()         { *m = Pipeline{} }
func (m *Pipeline) String() string { return proto.CompactTextString(m) }
func (*Pipeline) ProtoMessage()    {}
func (*Pipeline) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{0}
}

func (m *Pipeline) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Pipeline.Unmarshal(m, b)
}
func (m *Pipeline) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Pipeline.Marshal(b, m, deterministic)
}
func (m *Pipeline) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Pipeline.Merge(m, src)
}
func (m *Pipeline) XXX_Size() int {
	return xxx_messageInfo_Pipeline.Size(m)
}
func (m *Pipeline) XXX_DiscardUnknown() {
	xxx_messageInfo_Pipeline.DiscardUnknown(m)
}

var xxx_messageInfo_Pipeline proto.InternalMessageInfo

func (m *Pipeline) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Pipeline) GetProjectId() string {
	if m != nil {
		return m.ProjectId
	}
	return ""
}

func (m *Pipeline) GetCommitId() string {
	if m != nil {
		return m.CommitId
	}
	return ""
}

func (m *Pipeline) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Pipeline) GetStages() []*Stage {
	if m != nil {
		return m.Stages
	}
	return nil
}

type Stage struct {
	Name                 string       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Blueprints           []*Blueprint `protobuf:"bytes,2,rep,name=blueprints,proto3" json:"blueprints,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Stage) Reset()         { *m = Stage{} }
func (m *Stage) String() string { return proto.CompactTextString(m) }
func (*Stage) ProtoMessage()    {}
func (*Stage) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{1}
}

func (m *Stage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stage.Unmarshal(m, b)
}
func (m *Stage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stage.Marshal(b, m, deterministic)
}
func (m *Stage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stage.Merge(m, src)
}
func (m *Stage) XXX_Size() int {
	return xxx_messageInfo_Stage.Size(m)
}
func (m *Stage) XXX_DiscardUnknown() {
	xxx_messageInfo_Stage.DiscardUnknown(m)
}

var xxx_messageInfo_Stage proto.InternalMessageInfo

func (m *Stage) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Stage) GetBlueprints() []*Blueprint {
	if m != nil {
		return m.Blueprints
	}
	return nil
}

type PipelineQuery struct {
	Ids                  []string `protobuf:"bytes,1,rep,name=ids,proto3" json:"ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PipelineQuery) Reset()         { *m = PipelineQuery{} }
func (m *PipelineQuery) String() string { return proto.CompactTextString(m) }
func (*PipelineQuery) ProtoMessage()    {}
func (*PipelineQuery) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{2}
}

func (m *PipelineQuery) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PipelineQuery.Unmarshal(m, b)
}
func (m *PipelineQuery) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PipelineQuery.Marshal(b, m, deterministic)
}
func (m *PipelineQuery) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PipelineQuery.Merge(m, src)
}
func (m *PipelineQuery) XXX_Size() int {
	return xxx_messageInfo_PipelineQuery.Size(m)
}
func (m *PipelineQuery) XXX_DiscardUnknown() {
	xxx_messageInfo_PipelineQuery.DiscardUnknown(m)
}

var xxx_messageInfo_PipelineQuery proto.InternalMessageInfo

func (m *PipelineQuery) GetIds() []string {
	if m != nil {
		return m.Ids
	}
	return nil
}

type GetPipelineRequest struct {
	// The id of the project in the form of
	// `[PROJECT_ID]`.
	ProjectId string `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	// The id of the Pipeline in the form of
	// `[PIPELINE_ID]`.
	PipelineId           string   `protobuf:"bytes,2,opt,name=Pipeline_id,json=PipelineId,proto3" json:"Pipeline_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPipelineRequest) Reset()         { *m = GetPipelineRequest{} }
func (m *GetPipelineRequest) String() string { return proto.CompactTextString(m) }
func (*GetPipelineRequest) ProtoMessage()    {}
func (*GetPipelineRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{3}
}

func (m *GetPipelineRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPipelineRequest.Unmarshal(m, b)
}
func (m *GetPipelineRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPipelineRequest.Marshal(b, m, deterministic)
}
func (m *GetPipelineRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPipelineRequest.Merge(m, src)
}
func (m *GetPipelineRequest) XXX_Size() int {
	return xxx_messageInfo_GetPipelineRequest.Size(m)
}
func (m *GetPipelineRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPipelineRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPipelineRequest proto.InternalMessageInfo

func (m *GetPipelineRequest) GetProjectId() string {
	if m != nil {
		return m.ProjectId
	}
	return ""
}

func (m *GetPipelineRequest) GetPipelineId() string {
	if m != nil {
		return m.PipelineId
	}
	return ""
}

type ListPipelinesRequest struct {
	RepoQuery            *RepoQuery `protobuf:"bytes,1,opt,name=repo_query,json=repoQuery,proto3" json:"repo_query,omitempty"`
	PageQuery            *PageQuery `protobuf:"bytes,99,opt,name=page_query,json=pageQuery,proto3" json:"page_query,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ListPipelinesRequest) Reset()         { *m = ListPipelinesRequest{} }
func (m *ListPipelinesRequest) String() string { return proto.CompactTextString(m) }
func (*ListPipelinesRequest) ProtoMessage()    {}
func (*ListPipelinesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{4}
}

func (m *ListPipelinesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListPipelinesRequest.Unmarshal(m, b)
}
func (m *ListPipelinesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListPipelinesRequest.Marshal(b, m, deterministic)
}
func (m *ListPipelinesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListPipelinesRequest.Merge(m, src)
}
func (m *ListPipelinesRequest) XXX_Size() int {
	return xxx_messageInfo_ListPipelinesRequest.Size(m)
}
func (m *ListPipelinesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListPipelinesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListPipelinesRequest proto.InternalMessageInfo

func (m *ListPipelinesRequest) GetRepoQuery() *RepoQuery {
	if m != nil {
		return m.RepoQuery
	}
	return nil
}

func (m *ListPipelinesRequest) GetPageQuery() *PageQuery {
	if m != nil {
		return m.PageQuery
	}
	return nil
}

type ListPipelinesResponse struct {
	Pipelines            []*Pipeline `protobuf:"bytes,1,rep,name=Pipelines,proto3" json:"Pipelines,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ListPipelinesResponse) Reset()         { *m = ListPipelinesResponse{} }
func (m *ListPipelinesResponse) String() string { return proto.CompactTextString(m) }
func (*ListPipelinesResponse) ProtoMessage()    {}
func (*ListPipelinesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7ac67a7adf3df9c7, []int{5}
}

func (m *ListPipelinesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListPipelinesResponse.Unmarshal(m, b)
}
func (m *ListPipelinesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListPipelinesResponse.Marshal(b, m, deterministic)
}
func (m *ListPipelinesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListPipelinesResponse.Merge(m, src)
}
func (m *ListPipelinesResponse) XXX_Size() int {
	return xxx_messageInfo_ListPipelinesResponse.Size(m)
}
func (m *ListPipelinesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListPipelinesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListPipelinesResponse proto.InternalMessageInfo

func (m *ListPipelinesResponse) GetPipelines() []*Pipeline {
	if m != nil {
		return m.Pipelines
	}
	return nil
}

func init() {
	proto.RegisterType((*Pipeline)(nil), "velocity.v1.Pipeline")
	proto.RegisterType((*Stage)(nil), "velocity.v1.Stage")
	proto.RegisterType((*PipelineQuery)(nil), "velocity.v1.PipelineQuery")
	proto.RegisterType((*GetPipelineRequest)(nil), "velocity.v1.GetPipelineRequest")
	proto.RegisterType((*ListPipelinesRequest)(nil), "velocity.v1.ListPipelinesRequest")
	proto.RegisterType((*ListPipelinesResponse)(nil), "velocity.v1.ListPipelinesResponse")
}

func init() { proto.RegisterFile("pipeline.proto", fileDescriptor_7ac67a7adf3df9c7) }

var fileDescriptor_7ac67a7adf3df9c7 = []byte{
	// 473 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0x4d, 0x6b, 0xdb, 0x40,
	0x10, 0x45, 0xb2, 0x13, 0xa2, 0x11, 0xf9, 0x60, 0x68, 0x8a, 0x70, 0x5b, 0x6c, 0x6f, 0x0f, 0x75,
	0x7d, 0xb0, 0xb0, 0x43, 0x7b, 0xeb, 0x25, 0x97, 0x62, 0xc8, 0x21, 0x95, 0x7b, 0xea, 0x25, 0x28,
	0xd2, 0x20, 0xb6, 0xd8, 0xda, 0x8d, 0x76, 0x2d, 0x30, 0x25, 0x14, 0x4a, 0xfe, 0x41, 0xa1, 0x7f,
	0xac, 0x7f, 0xa1, 0x3f, 0xa4, 0x68, 0xad, 0xb5, 0x2d, 0xd7, 0xf4, 0x36, 0xfb, 0x66, 0xe6, 0xed,
	0x7b, 0x6f, 0x59, 0x38, 0x93, 0x5c, 0xd2, 0x9c, 0xe7, 0x34, 0x92, 0x85, 0xd0, 0x02, 0xfd, 0x92,
	0xe6, 0x22, 0xe1, 0x7a, 0x35, 0x2a, 0xc7, 0x1d, 0x2f, 0x96, 0x7c, 0x8d, 0x77, 0xce, 0xef, 0xe7,
	0x4b, 0x92, 0x05, 0xcf, 0x75, 0x0d, 0x5c, 0x14, 0x24, 0x85, 0xe2, 0x5a, 0x14, 0xab, 0x1a, 0x79,
	0x99, 0x09, 0x91, 0xcd, 0x29, 0x8c, 0x25, 0x0f, 0xe3, 0x3c, 0x17, 0x3a, 0xd6, 0x5c, 0xe4, 0x6a,
	0xdd, 0x65, 0xbf, 0x1c, 0x38, 0xb9, 0xad, 0xef, 0xc2, 0x33, 0x70, 0x79, 0x1a, 0x38, 0x3d, 0x67,
	0xe0, 0x45, 0x2e, 0x4f, 0xf1, 0x15, 0x80, 0x2c, 0xc4, 0x57, 0x4a, 0xf4, 0x1d, 0x4f, 0x03, 0xd7,
	0xe0, 0x5e, 0x8d, 0x4c, 0x53, 0x7c, 0x01, 0x5e, 0x22, 0x16, 0x0b, 0x6e, 0xba, 0x2d, 0xd3, 0x3d,
	0x59, 0x03, 0xd3, 0x14, 0x11, 0xda, 0x79, 0xbc, 0xa0, 0xa0, 0x6d, 0x70, 0x53, 0xe3, 0x10, 0x8e,
	0x95, 0x8e, 0x33, 0x52, 0xc1, 0x51, 0xaf, 0x35, 0xf0, 0x27, 0x38, 0xda, 0xb1, 0x35, 0x9a, 0x55,
	0xad, 0xa8, 0x9e, 0x60, 0x33, 0x38, 0x32, 0xc0, 0x86, 0xc8, 0xd9, 0x21, 0x7a, 0x0f, 0xb0, 0x31,
	0xae, 0x02, 0xd7, 0x90, 0x3d, 0x6f, 0x90, 0x5d, 0xdb, 0x76, 0xb4, 0x33, 0xc9, 0xfa, 0x70, 0x6a,
	0xcd, 0x7e, 0x5a, 0x52, 0xb1, 0xc2, 0x0b, 0x68, 0xf1, 0x54, 0x05, 0x4e, 0xaf, 0x35, 0xf0, 0xa2,
	0xaa, 0x64, 0x9f, 0x01, 0x3f, 0x92, 0xb6, 0x53, 0x11, 0x3d, 0x2c, 0x49, 0xe9, 0xbd, 0x24, 0x9c,
	0xfd, 0x24, 0xba, 0xe0, 0xdb, 0x8d, 0x6d, 0x52, 0x60, 0xa1, 0x69, 0xca, 0x9e, 0x1c, 0x78, 0x76,
	0xc3, 0xd5, 0x86, 0x57, 0x59, 0xe2, 0x77, 0x00, 0xd5, 0x8b, 0xdd, 0x3d, 0x54, 0x72, 0x0c, 0xf1,
	0xbe, 0x93, 0x88, 0xa4, 0x30, 0x62, 0x23, 0xaf, 0xb0, 0x65, 0xb5, 0x26, 0xe3, 0x8c, 0xea, 0xb5,
	0xe4, 0xc0, 0xda, 0x6d, 0x9c, 0x51, 0xbd, 0x26, 0x6d, 0xc9, 0x6e, 0xe0, 0x72, 0x4f, 0x85, 0x92,
	0x22, 0x57, 0x84, 0x57, 0xe0, 0x6d, 0x40, 0x93, 0x86, 0x3f, 0xb9, 0x6c, 0xd2, 0xd9, 0x40, 0xb6,
	0x73, 0x93, 0x27, 0x17, 0xce, 0xed, 0x69, 0x46, 0x45, 0xc9, 0x13, 0x42, 0x09, 0xfe, 0x4e, 0x7c,
	0xd8, 0x6d, 0x90, 0xfc, 0x1b, 0x6c, 0xe7, 0xf0, 0x2d, 0xec, 0xed, 0x8f, 0xdf, 0x7f, 0x7e, 0xba,
	0xaf, 0xb1, 0x1f, 0x96, 0xe3, 0xf0, 0x5b, 0xf5, 0xe6, 0x1f, 0xea, 0xb4, 0x55, 0x38, 0x0c, 0xed,
	0xe7, 0x50, 0xe1, 0xf0, 0x11, 0xbf, 0xc3, 0x69, 0xc3, 0x13, 0xf6, 0x1b, 0x94, 0x87, 0x52, 0xef,
	0xb0, 0xff, 0x8d, 0xac, 0x23, 0x61, 0x6f, 0x8c, 0x84, 0x3e, 0x76, 0x0f, 0x49, 0x78, 0xdc, 0x6a,
	0xb8, 0x6e, 0x7f, 0x71, 0xcb, 0xf1, 0xfd, 0xb1, 0xf9, 0x4f, 0x57, 0x7f, 0x03, 0x00, 0x00, 0xff,
	0xff, 0xea, 0x28, 0xaa, 0xa8, 0xba, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PipelineServiceClient is the client API for PipelineService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PipelineServiceClient interface {
	GetPipeline(ctx context.Context, in *GetPipelineRequest, opts ...grpc.CallOption) (*Pipeline, error)
	ListPipelines(ctx context.Context, in *ListPipelinesRequest, opts ...grpc.CallOption) (*ListPipelinesResponse, error)
}

type pipelineServiceClient struct {
	cc *grpc.ClientConn
}

func NewPipelineServiceClient(cc *grpc.ClientConn) PipelineServiceClient {
	return &pipelineServiceClient{cc}
}

func (c *pipelineServiceClient) GetPipeline(ctx context.Context, in *GetPipelineRequest, opts ...grpc.CallOption) (*Pipeline, error) {
	out := new(Pipeline)
	err := c.cc.Invoke(ctx, "/velocity.v1.PipelineService/GetPipeline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pipelineServiceClient) ListPipelines(ctx context.Context, in *ListPipelinesRequest, opts ...grpc.CallOption) (*ListPipelinesResponse, error) {
	out := new(ListPipelinesResponse)
	err := c.cc.Invoke(ctx, "/velocity.v1.PipelineService/ListPipelines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PipelineServiceServer is the server API for PipelineService service.
type PipelineServiceServer interface {
	GetPipeline(context.Context, *GetPipelineRequest) (*Pipeline, error)
	ListPipelines(context.Context, *ListPipelinesRequest) (*ListPipelinesResponse, error)
}

// UnimplementedPipelineServiceServer can be embedded to have forward compatible implementations.
type UnimplementedPipelineServiceServer struct {
}

func (*UnimplementedPipelineServiceServer) GetPipeline(ctx context.Context, req *GetPipelineRequest) (*Pipeline, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPipeline not implemented")
}
func (*UnimplementedPipelineServiceServer) ListPipelines(ctx context.Context, req *ListPipelinesRequest) (*ListPipelinesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPipelines not implemented")
}

func RegisterPipelineServiceServer(s *grpc.Server, srv PipelineServiceServer) {
	s.RegisterService(&_PipelineService_serviceDesc, srv)
}

func _PipelineService_GetPipeline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPipelineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PipelineServiceServer).GetPipeline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/velocity.v1.PipelineService/GetPipeline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PipelineServiceServer).GetPipeline(ctx, req.(*GetPipelineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PipelineService_ListPipelines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPipelinesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PipelineServiceServer).ListPipelines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/velocity.v1.PipelineService/ListPipelines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PipelineServiceServer).ListPipelines(ctx, req.(*ListPipelinesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PipelineService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "velocity.v1.PipelineService",
	HandlerType: (*PipelineServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPipeline",
			Handler:    _PipelineService_GetPipeline_Handler,
		},
		{
			MethodName: "ListPipelines",
			Handler:    _PipelineService_ListPipelines_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pipeline.proto",
}
