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
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type DistSQLTypeResolver struct {
	descriptors *Collection
	txn         *kv.Txn
}

func NewDistSQLTypeResolver(descs *Collection, txn *kv.Txn) DistSQLTypeResolver {
	__antithesis_instrumentation__.Notify(264473)
	return DistSQLTypeResolver{
		descriptors: descs,
		txn:         txn,
	}
}

func (dt *DistSQLTypeResolver) ResolveType(
	context.Context, *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(264474)
	return nil, errors.AssertionFailedf("cannot resolve types in DistSQL by name")
}

func (dt *DistSQLTypeResolver) ResolveTypeByOID(
	ctx context.Context, oid oid.Oid,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(264475)
	id, err := typedesc.UserDefinedTypeOIDToID(oid)
	if err != nil {
		__antithesis_instrumentation__.Notify(264478)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264479)
	}
	__antithesis_instrumentation__.Notify(264476)
	name, desc, err := dt.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(264480)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264481)
	}
	__antithesis_instrumentation__.Notify(264477)
	return desc.MakeTypesT(ctx, &name, dt)
}

func (dt *DistSQLTypeResolver) GetTypeDescriptor(
	ctx context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(264482)
	flags := tree.CommonLookupFlags{
		Required: true,
	}
	descs, err := dt.descriptors.getDescriptorsByID(ctx, dt.txn, flags, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(264485)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264486)
	}
	__antithesis_instrumentation__.Notify(264483)
	var typeDesc catalog.TypeDescriptor
	switch t := descs[0].(type) {
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(264487)

		typeDesc = t
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(264488)

		t, err = dt.descriptors.hydrateTypesInTableDesc(ctx, dt.txn, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(264491)
			return tree.TypeName{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264492)
		}
		__antithesis_instrumentation__.Notify(264489)
		typeDesc, err = typedesc.CreateImplicitRecordTypeFromTableDesc(t)
		if err != nil {
			__antithesis_instrumentation__.Notify(264493)
			return tree.TypeName{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264494)
		}
	default:
		__antithesis_instrumentation__.Notify(264490)
		return tree.TypeName{}, nil, pgerror.Newf(pgcode.WrongObjectType,
			"descriptor %d is a %s not a %s", id, t.DescriptorType(), catalog.Type)
	}
	__antithesis_instrumentation__.Notify(264484)
	name := tree.MakeUnqualifiedTypeName(typeDesc.GetName())
	return name, typeDesc, nil
}

func (dt *DistSQLTypeResolver) HydrateTypeSlice(ctx context.Context, typs []*types.T) error {
	__antithesis_instrumentation__.Notify(264495)
	for _, t := range typs {
		__antithesis_instrumentation__.Notify(264497)
		if err := typedesc.EnsureTypeIsHydrated(ctx, t, dt); err != nil {
			__antithesis_instrumentation__.Notify(264498)
			return err
		} else {
			__antithesis_instrumentation__.Notify(264499)
		}
	}
	__antithesis_instrumentation__.Notify(264496)
	return nil
}
