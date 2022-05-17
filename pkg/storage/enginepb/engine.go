package enginepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/redact"
)

func (e *EngineType) Type() string { __antithesis_instrumentation__.Notify(633872); return "string" }

func (e *EngineType) String() string {
	__antithesis_instrumentation__.Notify(633873)
	return redact.StringWithoutMarkers(e)
}

func (e *EngineType) SafeFormat(p redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(633874)
	switch *e {
	case EngineTypeDefault:
		__antithesis_instrumentation__.Notify(633875)
		p.SafeString("default")
	case EngineTypePebble:
		__antithesis_instrumentation__.Notify(633876)
		p.SafeString("pebble")
	default:
		__antithesis_instrumentation__.Notify(633877)
		p.Printf("<unknown engine %d>", int32(*e))
	}
}

func (e *EngineType) Set(s string) error {
	__antithesis_instrumentation__.Notify(633878)
	switch s {
	case "default":
		__antithesis_instrumentation__.Notify(633880)
		*e = EngineTypeDefault
	case "pebble":
		__antithesis_instrumentation__.Notify(633881)
		*e = EngineTypePebble
	default:
		__antithesis_instrumentation__.Notify(633882)
		return fmt.Errorf("invalid storage engine: %s "+
			"(possible values: pebble)", s)
	}
	__antithesis_instrumentation__.Notify(633879)
	return nil
}
