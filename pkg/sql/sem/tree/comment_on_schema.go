package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnSchema struct {
	Name    Name
	Comment *string
}

func (n *CommentOnSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604440)
	ctx.WriteString("COMMENT ON SCHEMA ")
	ctx.FormatNode(&n.Name)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604441)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604442)
			ctx.WriteByte('_')
		} else {
			__antithesis_instrumentation__.Notify(604443)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604444)
		ctx.WriteString("NULL")
	}
}
