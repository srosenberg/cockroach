package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Delete struct {
	With      *With
	Table     TableExpr
	Where     *Where
	OrderBy   OrderBy
	Limit     *Limit
	Returning ReturningClause
}

func (node *Delete) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607602)
	ctx.FormatNode(node.With)
	ctx.WriteString("DELETE FROM ")
	ctx.FormatNode(node.Table)
	if node.Where != nil {
		__antithesis_instrumentation__.Notify(607606)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Where)
	} else {
		__antithesis_instrumentation__.Notify(607607)
	}
	__antithesis_instrumentation__.Notify(607603)
	if len(node.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(607608)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.OrderBy)
	} else {
		__antithesis_instrumentation__.Notify(607609)
	}
	__antithesis_instrumentation__.Notify(607604)
	if node.Limit != nil {
		__antithesis_instrumentation__.Notify(607610)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Limit)
	} else {
		__antithesis_instrumentation__.Notify(607611)
	}
	__antithesis_instrumentation__.Notify(607605)
	if HasReturningClause(node.Returning) {
		__antithesis_instrumentation__.Notify(607612)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Returning)
	} else {
		__antithesis_instrumentation__.Notify(607613)
	}
}
