package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/internal/sqlsmith"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
)

const statementTimeout = time.Minute

func registerTLP(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51275)
	r.Add(registry.TestSpec{
		Name:    "tlp",
		Owner:   registry.OwnerSQLQueries,
		Timeout: time.Hour * 1,
		Tags:    nil,
		Cluster: r.MakeClusterSpec(1),
		Run:     runTLP,
	})
}

func runTLP(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(51276)
	timeout := 10 * time.Minute

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, t.Spec().(*registry.TestSpec).Timeout-5*time.Minute)
	defer cancel()
	done := ctx.Done()
	shouldExit := func() bool {
		__antithesis_instrumentation__.Notify(51279)
		select {
		case <-done:
			__antithesis_instrumentation__.Notify(51280)
			return true
		default:
			__antithesis_instrumentation__.Notify(51281)
			return false
		}
	}
	__antithesis_instrumentation__.Notify(51277)

	c.Put(ctx, t.Cockroach(), "./cockroach")
	if err := c.PutLibraries(ctx, "./lib"); err != nil {
		__antithesis_instrumentation__.Notify(51282)
		t.Fatalf("could not initialize libraries: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(51283)
	}
	__antithesis_instrumentation__.Notify(51278)

	for i := 0; ; i++ {
		__antithesis_instrumentation__.Notify(51284)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
		if shouldExit() {
			__antithesis_instrumentation__.Notify(51287)
			return
		} else {
			__antithesis_instrumentation__.Notify(51288)
		}
		__antithesis_instrumentation__.Notify(51285)

		runOneTLP(ctx, i, timeout, t, c)

		if shouldExit() {
			__antithesis_instrumentation__.Notify(51289)
			return
		} else {
			__antithesis_instrumentation__.Notify(51290)
		}
		__antithesis_instrumentation__.Notify(51286)
		c.Stop(ctx, t.L(), option.DefaultStopOpts())
		c.Wipe(ctx)
	}
}

func runOneTLP(
	ctx context.Context, iter int, timeout time.Duration, t test.Test, c cluster.Cluster,
) {
	__antithesis_instrumentation__.Notify(51291)

	tlpLog, err := os.Create(filepath.Join(t.ArtifactsDir(), fmt.Sprintf("tlp%03d.log", iter)))
	if err != nil {
		__antithesis_instrumentation__.Notify(51297)
		t.Fatalf("could not create tlp.log: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(51298)
	}
	__antithesis_instrumentation__.Notify(51292)
	defer tlpLog.Close()
	logStmt := func(stmt string) {
		__antithesis_instrumentation__.Notify(51299)
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			__antithesis_instrumentation__.Notify(51302)
			return
		} else {
			__antithesis_instrumentation__.Notify(51303)
		}
		__antithesis_instrumentation__.Notify(51300)
		fmt.Fprint(tlpLog, stmt)
		if !strings.HasSuffix(stmt, ";") {
			__antithesis_instrumentation__.Notify(51304)
			fmt.Fprint(tlpLog, ";")
		} else {
			__antithesis_instrumentation__.Notify(51305)
		}
		__antithesis_instrumentation__.Notify(51301)
		fmt.Fprint(tlpLog, "\n\n")
	}
	__antithesis_instrumentation__.Notify(51293)

	conn := c.Conn(ctx, t.L(), 1)

	rnd, seed := randutil.NewTestRand()
	t.L().Printf("seed: %d", seed)

	setup := sqlsmith.Setups[sqlsmith.RandTableSetupName](rnd)

	t.Status("executing setup")
	t.L().Printf("setup:\n%s", strings.Join(setup, "\n"))
	for _, stmt := range setup {
		__antithesis_instrumentation__.Notify(51306)
		if _, err := conn.Exec(stmt); err != nil {
			__antithesis_instrumentation__.Notify(51307)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51308)
			logStmt(stmt)
		}
	}
	__antithesis_instrumentation__.Notify(51294)

	setStmtTimeout := fmt.Sprintf("SET statement_timeout='%s';", statementTimeout.String())
	t.Status("setting statement_timeout")
	t.L().Printf("statement timeout:\n%s", setStmtTimeout)
	if _, err := conn.Exec(setStmtTimeout); err != nil {
		__antithesis_instrumentation__.Notify(51309)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51310)
	}
	__antithesis_instrumentation__.Notify(51295)
	logStmt(setStmtTimeout)

	smither, err := sqlsmith.NewSmither(conn, rnd, sqlsmith.MutationsOnly())
	if err != nil {
		__antithesis_instrumentation__.Notify(51311)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51312)
	}
	__antithesis_instrumentation__.Notify(51296)
	defer smither.Close()

	t.Status("running TLP")
	until := time.After(timeout)
	done := ctx.Done()
	for i := 1; ; i++ {
		__antithesis_instrumentation__.Notify(51313)
		select {
		case <-until:
			__antithesis_instrumentation__.Notify(51317)
			return
		case <-done:
			__antithesis_instrumentation__.Notify(51318)
			return
		default:
			__antithesis_instrumentation__.Notify(51319)
		}
		__antithesis_instrumentation__.Notify(51314)

		if i%1000 == 0 {
			__antithesis_instrumentation__.Notify(51320)
			t.Status("running TLP: ", i, " statements completed")
		} else {
			__antithesis_instrumentation__.Notify(51321)
		}
		__antithesis_instrumentation__.Notify(51315)

		if i < 1000 || func() bool {
			__antithesis_instrumentation__.Notify(51322)
			return i%10 == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(51323)
			runMutationStatement(conn, smither, logStmt)
			continue
		} else {
			__antithesis_instrumentation__.Notify(51324)
		}
		__antithesis_instrumentation__.Notify(51316)

		if err := runTLPQuery(conn, smither, logStmt); err != nil {
			__antithesis_instrumentation__.Notify(51325)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51326)
		}
	}
}

