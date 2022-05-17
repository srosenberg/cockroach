package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	gosql "database/sql"
	gojson "encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type Validator interface {
	NoteRow(partition string, key, value string, updated hlc.Timestamp) error

	NoteResolved(partition string, resolved hlc.Timestamp) error

	Failures() []string
}

type StreamValidator interface {
	Validator

	GetValuesForKeyBelowTimestamp(key string, timestamp hlc.Timestamp) ([]roachpb.KeyValue, error)
}

type timestampValue struct {
	ts    hlc.Timestamp
	value string
}

type orderValidator struct {
	topic                 string
	partitionForKey       map[string]string
	keyTimestampAndValues map[string][]timestampValue
	resolved              map[string]hlc.Timestamp

	failures []string
}

var NoOpValidator = &noOpValidator{}

var _ Validator = &orderValidator{}
var _ Validator = &noOpValidator{}
var _ StreamValidator = &orderValidator{}

type noOpValidator struct{}

func (v *noOpValidator) NoteRow(string, string, string, hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15057)
	return nil
}

func (v *noOpValidator) NoteResolved(string, hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15058)
	return nil
}

func (v *noOpValidator) Failures() []string { __antithesis_instrumentation__.Notify(15059); return nil }

func NewOrderValidator(topic string) Validator {
	__antithesis_instrumentation__.Notify(15060)
	return &orderValidator{
		topic:                 topic,
		partitionForKey:       make(map[string]string),
		keyTimestampAndValues: make(map[string][]timestampValue),
		resolved:              make(map[string]hlc.Timestamp),
	}
}

func NewStreamOrderValidator() StreamValidator {
	__antithesis_instrumentation__.Notify(15061)
	return &orderValidator{
		topic:                 "unused",
		partitionForKey:       make(map[string]string),
		keyTimestampAndValues: make(map[string][]timestampValue),
		resolved:              make(map[string]hlc.Timestamp),
	}
}

func (v *orderValidator) GetValuesForKeyBelowTimestamp(
	key string, timestamp hlc.Timestamp,
) ([]roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(15062)
	timestampValueTuples := v.keyTimestampAndValues[key]
	timestampsIdx := sort.Search(len(timestampValueTuples), func(i int) bool {
		__antithesis_instrumentation__.Notify(15065)
		return timestamp.Less(timestampValueTuples[i].ts)
	})
	__antithesis_instrumentation__.Notify(15063)
	var kv []roachpb.KeyValue
	for _, tsValue := range timestampValueTuples[:timestampsIdx] {
		__antithesis_instrumentation__.Notify(15066)
		byteRep := []byte(key)
		kv = append(kv, roachpb.KeyValue{
			Key: byteRep,
			Value: roachpb.Value{
				RawBytes:  []byte(tsValue.value),
				Timestamp: tsValue.ts,
			},
		})
	}
	__antithesis_instrumentation__.Notify(15064)

	return kv, nil
}

