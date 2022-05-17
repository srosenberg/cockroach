package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type renderNode struct {
	_ util.NoCopy

	source planDataSource

	ivarHelper tree.IndexedVarHelper

	render []tree.TypedExpr

	columns colinfo.ResultColumns

	serialize bool

	reqOrdering ReqOrdering
}

func (r *renderNode) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(566111)
	panic("renderNode can't be run in local mode")
}

func (r *renderNode) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(566112)
	return r.source.columns[idx].Typ
}

func (r *renderNode) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(566113)
	return r.source.columns.NodeFormatter(idx)
}

func (r *renderNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(566114)
	panic("renderNode can't be run in local mode")
}

func (r *renderNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(566115)
	panic("renderNode can't be run in local mode")
}

func (r *renderNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(566116)
	panic("renderNode can't be run in local mode")
}

func (r *renderNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(566117)
	r.source.plan.Close(ctx)
}

func (p *planner) getTimestamp(
	ctx context.Context, asOfClause tree.AsOfClause,
) (hlc.Timestamp, bool, error) {
	__antithesis_instrumentation__.Notify(566118)
	if asOfClause.Expr != nil {
		__antithesis_instrumentation__.Notify(566120)

		if p.EvalContext().AsOfSystemTime == nil {
			__antithesis_instrumentation__.Notify(566124)
			return hlc.MaxTimestamp, false,
				pgerror.Newf(pgcode.Syntax,
					"AS OF SYSTEM TIME must be provided on a top-level statement")
		} else {
			__antithesis_instrumentation__.Notify(566125)
		}
		__antithesis_instrumentation__.Notify(566121)

		asOf, err := p.EvalAsOfTimestamp(ctx, asOfClause)
		if err != nil {
			__antithesis_instrumentation__.Notify(566126)
			return hlc.MaxTimestamp, false, err
		} else {
			__antithesis_instrumentation__.Notify(566127)
		}
		__antithesis_instrumentation__.Notify(566122)

		if asOf != *p.EvalContext().AsOfSystemTime && func() bool {
			__antithesis_instrumentation__.Notify(566128)
			return p.EvalContext().AsOfSystemTime.MaxTimestampBound.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(566129)
			return hlc.MaxTimestamp, false,
				unimplemented.NewWithIssue(35712,
					"cannot specify AS OF SYSTEM TIME with different timestamps")
		} else {
			__antithesis_instrumentation__.Notify(566130)
		}
		__antithesis_instrumentation__.Notify(566123)
		return asOf.Timestamp, true, nil
	} else {
		__antithesis_instrumentation__.Notify(566131)
	}
	__antithesis_instrumentation__.Notify(566119)
	return hlc.MaxTimestamp, false, nil
}
