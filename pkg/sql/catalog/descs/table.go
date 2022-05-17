package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetMutableTableByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ *tabledesc.Mutable, _ error) {
	__antithesis_instrumentation__.Notify(264913)
	flags.RequireMutable = true
	found, desc, err := tc.getTableByName(ctx, txn, name, flags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264915)
		return !found == true
	}() == true {
		__antithesis_instrumentation__.Notify(264916)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264917)
	}
	__antithesis_instrumentation__.Notify(264914)
	return true, desc.(*tabledesc.Mutable), nil
}

func (tc *Collection) GetImmutableTableByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ catalog.TableDescriptor, _ error) {
	__antithesis_instrumentation__.Notify(264918)
	flags.RequireMutable = false
	return tc.getTableByName(ctx, txn, name, flags)
}

func (tc *Collection) getTableByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(264919)
	flags.DesiredObjectKind = tree.TableObject
	_, desc, err := tc.getObjectByName(
		ctx, txn, name.Catalog(), name.Schema(), name.Object(), flags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264921)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264922)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264923)
	}
	__antithesis_instrumentation__.Notify(264920)
	return true, desc.(catalog.TableDescriptor), nil
}

func (tc *Collection) GetLeasedImmutableTableByID(
	ctx context.Context, txn *kv.Txn, tableID descpb.ID,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264924)
	desc, _, err := tc.leased.getByID(ctx, tc.deadlineHolder(txn), tableID)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264928)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264929)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264930)
	}
	__antithesis_instrumentation__.Notify(264925)
	table, err := catalog.AsTableDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(264931)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264932)
	}
	__antithesis_instrumentation__.Notify(264926)
	hydrated, err := tc.hydrateTypesInTableDesc(ctx, txn, table)
	if err != nil {
		__antithesis_instrumentation__.Notify(264933)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264934)
	}
	__antithesis_instrumentation__.Notify(264927)
	return hydrated, nil
}

func (tc *Collection) GetUncommittedMutableTableByID(id descpb.ID) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(264935)
	if imm, status := tc.uncommitted.getImmutableByID(id); imm == nil || func() bool {
		__antithesis_instrumentation__.Notify(264939)
		return status == notValidatedYet == true
	}() == true {
		__antithesis_instrumentation__.Notify(264940)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264941)
	}
	__antithesis_instrumentation__.Notify(264936)
	mut, err := tc.uncommitted.checkOut(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(264942)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264943)
	}
	__antithesis_instrumentation__.Notify(264937)
	if table, ok := mut.(*tabledesc.Mutable); ok {
		__antithesis_instrumentation__.Notify(264944)
		return table, nil
	} else {
		__antithesis_instrumentation__.Notify(264945)
	}
	__antithesis_instrumentation__.Notify(264938)

	return nil, tc.uncommitted.checkIn(mut)
}

func (tc *Collection) GetMutableTableByID(
	ctx context.Context, txn *kv.Txn, tableID descpb.ID, flags tree.ObjectLookupFlags,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(264946)
	flags.RequireMutable = true
	desc, err := tc.getTableByID(ctx, txn, tableID, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(264948)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264949)
	}
	__antithesis_instrumentation__.Notify(264947)
	return desc.(*tabledesc.Mutable), nil
}

func (tc *Collection) GetMutableTableVersionByID(
	ctx context.Context, tableID descpb.ID, txn *kv.Txn,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(264950)
	return tc.GetMutableTableByID(ctx, txn, tableID, tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			IncludeOffline: true,
			IncludeDropped: true,
		},
	})
}

func (tc *Collection) GetImmutableTableByID(
	ctx context.Context, txn *kv.Txn, tableID descpb.ID, flags tree.ObjectLookupFlags,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264951)
	flags.RequireMutable = false
	desc, err := tc.getTableByID(ctx, txn, tableID, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(264953)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264954)
	}
	__antithesis_instrumentation__.Notify(264952)
	return desc, nil
}

func (tc *Collection) getTableByID(
	ctx context.Context, txn *kv.Txn, tableID descpb.ID, flags tree.ObjectLookupFlags,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264955)
	descs, err := tc.getDescriptorsByID(ctx, txn, flags.CommonLookupFlags, tableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(264959)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(264961)
			return nil, sqlerrors.NewUndefinedRelationError(
				&tree.TableRef{TableID: int64(tableID)})
		} else {
			__antithesis_instrumentation__.Notify(264962)
		}
		__antithesis_instrumentation__.Notify(264960)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264963)
	}
	__antithesis_instrumentation__.Notify(264956)
	table, ok := descs[0].(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(264964)
		return nil, sqlerrors.NewUndefinedRelationError(
			&tree.TableRef{TableID: int64(tableID)})
	} else {
		__antithesis_instrumentation__.Notify(264965)
	}
	__antithesis_instrumentation__.Notify(264957)
	hydrated, err := tc.hydrateTypesInTableDesc(ctx, txn, table)
	if err != nil {
		__antithesis_instrumentation__.Notify(264966)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264967)
	}
	__antithesis_instrumentation__.Notify(264958)
	return hydrated, nil
}
