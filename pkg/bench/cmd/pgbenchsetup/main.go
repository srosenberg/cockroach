package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/cockroachdb/cockroach/pkg/bench"
	_ "github.com/lib/pq"
)

var usage = func() {
	__antithesis_instrumentation__.Notify(1771)
	fmt.Fprintln(os.Stderr, "Creates the schema and initial data used by the `pgbench` tool")
	fmt.Fprintf(os.Stderr, "\nUsage: %s <db URL>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	__antithesis_instrumentation__.Notify(1772)
	accounts := flag.Int("accounts", 100000, "number of accounts to create")

	createDb := flag.Bool("createdb", false, "attempt to create named db, dropping first if exists (must be able to connect to default db to do so).")

	flag.Parse()
	flag.Usage = usage
	if flag.NArg() != 1 {
		__antithesis_instrumentation__.Notify(1776)
		flag.Usage()
		os.Exit(2)
	} else {
		__antithesis_instrumentation__.Notify(1777)
	}
	__antithesis_instrumentation__.Notify(1773)

	var db *gosql.DB
	var err error

	if *createDb {
		__antithesis_instrumentation__.Notify(1778)
		name := ""
		parsed, parseErr := url.Parse(flag.Arg(0))
		if parseErr != nil {
			__antithesis_instrumentation__.Notify(1780)
			panic(parseErr)
		} else {
			__antithesis_instrumentation__.Notify(1781)
			if len(parsed.Path) < 2 {
				__antithesis_instrumentation__.Notify(1782)
				panic("URL must include db name")
			} else {
				__antithesis_instrumentation__.Notify(1783)
				name = parsed.Path[1:]
			}
		}
		__antithesis_instrumentation__.Notify(1779)

		db, err = bench.CreateAndConnect(*parsed, name)
	} else {
		__antithesis_instrumentation__.Notify(1784)
		db, err = gosql.Open("postgres", flag.Arg(0))
	}
	__antithesis_instrumentation__.Notify(1774)
	if err != nil {
		__antithesis_instrumentation__.Notify(1785)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(1786)
	}
	__antithesis_instrumentation__.Notify(1775)

	defer db.Close()

	if err := bench.SetupBenchDB(db, *accounts, false); err != nil {
		__antithesis_instrumentation__.Notify(1787)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(1788)
	}
}
