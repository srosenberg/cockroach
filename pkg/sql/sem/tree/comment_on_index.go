package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnIndex struct {
	Index   TableIndexName
	Comment *string
}

func (n *CommentOnIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604435)
	ctx.WriteString("COMMENT ON INDEX ")
	ctx.FormatNode(&n.Index)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604436)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604437)
			ctx.WriteString("'_'")
		} else {
			__antithesis_instrumentation__.Notify(604438)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604439)
		ctx.WriteString("NULL")
	}
}
