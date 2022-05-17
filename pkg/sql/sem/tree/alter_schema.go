package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterSchema struct {
	Schema ObjectNamePrefix
	Cmd    AlterSchemaCmd
}

var _ Statement = &AlterSchema{}

func (node *AlterSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603057)
	ctx.WriteString("ALTER SCHEMA ")
	ctx.FormatNode(&node.Schema)
	ctx.FormatNode(node.Cmd)
}

type AlterSchemaCmd interface {
	NodeFormatter
	alterSchemaCmd()
}

func (*AlterSchemaRename) alterSchemaCmd() { __antithesis_instrumentation__.Notify(603058) }

type AlterSchemaRename struct {
	NewName Name
}

func (node *AlterSchemaRename) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603059)
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(&node.NewName)
}

func (*AlterSchemaOwner) alterSchemaCmd() { __antithesis_instrumentation__.Notify(603060) }

type AlterSchemaOwner struct {
	Owner RoleSpec
}

func (node *AlterSchemaOwner) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603061)
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNode(&node.Owner)
}
