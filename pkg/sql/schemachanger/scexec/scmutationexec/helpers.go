package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/errors"
)

func (m *visitor) checkOutTable(ctx context.Context, id descpb.ID) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(581911)
	desc, err := m.s.CheckOutDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581914)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581915)
	}
	__antithesis_instrumentation__.Notify(581912)
	mut, ok := desc.(*tabledesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(581916)
		return nil, catalog.WrapTableDescRefErr(id, catalog.NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(581917)
	}
	__antithesis_instrumentation__.Notify(581913)
	return mut, nil
}

func (m *visitor) checkOutDatabase(ctx context.Context, id descpb.ID) (*dbdesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(581918)
	desc, err := m.s.CheckOutDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581921)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581922)
	}
	__antithesis_instrumentation__.Notify(581919)
	mut, ok := desc.(*dbdesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(581923)
		return nil, catalog.WrapDatabaseDescRefErr(id, catalog.NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(581924)
	}
	__antithesis_instrumentation__.Notify(581920)
	return mut, nil
}

func (m *visitor) checkOutSchema(ctx context.Context, id descpb.ID) (*schemadesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(581925)
	desc, err := m.s.CheckOutDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581928)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581929)
	}
	__antithesis_instrumentation__.Notify(581926)
	mut, ok := desc.(*schemadesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(581930)
		return nil, catalog.WrapSchemaDescRefErr(id, catalog.NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(581931)
	}
	__antithesis_instrumentation__.Notify(581927)
	return mut, nil
}

var _ = ((*visitor)(nil)).checkOutSchema

func (m *visitor) checkOutType(ctx context.Context, id descpb.ID) (*typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(581932)
	desc, err := m.s.CheckOutDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581935)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581936)
	}
	__antithesis_instrumentation__.Notify(581933)
	mut, ok := desc.(*typedesc.Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(581937)
		return nil, catalog.WrapTypeDescRefErr(id, catalog.NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(581938)
	}
	__antithesis_instrumentation__.Notify(581934)
	return mut, nil
}

func mutationStateChange(
	tbl *tabledesc.Mutable, f MutationSelector, exp, next descpb.DescriptorMutation_State,
) error {
	__antithesis_instrumentation__.Notify(581939)
	mut, err := FindMutation(tbl, f)
	if err != nil {
		__antithesis_instrumentation__.Notify(581942)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581943)
	}
	__antithesis_instrumentation__.Notify(581940)
	m := &tbl.TableDesc().Mutations[mut.MutationOrdinal()]
	if m.State != exp {
		__antithesis_instrumentation__.Notify(581944)
		return errors.AssertionFailedf("update mutation for %d from %v to %v: unexpected state: %v",
			tbl.GetID(), exp, m.State, tbl)
	} else {
		__antithesis_instrumentation__.Notify(581945)
	}
	__antithesis_instrumentation__.Notify(581941)
	m.State = next
	return nil
}

func removeMutation(
	tbl *tabledesc.Mutable, f MutationSelector, exp descpb.DescriptorMutation_State,
) (descpb.DescriptorMutation, error) {
	__antithesis_instrumentation__.Notify(581946)
	mut, err := FindMutation(tbl, f)
	if err != nil {
		__antithesis_instrumentation__.Notify(581949)
		return descpb.DescriptorMutation{}, err
	} else {
		__antithesis_instrumentation__.Notify(581950)
	}
	__antithesis_instrumentation__.Notify(581947)
	foundIdx := mut.MutationOrdinal()
	cpy := tbl.Mutations[foundIdx]
	if cpy.State != exp {
		__antithesis_instrumentation__.Notify(581951)
		return descpb.DescriptorMutation{}, errors.AssertionFailedf(
			"remove mutation from %d: unexpected state: got %v, expected %v: %v",
			tbl.GetID(), cpy.State, exp, tbl,
		)
	} else {
		__antithesis_instrumentation__.Notify(581952)
	}
	__antithesis_instrumentation__.Notify(581948)
	tbl.Mutations = append(tbl.Mutations[:foundIdx], tbl.Mutations[foundIdx+1:]...)
	return cpy, nil
}

func columnNamesFromIDs(tbl *tabledesc.Mutable, columnIDs descpb.ColumnIDs) ([]string, error) {
	__antithesis_instrumentation__.Notify(581953)
	storeColNames := make([]string, 0, len(columnIDs))
	for _, colID := range columnIDs {
		__antithesis_instrumentation__.Notify(581955)
		column, err := tbl.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(581957)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581958)
		}
		__antithesis_instrumentation__.Notify(581956)
		storeColNames = append(storeColNames, column.GetName())
	}
	__antithesis_instrumentation__.Notify(581954)
	return storeColNames, nil
}

type MutationSelector func(mutation catalog.Mutation) (matches bool)

