package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(85206)

	var x [1]struct{}
	_ = x[DatabaseCommentType-0]
	_ = x[TableCommentType-1]
	_ = x[ColumnCommentType-2]
	_ = x[IndexCommentType-3]
	_ = x[SchemaCommentType-4]
	_ = x[ConstraintCommentType-5]
}

const _CommentType_name = "DatabaseCommentTypeTableCommentTypeColumnCommentTypeIndexCommentTypeSchemaCommentTypeConstraintCommentType"

var _CommentType_index = [...]uint8{0, 19, 35, 52, 68, 85, 106}

func (i CommentType) String() string {
	__antithesis_instrumentation__.Notify(85207)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(85209)
		return i >= CommentType(len(_CommentType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85210)
		return "CommentType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(85211)
	}
	__antithesis_instrumentation__.Notify(85208)
	return _CommentType_name[_CommentType_index[i]:_CommentType_index[i+1]]
}
