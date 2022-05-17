// Package tabledesc provides concrete implementations of catalog.TableDesc.
package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

var _ catalog.TableDescriptor = (*immutable)(nil)
var _ catalog.TableDescriptor = (*Mutable)(nil)
var _ catalog.MutableDescriptor = (*Mutable)(nil)
var _ catalog.TableDescriptor = (*wrapper)(nil)

const ConstraintIDsAddedToTableDescsVersion = clusterversion.RemoveIncompatibleDatabasePrivileges

type wrapper struct {
	descpb.TableDescriptor

	mutationCache *mutationCache
	indexCache    *indexCache
	columnCache   *columnCache

	changes catalog.PostDeserializationChanges
}

func (*wrapper) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(270541)
	return false
}

func (desc *wrapper) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(270542)
	return desc.changes
}

func (desc *wrapper) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(270543)
	return (desc.DeclarativeSchemaChangerState != nil && func() bool {
		__antithesis_instrumentation__.Notify(270544)
		return desc.DeclarativeSchemaChangerState.JobID != catpb.InvalidJobID == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(270545)
		return len(desc.Mutations) > 0 == true
	}() == true
}

func (desc *wrapper) ActiveChecks() []descpb.TableDescriptor_CheckConstraint {
	__antithesis_instrumentation__.Notify(270546)
	checks := make([]descpb.TableDescriptor_CheckConstraint, len(desc.Checks))
	for i, c := range desc.Checks {
		__antithesis_instrumentation__.Notify(270548)
		checks[i] = *c
	}
	__antithesis_instrumentation__.Notify(270547)
	return checks
}

type immutable struct {
	wrapper

	allChecks []descpb.TableDescriptor_CheckConstraint

	isUncommittedVersion bool
}

func (desc *immutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(270549)
	return desc.isUncommittedVersion
}

func (desc *wrapper) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(270550)
	return &descpb.Descriptor{
		Union: &descpb.Descriptor_Table{Table: &desc.TableDescriptor},
	}
}

func (desc *wrapper) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(270551)
	return int64(desc.Size())
}

func (desc *wrapper) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(270552)
	return newBuilder(desc.TableDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *wrapper) GetPrimaryIndexID() descpb.IndexID {
	__antithesis_instrumentation__.Notify(270553)
	return desc.PrimaryIndex.ID
}

func (desc *wrapper) IsTemporary() bool {
	__antithesis_instrumentation__.Notify(270554)
	return desc.GetTemporary()
}

func (desc *Mutable) ImmutableCopy() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(270555)
	return desc.NewBuilder().BuildImmutable()
}

