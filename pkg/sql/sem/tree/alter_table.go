package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

type AlterTable struct {
	IfExists bool
	Table    *UnresolvedObjectName
	Cmds     AlterTableCmds
}

func (node *AlterTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603066)
	ctx.WriteString("ALTER TABLE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603068)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603069)
	}
	__antithesis_instrumentation__.Notify(603067)
	ctx.FormatNode(node.Table)
	ctx.FormatNode(&node.Cmds)
}

type AlterTableCmds []AlterTableCmd

func (node *AlterTableCmds) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603070)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(603071)
		if i > 0 {
			__antithesis_instrumentation__.Notify(603073)
			ctx.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(603074)
		}
		__antithesis_instrumentation__.Notify(603072)
		ctx.FormatNode(n)
	}
}

type AlterTableCmd interface {
	NodeFormatter

	TelemetryCounter() telemetry.Counter

	alterTableCmd()
}

func (*AlterTableAddColumn) alterTableCmd()          { __antithesis_instrumentation__.Notify(603075) }
func (*AlterTableAddConstraint) alterTableCmd()      { __antithesis_instrumentation__.Notify(603076) }
func (*AlterTableAlterColumnType) alterTableCmd()    { __antithesis_instrumentation__.Notify(603077) }
func (*AlterTableAlterPrimaryKey) alterTableCmd()    { __antithesis_instrumentation__.Notify(603078) }
func (*AlterTableDropColumn) alterTableCmd()         { __antithesis_instrumentation__.Notify(603079) }
func (*AlterTableDropConstraint) alterTableCmd()     { __antithesis_instrumentation__.Notify(603080) }
func (*AlterTableDropNotNull) alterTableCmd()        { __antithesis_instrumentation__.Notify(603081) }
func (*AlterTableDropStored) alterTableCmd()         { __antithesis_instrumentation__.Notify(603082) }
func (*AlterTableSetNotNull) alterTableCmd()         { __antithesis_instrumentation__.Notify(603083) }
func (*AlterTableRenameColumn) alterTableCmd()       { __antithesis_instrumentation__.Notify(603084) }
func (*AlterTableRenameConstraint) alterTableCmd()   { __antithesis_instrumentation__.Notify(603085) }
func (*AlterTableSetAudit) alterTableCmd()           { __antithesis_instrumentation__.Notify(603086) }
func (*AlterTableSetDefault) alterTableCmd()         { __antithesis_instrumentation__.Notify(603087) }
func (*AlterTableSetOnUpdate) alterTableCmd()        { __antithesis_instrumentation__.Notify(603088) }
func (*AlterTableSetVisible) alterTableCmd()         { __antithesis_instrumentation__.Notify(603089) }
func (*AlterTableValidateConstraint) alterTableCmd() { __antithesis_instrumentation__.Notify(603090) }
func (*AlterTablePartitionByTable) alterTableCmd()   { __antithesis_instrumentation__.Notify(603091) }
func (*AlterTableInjectStats) alterTableCmd()        { __antithesis_instrumentation__.Notify(603092) }
func (*AlterTableSetStorageParams) alterTableCmd()   { __antithesis_instrumentation__.Notify(603093) }
func (*AlterTableResetStorageParams) alterTableCmd() { __antithesis_instrumentation__.Notify(603094) }

var _ AlterTableCmd = &AlterTableAddColumn{}
var _ AlterTableCmd = &AlterTableAddConstraint{}
var _ AlterTableCmd = &AlterTableAlterColumnType{}
var _ AlterTableCmd = &AlterTableDropColumn{}
var _ AlterTableCmd = &AlterTableDropConstraint{}
var _ AlterTableCmd = &AlterTableDropNotNull{}
var _ AlterTableCmd = &AlterTableDropStored{}
var _ AlterTableCmd = &AlterTableSetNotNull{}
var _ AlterTableCmd = &AlterTableRenameColumn{}
var _ AlterTableCmd = &AlterTableRenameConstraint{}
var _ AlterTableCmd = &AlterTableSetAudit{}
var _ AlterTableCmd = &AlterTableSetDefault{}
var _ AlterTableCmd = &AlterTableSetOnUpdate{}
var _ AlterTableCmd = &AlterTableSetVisible{}
var _ AlterTableCmd = &AlterTableValidateConstraint{}
var _ AlterTableCmd = &AlterTablePartitionByTable{}
var _ AlterTableCmd = &AlterTableInjectStats{}
var _ AlterTableCmd = &AlterTableSetStorageParams{}
var _ AlterTableCmd = &AlterTableResetStorageParams{}

