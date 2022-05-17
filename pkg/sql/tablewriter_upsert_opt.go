package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type optTableUpserter struct {
	tableWriterBase

	ri row.Inserter

	rowsNeeded bool

	colIDToReturnIndex catalog.TableColMap

	insertReorderingRequired bool

	fetchCols []catalog.Column

	updateCols []catalog.Column

	returnCols []catalog.Column

	canaryOrdinal int

	resultRow tree.Datums

	ru row.Updater

	tabColIdxToRetIdx []int
}

var _ tableWriter = &optTableUpserter{}

func (tu *optTableUpserter) init(
	ctx context.Context, txn *kv.Txn, evalCtx *tree.EvalContext, sv *settings.Values,
) error {
	tu.tableWriterBase.init(txn, tu.ri.Helper.TableDesc, evalCtx, sv)

	if tu.rowsNeeded {
		tu.resultRow = make(tree.Datums, len(tu.returnCols))
		tu.rows = rowcontainer.NewRowContainer(
			evalCtx.Mon.MakeBoundAccount(),
			colinfo.ColTypeInfoFromColumns(tu.returnCols),
		)

		tu.colIDToReturnIndex = catalog.ColumnIDToOrdinalMap(tu.tableDesc().PublicColumns())
		if tu.ri.InsertColIDtoRowIndex.Len() == tu.colIDToReturnIndex.Len() {
			for i := range tu.ri.InsertCols {
				colID := tu.ri.InsertCols[i].GetID()
				resultIndex, ok := tu.colIDToReturnIndex.Get(colID)
				if !ok || resultIndex != tu.ri.InsertColIDtoRowIndex.GetDefault(colID) {
					tu.insertReorderingRequired = true
					break
				}
			}
		} else {
			tu.insertReorderingRequired = true
		}
	}

	return nil
}

func (tu *optTableUpserter) makeResultFromRow(
	row tree.Datums, colIDToRowIndex catalog.TableColMap,
) tree.Datums {
	__antithesis_instrumentation__.Notify(627774)
	resultRow := make(tree.Datums, tu.colIDToReturnIndex.Len())
	tu.colIDToReturnIndex.ForEach(func(colID descpb.ColumnID, returnIndex int) {
		__antithesis_instrumentation__.Notify(627776)
		rowIndex, ok := colIDToRowIndex.Get(colID)
		if ok {
			__antithesis_instrumentation__.Notify(627777)
			resultRow[returnIndex] = row[rowIndex]
		} else {
			__antithesis_instrumentation__.Notify(627778)

			resultRow[returnIndex] = tree.DNull
		}
	})
	__antithesis_instrumentation__.Notify(627775)
	return resultRow
}

func (*optTableUpserter) desc() string {
	__antithesis_instrumentation__.Notify(627779)
	return "opt upserter"
}

func (tu *optTableUpserter) row(
	ctx context.Context, row tree.Datums, pm row.PartialIndexUpdateHelper, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(627780)
	tu.currentBatchSize++

	insertEnd := len(tu.ri.InsertCols)
	if tu.canaryOrdinal == -1 {
		__antithesis_instrumentation__.Notify(627784)

		return tu.insertNonConflictingRow(ctx, tu.b, row[:insertEnd], pm, true, traceKV)
	} else {
		__antithesis_instrumentation__.Notify(627785)
	}
	__antithesis_instrumentation__.Notify(627781)
	if row[tu.canaryOrdinal] == tree.DNull {
		__antithesis_instrumentation__.Notify(627786)

		return tu.insertNonConflictingRow(ctx, tu.b, row[:insertEnd], pm, false, traceKV)
	} else {
		__antithesis_instrumentation__.Notify(627787)
	}
	__antithesis_instrumentation__.Notify(627782)

	fetchEnd := insertEnd + len(tu.fetchCols)
	if len(tu.updateCols) == 0 {
		__antithesis_instrumentation__.Notify(627788)
		if !tu.rowsNeeded {
			__antithesis_instrumentation__.Notify(627790)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(627791)
		}
		__antithesis_instrumentation__.Notify(627789)
		_, err := tu.rows.AddRow(ctx, row[insertEnd:fetchEnd])
		return err
	} else {
		__antithesis_instrumentation__.Notify(627792)
	}
	__antithesis_instrumentation__.Notify(627783)

	updateEnd := fetchEnd + len(tu.updateCols)
	return tu.updateConflictingRow(
		ctx,
		tu.b,
		row[insertEnd:fetchEnd],
		row[fetchEnd:updateEnd],
		pm,
		traceKV,
	)
}

