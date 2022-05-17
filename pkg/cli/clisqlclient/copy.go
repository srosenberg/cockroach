package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
)

type copyFromer interface {
	CopyData(ctx context.Context, line string) (r driver.Result, err error)
	Exec(v []driver.Value) (r driver.Result, err error)
	Close() error
}

type CopyFromState struct {
	driver.Tx
	copyFromer
}

func BeginCopyFrom(ctx context.Context, conn Conn, query string) (*CopyFromState, error) {
	__antithesis_instrumentation__.Notify(28779)
	txn, err := conn.(*sqlConn).conn.(driver.ConnBeginTx).BeginTx(ctx, driver.TxOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(28782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28783)
	}
	__antithesis_instrumentation__.Notify(28780)
	stmt, err := txn.(driver.Conn).Prepare(query)
	if err != nil {
		__antithesis_instrumentation__.Notify(28784)
		return nil, errors.CombineErrors(err, txn.Rollback())
	} else {
		__antithesis_instrumentation__.Notify(28785)
	}
	__antithesis_instrumentation__.Notify(28781)
	return &CopyFromState{Tx: txn, copyFromer: stmt.(copyFromer)}, nil
}

type copyFromRows struct {
	r driver.Result
}

func (c copyFromRows) Close() error {
	__antithesis_instrumentation__.Notify(28786)
	return nil
}

func (c copyFromRows) Columns() []string {
	__antithesis_instrumentation__.Notify(28787)
	return nil
}

func (c copyFromRows) ColumnTypeScanType(index int) reflect.Type {
	__antithesis_instrumentation__.Notify(28788)
	return nil
}

func (c copyFromRows) ColumnTypeDatabaseTypeName(index int) string {
	__antithesis_instrumentation__.Notify(28789)
	return ""
}

func (c copyFromRows) ColumnTypeNames() []string {
	__antithesis_instrumentation__.Notify(28790)
	return nil
}

func (c copyFromRows) Result() driver.Result {
	__antithesis_instrumentation__.Notify(28791)
	return c.r
}

func (c copyFromRows) Tag() string {
	__antithesis_instrumentation__.Notify(28792)
	return "COPY"
}

func (c copyFromRows) Next(values []driver.Value) error {
	__antithesis_instrumentation__.Notify(28793)
	return io.EOF
}

func (c copyFromRows) NextResultSet() (bool, error) {
	__antithesis_instrumentation__.Notify(28794)
	return false, nil
}

func (c *CopyFromState) Cancel() error {
	__antithesis_instrumentation__.Notify(28795)
	return errors.CombineErrors(c.copyFromer.Close(), c.Tx.Rollback())
}

func (c *CopyFromState) Commit(ctx context.Context, cleanupFunc func(), lines string) QueryFn {
	__antithesis_instrumentation__.Notify(28796)
	return func(ctx context.Context, conn Conn) (Rows, bool, error) {
		__antithesis_instrumentation__.Notify(28797)
		defer cleanupFunc()
		rows, isMulti, err := func() (Rows, bool, error) {
			__antithesis_instrumentation__.Notify(28800)
			for _, l := range strings.Split(lines, "\n") {
				__antithesis_instrumentation__.Notify(28803)
				_, err := c.copyFromer.CopyData(ctx, l)
				if err != nil {
					__antithesis_instrumentation__.Notify(28804)
					return nil, false, err
				} else {
					__antithesis_instrumentation__.Notify(28805)
				}
			}
			__antithesis_instrumentation__.Notify(28801)
			r, err := c.copyFromer.Exec(nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(28806)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(28807)
			}
			__antithesis_instrumentation__.Notify(28802)
			return copyFromRows{r: r}, false, c.Tx.Commit()
		}()
		__antithesis_instrumentation__.Notify(28798)
		if err != nil {
			__antithesis_instrumentation__.Notify(28808)
			return rows, isMulti, errors.CombineErrors(err, errors.CombineErrors(c.copyFromer.Close(), c.Tx.Rollback()))
		} else {
			__antithesis_instrumentation__.Notify(28809)
		}
		__antithesis_instrumentation__.Notify(28799)
		return rows, isMulti, err
	}
}