type ColumnMutationCmd interface {
	AlterTableCmd
	GetColumn() Name
}

type AlterTableAddColumn struct {
	IfNotExists bool
	ColumnDef   *ColumnTableDef
}

func (node *AlterTableAddColumn) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603095)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "add_column")
}

func (node *AlterTableAddColumn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603096)
	ctx.WriteString(" ADD COLUMN ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(603098)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603099)
	}
	__antithesis_instrumentation__.Notify(603097)
	ctx.FormatNode(node.ColumnDef)
}

func (node *AlterTable) HoistAddColumnConstraints() {
	__antithesis_instrumentation__.Notify(603100)
	var normalizedCmds AlterTableCmds

	for _, cmd := range node.Cmds {
		__antithesis_instrumentation__.Notify(603102)
		normalizedCmds = append(normalizedCmds, cmd)

		if t, ok := cmd.(*AlterTableAddColumn); ok {
			__antithesis_instrumentation__.Notify(603103)
			d := t.ColumnDef
			for _, checkExpr := range d.CheckExprs {
				__antithesis_instrumentation__.Notify(603105)
				normalizedCmds = append(normalizedCmds,
					&AlterTableAddConstraint{
						ConstraintDef: &CheckConstraintTableDef{
							Expr: checkExpr.Expr,
							Name: checkExpr.ConstraintName,
						},
						ValidationBehavior: ValidationDefault,
					},
				)
			}
			__antithesis_instrumentation__.Notify(603104)
			d.CheckExprs = nil
			if d.HasFKConstraint() {
				__antithesis_instrumentation__.Notify(603106)
				var targetCol NameList
				if d.References.Col != "" {
					__antithesis_instrumentation__.Notify(603108)
					targetCol = append(targetCol, d.References.Col)
				} else {
					__antithesis_instrumentation__.Notify(603109)
				}
				__antithesis_instrumentation__.Notify(603107)
				fk := &ForeignKeyConstraintTableDef{
					Table:    *d.References.Table,
					FromCols: NameList{d.Name},
					ToCols:   targetCol,
					Name:     d.References.ConstraintName,
					Actions:  d.References.Actions,
					Match:    d.References.Match,
				}
				constraint := &AlterTableAddConstraint{
					ConstraintDef:      fk,
					ValidationBehavior: ValidationDefault,
				}
				normalizedCmds = append(normalizedCmds, constraint)
				d.References.Table = nil
				telemetry.Inc(sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "add_column.references"))
			} else {
				__antithesis_instrumentation__.Notify(603110)
			}
		} else {
			__antithesis_instrumentation__.Notify(603111)
		}
	}
	__antithesis_instrumentation__.Notify(603101)
	node.Cmds = normalizedCmds
}

type ValidationBehavior int

const (
	ValidationDefault ValidationBehavior = iota

	ValidationSkip
)

type AlterTableAddConstraint struct {
	ConstraintDef      ConstraintTableDef
	ValidationBehavior ValidationBehavior
}

func (node *AlterTableAddConstraint) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603112)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "add_constraint")
}

func (node *AlterTableAddConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603113)
	ctx.WriteString(" ADD ")
	ctx.FormatNode(node.ConstraintDef)
	if node.ValidationBehavior == ValidationSkip {
		__antithesis_instrumentation__.Notify(603114)
		ctx.WriteString(" NOT VALID")
	} else {
		__antithesis_instrumentation__.Notify(603115)
	}
}

type AlterTableAlterColumnType struct {
	Collation string
	Column    Name
	ToType    ResolvableTypeReference
	Using     Expr
}

func (node *AlterTableAlterColumnType) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603116)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "alter_column_type")
}

