package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type StreamIngestion struct {
	Targets TargetList
	From    StringOrPlaceholderOptList
	AsOf    AsOfClause
}

var _ Statement = &StreamIngestion{}

func (node *StreamIngestion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614439)
	ctx.WriteString("RESTORE ")
	ctx.FormatNode(&node.Targets)
	ctx.WriteString(" ")
	ctx.WriteString("FROM REPLICATION STREAM FROM ")
	ctx.FormatNode(&node.From)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(614440)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.AsOf)
	} else {
		__antithesis_instrumentation__.Notify(614441)
	}
}
