package scexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func ExecuteStage(ctx context.Context, deps Dependencies, ops []scop.Op) error {
	__antithesis_instrumentation__.Notify(581612)

	if len(ops) == 0 {
		__antithesis_instrumentation__.Notify(581614)
		log.Infof(ctx, "skipping execution, no operations in this stage")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581615)
	}
	__antithesis_instrumentation__.Notify(581613)
	typ := ops[0].Type()
	switch typ {
	case scop.MutationType:
		__antithesis_instrumentation__.Notify(581616)
		return executeDescriptorMutationOps(ctx, deps, ops)
	case scop.BackfillType:
		__antithesis_instrumentation__.Notify(581617)
		return executeBackfillOps(ctx, deps, ops)
	case scop.ValidationType:
		__antithesis_instrumentation__.Notify(581618)
		return executeValidationOps(ctx, deps, ops)
	default:
		__antithesis_instrumentation__.Notify(581619)
		return errors.AssertionFailedf("unknown ops type %d", typ)
	}
}
