package protoreflect

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"
	"strings"

	jsonb "github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

var shorthands map[string]protoutil.Message = map[string]protoutil.Message{}

func RegisterShorthands(msg protoutil.Message, names ...string) {
	__antithesis_instrumentation__.Notify(563644)
	for _, name := range names {
		__antithesis_instrumentation__.Notify(563645)
		name = foldShorthand(name)
		if existing, ok := shorthands[name]; ok {
			__antithesis_instrumentation__.Notify(563647)
			panic(errors.AssertionFailedf("shorthand %s already registered to %T", name, existing))
		} else {
			__antithesis_instrumentation__.Notify(563648)
		}
		__antithesis_instrumentation__.Notify(563646)
		shorthands[name] = msg
	}
}

func foldShorthand(name string) string {
	__antithesis_instrumentation__.Notify(563649)
	return strings.ReplaceAll(strings.ToLower(name), "_", "")
}

func NewMessage(name string) (protoutil.Message, error) {
	__antithesis_instrumentation__.Notify(563650)

	rt := proto.MessageType(name)
	if rt == nil {
		__antithesis_instrumentation__.Notify(563654)
		if msg, ok := shorthands[foldShorthand(name)]; ok {
			__antithesis_instrumentation__.Notify(563655)
			fullName := proto.MessageName(msg)
			rt = proto.MessageType(fullName)
			if rt == nil {
				__antithesis_instrumentation__.Notify(563656)
				return nil, errors.Newf("unknown proto message type %s", fullName)
			} else {
				__antithesis_instrumentation__.Notify(563657)
			}
		} else {
			__antithesis_instrumentation__.Notify(563658)
			return nil, errors.Newf("unknown proto message type %s", name)
		}
	} else {
		__antithesis_instrumentation__.Notify(563659)
	}
	__antithesis_instrumentation__.Notify(563651)

	if rt.Kind() != reflect.Ptr {
		__antithesis_instrumentation__.Notify(563660)
		return nil, errors.AssertionFailedf(
			"expected ptr to message, got %s instead", rt.Kind().String())
	} else {
		__antithesis_instrumentation__.Notify(563661)
	}
	__antithesis_instrumentation__.Notify(563652)

	rv := reflect.New(rt.Elem())
	msg, ok := rv.Interface().(protoutil.Message)

	if !ok {
		__antithesis_instrumentation__.Notify(563662)

		return nil, errors.AssertionFailedf(
			"unexpected proto type for %s; expected protoutil.Message, got %T",
			name, rv.Interface())
	} else {
		__antithesis_instrumentation__.Notify(563663)
	}
	__antithesis_instrumentation__.Notify(563653)
	return msg, nil
}

func DecodeMessage(name string, data []byte) (protoutil.Message, error) {
	__antithesis_instrumentation__.Notify(563664)
	msg, err := NewMessage(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(563667)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563668)
	}
	__antithesis_instrumentation__.Notify(563665)

	if err := protoutil.Unmarshal(data, msg); err != nil {
		__antithesis_instrumentation__.Notify(563669)
		return nil, errors.Wrapf(err, "failed to unmarshal proto %s", name)
	} else {
		__antithesis_instrumentation__.Notify(563670)
	}
	__antithesis_instrumentation__.Notify(563666)
	return msg, nil
}

type FmtFlags struct {
	EmitDefaults bool

	EmitRedacted bool
}

func marshalToJSON(msg protoutil.Message, emitDefaults bool) (jsonb.JSON, error) {
	__antithesis_instrumentation__.Notify(563671)
	jsonEncoder := jsonpb.Marshaler{EmitDefaults: emitDefaults}
	msgJSON, err := jsonEncoder.MarshalToString(msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(563673)
		return nil, errors.Wrapf(err, "marshaling %s; msg=%+v", proto.MessageName(msg), msg)
	} else {
		__antithesis_instrumentation__.Notify(563674)
	}
	__antithesis_instrumentation__.Notify(563672)
	return jsonb.ParseJSON(msgJSON)
}

func MessageToJSON(msg protoutil.Message, flags FmtFlags) (jsonb.JSON, error) {
	__antithesis_instrumentation__.Notify(563675)
	if flags.EmitRedacted {
		__antithesis_instrumentation__.Notify(563677)
		return marshalToJSONRedacted(msg, flags.EmitDefaults)
	} else {
		__antithesis_instrumentation__.Notify(563678)
	}
	__antithesis_instrumentation__.Notify(563676)
	return marshalToJSON(msg, flags.EmitDefaults)
}

func JSONBMarshalToMessage(input jsonb.JSON, target protoutil.Message) ([]byte, error) {
	__antithesis_instrumentation__.Notify(563679)
	json := &jsonpb.Unmarshaler{}
	if err := json.Unmarshal(strings.NewReader(input.String()), target); err != nil {
		__antithesis_instrumentation__.Notify(563682)
		return nil, errors.Wrapf(err, "unmarshaling json to %s", proto.MessageName(target))
	} else {
		__antithesis_instrumentation__.Notify(563683)
	}
	__antithesis_instrumentation__.Notify(563680)
	data, err := protoutil.Marshal(target)
	if err != nil {
		__antithesis_instrumentation__.Notify(563684)
		return nil, errors.Wrapf(err, "marshaling to proto %s", proto.MessageName(target))
	} else {
		__antithesis_instrumentation__.Notify(563685)
	}
	__antithesis_instrumentation__.Notify(563681)
	return data, nil
}
