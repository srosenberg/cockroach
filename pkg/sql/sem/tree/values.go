package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ValuesClause struct {
	Rows []Exprs
}

func (node *ValuesClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615896)
	ctx.WriteString("VALUES ")
	comma := ""
	for i := range node.Rows {
		__antithesis_instrumentation__.Notify(615897)
		ctx.WriteString(comma)
		ctx.WriteByte('(')
		ctx.FormatNode(&node.Rows[i])
		ctx.WriteByte(')')
		comma = ", "
	}
}