func (v *orderValidator) NoteRow(partition string, key, value string, updated hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15067)
	if prev, ok := v.partitionForKey[key]; ok && func() bool {
		__antithesis_instrumentation__.Notify(15073)
		return prev != partition == true
	}() == true {
		__antithesis_instrumentation__.Notify(15074)
		v.failures = append(v.failures, fmt.Sprintf(
			`key [%s] received on two partitions: %s and %s`, key, prev, partition,
		))
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15075)
	}
	__antithesis_instrumentation__.Notify(15068)
	v.partitionForKey[key] = partition

	timestampValueTuples := v.keyTimestampAndValues[key]
	timestampsIdx := sort.Search(len(timestampValueTuples), func(i int) bool {
		__antithesis_instrumentation__.Notify(15076)
		return updated.LessEq(timestampValueTuples[i].ts)
	})
	__antithesis_instrumentation__.Notify(15069)
	seen := timestampsIdx < len(timestampValueTuples) && func() bool {
		__antithesis_instrumentation__.Notify(15077)
		return timestampValueTuples[timestampsIdx].ts == updated == true
	}() == true

	if !seen && func() bool {
		__antithesis_instrumentation__.Notify(15078)
		return len(timestampValueTuples) > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(15079)
		return updated.Less(timestampValueTuples[len(timestampValueTuples)-1].ts) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15080)
		v.failures = append(v.failures, fmt.Sprintf(
			`topic %s partition %s: saw new row timestamp %s after %s was seen`,
			v.topic, partition,
			updated.AsOfSystemTime(),
			timestampValueTuples[len(timestampValueTuples)-1].ts.AsOfSystemTime(),
		))
	} else {
		__antithesis_instrumentation__.Notify(15081)
	}
	__antithesis_instrumentation__.Notify(15070)
	if !seen && func() bool {
		__antithesis_instrumentation__.Notify(15082)
		return updated.Less(v.resolved[partition]) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15083)
		v.failures = append(v.failures, fmt.Sprintf(
			`topic %s partition %s: saw new row timestamp %s after %s was resolved`,
			v.topic, partition, updated.AsOfSystemTime(), v.resolved[partition].AsOfSystemTime(),
		))
	} else {
		__antithesis_instrumentation__.Notify(15084)
	}
	__antithesis_instrumentation__.Notify(15071)

	if !seen {
		__antithesis_instrumentation__.Notify(15085)
		v.keyTimestampAndValues[key] = append(
			append(timestampValueTuples[:timestampsIdx], timestampValue{
				ts:    updated,
				value: value,
			}),
			timestampValueTuples[timestampsIdx:]...)
	} else {
		__antithesis_instrumentation__.Notify(15086)
	}
	__antithesis_instrumentation__.Notify(15072)
	return nil
}

func (v *orderValidator) NoteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15087)
	prev := v.resolved[partition]
	if prev.Less(resolved) {
		__antithesis_instrumentation__.Notify(15089)
		v.resolved[partition] = resolved
	} else {
		__antithesis_instrumentation__.Notify(15090)
	}
	__antithesis_instrumentation__.Notify(15088)
	return nil
}

func (v *orderValidator) Failures() []string {
	__antithesis_instrumentation__.Notify(15091)
	return v.failures
}

type beforeAfterValidator struct {
	sqlDB          *gosql.DB
	table          string
	primaryKeyCols []string
	resolved       map[string]hlc.Timestamp

	failures []string
}

func NewBeforeAfterValidator(sqlDB *gosql.DB, table string) (Validator, error) {
	__antithesis_instrumentation__.Notify(15092)
	primaryKeyCols, err := fetchPrimaryKeyCols(sqlDB, table)
	if err != nil {
		__antithesis_instrumentation__.Notify(15094)
		return nil, errors.Wrap(err, "fetchPrimaryKeyCols failed")
	} else {
		__antithesis_instrumentation__.Notify(15095)
	}
	__antithesis_instrumentation__.Notify(15093)

	return &beforeAfterValidator{
		sqlDB:          sqlDB,
		table:          table,
		primaryKeyCols: primaryKeyCols,
		resolved:       make(map[string]hlc.Timestamp),
	}, nil
}

