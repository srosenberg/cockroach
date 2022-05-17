package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetMutableTypeByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ *typedesc.Mutable, _ error) {
	__antithesis_instrumentation__.Notify(265087)
	flags.RequireMutable = true
	found, desc, err := tc.getTypeByName(ctx, txn, name, flags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(265089)
		return !found == true
	}() == true {
		__antithesis_instrumentation__.Notify(265090)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(265091)
	}
	__antithesis_instrumentation__.Notify(265088)
	return true, desc.(*typedesc.Mutable), nil
}

func (tc *Collection) GetImmutableTypeByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ catalog.TypeDescriptor, _ error) {
	__antithesis_instrumentation__.Notify(265092)
	flags.RequireMutable = false
	return tc.getTypeByName(ctx, txn, name, flags)
}

func (tc *Collection) getTypeByName(
	ctx context.Context, txn *kv.Txn, name tree.ObjectName, flags tree.ObjectLookupFlags,
) (found bool, _ catalog.TypeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(265093)
	flags.DesiredObjectKind = tree.TypeObject
	_, desc, err := tc.getObjectByName(
		ctx, txn, name.Catalog(), name.Schema(), name.Object(), flags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(265095)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(265096)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(265097)
	}
	__antithesis_instrumentation__.Notify(265094)
	return true, desc.(catalog.TypeDescriptor), nil
}

func (tc *Collection) GetMutableTypeVersionByID(
	ctx context.Context, txn *kv.Txn, typeID descpb.ID,
) (*typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(265098)
	return tc.GetMutableTypeByID(ctx, txn, typeID, tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			IncludeOffline: true,
			IncludeDropped: true,
		},
	})
}

func (tc *Collection) GetMutableTypeByID(
	ctx context.Context, txn *kv.Txn, typeID descpb.ID, flags tree.ObjectLookupFlags,
) (*typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(265099)
	flags.RequireMutable = true
	desc, err := tc.getTypeByID(ctx, txn, typeID, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(265102)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265103)
	}
	__antithesis_instrumentation__.Notify(265100)
	switch t := desc.(type) {
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(265104)
		return t, nil
	case *typedesc.TableImplicitRecordType:
		__antithesis_instrumentation__.Notify(265105)
		return nil, pgerror.Newf(pgcode.DependentObjectsStillExist, "cannot modify table record type %q", desc.GetName())
	}
	__antithesis_instrumentation__.Notify(265101)
	return nil,
		errors.AssertionFailedf("unhandled type descriptor type %T during GetMutableTypeByID", desc)
}

func (tc *Collection) GetImmutableTypeByID(
	ctx context.Context, txn *kv.Txn, typeID descpb.ID, flags tree.ObjectLookupFlags,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(265106)
	flags.RequireMutable = false
	return tc.getTypeByID(ctx, txn, typeID, flags)
}

func (tc *Collection) getTypeByID(
	ctx context.Context, txn *kv.Txn, typeID descpb.ID, flags tree.ObjectLookupFlags,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(265107)
	descs, err := tc.getDescriptorsByID(ctx, txn, flags.CommonLookupFlags, typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(265110)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(265112)
			return nil, pgerror.Newf(
				pgcode.UndefinedObject, "type with ID %d does not exist", typeID)
		} else {
			__antithesis_instrumentation__.Notify(265113)
		}
		__antithesis_instrumentation__.Notify(265111)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265114)
	}
	__antithesis_instrumentation__.Notify(265108)
	switch t := descs[0].(type) {
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(265115)

		return t, nil
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(265116)

		t, err = tc.hydrateTypesInTableDesc(ctx, txn, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(265118)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265119)
		}
		__antithesis_instrumentation__.Notify(265117)
		return typedesc.CreateImplicitRecordTypeFromTableDesc(t)
	}
	__antithesis_instrumentation__.Notify(265109)
	return nil, pgerror.Newf(
		pgcode.UndefinedObject, "type with ID %d does not exist", typeID)
}