func (desc *Mutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(270556)
	return newBuilder(desc.TableDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *Mutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(270557)
	clusterVersion := desc.ClusterVersion()
	return desc.IsNew() || func() bool {
		__antithesis_instrumentation__.Notify(270558)
		return desc.GetVersion() != clusterVersion.GetVersion() == true
	}() == true
}

func (desc *Mutable) SetDrainingNames(names []descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(270559)
	desc.DrainingNames = names
}

func (desc *Mutable) RemovePublicNonPrimaryIndex(indexOrdinal int) {
	__antithesis_instrumentation__.Notify(270560)
	desc.Indexes = append(desc.Indexes[:indexOrdinal-1], desc.Indexes[indexOrdinal:]...)
}

func (desc *Mutable) SetPublicNonPrimaryIndexes(indexes []descpb.IndexDescriptor) {
	__antithesis_instrumentation__.Notify(270561)
	desc.Indexes = append(desc.Indexes[:0], indexes...)
}

func (desc *Mutable) AddPublicNonPrimaryIndex(index descpb.IndexDescriptor) {
	__antithesis_instrumentation__.Notify(270562)
	desc.Indexes = append(desc.Indexes, index)
}

func (desc *Mutable) SetPrimaryIndex(index descpb.IndexDescriptor) {
	__antithesis_instrumentation__.Notify(270563)
	desc.PrimaryIndex = index
}

func (desc *Mutable) SetPublicNonPrimaryIndex(indexOrdinal int, index descpb.IndexDescriptor) {
	__antithesis_instrumentation__.Notify(270564)
	desc.Indexes[indexOrdinal-1] = index
}

func UpdateIndexPartitioning(
	idx *descpb.IndexDescriptor,
	isIndexPrimary bool,
	newImplicitCols []catalog.Column,
	newPartitioning catpb.PartitioningDescriptor,
) bool {
	__antithesis_instrumentation__.Notify(270565)
	oldNumImplicitCols := int(idx.Partitioning.NumImplicitColumns)
	isNoOp := oldNumImplicitCols == len(newImplicitCols) && func() bool {
		__antithesis_instrumentation__.Notify(270571)
		return idx.Partitioning.Equal(newPartitioning) == true
	}() == true
	numCols := len(idx.KeyColumnIDs)
	newCap := numCols + len(newImplicitCols) - oldNumImplicitCols
	newColumnIDs := make([]descpb.ColumnID, len(newImplicitCols), newCap)
	newColumnNames := make([]string, len(newImplicitCols), newCap)
	newColumnDirections := make([]descpb.IndexDescriptor_Direction, len(newImplicitCols), newCap)
	for i, col := range newImplicitCols {
		__antithesis_instrumentation__.Notify(270572)
		newColumnIDs[i] = col.GetID()
		newColumnNames[i] = col.GetName()
		newColumnDirections[i] = descpb.IndexDescriptor_ASC
		if isNoOp && func() bool {
			__antithesis_instrumentation__.Notify(270573)
			return (idx.KeyColumnIDs[i] != newColumnIDs[i] || func() bool {
				__antithesis_instrumentation__.Notify(270574)
				return idx.KeyColumnNames[i] != newColumnNames[i] == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(270575)
				return idx.KeyColumnDirections[i] != newColumnDirections[i] == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(270576)
			isNoOp = false
		} else {
			__antithesis_instrumentation__.Notify(270577)
		}
	}
	__antithesis_instrumentation__.Notify(270566)
	if isNoOp {
		__antithesis_instrumentation__.Notify(270578)
		return false
	} else {
		__antithesis_instrumentation__.Notify(270579)
	}
	__antithesis_instrumentation__.Notify(270567)
	idx.KeyColumnIDs = append(newColumnIDs, idx.KeyColumnIDs[oldNumImplicitCols:]...)
	idx.KeyColumnNames = append(newColumnNames, idx.KeyColumnNames[oldNumImplicitCols:]...)
	idx.KeyColumnDirections = append(newColumnDirections, idx.KeyColumnDirections[oldNumImplicitCols:]...)
	idx.Partitioning = newPartitioning
	if !isIndexPrimary {
		__antithesis_instrumentation__.Notify(270580)
		return true
	} else {
		__antithesis_instrumentation__.Notify(270581)
	}
	__antithesis_instrumentation__.Notify(270568)

	newStoreColumnIDs := make([]descpb.ColumnID, 0, len(idx.StoreColumnIDs))
	newStoreColumnNames := make([]string, 0, len(idx.StoreColumnNames))
	for i := range idx.StoreColumnIDs {
		__antithesis_instrumentation__.Notify(270582)
		id := idx.StoreColumnIDs[i]
		name := idx.StoreColumnNames[i]
		found := false
		for _, newColumnName := range newColumnNames {
			__antithesis_instrumentation__.Notify(270584)
			if newColumnName == name {
				__antithesis_instrumentation__.Notify(270585)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(270586)
			}
		}
		__antithesis_instrumentation__.Notify(270583)
		if !found {
			__antithesis_instrumentation__.Notify(270587)
			newStoreColumnIDs = append(newStoreColumnIDs, id)
			newStoreColumnNames = append(newStoreColumnNames, name)
		} else {
			__antithesis_instrumentation__.Notify(270588)
		}
	}
	__antithesis_instrumentation__.Notify(270569)
	idx.StoreColumnIDs = newStoreColumnIDs
	idx.StoreColumnNames = newStoreColumnNames
	if len(idx.StoreColumnNames) == 0 {
		__antithesis_instrumentation__.Notify(270589)
		idx.StoreColumnIDs = nil
		idx.StoreColumnNames = nil
	} else {
		__antithesis_instrumentation__.Notify(270590)
	}
	__antithesis_instrumentation__.Notify(270570)
	return true
}

func (desc *wrapper) GetPrimaryIndex() catalog.Index {
	__antithesis_instrumentation__.Notify(270591)
	return desc.getExistingOrNewIndexCache().primary
}

func (desc *wrapper) getExistingOrNewIndexCache() *indexCache {
	__antithesis_instrumentation__.Notify(270592)
	if desc.indexCache != nil {
		__antithesis_instrumentation__.Notify(270594)
		return desc.indexCache
	} else {
		__antithesis_instrumentation__.Notify(270595)
	}
	__antithesis_instrumentation__.Notify(270593)
	return newIndexCache(desc.TableDesc(), desc.getExistingOrNewMutationCache())
}

func (desc *wrapper) AllIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270596)
	return desc.getExistingOrNewIndexCache().all
}

func (desc *wrapper) ActiveIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270597)
	return desc.getExistingOrNewIndexCache().active
}

