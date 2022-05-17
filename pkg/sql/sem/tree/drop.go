package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

type DropBehavior int

const (
	DropDefault DropBehavior = iota
	DropRestrict
	DropCascade
)

var dropBehaviorName = [...]string{
	DropDefault:  "",
	DropRestrict: "RESTRICT",
	DropCascade:  "CASCADE",
}

func (d DropBehavior) String() string {
	__antithesis_instrumentation__.Notify(607618)
	return dropBehaviorName[d]
}

type DropDatabase struct {
	Name         Name
	IfExists     bool
	DropBehavior DropBehavior
}

func (node *DropDatabase) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607619)
	ctx.WriteString("DROP DATABASE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607621)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607622)
	}
	__antithesis_instrumentation__.Notify(607620)
	ctx.FormatNode(&node.Name)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607623)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607624)
	}
}

type DropIndex struct {
	IndexList    TableIndexNames
	IfExists     bool
	DropBehavior DropBehavior
	Concurrently bool
}

func (node *DropIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607625)
	ctx.WriteString("DROP INDEX ")
	if node.Concurrently {
		__antithesis_instrumentation__.Notify(607628)
		ctx.WriteString("CONCURRENTLY ")
	} else {
		__antithesis_instrumentation__.Notify(607629)
	}
	__antithesis_instrumentation__.Notify(607626)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607630)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607631)
	}
	__antithesis_instrumentation__.Notify(607627)
	ctx.FormatNode(&node.IndexList)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607632)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607633)
	}
}

type DropTable struct {
	Names        TableNames
	IfExists     bool
	DropBehavior DropBehavior
}

func (node *DropTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607634)
	ctx.WriteString("DROP TABLE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607636)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607637)
	}
	__antithesis_instrumentation__.Notify(607635)
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607638)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607639)
	}
}

type DropView struct {
	Names          TableNames
	IfExists       bool
	DropBehavior   DropBehavior
	IsMaterialized bool
}

func (node *DropView) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607640)
	ctx.WriteString("DROP ")
	if node.IsMaterialized {
		__antithesis_instrumentation__.Notify(607643)
		ctx.WriteString("MATERIALIZED ")
	} else {
		__antithesis_instrumentation__.Notify(607644)
	}
	__antithesis_instrumentation__.Notify(607641)
	ctx.WriteString("VIEW ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607645)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607646)
	}
	__antithesis_instrumentation__.Notify(607642)
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607647)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607648)
	}
}

func (node *DropView) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(607649)
	return sqltelemetry.SchemaChangeDropCounter(
		GetTableType(
			false, true,
			node.IsMaterialized))
}

type DropSequence struct {
	Names        TableNames
	IfExists     bool
	DropBehavior DropBehavior
}

func (node *DropSequence) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607650)
	ctx.WriteString("DROP SEQUENCE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607652)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607653)
	}
	__antithesis_instrumentation__.Notify(607651)
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607654)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607655)
	}
}

type DropRole struct {
	Names    RoleSpecList
	IsRole   bool
	IfExists bool
}

func (node *DropRole) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607656)
	ctx.WriteString("DROP")
	if node.IsRole {
		__antithesis_instrumentation__.Notify(607659)
		ctx.WriteString(" ROLE ")
	} else {
		__antithesis_instrumentation__.Notify(607660)
		ctx.WriteString(" USER ")
	}
	__antithesis_instrumentation__.Notify(607657)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607661)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607662)
	}
	__antithesis_instrumentation__.Notify(607658)
	ctx.FormatNode(&node.Names)
}

type DropType struct {
	Names        []*UnresolvedObjectName
	IfExists     bool
	DropBehavior DropBehavior
}

var _ Statement = &DropType{}

func (node *DropType) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607663)
	ctx.WriteString("DROP TYPE ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607666)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607667)
	}
	__antithesis_instrumentation__.Notify(607664)
	for i := range node.Names {
		__antithesis_instrumentation__.Notify(607668)
		if i > 0 {
			__antithesis_instrumentation__.Notify(607670)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(607671)
		}
		__antithesis_instrumentation__.Notify(607669)
		ctx.FormatNode(node.Names[i])
	}
	__antithesis_instrumentation__.Notify(607665)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607672)
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607673)
	}
}

type DropSchema struct {
	Names        ObjectNamePrefixList
	IfExists     bool
	DropBehavior DropBehavior
}

var _ Statement = &DropSchema{}

func (node *DropSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607674)
	ctx.WriteString("DROP SCHEMA ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(607676)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(607677)
	}
	__antithesis_instrumentation__.Notify(607675)
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		__antithesis_instrumentation__.Notify(607678)
		ctx.WriteString(" ")
		ctx.WriteString(node.DropBehavior.String())
	} else {
		__antithesis_instrumentation__.Notify(607679)
	}
}
