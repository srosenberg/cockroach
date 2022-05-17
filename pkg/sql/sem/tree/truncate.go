package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Truncate struct {
	Tables       TableNames
	DropBehavior DropBehavior
}

func (node *Truncate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614650)
	ctx.WriteString("TRUNCATE TABLE ")
	sep := ""
	for i := range node.Tables {
		__antithesis_instrumentation__.Notify(614652)
		ctx.WriteString(sep)
		ctx.FormatNode(&node.Tables[i])
		sep = ", "
	}
	__antithesis_instrumentation__.Notify(614651)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(614653)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(614654)
	}
}
