package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/redact"
	"github.com/cockroachdb/redact/interfaces"
)

func (v EncodedValue) String() string {
	__antithesis_instrumentation__.Notify(239684)
	return redact.Sprint(v).StripMarkers()
}

func (v EncodedValue) SafeFormat(s interfaces.SafePrinter, verb rune) {
	__antithesis_instrumentation__.Notify(239685)
	s.Printf("%q (%s)", v.Value, redact.SafeString(v.Type))
}

var _ redact.SafeFormatter = EncodedValue{}
