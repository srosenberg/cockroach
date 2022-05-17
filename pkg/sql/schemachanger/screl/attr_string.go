package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(594990)

	var x [1]struct{}
	_ = x[DescID-1]
	_ = x[IndexID-2]
	_ = x[ColumnFamilyID-3]
	_ = x[ColumnID-4]
	_ = x[ConstraintID-5]
	_ = x[Name-6]
	_ = x[ReferencedDescID-7]
	_ = x[TargetStatus-8]
	_ = x[CurrentStatus-9]
	_ = x[Element-10]
	_ = x[Target-11]
}

const _Attr_name = "DescIDIndexIDColumnFamilyIDColumnIDConstraintIDNameReferencedDescIDTargetStatusCurrentStatusElementTarget"

var _Attr_index = [...]uint8{0, 6, 13, 27, 35, 47, 51, 67, 79, 92, 99, 105}

func (i Attr) String() string {
	__antithesis_instrumentation__.Notify(594991)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(594993)
		return i >= Attr(len(_Attr_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(594994)
		return "Attr(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(594995)
	}
	__antithesis_instrumentation__.Notify(594992)
	return _Attr_name[_Attr_index[i]:_Attr_index[i+1]]
}
