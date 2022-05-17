package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
)

type TableElementMaybeMutation interface {
	IsMutation() bool

	IsRollback() bool

	MutationID() descpb.MutationID

	WriteAndDeleteOnly() bool

	DeleteOnly() bool

	Backfilling() bool

	Merging() bool

	Adding() bool

	Dropped() bool
}

type Mutation interface {
	TableElementMaybeMutation

	AsColumn() Column

	AsIndex() Index

	AsConstraint() ConstraintToUpdate

	AsPrimaryKeySwap() PrimaryKeySwap

	AsComputedColumnSwap() ComputedColumnSwap

	AsMaterializedViewRefresh() MaterializedViewRefresh

	AsModifyRowLevelTTL() ModifyRowLevelTTL

	MutationOrdinal() int
}

type Index interface {
	TableElementMaybeMutation

	IndexDesc() *descpb.IndexDescriptor

	IndexDescDeepCopy() descpb.IndexDescriptor

	Ordinal() int

	Primary() bool

	Public() bool

	GetID() descpb.IndexID
	GetConstraintID() descpb.ConstraintID
	GetName() string
	IsPartial() bool
	IsUnique() bool
	IsDisabled() bool
	IsSharded() bool
	IsCreatedExplicitly() bool
	GetPredicate() string
	GetType() descpb.IndexDescriptor_Type
	GetGeoConfig() geoindex.Config
	GetVersion() descpb.IndexDescriptorVersion
	GetEncodingType() descpb.IndexDescriptorEncodingType

	GetSharded() catpb.ShardedDescriptor
	GetShardColumnName() string

	IsValidOriginIndex(originColIDs descpb.ColumnIDs) bool
	IsValidReferencedUniqueConstraint(referencedColIDs descpb.ColumnIDs) bool

	GetPartitioning() Partitioning

	ExplicitColumnStartIdx() int

	NumKeyColumns() int
	GetKeyColumnID(columnOrdinal int) descpb.ColumnID
	GetKeyColumnName(columnOrdinal int) string
	GetKeyColumnDirection(columnOrdinal int) descpb.IndexDescriptor_Direction

	CollectKeyColumnIDs() TableColSet
	CollectKeySuffixColumnIDs() TableColSet
	CollectPrimaryStoredColumnIDs() TableColSet
	CollectSecondaryStoredColumnIDs() TableColSet
	CollectCompositeColumnIDs() TableColSet

	InvertedColumnID() descpb.ColumnID

	InvertedColumnName() string

	InvertedColumnKeyType() *types.T

	NumPrimaryStoredColumns() int
	NumSecondaryStoredColumns() int
	GetStoredColumnID(storedColumnOrdinal int) descpb.ColumnID
	GetStoredColumnName(storedColumnOrdinal int) string
	HasOldStoredColumns() bool

	NumKeySuffixColumns() int
	GetKeySuffixColumnID(extraColumnOrdinal int) descpb.ColumnID

	NumCompositeColumns() int
	GetCompositeColumnID(compositeColumnOrdinal int) descpb.ColumnID
	UseDeletePreservingEncoding() bool

	ForcePut() bool

	CreatedAt() time.Time

	IsTemporaryIndexForBackfill() bool
}

type Column interface {
	TableElementMaybeMutation

	ColumnDesc() *descpb.ColumnDescriptor

	ColumnDescDeepCopy() descpb.ColumnDescriptor

	DeepCopy() Column

	Ordinal() int

	Public() bool

	GetID() descpb.ColumnID

	GetName() string

	ColName() tree.Name

	HasType() bool

	GetType() *types.T

	IsNullable() bool

	HasDefault() bool

	GetDefaultExpr() string

	HasOnUpdate() bool

	GetOnUpdateExpr() string

	IsComputed() bool

	GetComputeExpr() string

	IsHidden() bool

	IsInaccessible() bool

	IsExpressionIndexColumn() bool

	NumUsesSequences() int

	GetUsesSequenceID(usesSequenceOrdinal int) descpb.ID

	NumOwnsSequences() int

	GetOwnsSequenceID(ownsSequenceOrdinal int) descpb.ID

	IsVirtual() bool

	CheckCanBeInboundFKRef() error

	CheckCanBeOutboundFKRef() error

	GetPGAttributeNum() uint32

	IsSystemColumn() bool

	IsGeneratedAsIdentity() bool

	IsGeneratedAlwaysAsIdentity() bool

	IsGeneratedByDefaultAsIdentity() bool

	GetGeneratedAsIdentityType() catpb.GeneratedAsIdentityType

	HasGeneratedAsIdentitySequenceOption() bool

	GetGeneratedAsIdentitySequenceOption() string
}

