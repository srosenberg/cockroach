package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/lib/pq/oid"
)

const (
	MySQL    = "mysql"
	Postgres = "postgres"
)

const GetPGMetadataSQL = `
	SELECT
		c.relname AS table_name,
		a.attname AS column_name,
		t.typname AS data_type,
		t.oid AS data_type_oid
	FROM pg_class c
	JOIN pg_attribute a ON a.attrelid = c.oid
	JOIN pg_type t ON t.oid = a.atttypid
	JOIN pg_namespace n ON n.oid = c.relnamespace
	WHERE n.nspname = $1
	AND a.attnum > 0
  AND c.relkind != 'i';
`

type Summary struct {
	TotalTables        int
	TotalColumns       int
	MissingTables      int
	MissingColumns     int
	DatatypeMismatches int
}

type PGMetadataColumnDiff struct {
	Oid              uint32 `json:"oid"`
	DataType         string `json:"dataType"`
	ExpectedOid      uint32 `json:"expectedOid"`
	ExpectedDataType string `json:"expectedDataType"`
}

type PGMetadataColumnDiffs map[string]*PGMetadataColumnDiff

type PGMetadataTableDiffs map[string]PGMetadataColumnDiffs

type PGMetadataColumnType struct {
	Oid      uint32 `json:"oid"`
	DataType string `json:"dataType"`
}

type PGMetadataColumns map[string]*PGMetadataColumnType

type PGMetadataTableInfo struct {
	ColumnNames []string          `json:"columnNames"`
	Columns     PGMetadataColumns `json:"columns"`
}

type PGMetadataTables map[string]PGMetadataTableInfo

type PGMetadataFile struct {
	Version    string           `json:"version"`
	PGMetadata PGMetadataTables `json:"tables"`
}

type PGMetadataDiffFile struct {
	Version            string               `json:"version"`
	DiffSummary        Summary              `json:"diffSummary"`
	Diffs              PGMetadataTableDiffs `json:"diffs"`
	UnimplementedTypes map[oid.Oid]string   `json:"unimplementedTypes"`
}

func (d PGMetadataTableDiffs) addColumn(
	tableName, columnName string, column *PGMetadataColumnDiff,
) {
	__antithesis_instrumentation__.Notify(558781)
	columns, ok := d[tableName]

	if !ok {
		__antithesis_instrumentation__.Notify(558783)
		columns = make(PGMetadataColumnDiffs)
		d[tableName] = columns
	} else {
		__antithesis_instrumentation__.Notify(558784)
	}
	__antithesis_instrumentation__.Notify(558782)

	columns[columnName] = column
}

func (p PGMetadataTables) addColumn(
	tableName string, columnName string, column *PGMetadataColumnType,
) {
	__antithesis_instrumentation__.Notify(558785)
	tableInfo, ok := p[tableName]

	if !ok {
		__antithesis_instrumentation__.Notify(558787)
		tableInfo = PGMetadataTableInfo{
			ColumnNames: []string{},
			Columns:     make(PGMetadataColumns),
		}
		p[tableName] = tableInfo
	} else {
		__antithesis_instrumentation__.Notify(558788)
	}
	__antithesis_instrumentation__.Notify(558786)

	tableInfo.ColumnNames = append(tableInfo.ColumnNames, columnName)
	tableInfo.Columns[columnName] = column
	p[tableName] = tableInfo
}

func (p PGMetadataTables) AddColumnMetadata(
	tableName string, columnName string, dataType string, dataTypeOid uint32,
) {
	__antithesis_instrumentation__.Notify(558789)
	p.addColumn(tableName, columnName, &PGMetadataColumnType{
		dataTypeOid,
		dataType,
	})
}

func (d PGMetadataTableDiffs) addDiff(
	tableName string, columnName string, expected *PGMetadataColumnType, actual *PGMetadataColumnType,
) {
	__antithesis_instrumentation__.Notify(558790)
	d.addColumn(tableName, columnName, &PGMetadataColumnDiff{
		actual.Oid,
		actual.DataType,
		expected.Oid,
		expected.DataType,
	})
}

type compareResult int

const (
	success compareResult = 1 + iota
	expectedDiffError
	diffError
)

