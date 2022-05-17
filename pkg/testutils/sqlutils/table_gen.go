package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	gosql "database/sql"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

const rowsPerInsert = 100

const TestDB = "test"

type GenRowFn func(row int) []tree.Datum

func genValues(w io.Writer, firstRow, lastRow int, fn GenRowFn, shouldPrint bool) {
	__antithesis_instrumentation__.Notify(646326)
	for rowIdx := firstRow; rowIdx <= lastRow; rowIdx++ {
		__antithesis_instrumentation__.Notify(646327)
		if rowIdx > firstRow {
			__antithesis_instrumentation__.Notify(646331)
			fmt.Fprint(w, ",")
		} else {
			__antithesis_instrumentation__.Notify(646332)
		}
		__antithesis_instrumentation__.Notify(646328)
		row := fn(rowIdx)
		if shouldPrint {
			__antithesis_instrumentation__.Notify(646333)
			var strs []string
			for _, v := range row {
				__antithesis_instrumentation__.Notify(646335)
				strs = append(strs, v.String())
			}
			__antithesis_instrumentation__.Notify(646334)
			fmt.Printf("(%v),\n", strings.Join(strs, ","))
		} else {
			__antithesis_instrumentation__.Notify(646336)
		}
		__antithesis_instrumentation__.Notify(646329)
		fmt.Fprintf(w, "(%s", tree.Serialize(row[0]))
		for _, v := range row[1:] {
			__antithesis_instrumentation__.Notify(646337)
			fmt.Fprintf(w, ",%s", tree.Serialize(v))
		}
		__antithesis_instrumentation__.Notify(646330)
		fmt.Fprint(w, ")")
	}
}

func CreateTable(
	tb testing.TB, sqlDB *gosql.DB, tableName, schema string, numRows int, fn GenRowFn,
) {
	__antithesis_instrumentation__.Notify(646338)
	CreateTableDebug(tb, sqlDB, tableName, schema, numRows, fn, false)
}

func CreateTableDebug(
	tb testing.TB,
	sqlDB *gosql.DB,
	tableName, schema string,
	numRows int,
	fn GenRowFn,
	shouldPrint bool,
) {
	__antithesis_instrumentation__.Notify(646339)
	r := MakeSQLRunner(sqlDB)
	stmt := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s;`, TestDB)
	stmt += fmt.Sprintf(`CREATE TABLE %s.%s (%s);`, TestDB, tableName, schema)
	r.Exec(tb, stmt)
	if shouldPrint {
		__antithesis_instrumentation__.Notify(646341)
		fmt.Printf("Creating table: %s\n%s\n", tableName, schema)
	} else {
		__antithesis_instrumentation__.Notify(646342)
	}
	__antithesis_instrumentation__.Notify(646340)
	for i := 1; i <= numRows; {
		__antithesis_instrumentation__.Notify(646343)
		var buf bytes.Buffer
		fmt.Fprintf(&buf, `INSERT INTO %s.%s VALUES `, TestDB, tableName)
		batchEnd := i + rowsPerInsert
		if batchEnd > numRows {
			__antithesis_instrumentation__.Notify(646345)
			batchEnd = numRows
		} else {
			__antithesis_instrumentation__.Notify(646346)
		}
		__antithesis_instrumentation__.Notify(646344)
		genValues(&buf, i, batchEnd, fn, shouldPrint)

		r.Exec(tb, buf.String())
		i = batchEnd + 1
	}
}

type GenValueFn func(row int) tree.Datum

func RowIdxFn(row int) tree.Datum {
	__antithesis_instrumentation__.Notify(646347)
	return tree.NewDInt(tree.DInt(row))
}

func RowModuloFn(modulo int) GenValueFn {
	__antithesis_instrumentation__.Notify(646348)
	return func(row int) tree.Datum {
		__antithesis_instrumentation__.Notify(646349)
		return tree.NewDInt(tree.DInt(row % modulo))
	}
}

func IntToEnglish(val int) string {
	__antithesis_instrumentation__.Notify(646350)
	if val < 0 {
		__antithesis_instrumentation__.Notify(646354)
		panic(val)
	} else {
		__antithesis_instrumentation__.Notify(646355)
	}
	__antithesis_instrumentation__.Notify(646351)
	d := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

	var digits []string
	digits = append(digits, d[val%10])
	for val > 9 {
		__antithesis_instrumentation__.Notify(646356)
		val /= 10
		digits = append(digits, d[val%10])
	}
	__antithesis_instrumentation__.Notify(646352)
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		__antithesis_instrumentation__.Notify(646357)
		digits[i], digits[j] = digits[j], digits[i]
	}
	__antithesis_instrumentation__.Notify(646353)
	return strings.Join(digits, "-")
}

func RowEnglishFn(row int) tree.Datum {
	__antithesis_instrumentation__.Notify(646358)
	return tree.NewDString(IntToEnglish(row))
}

func ToRowFn(fn ...GenValueFn) GenRowFn {
	__antithesis_instrumentation__.Notify(646359)
	return func(row int) []tree.Datum {
		__antithesis_instrumentation__.Notify(646360)
		res := make([]tree.Datum, 0, len(fn))
		for _, f := range fn {
			__antithesis_instrumentation__.Notify(646362)
			res = append(res, f(row))
		}
		__antithesis_instrumentation__.Notify(646361)
		return res
	}
}
