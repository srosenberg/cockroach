package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
)

type StatsCompactor struct {
	st *cluster.Settings
	db *kv.DB
	ie sqlutil.InternalExecutor

	rowsRemovedCounter *metric.Counter

	knobs *sqlstats.TestingKnobs

	scratch struct {
		qargs []interface{}
	}
}

func NewStatsCompactor(
	setting *cluster.Settings,
	internalEx sqlutil.InternalExecutor,
	db *kv.DB,
	rowsRemovedCounter *metric.Counter,
	knobs *sqlstats.TestingKnobs,
) *StatsCompactor {
	__antithesis_instrumentation__.Notify(624585)
	return &StatsCompactor{
		st:                 setting,
		db:                 db,
		ie:                 internalEx,
		rowsRemovedCounter: rowsRemovedCounter,
		knobs:              knobs,
	}
}

func (c *StatsCompactor) DeleteOldestEntries(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(624586)
	if err := c.removeStaleRowsPerShard(
		ctx,
		stmtStatsCleanupOps,
	); err != nil {
		__antithesis_instrumentation__.Notify(624588)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624589)
	}
	__antithesis_instrumentation__.Notify(624587)

	return c.removeStaleRowsPerShard(
		ctx,
		txnStatsCleanupOps,
	)
}

func (c *StatsCompactor) removeStaleRowsPerShard(
	ctx context.Context, ops *cleanupOperations,
) error {
	__antithesis_instrumentation__.Notify(624590)
	rowLimitPerShard := c.getRowLimitPerShard()
	for shardIdx, rowLimit := range rowLimitPerShard {
		__antithesis_instrumentation__.Notify(624592)
		var existingRowCount int64

		if err := c.getRowCountForShard(
			ctx,
			ops.getScanStmt(c.knobs),
			shardIdx,
			&existingRowCount,
		); err != nil {
			__antithesis_instrumentation__.Notify(624595)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624596)
		}
		__antithesis_instrumentation__.Notify(624593)

		if c.knobs != nil && func() bool {
			__antithesis_instrumentation__.Notify(624597)
			return c.knobs.OnCleanupStartForShard != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(624598)
			c.knobs.OnCleanupStartForShard(shardIdx, existingRowCount, rowLimit)
		} else {
			__antithesis_instrumentation__.Notify(624599)
		}
		__antithesis_instrumentation__.Notify(624594)

		if err := c.removeStaleRowsForShard(
			ctx,
			ops,
			int64(shardIdx),
			existingRowCount,
			rowLimit,
		); err != nil {
			__antithesis_instrumentation__.Notify(624600)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624601)
		}
	}
	__antithesis_instrumentation__.Notify(624591)

	return nil
}

func (c *StatsCompactor) getRowCountForShard(
	ctx context.Context, stmt string, shardIdx int, count *int64,
) error {
	__antithesis_instrumentation__.Notify(624602)
	row, err := c.ie.QueryRowEx(ctx,
		"scan-row-count",
		nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		stmt,
		shardIdx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(624605)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624606)
	}
	__antithesis_instrumentation__.Notify(624603)

	if row.Len() != 1 {
		__antithesis_instrumentation__.Notify(624607)
		return errors.AssertionFailedf("unexpected number of column returned")
	} else {
		__antithesis_instrumentation__.Notify(624608)
	}
	__antithesis_instrumentation__.Notify(624604)
	*count = int64(tree.MustBeDInt(row[0]))

	return nil
}

func (c *StatsCompactor) getRowLimitPerShard() []int64 {
	__antithesis_instrumentation__.Notify(624609)
	limitPerShard := make([]int64, systemschema.SQLStatsHashShardBucketCount)
	maxPersistedRows := SQLStatsMaxPersistedRows.Get(&c.st.SV)

	for shardIdx := int64(0); shardIdx < systemschema.SQLStatsHashShardBucketCount; shardIdx++ {
		__antithesis_instrumentation__.Notify(624611)
		limitPerShard[shardIdx] = maxPersistedRows / (systemschema.SQLStatsHashShardBucketCount - shardIdx)
		maxPersistedRows -= limitPerShard[shardIdx]
	}
	__antithesis_instrumentation__.Notify(624610)

	return limitPerShard
}

