package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func InsertNewStats(
	ctx context.Context,
	settings *cluster.Settings,
	executor sqlutil.InternalExecutor,
	txn *kv.Txn,
	tableStats []*TableStatisticProto,
) error {
	__antithesis_instrumentation__.Notify(626713)
	var err error
	for _, statistic := range tableStats {
		__antithesis_instrumentation__.Notify(626715)
		err = InsertNewStat(
			ctx,
			settings,
			executor,
			txn,
			statistic.TableID,
			statistic.Name,
			statistic.ColumnIDs,
			int64(statistic.RowCount),
			int64(statistic.DistinctCount),
			int64(statistic.NullCount),
			int64(statistic.AvgSize),
			statistic.HistogramData,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(626716)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626717)
		}
	}
	__antithesis_instrumentation__.Notify(626714)
	return nil
}

func InsertNewStat(
	ctx context.Context,
	settings *cluster.Settings,
	executor sqlutil.InternalExecutor,
	txn *kv.Txn,
	tableID descpb.ID,
	name string,
	columnIDs []descpb.ColumnID,
	rowCount, distinctCount, nullCount, avgSize int64,
	h *HistogramData,
) error {
	__antithesis_instrumentation__.Notify(626718)

	var nameVal, histogramVal interface{}
	if name != "" {
		__antithesis_instrumentation__.Notify(626723)
		nameVal = name
	} else {
		__antithesis_instrumentation__.Notify(626724)
	}
	__antithesis_instrumentation__.Notify(626719)
	if h != nil {
		__antithesis_instrumentation__.Notify(626725)
		var err error
		histogramVal, err = protoutil.Marshal(h)
		if err != nil {
			__antithesis_instrumentation__.Notify(626726)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626727)
		}
	} else {
		__antithesis_instrumentation__.Notify(626728)
	}
	__antithesis_instrumentation__.Notify(626720)

	columnIDsVal := tree.NewDArray(types.Int)
	for _, c := range columnIDs {
		__antithesis_instrumentation__.Notify(626729)
		if err := columnIDsVal.Append(tree.NewDInt(tree.DInt(int(c)))); err != nil {
			__antithesis_instrumentation__.Notify(626730)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626731)
		}
	}
	__antithesis_instrumentation__.Notify(626721)
	if !settings.Version.IsActive(ctx, clusterversion.AlterSystemTableStatisticsAddAvgSizeCol) {
		__antithesis_instrumentation__.Notify(626732)
		_, err := executor.Exec(
			ctx, "insert-statistic", txn,
			`INSERT INTO system.table_statistics (
					"tableID",
					"name",
					"columnIDs",
					"rowCount",
					"distinctCount",
					"nullCount",
					histogram
				) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tableID,
			nameVal,
			columnIDsVal,
			rowCount,
			distinctCount,
			nullCount,
			histogramVal,
		)
		return err
	} else {
		__antithesis_instrumentation__.Notify(626733)
	}
	__antithesis_instrumentation__.Notify(626722)
	_, err := executor.Exec(
		ctx, "insert-statistic", txn,
		`INSERT INTO system.table_statistics (
					"tableID",
					"name",
					"columnIDs",
					"rowCount",
					"distinctCount",
					"nullCount",
					"avgSize",
					histogram
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tableID,
		nameVal,
		columnIDsVal,
		rowCount,
		distinctCount,
		nullCount,
		avgSize,
		histogramVal,
	)
	return err
}
