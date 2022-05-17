package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(578288)

	var x [1]struct{}
	_ = x[ColumnConversionDangerous - -1]
	_ = x[ColumnConversionImpossible-0]
	_ = x[ColumnConversionTrivial-1]
	_ = x[ColumnConversionValidate-2]
	_ = x[ColumnConversionGeneral-3]
}

const _ColumnConversionKind_name = "DangerousImpossibleTrivialValidateGeneral"

var _ColumnConversionKind_index = [...]uint8{0, 9, 19, 26, 34, 41}

func (i ColumnConversionKind) String() string {
	__antithesis_instrumentation__.Notify(578289)
	i -= -1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(578291)
		return i >= ColumnConversionKind(len(_ColumnConversionKind_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(578292)
		return "ColumnConversionKind(" + strconv.FormatInt(int64(i+-1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(578293)
	}
	__antithesis_instrumentation__.Notify(578290)
	return _ColumnConversionKind_name[_ColumnConversionKind_index[i]:_ColumnConversionKind_index[i+1]]
}