func (node *AlterTableAlterColumnType) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603117)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET DATA TYPE ")
	ctx.WriteString(node.ToType.SQLString())
	if len(node.Collation) > 0 {
		__antithesis_instrumentation__.Notify(603119)
		ctx.WriteString(" COLLATE ")
		lex.EncodeLocaleName(&ctx.Buffer, node.Collation)
	} else {
		__antithesis_instrumentation__.Notify(603120)
	}
	__antithesis_instrumentation__.Notify(603118)
	if node.Using != nil {
		__antithesis_instrumentation__.Notify(603121)
		ctx.WriteString(" USING ")
		ctx.FormatNode(node.Using)
	} else {
		__antithesis_instrumentation__.Notify(603122)
	}
}

func (node *AlterTableAlterColumnType) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603123)
	return node.Column
}

type AlterTableAlterPrimaryKey struct {
	Columns       IndexElemList
	Sharded       *ShardedIndexDef
	Name          Name
	StorageParams StorageParams
}

func (node *AlterTableAlterPrimaryKey) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603124)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "alter_primary_key")
}

func (node *AlterTableAlterPrimaryKey) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603125)
	ctx.WriteString(" ALTER PRIMARY KEY USING COLUMNS (")
	ctx.FormatNode(&node.Columns)
	ctx.WriteString(")")
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(603126)
		ctx.FormatNode(node.Sharded)
	} else {
		__antithesis_instrumentation__.Notify(603127)
	}
}

type AlterTableDropColumn struct {
	IfExists     bool
	Column       Name
	DropBehavior DropBehavior
}

func (node *AlterTableDropColumn) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603128)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "drop_column")
}

func (node *AlterTableDropColumn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603129)
	ctx.WriteString(" DROP COLUMN ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603131)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603132)
	}
	__antithesis_instrumentation__.Notify(603130)
	ctx.FormatNode(&node.Column)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(603133)
		ctx.Printf(" %s", node.DropBehavior)
	} else {
		__antithesis_instrumentation__.Notify(603134)
	}
}

type AlterTableDropConstraint struct {
	IfExists     bool
	Constraint   Name
	DropBehavior DropBehavior
}

func (node *AlterTableDropConstraint) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603135)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "drop_constraint")
}

func (node *AlterTableDropConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603136)
	ctx.WriteString(" DROP CONSTRAINT ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603138)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603139)
	}
	__antithesis_instrumentation__.Notify(603137)
	ctx.FormatNode(&node.Constraint)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(603140)
		ctx.Printf(" %s", node.DropBehavior)
	} else {
		__antithesis_instrumentation__.Notify(603141)
	}
}

type AlterTableValidateConstraint struct {
	Constraint Name
}

func (node *AlterTableValidateConstraint) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603142)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "validate_constraint")
}

func (node *AlterTableValidateConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603143)
	ctx.WriteString(" VALIDATE CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
}

type AlterTableRenameColumn struct {
	Column  Name
	NewName Name
}

func (node *AlterTableRenameColumn) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603144)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "rename_column")
}

func (node *AlterTableRenameColumn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603145)
	ctx.WriteString(" RENAME COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}

type AlterTableRenameConstraint struct {
	Constraint Name
	NewName    Name
}

func (node *AlterTableRenameConstraint) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603146)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "rename_constraint")
}

func (node *AlterTableRenameConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603147)
	ctx.WriteString(" RENAME CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}

type AlterTableSetDefault struct {
	Column  Name
	Default Expr
}

func (node *AlterTableSetDefault) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603148)
	return node.Column
}

func (node *AlterTableSetDefault) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603149)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "set_default")
}

func (node *AlterTableSetDefault) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603150)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.Default == nil {
		__antithesis_instrumentation__.Notify(603151)
		ctx.WriteString(" DROP DEFAULT")
	} else {
		__antithesis_instrumentation__.Notify(603152)
		ctx.WriteString(" SET DEFAULT ")
		ctx.FormatNode(node.Default)
	}
}

type AlterTableSetOnUpdate struct {
	Column Name
	Expr   Expr
}

func (node *AlterTableSetOnUpdate) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603153)
	return node.Column
}

func (node *AlterTableSetOnUpdate) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603154)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "set_on_update")
}

