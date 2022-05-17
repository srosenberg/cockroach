package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/errors"
)

func (m *visitor) MakeAddedIndexDeleteOnly(
	ctx context.Context, op scop.MakeAddedIndexDeleteOnly,
) error {
	__antithesis_instrumentation__.Notify(582002)
	tbl, err := m.checkOutTable(ctx, op.Index.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582011)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582012)
	}
	__antithesis_instrumentation__.Notify(582003)

	if op.Index.IndexID >= tbl.NextIndexID {
		__antithesis_instrumentation__.Notify(582013)
		tbl.NextIndexID = op.Index.IndexID + 1
	} else {
		__antithesis_instrumentation__.Notify(582014)
	}
	__antithesis_instrumentation__.Notify(582004)

	colNames, err := columnNamesFromIDs(tbl, op.Index.KeyColumnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(582015)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582016)
	}
	__antithesis_instrumentation__.Notify(582005)
	storeColNames, err := columnNamesFromIDs(tbl, op.Index.StoringColumnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(582017)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582018)
	}
	__antithesis_instrumentation__.Notify(582006)
	colDirs := make([]descpb.IndexDescriptor_Direction, len(op.Index.KeyColumnIDs))
	for i, dir := range op.Index.KeyColumnDirections {
		__antithesis_instrumentation__.Notify(582019)
		if dir == scpb.Index_DESC {
			__antithesis_instrumentation__.Notify(582020)
			colDirs[i] = descpb.IndexDescriptor_DESC
		} else {
			__antithesis_instrumentation__.Notify(582021)
		}
	}
	__antithesis_instrumentation__.Notify(582007)

	indexType := descpb.IndexDescriptor_FORWARD
	if op.Index.IsInverted {
		__antithesis_instrumentation__.Notify(582022)
		indexType = descpb.IndexDescriptor_INVERTED
	} else {
		__antithesis_instrumentation__.Notify(582023)
	}
	__antithesis_instrumentation__.Notify(582008)

	encodingType := descpb.PrimaryIndexEncoding
	indexVersion := descpb.LatestIndexDescriptorVersion
	if op.IsSecondaryIndex {
		__antithesis_instrumentation__.Notify(582024)
		encodingType = descpb.SecondaryIndexEncoding
	} else {
		__antithesis_instrumentation__.Notify(582025)
	}
	__antithesis_instrumentation__.Notify(582009)

	idx := &descpb.IndexDescriptor{
		ID:                          op.Index.IndexID,
		Name:                        tabledesc.IndexNamePlaceholder(op.Index.IndexID),
		Unique:                      op.Index.IsUnique,
		Version:                     indexVersion,
		KeyColumnNames:              colNames,
		KeyColumnIDs:                op.Index.KeyColumnIDs,
		StoreColumnIDs:              op.Index.StoringColumnIDs,
		StoreColumnNames:            storeColNames,
		KeyColumnDirections:         colDirs,
		Type:                        indexType,
		KeySuffixColumnIDs:          op.Index.KeySuffixColumnIDs,
		CompositeColumnIDs:          op.Index.CompositeColumnIDs,
		CreatedExplicitly:           true,
		EncodingType:                encodingType,
		ConstraintID:                tbl.GetNextConstraintID(),
		UseDeletePreservingEncoding: op.IsDeletePreserving,
	}
	if op.Index.Sharding != nil {
		__antithesis_instrumentation__.Notify(582026)
		idx.Sharded = *op.Index.Sharding
	} else {
		__antithesis_instrumentation__.Notify(582027)
	}
	__antithesis_instrumentation__.Notify(582010)
	tbl.NextConstraintID++
	return enqueueAddIndexMutation(tbl, idx)
}

