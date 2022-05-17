package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/errors"
)

type scatterNode struct {
	optColumnsSlot

	run scatterRun
}

func (p *planner) Scatter(ctx context.Context, n *tree.Scatter) (planNode, error) {
	__antithesis_instrumentation__.Notify(576860)
	if !p.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(576864)
		return nil, errorutil.UnsupportedWithMultiTenancy(54255)
	} else {
		__antithesis_instrumentation__.Notify(576865)
	}
	__antithesis_instrumentation__.Notify(576861)

	tableDesc, index, err := p.getTableAndIndex(ctx, &n.TableOrIndex, privilege.INSERT)
	if err != nil {
		__antithesis_instrumentation__.Notify(576866)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(576867)
	}
	__antithesis_instrumentation__.Notify(576862)

	var span roachpb.Span
	if n.From == nil {
		__antithesis_instrumentation__.Notify(576868)

		span = tableDesc.IndexSpan(p.ExecCfg().Codec, index.GetID())
	} else {
		__antithesis_instrumentation__.Notify(576869)
		switch {
		case len(n.From) == 0:
			__antithesis_instrumentation__.Notify(576876)
			return nil, errors.Errorf("no columns in SCATTER FROM expression")
		case len(n.From) > index.NumKeyColumns():
			__antithesis_instrumentation__.Notify(576877)
			return nil, errors.Errorf("too many columns in SCATTER FROM expression")
		case len(n.To) == 0:
			__antithesis_instrumentation__.Notify(576878)
			return nil, errors.Errorf("no columns in SCATTER TO expression")
		case len(n.To) > index.NumKeyColumns():
			__antithesis_instrumentation__.Notify(576879)
			return nil, errors.Errorf("too many columns in SCATTER TO expression")
		default:
			__antithesis_instrumentation__.Notify(576880)
		}
		__antithesis_instrumentation__.Notify(576870)

		desiredTypes := make([]*types.T, index.NumKeyColumns())
		for i := 0; i < index.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(576881)
			colID := index.GetKeyColumnID(i)
			c, err := tableDesc.FindColumnWithID(colID)
			if err != nil {
				__antithesis_instrumentation__.Notify(576883)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576884)
			}
			__antithesis_instrumentation__.Notify(576882)
			desiredTypes[i] = c.GetType()
		}
		__antithesis_instrumentation__.Notify(576871)
		fromVals := make([]tree.Datum, len(n.From))
		for i, expr := range n.From {
			__antithesis_instrumentation__.Notify(576885)
			typedExpr, err := p.analyzeExpr(
				ctx, expr, nil, tree.IndexedVarHelper{}, desiredTypes[i], true, "SCATTER",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(576887)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576888)
			}
			__antithesis_instrumentation__.Notify(576886)
			fromVals[i], err = typedExpr.Eval(p.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(576889)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576890)
			}
		}
		__antithesis_instrumentation__.Notify(576872)
		toVals := make([]tree.Datum, len(n.From))
		for i, expr := range n.To {
			__antithesis_instrumentation__.Notify(576891)
			typedExpr, err := p.analyzeExpr(
				ctx, expr, nil, tree.IndexedVarHelper{}, desiredTypes[i], true, "SCATTER",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(576893)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576894)
			}
			__antithesis_instrumentation__.Notify(576892)
			toVals[i], err = typedExpr.Eval(p.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(576895)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576896)
			}
		}
		__antithesis_instrumentation__.Notify(576873)

		span.Key, err = getRowKey(p.ExecCfg().Codec, tableDesc, index, fromVals)
		if err != nil {
			__antithesis_instrumentation__.Notify(576897)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(576898)
		}
		__antithesis_instrumentation__.Notify(576874)
		span.EndKey, err = getRowKey(p.ExecCfg().Codec, tableDesc, index, toVals)
		if err != nil {
			__antithesis_instrumentation__.Notify(576899)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(576900)
		}
		__antithesis_instrumentation__.Notify(576875)

		if cmp := span.Key.Compare(span.EndKey); cmp > 0 {
			__antithesis_instrumentation__.Notify(576901)
			span.Key, span.EndKey = span.EndKey, span.Key
		} else {
			__antithesis_instrumentation__.Notify(576902)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(576903)

				span.EndKey = span.EndKey.Next()
			} else {
				__antithesis_instrumentation__.Notify(576904)
			}
		}
	}
	__antithesis_instrumentation__.Notify(576863)

	return &scatterNode{
		run: scatterRun{
			span: span,
		},
	}, nil
}

type scatterRun struct {
	span roachpb.Span

	rangeIdx int
	ranges   []roachpb.Span
}

func (n *scatterNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(576905)
	db := params.p.ExecCfg().DB
	req := &roachpb.AdminScatterRequest{
		RequestHeader:   roachpb.RequestHeader{Key: n.run.span.Key, EndKey: n.run.span.EndKey},
		RandomizeLeases: true,
	}
	res, pErr := kv.SendWrapped(params.ctx, db.NonTransactionalSender(), req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(576908)
		return pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(576909)
	}
	__antithesis_instrumentation__.Notify(576906)
	scatterRes := res.(*roachpb.AdminScatterResponse)
	n.run.rangeIdx = -1
	n.run.ranges = make([]roachpb.Span, len(scatterRes.RangeInfos))
	for i, rangeInfo := range scatterRes.RangeInfos {
		__antithesis_instrumentation__.Notify(576910)
		n.run.ranges[i] = roachpb.Span{
			Key:    rangeInfo.Desc.StartKey.AsRawKey(),
			EndKey: rangeInfo.Desc.EndKey.AsRawKey(),
		}
	}
	__antithesis_instrumentation__.Notify(576907)
	return nil
}

func (n *scatterNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(576911)
	n.run.rangeIdx++
	hasNext := n.run.rangeIdx < len(n.run.ranges)
	return hasNext, nil
}

func (n *scatterNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(576912)
	r := n.run.ranges[n.run.rangeIdx]
	return tree.Datums{
		tree.NewDBytes(tree.DBytes(r.Key)),
		tree.NewDString(keys.PrettyPrint(nil, r.Key)),
	}
}

func (*scatterNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(576913) }
