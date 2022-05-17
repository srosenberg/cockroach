package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func (m *visitor) MakeAddedColumnDeleteOnly(
	ctx context.Context, op scop.MakeAddedColumnDeleteOnly,
) error {
	__antithesis_instrumentation__.Notify(581697)
	col := &descpb.ColumnDescriptor{
		ID:                      op.Column.ColumnID,
		Name:                    tabledesc.ColumnNamePlaceholder(op.Column.ColumnID),
		Hidden:                  op.Column.IsHidden,
		Inaccessible:            op.Column.IsInaccessible,
		GeneratedAsIdentityType: op.Column.GeneratedAsIdentityType,
		PGAttributeNum:          op.Column.PgAttributeNum,
	}
	if o := op.Column.GeneratedAsIdentitySequenceOption; o != "" {
		__antithesis_instrumentation__.Notify(581701)
		col.GeneratedAsIdentitySequenceOption = &o
	} else {
		__antithesis_instrumentation__.Notify(581702)
	}
	__antithesis_instrumentation__.Notify(581698)
	tbl, err := m.checkOutTable(ctx, op.Column.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581703)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581704)
	}
	__antithesis_instrumentation__.Notify(581699)
	if col.ID >= tbl.NextColumnID {
		__antithesis_instrumentation__.Notify(581705)
		tbl.NextColumnID = col.ID + 1
	} else {
		__antithesis_instrumentation__.Notify(581706)
	}
	__antithesis_instrumentation__.Notify(581700)
	return enqueueAddColumnMutation(tbl, col)
}

func (m *visitor) SetAddedColumnType(ctx context.Context, op scop.SetAddedColumnType) error {
	__antithesis_instrumentation__.Notify(581707)
	tbl, err := m.checkOutTable(ctx, op.ColumnType.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581712)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581713)
	}
	__antithesis_instrumentation__.Notify(581708)
	mut, err := FindMutation(tbl, MakeColumnIDMutationSelector(op.ColumnType.ColumnID))
	if err != nil {
		__antithesis_instrumentation__.Notify(581714)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581715)
	}
	__antithesis_instrumentation__.Notify(581709)
	col := mut.AsColumn().ColumnDesc()
	col.Type = op.ColumnType.Type
	col.Nullable = op.ColumnType.IsNullable
	col.Virtual = op.ColumnType.IsVirtual
	if ce := op.ColumnType.ComputeExpr; ce != nil {
		__antithesis_instrumentation__.Notify(581716)
		expr := string(ce.Expr)
		col.ComputeExpr = &expr
		col.UsesSequenceIds = ce.UsesSequenceIDs
	} else {
		__antithesis_instrumentation__.Notify(581717)
	}
	__antithesis_instrumentation__.Notify(581710)
	if col.ComputeExpr == nil || func() bool {
		__antithesis_instrumentation__.Notify(581718)
		return !col.Virtual == true
	}() == true {
		__antithesis_instrumentation__.Notify(581719)
		for i := range tbl.Families {
			__antithesis_instrumentation__.Notify(581720)
			fam := &tbl.Families[i]
			if fam.ID == op.ColumnType.FamilyID {
				__antithesis_instrumentation__.Notify(581721)
				fam.ColumnIDs = append(fam.ColumnIDs, col.ID)
				fam.ColumnNames = append(fam.ColumnNames, col.Name)
				break
			} else {
				__antithesis_instrumentation__.Notify(581722)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(581723)
	}
	__antithesis_instrumentation__.Notify(581711)
	return nil
}

func (m *visitor) MakeAddedColumnDeleteAndWriteOnly(
	ctx context.Context, op scop.MakeAddedColumnDeleteAndWriteOnly,
) error {
	__antithesis_instrumentation__.Notify(581724)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581726)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581727)
	}
	__antithesis_instrumentation__.Notify(581725)
	return mutationStateChange(
		tbl,
		MakeColumnIDMutationSelector(op.ColumnID),
		descpb.DescriptorMutation_DELETE_ONLY,
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
	)
}

