package sqltestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/tests"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/stretchr/testify/require"
)

type ShowCreateTableTestCase struct {
	CreateStatement string

	Expect string

	Database string
}

func ShowCreateTableTest(
	t *testing.T, extraQuerySetup string, testCases []ShowCreateTableTestCase,
) {
	__antithesis_instrumentation__.Notify(625851)
	params, _ := tests.CreateTestServerParams()
	params.Locality.Tiers = []roachpb.Tier{
		{Key: "region", Value: "us-west1"},
	}
	s, sqlDB, _ := serverutils.StartServer(t, params)
	defer s.Stopper().Stop(context.Background())

	if _, err := sqlDB.Exec(`
    SET CLUSTER SETTING sql.cross_db_fks.enabled = TRUE;
		CREATE DATABASE d;
		USE d;
		-- Create a table we can point FKs to.
		CREATE TABLE items (
			a int8,
			b int8,
			c int8 unique,
			primary key (a, b)
		);
		-- Create a database we can cross reference.
		CREATE DATABASE o;
		CREATE TABLE o.foo(x int primary key);
	`); err != nil {
		__antithesis_instrumentation__.Notify(625854)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(625855)
	}
	__antithesis_instrumentation__.Notify(625852)
	if extraQuerySetup != "" {
		__antithesis_instrumentation__.Notify(625856)
		if _, err := sqlDB.Exec(extraQuerySetup); err != nil {
			__antithesis_instrumentation__.Notify(625857)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(625858)
		}
	} else {
		__antithesis_instrumentation__.Notify(625859)
	}
	__antithesis_instrumentation__.Notify(625853)
	for i, test := range testCases {
		__antithesis_instrumentation__.Notify(625860)
		name := fmt.Sprintf("t%d", i)
		t.Run(name, func(t *testing.T) {
			__antithesis_instrumentation__.Notify(625861)
			if test.Expect == "" {
				__antithesis_instrumentation__.Notify(625872)
				test.Expect = test.CreateStatement
			} else {
				__antithesis_instrumentation__.Notify(625873)
			}
			__antithesis_instrumentation__.Notify(625862)
			db := test.Database
			if db == "" {
				__antithesis_instrumentation__.Notify(625874)
				db = "d"
			} else {
				__antithesis_instrumentation__.Notify(625875)
			}
			__antithesis_instrumentation__.Notify(625863)
			_, err := sqlDB.Exec("USE $1", db)
			require.NoError(t, err)
			stmt := fmt.Sprintf(test.CreateStatement, name)
			expect := fmt.Sprintf(test.Expect, name)
			if _, err := sqlDB.Exec(stmt); err != nil {
				__antithesis_instrumentation__.Notify(625876)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(625877)
			}
			__antithesis_instrumentation__.Notify(625864)
			row := sqlDB.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", name))
			var scanName, create string
			if err := row.Scan(&scanName, &create); err != nil {
				__antithesis_instrumentation__.Notify(625878)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(625879)
			}
			__antithesis_instrumentation__.Notify(625865)
			if scanName != name {
				__antithesis_instrumentation__.Notify(625880)
				t.Fatalf("expected table name %s, got %s", name, scanName)
			} else {
				__antithesis_instrumentation__.Notify(625881)
			}
			__antithesis_instrumentation__.Notify(625866)
			if create != expect {
				__antithesis_instrumentation__.Notify(625882)
				t.Fatalf("statement: %s\ngot: %s\nexpected: %s", stmt, create, expect)
			} else {
				__antithesis_instrumentation__.Notify(625883)
			}
			__antithesis_instrumentation__.Notify(625867)
			if _, err := sqlDB.Exec(fmt.Sprintf("DROP TABLE %s", name)); err != nil {
				__antithesis_instrumentation__.Notify(625884)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(625885)
			}
			__antithesis_instrumentation__.Notify(625868)

			name += "_roundtrip"
			expect = fmt.Sprintf(test.Expect, name)
			if _, err := sqlDB.Exec(expect); err != nil {
				__antithesis_instrumentation__.Notify(625886)
				t.Fatalf("reinsert failure: %s: %s", expect, err)
			} else {
				__antithesis_instrumentation__.Notify(625887)
			}
			__antithesis_instrumentation__.Notify(625869)
			row = sqlDB.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", name))
			if err := row.Scan(&scanName, &create); err != nil {
				__antithesis_instrumentation__.Notify(625888)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(625889)
			}
			__antithesis_instrumentation__.Notify(625870)
			if create != expect {
				__antithesis_instrumentation__.Notify(625890)
				t.Fatalf("round trip statement: %s\ngot: %s", expect, create)
			} else {
				__antithesis_instrumentation__.Notify(625891)
			}
			__antithesis_instrumentation__.Notify(625871)
			if _, err := sqlDB.Exec(fmt.Sprintf("DROP TABLE %s", name)); err != nil {
				__antithesis_instrumentation__.Notify(625892)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(625893)
			}
		})
	}
}
