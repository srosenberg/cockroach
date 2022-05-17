package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

func BuiltinCounter(name, signature string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625820)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.builtins.%s%s", name, signature))
}

func UnaryOpCounter(op, typ string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625821)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.ops.un.%s %s", op, typ))
}

func CmpOpCounter(op, ltyp, rtyp string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625822)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.ops.cmp.%s %s %s", ltyp, op, rtyp))
}

func BinOpCounter(op, ltyp, rtyp string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625823)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.ops.bin.%s %s %s", ltyp, op, rtyp))
}

func CastOpCounter(ftyp, ttyp string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625824)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.ops.cast.%s::%s", ftyp, ttyp))
}

var ArrayCastCounter = telemetry.GetCounterOnce("sql.plan.ops.cast.arrays")

var TupleCastCounter = telemetry.GetCounterOnce("sql.plan.ops.cast.tuples")

var EnumCastCounter = telemetry.GetCounterOnce("sql.plan.ops.cast.enums")

var ArrayConstructorCounter = telemetry.GetCounterOnce("sql.plan.ops.array.cons")

var ArrayFlattenCounter = telemetry.GetCounterOnce("sql.plan.ops.array.flatten")

var ArraySubscriptCounter = telemetry.GetCounterOnce("sql.plan.ops.array.ind")

var IfErrCounter = telemetry.GetCounterOnce("sql.plan.ops.iferr")

var LargeLShiftArgumentCounter = telemetry.GetCounterOnce("sql.large_lshift_argument")

var LargeRShiftArgumentCounter = telemetry.GetCounterOnce("sql.large_rshift_argument")
