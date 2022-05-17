// skiperrs connects to a postgres-compatible server with its URL specified as
// the first argument. It then splits stdin into SQL statements and executes
// them on the connection. Errors are printed but do not stop execution.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cmd/cr2pg/sqlstream"
	"github.com/lib/pq"
)

func main() {
	__antithesis_instrumentation__.Notify(52792)
	if len(os.Args) != 2 {
		__antithesis_instrumentation__.Notify(52795)
		fmt.Printf("usage: %s <url>\n", os.Args[0])
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(52796)
	}
	__antithesis_instrumentation__.Notify(52793)
	url := os.Args[1]

	connector, err := pq.NewConnector(url)
	if err != nil {
		__antithesis_instrumentation__.Notify(52797)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52798)
	}
	__antithesis_instrumentation__.Notify(52794)
	db := gosql.OpenDB(connector)
	defer db.Close()

	stream := sqlstream.NewStream(os.Stdin)
	for {
		__antithesis_instrumentation__.Notify(52799)
		stmt, err := stream.Next()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(52801)
			break
		} else {
			__antithesis_instrumentation__.Notify(52802)
			if err != nil {
				__antithesis_instrumentation__.Notify(52803)
				log.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52804)
			}
		}
		__antithesis_instrumentation__.Notify(52800)
		if _, err := db.Exec(stmt.String()); err != nil {
			__antithesis_instrumentation__.Notify(52805)
			fmt.Println(err)
		} else {
			__antithesis_instrumentation__.Notify(52806)
		}
	}
}
