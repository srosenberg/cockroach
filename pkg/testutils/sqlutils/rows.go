package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v4"
)

func RowsToDataDrivenOutput(rows *gosql.Rows) (string, error) {
	__antithesis_instrumentation__.Notify(646195)

	cols, err := rows.Columns()
	if err != nil {
		__antithesis_instrumentation__.Notify(646200)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(646201)
	}
	__antithesis_instrumentation__.Notify(646196)

	elemsI := make([]interface{}, len(cols))
	for i := range elemsI {
		__antithesis_instrumentation__.Notify(646202)
		elemsI[i] = new(interface{})
	}
	__antithesis_instrumentation__.Notify(646197)
	elems := make([]string, len(cols))

	var output strings.Builder
	for rows.Next() {
		__antithesis_instrumentation__.Notify(646203)
		if err := rows.Scan(elemsI...); err != nil {
			__antithesis_instrumentation__.Notify(646206)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(646207)
		}
		__antithesis_instrumentation__.Notify(646204)
		for i, elem := range elemsI {
			__antithesis_instrumentation__.Notify(646208)
			val := *(elem.(*interface{}))
			switch t := val.(type) {
			case []byte:
				__antithesis_instrumentation__.Notify(646209)

				if str := string(t); utf8.ValidString(str) {
					__antithesis_instrumentation__.Notify(646211)
					elems[i] = str
				} else {
					__antithesis_instrumentation__.Notify(646212)
				}
			default:
				__antithesis_instrumentation__.Notify(646210)
				elems[i] = fmt.Sprintf("%v", val)
			}
		}
		__antithesis_instrumentation__.Notify(646205)
		output.WriteString(strings.Join(elems, " "))
		output.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(646198)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(646213)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(646214)
	}
	__antithesis_instrumentation__.Notify(646199)
	return output.String(), nil
}

func PGXRowsToDataDrivenOutput(rows pgx.Rows) (string, error) {
	__antithesis_instrumentation__.Notify(646215)

	cols := rows.FieldDescriptions()

	elemsI := make([]interface{}, len(cols))
	for i := range elemsI {
		__antithesis_instrumentation__.Notify(646219)
		elemsI[i] = new(interface{})
	}
	__antithesis_instrumentation__.Notify(646216)
	elems := make([]string, len(cols))

	var output strings.Builder
	for rows.Next() {
		__antithesis_instrumentation__.Notify(646220)
		if err := rows.Scan(elemsI...); err != nil {
			__antithesis_instrumentation__.Notify(646223)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(646224)
		}
		__antithesis_instrumentation__.Notify(646221)
		for i, elem := range elemsI {
			__antithesis_instrumentation__.Notify(646225)
			val := *(elem.(*interface{}))
			switch t := val.(type) {
			case []byte:
				__antithesis_instrumentation__.Notify(646226)

				if str := string(t); utf8.ValidString(str) {
					__antithesis_instrumentation__.Notify(646228)
					elems[i] = str
				} else {
					__antithesis_instrumentation__.Notify(646229)
				}
			default:
				__antithesis_instrumentation__.Notify(646227)
				elems[i] = fmt.Sprintf("%v", val)
			}
		}
		__antithesis_instrumentation__.Notify(646222)
		output.WriteString(strings.Join(elems, " "))
		output.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(646217)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(646230)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(646231)
	}
	__antithesis_instrumentation__.Notify(646218)
	return output.String(), nil
}
