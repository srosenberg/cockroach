package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(250653)

	var x [1]struct{}
	_ = x[InitialVersion-0]
	_ = x[OwnerVersion-1]
	_ = x[Version21_2-2]
}

const _PrivilegeDescVersion_name = "InitialVersionOwnerVersionVersion21_2"

var _PrivilegeDescVersion_index = [...]uint8{0, 14, 26, 37}

func (i PrivilegeDescVersion) String() string {
	__antithesis_instrumentation__.Notify(250654)
	if i >= PrivilegeDescVersion(len(_PrivilegeDescVersion_index)-1) {
		__antithesis_instrumentation__.Notify(250656)
		return "PrivilegeDescVersion(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(250657)
	}
	__antithesis_instrumentation__.Notify(250655)
	return _PrivilegeDescVersion_name[_PrivilegeDescVersion_index[i]:_PrivilegeDescVersion_index[i+1]]
}
