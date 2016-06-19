// Code generated by protoc-gen-gogo.
// source: services/svcmetadata/service_desc.proto
// DO NOT EDIT!

/*
	Package svcmetadata is a generated protocol buffer package.

	It is generated from these files:
		services/svcmetadata/service_desc.proto

	It has these top-level messages:
		ServiceDescription
*/
package queue_info

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.GoGoProtoPackageIsVersion1

type ServiceDescription struct {
	ExportId  uint64 `protobuf:"varint,1,req,name=export_id,json=exportId" json:"export_id"`
	SType     string `protobuf:"bytes,2,req,name=s_type,json=sType" json:"s_type"`
	Name      string `protobuf:"bytes,3,req,name=name" json:"name"`
	CreateTs  int64  `protobuf:"varint,4,req,name=create_ts,json=createTs" json:"create_ts"`
	Disabled  bool   `protobuf:"varint,5,req,name=disabled" json:"disabled"`
	ToDelete  bool   `protobuf:"varint,6,req,name=to_delete,json=toDelete" json:"to_delete"`
	ServiceId string `protobuf:"bytes,7,req,name=service_id,json=serviceId" json:"service_id"`
}

func (m *ServiceDescription) Reset()                    { *m = ServiceDescription{} }
func (*ServiceDescription) ProtoMessage()               {}
func (*ServiceDescription) Descriptor() ([]byte, []int) { return fileDescriptorServiceDesc, []int{0} }

func (m *ServiceDescription) GetExportId() uint64 {
	if m != nil {
		return m.ExportId
	}
	return 0
}

func (m *ServiceDescription) GetSType() string {
	if m != nil {
		return m.SType
	}
	return ""
}

func (m *ServiceDescription) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ServiceDescription) GetCreateTs() int64 {
	if m != nil {
		return m.CreateTs
	}
	return 0
}

func (m *ServiceDescription) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func (m *ServiceDescription) GetToDelete() bool {
	if m != nil {
		return m.ToDelete
	}
	return false
}

func (m *ServiceDescription) GetServiceId() string {
	if m != nil {
		return m.ServiceId
	}
	return ""
}

func init() {
	proto.RegisterType((*ServiceDescription)(nil), "svcmetadata.ServiceDescription")
}
func (this *ServiceDescription) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*ServiceDescription)
	if !ok {
		that2, ok := that.(ServiceDescription)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.ExportId != that1.ExportId {
		return false
	}
	if this.SType != that1.SType {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.CreateTs != that1.CreateTs {
		return false
	}
	if this.Disabled != that1.Disabled {
		return false
	}
	if this.ToDelete != that1.ToDelete {
		return false
	}
	if this.ServiceId != that1.ServiceId {
		return false
	}
	return true
}
func (this *ServiceDescription) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 11)
	s = append(s, "&svcmetadata.ServiceDescription{")
	s = append(s, "ExportId: "+fmt.Sprintf("%#v", this.ExportId)+",\n")
	s = append(s, "SType: "+fmt.Sprintf("%#v", this.SType)+",\n")
	s = append(s, "Name: "+fmt.Sprintf("%#v", this.Name)+",\n")
	s = append(s, "CreateTs: "+fmt.Sprintf("%#v", this.CreateTs)+",\n")
	s = append(s, "Disabled: "+fmt.Sprintf("%#v", this.Disabled)+",\n")
	s = append(s, "ToDelete: "+fmt.Sprintf("%#v", this.ToDelete)+",\n")
	s = append(s, "ServiceId: "+fmt.Sprintf("%#v", this.ServiceId)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringServiceDesc(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringServiceDesc(e map[int32]github_com_gogo_protobuf_proto.Extension) string {
	if e == nil {
		return "nil"
	}
	s := "map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "}"
	return s
}
func (m *ServiceDescription) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *ServiceDescription) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0x8
	i++
	i = encodeVarintServiceDesc(data, i, uint64(m.ExportId))
	data[i] = 0x12
	i++
	i = encodeVarintServiceDesc(data, i, uint64(len(m.SType)))
	i += copy(data[i:], m.SType)
	data[i] = 0x1a
	i++
	i = encodeVarintServiceDesc(data, i, uint64(len(m.Name)))
	i += copy(data[i:], m.Name)
	data[i] = 0x20
	i++
	i = encodeVarintServiceDesc(data, i, uint64(m.CreateTs))
	data[i] = 0x28
	i++
	if m.Disabled {
		data[i] = 1
	} else {
		data[i] = 0
	}
	i++
	data[i] = 0x30
	i++
	if m.ToDelete {
		data[i] = 1
	} else {
		data[i] = 0
	}
	i++
	data[i] = 0x3a
	i++
	i = encodeVarintServiceDesc(data, i, uint64(len(m.ServiceId)))
	i += copy(data[i:], m.ServiceId)
	return i, nil
}