func (m *visitor) MakeColumnPublic(ctx context.Context, op scop.MakeColumnPublic) error {
	__antithesis_instrumentation__.Notify(581728)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581731)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581732)
	}
	__antithesis_instrumentation__.Notify(581729)
	mut, err := removeMutation(
		tbl,
		MakeColumnIDMutationSelector(op.ColumnID),
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(581733)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581734)
	}
	__antithesis_instrumentation__.Notify(581730)

	tbl.Columns = append(tbl.Columns,
		*(protoutil.Clone(mut.GetColumn())).(*descpb.ColumnDescriptor))
	return nil
}

func (m *visitor) MakeDroppedColumnDeleteAndWriteOnly(
	ctx context.Context, op scop.MakeDroppedColumnDeleteAndWriteOnly,
) error {
	__antithesis_instrumentation__.Notify(581735)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581738)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581739)
	}
	__antithesis_instrumentation__.Notify(581736)
	for i, col := range tbl.PublicColumns() {
		__antithesis_instrumentation__.Notify(581740)
		if col.GetID() == op.ColumnID {
			__antithesis_instrumentation__.Notify(581741)
			desc := col.ColumnDescDeepCopy()
			tbl.Columns = append(tbl.Columns[:i], tbl.Columns[i+1:]...)
			return enqueueDropColumnMutation(tbl, &desc)
		} else {
			__antithesis_instrumentation__.Notify(581742)
		}
	}
	__antithesis_instrumentation__.Notify(581737)
	return errors.AssertionFailedf("failed to find column %d in table %q (%d)",
		op.ColumnID, tbl.GetName(), tbl.GetID())
}

func (m *visitor) MakeDroppedColumnDeleteOnly(
	ctx context.Context, op scop.MakeDroppedColumnDeleteOnly,
) error {
	__antithesis_instrumentation__.Notify(581743)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581745)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581746)
	}
	__antithesis_instrumentation__.Notify(581744)
	return mutationStateChange(
		tbl,
		MakeColumnIDMutationSelector(op.ColumnID),
		descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		descpb.DescriptorMutation_DELETE_ONLY,
	)
}

func (m *visitor) RemoveDroppedColumnType(
	ctx context.Context, op scop.RemoveDroppedColumnType,
) error {
	__antithesis_instrumentation__.Notify(581747)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(581751)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581752)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581753)
	}
	__antithesis_instrumentation__.Notify(581748)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581754)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581755)
	}
	__antithesis_instrumentation__.Notify(581749)
	mut, err := FindMutation(tbl, MakeColumnIDMutationSelector(op.ColumnID))
	if err != nil {
		__antithesis_instrumentation__.Notify(581756)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581757)
	}
	__antithesis_instrumentation__.Notify(581750)
	col := mut.AsColumn().ColumnDesc()
	col.ComputeExpr = nil
	col.Type = types.Any
	return nil
}

func (m *visitor) MakeColumnAbsent(ctx context.Context, op scop.MakeColumnAbsent) error {
	__antithesis_instrumentation__.Notify(581758)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(581762)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581763)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581764)
	}
	__antithesis_instrumentation__.Notify(581759)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581765)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581766)
	}
	__antithesis_instrumentation__.Notify(581760)
	mut, err := removeMutation(
		tbl,
		MakeColumnIDMutationSelector(op.ColumnID),
		descpb.DescriptorMutation_DELETE_ONLY,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(581767)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581768)
	}
	__antithesis_instrumentation__.Notify(581761)
	col := mut.GetColumn()
	tbl.RemoveColumnFromFamilyAndPrimaryIndex(col.ID)
	return nil
}

func (m *visitor) AddColumnFamily(ctx context.Context, op scop.AddColumnFamily) error {
	__antithesis_instrumentation__.Notify(581769)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581772)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581773)
	}
	__antithesis_instrumentation__.Notify(581770)
	family := descpb.ColumnFamilyDescriptor{
		Name: op.Name,
		ID:   op.FamilyID,
	}
	tbl.AddFamily(family)
	if family.ID >= tbl.NextFamilyID {
		__antithesis_instrumentation__.Notify(581774)
		tbl.NextFamilyID = family.ID + 1
	} else {
		__antithesis_instrumentation__.Notify(581775)
	}
	__antithesis_instrumentation__.Notify(581771)
	return nil
}