func (v *beforeAfterValidator) NoteRow(
	partition string, key, value string, updated hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(15096)
	var primaryKeyDatums []interface{}
	if err := gojson.Unmarshal([]byte(key), &primaryKeyDatums); err != nil {
		__antithesis_instrumentation__.Notify(15102)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15103)
	}
	__antithesis_instrumentation__.Notify(15097)
	if len(primaryKeyDatums) != len(v.primaryKeyCols) {
		__antithesis_instrumentation__.Notify(15104)
		return errors.Errorf(
			`expected primary key columns %s got datums %s`, v.primaryKeyCols, primaryKeyDatums)
	} else {
		__antithesis_instrumentation__.Notify(15105)
	}
	__antithesis_instrumentation__.Notify(15098)

	type wrapper struct {
		After  map[string]interface{} `json:"after"`
		Before map[string]interface{} `json:"before"`
	}
	var rowJSON wrapper
	if err := gojson.Unmarshal([]byte(value), &rowJSON); err != nil {
		__antithesis_instrumentation__.Notify(15106)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15107)
	}
	__antithesis_instrumentation__.Notify(15099)

	if err := v.checkRowAt("after", primaryKeyDatums, rowJSON.After, updated); err != nil {
		__antithesis_instrumentation__.Notify(15108)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15109)
	}
	__antithesis_instrumentation__.Notify(15100)

	if v.resolved[partition].IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(15110)
		return rowJSON.Before == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(15111)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(15112)
	}
	__antithesis_instrumentation__.Notify(15101)

	return v.checkRowAt("before", primaryKeyDatums, rowJSON.Before, updated.Prev())
}

func (v *beforeAfterValidator) checkRowAt(
	field string, primaryKeyDatums []interface{}, rowDatums map[string]interface{}, ts hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(15113)
	var stmtBuf bytes.Buffer
	var args []interface{}
	if rowDatums == nil {
		__antithesis_instrumentation__.Notify(15118)

		stmtBuf.WriteString(`SELECT count(*) = 0 `)
	} else {
		__antithesis_instrumentation__.Notify(15119)

		stmtBuf.WriteString(`SELECT count(*) = 1 `)
	}
	__antithesis_instrumentation__.Notify(15114)
	fmt.Fprintf(&stmtBuf, `FROM %s AS OF SYSTEM TIME '%s' WHERE `, v.table, ts.AsOfSystemTime())
	if rowDatums == nil {
		__antithesis_instrumentation__.Notify(15120)

		for i, datum := range primaryKeyDatums {
			__antithesis_instrumentation__.Notify(15121)
			if len(args) != 0 {
				__antithesis_instrumentation__.Notify(15123)
				stmtBuf.WriteString(` AND `)
			} else {
				__antithesis_instrumentation__.Notify(15124)
			}
			__antithesis_instrumentation__.Notify(15122)
			fmt.Fprintf(&stmtBuf, `%s = $%d`, v.primaryKeyCols[i], i+1)
			args = append(args, datum)
		}
	} else {
		__antithesis_instrumentation__.Notify(15125)

		colNames := make([]string, 0, len(rowDatums))
		for col := range rowDatums {
			__antithesis_instrumentation__.Notify(15127)
			colNames = append(colNames, col)
		}
		__antithesis_instrumentation__.Notify(15126)
		sort.Strings(colNames)
		for i, col := range colNames {
			__antithesis_instrumentation__.Notify(15128)
			if len(args) != 0 {
				__antithesis_instrumentation__.Notify(15130)
				stmtBuf.WriteString(` AND `)
			} else {
				__antithesis_instrumentation__.Notify(15131)
			}
			__antithesis_instrumentation__.Notify(15129)
			fmt.Fprintf(&stmtBuf, `%s = $%d`, col, i+1)
			args = append(args, rowDatums[col])
		}
	}
	__antithesis_instrumentation__.Notify(15115)

	var valid bool
	stmt := stmtBuf.String()
	if err := v.sqlDB.QueryRow(stmt, args...).Scan(&valid); err != nil {
		__antithesis_instrumentation__.Notify(15132)
		return errors.Wrapf(err, "while executing %s", stmt)
	} else {
		__antithesis_instrumentation__.Notify(15133)
	}
	__antithesis_instrumentation__.Notify(15116)
	if !valid {
		__antithesis_instrumentation__.Notify(15134)
		v.failures = append(v.failures, fmt.Sprintf(
			"%q field did not agree with row at %s: %s %v",
			field, ts.AsOfSystemTime(), stmt, args))
	} else {
		__antithesis_instrumentation__.Notify(15135)
	}
	__antithesis_instrumentation__.Notify(15117)
	return nil
}