func (m *visitor) SetAddedIndexPartialPredicate(
	ctx context.Context, op scop.SetAddedIndexPartialPredicate,
) error {
	__antithesis_instrumentation__.Notify(582028)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582031)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582032)
	}
	__antithesis_instrumentation__.Notify(582029)
	mut, err := FindMutation(tbl, MakeIndexIDMutationSelector(op.IndexID))
	if err != nil {
		__antithesis_instrumentation__.Notify(582033)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582034)
	}
	__antithesis_instrumentation__.Notify(582030)
	idx := mut.AsIndex().IndexDesc()
	idx.Predicate = string(op.Expr)
	return nil
}

func (m *visitor) MakeAddedIndexDeleteAndWriteOnly(
	ctx context.Context, op scop.MakeAddedIndexDeleteAndWriteOnly,
) error {
	__antithesis_instrumentation__.Notify(582035)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582037)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582038)
	}
	__antithesis_instrumentation__.Notify(582036)
	return mutationStateChange(
		tbl,
		MakeIndexIDMutationSelector(op.IndexID),
		descpb.DescriptorMutation_DELETE_ONLY,
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
	)
}

func (m *visitor) MakeAddedPrimaryIndexPublic(
	ctx context.Context, op scop.MakeAddedPrimaryIndexPublic,
) error {
	__antithesis_instrumentation__.Notify(582039)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582043)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582044)
	}
	__antithesis_instrumentation__.Notify(582040)
	index, err := tbl.FindIndexWithID(op.IndexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582045)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582046)
	}
	__antithesis_instrumentation__.Notify(582041)
	indexDesc := index.IndexDescDeepCopy()
	if _, err := removeMutation(
		tbl,
		MakeIndexIDMutationSelector(op.IndexID),
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
	); err != nil {
		__antithesis_instrumentation__.Notify(582047)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582048)
	}
	__antithesis_instrumentation__.Notify(582042)
	tbl.PrimaryIndex = indexDesc
	return nil
}

func (m *visitor) MakeAddedSecondaryIndexPublic(
	ctx context.Context, op scop.MakeAddedSecondaryIndexPublic,
) error {
	__antithesis_instrumentation__.Notify(582049)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582053)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582054)
	}
	__antithesis_instrumentation__.Notify(582050)

	for idx, idxMutation := range tbl.GetMutations() {
		__antithesis_instrumentation__.Notify(582055)
		if idxMutation.GetIndex() != nil && func() bool {
			__antithesis_instrumentation__.Notify(582056)
			return idxMutation.GetIndex().ID == op.IndexID == true
		}() == true {
			__antithesis_instrumentation__.Notify(582057)
			err := tbl.MakeMutationComplete(idxMutation)
			if err != nil {
				__antithesis_instrumentation__.Notify(582059)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582060)
			}
			__antithesis_instrumentation__.Notify(582058)
			tbl.Mutations = append(tbl.Mutations[:idx], tbl.Mutations[idx+1:]...)
			break
		} else {
			__antithesis_instrumentation__.Notify(582061)
		}
	}
	__antithesis_instrumentation__.Notify(582051)
	if len(tbl.Mutations) == 0 {
		__antithesis_instrumentation__.Notify(582062)
		tbl.Mutations = nil
	} else {
		__antithesis_instrumentation__.Notify(582063)
	}
	__antithesis_instrumentation__.Notify(582052)
	return nil
}

func (m *visitor) MakeDroppedPrimaryIndexDeleteAndWriteOnly(
	ctx context.Context, op scop.MakeDroppedPrimaryIndexDeleteAndWriteOnly,
) error {
	__antithesis_instrumentation__.Notify(582064)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582067)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582068)
	}
	__antithesis_instrumentation__.Notify(582065)
	if tbl.GetPrimaryIndexID() != op.IndexID {
		__antithesis_instrumentation__.Notify(582069)
		return errors.AssertionFailedf("index being dropped (%d) does not match existing primary index (%d).", op.IndexID, tbl.PrimaryIndex.ID)
	} else {
		__antithesis_instrumentation__.Notify(582070)
	}
	__antithesis_instrumentation__.Notify(582066)
	desc := tbl.GetPrimaryIndex().IndexDescDeepCopy()
	return enqueueDropIndexMutation(tbl, &desc)
}

