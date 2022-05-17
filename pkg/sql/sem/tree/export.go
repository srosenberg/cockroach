package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Export struct {
	Query      *Select
	FileFormat string
	File       Expr
	Options    KVOptions
}

var _ Statement = &Export{}

func (node *Export) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609329)
	ctx.WriteString("EXPORT INTO ")
	ctx.WriteString(node.FileFormat)
	ctx.WriteString(" ")
	ctx.FormatNode(node.File)
	if node.Options != nil {
		__antithesis_instrumentation__.Notify(609331)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(609332)
	}
	__antithesis_instrumentation__.Notify(609330)
	ctx.WriteString(" FROM ")
	ctx.FormatNode(node.Query)
}
