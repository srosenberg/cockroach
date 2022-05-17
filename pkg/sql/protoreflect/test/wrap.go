package protoreflecttest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"

	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/gogo/protobuf/jsonpb"
)

const SecretMessage = "secret message"

const RedactedMessage = "nothing to see here"

func (m Inner) MarshalJSONPB(marshaller *jsonpb.Marshaler) ([]byte, error) {
	__antithesis_instrumentation__.Notify(563639)
	if protoreflect.ShouldRedact(marshaller) && func() bool {
		__antithesis_instrumentation__.Notify(563641)
		return m.Value == SecretMessage == true
	}() == true {
		__antithesis_instrumentation__.Notify(563642)
		m.Value = RedactedMessage
	} else {
		__antithesis_instrumentation__.Notify(563643)
	}
	__antithesis_instrumentation__.Notify(563640)
	return json.Marshal(m)
}