func (node *AlterTableSetOnUpdate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603155)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.Expr == nil {
		__antithesis_instrumentation__.Notify(603156)
		ctx.WriteString(" DROP ON UPDATE")
	} else {
		__antithesis_instrumentation__.Notify(603157)
		ctx.WriteString(" SET ON UPDATE ")
		ctx.FormatNode(node.Expr)
	}
}

type AlterTableSetVisible struct {
	Column  Name
	Visible bool
}

func (node *AlterTableSetVisible) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603158)
	return node.Column
}

func (node *AlterTableSetVisible) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603159)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "set_visible")
}

func (node *AlterTableSetVisible) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603160)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET ")
	if !node.Visible {
		__antithesis_instrumentation__.Notify(603162)
		ctx.WriteString("NOT ")
	} else {
		__antithesis_instrumentation__.Notify(603163)
	}
	__antithesis_instrumentation__.Notify(603161)
	ctx.WriteString("VISIBLE")
}

type AlterTableSetNotNull struct {
	Column Name
}

func (node *AlterTableSetNotNull) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603164)
	return node.Column
}

func (node *AlterTableSetNotNull) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603165)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "set_not_null")
}

func (node *AlterTableSetNotNull) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603166)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET NOT NULL")
}

type AlterTableDropNotNull struct {
	Column Name
}

func (node *AlterTableDropNotNull) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603167)
	return node.Column
}

func (node *AlterTableDropNotNull) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603168)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "drop_not_null")
}

func (node *AlterTableDropNotNull) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603169)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" DROP NOT NULL")
}

type AlterTableDropStored struct {
	Column Name
}

func (node *AlterTableDropStored) GetColumn() Name {
	__antithesis_instrumentation__.Notify(603170)
	return node.Column
}

func (node *AlterTableDropStored) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603171)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "drop_stored")
}

func (node *AlterTableDropStored) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603172)
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" DROP STORED")
}

type AlterTablePartitionByTable struct {
	*PartitionByTable
}

func (node *AlterTablePartitionByTable) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603173)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "partition_by")
}

func (node *AlterTablePartitionByTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603174)
	ctx.FormatNode(node.PartitionByTable)
}

type AuditMode int

const (
	AuditModeDisable AuditMode = iota

	AuditModeReadWrite
)

var auditModeName = [...]string{
	AuditModeDisable:   "OFF",
	AuditModeReadWrite: "READ WRITE",
}

func (m AuditMode) String() string {
	__antithesis_instrumentation__.Notify(603175)
	return auditModeName[m]
}

func (m AuditMode) TelemetryName() string {
	__antithesis_instrumentation__.Notify(603176)
	return strings.ReplaceAll(strings.ToLower(m.String()), " ", "_")
}

type AlterTableSetAudit struct {
	Mode AuditMode
}

func (node *AlterTableSetAudit) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603177)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "set_audit")
}

func (node *AlterTableSetAudit) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603178)
	ctx.WriteString(" EXPERIMENTAL_AUDIT SET ")
	ctx.WriteString(node.Mode.String())
}

type AlterTableInjectStats struct {
	Stats Expr
}

func (node *AlterTableInjectStats) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603179)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra("table", "inject_stats")
}

func (node *AlterTableInjectStats) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603180)
	ctx.WriteString(" INJECT STATISTICS ")
	ctx.FormatNode(node.Stats)
}

type AlterTableSetStorageParams struct {
	StorageParams StorageParams
}

func (node *AlterTableSetStorageParams) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603181)
	return sqltelemetry.SchemaChangeAlterCounter("set_storage_param")
}

func (node *AlterTableSetStorageParams) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603182)
	ctx.WriteString(" SET (")
	ctx.FormatNode(&node.StorageParams)
	ctx.WriteString(")")
}

type AlterTableResetStorageParams struct {
	Params NameList
}

func (node *AlterTableResetStorageParams) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603183)
	return sqltelemetry.SchemaChangeAlterCounter("set_storage_param")
}

func (node *AlterTableResetStorageParams) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603184)
	ctx.WriteString(" RESET (")
	ctx.FormatNode(&node.Params)
	ctx.WriteString(")")
}