func runMutationStatement(conn *gosql.DB, smither *sqlsmith.Smither, logStmt func(string)) {
	__antithesis_instrumentation__.Notify(51327)

	defer func() {
		__antithesis_instrumentation__.Notify(51329)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(51330)
			return
		} else {
			__antithesis_instrumentation__.Notify(51331)
		}
	}()
	__antithesis_instrumentation__.Notify(51328)

	stmt := smither.Generate()

	_ = runWithTimeout(func() error {
		__antithesis_instrumentation__.Notify(51332)

		if _, err := conn.Exec(stmt); err == nil {
			__antithesis_instrumentation__.Notify(51334)
			logStmt(stmt)
		} else {
			__antithesis_instrumentation__.Notify(51335)
		}
		__antithesis_instrumentation__.Notify(51333)
		return nil
	})
}

func runTLPQuery(conn *gosql.DB, smither *sqlsmith.Smither, logStmt func(string)) error {
	__antithesis_instrumentation__.Notify(51336)

	defer func() {
		__antithesis_instrumentation__.Notify(51338)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(51339)
			return
		} else {
			__antithesis_instrumentation__.Notify(51340)
		}
	}()
	__antithesis_instrumentation__.Notify(51337)

	unpartitioned, partitioned, args := smither.GenerateTLP()

	return runWithTimeout(func() error {
		__antithesis_instrumentation__.Notify(51341)
		rows1, err := conn.Query(unpartitioned)
		if err != nil {
			__antithesis_instrumentation__.Notify(51347)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(51348)
		}
		__antithesis_instrumentation__.Notify(51342)
		defer rows1.Close()
		unpartitionedRows, err := sqlutils.RowsToStrMatrix(rows1)
		if err != nil {
			__antithesis_instrumentation__.Notify(51349)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(51350)
		}
		__antithesis_instrumentation__.Notify(51343)
		rows2, err := conn.Query(partitioned, args...)
		if err != nil {
			__antithesis_instrumentation__.Notify(51351)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(51352)
		}
		__antithesis_instrumentation__.Notify(51344)
		defer rows2.Close()
		partitionedRows, err := sqlutils.RowsToStrMatrix(rows2)
		if err != nil {
			__antithesis_instrumentation__.Notify(51353)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(51354)
		}
		__antithesis_instrumentation__.Notify(51345)

		if diff := unsortedMatricesDiff(unpartitionedRows, partitionedRows); diff != "" {
			__antithesis_instrumentation__.Notify(51355)
			logStmt(unpartitioned)
			logStmt(partitioned)
			return errors.Newf(
				"expected unpartitioned and partitioned results to be equal\n%s\nsql: %s\n%s\nwith args: %s",
				diff, unpartitioned, partitioned, args)
		} else {
			__antithesis_instrumentation__.Notify(51356)
		}
		__antithesis_instrumentation__.Notify(51346)
		return nil
	})
}

func unsortedMatricesDiff(rowMatrix1, rowMatrix2 [][]string) string {
	__antithesis_instrumentation__.Notify(51357)
	var rows1 []string
	for _, row := range rowMatrix1 {
		__antithesis_instrumentation__.Notify(51360)
		rows1 = append(rows1, strings.Join(row[:], ","))
	}
	__antithesis_instrumentation__.Notify(51358)
	var rows2 []string
	for _, row := range rowMatrix2 {
		__antithesis_instrumentation__.Notify(51361)
		rows2 = append(rows2, strings.Join(row[:], ","))
	}
	__antithesis_instrumentation__.Notify(51359)
	sort.Strings(rows1)
	sort.Strings(rows2)
	return cmp.Diff(rows1, rows2)
}

func runWithTimeout(f func() error) error {
	__antithesis_instrumentation__.Notify(51362)
	done := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(51364)
		err := f()
		done <- err
	}()
	__antithesis_instrumentation__.Notify(51363)
	select {
	case <-time.After(statementTimeout + time.Second*5):
		__antithesis_instrumentation__.Notify(51365)

		return nil
	case err := <-done:
		__antithesis_instrumentation__.Notify(51366)
		return err
	}
}
