package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnTable struct {
	Table   *UnresolvedObjectName
	Comment *string
}

func (n *CommentOnTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604445)
	ctx.WriteString("COMMENT ON TABLE ")
	ctx.FormatNode(n.Table)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604446)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604447)
			ctx.WriteString("'_'")
		} else {
			__antithesis_instrumentation__.Notify(604448)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604449)
		ctx.WriteString("NULL")
	}
}
