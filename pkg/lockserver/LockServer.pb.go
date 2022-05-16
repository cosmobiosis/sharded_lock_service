// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.9.1
// source: pkg/lockserver/LockServer.proto

package lockserver

import (
	_ "github.com/golang/protobuf/ptypes/empty"
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

type AcquireLocksInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId  string   `protobuf:"bytes,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	ReadKeys  []string `protobuf:"bytes,2,rep,name=readKeys,proto3" json:"readKeys,omitempty"`
	WriteKeys []string `protobuf:"bytes,3,rep,name=writeKeys,proto3" json:"writeKeys,omitempty"`
}

func (x *AcquireLocksInfo) Reset() {
	*x = AcquireLocksInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_lockserver_LockServer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AcquireLocksInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcquireLocksInfo) ProtoMessage() {}

func (x *AcquireLocksInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_lockserver_LockServer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcquireLocksInfo.ProtoReflect.Descriptor instead.
func (*AcquireLocksInfo) Descriptor() ([]byte, []int) {
	return file_pkg_lockserver_LockServer_proto_rawDescGZIP(), []int{0}
}

func (x *AcquireLocksInfo) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *AcquireLocksInfo) GetReadKeys() []string {
	if x != nil {
		return x.ReadKeys
	}
	return nil
}

func (x *AcquireLocksInfo) GetWriteKeys() []string {
	if x != nil {
		return x.WriteKeys
	}
	return nil
}

type ReleaseLocksInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId  string   `protobuf:"bytes,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	ReadKeys  []string `protobuf:"bytes,2,rep,name=readKeys,proto3" json:"readKeys,omitempty"`
	WriteKeys []string `protobuf:"bytes,3,rep,name=writeKeys,proto3" json:"writeKeys,omitempty"`
}

func (x *ReleaseLocksInfo) Reset() {
	*x = ReleaseLocksInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_lockserver_LockServer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReleaseLocksInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReleaseLocksInfo) ProtoMessage() {}

func (x *ReleaseLocksInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_lockserver_LockServer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReleaseLocksInfo.ProtoReflect.Descriptor instead.
func (*ReleaseLocksInfo) Descriptor() ([]byte, []int) {
	return file_pkg_lockserver_LockServer_proto_rawDescGZIP(), []int{1}
}

func (x *ReleaseLocksInfo) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *ReleaseLocksInfo) GetReadKeys() []string {
	if x != nil {
		return x.ReadKeys
	}
	return nil
}

func (x *ReleaseLocksInfo) GetWriteKeys() []string {
	if x != nil {
		return x.WriteKeys
	}
	return nil
}

type Success struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Flag bool `protobuf:"varint,1,opt,name=flag,proto3" json:"flag,omitempty"`
}

func (x *Success) Reset() {
	*x = Success{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_lockserver_LockServer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Success) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Success) ProtoMessage() {}

func (x *Success) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_lockserver_LockServer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Success.ProtoReflect.Descriptor instead.
func (*Success) Descriptor() ([]byte, []int) {
	return file_pkg_lockserver_LockServer_proto_rawDescGZIP(), []int{2}
}

func (x *Success) GetFlag() bool {
	if x != nil {
		return x.Flag
	}
	return false
}

var File_pkg_lockserver_LockServer_proto protoreflect.FileDescriptor

var file_pkg_lockserver_LockServer_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2f, 0x4c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0a, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x1a, 0x1b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x68, 0x0a, 0x10, 0x41, 0x63,
	0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x6f, 0x63, 0x6b, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65,
	0x61, 0x64, 0x4b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65,
	0x61, 0x64, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x77, 0x72, 0x69, 0x74, 0x65, 0x4b,
	0x65, 0x79, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x77, 0x72, 0x69, 0x74, 0x65,
	0x4b, 0x65, 0x79, 0x73, 0x22, 0x68, 0x0a, 0x10, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x4c,
	0x6f, 0x63, 0x6b, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x61, 0x64, 0x4b, 0x65, 0x79, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x61, 0x64, 0x4b, 0x65, 0x79, 0x73,
	0x12, 0x1c, 0x0a, 0x09, 0x77, 0x72, 0x69, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x09, 0x77, 0x72, 0x69, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x73, 0x22, 0x1d,
	0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x6c, 0x61,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x32, 0x8d, 0x01,
	0x0a, 0x0b, 0x4c, 0x6f, 0x63, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3e, 0x0a,
	0x07, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x12, 0x1c, 0x2e, 0x6c, 0x6f, 0x63, 0x6b, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x6f, 0x63,
	0x6b, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x13, 0x2e, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x00, 0x12, 0x3e, 0x0a,
	0x07, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x12, 0x1c, 0x2e, 0x6c, 0x6f, 0x63, 0x6b, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x4c, 0x6f, 0x63,
	0x6b, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x13, 0x2e, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x00, 0x42, 0x25, 0x5a,
	0x23, 0x73, 0x68, 0x61, 0x72, 0x64, 0x65, 0x64, 0x5f, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_lockserver_LockServer_proto_rawDescOnce sync.Once
	file_pkg_lockserver_LockServer_proto_rawDescData = file_pkg_lockserver_LockServer_proto_rawDesc
)

func file_pkg_lockserver_LockServer_proto_rawDescGZIP() []byte {
	file_pkg_lockserver_LockServer_proto_rawDescOnce.Do(func() {
		file_pkg_lockserver_LockServer_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_lockserver_LockServer_proto_rawDescData)
	})
	return file_pkg_lockserver_LockServer_proto_rawDescData
}

var file_pkg_lockserver_LockServer_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_pkg_lockserver_LockServer_proto_goTypes = []interface{}{
	(*AcquireLocksInfo)(nil), // 0: lockserver.AcquireLocksInfo
	(*ReleaseLocksInfo)(nil), // 1: lockserver.ReleaseLocksInfo
	(*Success)(nil),          // 2: lockserver.Success
}
var file_pkg_lockserver_LockServer_proto_depIdxs = []int32{
	0, // 0: lockserver.LockService.Acquire:input_type -> lockserver.AcquireLocksInfo
	1, // 1: lockserver.LockService.Release:input_type -> lockserver.ReleaseLocksInfo
	2, // 2: lockserver.LockService.Acquire:output_type -> lockserver.Success
	2, // 3: lockserver.LockService.Release:output_type -> lockserver.Success
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_lockserver_LockServer_proto_init() }
func file_pkg_lockserver_LockServer_proto_init() {
	if File_pkg_lockserver_LockServer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_lockserver_LockServer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AcquireLocksInfo); i {
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
		file_pkg_lockserver_LockServer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReleaseLocksInfo); i {
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
		file_pkg_lockserver_LockServer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Success); i {
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
			RawDescriptor: file_pkg_lockserver_LockServer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_lockserver_LockServer_proto_goTypes,
		DependencyIndexes: file_pkg_lockserver_LockServer_proto_depIdxs,
		MessageInfos:      file_pkg_lockserver_LockServer_proto_msgTypes,
	}.Build()
	File_pkg_lockserver_LockServer_proto = out.File
	file_pkg_lockserver_LockServer_proto_rawDesc = nil
	file_pkg_lockserver_LockServer_proto_goTypes = nil
	file_pkg_lockserver_LockServer_proto_depIdxs = nil
}
