package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var deleteNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(465891)
		return &deleteNode{}
	},
}

type deleteNode struct {
	source planNode

	columns colinfo.ResultColumns

	run deleteRun
}

type deleteRun struct {
	td         tableDeleter
	rowsNeeded bool

	done bool

	traceKV bool

	partialIndexDelValsOffset int

	rowIdxToRetIdx []int
}

var _ mutationPlanNode = &deleteNode{}

func (d *deleteNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(465892)

	d.run.traceKV = params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()

	if d.run.rowsNeeded {
		__antithesis_instrumentation__.Notify(465894)
		d.run.td.rows = rowcontainer.NewRowContainer(
			params.EvalContext().Mon.MakeBoundAccount(),
			colinfo.ColTypeInfoFromResCols(d.columns))
	} else {
		__antithesis_instrumentation__.Notify(465895)
	}
	__antithesis_instrumentation__.Notify(465893)
	return d.run.td.init(params.ctx, params.p.txn, params.EvalContext(), &params.EvalContext().Settings.SV)
}

func (d *deleteNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465896)
	panic("not valid")
}

func (d *deleteNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(465897)
	panic("not valid")
}

func (d *deleteNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465898)
	if d.run.done {
		__antithesis_instrumentation__.Notify(465903)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(465904)
	}
	__antithesis_instrumentation__.Notify(465899)

	d.run.td.clearLastBatch(params.ctx)

	lastBatch := false
	for {
		__antithesis_instrumentation__.Notify(465905)
		if err := params.p.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(465909)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(465910)
		}
		__antithesis_instrumentation__.Notify(465906)

		if next, err := d.source.Next(params); !next {
			__antithesis_instrumentation__.Notify(465911)
			lastBatch = true
			if err != nil {
				__antithesis_instrumentation__.Notify(465913)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(465914)
			}
			__antithesis_instrumentation__.Notify(465912)
			break
		} else {
			__antithesis_instrumentation__.Notify(465915)
		}
		__antithesis_instrumentation__.Notify(465907)

		if err := d.processSourceRow(params, d.source.Values()); err != nil {
			__antithesis_instrumentation__.Notify(465916)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(465917)
		}
		__antithesis_instrumentation__.Notify(465908)

		if d.run.td.currentBatchSize >= d.run.td.maxBatchSize || func() bool {
			__antithesis_instrumentation__.Notify(465918)
			return d.run.td.b.ApproximateMutationBytes() >= d.run.td.maxBatchByteSize == true
		}() == true {
			__antithesis_instrumentation__.Notify(465919)
			break
		} else {
			__antithesis_instrumentation__.Notify(465920)
		}
	}
	__antithesis_instrumentation__.Notify(465900)

	if d.run.td.currentBatchSize > 0 {
		__antithesis_instrumentation__.Notify(465921)
		if !lastBatch {
			__antithesis_instrumentation__.Notify(465922)

			if err := d.run.td.flushAndStartNewBatch(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(465923)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(465924)
			}
		} else {
			__antithesis_instrumentation__.Notify(465925)
		}
	} else {
		__antithesis_instrumentation__.Notify(465926)
	}
	__antithesis_instrumentation__.Notify(465901)

	if lastBatch {
		__antithesis_instrumentation__.Notify(465927)
		d.run.td.setRowsWrittenLimit(params.extendedEvalCtx.SessionData())
		if err := d.run.td.finalize(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(465929)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(465930)
		}
		__antithesis_instrumentation__.Notify(465928)

		d.run.done = true
	} else {
		__antithesis_instrumentation__.Notify(465931)
	}
	__antithesis_instrumentation__.Notify(465902)

	params.ExecCfg().StatsRefresher.NotifyMutation(d.run.td.tableDesc(), d.run.td.lastBatchSize)

	return d.run.td.lastBatchSize > 0, nil
}

func (d *deleteNode) processSourceRow(params runParams, sourceVals tree.Datums) error {
	__antithesis_instrumentation__.Notify(465932)

	var pm row.PartialIndexUpdateHelper
	if n := len(d.run.td.tableDesc().PartialIndexes()); n > 0 {
		__antithesis_instrumentation__.Notify(465936)
		offset := d.run.partialIndexDelValsOffset
		partialIndexDelVals := sourceVals[offset : offset+n]

		err := pm.Init(tree.Datums{}, partialIndexDelVals, d.run.td.tableDesc())
		if err != nil {
			__antithesis_instrumentation__.Notify(465938)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465939)
		}
		__antithesis_instrumentation__.Notify(465937)

		sourceVals = sourceVals[:d.run.partialIndexDelValsOffset]
	} else {
		__antithesis_instrumentation__.Notify(465940)
	}
	__antithesis_instrumentation__.Notify(465933)

	if err := d.run.td.row(params.ctx, sourceVals, pm, d.run.traceKV); err != nil {
		__antithesis_instrumentation__.Notify(465941)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465942)
	}
	__antithesis_instrumentation__.Notify(465934)

	if d.run.td.rows != nil {
		__antithesis_instrumentation__.Notify(465943)

		resultValues := make(tree.Datums, d.run.td.rows.NumCols())
		for i, retIdx := range d.run.rowIdxToRetIdx {
			__antithesis_instrumentation__.Notify(465945)
			if retIdx >= 0 {
				__antithesis_instrumentation__.Notify(465946)
				resultValues[retIdx] = sourceVals[i]
			} else {
				__antithesis_instrumentation__.Notify(465947)
			}
		}
		__antithesis_instrumentation__.Notify(465944)

		if _, err := d.run.td.rows.AddRow(params.ctx, resultValues); err != nil {
			__antithesis_instrumentation__.Notify(465948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465949)
		}
	} else {
		__antithesis_instrumentation__.Notify(465950)
	}
	__antithesis_instrumentation__.Notify(465935)

	return nil
}

func (d *deleteNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(465951)
	return d.run.td.lastBatchSize
}

func (d *deleteNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(465952)
	return d.run.td.rows.At(rowIdx)
}

func (d *deleteNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(465953)
	d.source.Close(ctx)
	d.run.td.close(ctx)
	*d = deleteNode{}
	deleteNodePool.Put(d)
}

func (d *deleteNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(465954)
	return d.run.td.rowsWritten
}

func (d *deleteNode) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(465955)
	d.run.td.enableAutoCommit()
}