func (desc *wrapper) NonDropIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270598)
	return desc.getExistingOrNewIndexCache().nonDrop
}

func (desc *wrapper) PartialIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270599)
	return desc.getExistingOrNewIndexCache().partial
}

func (desc *wrapper) NonPrimaryIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270600)
	return desc.getExistingOrNewIndexCache().nonPrimary
}

func (desc *wrapper) PublicNonPrimaryIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270601)
	return desc.getExistingOrNewIndexCache().publicNonPrimary
}

func (desc *wrapper) WritableNonPrimaryIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270602)
	return desc.getExistingOrNewIndexCache().writableNonPrimary
}

func (desc *wrapper) DeletableNonPrimaryIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270603)
	return desc.getExistingOrNewIndexCache().deletableNonPrimary
}

func (desc *wrapper) DeleteOnlyNonPrimaryIndexes() []catalog.Index {
	__antithesis_instrumentation__.Notify(270604)
	return desc.getExistingOrNewIndexCache().deleteOnlyNonPrimary
}

func (desc *wrapper) FindIndexWithID(id descpb.IndexID) (catalog.Index, error) {
	__antithesis_instrumentation__.Notify(270605)
	if idx := catalog.FindIndex(desc, catalog.IndexOpts{
		NonPhysicalPrimaryIndex: true,
		DropMutations:           true,
		AddMutations:            true,
	}, func(idx catalog.Index) bool {
		__antithesis_instrumentation__.Notify(270607)
		return idx.GetID() == id
	}); idx != nil {
		__antithesis_instrumentation__.Notify(270608)
		return idx, nil
	} else {
		__antithesis_instrumentation__.Notify(270609)
	}
	__antithesis_instrumentation__.Notify(270606)
	return nil, errors.Errorf("index-id \"%d\" does not exist", id)
}

func (desc *wrapper) FindIndexWithName(name string) (catalog.Index, error) {
	__antithesis_instrumentation__.Notify(270610)
	if idx := catalog.FindIndex(desc, catalog.IndexOpts{
		NonPhysicalPrimaryIndex: false,
		DropMutations:           true,
		AddMutations:            true,
	}, func(idx catalog.Index) bool {
		__antithesis_instrumentation__.Notify(270612)
		return idx.GetName() == name
	}); idx != nil {
		__antithesis_instrumentation__.Notify(270613)
		return idx, nil
	} else {
		__antithesis_instrumentation__.Notify(270614)
	}
	__antithesis_instrumentation__.Notify(270611)
	return nil, errors.Errorf("index %q does not exist", name)
}

