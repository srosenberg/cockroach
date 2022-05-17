package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

type AlterType struct {
	Type *UnresolvedObjectName
	Cmd  AlterTypeCmd
}

func (node *AlterType) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603223)
	ctx.WriteString("ALTER TYPE ")
	ctx.FormatNode(node.Type)
	ctx.FormatNode(node.Cmd)
}

type AlterTypeCmd interface {
	NodeFormatter
	alterTypeCmd()

	TelemetryCounter() telemetry.Counter
}

func (*AlterTypeAddValue) alterTypeCmd()    { __antithesis_instrumentation__.Notify(603224) }
func (*AlterTypeRenameValue) alterTypeCmd() { __antithesis_instrumentation__.Notify(603225) }
func (*AlterTypeRename) alterTypeCmd()      { __antithesis_instrumentation__.Notify(603226) }
func (*AlterTypeSetSchema) alterTypeCmd()   { __antithesis_instrumentation__.Notify(603227) }
func (*AlterTypeOwner) alterTypeCmd()       { __antithesis_instrumentation__.Notify(603228) }
func (*AlterTypeDropValue) alterTypeCmd()   { __antithesis_instrumentation__.Notify(603229) }

var _ AlterTypeCmd = &AlterTypeAddValue{}
var _ AlterTypeCmd = &AlterTypeRenameValue{}
var _ AlterTypeCmd = &AlterTypeRename{}
var _ AlterTypeCmd = &AlterTypeSetSchema{}
var _ AlterTypeCmd = &AlterTypeOwner{}
var _ AlterTypeCmd = &AlterTypeDropValue{}

type AlterTypeAddValue struct {
	NewVal      EnumValue
	IfNotExists bool
	Placement   *AlterTypeAddValuePlacement
}

func (node *AlterTypeAddValue) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603230)
	ctx.WriteString(" ADD VALUE ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(603232)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603233)
	}
	__antithesis_instrumentation__.Notify(603231)
	ctx.FormatNode(&node.NewVal)
	if node.Placement != nil {
		__antithesis_instrumentation__.Notify(603234)
		if node.Placement.Before {
			__antithesis_instrumentation__.Notify(603236)
			ctx.WriteString(" BEFORE ")
		} else {
			__antithesis_instrumentation__.Notify(603237)
			ctx.WriteString(" AFTER ")
		}
		__antithesis_instrumentation__.Notify(603235)
		ctx.FormatNode(&node.Placement.ExistingVal)
	} else {
		__antithesis_instrumentation__.Notify(603238)
	}
}

func (node *AlterTypeAddValue) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603239)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "add_value")
}

type AlterTypeAddValuePlacement struct {
	Before      bool
	ExistingVal EnumValue
}

type AlterTypeRenameValue struct {
	OldVal EnumValue
	NewVal EnumValue
}

func (node *AlterTypeRenameValue) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603240)
	ctx.WriteString(" RENAME VALUE ")
	ctx.FormatNode(&node.OldVal)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewVal)
}

func (node *AlterTypeRenameValue) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603241)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "rename_value")
}

type AlterTypeDropValue struct {
	Val EnumValue
}

func (node *AlterTypeDropValue) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603242)
	ctx.WriteString(" DROP VALUE ")
	ctx.FormatNode(&node.Val)
}

func (node *AlterTypeDropValue) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603243)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "drop_value")
}

type AlterTypeRename struct {
	NewName Name
}

func (node *AlterTypeRename) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603244)
	ctx.WriteString(" RENAME TO ")
	ctx.FormatNode(&node.NewName)
}

func (node *AlterTypeRename) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603245)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "rename")
}

type AlterTypeSetSchema struct {
	Schema Name
}

func (node *AlterTypeSetSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603246)
	ctx.WriteString(" SET SCHEMA ")
	ctx.FormatNode(&node.Schema)
}

func (node *AlterTypeSetSchema) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603247)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "set_schema")
}

type AlterTypeOwner struct {
	Owner RoleSpec
}

func (node *AlterTypeOwner) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603248)
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNode(&node.Owner)
}

func (node *AlterTypeOwner) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603249)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("type", "owner")
}
