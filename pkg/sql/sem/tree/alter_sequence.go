package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterSequence struct {
	IfExists bool
	Name     *UnresolvedObjectName
	Options  SequenceOptions
}

func (node *AlterSequence) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603062)
	ctx.WriteString("ALTER SEQUENCE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603064)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603065)
	}
	__antithesis_instrumentation__.Notify(603063)
	ctx.FormatNode(node.Name)
	ctx.FormatNode(&node.Options)
}
