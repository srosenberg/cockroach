package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var upsertNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(631390)
		return &upsertNode{}
	},
}

type upsertNode struct {
	source planNode

	columns colinfo.ResultColumns

	run upsertRun
}

var _ mutationPlanNode = &upsertNode{}

type upsertRun struct {
	tw        optTableUpserter
	checkOrds checkSet

	insertCols []catalog.Column

	done bool

	traceKV bool
}

func (n *upsertNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631391)

	n.run.traceKV = params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()

	return n.run.tw.init(params.ctx, params.p.txn, params.EvalContext(), &params.EvalContext().Settings.SV)
}

func (n *upsertNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631392)
	panic("not valid")
}

func (n *upsertNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631393)
	panic("not valid")
}

func (n *upsertNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631394)
	if n.run.done {
		__antithesis_instrumentation__.Notify(631399)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(631400)
	}
	__antithesis_instrumentation__.Notify(631395)

	n.run.tw.clearLastBatch(params.ctx)

	lastBatch := false
	for {
		__antithesis_instrumentation__.Notify(631401)
		if err := params.p.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(631405)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631406)
		}
		__antithesis_instrumentation__.Notify(631402)

		if next, err := n.source.Next(params); !next {
			__antithesis_instrumentation__.Notify(631407)
			lastBatch = true
			if err != nil {
				__antithesis_instrumentation__.Notify(631409)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(631410)
			}
			__antithesis_instrumentation__.Notify(631408)
			break
		} else {
			__antithesis_instrumentation__.Notify(631411)
		}
		__antithesis_instrumentation__.Notify(631403)

		if err := n.processSourceRow(params, n.source.Values()); err != nil {
			__antithesis_instrumentation__.Notify(631412)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631413)
		}
		__antithesis_instrumentation__.Notify(631404)

		if n.run.tw.currentBatchSize >= n.run.tw.maxBatchSize {
			__antithesis_instrumentation__.Notify(631414)
			break
		} else {
			__antithesis_instrumentation__.Notify(631415)
		}
	}
	__antithesis_instrumentation__.Notify(631396)

	if n.run.tw.currentBatchSize > 0 {
		__antithesis_instrumentation__.Notify(631416)
		if !lastBatch {
			__antithesis_instrumentation__.Notify(631417)

			if err := n.run.tw.flushAndStartNewBatch(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(631418)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(631419)
			}
		} else {
			__antithesis_instrumentation__.Notify(631420)
		}
	} else {
		__antithesis_instrumentation__.Notify(631421)
	}
	__antithesis_instrumentation__.Notify(631397)

	if lastBatch {
		__antithesis_instrumentation__.Notify(631422)
		n.run.tw.setRowsWrittenLimit(params.extendedEvalCtx.SessionData())
		if err := n.run.tw.finalize(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(631424)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631425)
		}
		__antithesis_instrumentation__.Notify(631423)

		n.run.done = true
	} else {
		__antithesis_instrumentation__.Notify(631426)
	}
	__antithesis_instrumentation__.Notify(631398)

	params.ExecCfg().StatsRefresher.NotifyMutation(n.run.tw.tableDesc(), n.run.tw.lastBatchSize)

	return n.run.tw.lastBatchSize > 0, nil
}

func (n *upsertNode) processSourceRow(params runParams, rowVals tree.Datums) error {
	__antithesis_instrumentation__.Notify(631427)
	if err := enforceLocalColumnConstraints(rowVals, n.run.insertCols); err != nil {
		__antithesis_instrumentation__.Notify(631431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631432)
	}
	__antithesis_instrumentation__.Notify(631428)

	var pm row.PartialIndexUpdateHelper
	if numPartialIndexes := len(n.run.tw.tableDesc().PartialIndexes()); numPartialIndexes > 0 {
		__antithesis_instrumentation__.Notify(631433)
		offset := len(n.run.insertCols) + len(n.run.tw.fetchCols) + len(n.run.tw.updateCols) + n.run.checkOrds.Len()
		if n.run.tw.canaryOrdinal != -1 {
			__antithesis_instrumentation__.Notify(631436)
			offset++
		} else {
			__antithesis_instrumentation__.Notify(631437)
		}
		__antithesis_instrumentation__.Notify(631434)
		partialIndexVals := rowVals[offset:]
		partialIndexPutVals := partialIndexVals[:numPartialIndexes]
		partialIndexDelVals := partialIndexVals[numPartialIndexes : numPartialIndexes*2]

		err := pm.Init(partialIndexPutVals, partialIndexDelVals, n.run.tw.tableDesc())
		if err != nil {
			__antithesis_instrumentation__.Notify(631438)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631439)
		}
		__antithesis_instrumentation__.Notify(631435)

		rowVals = rowVals[:offset]
	} else {
		__antithesis_instrumentation__.Notify(631440)
	}
	__antithesis_instrumentation__.Notify(631429)

	if !n.run.checkOrds.Empty() {
		__antithesis_instrumentation__.Notify(631441)
		ord := len(n.run.insertCols) + len(n.run.tw.fetchCols) + len(n.run.tw.updateCols)
		if n.run.tw.canaryOrdinal != -1 {
			__antithesis_instrumentation__.Notify(631444)
			ord++
		} else {
			__antithesis_instrumentation__.Notify(631445)
		}
		__antithesis_instrumentation__.Notify(631442)
		checkVals := rowVals[ord:]
		if err := checkMutationInput(
			params.ctx, &params.p.semaCtx, params.p.SessionData(), n.run.tw.tableDesc(), n.run.checkOrds, checkVals,
		); err != nil {
			__antithesis_instrumentation__.Notify(631446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631447)
		}
		__antithesis_instrumentation__.Notify(631443)
		rowVals = rowVals[:ord]
	} else {
		__antithesis_instrumentation__.Notify(631448)
	}
	__antithesis_instrumentation__.Notify(631430)

	return n.run.tw.row(params.ctx, rowVals, pm, n.run.traceKV)
}

func (n *upsertNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(631449)
	return n.run.tw.lastBatchSize
}

func (n *upsertNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(631450)
	return n.run.tw.rows.At(rowIdx)
}

func (n *upsertNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(631451)
	n.source.Close(ctx)
	n.run.tw.close(ctx)
	*n = upsertNode{}
	upsertNodePool.Put(n)
}

func (n *upsertNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(631452)
	return n.run.tw.rowsWritten
}

func (n *upsertNode) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(631453)
	n.run.tw.enableAutoCommit()
}
