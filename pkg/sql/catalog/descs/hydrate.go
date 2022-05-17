package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
)

func (tc *Collection) hydrateTypesInTableDesc(
	ctx context.Context, txn *kv.Txn, desc catalog.TableDescriptor,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264504)
	if desc.Dropped() {
		__antithesis_instrumentation__.Notify(264506)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(264507)
	}
	__antithesis_instrumentation__.Notify(264505)
	switch t := desc.(type) {
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(264508)

		getType := func(ctx context.Context, id descpb.ID) (tree.TypeName, catalog.TypeDescriptor, error) {
			__antithesis_instrumentation__.Notify(264516)
			desc, err := tc.GetMutableTypeVersionByID(ctx, txn, id)
			if err != nil {
				__antithesis_instrumentation__.Notify(264520)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264521)
			}
			__antithesis_instrumentation__.Notify(264517)
			dbDesc, err := tc.GetMutableDescriptorByID(ctx, txn, desc.ParentID)
			if err != nil {
				__antithesis_instrumentation__.Notify(264522)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264523)
			}
			__antithesis_instrumentation__.Notify(264518)
			sc, err := tc.getSchemaByID(
				ctx, txn, desc.ParentSchemaID,
				tree.SchemaLookupFlags{
					Required:       true,
					IncludeOffline: true,
					RequireMutable: true,
				},
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(264524)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264525)
			}
			__antithesis_instrumentation__.Notify(264519)
			name := tree.MakeQualifiedTypeName(dbDesc.GetName(), sc.GetName(), desc.Name)
			return name, desc, nil
		}
		__antithesis_instrumentation__.Notify(264509)

		return desc, typedesc.HydrateTypesInTableDescriptor(ctx, t.TableDesc(), typedesc.TypeLookupFunc(getType))
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(264510)

		if !t.ContainsUserDefinedTypes() {
			__antithesis_instrumentation__.Notify(264526)
			return desc, nil
		} else {
			__antithesis_instrumentation__.Notify(264527)
		}
		__antithesis_instrumentation__.Notify(264511)

		getType := typedesc.TypeLookupFunc(func(
			ctx context.Context, id descpb.ID,
		) (tree.TypeName, catalog.TypeDescriptor, error) {
			__antithesis_instrumentation__.Notify(264528)
			desc, err := tc.GetImmutableTypeByID(ctx, txn, id, tree.ObjectLookupFlags{})
			if err != nil {
				__antithesis_instrumentation__.Notify(264532)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264533)
			}
			__antithesis_instrumentation__.Notify(264529)
			_, dbDesc, err := tc.GetImmutableDatabaseByID(ctx, txn, desc.GetParentID(),
				tree.DatabaseLookupFlags{Required: true})
			if err != nil {
				__antithesis_instrumentation__.Notify(264534)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264535)
			}
			__antithesis_instrumentation__.Notify(264530)
			sc, err := tc.GetImmutableSchemaByID(
				ctx, txn, desc.GetParentSchemaID(), tree.SchemaLookupFlags{Required: true})
			if err != nil {
				__antithesis_instrumentation__.Notify(264536)
				return tree.TypeName{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264537)
			}
			__antithesis_instrumentation__.Notify(264531)
			name := tree.MakeQualifiedTypeName(dbDesc.GetName(), sc.GetName(), desc.GetName())
			return name, desc, nil
		})
		__antithesis_instrumentation__.Notify(264512)

		if tc.hydratedTables != nil && func() bool {
			__antithesis_instrumentation__.Notify(264538)
			return tc.uncommitted.descs.GetByID(desc.GetID()) == nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(264539)
			return tc.synthetic.descs.GetByID(desc.GetID()) == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(264540)
			hydrated, err := tc.hydratedTables.GetHydratedTableDescriptor(ctx, t, getType)
			if err != nil {
				__antithesis_instrumentation__.Notify(264542)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(264543)
			}
			__antithesis_instrumentation__.Notify(264541)
			if hydrated != nil {
				__antithesis_instrumentation__.Notify(264544)
				return hydrated, nil
			} else {
				__antithesis_instrumentation__.Notify(264545)
			}

		} else {
			__antithesis_instrumentation__.Notify(264546)
		}
		__antithesis_instrumentation__.Notify(264513)

		mut := t.NewBuilder().BuildExistingMutable().(*tabledesc.Mutable)
		if err := typedesc.HydrateTypesInTableDescriptor(ctx, mut.TableDesc(), getType); err != nil {
			__antithesis_instrumentation__.Notify(264547)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264548)
		}
		__antithesis_instrumentation__.Notify(264514)
		return mut.ImmutableCopy().(catalog.TableDescriptor), nil
	default:
		__antithesis_instrumentation__.Notify(264515)
		return desc, nil
	}
}

