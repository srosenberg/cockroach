package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterChangefeed struct {
	Jobs Expr
	Cmds AlterChangefeedCmds
}

var _ Statement = &AlterChangefeed{}

func (node *AlterChangefeed) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602932)
	ctx.WriteString(`ALTER CHANGEFEED `)
	ctx.FormatNode(node.Jobs)
	ctx.FormatNode(&node.Cmds)
}

type AlterChangefeedCmds []AlterChangefeedCmd

func (node *AlterChangefeedCmds) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602933)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(602934)
		if i > 0 {
			__antithesis_instrumentation__.Notify(602936)
			ctx.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(602937)
		}
		__antithesis_instrumentation__.Notify(602935)
		ctx.FormatNode(n)
	}
}

type AlterChangefeedCmd interface {
	NodeFormatter

	alterChangefeedCmd()
}

func (*AlterChangefeedAddTarget) alterChangefeedCmd()  { __antithesis_instrumentation__.Notify(602938) }
func (*AlterChangefeedDropTarget) alterChangefeedCmd() { __antithesis_instrumentation__.Notify(602939) }
func (*AlterChangefeedSetOptions) alterChangefeedCmd() { __antithesis_instrumentation__.Notify(602940) }
func (*AlterChangefeedUnsetOptions) alterChangefeedCmd() {
	__antithesis_instrumentation__.Notify(602941)
}

var _ AlterChangefeedCmd = &AlterChangefeedAddTarget{}
var _ AlterChangefeedCmd = &AlterChangefeedDropTarget{}
var _ AlterChangefeedCmd = &AlterChangefeedSetOptions{}
var _ AlterChangefeedCmd = &AlterChangefeedUnsetOptions{}

type AlterChangefeedAddTarget struct {
	Targets ChangefeedTargets
	Options KVOptions
}

func (node *AlterChangefeedAddTarget) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602942)
	ctx.WriteString(" ADD ")
	ctx.FormatNode(&node.Targets)
	if node.Options != nil {
		__antithesis_instrumentation__.Notify(602943)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(602944)
	}
}

type AlterChangefeedDropTarget struct {
	Targets ChangefeedTargets
}

func (node *AlterChangefeedDropTarget) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602945)
	ctx.WriteString(" DROP ")
	ctx.FormatNode(&node.Targets)
}

type AlterChangefeedSetOptions struct {
	Options KVOptions
}

func (node *AlterChangefeedSetOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602946)
	ctx.WriteString(" SET ")
	ctx.FormatNode(&node.Options)
}

type AlterChangefeedUnsetOptions struct {
	Options NameList
}

func (node *AlterChangefeedUnsetOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602947)
	ctx.WriteString(" UNSET ")
	ctx.FormatNode(&node.Options)
}
