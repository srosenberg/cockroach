package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"database/sql/driver"
	"io"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
)

func getAllRowStrings(
	rows clisqlclient.Rows, colTypes []string, showMoreChars bool,
) ([][]string, error) {
	__antithesis_instrumentation__.Notify(29279)
	var allRows [][]string

	for {
		__antithesis_instrumentation__.Notify(29281)
		rowStrings, err := getNextRowStrings(rows, colTypes, showMoreChars)
		if err != nil {
			__antithesis_instrumentation__.Notify(29284)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(29285)
		}
		__antithesis_instrumentation__.Notify(29282)
		if rowStrings == nil {
			__antithesis_instrumentation__.Notify(29286)
			break
		} else {
			__antithesis_instrumentation__.Notify(29287)
		}
		__antithesis_instrumentation__.Notify(29283)
		allRows = append(allRows, rowStrings)
	}
	__antithesis_instrumentation__.Notify(29280)

	return allRows, nil
}

func getNextRowStrings(
	rows clisqlclient.Rows, colTypes []string, showMoreChars bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(29288)
	cols := rows.Columns()
	var vals []driver.Value
	if len(cols) > 0 {
		__antithesis_instrumentation__.Notify(29293)
		vals = make([]driver.Value, len(cols))
	} else {
		__antithesis_instrumentation__.Notify(29294)
	}
	__antithesis_instrumentation__.Notify(29289)

	err := rows.Next(vals)
	if err == io.EOF {
		__antithesis_instrumentation__.Notify(29295)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(29296)
	}
	__antithesis_instrumentation__.Notify(29290)
	if err != nil {
		__antithesis_instrumentation__.Notify(29297)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(29298)
	}
	__antithesis_instrumentation__.Notify(29291)

	rowStrings := make([]string, len(cols))
	for i, v := range vals {
		__antithesis_instrumentation__.Notify(29299)
		rowStrings[i] = FormatVal(v, colTypes[i], showMoreChars, showMoreChars)
	}
	__antithesis_instrumentation__.Notify(29292)
	return rowStrings, nil
}
