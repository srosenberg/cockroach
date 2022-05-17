package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Analyze struct {
	Table TableExpr
}

func (node *Analyze) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603250)
	ctx.WriteString("ANALYZE ")
	ctx.FormatNode(node.Table)
}