func (v *beforeAfterValidator) NoteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15136)
	prev := v.resolved[partition]
	if prev.Less(resolved) {
		__antithesis_instrumentation__.Notify(15138)
		v.resolved[partition] = resolved
	} else {
		__antithesis_instrumentation__.Notify(15139)
	}
	__antithesis_instrumentation__.Notify(15137)
	return nil
}

func (v *beforeAfterValidator) Failures() []string {
	__antithesis_instrumentation__.Notify(15140)
	return v.failures
}

type validatorRow struct {
	key, value string
	updated    hlc.Timestamp
}

type fingerprintValidator struct {
	sqlDB                  *gosql.DB
	origTable, fprintTable string
	primaryKeyCols         []string
	partitionResolved      map[string]hlc.Timestamp
	resolved               hlc.Timestamp

	firstRowTimestamp hlc.Timestamp

	previousRowUpdateTs hlc.Timestamp

	fprintOrigColumns int
	fprintTestColumns int
	buffer            []validatorRow

	failures []string
}

func NewFingerprintValidator(
	sqlDB *gosql.DB, origTable, fprintTable string, partitions []string, maxTestColumnCount int,
) (Validator, error) {
	__antithesis_instrumentation__.Notify(15141)

	primaryKeyCols, err := fetchPrimaryKeyCols(sqlDB, fprintTable)
	if err != nil {
		__antithesis_instrumentation__.Notify(15146)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15147)
	}
	__antithesis_instrumentation__.Notify(15142)

	var fprintOrigColumns int
	if err := sqlDB.QueryRow(`
		SELECT count(column_name)
		FROM information_schema.columns
		WHERE table_name=$1
	`, fprintTable).Scan(&fprintOrigColumns); err != nil {
		__antithesis_instrumentation__.Notify(15148)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15149)
	}
	__antithesis_instrumentation__.Notify(15143)

	if maxTestColumnCount > 0 {
		__antithesis_instrumentation__.Notify(15150)
		var addColumnStmt bytes.Buffer
		addColumnStmt.WriteString(`ALTER TABLE fprint `)
		for i := 0; i < maxTestColumnCount; i++ {
			__antithesis_instrumentation__.Notify(15152)
			if i != 0 {
				__antithesis_instrumentation__.Notify(15154)
				addColumnStmt.WriteString(`, `)
			} else {
				__antithesis_instrumentation__.Notify(15155)
			}
			__antithesis_instrumentation__.Notify(15153)
			fmt.Fprintf(&addColumnStmt, `ADD COLUMN test%d STRING`, i)
		}
		__antithesis_instrumentation__.Notify(15151)
		if _, err := sqlDB.Exec(addColumnStmt.String()); err != nil {
			__antithesis_instrumentation__.Notify(15156)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(15157)
		}
	} else {
		__antithesis_instrumentation__.Notify(15158)
	}
	__antithesis_instrumentation__.Notify(15144)

	v := &fingerprintValidator{
		sqlDB:             sqlDB,
		origTable:         origTable,
		fprintTable:       fprintTable,
		primaryKeyCols:    primaryKeyCols,
		fprintOrigColumns: fprintOrigColumns,
		fprintTestColumns: maxTestColumnCount,
	}
	v.partitionResolved = make(map[string]hlc.Timestamp)
	for _, partition := range partitions {
		__antithesis_instrumentation__.Notify(15159)
		v.partitionResolved[partition] = hlc.Timestamp{}
	}
	__antithesis_instrumentation__.Notify(15145)
	return v, nil
}

func (v *fingerprintValidator) NoteRow(
	ignoredPartition string, key, value string, updated hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(15160)
	if v.firstRowTimestamp.IsEmpty() || func() bool {
		__antithesis_instrumentation__.Notify(15162)
		return updated.Less(v.firstRowTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15163)
		v.firstRowTimestamp = updated
	} else {
		__antithesis_instrumentation__.Notify(15164)
	}
	__antithesis_instrumentation__.Notify(15161)
	v.buffer = append(v.buffer, validatorRow{
		key:     key,
		value:   value,
		updated: updated,
	})
	return nil
}

