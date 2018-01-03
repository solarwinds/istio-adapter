// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: mixer/adapter/appoptics/config/config.proto

/*
	Package config is a generated protocol buffer package.

	It is generated from these files:
		mixer/adapter/appoptics/config/config.proto

	It has these top-level messages:
		Params
*/
package config

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Params struct {
	AppopticsAccessToken     string `protobuf:"bytes,1,opt,name=appoptics_access_token,json=appopticsAccessToken,proto3" json:"appoptics_access_token,omitempty"`
	PapertrailUrl            string `protobuf:"bytes,2,opt,name=papertrail_url,json=papertrailUrl,proto3" json:"papertrail_url,omitempty"`
	PapertrailLocalRetention string `protobuf:"bytes,3,opt,name=papertrail_local_retention,json=papertrailLocalRetention,proto3" json:"papertrail_local_retention,omitempty"`
	// The set of metrics to represent in Prometheus. If a metric is defined in Istio but doesn't have a corresponding
	// shape here, it will not be populated at runtime.
	Metrics []*Params_MetricInfo `protobuf:"bytes,4,rep,name=metrics" json:"metrics,omitempty"`
	Logs    []*Params_LogInfo    `protobuf:"bytes,5,rep,name=logs" json:"logs,omitempty"`
}

func (m *Params) Reset()                    { *m = Params{} }
func (*Params) ProtoMessage()               {}
func (*Params) Descriptor() ([]byte, []int) { return fileDescriptorConfig, []int{0} }

// Describes how a metric should be represented in Prometheus.
type Params_MetricInfo struct {
	// Recommended. The name is used to register the prometheus metric.
	// It must be unique across all prometheus metrics as prometheus does not allow duplicate names.
	// If name is not specified a sanitized version of instance_name is used.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Required. The name is the fully qualified name of the Istio metric instance
	// that this MetricInfo processes.
	InstanceName string `protobuf:"bytes,2,opt,name=instance_name,json=instanceName,proto3" json:"instance_name,omitempty"`
	// Optional. A human readable description of this metric.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// The names of labels to use: these need to match the dimensions of the Istio metric.
	// TODO: see if we can remove this and rely on only the dimensions in the future.
	LabelNames []string `protobuf:"bytes,4,rep,name=label_names,json=labelNames" json:"label_names,omitempty"`
}

func (m *Params_MetricInfo) Reset()                    { *m = Params_MetricInfo{} }
func (*Params_MetricInfo) ProtoMessage()               {}
func (*Params_MetricInfo) Descriptor() ([]byte, []int) { return fileDescriptorConfig, []int{0, 0} }

type Params_LogInfo struct {
	// The logging template provides a set of variables; these list the subset of variables that should be used to
	// form Stackdriver labels for the log entry.
	LabelNames []string `protobuf:"bytes,1,rep,name=label_names,json=labelNames" json:"label_names,omitempty"`
	// A golang text/template template that will be executed to construct the payload for this log entry.
	// It will be given the full set of variables for the log to use to construct its result.
	PayloadTemplate string `protobuf:"bytes,2,opt,name=payload_template,json=payloadTemplate,proto3" json:"payload_template,omitempty"`
	InstanceName    string `protobuf:"bytes,3,opt,name=instance_name,json=instanceName,proto3" json:"instance_name,omitempty"`
}

func (m *Params_LogInfo) Reset()                    { *m = Params_LogInfo{} }
func (*Params_LogInfo) ProtoMessage()               {}
func (*Params_LogInfo) Descriptor() ([]byte, []int) { return fileDescriptorConfig, []int{0, 1} }

func init() {
	proto.RegisterType((*Params)(nil), "adapter.appoptics.config.Params")
	proto.RegisterType((*Params_MetricInfo)(nil), "adapter.appoptics.config.Params.MetricInfo")
	proto.RegisterType((*Params_LogInfo)(nil), "adapter.appoptics.config.Params.LogInfo")
}
func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.AppopticsAccessToken) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.AppopticsAccessToken)))
		i += copy(dAtA[i:], m.AppopticsAccessToken)
	}
	if len(m.PapertrailUrl) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.PapertrailUrl)))
		i += copy(dAtA[i:], m.PapertrailUrl)
	}
	if len(m.PapertrailLocalRetention) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.PapertrailLocalRetention)))
		i += copy(dAtA[i:], m.PapertrailLocalRetention)
	}
	if len(m.Metrics) > 0 {
		for _, msg := range m.Metrics {
			dAtA[i] = 0x22
			i++
			i = encodeVarintConfig(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Logs) > 0 {
		for _, msg := range m.Logs {
			dAtA[i] = 0x2a
			i++
			i = encodeVarintConfig(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *Params_MetricInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params_MetricInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.InstanceName) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.InstanceName)))
		i += copy(dAtA[i:], m.InstanceName)
	}
	if len(m.Description) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.Description)))
		i += copy(dAtA[i:], m.Description)
	}
	if len(m.LabelNames) > 0 {
		for _, s := range m.LabelNames {
			dAtA[i] = 0x22
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	return i, nil
}

func (m *Params_LogInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params_LogInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.LabelNames) > 0 {
		for _, s := range m.LabelNames {
			dAtA[i] = 0xa
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if len(m.PayloadTemplate) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.PayloadTemplate)))
		i += copy(dAtA[i:], m.PayloadTemplate)
	}
	if len(m.InstanceName) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintConfig(dAtA, i, uint64(len(m.InstanceName)))
		i += copy(dAtA[i:], m.InstanceName)
	}
	return i, nil
}

