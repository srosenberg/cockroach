package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

const (
	MVCCTimestampColumnID descpb.ColumnID = math.MaxUint32 - iota

	TableOIDColumnID

	numSystemColumns = iota
)

func init() {
	if len(AllSystemColumnDescs) != numSystemColumns {
		panic("need to update system column IDs or descriptors")
	}
	if catalog.NumSystemColumns != numSystemColumns {
		panic("need to update catalog.NumSystemColumns")
	}
	if catalog.SmallestSystemColumnColumnID != math.MaxUint32-numSystemColumns+1 {
		panic("need to update catalog.SmallestSystemColumnColumnID")
	}
	for _, desc := range AllSystemColumnDescs {
		if desc.SystemColumnKind != GetSystemColumnKindFromColumnID(desc.ID) {
			panic("system column ID ordering must match SystemColumnKind value ordering")
		}
	}
}

var AllSystemColumnDescs = []descpb.ColumnDescriptor{
	MVCCTimestampColumnDesc,
	TableOIDColumnDesc,
}

var MVCCTimestampColumnDesc = descpb.ColumnDescriptor{
	Name:             MVCCTimestampColumnName,
	Type:             MVCCTimestampColumnType,
	Hidden:           true,
	Nullable:         true,
	SystemColumnKind: catpb.SystemColumnKind_MVCCTIMESTAMP,
	ID:               MVCCTimestampColumnID,
}

const MVCCTimestampColumnName = "crdb_internal_mvcc_timestamp"

var MVCCTimestampColumnType = types.Decimal

var TableOIDColumnDesc = descpb.ColumnDescriptor{
	Name:             TableOIDColumnName,
	Type:             types.Oid,
	Hidden:           true,
	Nullable:         true,
	SystemColumnKind: catpb.SystemColumnKind_TABLEOID,
	ID:               TableOIDColumnID,
}

const TableOIDColumnName = "tableoid"

func IsColIDSystemColumn(colID descpb.ColumnID) bool {
	__antithesis_instrumentation__.Notify(251221)
	return colID > math.MaxUint32-numSystemColumns
}

func GetSystemColumnKindFromColumnID(colID descpb.ColumnID) catpb.SystemColumnKind {
	__antithesis_instrumentation__.Notify(251222)
	i := math.MaxUint32 - uint32(colID)
	if i >= numSystemColumns {
		__antithesis_instrumentation__.Notify(251224)
		return catpb.SystemColumnKind_NONE
	} else {
		__antithesis_instrumentation__.Notify(251225)
	}
	__antithesis_instrumentation__.Notify(251223)
	return catpb.SystemColumnKind_NONE + 1 + catpb.SystemColumnKind(i)
}

func IsSystemColumnName(name string) bool {
	__antithesis_instrumentation__.Notify(251226)
	for i := range AllSystemColumnDescs {
		__antithesis_instrumentation__.Notify(251228)
		if AllSystemColumnDescs[i].Name == name {
			__antithesis_instrumentation__.Notify(251229)
			return true
		} else {
			__antithesis_instrumentation__.Notify(251230)
		}
	}
	__antithesis_instrumentation__.Notify(251227)
	return false
}
