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

type tableUpdater struct {
	tableWriterBase
	ru row.Updater
}

var _ tableWriter = &tableUpdater{}

func (*tableUpdater) desc() string { __antithesis_instrumentation__.Notify(627769); return "updater" }

func (tu *tableUpdater) init(
	_ context.Context, txn *kv.Txn, evalCtx *tree.EvalContext, sv *settings.Values,
) error {
	tu.tableWriterBase.init(txn, tu.tableDesc(), evalCtx, sv)
	return nil
}

func (tu *tableUpdater) row(
	context.Context, tree.Datums, row.PartialIndexUpdateHelper, bool,
) error {
	__antithesis_instrumentation__.Notify(627770)
	panic("unimplemented")
}

func (tu *tableUpdater) rowForUpdate(
	ctx context.Context,
	oldValues, updateValues tree.Datums,
	pm row.PartialIndexUpdateHelper,
	traceKV bool,
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(627771)
	tu.currentBatchSize++
	return tu.ru.UpdateRow(ctx, tu.b, oldValues, updateValues, pm, traceKV)
}

func (tu *tableUpdater) tableDesc() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(627772)
	return tu.ru.Helper.TableDesc
}

func (tu *tableUpdater) walkExprs(_ func(desc string, index int, expr tree.TypedExpr)) {
	__antithesis_instrumentation__.Notify(627773)
}