func (c *StatsCompactor) removeStaleRowsForShard(
	ctx context.Context,
	ops *cleanupOperations,
	shardIdx int64,
	existingRowCountPerShard, maxRowLimitPerShard int64,
) error {
	__antithesis_instrumentation__.Notify(624612)
	var err error
	var lastDeletedRow tree.Datums
	maxDeleteRowsPerTxn := CompactionJobRowsToDeletePerTxn.Get(&c.st.SV)

	if rowsToRemove := existingRowCountPerShard - maxRowLimitPerShard; rowsToRemove > 0 {
		__antithesis_instrumentation__.Notify(624614)
		for remainToBeRemoved := rowsToRemove; remainToBeRemoved > 0; {
			__antithesis_instrumentation__.Notify(624615)
			rowsToRemovePerTxn := remainToBeRemoved
			if remainToBeRemoved > maxDeleteRowsPerTxn {
				__antithesis_instrumentation__.Notify(624619)
				rowsToRemovePerTxn = maxDeleteRowsPerTxn
			} else {
				__antithesis_instrumentation__.Notify(624620)
			}
			__antithesis_instrumentation__.Notify(624616)

			stmt := ops.getDeleteStmt(lastDeletedRow)
			qargs := c.getQargs(shardIdx, rowsToRemovePerTxn, lastDeletedRow)
			var rowsRemoved int64

			lastDeletedRow, rowsRemoved, err = c.executeDeleteStmt(
				ctx,
				stmt,
				qargs,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(624621)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624622)
			}
			__antithesis_instrumentation__.Notify(624617)
			c.rowsRemovedCounter.Inc(rowsToRemovePerTxn)

			if rowsRemoved < rowsToRemovePerTxn {
				__antithesis_instrumentation__.Notify(624623)
				break
			} else {
				__antithesis_instrumentation__.Notify(624624)
			}
			__antithesis_instrumentation__.Notify(624618)

			remainToBeRemoved -= rowsToRemovePerTxn

		}
	} else {
		__antithesis_instrumentation__.Notify(624625)
	}
	__antithesis_instrumentation__.Notify(624613)

	return nil
}

func (c *StatsCompactor) executeDeleteStmt(
	ctx context.Context, delStmt string, qargs []interface{},
) (lastRow tree.Datums, rowsDeleted int64, err error) {
	__antithesis_instrumentation__.Notify(624626)
	it, err := c.ie.QueryIteratorEx(ctx,
		"delete-old-sql-stats",
		nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		delStmt,
		qargs...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(624630)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(624631)
	}
	__antithesis_instrumentation__.Notify(624627)
	defer func() {
		__antithesis_instrumentation__.Notify(624632)
		err = errors.CombineErrors(err, it.Close())
	}()
	__antithesis_instrumentation__.Notify(624628)

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(624633)
		lastRow = it.Cur()
		rowsDeleted++
	}
	__antithesis_instrumentation__.Notify(624629)

	return lastRow, rowsDeleted, err
}

func (c *StatsCompactor) getQargs(shardIdx, limit int64, lastDeletedRow tree.Datums) []interface{} {
	__antithesis_instrumentation__.Notify(624634)
	size := len(lastDeletedRow) + 2
	if cap(c.scratch.qargs) < size {
		__antithesis_instrumentation__.Notify(624637)
		c.scratch.qargs = make([]interface{}, 0, size)
	} else {
		__antithesis_instrumentation__.Notify(624638)
	}
	__antithesis_instrumentation__.Notify(624635)
	c.scratch.qargs = c.scratch.qargs[:0]

	c.scratch.qargs = append(c.scratch.qargs, tree.NewDInt(tree.DInt(shardIdx)))
	c.scratch.qargs = append(c.scratch.qargs, tree.NewDInt(tree.DInt(limit)))

	for _, value := range lastDeletedRow {
		__antithesis_instrumentation__.Notify(624639)
		c.scratch.qargs = append(c.scratch.qargs, value)
	}
	__antithesis_instrumentation__.Notify(624636)

	return c.scratch.qargs
}

