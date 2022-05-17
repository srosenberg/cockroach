package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterIndex struct {
	IfExists bool
	Index    TableIndexName
	Cmds     AlterIndexCmds
}

var _ Statement = &AlterIndex{}

func (node *AlterIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603014)
	ctx.WriteString("ALTER INDEX ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603016)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603017)
	}
	__antithesis_instrumentation__.Notify(603015)
	ctx.FormatNode(&node.Index)
	ctx.FormatNode(&node.Cmds)
}

type AlterIndexCmds []AlterIndexCmd

func (node *AlterIndexCmds) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603018)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(603019)
		if i > 0 {
			__antithesis_instrumentation__.Notify(603021)
			ctx.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(603022)
		}
		__antithesis_instrumentation__.Notify(603020)
		ctx.FormatNode(n)
	}
}

type AlterIndexCmd interface {
	NodeFormatter

	alterIndexCmd()
}

func (*AlterIndexPartitionBy) alterIndexCmd() { __antithesis_instrumentation__.Notify(603023) }

var _ AlterIndexCmd = &AlterIndexPartitionBy{}

type AlterIndexPartitionBy struct {
	*PartitionByIndex
}

func (node *AlterIndexPartitionBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603024)
	ctx.FormatNode(node.PartitionByIndex)
}
