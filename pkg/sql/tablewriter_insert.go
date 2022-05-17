package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type tableInserter struct {
	tableWriterBase
	ri row.Inserter
}

var _ tableWriter = &tableInserter{}

func (*tableInserter) desc() string { __antithesis_instrumentation__.Notify(627765); return "inserter" }

func (ti *tableInserter) init(
	_ context.Context, txn *kv.Txn, evalCtx *tree.EvalContext, sv *settings.Values,
) error {
	ti.tableWriterBase.init(txn, ti.tableDesc(), evalCtx, sv)
	return nil
}

func (ti *tableInserter) row(
	ctx context.Context, values tree.Datums, pm row.PartialIndexUpdateHelper, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(627766)
	ti.currentBatchSize++
	return ti.ri.InsertRow(ctx, ti.b, values, pm, false, traceKV)
}

func (ti *tableInserter) tableDesc() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(627767)
	return ti.ri.Helper.TableDesc
}

func (ti *tableInserter) walkExprs(_ func(desc string, index int, expr tree.TypedExpr)) {
	__antithesis_instrumentation__.Notify(627768)
}