type cleanupOperations struct {
	initialScanStmtTemplate string
	unconstrainedDeleteStmt string
	constrainedDeleteStmt   string
}

var (
	stmtStatsCleanupOps = &cleanupOperations{
		initialScanStmtTemplate: `
      SELECT count(*)
      FROM system.statement_statistics
      %s
      WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_plan_hash_transaction_fingerprint_id_shard_8 = $1`,
		unconstrainedDeleteStmt: `
      DELETE FROM system.statement_statistics
      WHERE (aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id) IN (
        SELECT aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id
        FROM system.statement_statistics
        WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_plan_hash_transaction_fingerprint_id_shard_8 = $1
        ORDER BY aggregated_ts ASC
        LIMIT $2
      ) RETURNING aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id`,
		constrainedDeleteStmt: `
    DELETE FROM system.statement_statistics
    WHERE (aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id) IN (
    SELECT aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id
    FROM system.statement_statistics
    WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_plan_hash_transaction_fingerprint_id_shard_8 = $1
    AND (
      (
        aggregated_ts,
        fingerprint_id,
        transaction_fingerprint_id,
        plan_hash,
        app_name,
        node_id
        ) >= ($3, $4, $5, $6, $7, $8)
      )
      ORDER BY aggregated_ts ASC
      LIMIT $2
    ) RETURNING aggregated_ts, fingerprint_id, transaction_fingerprint_id, plan_hash, app_name, node_id`,
	}
	txnStatsCleanupOps = &cleanupOperations{
		initialScanStmtTemplate: `
      SELECT count(*)
      FROM system.transaction_statistics
      %s
      WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_shard_8 = $1`,
		unconstrainedDeleteStmt: `
    DELETE FROM system.transaction_statistics
    WHERE (aggregated_ts, fingerprint_id, app_name, node_id) IN (
      SELECT aggregated_ts, fingerprint_id, app_name, node_id
      FROM system.transaction_statistics
      WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_shard_8 = $1
      ORDER BY aggregated_ts ASC
      LIMIT $2
    ) RETURNING aggregated_ts, fingerprint_id, app_name, node_id`,
		constrainedDeleteStmt: `
    DELETE FROM system.transaction_statistics
      WHERE (aggregated_ts, fingerprint_id, app_name, node_id) IN (
      SELECT aggregated_ts, fingerprint_id, app_name, node_id
      FROM system.transaction_statistics
      WHERE crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_shard_8 = $1
      AND (
        (
        aggregated_ts,
        fingerprint_id,
        app_name,
        node_id
        ) >= ($3, $4, $5, $6)
      )
      ORDER BY aggregated_ts ASC
      LIMIT $2
    ) RETURNING aggregated_ts, fingerprint_id, app_name, node_id`,
	}
)

func (c *cleanupOperations) getScanStmt(knobs *sqlstats.TestingKnobs) string {
	__antithesis_instrumentation__.Notify(624640)
	return fmt.Sprintf(c.initialScanStmtTemplate, knobs.GetAOSTClause())
}

func (c *cleanupOperations) getDeleteStmt(lastDeletedRow tree.Datums) string {
	__antithesis_instrumentation__.Notify(624641)
	if len(lastDeletedRow) == 0 {
		__antithesis_instrumentation__.Notify(624643)
		return c.unconstrainedDeleteStmt
	} else {
		__antithesis_instrumentation__.Notify(624644)
	}
	__antithesis_instrumentation__.Notify(624642)

	return c.constrainedDeleteStmt
}
