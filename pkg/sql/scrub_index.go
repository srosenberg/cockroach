package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type indexCheckOperation struct {
	tableName *tree.TableName
	tableDesc catalog.TableDescriptor
	index     catalog.Index
	asOf      hlc.Timestamp

	columns []catalog.Column

	primaryColIdxs []int

	run indexCheckRun
}

type indexCheckRun struct {
	started  bool
	rows     []tree.Datums
	rowIndex int
}

func newIndexCheckOperation(
	tableName *tree.TableName,
	tableDesc catalog.TableDescriptor,
	index catalog.Index,
	asOf hlc.Timestamp,
) *indexCheckOperation {
	__antithesis_instrumentation__.Notify(595699)
	return &indexCheckOperation{
		tableName: tableName,
		tableDesc: tableDesc,
		index:     index,
		asOf:      asOf,
	}
}

func (o *indexCheckOperation) Start(params runParams) error {
	__antithesis_instrumentation__.Notify(595700)
	ctx := params.ctx

	var colToIdx catalog.TableColMap
	for _, c := range o.tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(595707)
		colToIdx.Set(c.GetID(), c.Ordinal())
	}
	__antithesis_instrumentation__.Notify(595701)

	var pkColumns, otherColumns []catalog.Column

	for i := 0; i < o.tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(595708)
		colID := o.tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		col := o.tableDesc.PublicColumns()[colToIdx.GetDefault(colID)]
		pkColumns = append(pkColumns, col)
		colToIdx.Set(colID, -1)
	}
	__antithesis_instrumentation__.Notify(595702)

	colIDs := catalog.TableColSet{}
	colIDs.UnionWith(o.index.CollectKeyColumnIDs())
	colIDs.UnionWith(o.index.CollectSecondaryStoredColumnIDs())
	colIDs.UnionWith(o.index.CollectKeySuffixColumnIDs())
	colIDs.ForEach(func(colID descpb.ColumnID) {
		__antithesis_instrumentation__.Notify(595709)
		pos := colToIdx.GetDefault(colID)
		if pos == -1 {
			__antithesis_instrumentation__.Notify(595711)
			return
		} else {
			__antithesis_instrumentation__.Notify(595712)
		}
		__antithesis_instrumentation__.Notify(595710)
		col := o.tableDesc.PublicColumns()[pos]
		otherColumns = append(otherColumns, col)
	})
	__antithesis_instrumentation__.Notify(595703)

	colNames := func(cols []catalog.Column) []string {
		__antithesis_instrumentation__.Notify(595713)
		res := make([]string, len(cols))
		for i := range cols {
			__antithesis_instrumentation__.Notify(595715)
			res[i] = cols[i].GetName()
		}
		__antithesis_instrumentation__.Notify(595714)
		return res
	}
	__antithesis_instrumentation__.Notify(595704)

	checkQuery := createIndexCheckQuery(
		colNames(pkColumns), colNames(otherColumns), o.tableDesc.GetID(), o.index.GetID(), o.tableDesc.GetPrimaryIndexID(),
	)

	rows, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBuffered(
		ctx, "scrub-index", params.p.txn, checkQuery,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595716)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595717)
	}
	__antithesis_instrumentation__.Notify(595705)

	o.run.started = true
	o.run.rows = rows
	o.primaryColIdxs = make([]int, len(pkColumns))
	for i := range o.primaryColIdxs {
		__antithesis_instrumentation__.Notify(595718)
		o.primaryColIdxs[i] = i
	}
	__antithesis_instrumentation__.Notify(595706)
	o.columns = append(pkColumns, otherColumns...)
	return nil
}

