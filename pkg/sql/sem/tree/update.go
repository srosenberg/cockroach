package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Update struct {
	With      *With
	Table     TableExpr
	Exprs     UpdateExprs
	From      TableExprs
	Where     *Where
	OrderBy   OrderBy
	Limit     *Limit
	Returning ReturningClause
}

func (node *Update) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615872)
	ctx.FormatNode(node.With)
	ctx.WriteString("UPDATE ")
	ctx.FormatNode(node.Table)
	ctx.WriteString(" SET ")
	ctx.FormatNode(&node.Exprs)
	if len(node.From) > 0 {
		__antithesis_instrumentation__.Notify(615877)
		ctx.WriteString(" FROM ")
		ctx.FormatNode(&node.From)
	} else {
		__antithesis_instrumentation__.Notify(615878)
	}
	__antithesis_instrumentation__.Notify(615873)
	if node.Where != nil {
		__antithesis_instrumentation__.Notify(615879)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Where)
	} else {
		__antithesis_instrumentation__.Notify(615880)
	}
	__antithesis_instrumentation__.Notify(615874)
	if len(node.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(615881)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.OrderBy)
	} else {
		__antithesis_instrumentation__.Notify(615882)
	}
	__antithesis_instrumentation__.Notify(615875)
	if node.Limit != nil {
		__antithesis_instrumentation__.Notify(615883)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Limit)
	} else {
		__antithesis_instrumentation__.Notify(615884)
	}
	__antithesis_instrumentation__.Notify(615876)
	if HasReturningClause(node.Returning) {
		__antithesis_instrumentation__.Notify(615885)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Returning)
	} else {
		__antithesis_instrumentation__.Notify(615886)
	}
}

type UpdateExprs []*UpdateExpr

func (node *UpdateExprs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615887)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(615888)
		if i > 0 {
			__antithesis_instrumentation__.Notify(615890)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(615891)
		}
		__antithesis_instrumentation__.Notify(615889)
		ctx.FormatNode(n)
	}
}

type UpdateExpr struct {
	Tuple bool
	Names NameList
	Expr  Expr
}

func (node *UpdateExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615892)
	open, close := "", ""
	if node.Tuple {
		__antithesis_instrumentation__.Notify(615894)
		open, close = "(", ")"
	} else {
		__antithesis_instrumentation__.Notify(615895)
	}
	__antithesis_instrumentation__.Notify(615893)
	ctx.WriteString(open)
	ctx.FormatNode(&node.Names)
	ctx.WriteString(close)
	ctx.WriteString(" = ")
	ctx.FormatNode(node.Expr)
}
