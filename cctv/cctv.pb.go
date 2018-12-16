// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cctv.proto

package cctv

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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

type Channels struct {
	Mask                 int32    `protobuf:"varint,1,opt,name=mask,proto3" json:"mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Channels) Reset()         { *m = Channels{} }
func (m *Channels) String() string { return proto.CompactTextString(m) }
func (*Channels) ProtoMessage()    {}
func (*Channels) Descriptor() ([]byte, []int) {
	return fileDescriptor_cctv_80369915c06fdeb4, []int{0}
}
func (m *Channels) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Channels.Unmarshal(m, b)
}
func (m *Channels) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Channels.Marshal(b, m, deterministic)
}
func (dst *Channels) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Channels.Merge(dst, src)
}
func (m *Channels) XXX_Size() int {
	return xxx_messageInfo_Channels.Size(m)
}
func (m *Channels) XXX_DiscardUnknown() {
	xxx_messageInfo_Channels.DiscardUnknown(m)
}

var xxx_messageInfo_Channels proto.InternalMessageInfo

func (m *Channels) GetMask() int32 {
	if m != nil {
		return m.Mask
	}
	return 0
}

type Frame struct {
	Channel              int32    `protobuf:"varint,1,opt,name=channel,proto3" json:"channel,omitempty"`
	Ts                   int64    `protobuf:"varint,2,opt,name=ts,proto3" json:"ts,omitempty"`
	Image                []byte   `protobuf:"bytes,5,opt,name=image,proto3" json:"image,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Frame) Reset()         { *m = Frame{} }
func (m *Frame) String() string { return proto.CompactTextString(m) }
func (*Frame) ProtoMessage()    {}
func (*Frame) Descriptor() ([]byte, []int) {
	return fileDescriptor_cctv_80369915c06fdeb4, []int{1}
}
func (m *Frame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Frame.Unmarshal(m, b)
}
func (m *Frame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Frame.Marshal(b, m, deterministic)
}
func (dst *Frame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Frame.Merge(dst, src)
}
func (m *Frame) XXX_Size() int {
	return xxx_messageInfo_Frame.Size(m)
}
func (m *Frame) XXX_DiscardUnknown() {
	xxx_messageInfo_Frame.DiscardUnknown(m)
}

var xxx_messageInfo_Frame proto.InternalMessageInfo

func (m *Frame) GetChannel() int32 {
	if m != nil {
		return m.Channel
	}
	return 0
}

func (m *Frame) GetTs() int64 {
	if m != nil {
		return m.Ts
	}
	return 0
}

func (m *Frame) GetImage() []byte {
	if m != nil {
		return m.Image
	}
	return nil
}

func init() {
	proto.RegisterType((*Channels)(nil), "cctv.Channels")
	proto.RegisterType((*Frame)(nil), "cctv.Frame")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CCTVClient is the client API for CCTV service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CCTVClient interface {
	Feeds(ctx context.Context, in *Channels, opts ...grpc.CallOption) (CCTV_FeedsClient, error)
}

type cCTVClient struct {
	cc *grpc.ClientConn
}

func NewCCTVClient(cc *grpc.ClientConn) CCTVClient {
	return &cCTVClient{cc}
}

func (c *cCTVClient) Feeds(ctx context.Context, in *Channels, opts ...grpc.CallOption) (CCTV_FeedsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_CCTV_serviceDesc.Streams[0], "/cctv.CCTV/Feeds", opts...)
	if err != nil {
		return nil, err
	}
	x := &cCTVFeedsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CCTV_FeedsClient interface {
	Recv() (*Frame, error)
	grpc.ClientStream
}

type cCTVFeedsClient struct {
	grpc.ClientStream
}

func (x *cCTVFeedsClient) Recv() (*Frame, error) {
	m := new(Frame)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CCTVServer is the server API for CCTV service.
type CCTVServer interface {
	Feeds(*Channels, CCTV_FeedsServer) error
}

func RegisterCCTVServer(s *grpc.Server, srv CCTVServer) {
	s.RegisterService(&_CCTV_serviceDesc, srv)
}

func _CCTV_Feeds_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Channels)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CCTVServer).Feeds(m, &cCTVFeedsServer{stream})
}

type CCTV_FeedsServer interface {
	Send(*Frame) error
	grpc.ServerStream
}

type cCTVFeedsServer struct {
	grpc.ServerStream
}

func (x *cCTVFeedsServer) Send(m *Frame) error {
	return x.ServerStream.SendMsg(m)
}

var _CCTV_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cctv.CCTV",
	HandlerType: (*CCTVServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Feeds",
			Handler:       _CCTV_Feeds_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cctv.proto",
}

func init() { proto.RegisterFile("cctv.proto", fileDescriptor_cctv_80369915c06fdeb4) }

var fileDescriptor_cctv_80369915c06fdeb4 = []byte{
	// 161 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4a, 0x4e, 0x2e, 0x29,
	0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x01, 0xb1, 0x95, 0xe4, 0xb8, 0x38, 0x9c, 0x33,
	0x12, 0xf3, 0xf2, 0x52, 0x73, 0x8a, 0x85, 0x84, 0xb8, 0x58, 0x72, 0x13, 0x8b, 0xb3, 0x25, 0x18,
	0x15, 0x18, 0x35, 0x58, 0x83, 0xc0, 0x6c, 0x25, 0x77, 0x2e, 0x56, 0xb7, 0xa2, 0xc4, 0xdc, 0x54,
	0x21, 0x09, 0x2e, 0xf6, 0x64, 0x88, 0x42, 0xa8, 0x3c, 0x8c, 0x2b, 0xc4, 0xc7, 0xc5, 0x54, 0x52,
	0x2c, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x1c, 0xc4, 0x54, 0x52, 0x2c, 0x24, 0xc2, 0xc5, 0x9a, 0x99,
	0x9b, 0x98, 0x9e, 0x2a, 0xc1, 0xaa, 0xc0, 0xa8, 0xc1, 0x13, 0x04, 0xe1, 0x18, 0x19, 0x70, 0xb1,
	0x38, 0x3b, 0x87, 0x84, 0x09, 0x69, 0x70, 0xb1, 0xba, 0xa5, 0xa6, 0xa6, 0x14, 0x0b, 0xf1, 0xe9,
	0x81, 0x1d, 0x03, 0xb3, 0x5d, 0x8a, 0x1b, 0xc2, 0x07, 0xdb, 0xa6, 0xc4, 0x60, 0xc0, 0x98, 0xc4,
	0x06, 0x76, 0xa7, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x21, 0xb2, 0x5b, 0xcc, 0xb5, 0x00, 0x00,
	0x00,
}
