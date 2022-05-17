package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

type LocalityLevel int

const (
	LocalityLevelGlobal LocalityLevel = iota

	LocalityLevelTable

	LocalityLevelRow
)

const (
	RegionEnum string = "crdb_internal_region"

	RegionalByRowRegionDefaultCol string = "crdb_region"

	RegionalByRowRegionDefaultColName Name = Name(RegionalByRowRegionDefaultCol)

	RegionalByRowRegionNotSpecifiedName = ""

	PrimaryRegionNotSpecifiedName Name = ""
)

type Locality struct {
	LocalityLevel LocalityLevel

	TableRegion Name

	RegionalByRowColumn Name
}

const (
	TelemetryNameGlobal            = "global"
	TelemetryNameRegionalByTable   = "regional_by_table"
	TelemetryNameRegionalByTableIn = "regional_by_table_in"
	TelemetryNameRegionalByRow     = "regional_by_row"
	TelemetryNameRegionalByRowAs   = "regional_by_row_as"
)

func (node *Locality) TelemetryName() string {
	__antithesis_instrumentation__.Notify(612860)
	switch node.LocalityLevel {
	case LocalityLevelGlobal:
		__antithesis_instrumentation__.Notify(612861)
		return TelemetryNameGlobal
	case LocalityLevelTable:
		__antithesis_instrumentation__.Notify(612862)
		if node.TableRegion != RegionalByRowRegionNotSpecifiedName {
			__antithesis_instrumentation__.Notify(612867)
			return TelemetryNameRegionalByTableIn
		} else {
			__antithesis_instrumentation__.Notify(612868)
		}
		__antithesis_instrumentation__.Notify(612863)
		return TelemetryNameRegionalByTable
	case LocalityLevelRow:
		__antithesis_instrumentation__.Notify(612864)
		if node.RegionalByRowColumn != PrimaryRegionNotSpecifiedName {
			__antithesis_instrumentation__.Notify(612869)
			return TelemetryNameRegionalByRowAs
		} else {
			__antithesis_instrumentation__.Notify(612870)
		}
		__antithesis_instrumentation__.Notify(612865)
		return TelemetryNameRegionalByRow
	default:
		__antithesis_instrumentation__.Notify(612866)
		panic(fmt.Sprintf("unknown locality: %#v", node.LocalityLevel))
	}
}

func (node *Locality) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612871)
	ctx.WriteString("LOCALITY ")
	switch node.LocalityLevel {
	case LocalityLevelGlobal:
		__antithesis_instrumentation__.Notify(612872)
		ctx.WriteString("GLOBAL")
	case LocalityLevelTable:
		__antithesis_instrumentation__.Notify(612873)
		ctx.WriteString("REGIONAL BY TABLE IN ")
		if node.TableRegion != "" {
			__antithesis_instrumentation__.Notify(612876)
			ctx.FormatNode(&node.TableRegion)
		} else {
			__antithesis_instrumentation__.Notify(612877)
			ctx.WriteString("PRIMARY REGION")
		}
	case LocalityLevelRow:
		__antithesis_instrumentation__.Notify(612874)
		ctx.WriteString("REGIONAL BY ROW")
		if node.RegionalByRowColumn != "" {
			__antithesis_instrumentation__.Notify(612878)
			ctx.WriteString(" AS ")
			ctx.FormatNode(&node.RegionalByRowColumn)
		} else {
			__antithesis_instrumentation__.Notify(612879)
		}
	default:
		__antithesis_instrumentation__.Notify(612875)
		panic(fmt.Sprintf("unknown locality: %#v", node.LocalityLevel))
	}
}