func HydrateGivenDescriptors(ctx context.Context, descs []catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(264549)

	dbDescs := make(map[descpb.ID]catalog.DatabaseDescriptor)
	typDescs := make(map[descpb.ID]catalog.TypeDescriptor)
	schemaDescs := make(map[descpb.ID]catalog.SchemaDescriptor)
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(264552)
		switch desc := desc.(type) {
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(264553)
			dbDescs[desc.GetID()] = desc
		case catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(264554)
			typDescs[desc.GetID()] = desc
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(264555)
			schemaDescs[desc.GetID()] = desc
		}
	}
	__antithesis_instrumentation__.Notify(264550)

	if len(typDescs) > 0 {
		__antithesis_instrumentation__.Notify(264556)

		typeLookup := func(ctx context.Context, id descpb.ID) (tree.TypeName, catalog.TypeDescriptor, error) {
			__antithesis_instrumentation__.Notify(264558)
			typDesc, ok := typDescs[id]
			if !ok {
				__antithesis_instrumentation__.Notify(264562)
				n := tree.MakeUnresolvedName(fmt.Sprintf("[%d]", id))
				return tree.TypeName{}, nil, sqlerrors.NewUndefinedObjectError(&n,
					tree.TypeObject)
			} else {
				__antithesis_instrumentation__.Notify(264563)
			}
			__antithesis_instrumentation__.Notify(264559)
			dbDesc, ok := dbDescs[typDesc.GetParentID()]
			if !ok {
				__antithesis_instrumentation__.Notify(264564)
				n := fmt.Sprintf("[%d]", typDesc.GetParentID())
				return tree.TypeName{}, nil, sqlerrors.NewUndefinedDatabaseError(n)
			} else {
				__antithesis_instrumentation__.Notify(264565)
			}
			__antithesis_instrumentation__.Notify(264560)

			var scName string
			switch typDesc.GetParentSchemaID() {

			case keys.PublicSchemaID:
				__antithesis_instrumentation__.Notify(264566)
				scName = tree.PublicSchema
			default:
				__antithesis_instrumentation__.Notify(264567)
				scName = schemaDescs[typDesc.GetParentSchemaID()].GetName()
			}
			__antithesis_instrumentation__.Notify(264561)
			name := tree.MakeQualifiedTypeName(dbDesc.GetName(), scName, typDesc.GetName())
			return name, typDesc, nil
		}
		__antithesis_instrumentation__.Notify(264557)

		for i := range descs {
			__antithesis_instrumentation__.Notify(264568)
			desc := descs[i]

			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(264570)
				continue
			} else {
				__antithesis_instrumentation__.Notify(264571)
			}
			__antithesis_instrumentation__.Notify(264569)
			tblDesc, ok := desc.(catalog.TableDescriptor)
			if ok {
				__antithesis_instrumentation__.Notify(264572)
				if err := typedesc.HydrateTypesInTableDescriptor(
					ctx,
					tblDesc.TableDesc(),
					typedesc.TypeLookupFunc(typeLookup),
				); err != nil {
					__antithesis_instrumentation__.Notify(264573)
					return err
				} else {
					__antithesis_instrumentation__.Notify(264574)
				}
			} else {
				__antithesis_instrumentation__.Notify(264575)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(264576)
	}
	__antithesis_instrumentation__.Notify(264551)
	return nil
}