type AlterTableLocality struct {
	Name     *UnresolvedObjectName
	IfExists bool
	Locality *Locality
}

var _ Statement = &AlterTableLocality{}

func (node *AlterTableLocality) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603185)
	ctx.WriteString("ALTER TABLE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603187)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603188)
	}
	__antithesis_instrumentation__.Notify(603186)
	ctx.FormatNode(node.Name)
	ctx.WriteString(" SET ")
	ctx.FormatNode(node.Locality)
}

type AlterTableSetSchema struct {
	Name           *UnresolvedObjectName
	Schema         Name
	IfExists       bool
	IsView         bool
	IsMaterialized bool
	IsSequence     bool
}

func (node *AlterTableSetSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603189)
	ctx.WriteString("ALTER")
	if node.IsView {
		__antithesis_instrumentation__.Notify(603192)
		if node.IsMaterialized {
			__antithesis_instrumentation__.Notify(603194)
			ctx.WriteString(" MATERIALIZED")
		} else {
			__antithesis_instrumentation__.Notify(603195)
		}
		__antithesis_instrumentation__.Notify(603193)
		ctx.WriteString(" VIEW ")
	} else {
		__antithesis_instrumentation__.Notify(603196)
		if node.IsSequence {
			__antithesis_instrumentation__.Notify(603197)
			ctx.WriteString(" SEQUENCE ")
		} else {
			__antithesis_instrumentation__.Notify(603198)
			ctx.WriteString(" TABLE ")
		}
	}
	__antithesis_instrumentation__.Notify(603190)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603199)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603200)
	}
	__antithesis_instrumentation__.Notify(603191)
	ctx.FormatNode(node.Name)
	ctx.WriteString(" SET SCHEMA ")
	ctx.FormatNode(&node.Schema)
}

func (node *AlterTableSetSchema) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603201)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra(
		GetTableType(node.IsSequence, node.IsView, node.IsMaterialized),
		"set_schema")
}

type AlterTableOwner struct {
	Name           *UnresolvedObjectName
	Owner          RoleSpec
	IfExists       bool
	IsView         bool
	IsMaterialized bool
	IsSequence     bool
}

func (node *AlterTableOwner) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(603202)
	return sqltelemetry.SchemaChangeAlterCounterWithExtra(
		GetTableType(node.IsSequence, node.IsView, node.IsMaterialized),
		"owner_to",
	)
}

func (node *AlterTableOwner) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603203)
	ctx.WriteString("ALTER")
	if node.IsView {
		__antithesis_instrumentation__.Notify(603206)
		if node.IsMaterialized {
			__antithesis_instrumentation__.Notify(603208)
			ctx.WriteString(" MATERIALIZED")
		} else {
			__antithesis_instrumentation__.Notify(603209)
		}
		__antithesis_instrumentation__.Notify(603207)
		ctx.WriteString(" VIEW ")
	} else {
		__antithesis_instrumentation__.Notify(603210)
		if node.IsSequence {
			__antithesis_instrumentation__.Notify(603211)
			ctx.WriteString(" SEQUENCE ")
		} else {
			__antithesis_instrumentation__.Notify(603212)
			ctx.WriteString(" TABLE ")
		}
	}
	__antithesis_instrumentation__.Notify(603204)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603213)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603214)
	}
	__antithesis_instrumentation__.Notify(603205)
	ctx.FormatNode(node.Name)
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNode(&node.Owner)
}

func GetTableType(isSequence bool, isView bool, isMaterialized bool) string {
	__antithesis_instrumentation__.Notify(603215)
	tableType := "table"
	if isSequence {
		__antithesis_instrumentation__.Notify(603217)
		tableType = "sequence"
	} else {
		__antithesis_instrumentation__.Notify(603218)
		if isView {
			__antithesis_instrumentation__.Notify(603219)
			if isMaterialized {
				__antithesis_instrumentation__.Notify(603220)
				tableType = "materialized_view"
			} else {
				__antithesis_instrumentation__.Notify(603221)
				tableType = "view"
			}
		} else {
			__antithesis_instrumentation__.Notify(603222)
		}
	}
	__antithesis_instrumentation__.Notify(603216)

	return tableType
}
