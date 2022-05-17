package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

type UnionClause struct {
	Type        UnionType
	Left, Right *Select
	All         bool
}

type UnionType int

const (
	UnionOp UnionType = iota
	IntersectOp
	ExceptOp
)

var unionTypeName = [...]string{
	UnionOp:     "UNION",
	IntersectOp: "INTERSECT",
	ExceptOp:    "EXCEPT",
}

func (i UnionType) String() string {
	__antithesis_instrumentation__.Notify(615860)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(615862)
		return i > UnionType(len(unionTypeName)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(615863)
		return fmt.Sprintf("UnionType(%d)", i)
	} else {
		__antithesis_instrumentation__.Notify(615864)
	}
	__antithesis_instrumentation__.Notify(615861)
	return unionTypeName[i]
}

func (node *UnionClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615865)
	ctx.FormatNode(node.Left)
	ctx.WriteByte(' ')
	ctx.WriteString(node.Type.String())
	if node.All {
		__antithesis_instrumentation__.Notify(615867)
		ctx.WriteString(" ALL")
	} else {
		__antithesis_instrumentation__.Notify(615868)
	}
	__antithesis_instrumentation__.Notify(615866)
	ctx.WriteByte(' ')
	ctx.FormatNode(node.Right)
}
