package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type importTypeResolver struct {
	typeIDToDesc   map[descpb.ID]*descpb.TypeDescriptor
	typeNameToDesc map[string]*descpb.TypeDescriptor
}

func newImportTypeResolver(typeDescs []*descpb.TypeDescriptor) importTypeResolver {
	__antithesis_instrumentation__.Notify(494851)
	itr := importTypeResolver{
		typeIDToDesc:   make(map[descpb.ID]*descpb.TypeDescriptor),
		typeNameToDesc: make(map[string]*descpb.TypeDescriptor),
	}
	for _, typeDesc := range typeDescs {
		__antithesis_instrumentation__.Notify(494853)
		itr.typeIDToDesc[typeDesc.GetID()] = typeDesc
		itr.typeNameToDesc[typeDesc.GetName()] = typeDesc
	}
	__antithesis_instrumentation__.Notify(494852)
	return itr
}

var _ tree.TypeReferenceResolver = &importTypeResolver{}

func (i importTypeResolver) ResolveType(
	_ context.Context, _ *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(494854)
	return nil, errors.New("importTypeResolver does not implement ResolveType")
}

func (i importTypeResolver) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(494855)
	id, err := typedesc.UserDefinedTypeOIDToID(oid)
	if err != nil {
		__antithesis_instrumentation__.Notify(494858)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494859)
	}
	__antithesis_instrumentation__.Notify(494856)
	name, desc, err := i.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(494860)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494861)
	}
	__antithesis_instrumentation__.Notify(494857)
	return desc.MakeTypesT(ctx, &name, i)
}

var _ catalog.TypeDescriptorResolver = &importTypeResolver{}

func (i importTypeResolver) GetTypeDescriptor(
	_ context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(494862)
	var desc *descpb.TypeDescriptor
	var ok bool
	if desc, ok = i.typeIDToDesc[id]; !ok {
		__antithesis_instrumentation__.Notify(494864)
		return tree.TypeName{}, nil, errors.Newf("type descriptor could not be resolved for type id %d", id)
	} else {
		__antithesis_instrumentation__.Notify(494865)
	}
	__antithesis_instrumentation__.Notify(494863)
	typeDesc := typedesc.NewBuilder(desc).BuildImmutableType()
	name := tree.MakeUnqualifiedTypeName(desc.GetName())
	return name, typeDesc, nil
}
