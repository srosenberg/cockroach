package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

func registerCopy(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46830)

	runCopy := func(ctx context.Context, t test.Test, c cluster.Cluster, rows int, inTxn bool) {
		__antithesis_instrumentation__.Notify(46832)

		const payload = 100

		const rowOverheadEstimate = 160
		const rowEstimate = rowOverheadEstimate + payload

		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		m := c.NewMonitor(ctx, c.All())
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(46834)
			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			if err := disableLoadBasedSplitting(ctx, db); err != nil {
				__antithesis_instrumentation__.Notify(46844)
				return errors.Wrap(err, "disabling load-based splitting")
			} else {
				__antithesis_instrumentation__.Notify(46845)
			}
			__antithesis_instrumentation__.Notify(46835)

			t.Status("importing Bank fixture")
			c.Run(ctx, c.Node(1), fmt.Sprintf(
				"./workload fixtures load bank --rows=%d --payload-bytes=%d {pgurl:1}",
				rows, payload))
			if _, err := db.Exec("ALTER TABLE bank.bank RENAME TO bank.bank_orig"); err != nil {
				__antithesis_instrumentation__.Notify(46846)
				t.Fatalf("failed to rename table: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(46847)
			}
			__antithesis_instrumentation__.Notify(46836)

			t.Status("create copy of Bank schema")
			c.Run(ctx, c.Node(1), "./workload init bank --rows=0 --ranges=0 {pgurl:1}")

			rangeCount := func() int {
				__antithesis_instrumentation__.Notify(46848)
				var count int
				const q = "SELECT count(*) FROM [SHOW RANGES FROM TABLE bank.bank]"
				if err := db.QueryRow(q).Scan(&count); err != nil {
					__antithesis_instrumentation__.Notify(46850)
					t.Fatalf("failed to get range count: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(46851)
				}
				__antithesis_instrumentation__.Notify(46849)
				return count
			}
			__antithesis_instrumentation__.Notify(46837)
			if rc := rangeCount(); rc != 1 {
				__antithesis_instrumentation__.Notify(46852)
				return errors.Errorf("empty bank table split over multiple ranges")
			} else {
				__antithesis_instrumentation__.Notify(46853)
			}
			__antithesis_instrumentation__.Notify(46838)

			rowsPerInsert := (60 << 20) / rowEstimate
			t.Status("copying from bank_orig to bank")

			type querier interface {
				QueryRowContext(ctx context.Context, query string, args ...interface{}) *gosql.Row
			}
			runCopy := func(ctx context.Context, qu querier) error {
				__antithesis_instrumentation__.Notify(46854)
				ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
				defer cancel()

				for lastID := -1; lastID+1 < rows; {
					__antithesis_instrumentation__.Notify(46856)
					if lastID > 0 {
						__antithesis_instrumentation__.Notify(46858)
						t.Progress(float64(lastID+1) / float64(rows))
					} else {
						__antithesis_instrumentation__.Notify(46859)
					}
					__antithesis_instrumentation__.Notify(46857)
					q := fmt.Sprintf(`
						SELECT id FROM [
							INSERT INTO bank.bank
							SELECT * FROM bank.bank_orig
							WHERE id > %d
							ORDER BY id ASC
							LIMIT %d
							RETURNING ID
						]
						ORDER BY id DESC
						LIMIT 1`,
						lastID, rowsPerInsert)
					if err := qu.QueryRowContext(ctx, q).Scan(&lastID); err != nil {
						__antithesis_instrumentation__.Notify(46860)
						return err
					} else {
						__antithesis_instrumentation__.Notify(46861)
					}
				}
				__antithesis_instrumentation__.Notify(46855)
				return nil
			}
			__antithesis_instrumentation__.Notify(46839)

			var err error
			if inTxn {
				__antithesis_instrumentation__.Notify(46862)
				attempt := 0
				err = crdb.ExecuteTx(ctx, db, nil, func(tx *gosql.Tx) error {
					__antithesis_instrumentation__.Notify(46863)
					attempt++
					if attempt > 5 {
						__antithesis_instrumentation__.Notify(46865)
						return errors.Errorf("aborting after %v failed attempts", attempt-1)
					} else {
						__antithesis_instrumentation__.Notify(46866)
					}
					__antithesis_instrumentation__.Notify(46864)
					t.Status(fmt.Sprintf("copying (attempt %v)", attempt))
					return runCopy(ctx, tx)
				})
			} else {
				__antithesis_instrumentation__.Notify(46867)
				err = runCopy(ctx, db)
			}
			__antithesis_instrumentation__.Notify(46840)
			if err != nil {
				__antithesis_instrumentation__.Notify(46868)
				t.Fatalf("failed to copy rows: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(46869)
			}
			__antithesis_instrumentation__.Notify(46841)
			rangeMinBytes, rangeMaxBytes, err := getDefaultRangeSize(ctx, db)
			if err != nil {
				__antithesis_instrumentation__.Notify(46870)
				t.Fatalf("failed to get default range size: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(46871)
			}
			__antithesis_instrumentation__.Notify(46842)
			rc := rangeCount()
			t.L().Printf("range count after copy = %d\n", rc)
			lowExp := (rows * rowEstimate) / rangeMaxBytes
			highExp := int(math.Ceil(float64(rows*rowEstimate) / float64(rangeMinBytes)))
			if rc > highExp || func() bool {
				__antithesis_instrumentation__.Notify(46872)
				return rc < lowExp == true
			}() == true {
				__antithesis_instrumentation__.Notify(46873)
				return errors.Errorf("expected range count for table between %d and %d, found %d",
					lowExp, highExp, rc)
			} else {
				__antithesis_instrumentation__.Notify(46874)
			}
			__antithesis_instrumentation__.Notify(46843)
			return nil
		})
		__antithesis_instrumentation__.Notify(46833)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(46831)

	testcases := []struct {
		rows  int
		nodes int
		txn   bool
	}{
		{rows: int(1e6), nodes: 9, txn: false},
		{rows: int(1e5), nodes: 5, txn: true},
	}

	for _, tc := range testcases {
		__antithesis_instrumentation__.Notify(46875)
		tc := tc
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("copy/bank/rows=%d,nodes=%d,txn=%t", tc.rows, tc.nodes, tc.txn),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(tc.nodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46876)
				runCopy(ctx, t, c, tc.rows, tc.txn)
			},
		})
	}
}

func getDefaultRangeSize(
	ctx context.Context, db *gosql.DB,
) (rangeMinBytes, rangeMaxBytes int, err error) {
	__antithesis_instrumentation__.Notify(46877)
	err = db.QueryRow(`SELECT
    regexp_extract(regexp_extract(raw_config_sql, e'range_min_bytes = \\d+'), e'\\d+')::INT8
        AS range_min_bytes,
    regexp_extract(regexp_extract(raw_config_sql, e'range_max_bytes = \\d+'), e'\\d+')::INT8
        AS range_max_bytes
FROM
    [SHOW ZONE CONFIGURATION FOR RANGE default];`).Scan(&rangeMinBytes, &rangeMaxBytes)

	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(46879)
		return strings.Contains(err.Error(), `column "raw_config_sql" does not exist`) == true
	}() == true {
		__antithesis_instrumentation__.Notify(46880)
		rangeMinBytes, rangeMaxBytes, err = 32<<20, 64<<20, nil
	} else {
		__antithesis_instrumentation__.Notify(46881)
	}
	__antithesis_instrumentation__.Notify(46878)
	return rangeMinBytes, rangeMaxBytes, err
}
