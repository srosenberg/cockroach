package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(251594)

	var x [1]struct{}
	_ = x[BaseFormatVersion-1]
	_ = x[FamilyFormatVersion-2]
	_ = x[InterleavedFormatVersion-3]
}

const _FormatVersion_name = "BaseFormatVersionFamilyFormatVersionInterleavedFormatVersion"

var _FormatVersion_index = [...]uint8{0, 17, 36, 60}

func (i FormatVersion) String() string {
	__antithesis_instrumentation__.Notify(251595)
	i -= 1
	if i >= FormatVersion(len(_FormatVersion_index)-1) {
		__antithesis_instrumentation__.Notify(251597)
		return "FormatVersion(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(251598)
	}
	__antithesis_instrumentation__.Notify(251596)
	return _FormatVersion_name[_FormatVersion_index[i]:_FormatVersion_index[i+1]]
}
