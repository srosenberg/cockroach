// smithcmp is a tool to execute random queries on a database. A TOML
// file provides configuration for which databases to connect to. If there
// is more than one, only non-mutating statements are generated, and the
// output is compared, exiting if there is a difference. If there is only
// one database, mutating and non-mutating statements are generated. A
// flag in the TOML controls whether Postgres-compatible output is generated.
//
// Explicit SQL statements can be specified (skipping sqlsmith generation)
// using the top-level SQL array. Placeholders (`$1`, etc.) are
// supported. Random datums of the correct type will be filled in.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cockroachdb/cockroach/pkg/cmd/cmpconn"
	"github.com/cockroachdb/cockroach/pkg/internal/sqlsmith"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/lib/pq/oid"
)

func usage() {
	__antithesis_instrumentation__.Notify(52862)
	const use = `Usage of %s:
	%[1]s config.toml
`

	fmt.Printf(use, os.Args[0])
	os.Exit(1)
}

type options struct {
	Postgres        bool
	InitSQL         string
	Smither         string
	Seed            int64
	TimeoutMins     int
	StmtTimeoutSecs int
	SQL             []string

	Databases map[string]struct {
		Addr           string
		InitSQL        string
		AllowMutations bool
	}
}

var sqlMutators = []randgen.Mutator{randgen.ColumnFamilyMutator}

func enableMutations(shouldEnable bool, mutations []randgen.Mutator) []randgen.Mutator {
	__antithesis_instrumentation__.Notify(52863)
	if shouldEnable {
		__antithesis_instrumentation__.Notify(52865)
		return mutations
	} else {
		__antithesis_instrumentation__.Notify(52866)
	}
	__antithesis_instrumentation__.Notify(52864)
	return nil
}

