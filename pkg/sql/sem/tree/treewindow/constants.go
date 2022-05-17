package treewindow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type WindowFrameMode int

const (
	RANGE WindowFrameMode = iota

	ROWS

	GROUPS
)

func (m WindowFrameMode) Name() string {
	__antithesis_instrumentation__.Notify(614602)
	switch m {
	case RANGE:
		__antithesis_instrumentation__.Notify(614604)
		return "Range"
	case ROWS:
		__antithesis_instrumentation__.Notify(614605)
		return "Rows"
	case GROUPS:
		__antithesis_instrumentation__.Notify(614606)
		return "Groups"
	default:
		__antithesis_instrumentation__.Notify(614607)
	}
	__antithesis_instrumentation__.Notify(614603)
	return ""
}

func (m WindowFrameMode) String() string {
	__antithesis_instrumentation__.Notify(614608)
	switch m {
	case RANGE:
		__antithesis_instrumentation__.Notify(614610)
		return "RANGE"
	case ROWS:
		__antithesis_instrumentation__.Notify(614611)
		return "ROWS"
	case GROUPS:
		__antithesis_instrumentation__.Notify(614612)
		return "GROUPS"
	default:
		__antithesis_instrumentation__.Notify(614613)
	}
	__antithesis_instrumentation__.Notify(614609)
	return ""
}

type WindowFrameBoundType int

const (
	UnboundedPreceding WindowFrameBoundType = iota

	OffsetPreceding

	CurrentRow

	OffsetFollowing

	UnboundedFollowing
)

func (ft WindowFrameBoundType) IsOffset() bool {
	__antithesis_instrumentation__.Notify(614614)
	return ft == OffsetPreceding || func() bool {
		__antithesis_instrumentation__.Notify(614615)
		return ft == OffsetFollowing == true
	}() == true
}

func (ft WindowFrameBoundType) Name() string {
	__antithesis_instrumentation__.Notify(614616)
	switch ft {
	case UnboundedPreceding:
		__antithesis_instrumentation__.Notify(614618)
		return "UnboundedPreceding"
	case OffsetPreceding:
		__antithesis_instrumentation__.Notify(614619)
		return "OffsetPreceding"
	case CurrentRow:
		__antithesis_instrumentation__.Notify(614620)
		return "CurrentRow"
	case OffsetFollowing:
		__antithesis_instrumentation__.Notify(614621)
		return "OffsetFollowing"
	case UnboundedFollowing:
		__antithesis_instrumentation__.Notify(614622)
		return "UnboundedFollowing"
	default:
		__antithesis_instrumentation__.Notify(614623)
	}
	__antithesis_instrumentation__.Notify(614617)
	return ""
}

func (ft WindowFrameBoundType) String() string {
	__antithesis_instrumentation__.Notify(614624)
	switch ft {
	case UnboundedPreceding:
		__antithesis_instrumentation__.Notify(614626)
		return "UNBOUNDED PRECEDING"
	case OffsetPreceding:
		__antithesis_instrumentation__.Notify(614627)
		return "OFFSET PRECEDING"
	case CurrentRow:
		__antithesis_instrumentation__.Notify(614628)
		return "CURRENT ROW"
	case OffsetFollowing:
		__antithesis_instrumentation__.Notify(614629)
		return "OFFSET FOLLOWING"
	case UnboundedFollowing:
		__antithesis_instrumentation__.Notify(614630)
		return "UNBOUNDED FOLLOWING"
	default:
		__antithesis_instrumentation__.Notify(614631)
	}
	__antithesis_instrumentation__.Notify(614625)
	return ""
}

type WindowFrameExclusion int

const (
	NoExclusion WindowFrameExclusion = iota

	ExcludeCurrentRow

	ExcludeGroup

	ExcludeTies
)

func (node WindowFrameExclusion) String() string {
	__antithesis_instrumentation__.Notify(614632)
	switch node {
	case NoExclusion:
		__antithesis_instrumentation__.Notify(614633)
		return "EXCLUDE NO ROWS"
	case ExcludeCurrentRow:
		__antithesis_instrumentation__.Notify(614634)
		return "EXCLUDE CURRENT ROW"
	case ExcludeGroup:
		__antithesis_instrumentation__.Notify(614635)
		return "EXCLUDE GROUP"
	case ExcludeTies:
		__antithesis_instrumentation__.Notify(614636)
		return "EXCLUDE TIES"
	default:
		__antithesis_instrumentation__.Notify(614637)
		panic(errors.AssertionFailedf("unhandled case: %d", redact.Safe(node)))
	}
}

func (node WindowFrameExclusion) Name() string {
	__antithesis_instrumentation__.Notify(614638)
	switch node {
	case NoExclusion:
		__antithesis_instrumentation__.Notify(614640)
		return "NoExclusion"
	case ExcludeCurrentRow:
		__antithesis_instrumentation__.Notify(614641)
		return "ExcludeCurrentRow"
	case ExcludeGroup:
		__antithesis_instrumentation__.Notify(614642)
		return "ExcludeGroup"
	case ExcludeTies:
		__antithesis_instrumentation__.Notify(614643)
		return "ExcludeTies"
	default:
		__antithesis_instrumentation__.Notify(614644)
	}
	__antithesis_instrumentation__.Notify(614639)
	return ""
}

func WindowModeName(mode WindowFrameMode) string {
	__antithesis_instrumentation__.Notify(614645)
	switch mode {
	case RANGE:
		__antithesis_instrumentation__.Notify(614646)
		return "RANGE"
	case ROWS:
		__antithesis_instrumentation__.Notify(614647)
		return "ROWS"
	case GROUPS:
		__antithesis_instrumentation__.Notify(614648)
		return "GROUPS"
	default:
		__antithesis_instrumentation__.Notify(614649)
		panic(errors.AssertionFailedf("unhandled case: %d", redact.Safe(mode)))
	}
}
