package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type DataPlacement uint32

const (
	DataPlacementUnspecified DataPlacement = iota

	DataPlacementDefault

	DataPlacementRestricted
)

func (node *DataPlacement) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605431)
	switch *node {
	case DataPlacementRestricted:
		__antithesis_instrumentation__.Notify(605432)
		ctx.WriteString("PLACEMENT RESTRICTED")
	case DataPlacementDefault:
		__antithesis_instrumentation__.Notify(605433)
		ctx.WriteString("PLACEMENT DEFAULT")
	case DataPlacementUnspecified:
		__antithesis_instrumentation__.Notify(605434)
	default:
		__antithesis_instrumentation__.Notify(605435)
		panic(errors.AssertionFailedf("unknown data placement strategy: %d", *node))
	}
}

func (node *DataPlacement) TelemetryName() string {
	__antithesis_instrumentation__.Notify(605436)
	switch *node {
	case DataPlacementRestricted:
		__antithesis_instrumentation__.Notify(605437)
		return "restricted"
	case DataPlacementDefault:
		__antithesis_instrumentation__.Notify(605438)
		return "default"
	case DataPlacementUnspecified:
		__antithesis_instrumentation__.Notify(605439)
		return "unspecified"
	default:
		__antithesis_instrumentation__.Notify(605440)
		panic(errors.AssertionFailedf("unknown data placement: %d", *node))
	}
}
