package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

type virtualDescriptors struct {
	vs catalog.VirtualSchemas
}

func makeVirtualDescriptors(schemas catalog.VirtualSchemas) virtualDescriptors {
	__antithesis_instrumentation__.Notify(265299)
	return virtualDescriptors{vs: schemas}
}

func (tc *virtualDescriptors) getSchemaByName(schemaName string) catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(265300)
	if tc.vs == nil {
		__antithesis_instrumentation__.Notify(265303)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265304)
	}
	__antithesis_instrumentation__.Notify(265301)
	if sc, ok := tc.vs.GetVirtualSchema(schemaName); ok {
		__antithesis_instrumentation__.Notify(265305)
		return sc.Desc()
	} else {
		__antithesis_instrumentation__.Notify(265306)
	}
	__antithesis_instrumentation__.Notify(265302)
	return nil
}

func (tc *virtualDescriptors) getObjectByName(
	schema string, object string, flags tree.ObjectLookupFlags, db string,
) (isVirtual bool, _ catalog.Descriptor, _ error) {
	__antithesis_instrumentation__.Notify(265307)
	if tc.vs == nil {
		__antithesis_instrumentation__.Notify(265313)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265314)
	}
	__antithesis_instrumentation__.Notify(265308)
	scEntry, ok := tc.vs.GetVirtualSchema(schema)
	if !ok {
		__antithesis_instrumentation__.Notify(265315)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265316)
	}
	__antithesis_instrumentation__.Notify(265309)
	desc, err := scEntry.GetObjectByName(object, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(265317)
		return true, nil, err
	} else {
		__antithesis_instrumentation__.Notify(265318)
	}
	__antithesis_instrumentation__.Notify(265310)
	if desc == nil {
		__antithesis_instrumentation__.Notify(265319)
		if flags.Required {
			__antithesis_instrumentation__.Notify(265321)
			obj := tree.NewQualifiedObjectName(db, schema, object, flags.DesiredObjectKind)
			return true, nil, sqlerrors.NewUndefinedObjectError(obj, flags.DesiredObjectKind)
		} else {
			__antithesis_instrumentation__.Notify(265322)
		}
		__antithesis_instrumentation__.Notify(265320)
		return true, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265323)
	}
	__antithesis_instrumentation__.Notify(265311)
	if flags.RequireMutable {
		__antithesis_instrumentation__.Notify(265324)
		return true, nil, catalog.NewMutableAccessToVirtualSchemaError(scEntry, object)
	} else {
		__antithesis_instrumentation__.Notify(265325)
	}
	__antithesis_instrumentation__.Notify(265312)
	return true, desc.Desc(), nil
}

func (tc virtualDescriptors) getByID(
	ctx context.Context, id descpb.ID, mutable bool,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265326)
	if tc.vs == nil {
		__antithesis_instrumentation__.Notify(265329)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265330)
	}
	__antithesis_instrumentation__.Notify(265327)
	if vd, found := tc.vs.GetVirtualObjectByID(id); found {
		__antithesis_instrumentation__.Notify(265331)
		if mutable {
			__antithesis_instrumentation__.Notify(265333)
			vs, found := tc.vs.GetVirtualSchemaByID(vd.Desc().GetParentSchemaID())
			if !found {
				__antithesis_instrumentation__.Notify(265335)
				return nil, errors.AssertionFailedf(
					"cannot resolve mutable virtual descriptor %d with unknown parent schema %d",
					id, vd.Desc().GetParentSchemaID(),
				)
			} else {
				__antithesis_instrumentation__.Notify(265336)
			}
			__antithesis_instrumentation__.Notify(265334)
			return nil, catalog.NewMutableAccessToVirtualSchemaError(vs, vd.Desc().GetName())
		} else {
			__antithesis_instrumentation__.Notify(265337)
		}
		__antithesis_instrumentation__.Notify(265332)
		return vd.Desc(), nil
	} else {
		__antithesis_instrumentation__.Notify(265338)
	}
	__antithesis_instrumentation__.Notify(265328)
	return tc.getSchemaByID(ctx, id, mutable)
}

func (tc virtualDescriptors) getSchemaByID(
	ctx context.Context, id descpb.ID, mutable bool,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(265339)
	if tc.vs == nil {
		__antithesis_instrumentation__.Notify(265341)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265342)
	}
	__antithesis_instrumentation__.Notify(265340)
	vs, found := tc.vs.GetVirtualSchemaByID(id)
	switch {
	case !found:
		__antithesis_instrumentation__.Notify(265343)
		return nil, nil
	case mutable:
		__antithesis_instrumentation__.Notify(265344)
		return nil, catalog.NewMutableAccessToVirtualSchemaError(vs, vs.Desc().GetName())
	default:
		__antithesis_instrumentation__.Notify(265345)
		return vs.Desc(), nil
	}
}

func (tc virtualDescriptors) maybeGetObjectNamesAndIDs(
	scName string, dbDesc catalog.DatabaseDescriptor, flags tree.DatabaseListFlags,
) (isVirtual bool, _ tree.TableNames, _ descpb.IDs) {
	__antithesis_instrumentation__.Notify(265346)
	if tc.vs == nil {
		__antithesis_instrumentation__.Notify(265350)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265351)
	}
	__antithesis_instrumentation__.Notify(265347)
	entry, ok := tc.vs.GetVirtualSchema(scName)
	if !ok {
		__antithesis_instrumentation__.Notify(265352)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265353)
	}
	__antithesis_instrumentation__.Notify(265348)
	names := make(tree.TableNames, 0, entry.NumTables())
	IDs := make(descpb.IDs, 0, entry.NumTables())
	schemaDesc := entry.Desc()
	entry.VisitTables(func(table catalog.VirtualObject) {
		__antithesis_instrumentation__.Notify(265354)
		name := tree.MakeTableNameWithSchema(
			tree.Name(dbDesc.GetName()), tree.Name(schemaDesc.GetName()), tree.Name(table.Desc().GetName()))
		name.ExplicitCatalog = flags.ExplicitPrefix
		name.ExplicitSchema = flags.ExplicitPrefix
		names = append(names, name)
		IDs = append(IDs, table.Desc().GetID())
	})
	__antithesis_instrumentation__.Notify(265349)
	return true, names, IDs
}
