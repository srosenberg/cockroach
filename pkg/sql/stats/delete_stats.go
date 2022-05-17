package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

const (
	keepCount = 4
)

func DeleteOldStatsForColumns(
	ctx context.Context,
	executor sqlutil.InternalExecutor,
	txn *kv.Txn,
	tableID descpb.ID,
	columnIDs []descpb.ColumnID,
) error {
	__antithesis_instrumentation__.Notify(626156)
	columnIDsVal := tree.NewDArray(types.Int)
	for _, c := range columnIDs {
		__antithesis_instrumentation__.Notify(626158)
		if err := columnIDsVal.Append(tree.NewDInt(tree.DInt(int(c)))); err != nil {
			__antithesis_instrumentation__.Notify(626159)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626160)
		}
	}
	__antithesis_instrumentation__.Notify(626157)

	_, err := executor.Exec(
		ctx, "delete-statistics", txn,
		`DELETE FROM system.table_statistics
               WHERE "tableID" = $1
               AND "columnIDs" = $3
               AND "statisticID" NOT IN (
                   SELECT "statisticID" FROM system.table_statistics
                   WHERE "tableID" = $1
                   AND "name" = $2
                   AND "columnIDs" = $3
                   ORDER BY "createdAt" DESC
                   LIMIT $4
               )`,
		tableID,
		jobspb.AutoStatsName,
		columnIDsVal,
		keepCount,
	)
	return err
}
