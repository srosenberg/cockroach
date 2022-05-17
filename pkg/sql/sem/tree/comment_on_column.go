package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnColumn struct {
	*ColumnItem
	Comment *string
}

func (n *CommentOnColumn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604420)
	ctx.WriteString("COMMENT ON COLUMN ")
	ctx.FormatNode(n.ColumnItem)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604421)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604422)
			ctx.WriteString("'_'")
		} else {
			__antithesis_instrumentation__.Notify(604423)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604424)
		ctx.WriteString("NULL")
	}
}
