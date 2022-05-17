package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func (p *planner) SetTransaction(ctx context.Context, n *tree.SetTransaction) (planNode, error) {
	__antithesis_instrumentation__.Notify(622044)
	var asOfTs hlc.Timestamp
	if n.Modes.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(622048)
		asOf, err := p.EvalAsOfTimestamp(ctx, n.Modes.AsOf)
		if err != nil {
			__antithesis_instrumentation__.Notify(622050)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(622051)
		}
		__antithesis_instrumentation__.Notify(622049)
		p.extendedEvalCtx.AsOfSystemTime = &asOf
		asOfTs = asOf.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(622052)
	}
	__antithesis_instrumentation__.Notify(622045)
	if n.Modes.Deferrable == tree.Deferrable {
		__antithesis_instrumentation__.Notify(622053)
		return nil, unimplemented.NewWithIssue(53432, "DEFERRABLE transactions")
	} else {
		__antithesis_instrumentation__.Notify(622054)
	}
	__antithesis_instrumentation__.Notify(622046)

	if err := p.extendedEvalCtx.TxnModesSetter.setTransactionModes(ctx, n.Modes, asOfTs); err != nil {
		__antithesis_instrumentation__.Notify(622055)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622056)
	}
	__antithesis_instrumentation__.Notify(622047)
	return newZeroNode(nil), nil
}
