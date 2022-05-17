package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"

	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
)

type QueryFn func(ctx context.Context, conn Conn) (rows Rows, isMultiStatementQuery bool, err error)

func MakeQuery(query string, parameters ...interface{}) QueryFn {
	__antithesis_instrumentation__.Notify(28818)
	return func(ctx context.Context, conn Conn) (Rows, bool, error) {
		__antithesis_instrumentation__.Notify(28819)
		isMultiStatementQuery, _ := scanner.HasMultipleStatements(query)
		rows, err := conn.Query(ctx, query, parameters...)
		return rows, isMultiStatementQuery, err
	}
}

func convertArgs(parameters []interface{}) ([]driver.NamedValue, error) {
	__antithesis_instrumentation__.Notify(28820)
	dVals := make([]driver.NamedValue, len(parameters))
	for i := range parameters {
		__antithesis_instrumentation__.Notify(28822)

		var err error
		dVals[i].Ordinal = i + 1
		dVals[i].Value, err = driver.DefaultParameterConverter.ConvertValue(parameters[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(28823)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(28824)
		}
	}
	__antithesis_instrumentation__.Notify(28821)
	return dVals, nil
}
