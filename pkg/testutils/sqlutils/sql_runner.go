package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type SQLRunner struct {
	DB                   DBHandle
	SucceedsSoonDuration time.Duration
}

type DBHandle interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (gosql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*gosql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *gosql.Row
}

var _ DBHandle = &gosql.DB{}
var _ DBHandle = &gosql.Conn{}
var _ DBHandle = &gosql.Tx{}

func MakeSQLRunner(db DBHandle) *SQLRunner {
	__antithesis_instrumentation__.Notify(646252)
	return &SQLRunner{DB: db}
}

func MakeRoundRobinSQLRunner(dbs ...DBHandle) *SQLRunner {
	__antithesis_instrumentation__.Notify(646253)
	return MakeSQLRunner(MakeRoundRobinDBHandle(dbs...))
}

func (sr *SQLRunner) Exec(t testutils.TB, query string, args ...interface{}) gosql.Result {
	__antithesis_instrumentation__.Notify(646254)
	t.Helper()
	r, err := sr.DB.ExecContext(context.Background(), query, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646256)
		t.Fatalf("error executing '%s': %s", query, err)
	} else {
		__antithesis_instrumentation__.Notify(646257)
	}
	__antithesis_instrumentation__.Notify(646255)
	return r
}

func (sr *SQLRunner) succeedsWithin(t testutils.TB, f func() error) {
	__antithesis_instrumentation__.Notify(646258)
	t.Helper()
	d := sr.SucceedsSoonDuration
	if d == 0 {
		__antithesis_instrumentation__.Notify(646260)
		d = testutils.DefaultSucceedsSoonDuration
	} else {
		__antithesis_instrumentation__.Notify(646261)
	}
	__antithesis_instrumentation__.Notify(646259)
	require.NoError(t, testutils.SucceedsWithinError(f, d))
}

func (sr *SQLRunner) ExecSucceedsSoon(t testutils.TB, query string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646262)
	t.Helper()
	sr.succeedsWithin(t, func() error {
		__antithesis_instrumentation__.Notify(646263)
		_, err := sr.DB.ExecContext(context.Background(), query, args...)
		return err
	})
}

func (sr *SQLRunner) ExecRowsAffected(
	t testutils.TB, expRowsAffected int, query string, args ...interface{},
) {
	__antithesis_instrumentation__.Notify(646264)
	t.Helper()
	r := sr.Exec(t, query, args...)
	numRows, err := r.RowsAffected()
	if err != nil {
		__antithesis_instrumentation__.Notify(646266)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646267)
	}
	__antithesis_instrumentation__.Notify(646265)
	if numRows != int64(expRowsAffected) {
		__antithesis_instrumentation__.Notify(646268)
		t.Fatalf("expected %d affected rows, got %d on '%s'", expRowsAffected, numRows, query)
	} else {
		__antithesis_instrumentation__.Notify(646269)
	}
}

func (sr *SQLRunner) ExpectErr(t testutils.TB, errRE string, query string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646270)
	t.Helper()
	_, err := sr.DB.ExecContext(context.Background(), query, args...)
	if !testutils.IsError(err, errRE) {
		__antithesis_instrumentation__.Notify(646271)
		t.Fatalf("expected error '%s', got: %s", errRE, pgerror.FullError(err))
	} else {
		__antithesis_instrumentation__.Notify(646272)
	}
}

func (sr *SQLRunner) ExpectErrSucceedsSoon(
	t testutils.TB, errRE string, query string, args ...interface{},
) {
	__antithesis_instrumentation__.Notify(646273)
	t.Helper()
	sr.succeedsWithin(t, func() error {
		__antithesis_instrumentation__.Notify(646274)
		_, err := sr.DB.ExecContext(context.Background(), query, args...)
		if !testutils.IsError(err, errRE) {
			__antithesis_instrumentation__.Notify(646276)
			return errors.Newf("expected error '%s', got: %s", errRE, pgerror.FullError(err))
		} else {
			__antithesis_instrumentation__.Notify(646277)
		}
		__antithesis_instrumentation__.Notify(646275)
		return nil
	})
}

func (sr *SQLRunner) Query(t testutils.TB, query string, args ...interface{}) *gosql.Rows {
	__antithesis_instrumentation__.Notify(646278)
	t.Helper()
	r, err := sr.DB.QueryContext(context.Background(), query, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646280)
		t.Fatalf("error executing '%s': %s", query, err)
	} else {
		__antithesis_instrumentation__.Notify(646281)
	}
	__antithesis_instrumentation__.Notify(646279)
	return r
}

type Row struct {
	testutils.TB
	row *gosql.Row
}

func (r *Row) Scan(dest ...interface{}) {
	__antithesis_instrumentation__.Notify(646282)
	r.Helper()
	if err := r.row.Scan(dest...); err != nil {
		__antithesis_instrumentation__.Notify(646283)
		r.Fatalf("error scanning '%v': %+v", r.row, err)
	} else {
		__antithesis_instrumentation__.Notify(646284)
	}
}

func (sr *SQLRunner) QueryRow(t testutils.TB, query string, args ...interface{}) *Row {
	__antithesis_instrumentation__.Notify(646285)
	t.Helper()
	return &Row{t, sr.DB.QueryRowContext(context.Background(), query, args...)}
}

