package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type SurvivalGoal uint32

const (
	SurvivalGoalDefault SurvivalGoal = iota

	SurvivalGoalRegionFailure

	SurvivalGoalZoneFailure
)

func (node *SurvivalGoal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614442)
	switch *node {
	case SurvivalGoalRegionFailure:
		__antithesis_instrumentation__.Notify(614443)
		ctx.WriteString("SURVIVE REGION FAILURE")
	case SurvivalGoalZoneFailure:
		__antithesis_instrumentation__.Notify(614444)
		ctx.WriteString("SURVIVE ZONE FAILURE")
	default:
		__antithesis_instrumentation__.Notify(614445)
		panic(errors.AssertionFailedf("unknown survival goal: %d", *node))
	}
}

func (node *SurvivalGoal) TelemetryName() string {
	__antithesis_instrumentation__.Notify(614446)
	switch *node {
	case SurvivalGoalDefault:
		__antithesis_instrumentation__.Notify(614447)
		return "survive_default"
	case SurvivalGoalRegionFailure:
		__antithesis_instrumentation__.Notify(614448)
		return "survive_region_failure"
	case SurvivalGoalZoneFailure:
		__antithesis_instrumentation__.Notify(614449)
		return "survive_zone_failure"
	default:
		__antithesis_instrumentation__.Notify(614450)
		panic(errors.AssertionFailedf("unknown survival goal: %d", *node))
	}
}
