package bench

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
)

const schema = `
DROP TABLE IF EXISTS pgbench_accounts;
DROP TABLE IF EXISTS pgbench_branches;
DROP TABLE IF EXISTS pgbench_tellers;
DROP TABLE IF EXISTS pgbench_history;

CREATE TABLE pgbench_accounts (
    aid integer NOT NULL PRIMARY KEY,
    bid integer,
    abalance integer,
    filler character(84)
);

CREATE TABLE pgbench_branches (
    bid integer NOT NULL PRIMARY KEY,
    bbalance integer,
    filler character(88)
);

CREATE TABLE pgbench_tellers (
    tid integer NOT NULL PRIMARY KEY,
    bid integer,
    tbalance integer,
    filler character(84)
);

CREATE TABLE pgbench_history (
    tid integer,
    bid integer,
    aid integer,
    delta integer,
    mtime timestamp,
    filler character(22)
);
`

func CreateAndConnect(pgURL url.URL, name string) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(2022)
	{
		__antithesis_instrumentation__.Notify(2025)
		pgURL.Path = ""
		db, err := gosql.Open("postgres", pgURL.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(2028)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(2029)
		}
		__antithesis_instrumentation__.Notify(2026)
		defer db.Close()

		if _, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)); err != nil {
			__antithesis_instrumentation__.Notify(2030)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(2031)
		}
		__antithesis_instrumentation__.Notify(2027)

		if _, err := db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, name)); err != nil {
			__antithesis_instrumentation__.Notify(2032)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(2033)
		}
	}
	__antithesis_instrumentation__.Notify(2023)

	pgURL.Path = name

	db, err := gosql.Open("postgres", pgURL.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(2034)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(2035)
	}
	__antithesis_instrumentation__.Notify(2024)
	return db, nil
}

