package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/gogo/protobuf/proto"

	gproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc/encoding"
)

const name = "proto"

type codec struct{}

var _ encoding.Codec = codec{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	__antithesis_instrumentation__.Notify(184276)
	if pm, ok := v.(proto.Marshaler); ok {
		__antithesis_instrumentation__.Notify(184278)
		return pm.Marshal()
	} else {
		__antithesis_instrumentation__.Notify(184279)
	}
	__antithesis_instrumentation__.Notify(184277)
	return gproto.Marshal(v.(gproto.Message))
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	__antithesis_instrumentation__.Notify(184280)
	if pm, ok := v.(proto.Unmarshaler); ok {
		__antithesis_instrumentation__.Notify(184282)
		return pm.Unmarshal(data)
	} else {
		__antithesis_instrumentation__.Notify(184283)
	}
	__antithesis_instrumentation__.Notify(184281)
	return gproto.Unmarshal(data, v.(gproto.Message))
}

func (codec) Name() string {
	__antithesis_instrumentation__.Notify(184284)
	return name
}
