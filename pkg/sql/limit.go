package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type limitNode struct {
	plan       planNode
	countExpr  tree.TypedExpr
	offsetExpr tree.TypedExpr
}

func (n *limitNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(500219)
	panic("limitNode cannot be run in local mode")
}

func (n *limitNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(500220)
	panic("limitNode cannot be run in local mode")
}

func (n *limitNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(500221)
	panic("limitNode cannot be run in local mode")
}

func (n *limitNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(500222)
	n.plan.Close(ctx)
}

func evalLimit(
	evalCtx *tree.EvalContext, countExpr, offsetExpr tree.TypedExpr,
) (count, offset int64, err error) {
	__antithesis_instrumentation__.Notify(500223)
	count = math.MaxInt64
	offset = 0

	data := []struct {
		name string
		src  tree.TypedExpr
		dst  *int64
	}{
		{"LIMIT", countExpr, &count},
		{"OFFSET", offsetExpr, &offset},
	}

	for _, datum := range data {
		__antithesis_instrumentation__.Notify(500225)
		if datum.src != nil {
			__antithesis_instrumentation__.Notify(500226)
			dstDatum, err := datum.src.Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(500230)
				return count, offset, err
			} else {
				__antithesis_instrumentation__.Notify(500231)
			}
			__antithesis_instrumentation__.Notify(500227)

			if dstDatum == tree.DNull {
				__antithesis_instrumentation__.Notify(500232)

				continue
			} else {
				__antithesis_instrumentation__.Notify(500233)
			}
			__antithesis_instrumentation__.Notify(500228)

			dstDInt := tree.MustBeDInt(dstDatum)
			val := int64(dstDInt)
			if val < 0 {
				__antithesis_instrumentation__.Notify(500234)
				return count, offset, fmt.Errorf("negative value for %s", datum.name)
			} else {
				__antithesis_instrumentation__.Notify(500235)
			}
			__antithesis_instrumentation__.Notify(500229)
			*datum.dst = val
		} else {
			__antithesis_instrumentation__.Notify(500236)
		}
	}
	__antithesis_instrumentation__.Notify(500224)
	return count, offset, nil
}