func SetupExec(pgURL url.URL, name string, accounts, transactions int) (*exec.Cmd, error) {
	__antithesis_instrumentation__.Notify(2036)
	db, err := CreateAndConnect(pgURL, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(2039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(2040)
	}
	__antithesis_instrumentation__.Notify(2037)
	defer db.Close()

	if err := SetupBenchDB(db, accounts, true); err != nil {
		__antithesis_instrumentation__.Notify(2041)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(2042)
	}
	__antithesis_instrumentation__.Notify(2038)

	return ExecPgbench(pgURL, name, transactions)
}

func SetupBenchDB(db sqlutils.DBHandle, accounts int, quiet bool) error {
	__antithesis_instrumentation__.Notify(2043)
	ctx := context.TODO()
	if _, err := db.ExecContext(ctx, schema); err != nil {
		__antithesis_instrumentation__.Notify(2045)
		return err
	} else {
		__antithesis_instrumentation__.Notify(2046)
	}
	__antithesis_instrumentation__.Notify(2044)
	return populateDB(ctx, db, accounts, quiet)
}

const tellers = 10

func populateDB(ctx context.Context, db sqlutils.DBHandle, accounts int, quiet bool) error {
	__antithesis_instrumentation__.Notify(2047)
	branches := `INSERT INTO pgbench_branches (bid, bbalance, filler) VALUES (1, 7354, NULL)`
	if r, err := db.ExecContext(ctx, branches); err != nil {
		__antithesis_instrumentation__.Notify(2052)
		return err
	} else {
		__antithesis_instrumentation__.Notify(2053)
		if x, err := r.RowsAffected(); err != nil {
			__antithesis_instrumentation__.Notify(2054)
			return err
		} else {
			__antithesis_instrumentation__.Notify(2055)
			if !quiet {
				__antithesis_instrumentation__.Notify(2056)
				fmt.Printf("Inserted %d branch records\n", x)
			} else {
				__antithesis_instrumentation__.Notify(2057)
			}
		}
	}
	__antithesis_instrumentation__.Notify(2048)

	tellers := `INSERT INTO pgbench_tellers VALUES (1, 1, 0, NULL),
	(2, 1, 955, NULL),
	(3, 1, -3338, NULL),
	(4, 1, -24, NULL),
	(5, 1, 2287, NULL),
	(6, 1, 4129, NULL),
	(7, 1, 0, NULL),
	(8, 1, 0, NULL),
	(9, 1, 0, NULL),
	(10, 1, 3345, NULL)
	`
	if r, err := db.ExecContext(ctx, tellers); err != nil {
		__antithesis_instrumentation__.Notify(2058)
		return err
	} else {
		__antithesis_instrumentation__.Notify(2059)
		if x, err := r.RowsAffected(); err != nil {
			__antithesis_instrumentation__.Notify(2060)
			return err
		} else {
			__antithesis_instrumentation__.Notify(2061)
			if !quiet {
				__antithesis_instrumentation__.Notify(2062)
				fmt.Printf("Inserted %d teller records\n", x)
			} else {
				__antithesis_instrumentation__.Notify(2063)
			}
		}
	}
	__antithesis_instrumentation__.Notify(2049)

	done := 0
	for {
		__antithesis_instrumentation__.Notify(2064)
		batch := 5000
		remaining := accounts - done
		if remaining < 1 {
			__antithesis_instrumentation__.Notify(2069)
			break
		} else {
			__antithesis_instrumentation__.Notify(2070)
		}
		__antithesis_instrumentation__.Notify(2065)
		if batch > remaining {
			__antithesis_instrumentation__.Notify(2071)
			batch = remaining
		} else {
			__antithesis_instrumentation__.Notify(2072)
		}
		__antithesis_instrumentation__.Notify(2066)
		var placeholders bytes.Buffer
		for i := 0; i < batch; i++ {
			__antithesis_instrumentation__.Notify(2073)
			if i > 0 {
				__antithesis_instrumentation__.Notify(2075)
				placeholders.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(2076)
			}
			__antithesis_instrumentation__.Notify(2074)
			fmt.Fprintf(&placeholders, "(%d, 1, 0, '                                                                                    ')", done+i)
		}
		__antithesis_instrumentation__.Notify(2067)
		stmt := fmt.Sprintf(`INSERT INTO pgbench_accounts VALUES %s`, placeholders.String())
		if r, err := db.ExecContext(ctx, stmt); err != nil {
			__antithesis_instrumentation__.Notify(2077)
			return err
		} else {
			__antithesis_instrumentation__.Notify(2078)
			if x, err := r.RowsAffected(); err != nil {
				__antithesis_instrumentation__.Notify(2079)
				return err
			} else {
				__antithesis_instrumentation__.Notify(2080)
				if !quiet {
					__antithesis_instrumentation__.Notify(2081)
					fmt.Printf("Inserted %d account records\n", x)
				} else {
					__antithesis_instrumentation__.Notify(2082)
				}
			}
		}
		__antithesis_instrumentation__.Notify(2068)
		done += batch
	}
	__antithesis_instrumentation__.Notify(2050)

	history := `
INSERT INTO pgbench_history VALUES
(5, 1, 36833, 407, CURRENT_TIMESTAMP, NULL),
(3, 1, 43082, -3338, CURRENT_TIMESTAMP, NULL),
(2, 1, 49129, 2872, CURRENT_TIMESTAMP, NULL),
(6, 1, 81223, 1064, CURRENT_TIMESTAMP, NULL),
(6, 1, 28316, 3065, CURRENT_TIMESTAMP, NULL),
(4, 1, 10146, -24, CURRENT_TIMESTAMP, NULL),
(10, 1, 12019, 2265, CURRENT_TIMESTAMP, NULL),
(2, 1, 46717, -1917, CURRENT_TIMESTAMP, NULL),
(5, 1, 68648, 1880, CURRENT_TIMESTAMP, NULL),
(10, 1, 46989, 1080, CURRENT_TIMESTAMP, NULL);`

	if r, err := db.ExecContext(ctx, history); err != nil {
		__antithesis_instrumentation__.Notify(2083)
		return err
	} else {
		__antithesis_instrumentation__.Notify(2084)
		if x, err := r.RowsAffected(); err != nil {
			__antithesis_instrumentation__.Notify(2085)
			return err
		} else {
			__antithesis_instrumentation__.Notify(2086)
			if !quiet {
				__antithesis_instrumentation__.Notify(2087)
				fmt.Printf("Inserted %d history records\n", x)
			} else {
				__antithesis_instrumentation__.Notify(2088)
			}
		}
	}
	__antithesis_instrumentation__.Notify(2051)

	return nil
}