func (sr *SQLRunner) QueryStr(t testutils.TB, query string, args ...interface{}) [][]string {
	__antithesis_instrumentation__.Notify(646286)
	t.Helper()
	rows := sr.Query(t, query, args...)
	r, err := RowsToStrMatrix(rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(646288)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646289)
	}
	__antithesis_instrumentation__.Notify(646287)
	return r
}

func RowsToStrMatrix(rows *gosql.Rows) ([][]string, error) {
	__antithesis_instrumentation__.Notify(646290)
	cols, err := rows.Columns()
	if err != nil {
		__antithesis_instrumentation__.Notify(646295)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646296)
	}
	__antithesis_instrumentation__.Notify(646291)
	vals := make([]interface{}, len(cols))
	for i := range vals {
		__antithesis_instrumentation__.Notify(646297)
		vals[i] = new(interface{})
	}
	__antithesis_instrumentation__.Notify(646292)
	res := [][]string{}
	for rows.Next() {
		__antithesis_instrumentation__.Notify(646298)
		if err := rows.Scan(vals...); err != nil {
			__antithesis_instrumentation__.Notify(646301)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(646302)
		}
		__antithesis_instrumentation__.Notify(646299)
		row := make([]string, len(vals))
		for j, v := range vals {
			__antithesis_instrumentation__.Notify(646303)
			if val := *v.(*interface{}); val != nil {
				__antithesis_instrumentation__.Notify(646304)
				switch t := val.(type) {
				case []byte:
					__antithesis_instrumentation__.Notify(646305)
					row[j] = string(t)
				default:
					__antithesis_instrumentation__.Notify(646306)
					row[j] = fmt.Sprint(val)
				}
			} else {
				__antithesis_instrumentation__.Notify(646307)
				row[j] = "NULL"
			}
		}
		__antithesis_instrumentation__.Notify(646300)
		res = append(res, row)
	}
	__antithesis_instrumentation__.Notify(646293)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(646308)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646309)
	}
	__antithesis_instrumentation__.Notify(646294)
	return res, nil
}

func MatrixToStr(rows [][]string) string {
	__antithesis_instrumentation__.Notify(646310)
	res := strings.Builder{}
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(646312)
		res.WriteString(strings.Join(row, ", "))
		res.WriteRune('\n')
	}
	__antithesis_instrumentation__.Notify(646311)
	return res.String()
}

func (sr *SQLRunner) CheckQueryResults(t testutils.TB, query string, expected [][]string) {
	__antithesis_instrumentation__.Notify(646313)
	t.Helper()
	res := sr.QueryStr(t, query)
	if !reflect.DeepEqual(res, expected) {
		__antithesis_instrumentation__.Notify(646314)
		t.Errorf("query '%s': expected:\n%v\ngot:\n%v\n",
			query, MatrixToStr(expected), MatrixToStr(res),
		)
	} else {
		__antithesis_instrumentation__.Notify(646315)
	}
}

func (sr *SQLRunner) CheckQueryResultsRetry(t testutils.TB, query string, expected [][]string) {
	__antithesis_instrumentation__.Notify(646316)
	t.Helper()
	sr.succeedsWithin(t, func() error {
		__antithesis_instrumentation__.Notify(646317)
		res := sr.QueryStr(t, query)
		if !reflect.DeepEqual(res, expected) {
			__antithesis_instrumentation__.Notify(646319)
			return errors.Errorf("query '%s': expected:\n%v\ngot:\n%v\n",
				query, MatrixToStr(expected), MatrixToStr(res),
			)
		} else {
			__antithesis_instrumentation__.Notify(646320)
		}
		__antithesis_instrumentation__.Notify(646318)
		return nil
	})
}

type RoundRobinDBHandle struct {
	handles []DBHandle
	current int
}

var _ DBHandle = &RoundRobinDBHandle{}

func MakeRoundRobinDBHandle(handles ...DBHandle) *RoundRobinDBHandle {
	__antithesis_instrumentation__.Notify(646321)
	return &RoundRobinDBHandle{handles: handles}
}

func (rr *RoundRobinDBHandle) next() DBHandle {
	__antithesis_instrumentation__.Notify(646322)
	h := rr.handles[rr.current]
	rr.current = (rr.current + 1) % len(rr.handles)
	return h
}

func (rr *RoundRobinDBHandle) ExecContext(
	ctx context.Context, query string, args ...interface{},
) (gosql.Result, error) {
	__antithesis_instrumentation__.Notify(646323)
	return rr.next().ExecContext(ctx, query, args...)
}

func (rr *RoundRobinDBHandle) QueryContext(
	ctx context.Context, query string, args ...interface{},
) (*gosql.Rows, error) {
	__antithesis_instrumentation__.Notify(646324)
	return rr.next().QueryContext(ctx, query, args...)
}

func (rr *RoundRobinDBHandle) QueryRowContext(
	ctx context.Context, query string, args ...interface{},
) *gosql.Row {
	__antithesis_instrumentation__.Notify(646325)
	return rr.next().QueryRowContext(ctx, query, args...)
}
