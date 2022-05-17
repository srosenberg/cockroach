package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(561406)

	var x [1]struct{}
	_ = x[ServerErrFieldSeverity-83]
	_ = x[ServerErrFieldSQLState-67]
	_ = x[ServerErrFieldMsgPrimary-77]
	_ = x[ServerErrFieldDetail-68]
	_ = x[ServerErrFieldHint-72]
	_ = x[ServerErrFieldSrcFile-70]
	_ = x[ServerErrFieldSrcLine-76]
	_ = x[ServerErrFieldSrcFunction-82]
	_ = x[ServerErrFieldConstraintName-110]
}

const (
	_ServerErrFieldType_name_0 = "ServerErrFieldSQLStateServerErrFieldDetail"
	_ServerErrFieldType_name_1 = "ServerErrFieldSrcFile"
	_ServerErrFieldType_name_2 = "ServerErrFieldHint"
	_ServerErrFieldType_name_3 = "ServerErrFieldSrcLineServerErrFieldMsgPrimary"
	_ServerErrFieldType_name_4 = "ServerErrFieldSrcFunctionServerErrFieldSeverity"
	_ServerErrFieldType_name_5 = "ServerErrFieldConstraintName"
)

var (
	_ServerErrFieldType_index_0 = [...]uint8{0, 22, 42}
	_ServerErrFieldType_index_3 = [...]uint8{0, 21, 45}
	_ServerErrFieldType_index_4 = [...]uint8{0, 25, 47}
)

func (i ServerErrFieldType) String() string {
	__antithesis_instrumentation__.Notify(561407)
	switch {
	case 67 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561415)
		return i <= 68 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561408)
		i -= 67
		return _ServerErrFieldType_name_0[_ServerErrFieldType_index_0[i]:_ServerErrFieldType_index_0[i+1]]
	case i == 70:
		__antithesis_instrumentation__.Notify(561409)
		return _ServerErrFieldType_name_1
	case i == 72:
		__antithesis_instrumentation__.Notify(561410)
		return _ServerErrFieldType_name_2
	case 76 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561416)
		return i <= 77 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561411)
		i -= 76
		return _ServerErrFieldType_name_3[_ServerErrFieldType_index_3[i]:_ServerErrFieldType_index_3[i+1]]
	case 82 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561417)
		return i <= 83 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561412)
		i -= 82
		return _ServerErrFieldType_name_4[_ServerErrFieldType_index_4[i]:_ServerErrFieldType_index_4[i+1]]
	case i == 110:
		__antithesis_instrumentation__.Notify(561413)
		return _ServerErrFieldType_name_5
	default:
		__antithesis_instrumentation__.Notify(561414)
		return "ServerErrFieldType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
