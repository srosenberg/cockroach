package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
)

var _ catalog.TableElementMaybeMutation = maybeMutation{}
var _ catalog.TableElementMaybeMutation = constraintToUpdate{}
var _ catalog.TableElementMaybeMutation = primaryKeySwap{}
var _ catalog.TableElementMaybeMutation = computedColumnSwap{}
var _ catalog.TableElementMaybeMutation = materializedViewRefresh{}
var _ catalog.Mutation = mutation{}

type maybeMutation struct {
	mutationID         descpb.MutationID
	mutationDirection  descpb.DescriptorMutation_Direction
	mutationState      descpb.DescriptorMutation_State
	mutationIsRollback bool
}

func (mm maybeMutation) IsMutation() bool {
	__antithesis_instrumentation__.Notify(268901)
	return mm.mutationState != descpb.DescriptorMutation_UNKNOWN
}

func (mm maybeMutation) IsRollback() bool {
	__antithesis_instrumentation__.Notify(268902)
	return mm.mutationIsRollback
}

func (mm maybeMutation) MutationID() descpb.MutationID {
	__antithesis_instrumentation__.Notify(268903)
	return mm.mutationID
}

func (mm maybeMutation) WriteAndDeleteOnly() bool {
	__antithesis_instrumentation__.Notify(268904)
	return mm.mutationState == descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
}

func (mm maybeMutation) DeleteOnly() bool {
	__antithesis_instrumentation__.Notify(268905)
	return mm.mutationState == descpb.DescriptorMutation_DELETE_ONLY
}

func (mm maybeMutation) Backfilling() bool {
	__antithesis_instrumentation__.Notify(268906)
	return mm.mutationState == descpb.DescriptorMutation_BACKFILLING
}

func (mm maybeMutation) Merging() bool {
	__antithesis_instrumentation__.Notify(268907)
	return mm.mutationState == descpb.DescriptorMutation_MERGING
}

func (mm maybeMutation) Adding() bool {
	__antithesis_instrumentation__.Notify(268908)
	return mm.mutationDirection == descpb.DescriptorMutation_ADD
}

func (mm maybeMutation) Dropped() bool {
	__antithesis_instrumentation__.Notify(268909)
	return mm.mutationDirection == descpb.DescriptorMutation_DROP
}

type constraintToUpdate struct {
	maybeMutation
	desc *descpb.ConstraintToUpdate
}

func (c constraintToUpdate) ConstraintToUpdateDesc() *descpb.ConstraintToUpdate {
	__antithesis_instrumentation__.Notify(268910)
	return c.desc
}

func (c constraintToUpdate) GetName() string {
	__antithesis_instrumentation__.Notify(268911)
	return c.desc.Name
}

func (c constraintToUpdate) IsCheck() bool {
	__antithesis_instrumentation__.Notify(268912)
	return c.desc.ConstraintType == descpb.ConstraintToUpdate_CHECK
}

func (c constraintToUpdate) Check() descpb.TableDescriptor_CheckConstraint {
	__antithesis_instrumentation__.Notify(268913)
	return c.desc.Check
}

func (c constraintToUpdate) IsForeignKey() bool {
	__antithesis_instrumentation__.Notify(268914)
	return c.desc.ConstraintType == descpb.ConstraintToUpdate_FOREIGN_KEY
}

func (c constraintToUpdate) ForeignKey() descpb.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(268915)
	return c.desc.ForeignKey
}

func (c constraintToUpdate) IsNotNull() bool {
	__antithesis_instrumentation__.Notify(268916)
	return c.desc.ConstraintType == descpb.ConstraintToUpdate_NOT_NULL
}

func (c constraintToUpdate) NotNullColumnID() descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268917)
	return c.desc.NotNullColumn
}

func (c constraintToUpdate) IsUniqueWithoutIndex() bool {
	__antithesis_instrumentation__.Notify(268918)
	return c.desc.ConstraintType == descpb.ConstraintToUpdate_UNIQUE_WITHOUT_INDEX
}

func (c constraintToUpdate) UniqueWithoutIndex() descpb.UniqueWithoutIndexConstraint {
	__antithesis_instrumentation__.Notify(268919)
	return c.desc.UniqueWithoutIndexConstraint
}

