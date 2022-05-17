package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

var showHistogramColumns = colinfo.ResultColumns{
	{Name: "upper_bound", Typ: types.String},
	{Name: "range_rows", Typ: types.Int},
	{Name: "distinct_range_rows", Typ: types.Float},
	{Name: "equal_rows", Typ: types.Int},
}

func (p *planner) ShowHistogram(ctx context.Context, n *tree.ShowHistogram) (planNode, error) {
	__antithesis_instrumentation__.Notify(623193)
	return &delayedNode{
		name:    fmt.Sprintf("SHOW HISTOGRAM %d", n.HistogramID),
		columns: showHistogramColumns,

		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(623194)
			row, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryRowEx(
				ctx,
				"read-histogram",
				p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				`SELECT histogram
				 FROM system.table_statistics
				 WHERE "statisticID" = $1`,
				n.HistogramID,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(623201)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623202)
			}
			__antithesis_instrumentation__.Notify(623195)
			if row == nil {
				__antithesis_instrumentation__.Notify(623203)
				return nil, fmt.Errorf("histogram %d not found", n.HistogramID)
			} else {
				__antithesis_instrumentation__.Notify(623204)
			}
			__antithesis_instrumentation__.Notify(623196)
			if len(row) != 1 {
				__antithesis_instrumentation__.Notify(623205)
				return nil, errors.AssertionFailedf("expected 1 column from internal query")
			} else {
				__antithesis_instrumentation__.Notify(623206)
			}
			__antithesis_instrumentation__.Notify(623197)
			if row[0] == tree.DNull {
				__antithesis_instrumentation__.Notify(623207)

				return nil, fmt.Errorf("histogram %d not found", n.HistogramID)
			} else {
				__antithesis_instrumentation__.Notify(623208)
			}
			__antithesis_instrumentation__.Notify(623198)

			histogram := &stats.HistogramData{}
			histData := *row[0].(*tree.DBytes)
			if err := protoutil.Unmarshal([]byte(histData), histogram); err != nil {
				__antithesis_instrumentation__.Notify(623209)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623210)
			}
			__antithesis_instrumentation__.Notify(623199)

			v := p.newContainerValuesNode(showHistogramColumns, 0)
			for _, b := range histogram.Buckets {
				__antithesis_instrumentation__.Notify(623211)
				ed, _, err := rowenc.EncDatumFromBuffer(
					histogram.ColumnType, descpb.DatumEncoding_ASCENDING_KEY, b.UpperBound,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(623213)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623214)
				}
				__antithesis_instrumentation__.Notify(623212)
				row := tree.Datums{
					tree.NewDString(ed.String(histogram.ColumnType)),
					tree.NewDInt(tree.DInt(b.NumRange)),
					tree.NewDFloat(tree.DFloat(b.DistinctRange)),
					tree.NewDInt(tree.DInt(b.NumEq)),
				}
				if _, err := v.rows.AddRow(ctx, row); err != nil {
					__antithesis_instrumentation__.Notify(623215)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623216)
				}
			}
			__antithesis_instrumentation__.Notify(623200)
			return v, nil
		},
	}, nil
}
