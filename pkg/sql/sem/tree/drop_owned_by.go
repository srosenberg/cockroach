package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type DropOwnedBy struct {
	Roles        RoleSpecList
	DropBehavior DropBehavior
}

var _ Statement = &DropOwnedBy{}

func (node *DropOwnedBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607680)
	ctx.WriteString("DROP OWNED BY ")
	for i := range node.Roles {
		__antithesis_instrumentation__.Notify(607682)
		if i > 0 {
			__antithesis_instrumentation__.Notify(607684)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(607685)
		}
		__antithesis_instrumentation__.Notify(607683)
		node.Roles[i].Format(ctx)
	}
	__antithesis_instrumentation__.Notify(607681)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607686)
		ctx.WriteString(" ")
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607687)
	}
}
