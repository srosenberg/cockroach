package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type ResultColumn struct {
	Name string
	Typ  *types.T

	Hidden bool

	TableID        descpb.ID
	PGAttributeNum uint32
}

type ResultColumns []ResultColumn

func ResultColumnsFromColumns(tableID descpb.ID, columns []catalog.Column) ResultColumns {
	__antithesis_instrumentation__.Notify(251176)
	return ResultColumnsFromColDescs(tableID, len(columns), func(i int) *descpb.ColumnDescriptor {
		__antithesis_instrumentation__.Notify(251177)
		return columns[i].ColumnDesc()
	})
}

func ResultColumnsFromColDescs(
	tableID descpb.ID, numCols int, getColDesc func(int) *descpb.ColumnDescriptor,
) ResultColumns {
	__antithesis_instrumentation__.Notify(251178)
	cols := make(ResultColumns, numCols)
	for i := range cols {
		__antithesis_instrumentation__.Notify(251180)
		colDesc := getColDesc(i)
		typ := colDesc.Type
		if typ == nil {
			__antithesis_instrumentation__.Notify(251182)
			panic(errors.AssertionFailedf("unsupported column type: %s", colDesc.Type.Family()))
		} else {
			__antithesis_instrumentation__.Notify(251183)
		}
		__antithesis_instrumentation__.Notify(251181)
		cols[i] = ResultColumn{
			Name:           colDesc.Name,
			Typ:            typ,
			Hidden:         colDesc.Hidden,
			TableID:        tableID,
			PGAttributeNum: colDesc.GetPGAttributeNum(),
		}
	}
	__antithesis_instrumentation__.Notify(251179)
	return cols
}

func (r ResultColumn) GetTypeModifier() int32 {
	__antithesis_instrumentation__.Notify(251184)
	return r.Typ.TypeModifier()
}

func (r ResultColumns) TypesEqual(other ResultColumns) bool {
	__antithesis_instrumentation__.Notify(251185)
	if len(r) != len(other) {
		__antithesis_instrumentation__.Notify(251188)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251189)
	}
	__antithesis_instrumentation__.Notify(251186)
	for i, c := range r {
		__antithesis_instrumentation__.Notify(251190)

		if other[i].Typ.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(251192)
			continue
		} else {
			__antithesis_instrumentation__.Notify(251193)
		}
		__antithesis_instrumentation__.Notify(251191)
		if !c.Typ.Equivalent(other[i].Typ) {
			__antithesis_instrumentation__.Notify(251194)
			return false
		} else {
			__antithesis_instrumentation__.Notify(251195)
		}
	}
	__antithesis_instrumentation__.Notify(251187)
	return true
}

func (r ResultColumns) NodeFormatter(colIdx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(251196)
	return &varFormatter{ColumnName: tree.Name(r[colIdx].Name)}
}

func (r ResultColumns) String(printTypes bool, showHidden bool) string {
	__antithesis_instrumentation__.Notify(251197)
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.WriteByte('(')
	for i := range r {
		__antithesis_instrumentation__.Notify(251199)
		rCol := &r[i]
		if i > 0 {
			__antithesis_instrumentation__.Notify(251204)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(251205)
		}
		__antithesis_instrumentation__.Notify(251200)
		f.FormatNameP(&rCol.Name)

		hasProps := false
		outputProp := func(prop string) {
			__antithesis_instrumentation__.Notify(251206)
			if hasProps {
				__antithesis_instrumentation__.Notify(251208)
				f.WriteByte(',')
			} else {
				__antithesis_instrumentation__.Notify(251209)
				f.WriteByte('[')
			}
			__antithesis_instrumentation__.Notify(251207)
			hasProps = true
			f.WriteString(prop)
		}
		__antithesis_instrumentation__.Notify(251201)
		if showHidden && func() bool {
			__antithesis_instrumentation__.Notify(251210)
			return rCol.Hidden == true
		}() == true {
			__antithesis_instrumentation__.Notify(251211)
			outputProp("hidden")
		} else {
			__antithesis_instrumentation__.Notify(251212)
		}
		__antithesis_instrumentation__.Notify(251202)
		if hasProps {
			__antithesis_instrumentation__.Notify(251213)
			f.WriteByte(']')
		} else {
			__antithesis_instrumentation__.Notify(251214)
		}
		__antithesis_instrumentation__.Notify(251203)

		if printTypes {
			__antithesis_instrumentation__.Notify(251215)
			f.WriteByte(' ')
			f.WriteString(rCol.Typ.String())
		} else {
			__antithesis_instrumentation__.Notify(251216)
		}
	}
	__antithesis_instrumentation__.Notify(251198)
	f.WriteByte(')')
	return f.CloseAndGetString()
}

