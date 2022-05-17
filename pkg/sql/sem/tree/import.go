package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Import struct {
	Table      *TableName
	Into       bool
	IntoCols   NameList
	FileFormat string
	Files      Exprs
	Bundle     bool
	Options    KVOptions
}

var _ Statement = &Import{}

func (node *Import) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609997)
	ctx.WriteString("IMPORT ")

	if node.Bundle {
		__antithesis_instrumentation__.Notify(609999)
		if node.Table != nil {
			__antithesis_instrumentation__.Notify(610001)
			ctx.WriteString("TABLE ")
			ctx.FormatNode(node.Table)
			ctx.WriteString(" FROM ")
		} else {
			__antithesis_instrumentation__.Notify(610002)
		}
		__antithesis_instrumentation__.Notify(610000)
		ctx.WriteString(node.FileFormat)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Files)
	} else {
		__antithesis_instrumentation__.Notify(610003)
		if node.Into {
			__antithesis_instrumentation__.Notify(610005)
			ctx.WriteString("INTO ")
			ctx.FormatNode(node.Table)
			if node.IntoCols != nil {
				__antithesis_instrumentation__.Notify(610006)
				ctx.WriteByte('(')
				ctx.FormatNode(&node.IntoCols)
				ctx.WriteString(") ")
			} else {
				__antithesis_instrumentation__.Notify(610007)
				ctx.WriteString(" ")
			}
		} else {
			__antithesis_instrumentation__.Notify(610008)
			ctx.WriteString("TABLE ")
			ctx.FormatNode(node.Table)
		}
		__antithesis_instrumentation__.Notify(610004)
		ctx.WriteString(node.FileFormat)
		ctx.WriteString(" DATA (")
		ctx.FormatNode(&node.Files)
		ctx.WriteString(")")
	}
	__antithesis_instrumentation__.Notify(609998)

	if node.Options != nil {
		__antithesis_instrumentation__.Notify(610009)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(610010)
	}
}
