package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"database/sql/driver"
	"reflect"

	"github.com/cockroachdb/errors"
)

type sqlRows struct {
	rows sqlRowsI
	conn *sqlConn
}

var _ Rows = (*sqlRows)(nil)

type sqlRowsI interface {
	driver.RowsColumnTypeScanType
	driver.RowsColumnTypeDatabaseTypeName
	Result() driver.Result
	Tag() string

	HasNextResultSet() bool
	NextResultSet() error
}

func (r *sqlRows) Columns() []string {
	__antithesis_instrumentation__.Notify(28829)
	return r.rows.Columns()
}

func (r *sqlRows) Result() driver.Result {
	__antithesis_instrumentation__.Notify(28830)
	return r.rows.Result()
}

func (r *sqlRows) Tag() string {
	__antithesis_instrumentation__.Notify(28831)
	return r.rows.Tag()
}

func (r *sqlRows) Close() error {
	__antithesis_instrumentation__.Notify(28832)
	r.conn.flushNotices()
	err := r.rows.Close()
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28834)
		r.conn.reconnecting = true
		r.conn.silentClose()
	} else {
		__antithesis_instrumentation__.Notify(28835)
	}
	__antithesis_instrumentation__.Notify(28833)
	return err
}

func (r *sqlRows) Next(values []driver.Value) error {
	__antithesis_instrumentation__.Notify(28836)
	err := r.rows.Next(values)
	if errors.Is(err, driver.ErrBadConn) {
		__antithesis_instrumentation__.Notify(28839)
		r.conn.reconnecting = true
		r.conn.silentClose()
	} else {
		__antithesis_instrumentation__.Notify(28840)
	}
	__antithesis_instrumentation__.Notify(28837)
	for i, v := range values {
		__antithesis_instrumentation__.Notify(28841)
		if b, ok := v.([]byte); ok {
			__antithesis_instrumentation__.Notify(28842)
			values[i] = append([]byte{}, b...)
		} else {
			__antithesis_instrumentation__.Notify(28843)
		}
	}
	__antithesis_instrumentation__.Notify(28838)

	r.conn.delayNotices = true
	return err
}

func (r *sqlRows) NextResultSet() (bool, error) {
	__antithesis_instrumentation__.Notify(28844)
	if !r.rows.HasNextResultSet() {
		__antithesis_instrumentation__.Notify(28846)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(28847)
	}
	__antithesis_instrumentation__.Notify(28845)
	return true, r.rows.NextResultSet()
}

func (r *sqlRows) ColumnTypeScanType(index int) reflect.Type {
	__antithesis_instrumentation__.Notify(28848)
	return r.rows.ColumnTypeScanType(index)
}

func (r *sqlRows) ColumnTypeDatabaseTypeName(index int) string {
	__antithesis_instrumentation__.Notify(28849)
	return r.rows.ColumnTypeDatabaseTypeName(index)
}

func (r *sqlRows) ColumnTypeNames() []string {
	__antithesis_instrumentation__.Notify(28850)
	colTypes := make([]string, len(r.Columns()))
	for i := range colTypes {
		__antithesis_instrumentation__.Notify(28852)
		colTypes[i] = r.ColumnTypeDatabaseTypeName(i)
	}
	__antithesis_instrumentation__.Notify(28851)
	return colTypes
}