func (v *fingerprintValidator) applyRowUpdate(row validatorRow) (_err error) {
	__antithesis_instrumentation__.Notify(15165)
	defer func() {
		__antithesis_instrumentation__.Notify(15171)
		_err = errors.Wrap(_err, "fingerprintValidator failed")
	}()
	__antithesis_instrumentation__.Notify(15166)

	var args []interface{}
	var primaryKeyDatums []interface{}
	if err := gojson.Unmarshal([]byte(row.key), &primaryKeyDatums); err != nil {
		__antithesis_instrumentation__.Notify(15172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15173)
	}
	__antithesis_instrumentation__.Notify(15167)
	if len(primaryKeyDatums) != len(v.primaryKeyCols) {
		__antithesis_instrumentation__.Notify(15174)
		return errors.Errorf(`expected primary key columns %s got datums %s`,
			v.primaryKeyCols, primaryKeyDatums)
	} else {
		__antithesis_instrumentation__.Notify(15175)
	}
	__antithesis_instrumentation__.Notify(15168)

	var stmtBuf bytes.Buffer
	type wrapper struct {
		After map[string]interface{} `json:"after"`
	}
	var value wrapper
	if err := gojson.Unmarshal([]byte(row.value), &value); err != nil {
		__antithesis_instrumentation__.Notify(15176)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15177)
	}
	__antithesis_instrumentation__.Notify(15169)
	if value.After != nil {
		__antithesis_instrumentation__.Notify(15178)

		fmt.Fprintf(&stmtBuf, `UPSERT INTO %s (`, v.fprintTable)
		for col, colValue := range value.After {
			__antithesis_instrumentation__.Notify(15184)
			if len(args) != 0 {
				__antithesis_instrumentation__.Notify(15186)
				stmtBuf.WriteString(`,`)
			} else {
				__antithesis_instrumentation__.Notify(15187)
			}
			__antithesis_instrumentation__.Notify(15185)
			stmtBuf.WriteString(col)
			args = append(args, colValue)
		}
		__antithesis_instrumentation__.Notify(15179)
		for i := len(value.After) - v.fprintOrigColumns; i < v.fprintTestColumns; i++ {
			__antithesis_instrumentation__.Notify(15188)
			fmt.Fprintf(&stmtBuf, `, test%d`, i)
			args = append(args, nil)
		}
		__antithesis_instrumentation__.Notify(15180)
		stmtBuf.WriteString(`) VALUES (`)
		for i := range args {
			__antithesis_instrumentation__.Notify(15189)
			if i != 0 {
				__antithesis_instrumentation__.Notify(15191)
				stmtBuf.WriteString(`,`)
			} else {
				__antithesis_instrumentation__.Notify(15192)
			}
			__antithesis_instrumentation__.Notify(15190)
			fmt.Fprintf(&stmtBuf, `$%d`, i+1)
		}
		__antithesis_instrumentation__.Notify(15181)
		stmtBuf.WriteString(`)`)

		primaryKeyDatums = make([]interface{}, len(v.primaryKeyCols))
		for idx, primaryKeyCol := range v.primaryKeyCols {
			__antithesis_instrumentation__.Notify(15193)
			primaryKeyDatums[idx] = value.After[primaryKeyCol]
		}
		__antithesis_instrumentation__.Notify(15182)
		primaryKeyJSON, err := gojson.Marshal(primaryKeyDatums)
		if err != nil {
			__antithesis_instrumentation__.Notify(15194)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15195)
		}
		__antithesis_instrumentation__.Notify(15183)

		if string(primaryKeyJSON) != row.key {
			__antithesis_instrumentation__.Notify(15196)
			v.failures = append(v.failures,
				fmt.Sprintf(`key %s did not match expected key %s for value %s`,
					row.key, primaryKeyJSON, row.value))
		} else {
			__antithesis_instrumentation__.Notify(15197)
		}
	} else {
		__antithesis_instrumentation__.Notify(15198)

		fmt.Fprintf(&stmtBuf, `DELETE FROM %s WHERE `, v.fprintTable)
		for i, datum := range primaryKeyDatums {
			__antithesis_instrumentation__.Notify(15199)
			if len(args) != 0 {
				__antithesis_instrumentation__.Notify(15201)
				stmtBuf.WriteString(` AND `)
			} else {
				__antithesis_instrumentation__.Notify(15202)
			}
			__antithesis_instrumentation__.Notify(15200)
			fmt.Fprintf(&stmtBuf, `%s = $%d`, v.primaryKeyCols[i], i+1)
			args = append(args, datum)
		}
	}
	__antithesis_instrumentation__.Notify(15170)
	_, err := v.sqlDB.Exec(stmtBuf.String(), args...)
	return err
}