func (d PGMetadataTableDiffs) compareColumns(
	tableName string, columnName string, expected *PGMetadataColumnType, actual *PGMetadataColumnType,
) compareResult {
	__antithesis_instrumentation__.Notify(558791)

	if expected.Oid == 0 {
		__antithesis_instrumentation__.Notify(558794)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(558795)
	}
	__antithesis_instrumentation__.Notify(558792)

	expectedDiff := d.getExpectedDiff(tableName, columnName)

	if actual.Oid == expected.Oid {
		__antithesis_instrumentation__.Notify(558796)
		if expectedDiff != nil {
			__antithesis_instrumentation__.Notify(558797)

			return expectedDiffError
		} else {
			__antithesis_instrumentation__.Notify(558798)
		}
	} else {
		__antithesis_instrumentation__.Notify(558799)
		if expectedDiff == nil || func() bool {
			__antithesis_instrumentation__.Notify(558800)
			return expectedDiff.Oid != actual.Oid == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(558801)
			return expectedDiff.ExpectedOid != expected.Oid == true
		}() == true {
			__antithesis_instrumentation__.Notify(558802)

			return diffError
		} else {
			__antithesis_instrumentation__.Notify(558803)
		}
	}
	__antithesis_instrumentation__.Notify(558793)

	return success
}

func (d PGMetadataTableDiffs) getExpectedDiff(tableName, columnName string) *PGMetadataColumnDiff {
	__antithesis_instrumentation__.Notify(558804)
	columns, ok := d[tableName]
	if !ok {
		__antithesis_instrumentation__.Notify(558806)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(558807)
	}
	__antithesis_instrumentation__.Notify(558805)

	return columns[columnName]
}

func (d PGMetadataTableDiffs) isExpectedMissingTable(tableName string) bool {
	__antithesis_instrumentation__.Notify(558808)
	if columns, ok := d[tableName]; !ok || func() bool {
		__antithesis_instrumentation__.Notify(558810)
		return len(columns) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(558811)
		return false
	} else {
		__antithesis_instrumentation__.Notify(558812)
	}
	__antithesis_instrumentation__.Notify(558809)

	return true
}

func (d PGMetadataTableDiffs) isExpectedMissingColumn(tableName string, columnName string) bool {
	__antithesis_instrumentation__.Notify(558813)
	columns, ok := d[tableName]
	if !ok {
		__antithesis_instrumentation__.Notify(558816)
		return false
	} else {
		__antithesis_instrumentation__.Notify(558817)
	}
	__antithesis_instrumentation__.Notify(558814)

	diff, ok := columns[columnName]
	if !ok {
		__antithesis_instrumentation__.Notify(558818)
		return false
	} else {
		__antithesis_instrumentation__.Notify(558819)
	}
	__antithesis_instrumentation__.Notify(558815)

	return diff == nil
}

func (d PGMetadataTableDiffs) addMissingTable(tableName string) {
	__antithesis_instrumentation__.Notify(558820)
	d[tableName] = make(PGMetadataColumnDiffs)
}

func (d PGMetadataTableDiffs) addMissingColumn(tableName string, columnName string) {
	__antithesis_instrumentation__.Notify(558821)
	columns, ok := d[tableName]

	if !ok {
		__antithesis_instrumentation__.Notify(558823)
		columns = make(PGMetadataColumnDiffs)
		d[tableName] = columns
	} else {
		__antithesis_instrumentation__.Notify(558824)
	}
	__antithesis_instrumentation__.Notify(558822)

	columns[columnName] = nil
}

func (d PGMetadataTableDiffs) rewriteDiffs(
	source PGMetadataFile, sum Summary, diffFile string,
) error {
	__antithesis_instrumentation__.Notify(558825)
	f, err := os.OpenFile(diffFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		__antithesis_instrumentation__.Notify(558827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(558828)
	}
	__antithesis_instrumentation__.Notify(558826)
	defer f.Close()

	mf := &PGMetadataDiffFile{
		Version:            source.Version,
		Diffs:              d,
		DiffSummary:        sum,
		UnimplementedTypes: source.PGMetadata.getUnimplementedTypes(),
	}
	mf.Save(f)
	return nil
}

func (df *PGMetadataDiffFile) Save(writer io.Writer) {
	__antithesis_instrumentation__.Notify(558829)
	Save(writer, df)
}

func (f *PGMetadataFile) Save(writer io.Writer) {
	__antithesis_instrumentation__.Notify(558830)
	Save(writer, f)
}

func Save(writer io.Writer, file interface{}) {
	__antithesis_instrumentation__.Notify(558831)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(file); err != nil {
		__antithesis_instrumentation__.Notify(558832)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(558833)
	}
}

func (d PGMetadataTableDiffs) getUnimplementedTables(source PGMetadataTables) PGMetadataTables {
	__antithesis_instrumentation__.Notify(558834)
	unimplementedTables := make(PGMetadataTables)
	for tableName := range d {
		__antithesis_instrumentation__.Notify(558836)
		if len(d[tableName]) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(558837)
			return len(source[tableName].Columns.getUnimplementedTypes()) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(558838)
			unimplementedTables[tableName] = source[tableName]
		} else {
			__antithesis_instrumentation__.Notify(558839)
		}
	}
	__antithesis_instrumentation__.Notify(558835)
	return unimplementedTables
}

