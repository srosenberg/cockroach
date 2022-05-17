package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type CreateChangefeed struct {
	Targets ChangefeedTargets
	SinkURI Expr
	Options KVOptions
}

var _ Statement = &CreateChangefeed{}

func (node *CreateChangefeed) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604311)
	if node.SinkURI != nil {
		__antithesis_instrumentation__.Notify(604314)
		ctx.WriteString("CREATE ")
	} else {
		__antithesis_instrumentation__.Notify(604315)

		ctx.WriteString("EXPERIMENTAL ")
	}
	__antithesis_instrumentation__.Notify(604312)
	ctx.WriteString("CHANGEFEED FOR ")
	ctx.FormatNode(&node.Targets)
	if node.SinkURI != nil {
		__antithesis_instrumentation__.Notify(604316)
		ctx.WriteString(" INTO ")
		ctx.FormatNode(node.SinkURI)
	} else {
		__antithesis_instrumentation__.Notify(604317)
	}
	__antithesis_instrumentation__.Notify(604313)
	if node.Options != nil {
		__antithesis_instrumentation__.Notify(604318)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(604319)
	}
}

type ChangefeedTarget struct {
	TableName  TablePattern
	FamilyName Name
}

func (ct *ChangefeedTarget) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604320)
	ctx.WriteString("TABLE ")
	ctx.FormatNode(ct.TableName)
	if ct.FamilyName != "" {
		__antithesis_instrumentation__.Notify(604321)
		ctx.WriteString(" FAMILY ")
		ctx.FormatNode(&ct.FamilyName)
	} else {
		__antithesis_instrumentation__.Notify(604322)
	}
}

type ChangefeedTargets []ChangefeedTarget

func (cts *ChangefeedTargets) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604323)
	for i, ct := range *cts {
		__antithesis_instrumentation__.Notify(604324)
		if i > 0 {
			__antithesis_instrumentation__.Notify(604326)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(604327)
		}
		__antithesis_instrumentation__.Notify(604325)
		ctx.FormatNode(&ct)
	}
}