func main() {
	__antithesis_instrumentation__.Notify(52867)
	ctx := context.Background()
	args := os.Args[1:]
	if len(args) != 1 {
		__antithesis_instrumentation__.Notify(52878)
		usage()
	} else {
		__antithesis_instrumentation__.Notify(52879)
	}
	__antithesis_instrumentation__.Notify(52868)

	tomlData, err := ioutil.ReadFile(args[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(52880)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52881)
	}
	__antithesis_instrumentation__.Notify(52869)

	var opts options
	if err := toml.Unmarshal(tomlData, &opts); err != nil {
		__antithesis_instrumentation__.Notify(52882)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52883)
	}
	__antithesis_instrumentation__.Notify(52870)
	timeout := time.Duration(opts.TimeoutMins) * time.Minute
	if timeout <= 0 {
		__antithesis_instrumentation__.Notify(52884)
		timeout = 15 * time.Minute
	} else {
		__antithesis_instrumentation__.Notify(52885)
	}
	__antithesis_instrumentation__.Notify(52871)
	stmtTimeout := time.Duration(opts.StmtTimeoutSecs) * time.Second
	if stmtTimeout <= 0 {
		__antithesis_instrumentation__.Notify(52886)
		stmtTimeout = time.Minute
	} else {
		__antithesis_instrumentation__.Notify(52887)
	}
	__antithesis_instrumentation__.Notify(52872)

	rng := rand.New(rand.NewSource(opts.Seed))
	conns := map[string]cmpconn.Conn{}
	for name, db := range opts.Databases {
		__antithesis_instrumentation__.Notify(52888)
		var err error
		mutators := enableMutations(opts.Databases[name].AllowMutations, sqlMutators)
		if opts.Postgres {
			__antithesis_instrumentation__.Notify(52890)
			mutators = append(mutators, randgen.PostgresMutator)
		} else {
			__antithesis_instrumentation__.Notify(52891)
		}
		__antithesis_instrumentation__.Notify(52889)
		conns[name], err = cmpconn.NewConnWithMutators(
			ctx, db.Addr, rng, mutators, db.InitSQL, opts.InitSQL)
		if err != nil {
			__antithesis_instrumentation__.Notify(52892)
			log.Fatalf("%s (%s): %+v", name, db.Addr, err)
		} else {
			__antithesis_instrumentation__.Notify(52893)
		}
	}
	__antithesis_instrumentation__.Notify(52873)
	compare := len(conns) > 1

	if opts.Seed < 0 {
		__antithesis_instrumentation__.Notify(52894)
		opts.Seed = timeutil.Now().UnixNano()
		fmt.Println("seed:", opts.Seed)
	} else {
		__antithesis_instrumentation__.Notify(52895)
	}
	__antithesis_instrumentation__.Notify(52874)
	smithOpts := []sqlsmith.SmitherOption{
		sqlsmith.AvoidConsts(),
	}
	if opts.Postgres {
		__antithesis_instrumentation__.Notify(52896)
		smithOpts = append(smithOpts, sqlsmith.PostgresMode())
	} else {
		__antithesis_instrumentation__.Notify(52897)
		if compare {
			__antithesis_instrumentation__.Notify(52898)
			smithOpts = append(smithOpts,
				sqlsmith.CompareMode(),
				sqlsmith.DisableCRDBFns(),
			)
		} else {
			__antithesis_instrumentation__.Notify(52899)
		}
	}
	__antithesis_instrumentation__.Notify(52875)
	if _, ok := conns[opts.Smither]; !ok {
		__antithesis_instrumentation__.Notify(52900)
		log.Fatalf("Smither option not present in databases: %s", opts.Smither)
	} else {
		__antithesis_instrumentation__.Notify(52901)
	}
	__antithesis_instrumentation__.Notify(52876)
	var smither *sqlsmith.Smither
	var stmts []statement
	if len(opts.SQL) == 0 {
		__antithesis_instrumentation__.Notify(52902)
		smither, err = sqlsmith.NewSmither(conns[opts.Smither].DB(), rng, smithOpts...)
		if err != nil {
			__antithesis_instrumentation__.Notify(52903)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52904)
		}
	} else {
		__antithesis_instrumentation__.Notify(52905)
		stmts = make([]statement, len(opts.SQL))
		for i, stmt := range opts.SQL {
			__antithesis_instrumentation__.Notify(52906)
			ps, err := conns[opts.Smither].PGX().Prepare(ctx, "", stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(52909)
				log.Fatalf("bad SQL statement on %s: %v\nSQL:\n%s", opts.Smither, stmt, err)
			} else {
				__antithesis_instrumentation__.Notify(52910)
			}
			__antithesis_instrumentation__.Notify(52907)
			var placeholders []*types.T
			for _, param := range ps.ParamOIDs {
				__antithesis_instrumentation__.Notify(52911)
				typ, ok := types.OidToType[oid.Oid(param)]
				if !ok {
					__antithesis_instrumentation__.Notify(52913)
					log.Fatalf("unknown oid: %v", param)
				} else {
					__antithesis_instrumentation__.Notify(52914)
				}
				__antithesis_instrumentation__.Notify(52912)
				placeholders = append(placeholders, typ)
			}
			__antithesis_instrumentation__.Notify(52908)
			stmts[i] = statement{
				stmt:         stmt,
				placeholders: placeholders,
			}
		}
	}
	__antithesis_instrumentation__.Notify(52877)

	var prep, exec string
	done := time.After(timeout)
	for i, ignoredErrCount := 0, 0; true; i++ {
		__antithesis_instrumentation__.Notify(52915)
		select {
		case <-done:
			__antithesis_instrumentation__.Notify(52919)
			fmt.Printf("executed query count: %d\nignored error count: %d\n", i, ignoredErrCount)
			return
		default:
			__antithesis_instrumentation__.Notify(52920)
		}
		__antithesis_instrumentation__.Notify(52916)
		fmt.Printf("stmt: %d\n", i)
		if smither != nil {
			__antithesis_instrumentation__.Notify(52921)
			exec = smither.Generate()
		} else {
			__antithesis_instrumentation__.Notify(52922)
			randStatement := stmts[rng.Intn(len(stmts))]
			name := fmt.Sprintf("s%d", i)
			prep = fmt.Sprintf("PREPARE %s AS\n%s;", name, randStatement.stmt)
			var sb strings.Builder
			fmt.Fprintf(&sb, "EXECUTE %s", name)
			for i, typ := range randStatement.placeholders {
				__antithesis_instrumentation__.Notify(52925)
				if i > 0 {
					__antithesis_instrumentation__.Notify(52927)
					sb.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(52928)
					sb.WriteString(" (")
				}
				__antithesis_instrumentation__.Notify(52926)
				d := randgen.RandDatum(rng, typ, true)
				fmt.Println(i, typ, d, tree.Serialize(d))
				sb.WriteString(tree.Serialize(d))
			}
			__antithesis_instrumentation__.Notify(52923)
			if len(randStatement.placeholders) > 0 {
				__antithesis_instrumentation__.Notify(52929)
				fmt.Fprintf(&sb, ")")
			} else {
				__antithesis_instrumentation__.Notify(52930)
			}
			__antithesis_instrumentation__.Notify(52924)
			fmt.Fprintf(&sb, ";")
			exec = sb.String()
			fmt.Println(exec)
		}
		__antithesis_instrumentation__.Notify(52917)
		if compare {
			__antithesis_instrumentation__.Notify(52931)
			if ignoredErr, err := cmpconn.CompareConns(
				ctx, stmtTimeout, conns, prep, exec, true,
			); err != nil {
				__antithesis_instrumentation__.Notify(52932)
				fmt.Printf("prep:\n%s;\nexec:\n%s;\nERR: %s\n\n", prep, exec, err)
				os.Exit(1)
			} else {
				__antithesis_instrumentation__.Notify(52933)
				if ignoredErr {
					__antithesis_instrumentation__.Notify(52934)
					ignoredErrCount++
				} else {
					__antithesis_instrumentation__.Notify(52935)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(52936)
			for _, conn := range conns {
				__antithesis_instrumentation__.Notify(52937)
				if err := conn.Exec(ctx, prep+exec); err != nil {
					__antithesis_instrumentation__.Notify(52938)
					fmt.Println(err)
				} else {
					__antithesis_instrumentation__.Notify(52939)
				}
			}
		}
		__antithesis_instrumentation__.Notify(52918)

		for name, conn := range conns {
			__antithesis_instrumentation__.Notify(52940)
			start := timeutil.Now()
			fmt.Printf("pinging %s...", name)
			if err := conn.Ping(ctx); err != nil {
				__antithesis_instrumentation__.Notify(52942)
				fmt.Printf("\n%s: ping failure: %v\nprevious SQL:\n%s;\n%s;\n", name, err, prep, exec)

				db := opts.Databases[name]
				newConn, err := cmpconn.NewConnWithMutators(
					ctx, db.Addr, rng, enableMutations(db.AllowMutations, sqlMutators),
					db.InitSQL, opts.InitSQL,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(52944)
					log.Fatalf("tried to reconnect: %v\n", err)
				} else {
					__antithesis_instrumentation__.Notify(52945)
				}
				__antithesis_instrumentation__.Notify(52943)
				conns[name] = newConn
			} else {
				__antithesis_instrumentation__.Notify(52946)
			}
			__antithesis_instrumentation__.Notify(52941)
			fmt.Printf(" %s\n", timeutil.Since(start))
		}
	}
}

type statement struct {
	stmt         string
	placeholders []*types.T
}