func (d PGMetadataTableDiffs) getUnimplementedColumns(target PGMetadataTables) PGMetadataTables {
	__antithesis_instrumentation__.Notify(558840)
	unimplementedColumns := make(PGMetadataTables)
	for tableName, columns := range d {
		__antithesis_instrumentation__.Notify(558842)
		for columnName, columnType := range columns {
			__antithesis_instrumentation__.Notify(558843)
			if columnType != nil {
				__antithesis_instrumentation__.Notify(558847)

				continue
			} else {
				__antithesis_instrumentation__.Notify(558848)
			}
			__antithesis_instrumentation__.Notify(558844)
			sourceType, ok := target[tableName].Columns[columnName]
			if !ok {
				__antithesis_instrumentation__.Notify(558849)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558850)
			}
			__antithesis_instrumentation__.Notify(558845)
			typeOid := oid.Oid(sourceType.Oid)
			if _, ok := types.OidToType[typeOid]; !ok || func() bool {
				__antithesis_instrumentation__.Notify(558851)
				return typeOid == oid.T_anyarray == true
			}() == true {
				__antithesis_instrumentation__.Notify(558852)

				continue
			} else {
				__antithesis_instrumentation__.Notify(558853)
			}
			__antithesis_instrumentation__.Notify(558846)
			unimplementedColumns.AddColumnMetadata(tableName, columnName, sourceType.DataType, sourceType.Oid)
		}
	}
	__antithesis_instrumentation__.Notify(558841)
	return unimplementedColumns
}

func (d PGMetadataTableDiffs) removeImplementedColumns(source PGMetadataTables) {
	__antithesis_instrumentation__.Notify(558854)
	for tableName, tableInfo := range source {
		__antithesis_instrumentation__.Notify(558855)
		pColumns, exists := d[tableName]
		if !exists {
			__antithesis_instrumentation__.Notify(558857)
			continue
		} else {
			__antithesis_instrumentation__.Notify(558858)
		}
		__antithesis_instrumentation__.Notify(558856)
		for _, columnName := range tableInfo.ColumnNames {
			__antithesis_instrumentation__.Notify(558859)
			columnType, exists := pColumns[columnName]
			if !exists {
				__antithesis_instrumentation__.Notify(558862)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558863)
			}
			__antithesis_instrumentation__.Notify(558860)
			if columnType != nil {
				__antithesis_instrumentation__.Notify(558864)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558865)
			}
			__antithesis_instrumentation__.Notify(558861)

			delete(pColumns, columnName)
		}
	}
}

func (c PGMetadataColumns) getUnimplementedTypes() map[oid.Oid]string {
	__antithesis_instrumentation__.Notify(558866)
	unimplemented := make(map[oid.Oid]string)
	for _, column := range c {
		__antithesis_instrumentation__.Notify(558868)
		typeOid := oid.Oid(column.Oid)
		if _, ok := types.OidToType[typeOid]; !ok || func() bool {
			__antithesis_instrumentation__.Notify(558869)
			return typeOid == oid.T_anyarray == true
		}() == true {
			__antithesis_instrumentation__.Notify(558870)
			unimplemented[typeOid] = column.DataType
		} else {
			__antithesis_instrumentation__.Notify(558871)
		}
	}
	__antithesis_instrumentation__.Notify(558867)

	return unimplemented
}

func (p PGMetadataTables) getUnimplementedTypes() map[oid.Oid]string {
	__antithesis_instrumentation__.Notify(558872)
	unimplemented := make(map[oid.Oid]string)
	for _, tableInfo := range p {
		__antithesis_instrumentation__.Notify(558874)
		for typeOid, dataType := range tableInfo.Columns.getUnimplementedTypes() {
			__antithesis_instrumentation__.Notify(558875)
			unimplemented[typeOid] = dataType
		}
	}
	__antithesis_instrumentation__.Notify(558873)

	return unimplemented
}

func TablesMetadataFilename(path, rdbms, schema string) string {
	__antithesis_instrumentation__.Notify(558876)
	return filepath.Join(
		path,
		fmt.Sprintf("%s_tables_from_%s.json", schema, rdbms),
	)
}
