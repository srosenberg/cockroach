package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(561396)

	var x [1]struct{}
	_ = x[PGNumericPos-0]
	_ = x[PGNumericNeg-16384]
}

const (
	_PGNumericSign_name_0 = "PGNumericPos"
	_PGNumericSign_name_1 = "PGNumericNeg"
)

func (i PGNumericSign) String() string {
	__antithesis_instrumentation__.Notify(561397)
	switch {
	case i == 0:
		__antithesis_instrumentation__.Notify(561398)
		return _PGNumericSign_name_0
	case i == 16384:
		__antithesis_instrumentation__.Notify(561399)
		return _PGNumericSign_name_1
	default:
		__antithesis_instrumentation__.Notify(561400)
		return "PGNumericSign(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