func (tu *optTableUpserter) insertNonConflictingRow(
	ctx context.Context,
	b *kv.Batch,
	insertRow tree.Datums,
	pm row.PartialIndexUpdateHelper,
	overwrite, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(627793)

	if err := tu.ri.InsertRow(ctx, b, insertRow, pm, overwrite, traceKV); err != nil {
		__antithesis_instrumentation__.Notify(627798)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627799)
	}
	__antithesis_instrumentation__.Notify(627794)

	if !tu.rowsNeeded {
		__antithesis_instrumentation__.Notify(627800)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(627801)
	}
	__antithesis_instrumentation__.Notify(627795)

	if tu.insertReorderingRequired {
		__antithesis_instrumentation__.Notify(627802)
		tableRow := tu.makeResultFromRow(insertRow, tu.ri.InsertColIDtoRowIndex)

		for tabIdx := range tableRow {
			__antithesis_instrumentation__.Notify(627804)
			if retIdx := tu.tabColIdxToRetIdx[tabIdx]; retIdx >= 0 {
				__antithesis_instrumentation__.Notify(627805)
				tu.resultRow[retIdx] = tableRow[tabIdx]
			} else {
				__antithesis_instrumentation__.Notify(627806)
			}
		}
		__antithesis_instrumentation__.Notify(627803)
		_, err := tu.rows.AddRow(ctx, tu.resultRow)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627807)
	}
	__antithesis_instrumentation__.Notify(627796)

	for tabIdx := range insertRow {
		__antithesis_instrumentation__.Notify(627808)
		if retIdx := tu.tabColIdxToRetIdx[tabIdx]; retIdx >= 0 {
			__antithesis_instrumentation__.Notify(627809)
			tu.resultRow[retIdx] = insertRow[tabIdx]
		} else {
			__antithesis_instrumentation__.Notify(627810)
		}
	}
	__antithesis_instrumentation__.Notify(627797)
	_, err := tu.rows.AddRow(ctx, tu.resultRow)
	return err
}

func (tu *optTableUpserter) updateConflictingRow(
	ctx context.Context,
	b *kv.Batch,
	fetchRow tree.Datums,
	updateValues tree.Datums,
	pm row.PartialIndexUpdateHelper,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(627811)

	if err := enforceLocalColumnConstraints(updateValues, tu.updateCols); err != nil {
		__antithesis_instrumentation__.Notify(627817)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627818)
	}
	__antithesis_instrumentation__.Notify(627812)

	_, err := tu.ru.UpdateRow(ctx, b, fetchRow, updateValues, pm, traceKV)
	if err != nil {
		__antithesis_instrumentation__.Notify(627819)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627820)
	}
	__antithesis_instrumentation__.Notify(627813)

	if !tu.rowsNeeded {
		__antithesis_instrumentation__.Notify(627821)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(627822)
	}
	__antithesis_instrumentation__.Notify(627814)

	tableRow := tu.makeResultFromRow(fetchRow, tu.ru.FetchColIDtoRowIndex)

	tu.colIDToReturnIndex.ForEach(func(colID descpb.ColumnID, returnIndex int) {
		__antithesis_instrumentation__.Notify(627823)

		rowIndex, ok := tu.ru.UpdateColIDtoRowIndex.Get(colID)
		if ok {
			__antithesis_instrumentation__.Notify(627824)
			tableRow[returnIndex] = updateValues[rowIndex]
		} else {
			__antithesis_instrumentation__.Notify(627825)
		}
	})
	__antithesis_instrumentation__.Notify(627815)

	for tabIdx := range tableRow {
		__antithesis_instrumentation__.Notify(627826)
		if retIdx := tu.tabColIdxToRetIdx[tabIdx]; retIdx >= 0 {
			__antithesis_instrumentation__.Notify(627827)
			tu.resultRow[retIdx] = tableRow[tabIdx]
		} else {
			__antithesis_instrumentation__.Notify(627828)
		}
	}
	__antithesis_instrumentation__.Notify(627816)

	_, err = tu.rows.AddRow(ctx, tu.resultRow)
	return err
}

func (tu *optTableUpserter) tableDesc() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(627829)
	return tu.ri.Helper.TableDesc
}

func (tu *optTableUpserter) walkExprs(walk func(desc string, index int, expr tree.TypedExpr)) {
	__antithesis_instrumentation__.Notify(627830)
}
