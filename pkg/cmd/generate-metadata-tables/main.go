// This connects to Postgresql and retrieves all the information about
// the tables in the pg_catalog schema. The result is printed into a
// comma separated lines which are meant to be store at a CSV and used
// to test that cockroachdb pg_catalog is up to date.
//
// This accepts the following arguments:
//
// --user:    to change default pg username of `postgres`
// --addr:    to change default pg address of `localhost:5432`
// --catalog: can be pg_catalog or information_schema. Default is pg_catalog
// --rdbms:   can be postgres or mysql. Default is postgres
// --stdout:  for testing purposes, use this flag to send the output to the
//            console
//
// Output of this file should generate (If not using --stout):
// pkg/sql/testdata/<catalog>_tables_from_<rdbms>.json
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/generate-metadata-tables/rdbms"
	"github.com/cockroachdb/cockroach/pkg/sql"
)

var testdataDir = filepath.Join("pkg", "sql", "testdata")

var (
	flagAddress = flag.String("addr", "localhost:5432", "RDBMS address")
	flagUser    = flag.String("user", "postgres", "RDBMS user")
	flagSchema  = flag.String("catalog", "pg_catalog", "Catalog or namespace")
	flagRDBMS   = flag.String("rdbms", sql.Postgres, "Determines which RDBMS it will connect")
	flagStdout  = flag.Bool("stdout", false, "Instead of re-writing the test data it will print the out to console")
)

func main() {
	__antithesis_instrumentation__.Notify(40326)
	ctx := context.Background()
	flag.Parse()
	conn, err := connect()
	if err != nil {
		__antithesis_instrumentation__.Notify(40333)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40334)
	}
	__antithesis_instrumentation__.Notify(40327)
	defer func() { __antithesis_instrumentation__.Notify(40335); _ = conn.Close(ctx) }()
	__antithesis_instrumentation__.Notify(40328)
	dbVersion, err := conn.DatabaseVersion(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(40336)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40337)
	}
	__antithesis_instrumentation__.Notify(40329)
	pgCatalogFile := &sql.PGMetadataFile{
		Version:    dbVersion,
		PGMetadata: sql.PGMetadataTables{},
	}

	rows, err := conn.DescribeSchema(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(40338)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40339)
	}
	__antithesis_instrumentation__.Notify(40330)

	rows.ForEachRow(func(tableName, columnName, dataTypeName string, dataTypeOid uint32) {
		__antithesis_instrumentation__.Notify(40340)
		pgCatalogFile.PGMetadata.AddColumnMetadata(tableName, columnName, dataTypeName, dataTypeOid)
	})
	__antithesis_instrumentation__.Notify(40331)

	writer, closeFn, err := getWriter()
	if err != nil {
		__antithesis_instrumentation__.Notify(40341)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40342)
	}
	__antithesis_instrumentation__.Notify(40332)
	defer closeFn()
	pgCatalogFile.Save(writer)
}

func connect() (rdbms.DBMetadataConnection, error) {
	__antithesis_instrumentation__.Notify(40343)
	connect, ok := rdbms.ConnectFns[*flagRDBMS]
	if !ok {
		__antithesis_instrumentation__.Notify(40345)
		return nil, fmt.Errorf("connect to %s is not supported", *flagRDBMS)
	} else {
		__antithesis_instrumentation__.Notify(40346)
	}
	__antithesis_instrumentation__.Notify(40344)
	return connect(*flagAddress, *flagUser, *flagSchema)
}

func testdata() string {
	__antithesis_instrumentation__.Notify(40347)
	path, err := os.Getwd()
	if err != nil {
		__antithesis_instrumentation__.Notify(40350)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40351)
	}
	__antithesis_instrumentation__.Notify(40348)
	if strings.HasSuffix(filepath.Join("pkg", "cmd", "generate-metadata-tables"), path) {
		__antithesis_instrumentation__.Notify(40352)
		return filepath.Join("..", "..", "..", testdataDir)
	} else {
		__antithesis_instrumentation__.Notify(40353)
	}
	__antithesis_instrumentation__.Notify(40349)
	return testdataDir
}

func getWriter() (io.Writer, func(), error) {
	__antithesis_instrumentation__.Notify(40354)
	if *flagStdout {
		__antithesis_instrumentation__.Notify(40356)
		return os.Stdout, func() { __antithesis_instrumentation__.Notify(40357) }, nil
	} else {
		__antithesis_instrumentation__.Notify(40358)
	}
	__antithesis_instrumentation__.Notify(40355)

	filename := sql.TablesMetadataFilename(testdata(), *flagRDBMS, *flagSchema)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	return file, func() { __antithesis_instrumentation__.Notify(40359); file.Close() }, err
}