var ExplainPlanColumns = ResultColumns{
	{Name: "info", Typ: types.String},
}

var ShowTraceColumns = ResultColumns{
	{Name: "timestamp", Typ: types.TimestampTZ},
	{Name: "age", Typ: types.Interval},
	{Name: "message", Typ: types.String},
	{Name: "tag", Typ: types.String},
	{Name: "location", Typ: types.String},
	{Name: "operation", Typ: types.String},
	{Name: "span", Typ: types.Int},
}

var ShowCompactTraceColumns = ResultColumns{
	{Name: "age", Typ: types.Interval},
	{Name: "message", Typ: types.String},
	{Name: "tag", Typ: types.String},
	{Name: "operation", Typ: types.String},
}

func GetTraceAgeColumnIdx(compact bool) int {
	__antithesis_instrumentation__.Notify(251217)
	if compact {
		__antithesis_instrumentation__.Notify(251219)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(251220)
	}
	__antithesis_instrumentation__.Notify(251218)
	return 1
}

var ShowReplicaTraceColumns = ResultColumns{
	{Name: "timestamp", Typ: types.TimestampTZ},
	{Name: "node_id", Typ: types.Int},
	{Name: "store_id", Typ: types.Int},
	{Name: "replica_id", Typ: types.Int},
}

var ShowSyntaxColumns = ResultColumns{
	{Name: "field", Typ: types.String},
	{Name: "message", Typ: types.String},
}

var ShowFingerprintsColumns = ResultColumns{
	{Name: "index_name", Typ: types.String},
	{Name: "fingerprint", Typ: types.String},
}

var AlterTableSplitColumns = ResultColumns{
	{Name: "key", Typ: types.Bytes},
	{Name: "pretty", Typ: types.String},
	{Name: "split_enforced_until", Typ: types.Timestamp},
}

var AlterTableUnsplitColumns = ResultColumns{
	{Name: "key", Typ: types.Bytes},
	{Name: "pretty", Typ: types.String},
}

var AlterTableRelocateColumns = ResultColumns{
	{Name: "key", Typ: types.Bytes},
	{Name: "pretty", Typ: types.String},
}

var AlterTableScatterColumns = ResultColumns{
	{Name: "key", Typ: types.Bytes},
	{Name: "pretty", Typ: types.String},
}

var AlterRangeRelocateColumns = ResultColumns{
	{Name: "range_id", Typ: types.Int},
	{Name: "pretty", Typ: types.String},
	{Name: "result", Typ: types.String},
}

var ScrubColumns = ResultColumns{
	{Name: "job_uuid", Typ: types.Uuid},
	{Name: "error_type", Typ: types.String},
	{Name: "database", Typ: types.String},
	{Name: "table", Typ: types.String},
	{Name: "primary_key", Typ: types.String},
	{Name: "timestamp", Typ: types.Timestamp},
	{Name: "repaired", Typ: types.Bool},
	{Name: "details", Typ: types.Jsonb},
}

var SequenceSelectColumns = ResultColumns{
	{Name: `last_value`, Typ: types.Int},
	{Name: `log_cnt`, Typ: types.Int},
	{Name: `is_called`, Typ: types.Bool},
}

var ExportColumns = ResultColumns{
	{Name: "filename", Typ: types.String},
	{Name: "rows", Typ: types.Int},
	{Name: "bytes", Typ: types.Int},
}