func (m *visitor) SetColumnName(ctx context.Context, op scop.SetColumnName) error {
	__antithesis_instrumentation__.Notify(581776)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581779)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581780)
	}
	__antithesis_instrumentation__.Notify(581777)
	col, err := tbl.FindColumnWithID(op.ColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581781)
		return errors.AssertionFailedf("column %d not found in table %q (%d)", op.ColumnID, tbl.GetName(), tbl.GetID())
	} else {
		__antithesis_instrumentation__.Notify(581782)
	}
	__antithesis_instrumentation__.Notify(581778)
	return tabledesc.RenameColumnInTable(tbl, col, tree.Name(op.Name), nil)
}

func (m *visitor) AddColumnDefaultExpression(
	ctx context.Context, op scop.AddColumnDefaultExpression,
) error {
	__antithesis_instrumentation__.Notify(581783)
	tbl, err := m.checkOutTable(ctx, op.Default.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581787)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581788)
	}
	__antithesis_instrumentation__.Notify(581784)
	col, err := tbl.FindColumnWithID(op.Default.ColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581789)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581790)
	}
	__antithesis_instrumentation__.Notify(581785)
	d := col.ColumnDesc()
	expr := string(op.Default.Expr)
	d.DefaultExpr = &expr
	refs := catalog.MakeDescriptorIDSet(d.UsesSequenceIds...)
	for _, seqID := range op.Default.UsesSequenceIDs {
		__antithesis_instrumentation__.Notify(581791)
		if refs.Contains(seqID) {
			__antithesis_instrumentation__.Notify(581793)
			continue
		} else {
			__antithesis_instrumentation__.Notify(581794)
		}
		__antithesis_instrumentation__.Notify(581792)
		d.UsesSequenceIds = append(d.UsesSequenceIds, seqID)
		refs.Add(seqID)
	}
	__antithesis_instrumentation__.Notify(581786)
	return nil
}

func (m *visitor) RemoveColumnDefaultExpression(
	ctx context.Context, op scop.RemoveColumnDefaultExpression,
) error {
	__antithesis_instrumentation__.Notify(581795)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(581799)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581800)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581801)
	}
	__antithesis_instrumentation__.Notify(581796)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581802)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581803)
	}
	__antithesis_instrumentation__.Notify(581797)
	col, err := tbl.FindColumnWithID(op.ColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581804)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581805)
	}
	__antithesis_instrumentation__.Notify(581798)
	d := col.ColumnDesc()
	d.DefaultExpr = nil
	return updateColumnExprSequenceUsage(d)
}

func (m *visitor) AddColumnOnUpdateExpression(
	ctx context.Context, op scop.AddColumnOnUpdateExpression,
) error {
	__antithesis_instrumentation__.Notify(581806)
	tbl, err := m.checkOutTable(ctx, op.OnUpdate.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581810)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581811)
	}
	__antithesis_instrumentation__.Notify(581807)
	col, err := tbl.FindColumnWithID(op.OnUpdate.ColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581812)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581813)
	}
	__antithesis_instrumentation__.Notify(581808)
	d := col.ColumnDesc()
	expr := string(op.OnUpdate.Expr)
	d.OnUpdateExpr = &expr
	refs := catalog.MakeDescriptorIDSet(d.UsesSequenceIds...)
	for _, seqID := range op.OnUpdate.UsesSequenceIDs {
		__antithesis_instrumentation__.Notify(581814)
		if refs.Contains(seqID) {
			__antithesis_instrumentation__.Notify(581816)
			continue
		} else {
			__antithesis_instrumentation__.Notify(581817)
		}
		__antithesis_instrumentation__.Notify(581815)
		d.UsesSequenceIds = append(d.UsesSequenceIds, seqID)
		refs.Add(seqID)
	}
	__antithesis_instrumentation__.Notify(581809)
	return nil
}

func (m *visitor) RemoveColumnOnUpdateExpression(
	ctx context.Context, op scop.RemoveColumnOnUpdateExpression,
) error {
	__antithesis_instrumentation__.Notify(581818)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(581822)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581823)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581824)
	}
	__antithesis_instrumentation__.Notify(581819)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581826)
	}
	__antithesis_instrumentation__.Notify(581820)
	col, err := tbl.FindColumnWithID(op.ColumnID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581827)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581828)
	}
	__antithesis_instrumentation__.Notify(581821)
	d := col.ColumnDesc()
	d.OnUpdateExpr = nil
	return updateColumnExprSequenceUsage(d)
}
