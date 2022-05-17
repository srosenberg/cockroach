package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
)

func loadTPCHDataset(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	sf int,
	m cluster.Monitor,
	roachNodes option.NodeListOption,
) error {
	__antithesis_instrumentation__.Notify(51459)
	db := c.Conn(ctx, t.L(), roachNodes[0])
	defer db.Close()

	if _, err := db.ExecContext(ctx, `USE tpch`); err == nil {
		__antithesis_instrumentation__.Notify(51462)
		t.L().Printf("found existing tpch dataset, verifying scale factor\n")

		var supplierCardinality int
		if err := db.QueryRowContext(
			ctx, `SELECT count(*) FROM tpch.supplier`,
		).Scan(&supplierCardinality); err != nil {
			__antithesis_instrumentation__.Notify(51465)
			if pqErr := (*pq.Error)(nil); !(errors.As(err, &pqErr) && func() bool {
				__antithesis_instrumentation__.Notify(51467)
				return pgcode.MakeCode(string(pqErr.Code)) == pgcode.UndefinedTable == true
			}() == true) {
				__antithesis_instrumentation__.Notify(51468)
				return err
			} else {
				__antithesis_instrumentation__.Notify(51469)
			}
			__antithesis_instrumentation__.Notify(51466)

			supplierCardinality = 0
		} else {
			__antithesis_instrumentation__.Notify(51470)
		}
		__antithesis_instrumentation__.Notify(51463)

		expectedSupplierCardinality := 10000 * sf
		if supplierCardinality >= expectedSupplierCardinality {
			__antithesis_instrumentation__.Notify(51471)
			t.L().Printf("dataset is at least of scale factor %d, continuing", sf)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(51472)
		}
		__antithesis_instrumentation__.Notify(51464)

		m.ExpectDeaths(int32(c.Spec().NodeCount))
		c.Wipe(ctx, roachNodes)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
		m.ResetDeaths()
	} else {
		__antithesis_instrumentation__.Notify(51473)
		if pqErr := (*pq.Error)(nil); !(errors.As(err, &pqErr) && func() bool {
			__antithesis_instrumentation__.Notify(51474)
			return pgcode.MakeCode(string(pqErr.Code)) == pgcode.InvalidCatalogName == true
		}() == true) {
			__antithesis_instrumentation__.Notify(51475)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51476)
		}
	}
	__antithesis_instrumentation__.Notify(51460)

	t.L().Printf("restoring tpch scale factor %d\n", sf)
	tpchURL := fmt.Sprintf("gs://cockroach-fixtures/workload/tpch/scalefactor=%d/backup?AUTH=implicit", sf)
	if _, err := db.ExecContext(ctx, `CREATE DATABASE IF NOT EXISTS tpch;`); err != nil {
		__antithesis_instrumentation__.Notify(51477)
		return err
	} else {
		__antithesis_instrumentation__.Notify(51478)
	}
	__antithesis_instrumentation__.Notify(51461)
	query := fmt.Sprintf(`RESTORE tpch.* FROM '%s' WITH into_db = 'tpch';`, tpchURL)
	_, err := db.ExecContext(ctx, query)
	return err
}

func scatterTables(t test.Test, conn *gosql.DB, tableNames []string) {
	__antithesis_instrumentation__.Notify(51479)
	t.Status("scattering the data")
	for _, table := range tableNames {
		__antithesis_instrumentation__.Notify(51480)
		scatter := fmt.Sprintf("ALTER TABLE %s SCATTER;", table)
		if _, err := conn.Exec(scatter); err != nil {
			__antithesis_instrumentation__.Notify(51481)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51482)
		}
	}
}

func disableAutoStats(t test.Test, conn *gosql.DB) {
	__antithesis_instrumentation__.Notify(51483)
	t.Status("disabling automatic collection of stats")
	if _, err := conn.Exec(
		`SET CLUSTER SETTING sql.stats.automatic_collection.enabled=false;`,
	); err != nil {
		__antithesis_instrumentation__.Notify(51484)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51485)
	}
}

func createStatsFromTables(t test.Test, conn *gosql.DB, tableNames []string) {
	__antithesis_instrumentation__.Notify(51486)
	t.Status("collecting stats")
	for _, tableName := range tableNames {
		__antithesis_instrumentation__.Notify(51487)
		t.Status(fmt.Sprintf("creating statistics from table %q", tableName))
		if _, err := conn.Exec(
			fmt.Sprintf(`CREATE STATISTICS %s FROM %s;`, tableName, tableName),
		); err != nil {
			__antithesis_instrumentation__.Notify(51488)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51489)
		}
	}
}