func (v *fingerprintValidator) NoteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15203)
	if r, ok := v.partitionResolved[partition]; !ok {
		__antithesis_instrumentation__.Notify(15210)
		return errors.Errorf(`unknown partition: %s`, partition)
	} else {
		__antithesis_instrumentation__.Notify(15211)
		if resolved.LessEq(r) {
			__antithesis_instrumentation__.Notify(15212)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(15213)
		}
	}
	__antithesis_instrumentation__.Notify(15204)
	v.partitionResolved[partition] = resolved

	newResolved := resolved
	for _, r := range v.partitionResolved {
		__antithesis_instrumentation__.Notify(15214)
		if r.Less(newResolved) {
			__antithesis_instrumentation__.Notify(15215)
			newResolved = r
		} else {
			__antithesis_instrumentation__.Notify(15216)
		}
	}
	__antithesis_instrumentation__.Notify(15205)
	if newResolved.LessEq(v.resolved) {
		__antithesis_instrumentation__.Notify(15217)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15218)
	}
	__antithesis_instrumentation__.Notify(15206)
	v.resolved = newResolved

	sort.Slice(v.buffer, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(15219)
		return v.buffer[i].updated.Less(v.buffer[j].updated)
	})
	__antithesis_instrumentation__.Notify(15207)

	var lastFingerprintedAt hlc.Timestamp

	for len(v.buffer) > 0 {
		__antithesis_instrumentation__.Notify(15220)
		if v.resolved.Less(v.buffer[0].updated) {
			__antithesis_instrumentation__.Notify(15225)
			break
		} else {
			__antithesis_instrumentation__.Notify(15226)
		}
		__antithesis_instrumentation__.Notify(15221)
		row := v.buffer[0]
		v.buffer = v.buffer[1:]

		if !v.previousRowUpdateTs.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(15227)
			return v.previousRowUpdateTs.Less(row.updated) == true
		}() == true {
			__antithesis_instrumentation__.Notify(15228)
			if err := v.fingerprint(row.updated.Prev()); err != nil {
				__antithesis_instrumentation__.Notify(15229)
				return err
			} else {
				__antithesis_instrumentation__.Notify(15230)
			}
		} else {
			__antithesis_instrumentation__.Notify(15231)
		}
		__antithesis_instrumentation__.Notify(15222)
		if err := v.applyRowUpdate(row); err != nil {
			__antithesis_instrumentation__.Notify(15232)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15233)
		}
		__antithesis_instrumentation__.Notify(15223)

		if len(v.buffer) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(15234)
			return v.buffer[0].updated != row.updated == true
		}() == true {
			__antithesis_instrumentation__.Notify(15235)
			lastFingerprintedAt = row.updated
			if err := v.fingerprint(row.updated); err != nil {
				__antithesis_instrumentation__.Notify(15236)
				return err
			} else {
				__antithesis_instrumentation__.Notify(15237)
			}
		} else {
			__antithesis_instrumentation__.Notify(15238)
		}
		__antithesis_instrumentation__.Notify(15224)
		v.previousRowUpdateTs = row.updated
	}
	__antithesis_instrumentation__.Notify(15208)

	if !v.firstRowTimestamp.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(15239)
		return v.firstRowTimestamp.LessEq(resolved) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(15240)
		return lastFingerprintedAt != resolved == true
	}() == true {
		__antithesis_instrumentation__.Notify(15241)
		return v.fingerprint(resolved)
	} else {
		__antithesis_instrumentation__.Notify(15242)
	}
	__antithesis_instrumentation__.Notify(15209)
	return nil
}