func (desc *wrapper) getExistingOrNewColumnCache() *columnCache {
	__antithesis_instrumentation__.Notify(270615)
	if desc.columnCache != nil {
		__antithesis_instrumentation__.Notify(270617)
		return desc.columnCache
	} else {
		__antithesis_instrumentation__.Notify(270618)
	}
	__antithesis_instrumentation__.Notify(270616)
	return newColumnCache(desc.TableDesc(), desc.getExistingOrNewMutationCache())
}

func (desc *wrapper) AllColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270619)
	return desc.getExistingOrNewColumnCache().all
}

func (desc *wrapper) PublicColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270620)
	return desc.getExistingOrNewColumnCache().public
}

func (desc *wrapper) WritableColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270621)
	return desc.getExistingOrNewColumnCache().writable
}

func (desc *wrapper) DeletableColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270622)
	return desc.getExistingOrNewColumnCache().deletable
}

func (desc *wrapper) NonDropColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270623)
	return desc.getExistingOrNewColumnCache().nonDrop
}

func (desc *wrapper) VisibleColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270624)
	return desc.getExistingOrNewColumnCache().visible
}

func (desc *wrapper) AccessibleColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270625)
	return desc.getExistingOrNewColumnCache().accessible
}

func (desc *wrapper) UserDefinedTypeColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270626)
	return desc.getExistingOrNewColumnCache().withUDTs
}

func (desc *wrapper) ReadableColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270627)
	return desc.getExistingOrNewColumnCache().readable
}

func (desc *wrapper) SystemColumns() []catalog.Column {
	__antithesis_instrumentation__.Notify(270628)
	return desc.getExistingOrNewColumnCache().system
}

func (desc *wrapper) FamilyDefaultColumns() []descpb.IndexFetchSpec_FamilyDefaultColumn {
	__antithesis_instrumentation__.Notify(270629)
	return desc.getExistingOrNewColumnCache().familyDefaultColumns
}

func (desc *wrapper) PublicColumnIDs() []descpb.ColumnID {
	__antithesis_instrumentation__.Notify(270630)
	cols := desc.PublicColumns()
	res := make([]descpb.ColumnID, len(cols))
	for i, c := range cols {
		__antithesis_instrumentation__.Notify(270632)
		res[i] = c.GetID()
	}
	__antithesis_instrumentation__.Notify(270631)
	return res
}

func (desc *wrapper) IndexColumns(idx catalog.Index) []catalog.Column {
	__antithesis_instrumentation__.Notify(270633)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270635)
		return ic.all
	} else {
		__antithesis_instrumentation__.Notify(270636)
	}
	__antithesis_instrumentation__.Notify(270634)
	return nil
}

func (desc *wrapper) IndexKeyColumns(idx catalog.Index) []catalog.Column {
	__antithesis_instrumentation__.Notify(270637)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270639)
		return ic.key
	} else {
		__antithesis_instrumentation__.Notify(270640)
	}
	__antithesis_instrumentation__.Notify(270638)
	return nil
}

func (desc *wrapper) IndexKeyColumnDirections(
	idx catalog.Index,
) []descpb.IndexDescriptor_Direction {
	__antithesis_instrumentation__.Notify(270641)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270643)
		return ic.keyDirs
	} else {
		__antithesis_instrumentation__.Notify(270644)
	}
	__antithesis_instrumentation__.Notify(270642)
	return nil
}

func (desc *wrapper) IndexKeySuffixColumns(idx catalog.Index) []catalog.Column {
	__antithesis_instrumentation__.Notify(270645)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270647)
		return ic.keySuffix
	} else {
		__antithesis_instrumentation__.Notify(270648)
	}
	__antithesis_instrumentation__.Notify(270646)
	return nil
}

