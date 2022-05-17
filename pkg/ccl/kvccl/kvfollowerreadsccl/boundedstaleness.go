package kvfollowerreadsccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

func checkBoundedStalenessEnabled(ctx *tree.EvalContext) error {
	__antithesis_instrumentation__.Notify(19595)
	st := ctx.Settings
	return utilccl.CheckEnterpriseEnabled(
		st,
		ctx.ClusterID,
		sql.ClusterOrganization.Get(&st.SV),
		"bounded staleness",
	)
}

func evalMaxStaleness(ctx *tree.EvalContext, d duration.Duration) (time.Time, error) {
	__antithesis_instrumentation__.Notify(19596)
	if err := checkBoundedStalenessEnabled(ctx); err != nil {
		__antithesis_instrumentation__.Notify(19599)
		return time.Time{}, err
	} else {
		__antithesis_instrumentation__.Notify(19600)
	}
	__antithesis_instrumentation__.Notify(19597)
	if d.Compare(duration.FromInt64(0)) < 0 {
		__antithesis_instrumentation__.Notify(19601)
		return time.Time{}, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"interval duration for %s must be greater or equal to 0",
			tree.WithMaxStalenessFunctionName,
		)
	} else {
		__antithesis_instrumentation__.Notify(19602)
	}
	__antithesis_instrumentation__.Notify(19598)
	return duration.Add(ctx.GetStmtTimestamp(), d.Mul(-1)), nil
}

func evalMinTimestamp(ctx *tree.EvalContext, t time.Time) (time.Time, error) {
	__antithesis_instrumentation__.Notify(19603)
	if err := checkBoundedStalenessEnabled(ctx); err != nil {
		__antithesis_instrumentation__.Notify(19606)
		return time.Time{}, err
	} else {
		__antithesis_instrumentation__.Notify(19607)
	}
	__antithesis_instrumentation__.Notify(19604)
	t = t.Round(time.Microsecond)
	if stmtTimestamp := ctx.GetStmtTimestamp().Round(time.Microsecond); t.After(stmtTimestamp) {
		__antithesis_instrumentation__.Notify(19608)
		return time.Time{}, errors.WithDetailf(
			pgerror.Newf(
				pgcode.InvalidParameterValue,
				"timestamp for %s must be less than or equal to statement_timestamp()",
				tree.WithMinTimestampFunctionName,
			),
			"statement timestamp: %d, min_timestamp: %d",
			stmtTimestamp.UnixNano(),
			t.UnixNano(),
		)
	} else {
		__antithesis_instrumentation__.Notify(19609)
	}
	__antithesis_instrumentation__.Notify(19605)
	return t, nil
}

func init() {
	builtins.WithMinTimestamp = evalMinTimestamp
	builtins.WithMaxStaleness = evalMaxStaleness
}
