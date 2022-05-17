package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

var updateNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(631290)
		return &updateNode{}
	},
}

type updateNode struct {
	source planNode

	columns colinfo.ResultColumns

	run updateRun
}

var _ mutationPlanNode = &updateNode{}

type updateRun struct {
	tu         tableUpdater
	rowsNeeded bool

	checkOrds checkSet

	done bool

	traceKV bool

	computedCols []catalog.Column

	computeExprs []tree.TypedExpr

	iVarContainerForComputedCols schemaexpr.RowIndexedVarContainer

	sourceSlots []sourceSlot

	updateValues tree.Datums

	updateColsIdx catalog.TableColMap

	rowIdxToRetIdx []int

	numPassthrough int
}

func (u *updateNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631291)

	u.run.traceKV = params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()

	if u.run.rowsNeeded {
		__antithesis_instrumentation__.Notify(631293)
		u.run.tu.rows = rowcontainer.NewRowContainer(
			params.EvalContext().Mon.MakeBoundAccount(),
			colinfo.ColTypeInfoFromResCols(u.columns),
		)
	} else {
		__antithesis_instrumentation__.Notify(631294)
	}
	__antithesis_instrumentation__.Notify(631292)
	return u.run.tu.init(params.ctx, params.p.txn, params.EvalContext(), &params.EvalContext().Settings.SV)
}

func (u *updateNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631295)
	panic("not valid")
}

func (u *updateNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631296)
	panic("not valid")
}

func (u *updateNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631297)
	if u.run.done {
		__antithesis_instrumentation__.Notify(631302)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(631303)
	}
	__antithesis_instrumentation__.Notify(631298)

	u.run.tu.clearLastBatch(params.ctx)

	lastBatch := false
	for {
		__antithesis_instrumentation__.Notify(631304)
		if err := params.p.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(631308)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631309)
		}
		__antithesis_instrumentation__.Notify(631305)

		if next, err := u.source.Next(params); !next {
			__antithesis_instrumentation__.Notify(631310)
			lastBatch = true
			if err != nil {
				__antithesis_instrumentation__.Notify(631312)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(631313)
			}
			__antithesis_instrumentation__.Notify(631311)
			break
		} else {
			__antithesis_instrumentation__.Notify(631314)
		}
		__antithesis_instrumentation__.Notify(631306)

		if err := u.processSourceRow(params, u.source.Values()); err != nil {
			__antithesis_instrumentation__.Notify(631315)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631316)
		}
		__antithesis_instrumentation__.Notify(631307)

		if u.run.tu.currentBatchSize >= u.run.tu.maxBatchSize || func() bool {
			__antithesis_instrumentation__.Notify(631317)
			return u.run.tu.b.ApproximateMutationBytes() >= u.run.tu.maxBatchByteSize == true
		}() == true {
			__antithesis_instrumentation__.Notify(631318)
			break
		} else {
			__antithesis_instrumentation__.Notify(631319)
		}
	}
	__antithesis_instrumentation__.Notify(631299)

	if u.run.tu.currentBatchSize > 0 {
		__antithesis_instrumentation__.Notify(631320)
		if !lastBatch {
			__antithesis_instrumentation__.Notify(631321)

			if err := u.run.tu.flushAndStartNewBatch(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(631322)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(631323)
			}
		} else {
			__antithesis_instrumentation__.Notify(631324)
		}
	} else {
		__antithesis_instrumentation__.Notify(631325)
	}
	__antithesis_instrumentation__.Notify(631300)

	if lastBatch {
		__antithesis_instrumentation__.Notify(631326)
		u.run.tu.setRowsWrittenLimit(params.extendedEvalCtx.SessionData())
		if err := u.run.tu.finalize(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(631328)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(631329)
		}
		__antithesis_instrumentation__.Notify(631327)

		u.run.done = true
	} else {
		__antithesis_instrumentation__.Notify(631330)
	}
	__antithesis_instrumentation__.Notify(631301)

	params.ExecCfg().StatsRefresher.NotifyMutation(u.run.tu.tableDesc(), u.run.tu.lastBatchSize)

	return u.run.tu.lastBatchSize > 0, nil
}