func (v *fingerprintValidator) fingerprint(ts hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15243)
	var orig string
	if err := v.sqlDB.QueryRow(`SELECT IFNULL(fingerprint, 'EMPTY') FROM [
		SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE ` + v.origTable + `
	] AS OF SYSTEM TIME '` + ts.AsOfSystemTime() + `'`).Scan(&orig); err != nil {
		__antithesis_instrumentation__.Notify(15247)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15248)
	}
	__antithesis_instrumentation__.Notify(15244)
	var check string
	if err := v.sqlDB.QueryRow(`SELECT IFNULL(fingerprint, 'EMPTY') FROM [
		SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE ` + v.fprintTable + `
	]`).Scan(&check); err != nil {
		__antithesis_instrumentation__.Notify(15249)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15250)
	}
	__antithesis_instrumentation__.Notify(15245)
	if orig != check {
		__antithesis_instrumentation__.Notify(15251)
		v.failures = append(v.failures, fmt.Sprintf(
			`fingerprints did not match at %s: %s vs %s`, ts.AsOfSystemTime(), orig, check))
	} else {
		__antithesis_instrumentation__.Notify(15252)
	}
	__antithesis_instrumentation__.Notify(15246)
	return nil
}

func (v *fingerprintValidator) Failures() []string {
	__antithesis_instrumentation__.Notify(15253)
	return v.failures
}

type Validators []Validator

func (vs Validators) NoteRow(partition string, key, value string, updated hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15254)
	for _, v := range vs {
		__antithesis_instrumentation__.Notify(15256)
		if err := v.NoteRow(partition, key, value, updated); err != nil {
			__antithesis_instrumentation__.Notify(15257)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15258)
		}
	}
	__antithesis_instrumentation__.Notify(15255)
	return nil
}

func (vs Validators) NoteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15259)
	for _, v := range vs {
		__antithesis_instrumentation__.Notify(15261)
		if err := v.NoteResolved(partition, resolved); err != nil {
			__antithesis_instrumentation__.Notify(15262)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15263)
		}
	}
	__antithesis_instrumentation__.Notify(15260)
	return nil
}

func (vs Validators) Failures() []string {
	__antithesis_instrumentation__.Notify(15264)
	var f []string
	for _, v := range vs {
		__antithesis_instrumentation__.Notify(15266)
		f = append(f, v.Failures()...)
	}
	__antithesis_instrumentation__.Notify(15265)
	return f
}

type CountValidator struct {
	v Validator

	NumRows, NumResolved                 int
	NumResolvedRows, NumResolvedWithRows int
	rowsSinceResolved                    int
}

func MakeCountValidator(v Validator) *CountValidator {
	__antithesis_instrumentation__.Notify(15267)
	return &CountValidator{v: v}
}

func (v *CountValidator) NoteRow(partition string, key, value string, updated hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15268)
	v.NumRows++
	v.rowsSinceResolved++
	return v.v.NoteRow(partition, key, value, updated)
}

func (v *CountValidator) NoteResolved(partition string, resolved hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(15269)
	v.NumResolved++
	if v.rowsSinceResolved > 0 {
		__antithesis_instrumentation__.Notify(15271)
		v.NumResolvedWithRows++
		v.NumResolvedRows += v.rowsSinceResolved
		v.rowsSinceResolved = 0
	} else {
		__antithesis_instrumentation__.Notify(15272)
	}
	__antithesis_instrumentation__.Notify(15270)
	return v.v.NoteResolved(partition, resolved)
}

