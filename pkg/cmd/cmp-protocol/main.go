// cmp-protocol connects to postgres and cockroach servers and compares
// the binary and text pgwire encodings of SQL statements. Statements can
// be specified in arguments (./cmp-protocol "select 1" "select 2") or will
// be generated randomly until a difference is found.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/cmp-protocol/pgconnect"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

var (
	pgAddr = flag.String("pg", "localhost:5432", "postgres address")
	pgUser = flag.String("pg-user", "postgres", "postgres user")
	crAddr = flag.String("cr", "localhost:26257", "cockroach address")
	crUser = flag.String("cr-user", "root", "cockroach user")
)

func main() {
	__antithesis_instrumentation__.Notify(37791)
	flag.Parse()

	stmtCh := make(chan string)
	if args := os.Args[1:]; len(args) > 0 {
		__antithesis_instrumentation__.Notify(37793)
		go func() {
			__antithesis_instrumentation__.Notify(37794)
			for _, arg := range os.Args[1:] {
				__antithesis_instrumentation__.Notify(37796)
				stmtCh <- arg
			}
			__antithesis_instrumentation__.Notify(37795)
			close(stmtCh)
		}()
	} else {
		__antithesis_instrumentation__.Notify(37797)
		go func() {
			__antithesis_instrumentation__.Notify(37798)
			rng, _ := randutil.NewPseudoRand()
			for {
				__antithesis_instrumentation__.Notify(37799)
				typ := randgen.RandType(rng)
				sem := typ.Family()
				switch sem {
				case types.DecimalFamily,
					types.CollatedStringFamily,
					types.OidFamily,
					types.FloatFamily,
					types.TimestampTZFamily,
					types.UnknownFamily,

					types.ArrayFamily,
					types.TupleFamily:
					__antithesis_instrumentation__.Notify(37802)
					continue
				default:
					__antithesis_instrumentation__.Notify(37803)
				}
				__antithesis_instrumentation__.Notify(37800)
				datum := randgen.RandDatum(rng, typ, false)
				if datum == tree.DNull {
					__antithesis_instrumentation__.Notify(37804)
					continue
				} else {
					__antithesis_instrumentation__.Notify(37805)
				}
				__antithesis_instrumentation__.Notify(37801)
				for _, format := range []string{
					"SELECT %s::%s;",
					"SELECT ARRAY[%s::%s];",
					"SELECT (%s::%s, NULL);",
				} {
					__antithesis_instrumentation__.Notify(37806)
					input := fmt.Sprintf(format, datum, pgTypeName(sem))
					stmtCh <- input
					fmt.Printf("\nTYP: %v, DATUM: %v\n", sem, datum)
				}
			}
		}()
	}
	__antithesis_instrumentation__.Notify(37792)

	for input := range stmtCh {
		__antithesis_instrumentation__.Notify(37807)
		fmt.Println("INPUT", input)
		if err := compare(os.Stdout, input, *pgAddr, *crAddr, *pgUser, *crUser); err != nil {
			__antithesis_instrumentation__.Notify(37808)
			fmt.Fprintln(os.Stderr, "ERROR:", input)
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(37809)
			fmt.Fprintln(os.Stderr, "OK", input)
		}
	}
}

func pgTypeName(sem types.Family) string {
	__antithesis_instrumentation__.Notify(37810)
	switch sem {
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(37811)
		return "TEXT"
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(37812)
		return "BYTEA"
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(37813)
		return "INT8"
	default:
		__antithesis_instrumentation__.Notify(37814)
		return sem.String()
	}
}

func compare(w io.Writer, input, pgAddr, crAddr, pgUser, crUser string) error {
	__antithesis_instrumentation__.Notify(37815)
	ctx := context.Background()
	for _, code := range []pgwirebase.FormatCode{
		pgwirebase.FormatText,
		pgwirebase.FormatBinary,
	} {
		__antithesis_instrumentation__.Notify(37817)

		if code == pgwirebase.FormatBinary && func() bool {
			__antithesis_instrumentation__.Notify(37819)
			return strings.HasPrefix(input, "SELECT (") == true
		}() == true {
			__antithesis_instrumentation__.Notify(37820)
			continue
		} else {
			__antithesis_instrumentation__.Notify(37821)
		}
		__antithesis_instrumentation__.Notify(37818)
		results := map[string][]byte{}
		for _, s := range []struct {
			user string
			addr string
		}{
			{user: pgUser, addr: pgAddr},
			{user: crUser, addr: crAddr},
		} {
			__antithesis_instrumentation__.Notify(37822)
			user := s.user
			addr := s.addr
			res, err := pgconnect.Connect(ctx, input, addr, user, code)
			if err != nil {
				__antithesis_instrumentation__.Notify(37825)
				return errors.Wrapf(err, "addr: %s, code: %s", addr, code)
			} else {
				__antithesis_instrumentation__.Notify(37826)
			}
			__antithesis_instrumentation__.Notify(37823)
			fmt.Printf("INPUT: %s, ADDR: %s, CODE: %s, res: %q, res: %v\n", input, addr, code, res, res)
			for k, v := range results {
				__antithesis_instrumentation__.Notify(37827)
				if !bytes.Equal(res, v) {
					__antithesis_instrumentation__.Notify(37828)
					return errors.Errorf("format: %s\naddr: %s\nstr: %q\nbytes: %[3]v\n!=\naddr: %s\nstr: %q\nbytes: %[5]v\n", code, k, v, addr, res)
				} else {
					__antithesis_instrumentation__.Notify(37829)
				}
			}
			__antithesis_instrumentation__.Notify(37824)
			results[addr] = res
		}
	}
	__antithesis_instrumentation__.Notify(37816)
	return nil
}