func (o *indexCheckOperation) Next(params runParams) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(595719)
	row := o.run.rows[o.run.rowIndex]
	o.run.rowIndex++

	var isMissingIndexReferenceError bool
	if row[o.primaryColIdxs[0]] != tree.DNull {
		__antithesis_instrumentation__.Notify(595725)
		isMissingIndexReferenceError = true
	} else {
		__antithesis_instrumentation__.Notify(595726)
	}
	__antithesis_instrumentation__.Notify(595720)

	colLen := len(o.columns)
	var errorType tree.Datum
	var primaryKeyDatums tree.Datums
	if isMissingIndexReferenceError {
		__antithesis_instrumentation__.Notify(595727)
		errorType = tree.NewDString(scrub.MissingIndexEntryError)

		for _, rowIdx := range o.primaryColIdxs {
			__antithesis_instrumentation__.Notify(595728)
			primaryKeyDatums = append(primaryKeyDatums, row[rowIdx])
		}
	} else {
		__antithesis_instrumentation__.Notify(595729)
		errorType = tree.NewDString(scrub.DanglingIndexReferenceError)

		for _, rowIdx := range o.primaryColIdxs {
			__antithesis_instrumentation__.Notify(595730)
			primaryKeyDatums = append(primaryKeyDatums, row[rowIdx+colLen])
		}
	}
	__antithesis_instrumentation__.Notify(595721)
	primaryKey := tree.NewDString(primaryKeyDatums.String())
	timestamp, err := tree.MakeDTimestamp(
		params.extendedEvalCtx.GetStmtTimestamp(), time.Nanosecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(595731)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595732)
	}
	__antithesis_instrumentation__.Notify(595722)

	details := make(map[string]interface{})
	rowDetails := make(map[string]interface{})
	details["row_data"] = rowDetails
	details["index_name"] = o.index.GetName()
	if isMissingIndexReferenceError {
		__antithesis_instrumentation__.Notify(595733)

		for rowIdx, col := range o.columns {
			__antithesis_instrumentation__.Notify(595734)

			rowDetails[col.GetName()] = row[rowIdx].String()
		}
	} else {
		__antithesis_instrumentation__.Notify(595735)

		for rowIdx, col := range o.columns {
			__antithesis_instrumentation__.Notify(595736)

			rowDetails[col.GetName()] = row[rowIdx+colLen].String()
		}
	}
	__antithesis_instrumentation__.Notify(595723)

	detailsJSON, err := tree.MakeDJSON(details)
	if err != nil {
		__antithesis_instrumentation__.Notify(595737)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595738)
	}
	__antithesis_instrumentation__.Notify(595724)

	return tree.Datums{

		tree.DNull,
		errorType,
		tree.NewDString(o.tableName.Catalog()),
		tree.NewDString(o.tableName.Table()),
		primaryKey,
		timestamp,
		tree.DBoolFalse,
		detailsJSON,
	}, nil
}

func (o *indexCheckOperation) Started() bool {
	__antithesis_instrumentation__.Notify(595739)
	return o.run.started
}

func (o *indexCheckOperation) Done(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(595740)
	return o.run.rows == nil || func() bool {
		__antithesis_instrumentation__.Notify(595741)
		return o.run.rowIndex >= len(o.run.rows) == true
	}() == true
}

func (o *indexCheckOperation) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595742)
	o.run.rows = nil
}

func createIndexCheckQuery(
	pkColumns []string,
	otherColumns []string,
	tableID descpb.ID,
	indexID descpb.IndexID,
	primaryIndexID descpb.IndexID,
) string {
	__antithesis_instrumentation__.Notify(595743)
	allColumns := append(pkColumns, otherColumns...)

	const checkIndexQuery = `
    SELECT %[1]s, %[2]s
    FROM
      (SELECT %[8]s FROM [%[3]d AS table_pri]@{FORCE_INDEX=[%[9]d]}) AS pri
    FULL OUTER JOIN
      (SELECT %[8]s FROM [%[3]d AS table_sec]@{FORCE_INDEX=[%[4]d]}) AS sec
    ON %[5]s
    WHERE %[6]s IS NULL OR %[7]s IS NULL`
	return fmt.Sprintf(
		checkIndexQuery,

		strings.Join(colRefs("pri", allColumns), ", "),

		strings.Join(colRefs("sec", allColumns), ", "),

		tableID,

		indexID,

		strings.Join(
			append(
				pairwiseOp(colRefs("pri", pkColumns), colRefs("sec", pkColumns), "="),
				pairwiseOp(colRefs("pri", otherColumns), colRefs("sec", otherColumns), "IS NOT DISTINCT FROM")...,
			),
			" AND ",
		),

		colRef("pri", pkColumns[0]),

		colRef("sec", pkColumns[0]),

		strings.Join(colRefs("", append(pkColumns, otherColumns...)), ", "),

		primaryIndexID,
	)
}
