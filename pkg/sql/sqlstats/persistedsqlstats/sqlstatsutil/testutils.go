package sqlstatsutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/stretchr/testify/require"
)

func GetRandomizedCollectedStatementStatisticsForTest(
	t *testing.T,
) (result roachpb.CollectedStatementStatistics) {
	__antithesis_instrumentation__.Notify(625220)
	data := genRandomData()
	fillObject(t, reflect.ValueOf(&result), &data)

	return result
}

type randomData struct {
	Bool        bool
	String      string
	Int64       int64
	Float       float64
	IntArray    []int64
	StringArray []string
	Time        time.Time
}

var alphabet = []rune("abcdefghijklmkopqrstuvwxyz")

func genRandomData() randomData {
	__antithesis_instrumentation__.Notify(625221)
	r := randomData{}
	r.Bool = rand.Float64() > 0.5

	b := strings.Builder{}
	for i := 0; i < 20; i++ {
		__antithesis_instrumentation__.Notify(625225)
		b.WriteRune(alphabet[rand.Intn(26)])
	}
	__antithesis_instrumentation__.Notify(625222)
	r.String = b.String()

	arrLen := 5
	r.StringArray = make([]string, arrLen)
	for i := 0; i < arrLen; i++ {
		__antithesis_instrumentation__.Notify(625226)

		b := strings.Builder{}
		for j := 0; j < 10; j++ {
			__antithesis_instrumentation__.Notify(625228)
			b.WriteRune(alphabet[rand.Intn(26)])
		}
		__antithesis_instrumentation__.Notify(625227)
		r.StringArray[i] = b.String()
	}
	__antithesis_instrumentation__.Notify(625223)

	r.Int64 = rand.Int63()
	r.Float = rand.Float64()

	r.IntArray = make([]int64, arrLen)
	for i := 0; i < arrLen; i++ {
		__antithesis_instrumentation__.Notify(625229)
		r.IntArray[i] = rand.Int63()
	}
	__antithesis_instrumentation__.Notify(625224)

	r.Time = timeutil.Now()
	return r
}

func fillTemplate(t *testing.T, tmplStr string, data randomData) string {
	__antithesis_instrumentation__.Notify(625230)
	joinInts := func(arr []int64) string {
		__antithesis_instrumentation__.Notify(625234)
		strArr := make([]string, len(arr))
		for i, val := range arr {
			__antithesis_instrumentation__.Notify(625236)
			strArr[i] = strconv.FormatInt(val, 10)
		}
		__antithesis_instrumentation__.Notify(625235)
		return strings.Join(strArr, ",")
	}
	__antithesis_instrumentation__.Notify(625231)
	joinStrings := func(arr []string) string {
		__antithesis_instrumentation__.Notify(625237)
		strArr := make([]string, len(arr))
		for i, val := range arr {
			__antithesis_instrumentation__.Notify(625239)
			strArr[i] = fmt.Sprintf("%q", val)
		}
		__antithesis_instrumentation__.Notify(625238)
		return strings.Join(strArr, ",")
	}
	__antithesis_instrumentation__.Notify(625232)
	stringifyTime := func(tm time.Time) string {
		__antithesis_instrumentation__.Notify(625240)
		s, err := tm.MarshalText()
		require.NoError(t, err)
		return string(s)
	}
	__antithesis_instrumentation__.Notify(625233)
	tmpl, err := template.
		New("").
		Funcs(template.FuncMap{
			"joinInts":      joinInts,
			"joinStrings":   joinStrings,
			"stringifyTime": stringifyTime,
		}).
		Parse(tmplStr)
	require.NoError(t, err)

	b := strings.Builder{}
	err = tmpl.Execute(&b, data)
	require.NoError(t, err)

	return b.String()
}

var fieldBlacklist = map[string]struct{}{
	"App":                     {},
	"SensitiveInfo":           {},
	"LegacyLastErr":           {},
	"LegacyLastErrRedacted":   {},
	"StatementFingerprintIDs": {},
	"AggregatedTs":            {},
	"AggregationInterval":     {},
}

func fillObject(t *testing.T, val reflect.Value, data *randomData) {
	__antithesis_instrumentation__.Notify(625241)

	if val.Kind() != reflect.Ptr {
		__antithesis_instrumentation__.Notify(625243)
		t.Fatal("not a pointer type")
	} else {
		__antithesis_instrumentation__.Notify(625244)
	}
	__antithesis_instrumentation__.Notify(625242)

	val = reflect.Indirect(val)

	switch val.Kind() {
	case reflect.Uint64:
		__antithesis_instrumentation__.Notify(625245)
		val.SetUint(uint64(0))
	case reflect.Int64:
		__antithesis_instrumentation__.Notify(625246)
		val.SetInt(data.Int64)
	case reflect.String:
		__antithesis_instrumentation__.Notify(625247)
		val.SetString(data.String)
	case reflect.Float64:
		__antithesis_instrumentation__.Notify(625248)
		val.SetFloat(data.Float)
	case reflect.Bool:
		__antithesis_instrumentation__.Notify(625249)
		val.SetBool(data.Bool)
	case reflect.Slice:
		__antithesis_instrumentation__.Notify(625250)
		switch val.Type().String() {
		case "[]string":
			__antithesis_instrumentation__.Notify(625253)
			for _, randString := range data.StringArray {
				__antithesis_instrumentation__.Notify(625256)
				val.Set(reflect.Append(val, reflect.ValueOf(randString)))
			}
		case "[]int64":
			__antithesis_instrumentation__.Notify(625254)
			for _, randInt := range data.IntArray {
				__antithesis_instrumentation__.Notify(625257)
				val.Set(reflect.Append(val, reflect.ValueOf(randInt)))
			}
		default:
			__antithesis_instrumentation__.Notify(625255)
		}
	case reflect.Struct:
		__antithesis_instrumentation__.Notify(625251)
		switch val.Type().Name() {

		case "Time":
			__antithesis_instrumentation__.Notify(625258)
			val.Set(reflect.ValueOf(data.Time))
			return
		default:
			__antithesis_instrumentation__.Notify(625259)
			numFields := val.NumField()
			for i := 0; i < numFields; i++ {
				__antithesis_instrumentation__.Notify(625260)
				fieldName := val.Type().Field(i).Name
				fieldAddr := val.Field(i).Addr()
				if _, ok := fieldBlacklist[fieldName]; ok {
					__antithesis_instrumentation__.Notify(625262)
					continue
				} else {
					__antithesis_instrumentation__.Notify(625263)
				}
				__antithesis_instrumentation__.Notify(625261)

				fillObject(t, fieldAddr, data)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(625252)
		t.Fatalf("unsupported type: %s", val.Kind().String())
	}
}
