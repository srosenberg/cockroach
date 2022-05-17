package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func WaitFor3XReplication(ctx context.Context, t test.Test, db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(52195)
	return WaitForReplication(ctx, t, db, 3)
}

func WaitForReplication(
	ctx context.Context, t test.Test, db *gosql.DB, replicationFactor int,
) error {
	__antithesis_instrumentation__.Notify(52196)
	t.L().Printf("waiting for initial up-replication...")
	tStart := timeutil.Now()
	var oldN int
	for {
		__antithesis_instrumentation__.Notify(52197)
		var n int
		if err := db.QueryRowContext(
			ctx,
			fmt.Sprintf(
				"SELECT count(1) FROM crdb_internal.ranges WHERE array_length(replicas, 1) < %d",
				replicationFactor,
			),
		).Scan(&n); err != nil {
			__antithesis_instrumentation__.Notify(52201)
			return err
		} else {
			__antithesis_instrumentation__.Notify(52202)
		}
		__antithesis_instrumentation__.Notify(52198)
		if n == 0 {
			__antithesis_instrumentation__.Notify(52203)
			t.L().Printf("up-replication complete")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(52204)
		}
		__antithesis_instrumentation__.Notify(52199)
		if timeutil.Since(tStart) > 30*time.Second || func() bool {
			__antithesis_instrumentation__.Notify(52205)
			return oldN != n == true
		}() == true {
			__antithesis_instrumentation__.Notify(52206)
			t.L().Printf("still waiting for full replication (%d ranges left)", n)
		} else {
			__antithesis_instrumentation__.Notify(52207)
		}
		__antithesis_instrumentation__.Notify(52200)
		oldN = n
		time.Sleep(time.Second)
	}
}

func WaitForUpdatedReplicationReport(ctx context.Context, t test.Test, db *gosql.DB) {
	__antithesis_instrumentation__.Notify(52208)
	t.L().Printf("waiting for updated replication report...")

	if _, err := db.ExecContext(
		ctx, `SET CLUSTER setting kv.replication_reports.interval = '2s'`,
	); err != nil {
		__antithesis_instrumentation__.Notify(52211)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52212)
	}
	__antithesis_instrumentation__.Notify(52209)
	defer func() {
		__antithesis_instrumentation__.Notify(52213)
		if _, err := db.ExecContext(
			ctx, `RESET CLUSTER setting kv.replication_reports.interval`,
		); err != nil {
			__antithesis_instrumentation__.Notify(52214)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52215)
		}
	}()
	__antithesis_instrumentation__.Notify(52210)

	tStart := timeutil.Now()
	for r := retry.StartWithCtx(ctx, retry.Options{}); r.Next(); {
		__antithesis_instrumentation__.Notify(52216)
		var count int
		var gen gosql.NullTime
		if err := db.QueryRowContext(
			ctx, `SELECT count(*), min(generated) FROM system.reports_meta`,
		).Scan(&count, &gen); err != nil {
			__antithesis_instrumentation__.Notify(52218)
			if !errors.Is(err, gosql.ErrNoRows) {
				__antithesis_instrumentation__.Notify(52219)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52220)
			}

		} else {
			__antithesis_instrumentation__.Notify(52221)
			if count == 3 && func() bool {
				__antithesis_instrumentation__.Notify(52222)
				return tStart.Before(gen.Time) == true
			}() == true {
				__antithesis_instrumentation__.Notify(52223)

				return
			} else {
				__antithesis_instrumentation__.Notify(52224)
			}
		}
		__antithesis_instrumentation__.Notify(52217)
		if timeutil.Since(tStart) > 30*time.Second {
			__antithesis_instrumentation__.Notify(52225)
			t.L().Printf("still waiting for updated replication report")
		} else {
			__antithesis_instrumentation__.Notify(52226)
		}
	}
}

func SetAdmissionControl(ctx context.Context, t test.Test, c cluster.Cluster, enabled bool) {
	__antithesis_instrumentation__.Notify(52227)
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()
	val := "true"
	if !enabled {
		__antithesis_instrumentation__.Notify(52229)
		val = "false"
	} else {
		__antithesis_instrumentation__.Notify(52230)
	}
	__antithesis_instrumentation__.Notify(52228)
	for _, setting := range []string{"admission.kv.enabled", "admission.sql_kv_response.enabled",
		"admission.sql_sql_response.enabled"} {
		__antithesis_instrumentation__.Notify(52231)
		if _, err := db.ExecContext(
			ctx, "SET CLUSTER SETTING "+setting+" = '"+val+"'"); err != nil {
			__antithesis_instrumentation__.Notify(52232)
			t.Fatalf("failed to set admission control to %t: %v", enabled, err)
		} else {
			__antithesis_instrumentation__.Notify(52233)
		}
	}
}