type ConstraintToUpdate interface {
	TableElementMaybeMutation

	ConstraintToUpdateDesc() *descpb.ConstraintToUpdate

	GetName() string

	IsCheck() bool

	IsForeignKey() bool

	IsNotNull() bool

	IsUniqueWithoutIndex() bool

	Check() descpb.TableDescriptor_CheckConstraint

	ForeignKey() descpb.ForeignKeyConstraint

	NotNullColumnID() descpb.ColumnID

	UniqueWithoutIndex() descpb.UniqueWithoutIndexConstraint

	GetConstraintID() descpb.ConstraintID
}

type PrimaryKeySwap interface {
	TableElementMaybeMutation

	PrimaryKeySwapDesc() *descpb.PrimaryKeySwap

	NumOldIndexes() int

	ForEachOldIndexIDs(fn func(id descpb.IndexID) error) error

	NumNewIndexes() int

	ForEachNewIndexIDs(fn func(id descpb.IndexID) error) error

	HasLocalityConfig() bool

	LocalityConfigSwap() descpb.PrimaryKeySwap_LocalityConfigSwap
}

type ComputedColumnSwap interface {
	TableElementMaybeMutation

	ComputedColumnSwapDesc() *descpb.ComputedColumnSwap
}

type MaterializedViewRefresh interface {
	TableElementMaybeMutation

	MaterializedViewRefreshDesc() *descpb.MaterializedViewRefresh

	ShouldBackfill() bool

	AsOf() hlc.Timestamp

	ForEachIndexID(func(id descpb.IndexID) error) error

	TableWithNewIndexes(tbl TableDescriptor) TableDescriptor
}

type ModifyRowLevelTTL interface {
	TableElementMaybeMutation

	RowLevelTTL() *catpb.RowLevelTTL
}

type Partitioning interface {
	PartitioningDesc() *catpb.PartitioningDescriptor

	DeepCopy() Partitioning

	FindPartitionByName(name string) Partitioning

	ForEachPartitionName(fn func(name string) error) error

	ForEachList(fn func(name string, values [][]byte, subPartitioning Partitioning) error) error

	ForEachRange(fn func(name string, from, to []byte) error) error

	NumColumns() int

	NumImplicitColumns() int

	NumLists() int

	NumRanges() int
}

func isIndexInSearchSet(desc TableDescriptor, opts IndexOpts, idx Index) bool {
	__antithesis_instrumentation__.Notify(268485)
	if !opts.NonPhysicalPrimaryIndex && func() bool {
		__antithesis_instrumentation__.Notify(268489)
		return idx.Primary() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(268490)
		return !desc.IsPhysicalTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(268491)
		return false
	} else {
		__antithesis_instrumentation__.Notify(268492)
	}
	__antithesis_instrumentation__.Notify(268486)
	if !opts.AddMutations && func() bool {
		__antithesis_instrumentation__.Notify(268493)
		return idx.Adding() == true
	}() == true {
		__antithesis_instrumentation__.Notify(268494)
		return false
	} else {
		__antithesis_instrumentation__.Notify(268495)
	}
	__antithesis_instrumentation__.Notify(268487)
	if !opts.DropMutations && func() bool {
		__antithesis_instrumentation__.Notify(268496)
		return idx.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(268497)
		return false
	} else {
		__antithesis_instrumentation__.Notify(268498)
	}
	__antithesis_instrumentation__.Notify(268488)
	return true
}

func ForEachIndex(desc TableDescriptor, opts IndexOpts, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268499)
	for _, idx := range desc.AllIndexes() {
		__antithesis_instrumentation__.Notify(268501)
		if !isIndexInSearchSet(desc, opts, idx) {
			__antithesis_instrumentation__.Notify(268503)
			continue
		} else {
			__antithesis_instrumentation__.Notify(268504)
		}
		__antithesis_instrumentation__.Notify(268502)
		if err := f(idx); err != nil {
			__antithesis_instrumentation__.Notify(268505)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268507)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268508)
			}
			__antithesis_instrumentation__.Notify(268506)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268509)
		}
	}
	__antithesis_instrumentation__.Notify(268500)
	return nil
}

