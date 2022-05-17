package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var insertNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(497383)
		return &insertNode{}
	},
}

var tableInserterPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(497384)
		return &tableInserter{}
	},
}

type insertNode struct {
	source planNode

	columns colinfo.ResultColumns

	run insertRun
}

var _ mutationPlanNode = &insertNode{}

type insertRun struct {
	ti         tableInserter
	rowsNeeded bool

	checkOrds checkSet

	insertCols []catalog.Column

	done bool

	resultRowBuffer tree.Datums

	rowIdxToTabColIdx []int

	tabColIdxToRetIdx []int

	traceKV bool
}

func (r *insertRun) initRowContainer(params runParams, columns colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(497385)
	if !r.rowsNeeded {
		__antithesis_instrumentation__.Notify(497388)
		return
	} else {
		__antithesis_instrumentation__.Notify(497389)
	}
	__antithesis_instrumentation__.Notify(497386)
	r.ti.rows = rowcontainer.NewRowContainer(
		params.EvalContext().Mon.MakeBoundAccount(),
		colinfo.ColTypeInfoFromResCols(columns),
	)

	r.resultRowBuffer = make(tree.Datums, len(columns))
	for i := range r.resultRowBuffer {
		__antithesis_instrumentation__.Notify(497390)
		r.resultRowBuffer[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(497387)

	colIDToRetIndex := catalog.ColumnIDToOrdinalMap(r.ti.tableDesc().PublicColumns())
	r.rowIdxToTabColIdx = make([]int, len(r.insertCols))
	for i, col := range r.insertCols {
		__antithesis_instrumentation__.Notify(497391)
		if idx, ok := colIDToRetIndex.Get(col.GetID()); !ok {
			__antithesis_instrumentation__.Notify(497392)

			r.rowIdxToTabColIdx[i] = -1
		} else {
			__antithesis_instrumentation__.Notify(497393)
			r.rowIdxToTabColIdx[i] = idx
		}
	}
}

func (r *insertRun) processSourceRow(params runParams, rowVals tree.Datums) error {
	__antithesis_instrumentation__.Notify(497394)
	if err := enforceLocalColumnConstraints(rowVals, r.insertCols); err != nil {
		__antithesis_instrumentation__.Notify(497400)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497401)
	}
	__antithesis_instrumentation__.Notify(497395)

	var pm row.PartialIndexUpdateHelper
	if n := len(r.ti.tableDesc().PartialIndexes()); n > 0 {
		__antithesis_instrumentation__.Notify(497402)
		offset := len(r.insertCols) + r.checkOrds.Len()
		partialIndexPutVals := rowVals[offset : offset+n]

		err := pm.Init(partialIndexPutVals, tree.Datums{}, r.ti.tableDesc())
		if err != nil {
			__antithesis_instrumentation__.Notify(497404)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497405)
		}
		__antithesis_instrumentation__.Notify(497403)

		rowVals = rowVals[:len(r.insertCols)+r.checkOrds.Len()]
	} else {
		__antithesis_instrumentation__.Notify(497406)
	}
	__antithesis_instrumentation__.Notify(497396)

	if !r.checkOrds.Empty() {
		__antithesis_instrumentation__.Notify(497407)
		checkVals := rowVals[len(r.insertCols):]
		if err := checkMutationInput(
			params.ctx, &params.p.semaCtx, params.p.SessionData(), r.ti.tableDesc(), r.checkOrds, checkVals,
		); err != nil {
			__antithesis_instrumentation__.Notify(497409)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497410)
		}
		__antithesis_instrumentation__.Notify(497408)
		rowVals = rowVals[:len(r.insertCols)]
	} else {
		__antithesis_instrumentation__.Notify(497411)
	}
	__antithesis_instrumentation__.Notify(497397)

	if err := r.ti.row(params.ctx, rowVals, pm, r.traceKV); err != nil {
		__antithesis_instrumentation__.Notify(497412)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497413)
	}
	__antithesis_instrumentation__.Notify(497398)

	if r.ti.rows != nil {
		__antithesis_instrumentation__.Notify(497414)
		for i, val := range rowVals {
			__antithesis_instrumentation__.Notify(497416)

			if tabIdx := r.rowIdxToTabColIdx[i]; tabIdx >= 0 {
				__antithesis_instrumentation__.Notify(497417)
				if retIdx := r.tabColIdxToRetIdx[tabIdx]; retIdx >= 0 {
					__antithesis_instrumentation__.Notify(497418)
					r.resultRowBuffer[retIdx] = val
				} else {
					__antithesis_instrumentation__.Notify(497419)
				}
			} else {
				__antithesis_instrumentation__.Notify(497420)
			}
		}
		__antithesis_instrumentation__.Notify(497415)

		if _, err := r.ti.rows.AddRow(params.ctx, r.resultRowBuffer); err != nil {
			__antithesis_instrumentation__.Notify(497421)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497422)
		}
	} else {
		__antithesis_instrumentation__.Notify(497423)
	}
	__antithesis_instrumentation__.Notify(497399)

	return nil
}