func FindMutation(
	tbl catalog.TableDescriptor, selector MutationSelector,
) (catalog.Mutation, error) {
	__antithesis_instrumentation__.Notify(581959)
	for _, mut := range tbl.AllMutations() {
		__antithesis_instrumentation__.Notify(581961)
		if selector(mut) {
			__antithesis_instrumentation__.Notify(581962)
			return mut, nil
		} else {
			__antithesis_instrumentation__.Notify(581963)
		}
	}
	__antithesis_instrumentation__.Notify(581960)
	return nil, errors.AssertionFailedf("matching mutation not found in table %d", tbl.GetID())
}

func MakeIndexIDMutationSelector(indexID descpb.IndexID) MutationSelector {
	__antithesis_instrumentation__.Notify(581964)
	return func(mut catalog.Mutation) bool {
		__antithesis_instrumentation__.Notify(581965)
		if mut.AsIndex() == nil {
			__antithesis_instrumentation__.Notify(581967)
			return false
		} else {
			__antithesis_instrumentation__.Notify(581968)
		}
		__antithesis_instrumentation__.Notify(581966)
		return mut.AsIndex().GetID() == indexID
	}
}

func MakeColumnIDMutationSelector(columnID descpb.ColumnID) MutationSelector {
	__antithesis_instrumentation__.Notify(581969)
	return func(mut catalog.Mutation) bool {
		__antithesis_instrumentation__.Notify(581970)
		if mut.AsColumn() == nil {
			__antithesis_instrumentation__.Notify(581972)
			return false
		} else {
			__antithesis_instrumentation__.Notify(581973)
		}
		__antithesis_instrumentation__.Notify(581971)
		return mut.AsColumn().GetID() == columnID
	}
}

func enqueueAddColumnMutation(tbl *tabledesc.Mutable, col *descpb.ColumnDescriptor) error {
	__antithesis_instrumentation__.Notify(581974)
	tbl.AddColumnMutation(col, descpb.DescriptorMutation_ADD)
	tbl.NextMutationID--
	return nil
}

func enqueueDropColumnMutation(tbl *tabledesc.Mutable, col *descpb.ColumnDescriptor) error {
	__antithesis_instrumentation__.Notify(581975)
	tbl.AddColumnMutation(col, descpb.DescriptorMutation_DROP)
	tbl.NextMutationID--
	return nil
}

func enqueueAddIndexMutation(tbl *tabledesc.Mutable, idx *descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(581976)
	if err := tbl.DeprecatedAddIndexMutation(idx, descpb.DescriptorMutation_ADD); err != nil {
		__antithesis_instrumentation__.Notify(581978)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581979)
	}
	__antithesis_instrumentation__.Notify(581977)
	tbl.NextMutationID--
	return nil
}

func enqueueDropIndexMutation(tbl *tabledesc.Mutable, idx *descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(581980)
	if err := tbl.DeprecatedAddIndexMutation(idx, descpb.DescriptorMutation_DROP); err != nil {
		__antithesis_instrumentation__.Notify(581982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581983)
	}
	__antithesis_instrumentation__.Notify(581981)
	tbl.NextMutationID--
	return nil
}

func updateColumnExprSequenceUsage(d *descpb.ColumnDescriptor) error {
	__antithesis_instrumentation__.Notify(581984)
	var all catalog.DescriptorIDSet
	for _, expr := range [3]*string{d.ComputeExpr, d.DefaultExpr, d.OnUpdateExpr} {
		__antithesis_instrumentation__.Notify(581986)
		if expr == nil {
			__antithesis_instrumentation__.Notify(581989)
			continue
		} else {
			__antithesis_instrumentation__.Notify(581990)
		}
		__antithesis_instrumentation__.Notify(581987)
		ids, err := sequenceIDsInExpr(*expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(581991)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581992)
		}
		__antithesis_instrumentation__.Notify(581988)
		ids.ForEach(all.Add)
	}
	__antithesis_instrumentation__.Notify(581985)
	d.UsesSequenceIds = all.Ordered()
	return nil
}

func sequenceIDsInExpr(expr string) (ids catalog.DescriptorIDSet, _ error) {
	__antithesis_instrumentation__.Notify(581993)
	e, err := parser.ParseExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(581997)
		return ids, err
	} else {
		__antithesis_instrumentation__.Notify(581998)
	}
	__antithesis_instrumentation__.Notify(581994)
	seqIdents, err := seqexpr.GetUsedSequences(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(581999)
		return ids, err
	} else {
		__antithesis_instrumentation__.Notify(582000)
	}
	__antithesis_instrumentation__.Notify(581995)
	for _, si := range seqIdents {
		__antithesis_instrumentation__.Notify(582001)
		ids.Add(descpb.ID(si.SeqID))
	}
	__antithesis_instrumentation__.Notify(581996)
	return ids, nil
}