func forEachIndex(slice []Index, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268510)
	for _, idx := range slice {
		__antithesis_instrumentation__.Notify(268512)
		if err := f(idx); err != nil {
			__antithesis_instrumentation__.Notify(268513)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268515)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268516)
			}
			__antithesis_instrumentation__.Notify(268514)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268517)
		}
	}
	__antithesis_instrumentation__.Notify(268511)
	return nil
}

func ForEachActiveIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268518)
	return forEachIndex(desc.ActiveIndexes(), f)
}

func ForEachNonDropIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268519)
	return forEachIndex(desc.NonDropIndexes(), f)
}

func ForEachPartialIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268520)
	return forEachIndex(desc.PartialIndexes(), f)
}

func ForEachNonPrimaryIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268521)
	return forEachIndex(desc.NonPrimaryIndexes(), f)
}

func ForEachPublicNonPrimaryIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268522)
	return forEachIndex(desc.PublicNonPrimaryIndexes(), f)
}

func ForEachWritableNonPrimaryIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268523)
	return forEachIndex(desc.WritableNonPrimaryIndexes(), f)
}

func ForEachDeletableNonPrimaryIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268524)
	return forEachIndex(desc.DeletableNonPrimaryIndexes(), f)
}

func ForEachDeleteOnlyNonPrimaryIndex(desc TableDescriptor, f func(idx Index) error) error {
	__antithesis_instrumentation__.Notify(268525)
	return forEachIndex(desc.DeleteOnlyNonPrimaryIndexes(), f)
}

func FindIndex(desc TableDescriptor, opts IndexOpts, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268526)
	for _, idx := range desc.AllIndexes() {
		__antithesis_instrumentation__.Notify(268528)
		if !isIndexInSearchSet(desc, opts, idx) {
			__antithesis_instrumentation__.Notify(268530)
			continue
		} else {
			__antithesis_instrumentation__.Notify(268531)
		}
		__antithesis_instrumentation__.Notify(268529)
		if test(idx) {
			__antithesis_instrumentation__.Notify(268532)
			return idx
		} else {
			__antithesis_instrumentation__.Notify(268533)
		}
	}
	__antithesis_instrumentation__.Notify(268527)
	return nil
}

func findIndex(slice []Index, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268534)
	for _, idx := range slice {
		__antithesis_instrumentation__.Notify(268536)
		if test(idx) {
			__antithesis_instrumentation__.Notify(268537)
			return idx
		} else {
			__antithesis_instrumentation__.Notify(268538)
		}
	}
	__antithesis_instrumentation__.Notify(268535)
	return nil
}

func FindActiveIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268539)
	return findIndex(desc.ActiveIndexes(), test)
}

func FindNonDropIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268540)
	return findIndex(desc.NonDropIndexes(), test)
}

func FindPartialIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268541)
	return findIndex(desc.PartialIndexes(), test)
}

func FindPublicNonPrimaryIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268542)
	return findIndex(desc.PublicNonPrimaryIndexes(), test)
}

func FindWritableNonPrimaryIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268543)
	return findIndex(desc.WritableNonPrimaryIndexes(), test)
}

func FindDeletableNonPrimaryIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268544)
	return findIndex(desc.DeletableNonPrimaryIndexes(), test)
}

func FindNonPrimaryIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268545)
	return findIndex(desc.NonPrimaryIndexes(), test)
}

func FindDeleteOnlyNonPrimaryIndex(desc TableDescriptor, test func(idx Index) bool) Index {
	__antithesis_instrumentation__.Notify(268546)
	return findIndex(desc.DeleteOnlyNonPrimaryIndexes(), test)
}