func encodeVarintConfig(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Params) Size() (n int) {
	var l int
	_ = l
	l = len(m.AppopticsAccessToken)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.PapertrailUrl)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.PapertrailLocalRetention)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	if len(m.Metrics) > 0 {
		for _, e := range m.Metrics {
			l = e.Size()
			n += 1 + l + sovConfig(uint64(l))
		}
	}
	if len(m.Logs) > 0 {
		for _, e := range m.Logs {
			l = e.Size()
			n += 1 + l + sovConfig(uint64(l))
		}
	}
	return n
}

func (m *Params_MetricInfo) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.InstanceName)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	if len(m.LabelNames) > 0 {
		for _, s := range m.LabelNames {
			l = len(s)
			n += 1 + l + sovConfig(uint64(l))
		}
	}
	return n
}

func (m *Params_LogInfo) Size() (n int) {
	var l int
	_ = l
	if len(m.LabelNames) > 0 {
		for _, s := range m.LabelNames {
			l = len(s)
			n += 1 + l + sovConfig(uint64(l))
		}
	}
	l = len(m.PayloadTemplate)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.InstanceName)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	return n
}

func sovConfig(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozConfig(x uint64) (n int) {
	return sovConfig(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Params) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Params{`,
		`AppopticsAccessToken:` + fmt.Sprintf("%v", this.AppopticsAccessToken) + `,`,
		`PapertrailUrl:` + fmt.Sprintf("%v", this.PapertrailUrl) + `,`,
		`PapertrailLocalRetention:` + fmt.Sprintf("%v", this.PapertrailLocalRetention) + `,`,
		`Metrics:` + strings.Replace(fmt.Sprintf("%v", this.Metrics), "Params_MetricInfo", "Params_MetricInfo", 1) + `,`,
		`Logs:` + strings.Replace(fmt.Sprintf("%v", this.Logs), "Params_LogInfo", "Params_LogInfo", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Params_MetricInfo) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Params_MetricInfo{`,
		`Name:` + fmt.Sprintf("%v", this.Name) + `,`,
		`InstanceName:` + fmt.Sprintf("%v", this.InstanceName) + `,`,
		`Description:` + fmt.Sprintf("%v", this.Description) + `,`,
		`LabelNames:` + fmt.Sprintf("%v", this.LabelNames) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Params_LogInfo) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Params_LogInfo{`,
		`LabelNames:` + fmt.Sprintf("%v", this.LabelNames) + `,`,
		`PayloadTemplate:` + fmt.Sprintf("%v", this.PayloadTemplate) + `,`,
		`InstanceName:` + fmt.Sprintf("%v", this.InstanceName) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringConfig(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppopticsAccessToken", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppopticsAccessToken = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PapertrailUrl", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PapertrailUrl = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PapertrailLocalRetention", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PapertrailLocalRetention = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metrics", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, &Params_MetricInfo{})
			if err := m.Metrics[len(m.Metrics)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Logs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Logs = append(m.Logs, &Params_LogInfo{})
			if err := m.Logs[len(m.Logs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthConfig
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Params_MetricInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MetricInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InstanceName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InstanceName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LabelNames", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LabelNames = append(m.LabelNames, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthConfig
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Params_LogInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LogInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LogInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LabelNames", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LabelNames = append(m.LabelNames, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PayloadTemplate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PayloadTemplate = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InstanceName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InstanceName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthConfig
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipConfig(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
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
					return 0, ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
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
					return 0, ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthConfig
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowConfig
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
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
				next, err := skipConfig(dAtA[start:])
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
	ErrInvalidLengthConfig = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConfig   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("mixer/adapter/appoptics/config/config.proto", fileDescriptorConfig) }

var fileDescriptorConfig = []byte{
	// 420 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x31, 0x8b, 0x14, 0x31,
	0x14, 0xc7, 0x27, 0xee, 0xb8, 0xc7, 0xbd, 0xf5, 0x54, 0xc2, 0x21, 0xc3, 0x14, 0x71, 0x51, 0x84,
	0x95, 0x83, 0x59, 0x50, 0x0b, 0x8b, 0x6b, 0x14, 0x2c, 0x84, 0x53, 0x64, 0x38, 0x1b, 0x9b, 0x21,
	0x97, 0xcd, 0x0d, 0xc1, 0x4c, 0x12, 0x92, 0x08, 0x6a, 0x65, 0x69, 0xe9, 0xc7, 0xb0, 0xf3, 0x6b,
	0x5c, 0xb9, 0xa5, 0xa5, 0x33, 0x36, 0x96, 0xfb, 0x11, 0x64, 0x32, 0x33, 0xbb, 0xcb, 0x2e, 0x62,
	0x95, 0xc7, 0xff, 0xff, 0x7b, 0x7f, 0xde, 0x7b, 0x04, 0x4e, 0x2a, 0xf1, 0x91, 0xdb, 0x39, 0x5d,
	0x50, 0xe3, 0xdb, 0xd7, 0x18, 0x6d, 0xbc, 0x60, 0x6e, 0xce, 0xb4, 0xba, 0x14, 0x65, 0xff, 0x64,
	0xc6, 0x6a, 0xaf, 0x71, 0xd2, 0x63, 0xd9, 0x1a, 0xcb, 0x3a, 0x3f, 0x3d, 0x2e, 0x75, 0xa9, 0x03,
	0x34, 0x6f, 0xab, 0x8e, 0xbf, 0xf7, 0x23, 0x86, 0xf1, 0x1b, 0x6a, 0x69, 0xe5, 0xf0, 0x13, 0xb8,
	0xb3, 0x6e, 0x2a, 0x28, 0x63, 0xdc, 0xb9, 0xc2, 0xeb, 0xf7, 0x5c, 0x25, 0x68, 0x8a, 0x66, 0x87,
	0xf9, 0xf1, 0xda, 0x7d, 0x16, 0xcc, 0xf3, 0xd6, 0xc3, 0x0f, 0xe0, 0xa6, 0xa1, 0x86, 0x5b, 0x6f,
	0xa9, 0x90, 0xc5, 0x07, 0x2b, 0x93, 0x6b, 0x81, 0x3e, 0xda, 0xa8, 0x6f, 0xad, 0xc4, 0xa7, 0x90,
	0x6e, 0x61, 0x52, 0x33, 0x2a, 0x0b, 0xcb, 0x3d, 0x57, 0x5e, 0x68, 0x95, 0x8c, 0x42, 0x4b, 0xb2,
	0x21, 0xce, 0x5a, 0x20, 0x1f, 0x7c, 0xfc, 0x02, 0x0e, 0x2a, 0xee, 0xad, 0x60, 0x2e, 0x89, 0xa7,
	0xa3, 0xd9, 0xe4, 0xd1, 0x49, 0xf6, 0xaf, 0x3d, 0xb3, 0x6e, 0x9b, 0xec, 0x55, 0xe0, 0x5f, 0xaa,
	0x4b, 0x9d, 0x0f, 0xbd, 0xf8, 0x14, 0x62, 0xa9, 0x4b, 0x97, 0x5c, 0x0f, 0x19, 0xb3, 0xff, 0x66,
	0x9c, 0xe9, 0x32, 0x04, 0x84, 0xae, 0xf4, 0x2b, 0x02, 0xd8, 0xa4, 0x62, 0x0c, 0xb1, 0xa2, 0x15,
	0xef, 0x8f, 0x13, 0x6a, 0x7c, 0x1f, 0x8e, 0x84, 0x72, 0x9e, 0x2a, 0xc6, 0x8b, 0x60, 0x76, 0xb7,
	0xb8, 0x31, 0x88, 0xaf, 0x5b, 0x68, 0x0a, 0x93, 0x05, 0x77, 0xcc, 0x0a, 0xb3, 0xb5, 0xfb, 0xb6,
	0x84, 0xef, 0xc2, 0x44, 0xd2, 0x0b, 0x2e, 0x43, 0x46, 0xb7, 0xf2, 0x61, 0x0e, 0x41, 0x6a, 0x13,
	0x5c, 0xfa, 0x19, 0x0e, 0xfa, 0xd9, 0x76, 0x59, 0xb4, 0xcb, 0xe2, 0x87, 0x70, 0xdb, 0xd0, 0x4f,
	0x52, 0xd3, 0x45, 0xe1, 0x79, 0x65, 0x24, 0xf5, 0xc3, 0x58, 0xb7, 0x7a, 0xfd, 0xbc, 0x97, 0xf7,
	0xc7, 0x1f, 0xed, 0x8f, 0xff, 0xfc, 0xe9, 0x55, 0x4d, 0xa2, 0x65, 0x4d, 0xa2, 0x9f, 0x35, 0x89,
	0x56, 0x35, 0x89, 0xbe, 0x34, 0x04, 0x7d, 0x6f, 0x48, 0x74, 0xd5, 0x10, 0xb4, 0x6c, 0x08, 0xfa,
	0xd5, 0x10, 0xf4, 0xa7, 0x21, 0xd1, 0xaa, 0x21, 0xe8, 0xdb, 0x6f, 0x12, 0xbd, 0x1b, 0x77, 0x57,
	0xbd, 0x18, 0x87, 0x2f, 0xf7, 0xf8, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc8, 0xab, 0xc3, 0x65,
	0xd1, 0x02, 0x00, 0x00,
}