func (c constraintToUpdate) GetConstraintID() descpb.ConstraintID {
	__antithesis_instrumentation__.Notify(268920)
	switch c.desc.ConstraintType {
	case descpb.ConstraintToUpdate_CHECK:
		__antithesis_instrumentation__.Notify(268922)
		return c.desc.Check.ConstraintID
	case descpb.ConstraintToUpdate_FOREIGN_KEY:
		__antithesis_instrumentation__.Notify(268923)
		return c.ForeignKey().ConstraintID
	case descpb.ConstraintToUpdate_NOT_NULL:
		__antithesis_instrumentation__.Notify(268924)
		return 0
	case descpb.ConstraintToUpdate_UNIQUE_WITHOUT_INDEX:
		__antithesis_instrumentation__.Notify(268925)
		return c.UniqueWithoutIndex().ConstraintID
	default:
		__antithesis_instrumentation__.Notify(268926)
	}
	__antithesis_instrumentation__.Notify(268921)
	panic("unknown constraint type")
}

type modifyRowLevelTTL struct {
	maybeMutation
	desc *descpb.ModifyRowLevelTTL
}

func (c modifyRowLevelTTL) RowLevelTTL() *catpb.RowLevelTTL {
	__antithesis_instrumentation__.Notify(268927)
	return c.desc.RowLevelTTL
}

type primaryKeySwap struct {
	maybeMutation
	desc *descpb.PrimaryKeySwap
}

func (c primaryKeySwap) PrimaryKeySwapDesc() *descpb.PrimaryKeySwap {
	__antithesis_instrumentation__.Notify(268928)
	return c.desc
}

func (c primaryKeySwap) NumOldIndexes() int {
	__antithesis_instrumentation__.Notify(268929)
	return 1 + len(c.desc.OldIndexes)
}

func (c primaryKeySwap) ForEachOldIndexIDs(fn func(id descpb.IndexID) error) error {
	__antithesis_instrumentation__.Notify(268930)
	return c.forEachIndexIDs(c.desc.OldPrimaryIndexId, c.desc.OldIndexes, fn)
}

func (c primaryKeySwap) NumNewIndexes() int {
	__antithesis_instrumentation__.Notify(268931)
	return 1 + len(c.desc.NewIndexes)
}

func (c primaryKeySwap) ForEachNewIndexIDs(fn func(id descpb.IndexID) error) error {
	__antithesis_instrumentation__.Notify(268932)
	return c.forEachIndexIDs(c.desc.NewPrimaryIndexId, c.desc.NewIndexes, fn)
}

func (c primaryKeySwap) forEachIndexIDs(
	pkID descpb.IndexID, secIDs []descpb.IndexID, fn func(id descpb.IndexID) error,
) error {
	__antithesis_instrumentation__.Notify(268933)
	err := fn(pkID)
	if err != nil {
		__antithesis_instrumentation__.Notify(268936)
		if iterutil.Done(err) {
			__antithesis_instrumentation__.Notify(268938)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(268939)
		}
		__antithesis_instrumentation__.Notify(268937)
		return err
	} else {
		__antithesis_instrumentation__.Notify(268940)
	}
	__antithesis_instrumentation__.Notify(268934)
	for _, id := range secIDs {
		__antithesis_instrumentation__.Notify(268941)
		err = fn(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(268942)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268944)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268945)
			}
			__antithesis_instrumentation__.Notify(268943)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268946)
		}
	}
	__antithesis_instrumentation__.Notify(268935)
	return nil
}

func (c primaryKeySwap) HasLocalityConfig() bool {
	__antithesis_instrumentation__.Notify(268947)
	return c.desc.LocalityConfigSwap != nil
}

func (c primaryKeySwap) LocalityConfigSwap() descpb.PrimaryKeySwap_LocalityConfigSwap {
	__antithesis_instrumentation__.Notify(268948)
	return *c.desc.LocalityConfigSwap
}

type computedColumnSwap struct {
	maybeMutation
	desc *descpb.ComputedColumnSwap
}

func (c computedColumnSwap) ComputedColumnSwapDesc() *descpb.ComputedColumnSwap {
	__antithesis_instrumentation__.Notify(268949)
	return c.desc
}

type materializedViewRefresh struct {
	maybeMutation
	desc *descpb.MaterializedViewRefresh
}

func (c materializedViewRefresh) MaterializedViewRefreshDesc() *descpb.MaterializedViewRefresh {
	__antithesis_instrumentation__.Notify(268950)
	return c.desc
}

func (c materializedViewRefresh) ShouldBackfill() bool {
	__antithesis_instrumentation__.Notify(268951)
	return c.desc.ShouldBackfill
}

