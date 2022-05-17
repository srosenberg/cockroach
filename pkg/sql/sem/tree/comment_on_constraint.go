package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type CommentOnConstraint struct {
	Constraint Name
	Table      *UnresolvedObjectName
	Comment    *string
}

func (n *CommentOnConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604425)
	ctx.WriteString("COMMENT ON CONSTRAINT ")
	ctx.FormatNode(&n.Constraint)
	ctx.WriteString(" ON ")
	ctx.FormatNode(n.Table)
	ctx.WriteString(" IS ")
	if n.Comment != nil {
		__antithesis_instrumentation__.Notify(604426)

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604427)
			ctx.WriteByte('_')
		} else {
			__antithesis_instrumentation__.Notify(604428)
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, *n.Comment, ctx.flags.EncodeFlags())
		}
	} else {
		__antithesis_instrumentation__.Notify(604429)
		ctx.WriteString("NULL")
	}
}
