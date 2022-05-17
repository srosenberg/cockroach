package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnDatabase struct {
	Name    Name
	Comment *string
}

func (n *CommentOnDatabase) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604430)
	ctx.WriteString("COMMENT ON DATABASE ")
	ctx.FormatNode(&n.Name)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604431)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604432)
			ctx.WriteString("'_'")
		} else {
			__antithesis_instrumentation__.Notify(604433)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604434)
		ctx.WriteString("NULL")
	}
}