func (c materializedViewRefresh) AsOf() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(268952)
	return c.desc.AsOf
}

func (c materializedViewRefresh) ForEachIndexID(fn func(id descpb.IndexID) error) error {
	__antithesis_instrumentation__.Notify(268953)
	err := fn(c.desc.NewPrimaryIndex.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(268956)
		if iterutil.Done(err) {
			__antithesis_instrumentation__.Notify(268958)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(268959)
		}
		__antithesis_instrumentation__.Notify(268957)
		return err
	} else {
		__antithesis_instrumentation__.Notify(268960)
	}
	__antithesis_instrumentation__.Notify(268954)
	for i := range c.desc.NewIndexes {
		__antithesis_instrumentation__.Notify(268961)
		err = fn(c.desc.NewIndexes[i].ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(268962)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(268964)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(268965)
			}
			__antithesis_instrumentation__.Notify(268963)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268966)
		}
	}
	__antithesis_instrumentation__.Notify(268955)
	return nil
}

func (c materializedViewRefresh) TableWithNewIndexes(
	tbl catalog.TableDescriptor,
) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(268967)
	deepCopy := NewBuilder(tbl.TableDesc()).BuildCreatedMutableTable().TableDesc()
	deepCopy.PrimaryIndex = c.desc.NewPrimaryIndex
	deepCopy.Indexes = c.desc.NewIndexes
	return NewBuilder(deepCopy).BuildImmutableTable()
}

type mutation struct {
	maybeMutation
	column            catalog.Column
	index             catalog.Index
	constraint        catalog.ConstraintToUpdate
	pkSwap            catalog.PrimaryKeySwap
	ccSwap            catalog.ComputedColumnSwap
	mvRefresh         catalog.MaterializedViewRefresh
	modifyRowLevelTTL catalog.ModifyRowLevelTTL
	mutationOrdinal   int
}

func (m mutation) AsColumn() catalog.Column {
	__antithesis_instrumentation__.Notify(268968)
	return m.column
}

func (m mutation) AsIndex() catalog.Index {
	__antithesis_instrumentation__.Notify(268969)
	return m.index
}

func (m mutation) AsConstraint() catalog.ConstraintToUpdate {
	__antithesis_instrumentation__.Notify(268970)
	return m.constraint
}

func (m mutation) AsPrimaryKeySwap() catalog.PrimaryKeySwap {
	__antithesis_instrumentation__.Notify(268971)
	return m.pkSwap
}

func (m mutation) AsModifyRowLevelTTL() catalog.ModifyRowLevelTTL {
	__antithesis_instrumentation__.Notify(268972)
	return m.modifyRowLevelTTL
}

func (m mutation) AsComputedColumnSwap() catalog.ComputedColumnSwap {
	__antithesis_instrumentation__.Notify(268973)
	return m.ccSwap
}

func (m mutation) AsMaterializedViewRefresh() catalog.MaterializedViewRefresh {
	__antithesis_instrumentation__.Notify(268974)
	return m.mvRefresh
}

func (m mutation) MutationOrdinal() int {
	__antithesis_instrumentation__.Notify(268975)
	return m.mutationOrdinal
}

type mutationCache struct {
	all     []catalog.Mutation
	columns []catalog.Mutation
	indexes []catalog.Mutation
}