func (u *updateNode) processSourceRow(params runParams, sourceVals tree.Datums) error {
	__antithesis_instrumentation__.Notify(631331)

	oldValues := sourceVals[:len(u.run.tu.ru.FetchCols)]

	valueIdx := 0

	for _, slot := range u.run.sourceSlots {
		__antithesis_instrumentation__.Notify(631339)
		for _, value := range slot.extractValues(sourceVals) {
			__antithesis_instrumentation__.Notify(631340)
			u.run.updateValues[valueIdx] = value
			valueIdx++
		}
	}
	__antithesis_instrumentation__.Notify(631332)

	if len(u.run.computeExprs) > 0 {
		__antithesis_instrumentation__.Notify(631341)

		copy(u.run.iVarContainerForComputedCols.CurSourceRow, oldValues)
		for i := range u.run.tu.ru.UpdateCols {
			__antithesis_instrumentation__.Notify(631344)
			id := u.run.tu.ru.UpdateCols[i].GetID()
			idx := u.run.tu.ru.FetchColIDtoRowIndex.GetDefault(id)
			u.run.iVarContainerForComputedCols.CurSourceRow[idx] = u.run.
				updateValues[i]
		}
		__antithesis_instrumentation__.Notify(631342)

		params.EvalContext().PushIVarContainer(&u.run.iVarContainerForComputedCols)
		for i := range u.run.computedCols {
			__antithesis_instrumentation__.Notify(631345)
			d, err := u.run.computeExprs[i].Eval(params.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(631347)
				params.EvalContext().IVarContainer = nil
				name := u.run.computedCols[i].GetName()
				return errors.Wrapf(err, "computed column %s", tree.ErrString((*tree.Name)(&name)))
			} else {
				__antithesis_instrumentation__.Notify(631348)
			}
			__antithesis_instrumentation__.Notify(631346)
			idx := u.run.updateColsIdx.GetDefault(u.run.computedCols[i].GetID())
			u.run.updateValues[idx] = d
		}
		__antithesis_instrumentation__.Notify(631343)
		params.EvalContext().PopIVarContainer()
	} else {
		__antithesis_instrumentation__.Notify(631349)
	}
	__antithesis_instrumentation__.Notify(631333)

	if err := enforceLocalColumnConstraints(u.run.updateValues, u.run.tu.ru.UpdateCols); err != nil {
		__antithesis_instrumentation__.Notify(631350)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631351)
	}
	__antithesis_instrumentation__.Notify(631334)

	if !u.run.checkOrds.Empty() {
		__antithesis_instrumentation__.Notify(631352)
		checkVals := sourceVals[len(u.run.tu.ru.FetchCols)+len(u.run.tu.ru.UpdateCols)+u.run.numPassthrough:]
		if err := checkMutationInput(
			params.ctx, &params.p.semaCtx, params.p.SessionData(), u.run.tu.tableDesc(), u.run.checkOrds, checkVals,
		); err != nil {
			__antithesis_instrumentation__.Notify(631353)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631354)
		}
	} else {
		__antithesis_instrumentation__.Notify(631355)
	}
	__antithesis_instrumentation__.Notify(631335)

	var pm row.PartialIndexUpdateHelper
	if n := len(u.run.tu.tableDesc().PartialIndexes()); n > 0 {
		__antithesis_instrumentation__.Notify(631356)
		offset := len(u.run.tu.ru.FetchCols) + len(u.run.tu.ru.UpdateCols) + u.run.checkOrds.Len() + u.run.numPassthrough
		partialIndexVals := sourceVals[offset:]
		partialIndexPutVals := partialIndexVals[:n]
		partialIndexDelVals := partialIndexVals[n : n*2]

		err := pm.Init(partialIndexPutVals, partialIndexDelVals, u.run.tu.tableDesc())
		if err != nil {
			__antithesis_instrumentation__.Notify(631357)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631358)
		}
	} else {
		__antithesis_instrumentation__.Notify(631359)
	}
	__antithesis_instrumentation__.Notify(631336)

	newValues, err := u.run.tu.rowForUpdate(params.ctx, oldValues, u.run.updateValues, pm, u.run.traceKV)
	if err != nil {
		__antithesis_instrumentation__.Notify(631360)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631361)
	}
	__antithesis_instrumentation__.Notify(631337)

	if u.run.tu.rows != nil {
		__antithesis_instrumentation__.Notify(631362)

		resultValues := make([]tree.Datum, len(u.columns))
		largestRetIdx := -1
		for i := range u.run.rowIdxToRetIdx {
			__antithesis_instrumentation__.Notify(631365)
			retIdx := u.run.rowIdxToRetIdx[i]
			if retIdx >= 0 {
				__antithesis_instrumentation__.Notify(631366)
				if retIdx >= largestRetIdx {
					__antithesis_instrumentation__.Notify(631368)
					largestRetIdx = retIdx
				} else {
					__antithesis_instrumentation__.Notify(631369)
				}
				__antithesis_instrumentation__.Notify(631367)
				resultValues[retIdx] = newValues[i]
			} else {
				__antithesis_instrumentation__.Notify(631370)
			}
		}
		__antithesis_instrumentation__.Notify(631363)

		if u.run.numPassthrough > 0 {
			__antithesis_instrumentation__.Notify(631371)
			passthroughBegin := len(u.run.tu.ru.FetchCols) + len(u.run.tu.ru.UpdateCols)
			passthroughEnd := passthroughBegin + u.run.numPassthrough
			passthroughValues := sourceVals[passthroughBegin:passthroughEnd]

			for i := 0; i < u.run.numPassthrough; i++ {
				__antithesis_instrumentation__.Notify(631372)
				largestRetIdx++
				resultValues[largestRetIdx] = passthroughValues[i]
			}
		} else {
			__antithesis_instrumentation__.Notify(631373)
		}
		__antithesis_instrumentation__.Notify(631364)

		if _, err := u.run.tu.rows.AddRow(params.ctx, resultValues); err != nil {
			__antithesis_instrumentation__.Notify(631374)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631375)
		}
	} else {
		__antithesis_instrumentation__.Notify(631376)
	}
	__antithesis_instrumentation__.Notify(631338)

	return nil
}

