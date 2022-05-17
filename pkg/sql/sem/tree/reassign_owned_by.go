package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ReassignOwnedBy struct {
	OldRoles RoleSpecList
	NewRole  RoleSpec
}

var _ Statement = &ReassignOwnedBy{}

func (node *ReassignOwnedBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612826)
	ctx.WriteString("REASSIGN OWNED BY ")
	for i := range node.OldRoles {
		__antithesis_instrumentation__.Notify(612828)
		if i > 0 {
			__antithesis_instrumentation__.Notify(612830)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(612831)
		}
		__antithesis_instrumentation__.Notify(612829)
		node.OldRoles[i].Format(ctx)
	}
	__antithesis_instrumentation__.Notify(612827)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewRole)
}
