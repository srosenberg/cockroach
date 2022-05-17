package protoreflect

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"strings"

	jsonb "github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

type anyResolver struct{}

func (r *anyResolver) Resolve(typeURL string) (proto.Message, error) {
	__antithesis_instrumentation__.Notify(563356)

	mname := typeURL
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		__antithesis_instrumentation__.Notify(563359)
		mname = mname[slash+1:]
	} else {
		__antithesis_instrumentation__.Notify(563360)
	}
	__antithesis_instrumentation__.Notify(563357)
	mt := proto.MessageType(mname)
	if mt == nil {
		__antithesis_instrumentation__.Notify(563361)
		return nil, fmt.Errorf("unknown message type %q", mname)
	} else {
		__antithesis_instrumentation__.Notify(563362)
	}
	__antithesis_instrumentation__.Notify(563358)
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

func ShouldRedact(m *jsonpb.Marshaler) bool {
	__antithesis_instrumentation__.Notify(563363)
	_, shouldRedact := m.AnyResolver.(*anyResolver)
	return shouldRedact
}

var redactionJSONBMarker = func() jsonb.JSON {
	__antithesis_instrumentation__.Notify(563364)
	jb, err := jsonb.ParseJSON(`{"__redacted__": true}`)
	if err != nil {
		__antithesis_instrumentation__.Notify(563366)
		panic("unexpected error parsing redaction JSON")
	} else {
		__antithesis_instrumentation__.Notify(563367)
	}
	__antithesis_instrumentation__.Notify(563365)
	return jb
}()

func marshalToJSONRedacted(msg protoutil.Message, emitDefaults bool) (jsonb.JSON, error) {
	__antithesis_instrumentation__.Notify(563368)
	jsonEncoder := jsonpb.Marshaler{EmitDefaults: emitDefaults}
	jsonEncoder.AnyResolver = &anyResolver{}
	msgJSON, err := jsonEncoder.MarshalToString(msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(563371)
		return nil, errors.Wrapf(err, "marshaling %s; msg=%+v", proto.MessageName(msg), msg)
	} else {
		__antithesis_instrumentation__.Notify(563372)
	}
	__antithesis_instrumentation__.Notify(563369)
	jb, err := jsonb.ParseJSON(msgJSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(563373)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563374)
	}
	__antithesis_instrumentation__.Notify(563370)
	return jb.Concat(redactionJSONBMarker)
}
