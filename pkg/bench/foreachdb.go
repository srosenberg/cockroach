package bench

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type BenchmarkFn func(b *testing.B, db *sqlutils.SQLRunner)

func benchmarkCockroach(b *testing.B, f BenchmarkFn) {
	__antithesis_instrumentation__.Notify(1789)
	s, db, _ := serverutils.StartServer(
		b, base.TestServerArgs{UseDatabase: "bench"})
	defer s.Stopper().Stop(context.TODO())

	if _, err := db.Exec(`CREATE DATABASE bench`); err != nil {
		__antithesis_instrumentation__.Notify(1791)
		b.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(1792)
	}
	__antithesis_instrumentation__.Notify(1790)

	f(b, sqlutils.MakeSQLRunner(db))
}

func benchmarkMultinodeCockroach(b *testing.B, f BenchmarkFn) {
	__antithesis_instrumentation__.Notify(1793)
	tc := testcluster.StartTestCluster(b, 3,
		base.TestClusterArgs{
			ReplicationMode: base.ReplicationAuto,
			ServerArgs: base.TestServerArgs{
				UseDatabase: "bench",
			},
		})
	if _, err := tc.Conns[0].Exec(`CREATE DATABASE bench`); err != nil {
		__antithesis_instrumentation__.Notify(1795)
		b.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(1796)
	}
	__antithesis_instrumentation__.Notify(1794)
	defer tc.Stopper().Stop(context.TODO())

	f(b, sqlutils.MakeRoundRobinSQLRunner(tc.Conns[0], tc.Conns[1], tc.Conns[2]))
}

func benchmarkPostgres(b *testing.B, f BenchmarkFn) {
	__antithesis_instrumentation__.Notify(1797)

	pgURL := url.URL{
		Scheme:   "postgres",
		Host:     "localhost:5432",
		RawQuery: "sslmode=require&dbname=postgres",
	}
	if conn, err := net.Dial("tcp", pgURL.Host); err != nil {
		__antithesis_instrumentation__.Notify(1800)
		skip.IgnoreLintf(b, "unable to connect to postgres server on %s: %s", pgURL.Host, err)
	} else {
		__antithesis_instrumentation__.Notify(1801)
		conn.Close()
	}
	__antithesis_instrumentation__.Notify(1798)

	db, err := gosql.Open("postgres", pgURL.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(1802)
		b.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(1803)
	}
	__antithesis_instrumentation__.Notify(1799)
	defer db.Close()

	r := sqlutils.MakeSQLRunner(db)
	r.Exec(b, `CREATE SCHEMA IF NOT EXISTS bench`)

	f(b, r)
}

func benchmarkMySQL(b *testing.B, f BenchmarkFn) {
	__antithesis_instrumentation__.Notify(1804)
	const addr = "localhost:3306"
	if conn, err := net.Dial("tcp", addr); err != nil {
		__antithesis_instrumentation__.Notify(1807)
		skip.IgnoreLintf(b, "unable to connect to mysql server on %s: %s", addr, err)
	} else {
		__antithesis_instrumentation__.Notify(1808)
		conn.Close()
	}
	__antithesis_instrumentation__.Notify(1805)

	db, err := gosql.Open("mysql", fmt.Sprintf("root@tcp(%s)/", addr))
	if err != nil {
		__antithesis_instrumentation__.Notify(1809)
		b.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(1810)
	}
	__antithesis_instrumentation__.Notify(1806)
	defer db.Close()

	r := sqlutils.MakeSQLRunner(db)
	r.Exec(b, `CREATE DATABASE IF NOT EXISTS bench`)

	f(b, r)
}

func ForEachDB(b *testing.B, fn BenchmarkFn) {
	__antithesis_instrumentation__.Notify(1811)
	for _, dbFn := range []func(*testing.B, BenchmarkFn){
		benchmarkCockroach,
		benchmarkMultinodeCockroach,
		benchmarkPostgres,
		benchmarkMySQL,
	} {
		__antithesis_instrumentation__.Notify(1812)
		dbName := runtime.FuncForPC(reflect.ValueOf(dbFn).Pointer()).Name()
		dbName = strings.TrimPrefix(dbName, "github.com/cockroachdb/cockroach/pkg/bench.benchmark")
		b.Run(dbName, func(b *testing.B) {
			__antithesis_instrumentation__.Notify(1813)
			dbFn(b, fn)
		})
	}
}