func (m *visitor) MakeDroppedNonPrimaryIndexDeleteAndWriteOnly(
	ctx context.Context, op scop.MakeDroppedNonPrimaryIndexDeleteAndWriteOnly,
) error {
	__antithesis_instrumentation__.Notify(582071)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582074)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582075)
	}
	__antithesis_instrumentation__.Notify(582072)
	for i, idx := range tbl.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(582076)
		if idx.GetID() == op.IndexID {
			__antithesis_instrumentation__.Notify(582077)
			desc := idx.IndexDescDeepCopy()
			tbl.Indexes = append(tbl.Indexes[:i], tbl.Indexes[i+1:]...)
			return enqueueDropIndexMutation(tbl, &desc)
		} else {
			__antithesis_instrumentation__.Notify(582078)
		}
	}
	__antithesis_instrumentation__.Notify(582073)
	return errors.AssertionFailedf("failed to find secondary index %d in descriptor %v", op.IndexID, tbl)
}

func (m *visitor) MakeDroppedIndexDeleteOnly(
	ctx context.Context, op scop.MakeDroppedIndexDeleteOnly,
) error {
	__antithesis_instrumentation__.Notify(582079)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582081)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582082)
	}
	__antithesis_instrumentation__.Notify(582080)
	return mutationStateChange(
		tbl,
		MakeIndexIDMutationSelector(op.IndexID),
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		descpb.DescriptorMutation_DELETE_ONLY,
	)
}

func (m *visitor) RemoveDroppedIndexPartialPredicate(
	ctx context.Context, op scop.RemoveDroppedIndexPartialPredicate,
) error {
	__antithesis_instrumentation__.Notify(582083)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582087)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582088)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582089)
	}
	__antithesis_instrumentation__.Notify(582084)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582090)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582091)
	}
	__antithesis_instrumentation__.Notify(582085)
	mut, err := FindMutation(tbl, MakeIndexIDMutationSelector(op.IndexID))
	if err != nil {
		__antithesis_instrumentation__.Notify(582092)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582093)
	}
	__antithesis_instrumentation__.Notify(582086)
	idx := mut.AsIndex().IndexDesc()
	idx.Predicate = ""
	return nil
}

func (m *visitor) MakeIndexAbsent(ctx context.Context, op scop.MakeIndexAbsent) error {
	__antithesis_instrumentation__.Notify(582094)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582096)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582097)
	}
	__antithesis_instrumentation__.Notify(582095)
	_, err = removeMutation(
		tbl,
		MakeIndexIDMutationSelector(op.IndexID),
		descpb.DescriptorMutation_DELETE_ONLY,
	)
	return err
}

func (m *visitor) AddIndexPartitionInfo(ctx context.Context, op scop.AddIndexPartitionInfo) error {
	__antithesis_instrumentation__.Notify(582098)
	tbl, err := m.checkOutTable(ctx, op.Partitioning.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582101)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582102)
	}
	__antithesis_instrumentation__.Notify(582099)
	index, err := tbl.FindIndexWithID(op.Partitioning.IndexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582103)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582104)
	}
	__antithesis_instrumentation__.Notify(582100)
	index.IndexDesc().Partitioning = op.Partitioning.PartitioningDescriptor
	return nil
}

func (m *visitor) SetIndexName(ctx context.Context, op scop.SetIndexName) error {
	__antithesis_instrumentation__.Notify(582105)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582108)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582109)
	}
	__antithesis_instrumentation__.Notify(582106)
	index, err := tbl.FindIndexWithID(op.IndexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582110)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582111)
	}
	__antithesis_instrumentation__.Notify(582107)
	index.IndexDesc().Name = op.Name
	return nil
}