func (n *insertNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(497424)

	n.run.traceKV = params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()

	n.run.initRowContainer(params, n.columns)

	return n.run.ti.init(params.ctx, params.p.txn, params.EvalContext(), &params.EvalContext().Settings.SV)
}

func (n *insertNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(497425)
	panic("not valid")
}

func (n *insertNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(497426)
	panic("not valid")
}

func (n *insertNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(497427)
	if n.run.done {
		__antithesis_instrumentation__.Notify(497432)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(497433)
	}
	__antithesis_instrumentation__.Notify(497428)

	n.run.ti.clearLastBatch(params.ctx)

	lastBatch := false
	for {
		__antithesis_instrumentation__.Notify(497434)
		if err := params.p.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(497438)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(497439)
		}
		__antithesis_instrumentation__.Notify(497435)

		if next, err := n.source.Next(params); !next {
			__antithesis_instrumentation__.Notify(497440)
			lastBatch = true
			if err != nil {
				__antithesis_instrumentation__.Notify(497442)

				err = interceptAlterColumnTypeParseError(n.run.insertCols, -1, err)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(497443)
			}
			__antithesis_instrumentation__.Notify(497441)
			break
		} else {
			__antithesis_instrumentation__.Notify(497444)
		}
		__antithesis_instrumentation__.Notify(497436)

		if err := n.run.processSourceRow(params, n.source.Values()); err != nil {
			__antithesis_instrumentation__.Notify(497445)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(497446)
		}
		__antithesis_instrumentation__.Notify(497437)

		if n.run.ti.currentBatchSize >= n.run.ti.maxBatchSize || func() bool {
			__antithesis_instrumentation__.Notify(497447)
			return n.run.ti.b.ApproximateMutationBytes() >= n.run.ti.maxBatchByteSize == true
		}() == true {
			__antithesis_instrumentation__.Notify(497448)
			break
		} else {
			__antithesis_instrumentation__.Notify(497449)
		}
	}
	__antithesis_instrumentation__.Notify(497429)

	if n.run.ti.currentBatchSize > 0 {
		__antithesis_instrumentation__.Notify(497450)
		if !lastBatch {
			__antithesis_instrumentation__.Notify(497451)

			if err := n.run.ti.flushAndStartNewBatch(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(497452)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(497453)
			}
		} else {
			__antithesis_instrumentation__.Notify(497454)
		}
	} else {
		__antithesis_instrumentation__.Notify(497455)
	}
	__antithesis_instrumentation__.Notify(497430)

	if lastBatch {
		__antithesis_instrumentation__.Notify(497456)
		n.run.ti.setRowsWrittenLimit(params.extendedEvalCtx.SessionData())
		if err := n.run.ti.finalize(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(497458)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(497459)
		}
		__antithesis_instrumentation__.Notify(497457)

		n.run.done = true
	} else {
		__antithesis_instrumentation__.Notify(497460)
	}
	__antithesis_instrumentation__.Notify(497431)

	params.ExecCfg().StatsRefresher.NotifyMutation(n.run.ti.tableDesc(), n.run.ti.lastBatchSize)

	return n.run.ti.lastBatchSize > 0, nil
}

func (n *insertNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(497461)
	return n.run.ti.lastBatchSize
}

func (n *insertNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(497462)
	return n.run.ti.rows.At(rowIdx)
}

func (n *insertNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(497463)
	n.source.Close(ctx)
	n.run.ti.close(ctx)
	*n = insertNode{}
	insertNodePool.Put(n)
}

func (n *insertNode) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(497464)
	n.run.ti.enableAutoCommit()
}

func (n *insertNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(497465)
	return n.run.ti.rowsWritten
}