func (desc *wrapper) IndexFullColumns(idx catalog.Index) []catalog.Column {
	__antithesis_instrumentation__.Notify(270649)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270651)
		return ic.full
	} else {
		__antithesis_instrumentation__.Notify(270652)
	}
	__antithesis_instrumentation__.Notify(270650)
	return nil
}

func (desc *wrapper) IndexFullColumnDirections(
	idx catalog.Index,
) []descpb.IndexDescriptor_Direction {
	__antithesis_instrumentation__.Notify(270653)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270655)
		return ic.fullDirs
	} else {
		__antithesis_instrumentation__.Notify(270656)
	}
	__antithesis_instrumentation__.Notify(270654)
	return nil
}

func (desc *wrapper) IndexStoredColumns(idx catalog.Index) []catalog.Column {
	__antithesis_instrumentation__.Notify(270657)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270659)
		return ic.stored
	} else {
		__antithesis_instrumentation__.Notify(270660)
	}
	__antithesis_instrumentation__.Notify(270658)
	return nil
}

func (desc *wrapper) IndexFetchSpecKeyAndSuffixColumns(
	idx catalog.Index,
) []descpb.IndexFetchSpec_KeyColumn {
	__antithesis_instrumentation__.Notify(270661)
	if ic := desc.getExistingOrNewIndexColumnCache(idx); ic != nil {
		__antithesis_instrumentation__.Notify(270663)
		return ic.keyAndSuffix
	} else {
		__antithesis_instrumentation__.Notify(270664)
	}
	__antithesis_instrumentation__.Notify(270662)
	return nil
}

func (desc *wrapper) getExistingOrNewIndexColumnCache(idx catalog.Index) *indexColumnCache {
	__antithesis_instrumentation__.Notify(270665)
	if idx == nil {
		__antithesis_instrumentation__.Notify(270668)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(270669)
	}
	__antithesis_instrumentation__.Notify(270666)
	c := desc.getExistingOrNewColumnCache()
	if idx.Ordinal() >= len(c.index) {
		__antithesis_instrumentation__.Notify(270670)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(270671)
	}
	__antithesis_instrumentation__.Notify(270667)
	return &c.index[idx.Ordinal()]
}

func (desc *wrapper) FindColumnWithID(id descpb.ColumnID) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(270672)
	for _, col := range desc.AllColumns() {
		__antithesis_instrumentation__.Notify(270674)
		if col.GetID() == id {
			__antithesis_instrumentation__.Notify(270675)
			return col, nil
		} else {
			__antithesis_instrumentation__.Notify(270676)
		}
	}
	__antithesis_instrumentation__.Notify(270673)

	return nil, pgerror.Newf(pgcode.UndefinedColumn, "column-id \"%d\" does not exist", id)
}

func (desc *wrapper) FindColumnWithName(name tree.Name) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(270677)
	for _, col := range desc.AllColumns() {
		__antithesis_instrumentation__.Notify(270679)
		if col.ColName() == name {
			__antithesis_instrumentation__.Notify(270680)
			return col, nil
		} else {
			__antithesis_instrumentation__.Notify(270681)
		}
	}
	__antithesis_instrumentation__.Notify(270678)
	return nil, colinfo.NewUndefinedColumnError(string(name))
}

func (desc *wrapper) getExistingOrNewMutationCache() *mutationCache {
	__antithesis_instrumentation__.Notify(270682)
	if desc.mutationCache != nil {
		__antithesis_instrumentation__.Notify(270684)
		return desc.mutationCache
	} else {
		__antithesis_instrumentation__.Notify(270685)
	}
	__antithesis_instrumentation__.Notify(270683)
	return newMutationCache(desc.TableDesc())
}

func (desc *wrapper) AllMutations() []catalog.Mutation {
	__antithesis_instrumentation__.Notify(270686)
	return desc.getExistingOrNewMutationCache().all
}
