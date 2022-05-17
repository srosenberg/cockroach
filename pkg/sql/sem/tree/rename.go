package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type RenameDatabase struct {
	Name    Name
	NewName Name
}

func (node *RenameDatabase) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612880)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(&node.NewName)
}

type ReparentDatabase struct {
	Name   Name
	Parent Name
}

func (node *ReparentDatabase) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612881)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" CONVERT TO SCHEMA WITH PARENT ")
	ctx.FormatNode(&node.Parent)
}

type RenameTable struct {
	Name           *UnresolvedObjectName
	NewName        *UnresolvedObjectName
	IfExists       bool
	IsView         bool
	IsMaterialized bool
	IsSequence     bool
}

func (node *RenameTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612882)
	ctx.WriteString("ALTER ")
	if node.IsView {
		__antithesis_instrumentation__.Notify(612885)
		if node.IsMaterialized {
			__antithesis_instrumentation__.Notify(612887)
			ctx.WriteString("MATERIALIZED ")
		} else {
			__antithesis_instrumentation__.Notify(612888)
		}
		__antithesis_instrumentation__.Notify(612886)
		ctx.WriteString("VIEW ")
	} else {
		__antithesis_instrumentation__.Notify(612889)
		if node.IsSequence {
			__antithesis_instrumentation__.Notify(612890)
			ctx.WriteString("SEQUENCE ")
		} else {
			__antithesis_instrumentation__.Notify(612891)
			ctx.WriteString("TABLE ")
		}
	}
	__antithesis_instrumentation__.Notify(612883)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612892)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(612893)
	}
	__antithesis_instrumentation__.Notify(612884)
	ctx.FormatNode(node.Name)
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(node.NewName)
}

type RenameIndex struct {
	Index    *TableIndexName
	NewName  UnrestrictedName
	IfExists bool
}

func (node *RenameIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612894)
	ctx.WriteString("ALTER INDEX ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612896)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(612897)
	}
	__antithesis_instrumentation__.Notify(612895)
	ctx.FormatNode(node.Index)
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(&node.NewName)
}

type RenameColumn struct {
	Table   TableName
	Name    Name
	NewName Name

	IfExists bool
}

func (node *RenameColumn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612898)
	ctx.WriteString("ALTER TABLE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612900)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(612901)
	}
	__antithesis_instrumentation__.Notify(612899)
	ctx.FormatNode(&node.Table)
	ctx.WriteString(" RENAME COLUMN ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}
