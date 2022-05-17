package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/redact"
)

func (x ForeignKeyAction) String() string {
	__antithesis_instrumentation__.Notify(249308)
	switch x {
	case ForeignKeyAction_RESTRICT:
		__antithesis_instrumentation__.Notify(249309)
		return "RESTRICT"
	case ForeignKeyAction_SET_DEFAULT:
		__antithesis_instrumentation__.Notify(249310)
		return "SET DEFAULT"
	case ForeignKeyAction_SET_NULL:
		__antithesis_instrumentation__.Notify(249311)
		return "SET NULL"
	case ForeignKeyAction_CASCADE:
		__antithesis_instrumentation__.Notify(249312)
		return "CASCADE"
	default:
		__antithesis_instrumentation__.Notify(249313)
		return strconv.Itoa(int(x))
	}
}

var _ redact.SafeValue = ForeignKeyAction(0)

func (x ForeignKeyAction) SafeValue() { __antithesis_instrumentation__.Notify(249314) }