func (u *updateNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(631377)
	return u.run.tu.lastBatchSize
}

func (u *updateNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(631378)
	return u.run.tu.rows.At(rowIdx)
}

func (u *updateNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(631379)
	u.source.Close(ctx)
	u.run.tu.close(ctx)
	*u = updateNode{}
	updateNodePool.Put(u)
}

func (u *updateNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(631380)
	return u.run.tu.rowsWritten
}

func (u *updateNode) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(631381)
	u.run.tu.enableAutoCommit()
}

type sourceSlot interface {
	extractValues(resultRow tree.Datums) tree.Datums

	checkColumnTypes(row []tree.TypedExpr) error
}

type scalarSlot struct {
	column      catalog.Column
	sourceIndex int
}

func (ss scalarSlot) extractValues(row tree.Datums) tree.Datums {
	__antithesis_instrumentation__.Notify(631382)
	return row[ss.sourceIndex : ss.sourceIndex+1]
}

func (ss scalarSlot) checkColumnTypes(row []tree.TypedExpr) error {
	__antithesis_instrumentation__.Notify(631383)
	renderedResult := row[ss.sourceIndex]
	typ := renderedResult.ResolvedType()
	return colinfo.CheckDatumTypeFitsColumnType(ss.column, typ)
}

func enforceLocalColumnConstraints(row tree.Datums, cols []catalog.Column) error {
	__antithesis_instrumentation__.Notify(631384)
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(631386)
		if !col.IsNullable() && func() bool {
			__antithesis_instrumentation__.Notify(631387)
			return row[i] == tree.DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(631388)
			return sqlerrors.NewNonNullViolationError(col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(631389)
		}
	}
	__antithesis_instrumentation__.Notify(631385)
	return nil
}