func (v *CountValidator) Failures() []string {
	__antithesis_instrumentation__.Notify(15273)
	return v.v.Failures()
}

func ParseJSONValueTimestamps(v []byte) (updated, resolved hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(15274)
	var valueRaw struct {
		Resolved string `json:"resolved"`
		Updated  string `json:"updated"`
	}
	if err := gojson.Unmarshal(v, &valueRaw); err != nil {
		__antithesis_instrumentation__.Notify(15278)
		return hlc.Timestamp{}, hlc.Timestamp{}, errors.Wrapf(err, "parsing [%s] as json", v)
	} else {
		__antithesis_instrumentation__.Notify(15279)
	}
	__antithesis_instrumentation__.Notify(15275)
	if valueRaw.Updated != `` {
		__antithesis_instrumentation__.Notify(15280)
		var err error
		updated, err = tree.ParseHLC(valueRaw.Updated)
		if err != nil {
			__antithesis_instrumentation__.Notify(15281)
			return hlc.Timestamp{}, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(15282)
		}
	} else {
		__antithesis_instrumentation__.Notify(15283)
	}
	__antithesis_instrumentation__.Notify(15276)
	if valueRaw.Resolved != `` {
		__antithesis_instrumentation__.Notify(15284)
		var err error
		resolved, err = tree.ParseHLC(valueRaw.Resolved)
		if err != nil {
			__antithesis_instrumentation__.Notify(15285)
			return hlc.Timestamp{}, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(15286)
		}
	} else {
		__antithesis_instrumentation__.Notify(15287)
	}
	__antithesis_instrumentation__.Notify(15277)
	return updated, resolved, nil
}

func fetchPrimaryKeyCols(sqlDB *gosql.DB, tableStr string) ([]string, error) {
	__antithesis_instrumentation__.Notify(15288)
	parts := strings.Split(tableStr, ".")
	var db, table string
	switch len(parts) {
	case 1:
		__antithesis_instrumentation__.Notify(15295)
		table = parts[0]
	case 2:
		__antithesis_instrumentation__.Notify(15296)
		db = parts[0] + "."
		table = parts[1]
	default:
		__antithesis_instrumentation__.Notify(15297)
		return nil, errors.Errorf("could not parse table %s", parts)
	}
	__antithesis_instrumentation__.Notify(15289)
	rows, err := sqlDB.Query(fmt.Sprintf(`
		SELECT column_name
		FROM %sinformation_schema.key_column_usage
		WHERE table_name=$1
			AND constraint_name=($1||'_pkey')
		ORDER BY ordinal_position`, db),
		table,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(15298)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15299)
	}
	__antithesis_instrumentation__.Notify(15290)
	defer func() { __antithesis_instrumentation__.Notify(15300); _ = rows.Close() }()
	__antithesis_instrumentation__.Notify(15291)
	var primaryKeyCols []string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(15301)
		var primaryKeyCol string
		if err := rows.Scan(&primaryKeyCol); err != nil {
			__antithesis_instrumentation__.Notify(15303)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(15304)
		}
		__antithesis_instrumentation__.Notify(15302)
		primaryKeyCols = append(primaryKeyCols, primaryKeyCol)
	}
	__antithesis_instrumentation__.Notify(15292)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(15305)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15306)
	}
	__antithesis_instrumentation__.Notify(15293)
	if len(primaryKeyCols) == 0 {
		__antithesis_instrumentation__.Notify(15307)
		return nil, errors.Errorf("no primary key information found for %s", tableStr)
	} else {
		__antithesis_instrumentation__.Notify(15308)
	}
	__antithesis_instrumentation__.Notify(15294)
	return primaryKeyCols, nil
}
