package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

func CheckKeyCount(t *testing.T, kvDB *kv.DB, span roachpb.Span, numKeys int) {
	__antithesis_instrumentation__.Notify(628338)
	t.Helper()
	if err := CheckKeyCountE(t, kvDB, span, numKeys); err != nil {
		__antithesis_instrumentation__.Notify(628339)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(628340)
	}
}

func CheckKeyCountE(t *testing.T, kvDB *kv.DB, span roachpb.Span, numKeys int) error {
	__antithesis_instrumentation__.Notify(628341)
	t.Helper()
	if kvs, err := kvDB.Scan(context.TODO(), span.Key, span.EndKey, 0); err != nil {
		__antithesis_instrumentation__.Notify(628343)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628344)
		if l := numKeys; len(kvs) != l {
			__antithesis_instrumentation__.Notify(628345)
			return errors.Newf("expected %d key value pairs, but got %d", l, len(kvs))
		} else {
			__antithesis_instrumentation__.Notify(628346)
		}
	}
	__antithesis_instrumentation__.Notify(628342)
	return nil
}

func CreateKVTable(sqlDB *gosql.DB, name string, numRows int) error {
	__antithesis_instrumentation__.Notify(628347)

	schemaStmts := []string{
		`CREATE DATABASE IF NOT EXISTS t;`,
		fmt.Sprintf(`CREATE TABLE t.%s (k INT PRIMARY KEY, v INT, FAMILY (k), FAMILY (v));`, name),
		fmt.Sprintf(`CREATE INDEX foo on t.%s (v);`, name),
	}

	for _, stmt := range schemaStmts {
		__antithesis_instrumentation__.Notify(628351)
		if _, err := sqlDB.Exec(stmt); err != nil {
			__antithesis_instrumentation__.Notify(628352)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628353)
		}
	}
	__antithesis_instrumentation__.Notify(628348)

	var insert bytes.Buffer
	if _, err := insert.WriteString(
		fmt.Sprintf(`INSERT INTO t.%s VALUES (%d, %d)`, name, 0, numRows-1)); err != nil {
		__antithesis_instrumentation__.Notify(628354)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628355)
	}
	__antithesis_instrumentation__.Notify(628349)
	for i := 1; i < numRows; i++ {
		__antithesis_instrumentation__.Notify(628356)
		if _, err := insert.WriteString(fmt.Sprintf(` ,(%d, %d)`, i, numRows-i)); err != nil {
			__antithesis_instrumentation__.Notify(628357)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628358)
		}
	}
	__antithesis_instrumentation__.Notify(628350)
	_, err := sqlDB.Exec(insert.String())
	return err
}