func encodeFixed64ServiceDesc(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32ServiceDesc(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintServiceDesc(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *ServiceDescription) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovServiceDesc(uint64(m.ExportId))
	l = len(m.SType)
	n += 1 + l + sovServiceDesc(uint64(l))
	l = len(m.Name)
	n += 1 + l + sovServiceDesc(uint64(l))
	n += 1 + sovServiceDesc(uint64(m.CreateTs))
	n += 2
	n += 2
	l = len(m.ServiceId)
	n += 1 + l + sovServiceDesc(uint64(l))
	return n
}

func sovServiceDesc(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozServiceDesc(x uint64) (n int) {
	return sovServiceDesc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *ServiceDescription) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ServiceDescription{`,
		`ExportId:` + fmt.Sprintf("%v", this.ExportId) + `,`,
		`SType:` + fmt.Sprintf("%v", this.SType) + `,`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`CreateTs:` + fmt.Sprintf("%v", this.CreateTs) + `,`,
		`Disabled:` + fmt.Sprintf("%v", this.Disabled) + `,`,
		`ToDelete:` + fmt.Sprintf("%v", this.ToDelete) + `,`,
		`ServiceId:` + fmt.Sprintf("%v", this.ServiceId) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringServiceDesc(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *ServiceDescription) Unmarshal(data []byte) error {
	var hasFields [1]uint64
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowServiceDesc
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ServiceDescription: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ServiceDescription: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExportId", wireType)
			}
			m.ExportId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.ExportId |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthServiceDesc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SType = string(data[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthServiceDesc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(data[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000004)
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreateTs", wireType)
			}
			m.CreateTs = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.CreateTs |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000008)
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Disabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Disabled = bool(v != 0)
			hasFields[0] |= uint64(0x00000010)
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ToDelete", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.ToDelete = bool(v != 0)
			hasFields[0] |= uint64(0x00000020)
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServiceId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthServiceDesc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ServiceId = string(data[iNdEx:postIndex])
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000040)
		default:
			iNdEx = preIndex
			skippy, err := skipServiceDesc(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthServiceDesc
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("export_id")
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("s_type")
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("name")
	}
	if hasFields[0]&uint64(0x00000008) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("create_ts")
	}
	if hasFields[0]&uint64(0x00000010) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("disabled")
	}
	if hasFields[0]&uint64(0x00000020) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("to_delete")
	}
	if hasFields[0]&uint64(0x00000040) == 0 {
		return github_com_gogo_protobuf_proto.NewRequiredNotSetError("service_id")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipServiceDesc(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowServiceDesc
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowServiceDesc
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthServiceDesc
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowServiceDesc
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipServiceDesc(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthServiceDesc = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowServiceDesc   = fmt.Errorf("proto: integer overflow")
)

var fileDescriptorServiceDesc = []byte{
	// 262 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x4c, 0x8f, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x93, 0x36, 0x2d, 0xc9, 0xb1, 0x79, 0xb2, 0x84, 0x14, 0x0a, 0x0c, 0x30, 0x20, 0xfa,
	0x0e, 0x55, 0x97, 0xae, 0xd0, 0x3d, 0x32, 0xf1, 0x0d, 0x96, 0xda, 0x3a, 0xf2, 0x9d, 0x2a, 0xd8,
	0x78, 0x04, 0x1e, 0x83, 0x47, 0xe9, 0xd8, 0x91, 0x09, 0xd1, 0xb2, 0x30, 0xb2, 0xb3, 0xf4, 0xda,
	0x24, 0x52, 0x86, 0x5f, 0x96, 0xbe, 0xef, 0xd7, 0xc9, 0x3f, 0xdc, 0x12, 0x86, 0xb5, 0x2b, 0x91,
	0xc6, 0xb4, 0x2e, 0x97, 0xc8, 0xc6, 0x1a, 0x36, 0xe3, 0x06, 0x16, 0x16, 0xa9, 0x7c, 0xa8, 0x82,
	0x67, 0xaf, 0xce, 0x3b, 0xfe, 0xfa, 0x3f, 0x06, 0xf5, 0x54, 0x77, 0xa6, 0x52, 0x09, 0xae, 0x62,
	0xe7, 0x57, 0xea, 0x0a, 0x32, 0x7c, 0xa9, 0x7c, 0xe0, 0xc2, 0x59, 0x1d, 0x8f, 0x7a, 0x77, 0xc9,
	0x24, 0xd9, 0x7c, 0x5d, 0x46, 0x8f, 0x69, 0x8d, 0x67, 0x56, 0x5d, 0xc0, 0x90, 0x0a, 0x7e, 0xad,
	0x50, 0xf7, 0xc4, 0x67, 0x8d, 0x1f, 0xd0, 0x5c, 0x90, 0xd2, 0x90, 0xac, 0xcc, 0x12, 0x75, 0xbf,
	0xa3, 0x4e, 0xe4, 0x78, 0xb9, 0x0c, 0x68, 0x18, 0x0b, 0x26, 0x9d, 0x88, 0xee, 0xb7, 0x97, 0x6b,
	0x3c, 0x27, 0x35, 0x82, 0xd4, 0x3a, 0x32, 0xcf, 0x0b, 0xb4, 0x7a, 0x20, 0x8d, 0xb4, 0x6d, 0xb4,
	0xf4, 0x78, 0x84, 0xbd, 0x6c, 0x5a, 0x20, 0xa3, 0x1e, 0x76, 0x2b, 0xec, 0xa7, 0x27, 0xaa, 0x6e,
	0x00, 0xda, 0xed, 0x32, 0xe1, 0xac, 0xf3, 0x8f, 0xac, 0xe1, 0x33, 0x3b, 0xb9, 0xdf, 0xee, 0xf2,
	0xe8, 0x53, 0xf2, 0xb7, 0xcb, 0xe3, 0xb7, 0x7d, 0x1e, 0x7f, 0x48, 0x36, 0x92, 0xad, 0xe4, 0x5b,
	0xf2, 0xbb, 0x17, 0x27, 0xef, 0xfb, 0x4f, 0x1e, 0x1d, 0x02, 0x00, 0x00, 0xff, 0xff, 0x85, 0xae,
	0xd0, 0x08, 0x62, 0x01, 0x00, 0x00,
}
