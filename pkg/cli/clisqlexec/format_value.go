package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	gosql "database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/lib/pq"
)

func isNotPrintableASCII(r rune) bool {
	__antithesis_instrumentation__.Notify(29206)
	return r < 0x20 || func() bool {
		__antithesis_instrumentation__.Notify(29207)
		return r > 0x7e == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(29208)
		return r == '"' == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(29209)
		return r == '\\' == true
	}() == true
}
func isNotGraphicUnicode(r rune) bool {
	__antithesis_instrumentation__.Notify(29210)
	return !unicode.IsGraphic(r)
}
func isNotGraphicUnicodeOrTabOrNewline(r rune) bool {
	__antithesis_instrumentation__.Notify(29211)
	return r != '\t' && func() bool {
		__antithesis_instrumentation__.Notify(29212)
		return r != '\n' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(29213)
		return !unicode.IsGraphic(r) == true
	}() == true
}

func FormatVal(
	val driver.Value, colType string, showPrintableUnicode bool, showNewLinesAndTabs bool,
) string {
	__antithesis_instrumentation__.Notify(29214)
	if b, ok := val.([]byte); ok {
		__antithesis_instrumentation__.Notify(29217)
		if strings.HasPrefix(colType, "_") && func() bool {
			__antithesis_instrumentation__.Notify(29219)
			return len(b) > 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(29220)
			return b[0] == '{' == true
		}() == true {
			__antithesis_instrumentation__.Notify(29221)
			return formatArray(b, colType[1:], showPrintableUnicode, showNewLinesAndTabs)
		} else {
			__antithesis_instrumentation__.Notify(29222)
		}
		__antithesis_instrumentation__.Notify(29218)

		if colType == "NAME" || func() bool {
			__antithesis_instrumentation__.Notify(29223)
			return colType == "RECORD" == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(29224)
			return colType == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(29225)
			val = string(b)
			colType = "VARCHAR"
		} else {
			__antithesis_instrumentation__.Notify(29226)
		}
	} else {
		__antithesis_instrumentation__.Notify(29227)
	}
	__antithesis_instrumentation__.Notify(29215)

	switch t := val.(type) {
	case nil:
		__antithesis_instrumentation__.Notify(29228)
		return "NULL"

	case float64:
		__antithesis_instrumentation__.Notify(29229)
		width := 64
		if colType == "FLOAT4" {
			__antithesis_instrumentation__.Notify(29238)
			width = 32
		} else {
			__antithesis_instrumentation__.Notify(29239)
		}
		__antithesis_instrumentation__.Notify(29230)
		if math.IsInf(t, 1) {
			__antithesis_instrumentation__.Notify(29240)
			return "Infinity"
		} else {
			__antithesis_instrumentation__.Notify(29241)
			if math.IsInf(t, -1) {
				__antithesis_instrumentation__.Notify(29242)
				return "-Infinity"
			} else {
				__antithesis_instrumentation__.Notify(29243)
			}
		}
		__antithesis_instrumentation__.Notify(29231)
		return strconv.FormatFloat(t, 'g', -1, width)

	case string:
		__antithesis_instrumentation__.Notify(29232)
		if showPrintableUnicode {
			__antithesis_instrumentation__.Notify(29244)
			pred := isNotGraphicUnicode
			if showNewLinesAndTabs {
				__antithesis_instrumentation__.Notify(29246)
				pred = isNotGraphicUnicodeOrTabOrNewline
			} else {
				__antithesis_instrumentation__.Notify(29247)
			}
			__antithesis_instrumentation__.Notify(29245)
			if utf8.ValidString(t) && func() bool {
				__antithesis_instrumentation__.Notify(29248)
				return strings.IndexFunc(t, pred) == -1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(29249)
				return t
			} else {
				__antithesis_instrumentation__.Notify(29250)
			}
		} else {
			__antithesis_instrumentation__.Notify(29251)
			if strings.IndexFunc(t, isNotPrintableASCII) == -1 {
				__antithesis_instrumentation__.Notify(29252)
				return t
			} else {
				__antithesis_instrumentation__.Notify(29253)
			}
		}
		__antithesis_instrumentation__.Notify(29233)
		s := fmt.Sprintf("%+q", t)

		return s[1 : len(s)-1]

	case []byte:
		__antithesis_instrumentation__.Notify(29234)

		var buf bytes.Buffer
		lexbase.EncodeSQLBytesInner(&buf, string(t))
		return buf.String()

	case time.Time:
		__antithesis_instrumentation__.Notify(29235)
		tfmt, ok := timeOutputFormats[colType]
		if !ok {
			__antithesis_instrumentation__.Notify(29254)

			tfmt = timeutil.FullTimeFormat
		} else {
			__antithesis_instrumentation__.Notify(29255)
		}
		__antithesis_instrumentation__.Notify(29236)
		if tfmt == timeutil.TimestampWithTZFormat || func() bool {
			__antithesis_instrumentation__.Notify(29256)
			return tfmt == timeutil.TimeWithTZFormat == true
		}() == true {
			__antithesis_instrumentation__.Notify(29257)
			if _, offsetSeconds := t.Zone(); offsetSeconds%60 != 0 {
				__antithesis_instrumentation__.Notify(29258)
				tfmt += ":00:00"
			} else {
				__antithesis_instrumentation__.Notify(29259)
				if offsetSeconds%3600 != 0 {
					__antithesis_instrumentation__.Notify(29260)
					tfmt += ":00"
				} else {
					__antithesis_instrumentation__.Notify(29261)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(29262)
		}
		__antithesis_instrumentation__.Notify(29237)
		return t.Format(tfmt)
	}
	__antithesis_instrumentation__.Notify(29216)

	return fmt.Sprint(val)
}

func formatArray(
	b []byte, colType string, showPrintableUnicode bool, showNewLinesAndTabs bool,
) string {
	__antithesis_instrumentation__.Notify(29263)

	var backingArray interface{}

	var parsingArray gosql.Scanner

	switch colType {
	case "BOOL":
		__antithesis_instrumentation__.Notify(29267)
		boolArray := []bool{}
		backingArray = &boolArray
		parsingArray = (*pq.BoolArray)(&boolArray)
	case "FLOAT4", "FLOAT8":
		__antithesis_instrumentation__.Notify(29268)
		floatArray := []float64{}
		backingArray = &floatArray
		parsingArray = (*pq.Float64Array)(&floatArray)
	case "INT2", "INT4", "INT8", "OID":
		__antithesis_instrumentation__.Notify(29269)
		intArray := []int64{}
		backingArray = &intArray
		parsingArray = (*pq.Int64Array)(&intArray)
	case "TEXT", "VARCHAR", "NAME", "CHAR", "BPCHAR", "RECORD":
		__antithesis_instrumentation__.Notify(29270)
		stringArray := []string{}
		backingArray = &stringArray
		parsingArray = (*pq.StringArray)(&stringArray)
	default:
		__antithesis_instrumentation__.Notify(29271)
		genArray := [][]byte{}
		backingArray = &genArray
		parsingArray = &pq.GenericArray{A: &genArray}
	}
	__antithesis_instrumentation__.Notify(29264)

	if err := parsingArray.Scan(b); err != nil {
		__antithesis_instrumentation__.Notify(29272)

		return FormatVal(b, "BYTEA", showPrintableUnicode, showNewLinesAndTabs)
	} else {
		__antithesis_instrumentation__.Notify(29273)
	}
	__antithesis_instrumentation__.Notify(29265)

	var buf strings.Builder
	buf.WriteByte('{')
	comma := ""
	v := reflect.ValueOf(backingArray).Elem()
	for i := 0; i < v.Len(); i++ {
		__antithesis_instrumentation__.Notify(29274)
		buf.WriteString(comma)

		arrayVal := driver.Value(v.Index(i).Interface())

		vs := FormatVal(arrayVal, colType, showPrintableUnicode, showNewLinesAndTabs)

		if strings.IndexByte(vs, ',') >= 0 || func() bool {
			__antithesis_instrumentation__.Notify(29276)
			return reArrayStringEscape.MatchString(vs) == true
		}() == true {
			__antithesis_instrumentation__.Notify(29277)
			vs = "\"" + reArrayStringEscape.ReplaceAllString(vs, "\\$1") + "\""
		} else {
			__antithesis_instrumentation__.Notify(29278)
		}
		__antithesis_instrumentation__.Notify(29275)

		buf.WriteString(vs)
		comma = ","
	}
	__antithesis_instrumentation__.Notify(29266)
	buf.WriteByte('}')
	return buf.String()
}

var reArrayStringEscape = regexp.MustCompile(`(["\\])`)

var timeOutputFormats = map[string]string{
	"TIMESTAMP":   timeutil.TimestampWithoutTZFormat,
	"TIMESTAMPTZ": timeutil.TimestampWithTZFormat,
	"TIME":        timeutil.TimeWithoutTZFormat,
	"TIMETZ":      timeutil.TimeWithTZFormat,
	"DATE":        timeutil.DateFormat,
}