func FindCorrespondingTemporaryIndexByID(desc TableDescriptor, id descpb.IndexID) Index {
	__antithesis_instrumentation__.Notify(268547)
	mutations := desc.AllMutations()
	var ord int
	for _, m := range mutations {
		__antithesis_instrumentation__.Notify(268551)
		idx := m.AsIndex()
		if idx != nil && func() bool {
			__antithesis_instrumentation__.Notify(268552)
			return idx.IndexDesc().ID == id == true
		}() == true {
			__antithesis_instrumentation__.Notify(268553)

			ord = m.MutationOrdinal() + 1
		} else {
			__antithesis_instrumentation__.Notify(268554)
		}
	}
	__antithesis_instrumentation__.Notify(268548)

	if ord == 0 {
		__antithesis_instrumentation__.Notify(268555)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(268556)
	}
	__antithesis_instrumentation__.Notify(268549)

	if len(mutations) >= ord+1 {
		__antithesis_instrumentation__.Notify(268557)
		candidateMutation := mutations[ord]
		if idx := candidateMutation.AsIndex(); idx != nil {
			__antithesis_instrumentation__.Notify(268558)
			if idx.IsTemporaryIndexForBackfill() {
				__antithesis_instrumentation__.Notify(268559)
				return idx
			} else {
				__antithesis_instrumentation__.Notify(268560)
			}
		} else {
			__antithesis_instrumentation__.Notify(268561)
		}
	} else {
		__antithesis_instrumentation__.Notify(268562)
	}
	__antithesis_instrumentation__.Notify(268550)
	return nil
}

func IsCorrespondingTemporaryIndex(idx Index, otherIdx Index) bool {
	__antithesis_instrumentation__.Notify(268563)
	return idx.IsTemporaryIndexForBackfill() && func() bool {
		__antithesis_instrumentation__.Notify(268564)
		return idx.Ordinal() == otherIdx.Ordinal()+1 == true
	}() == true
}

func UserDefinedTypeColsHaveSameVersion(desc TableDescriptor, otherDesc TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(268565)
	otherCols := otherDesc.UserDefinedTypeColumns()
	for i, thisCol := range desc.UserDefinedTypeColumns() {
		__antithesis_instrumentation__.Notify(268567)
		this, other := thisCol.GetType(), otherCols[i].GetType()
		if this.TypeMeta.Version != other.TypeMeta.Version {
			__antithesis_instrumentation__.Notify(268568)
			return false
		} else {
			__antithesis_instrumentation__.Notify(268569)
		}
	}
	__antithesis_instrumentation__.Notify(268566)
	return true
}

func ColumnIDToOrdinalMap(columns []Column) TableColMap {
	__antithesis_instrumentation__.Notify(268570)
	var m TableColMap
	for _, col := range columns {
		__antithesis_instrumentation__.Notify(268572)
		m.Set(col.GetID(), col.Ordinal())
	}
	__antithesis_instrumentation__.Notify(268571)
	return m
}

func ColumnTypes(columns []Column) []*types.T {
	__antithesis_instrumentation__.Notify(268573)
	return ColumnTypesWithInvertedCol(columns, nil)
}

func ColumnTypesWithInvertedCol(columns []Column, invertedCol Column) []*types.T {
	__antithesis_instrumentation__.Notify(268574)
	t := make([]*types.T, len(columns))
	for i, col := range columns {
		__antithesis_instrumentation__.Notify(268576)
		t[i] = col.GetType()
		if invertedCol != nil && func() bool {
			__antithesis_instrumentation__.Notify(268577)
			return col.GetID() == invertedCol.GetID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(268578)
			t[i] = invertedCol.GetType()
		} else {
			__antithesis_instrumentation__.Notify(268579)
		}
	}
	__antithesis_instrumentation__.Notify(268575)
	return t
}

func ColumnNeedsBackfill(col Column) bool {
	__antithesis_instrumentation__.Notify(268580)
	if col.IsVirtual() {
		__antithesis_instrumentation__.Notify(268584)

		return false
	} else {
		__antithesis_instrumentation__.Notify(268585)
	}
	__antithesis_instrumentation__.Notify(268581)
	if col.Dropped() {
		__antithesis_instrumentation__.Notify(268586)

		return true
	} else {
		__antithesis_instrumentation__.Notify(268587)
	}
	__antithesis_instrumentation__.Notify(268582)

	if col.ColumnDesc().HasNullDefault() {
		__antithesis_instrumentation__.Notify(268588)
		return false
	} else {
		__antithesis_instrumentation__.Notify(268589)
	}
	__antithesis_instrumentation__.Notify(268583)
	return col.HasDefault() || func() bool {
		__antithesis_instrumentation__.Notify(268590)
		return !col.IsNullable() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(268591)
		return col.IsComputed() == true
	}() == true
}
