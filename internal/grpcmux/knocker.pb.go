// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: knocker.proto

package grpcmux

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

type KnockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceId int64 `protobuf:"varint,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
}

func (x *KnockRequest) Reset() {
	*x = KnockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_knocker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KnockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KnockRequest) ProtoMessage() {}

func (x *KnockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_knocker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KnockRequest.ProtoReflect.Descriptor instead.
func (*KnockRequest) Descriptor() ([]byte, []int) {
	return file_knocker_proto_rawDescGZIP(), []int{0}
}

func (x *KnockRequest) GetServiceId() int64 {
	if x != nil {
		return x.ServiceId
	}
	return 0
}

type KnockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *KnockResponse) Reset() {
	*x = KnockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_knocker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KnockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KnockResponse) ProtoMessage() {}

func (x *KnockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_knocker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KnockResponse.ProtoReflect.Descriptor instead.
func (*KnockResponse) Descriptor() ([]byte, []int) {
	return file_knocker_proto_rawDescGZIP(), []int{1}
}

var File_knocker_proto protoreflect.FileDescriptor

var file_knocker_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6b, 0x6e, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x67, 0x72, 0x70, 0x63, 0x6d, 0x75, 0x78, 0x22, 0x2d, 0x0a, 0x0c, 0x4b, 0x6e, 0x6f, 0x63,
	0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x22, 0x0f, 0x0a, 0x0d, 0x4b, 0x6e, 0x6f, 0x63, 0x6b,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x41, 0x0a, 0x07, 0x4b, 0x6e, 0x6f, 0x63,
	0x6b, 0x65, 0x72, 0x12, 0x36, 0x0a, 0x05, 0x4b, 0x6e, 0x6f, 0x63, 0x6b, 0x12, 0x15, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x6d, 0x75, 0x78, 0x2e, 0x4b, 0x6e, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x6d, 0x75, 0x78, 0x2e, 0x4b, 0x6e,
	0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x0b, 0x5a, 0x09, 0x2e,
	0x2f, 0x67, 0x72, 0x70, 0x63, 0x6d, 0x75, 0x78, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_knocker_proto_rawDescOnce sync.Once
	file_knocker_proto_rawDescData = file_knocker_proto_rawDesc
)

func file_knocker_proto_rawDescGZIP() []byte {
	file_knocker_proto_rawDescOnce.Do(func() {
		file_knocker_proto_rawDescData = protoimpl.X.CompressGZIP(file_knocker_proto_rawDescData)
	})
	return file_knocker_proto_rawDescData
}

var file_knocker_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_knocker_proto_goTypes = []interface{}{
	(*KnockRequest)(nil),  // 0: grpcmux.KnockRequest
	(*KnockResponse)(nil), // 1: grpcmux.KnockResponse
}
var file_knocker_proto_depIdxs = []int32{
	0, // 0: grpcmux.Knocker.Knock:input_type -> grpcmux.KnockRequest
	1, // 1: grpcmux.Knocker.Knock:output_type -> grpcmux.KnockResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_knocker_proto_init() }
func file_knocker_proto_init() {
	if File_knocker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_knocker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KnockRequest); i {
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
		file_knocker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KnockResponse); i {
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
			RawDescriptor: file_knocker_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_knocker_proto_goTypes,
		DependencyIndexes: file_knocker_proto_depIdxs,
		MessageInfos:      file_knocker_proto_msgTypes,
	}.Build()
	File_knocker_proto = out.File
	file_knocker_proto_rawDesc = nil
	file_knocker_proto_goTypes = nil
	file_knocker_proto_depIdxs = nil
}