func newMutationCache(desc *descpb.TableDescriptor) *mutationCache {
	__antithesis_instrumentation__.Notify(268976)
	c := mutationCache{}
	if len(desc.Mutations) == 0 {
		__antithesis_instrumentation__.Notify(268983)
		return &c
	} else {
		__antithesis_instrumentation__.Notify(268984)
	}
	__antithesis_instrumentation__.Notify(268977)

	backingStructs := make([]mutation, len(desc.Mutations))
	var columns []column
	var indexes []index
	var constraints []constraintToUpdate
	var pkSwaps []primaryKeySwap
	var ccSwaps []computedColumnSwap
	var mvRefreshes []materializedViewRefresh
	var modifyRowLevelTTLs []modifyRowLevelTTL
	for i, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(268985)
		mm := maybeMutation{
			mutationID:         m.MutationID,
			mutationDirection:  m.Direction,
			mutationState:      m.State,
			mutationIsRollback: m.Rollback,
		}
		backingStructs[i] = mutation{
			maybeMutation:   mm,
			mutationOrdinal: i,
		}
		if pb := m.GetColumn(); pb != nil {
			__antithesis_instrumentation__.Notify(268986)
			columns = append(columns, column{
				maybeMutation: mm,
				desc:          pb,
				ordinal:       len(desc.Columns) + len(columns),
			})
			backingStructs[i].column = &columns[len(columns)-1]
		} else {
			__antithesis_instrumentation__.Notify(268987)
			if pb := m.GetIndex(); pb != nil {
				__antithesis_instrumentation__.Notify(268988)
				indexes = append(indexes, index{
					maybeMutation: mm,
					desc:          pb,
					ordinal:       1 + len(desc.Indexes) + len(indexes),
				})
				backingStructs[i].index = &indexes[len(indexes)-1]
			} else {
				__antithesis_instrumentation__.Notify(268989)
				if pb := m.GetConstraint(); pb != nil {
					__antithesis_instrumentation__.Notify(268990)
					constraints = append(constraints, constraintToUpdate{
						maybeMutation: mm,
						desc:          pb,
					})
					backingStructs[i].constraint = &constraints[len(constraints)-1]
				} else {
					__antithesis_instrumentation__.Notify(268991)
					if pb := m.GetPrimaryKeySwap(); pb != nil {
						__antithesis_instrumentation__.Notify(268992)
						pkSwaps = append(pkSwaps, primaryKeySwap{
							maybeMutation: mm,
							desc:          pb,
						})
						backingStructs[i].pkSwap = &pkSwaps[len(pkSwaps)-1]
					} else {
						__antithesis_instrumentation__.Notify(268993)
						if pb := m.GetComputedColumnSwap(); pb != nil {
							__antithesis_instrumentation__.Notify(268994)
							ccSwaps = append(ccSwaps, computedColumnSwap{
								maybeMutation: mm,
								desc:          pb,
							})
							backingStructs[i].ccSwap = &ccSwaps[len(ccSwaps)-1]
						} else {
							__antithesis_instrumentation__.Notify(268995)
							if pb := m.GetMaterializedViewRefresh(); pb != nil {
								__antithesis_instrumentation__.Notify(268996)
								mvRefreshes = append(mvRefreshes, materializedViewRefresh{
									maybeMutation: mm,
									desc:          pb,
								})
								backingStructs[i].mvRefresh = &mvRefreshes[len(mvRefreshes)-1]
							} else {
								__antithesis_instrumentation__.Notify(268997)
								if pb := m.GetModifyRowLevelTTL(); pb != nil {
									__antithesis_instrumentation__.Notify(268998)
									modifyRowLevelTTLs = append(modifyRowLevelTTLs, modifyRowLevelTTL{
										maybeMutation: mm,
										desc:          pb,
									})
									backingStructs[i].modifyRowLevelTTL = &modifyRowLevelTTLs[len(modifyRowLevelTTLs)-1]
								} else {
									__antithesis_instrumentation__.Notify(268999)
								}
							}
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(268978)

	c.all = make([]catalog.Mutation, len(backingStructs))
	for i := range backingStructs {
		__antithesis_instrumentation__.Notify(269000)
		c.all[i] = &backingStructs[i]
	}
	__antithesis_instrumentation__.Notify(268979)

	if len(columns) > 0 {
		__antithesis_instrumentation__.Notify(269001)
		c.columns = make([]catalog.Mutation, 0, len(columns))
	} else {
		__antithesis_instrumentation__.Notify(269002)
	}
	__antithesis_instrumentation__.Notify(268980)
	if len(indexes) > 0 {
		__antithesis_instrumentation__.Notify(269003)
		c.indexes = make([]catalog.Mutation, 0, len(indexes))
	} else {
		__antithesis_instrumentation__.Notify(269004)
	}
	__antithesis_instrumentation__.Notify(268981)
	for _, m := range c.all {
		__antithesis_instrumentation__.Notify(269005)
		if col := m.AsColumn(); col != nil {
			__antithesis_instrumentation__.Notify(269006)
			c.columns = append(c.columns, m)
		} else {
			__antithesis_instrumentation__.Notify(269007)
			if idx := m.AsIndex(); idx != nil {
				__antithesis_instrumentation__.Notify(269008)
				c.indexes = append(c.indexes, m)
			} else {
				__antithesis_instrumentation__.Notify(269009)
			}
		}
	}
	__antithesis_instrumentation__.Notify(268982)
	return &c
}
