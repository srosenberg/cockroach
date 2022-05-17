package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (m *visitor) CreateGcJobForTable(ctx context.Context, op scop.CreateGcJobForTable) error {
	__antithesis_instrumentation__.Notify(581829)
	desc, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581831)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581832)
	}
	__antithesis_instrumentation__.Notify(581830)
	m.s.AddNewGCJobForTable(op.StatementForDropJob, desc)
	return nil
}

func (m *visitor) CreateGcJobForDatabase(
	ctx context.Context, op scop.CreateGcJobForDatabase,
) error {
	__antithesis_instrumentation__.Notify(581833)
	desc, err := m.checkOutDatabase(ctx, op.DatabaseID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581835)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581836)
	}
	__antithesis_instrumentation__.Notify(581834)
	m.s.AddNewGCJobForDatabase(op.StatementForDropJob, desc)
	return nil
}

func (m *visitor) CreateGcJobForIndex(ctx context.Context, op scop.CreateGcJobForIndex) error {
	__antithesis_instrumentation__.Notify(581837)
	desc, err := m.s.GetDescriptor(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581841)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581842)
	}
	__antithesis_instrumentation__.Notify(581838)
	tbl, err := catalog.AsTableDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(581843)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581844)
	}
	__antithesis_instrumentation__.Notify(581839)
	idx, err := tbl.FindIndexWithID(op.IndexID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581845)
		return errors.AssertionFailedf("table %q (%d): could not find index %d", tbl.GetName(), tbl.GetID(), op.IndexID)
	} else {
		__antithesis_instrumentation__.Notify(581846)
	}
	__antithesis_instrumentation__.Notify(581840)
	m.s.AddNewGCJobForIndex(op.StatementForDropJob, tbl, idx)
	return nil
}

func (m *visitor) MarkDescriptorAsDropped(
	ctx context.Context, op scop.MarkDescriptorAsDropped,
) error {
	__antithesis_instrumentation__.Notify(581847)

	m.sd.RemoveSyntheticDescriptor(op.DescID)
	desc, err := m.s.CheckOutDescriptor(ctx, op.DescID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581850)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581851)
	}
	__antithesis_instrumentation__.Notify(581848)
	desc.SetDropped()

	if tableDesc, ok := desc.(*tabledesc.Mutable); ok && func() bool {
		__antithesis_instrumentation__.Notify(581852)
		return tableDesc.IsTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581853)
		tableDesc.DropTime = timeutil.Now().UnixNano()
	} else {
		__antithesis_instrumentation__.Notify(581854)
	}
	__antithesis_instrumentation__.Notify(581849)
	return nil
}

func (m *visitor) MarkDescriptorAsDroppedSynthetically(
	ctx context.Context, op scop.MarkDescriptorAsDroppedSynthetically,
) error {
	__antithesis_instrumentation__.Notify(581855)
	if co := m.s.MaybeCheckedOutDescriptor(op.DescID); co != nil {
		__antithesis_instrumentation__.Notify(581858)
		return errors.AssertionFailedf("cannot mark already checked-out descriptor %q (%d) as synthetic",
			co.GetName(), co.GetID())
	} else {
		__antithesis_instrumentation__.Notify(581859)
	}
	__antithesis_instrumentation__.Notify(581856)
	desc, err := m.s.GetDescriptor(ctx, op.DescID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581860)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581861)
	}
	__antithesis_instrumentation__.Notify(581857)
	mut := desc.NewBuilder().BuildCreatedMutable()
	mut.SetDropped()
	m.sd.AddSyntheticDescriptor(mut)
	return nil
}

func (m *visitor) DrainDescriptorName(_ context.Context, op scop.DrainDescriptorName) error {
	__antithesis_instrumentation__.Notify(581862)
	nameDetails := descpb.NameInfo{
		ParentID:       op.Namespace.DatabaseID,
		ParentSchemaID: op.Namespace.SchemaID,
		Name:           op.Namespace.Name,
	}
	m.s.AddDrainedName(op.Namespace.DescriptorID, nameDetails)
	return nil
}

func (m *visitor) DeleteDescriptor(_ context.Context, op scop.DeleteDescriptor) error {
	__antithesis_instrumentation__.Notify(581863)
	m.s.DeleteDescriptor(op.DescriptorID)
	return nil
}

func (m *visitor) RemoveAllTableComments(_ context.Context, op scop.RemoveAllTableComments) error {
	__antithesis_instrumentation__.Notify(581864)
	m.s.DeleteAllTableComments(op.TableID)
	return nil
}

func (m *visitor) RemoveTableComment(_ context.Context, op scop.RemoveTableComment) error {
	__antithesis_instrumentation__.Notify(581865)
	m.s.DeleteComment(op.TableID, 0, keys.TableCommentType)
	return nil
}

func (m *visitor) RemoveDatabaseComment(_ context.Context, op scop.RemoveDatabaseComment) error {
	__antithesis_instrumentation__.Notify(581866)
	m.s.DeleteComment(op.DatabaseID, 0, keys.DatabaseCommentType)
	return nil
}

func (m *visitor) RemoveSchemaComment(_ context.Context, op scop.RemoveSchemaComment) error {
	__antithesis_instrumentation__.Notify(581867)
	m.s.DeleteComment(op.SchemaID, 0, keys.SchemaCommentType)
	return nil
}

func (m *visitor) RemoveIndexComment(_ context.Context, op scop.RemoveIndexComment) error {
	__antithesis_instrumentation__.Notify(581868)
	m.s.DeleteComment(op.TableID, int(op.IndexID), keys.IndexCommentType)
	return nil
}

func (m *visitor) RemoveColumnComment(_ context.Context, op scop.RemoveColumnComment) error {
	__antithesis_instrumentation__.Notify(581869)
	m.s.DeleteComment(op.TableID, int(op.ColumnID), keys.ColumnCommentType)
	return nil
}

func (m *visitor) RemoveConstraintComment(
	ctx context.Context, op scop.RemoveConstraintComment,
) error {
	__antithesis_instrumentation__.Notify(581870)
	return m.s.DeleteConstraintComment(ctx, op.TableID, op.ConstraintID)
}

func (m *visitor) RemoveDatabaseRoleSettings(
	ctx context.Context, op scop.RemoveDatabaseRoleSettings,
) error {
	__antithesis_instrumentation__.Notify(581871)
	return m.s.DeleteDatabaseRoleSettings(ctx, op.DatabaseID)
}

func (m *visitor) DeleteSchedule(_ context.Context, op scop.DeleteSchedule) error {
	__antithesis_instrumentation__.Notify(581872)
	if op.ScheduleID != 0 {
		__antithesis_instrumentation__.Notify(581874)
		m.s.DeleteSchedule(op.ScheduleID)
	} else {
		__antithesis_instrumentation__.Notify(581875)
	}
	__antithesis_instrumentation__.Notify(581873)
	return nil
}
