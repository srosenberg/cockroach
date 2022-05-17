package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"go/constant"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type sqlCheckConstraintCheckOperation struct {
	tableName *tree.TableName
	tableDesc catalog.TableDescriptor
	checkDesc *descpb.TableDescriptor_CheckConstraint
	asOf      hlc.Timestamp

	columns []catalog.Column

	primaryColIdxs []int

	run sqlCheckConstraintCheckRun
}

type sqlCheckConstraintCheckRun struct {
	started  bool
	rows     []tree.Datums
	rowIndex int
}

func newSQLCheckConstraintCheckOperation(
	tableName *tree.TableName,
	tableDesc catalog.TableDescriptor,
	checkDesc *descpb.TableDescriptor_CheckConstraint,
	asOf hlc.Timestamp,
) *sqlCheckConstraintCheckOperation {
	__antithesis_instrumentation__.Notify(595616)
	return &sqlCheckConstraintCheckOperation{
		tableName: tableName,
		tableDesc: tableDesc,
		checkDesc: checkDesc,
		asOf:      asOf,
	}
}

func (o *sqlCheckConstraintCheckOperation) Start(params runParams) error {
	__antithesis_instrumentation__.Notify(595617)
	ctx := params.ctx
	expr, err := parser.ParseExpr(o.checkDesc.Expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(595621)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595622)
	}
	__antithesis_instrumentation__.Notify(595618)

	tn := *o.tableName
	tn.ExplicitCatalog = true
	tn.ExplicitSchema = true
	sel := &tree.SelectClause{
		Exprs: tabledesc.ColumnsSelectors(o.tableDesc.PublicColumns()),
		From: tree.From{
			Tables: tree.TableExprs{&tn},
		},
		Where: &tree.Where{
			Type: tree.AstWhere,
			Expr: &tree.NotExpr{Expr: expr},
		},
	}
	if o.asOf != hlc.MaxTimestamp {
		__antithesis_instrumentation__.Notify(595623)
		sel.From.AsOf = tree.AsOfClause{
			Expr: tree.NewNumVal(
				constant.MakeInt64(o.asOf.WallTime),
				"",
				false),
		}
	} else {
		__antithesis_instrumentation__.Notify(595624)
	}
	__antithesis_instrumentation__.Notify(595619)

	rows, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBuffered(
		ctx, "check-constraint", params.p.txn, tree.AsStringWithFlags(sel, tree.FmtParsable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595625)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595626)
	}
	__antithesis_instrumentation__.Notify(595620)

	o.run.started = true
	o.run.rows = rows

	o.columns = o.tableDesc.PublicColumns()

	o.primaryColIdxs, err = getPrimaryColIdxs(o.tableDesc, o.columns)
	return err
}

func (o *sqlCheckConstraintCheckOperation) Next(params runParams) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(595627)
	row := o.run.rows[o.run.rowIndex]
	o.run.rowIndex++
	timestamp, err := tree.MakeDTimestamp(
		params.extendedEvalCtx.GetStmtTimestamp(),
		time.Nanosecond,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595632)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595633)
	}
	__antithesis_instrumentation__.Notify(595628)

	var primaryKeyDatums tree.Datums
	for _, rowIdx := range o.primaryColIdxs {
		__antithesis_instrumentation__.Notify(595634)
		primaryKeyDatums = append(primaryKeyDatums, row[rowIdx])
	}
	__antithesis_instrumentation__.Notify(595629)

	details := make(map[string]interface{})
	rowDetails := make(map[string]interface{})
	details["row_data"] = rowDetails
	details["constraint_name"] = o.checkDesc.Name
	for rowIdx, col := range o.columns {
		__antithesis_instrumentation__.Notify(595635)

		rowDetails[col.GetName()] = row[rowIdx].String()
	}
	__antithesis_instrumentation__.Notify(595630)
	detailsJSON, err := tree.MakeDJSON(details)
	if err != nil {
		__antithesis_instrumentation__.Notify(595636)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595637)
	}
	__antithesis_instrumentation__.Notify(595631)

	return tree.Datums{

		tree.DNull,
		tree.NewDString(scrub.CheckConstraintViolation),
		tree.NewDString(o.tableName.Catalog()),
		tree.NewDString(o.tableName.Table()),
		tree.NewDString(primaryKeyDatums.String()),
		timestamp,
		tree.DBoolFalse,
		detailsJSON,
	}, nil
}

func (o *sqlCheckConstraintCheckOperation) Started() bool {
	__antithesis_instrumentation__.Notify(595638)
	return o.run.started
}

func (o *sqlCheckConstraintCheckOperation) Done(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(595639)
	return o.run.rows == nil || func() bool {
		__antithesis_instrumentation__.Notify(595640)
		return o.run.rowIndex >= len(o.run.rows) == true
	}() == true
}

func (o *sqlCheckConstraintCheckOperation) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595641)
	o.run.rows = nil
}
