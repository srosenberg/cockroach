package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type sqlForeignKeyCheckOperation struct {
	tableName           *tree.TableName
	tableDesc           catalog.TableDescriptor
	referencedTableDesc catalog.TableDescriptor
	constraint          *descpb.ConstraintDetail
	asOf                hlc.Timestamp

	colIDToRowIdx catalog.TableColMap

	run sqlForeignKeyConstraintCheckRun
}

type sqlForeignKeyConstraintCheckRun struct {
	started  bool
	rows     []tree.Datums
	rowIndex int
}

func newSQLForeignKeyCheckOperation(
	tableName *tree.TableName,
	tableDesc catalog.TableDescriptor,
	constraint descpb.ConstraintDetail,
	asOf hlc.Timestamp,
) *sqlForeignKeyCheckOperation {
	__antithesis_instrumentation__.Notify(595642)
	return &sqlForeignKeyCheckOperation{
		tableName:           tableName,
		tableDesc:           tableDesc,
		constraint:          &constraint,
		referencedTableDesc: tabledesc.NewBuilder(constraint.ReferencedTable).BuildImmutableTable(),
		asOf:                asOf,
	}
}

func (o *sqlForeignKeyCheckOperation) Start(params runParams) error {
	__antithesis_instrumentation__.Notify(595643)
	ctx := params.ctx

	checkQuery, _, err := nonMatchingRowQuery(
		o.tableDesc,
		o.constraint.FK,
		o.referencedTableDesc,
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595649)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595650)
	}
	__antithesis_instrumentation__.Notify(595644)

	rows, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBuffered(
		ctx, "scrub-fk", params.p.txn, checkQuery,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595651)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595652)
	}
	__antithesis_instrumentation__.Notify(595645)
	o.run.rows = rows

	if len(o.constraint.FK.OriginColumnIDs) > 1 && func() bool {
		__antithesis_instrumentation__.Notify(595653)
		return o.constraint.FK.Match == descpb.ForeignKeyReference_FULL == true
	}() == true {
		__antithesis_instrumentation__.Notify(595654)

		checkNullsQuery, _, err := matchFullUnacceptableKeyQuery(
			o.tableDesc,
			o.constraint.FK,
			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(595657)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595658)
		}
		__antithesis_instrumentation__.Notify(595655)
		rows, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryBuffered(
			ctx, "scrub-fk", params.p.txn, checkNullsQuery,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(595659)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595660)
		}
		__antithesis_instrumentation__.Notify(595656)
		o.run.rows = append(o.run.rows, rows...)
	} else {
		__antithesis_instrumentation__.Notify(595661)
	}
	__antithesis_instrumentation__.Notify(595646)

	var colIDs []descpb.ColumnID
	colIDs = append(colIDs, o.constraint.FK.OriginColumnIDs...)
	for i := 0; i < o.tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(595662)
		pkColID := o.tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		found := false
		for _, id := range o.constraint.FK.OriginColumnIDs {
			__antithesis_instrumentation__.Notify(595664)
			if pkColID == id {
				__antithesis_instrumentation__.Notify(595665)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(595666)
			}
		}
		__antithesis_instrumentation__.Notify(595663)
		if !found {
			__antithesis_instrumentation__.Notify(595667)
			colIDs = append(colIDs, pkColID)
		} else {
			__antithesis_instrumentation__.Notify(595668)
		}
	}
	__antithesis_instrumentation__.Notify(595647)

	for i, id := range colIDs {
		__antithesis_instrumentation__.Notify(595669)
		o.colIDToRowIdx.Set(id, i)
	}
	__antithesis_instrumentation__.Notify(595648)

	o.run.started = true
	return nil
}

func (o *sqlForeignKeyCheckOperation) Next(params runParams) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(595670)
	row := o.run.rows[o.run.rowIndex]
	o.run.rowIndex++

	details := make(map[string]interface{})
	rowDetails := make(map[string]interface{})
	details["row_data"] = rowDetails
	details["constraint_name"] = o.constraint.FK.Name

	primaryKeyDatums := make(tree.Datums, 0, o.tableDesc.GetPrimaryIndex().NumKeyColumns())
	for i := 0; i < o.tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(595676)
		id := o.tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		idx := o.colIDToRowIdx.GetDefault(id)
		primaryKeyDatums = append(primaryKeyDatums, row[idx])
	}
	__antithesis_instrumentation__.Notify(595671)

	for _, id := range o.constraint.FK.OriginColumnIDs {
		__antithesis_instrumentation__.Notify(595677)
		idx := o.colIDToRowIdx.GetDefault(id)
		col, err := tabledesc.FindPublicColumnWithID(o.tableDesc, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(595679)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(595680)
		}
		__antithesis_instrumentation__.Notify(595678)
		rowDetails[col.GetName()] = row[idx].String()
	}
	__antithesis_instrumentation__.Notify(595672)
	for i := 0; i < o.tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(595681)
		id := o.tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		found := false
		for _, fkID := range o.constraint.FK.OriginColumnIDs {
			__antithesis_instrumentation__.Notify(595683)
			if id == fkID {
				__antithesis_instrumentation__.Notify(595684)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(595685)
			}
		}
		__antithesis_instrumentation__.Notify(595682)
		if !found {
			__antithesis_instrumentation__.Notify(595686)
			idx := o.colIDToRowIdx.GetDefault(id)
			col, err := tabledesc.FindPublicColumnWithID(o.tableDesc, id)
			if err != nil {
				__antithesis_instrumentation__.Notify(595688)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(595689)
			}
			__antithesis_instrumentation__.Notify(595687)
			rowDetails[col.GetName()] = row[idx].String()
		} else {
			__antithesis_instrumentation__.Notify(595690)
		}
	}
	__antithesis_instrumentation__.Notify(595673)

	detailsJSON, err := tree.MakeDJSON(details)
	if err != nil {
		__antithesis_instrumentation__.Notify(595691)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595692)
	}
	__antithesis_instrumentation__.Notify(595674)

	ts, err := tree.MakeDTimestamp(
		params.extendedEvalCtx.GetStmtTimestamp(),
		time.Nanosecond,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(595693)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595694)
	}
	__antithesis_instrumentation__.Notify(595675)

	return tree.Datums{

		tree.DNull,
		tree.NewDString(scrub.ForeignKeyConstraintViolation),
		tree.NewDString(o.tableName.Catalog()),
		tree.NewDString(o.tableName.Table()),
		tree.NewDString(primaryKeyDatums.String()),
		ts,
		tree.DBoolFalse,
		detailsJSON,
	}, nil
}

func (o *sqlForeignKeyCheckOperation) Started() bool {
	__antithesis_instrumentation__.Notify(595695)
	return o.run.started
}

func (o *sqlForeignKeyCheckOperation) Done(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(595696)
	return o.run.rows == nil || func() bool {
		__antithesis_instrumentation__.Notify(595697)
		return o.run.rowIndex >= len(o.run.rows) == true
	}() == true
}

func (o *sqlForeignKeyCheckOperation) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595698)
	o.run.rows = nil
}
