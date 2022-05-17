package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetObjectDesc(
	ctx context.Context, txn *kv.Txn, db, schema, object string, flags tree.ObjectLookupFlags,
) (prefix catalog.ResolvedObjectPrefix, desc catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264745)
	return tc.getObjectByName(ctx, txn, db, schema, object, flags)
}

func (tc *Collection) getObjectByName(
	ctx context.Context,
	txn *kv.Txn,
	catalogName, schemaName, objectName string,
	flags tree.ObjectLookupFlags,
) (prefix catalog.ResolvedObjectPrefix, desc catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264746)
	defer func() {
		__antithesis_instrumentation__.Notify(264752)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(264754)
			return desc != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(264755)
			return !flags.Required == true
		}() == true {
			__antithesis_instrumentation__.Notify(264756)
			return
		} else {
			__antithesis_instrumentation__.Notify(264757)
		}
		__antithesis_instrumentation__.Notify(264753)
		if catalogName != "" && func() bool {
			__antithesis_instrumentation__.Notify(264758)
			return prefix.Database == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(264759)
			err = sqlerrors.NewUndefinedDatabaseError(catalogName)
		} else {
			__antithesis_instrumentation__.Notify(264760)
			if prefix.Schema == nil {
				__antithesis_instrumentation__.Notify(264761)
				err = sqlerrors.NewUndefinedSchemaError(schemaName)
			} else {
				__antithesis_instrumentation__.Notify(264762)
				tn := tree.MakeTableNameWithSchema(
					tree.Name(catalogName),
					tree.Name(schemaName),
					tree.Name(objectName))
				err = sqlerrors.NewUndefinedRelationError(&tn)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(264747)
	const alwaysLookupLeasedPublicSchema = false
	prefix, desc, err = tc.getObjectByNameIgnoringRequiredAndType(
		ctx, txn, catalogName, schemaName, objectName, flags,
		alwaysLookupLeasedPublicSchema,
	)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264763)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264764)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264765)
	}
	__antithesis_instrumentation__.Notify(264748)
	if desc.Adding() && func() bool {
		__antithesis_instrumentation__.Notify(264766)
		return desc.IsUncommittedVersion() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(264767)
		return (flags.RequireMutable || func() bool {
			__antithesis_instrumentation__.Notify(264768)
			return flags.CommonLookupFlags.AvoidLeased == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(264769)

		return prefix, desc, nil
	} else {
		__antithesis_instrumentation__.Notify(264770)
	}
	__antithesis_instrumentation__.Notify(264749)
	if dropped, err := filterDescriptorState(
		desc, flags.Required, flags.CommonLookupFlags,
	); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264771)
		return dropped == true
	}() == true {
		__antithesis_instrumentation__.Notify(264772)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264773)
	}
	__antithesis_instrumentation__.Notify(264750)
	switch t := desc.(type) {
	case catalog.TableDescriptor:
		__antithesis_instrumentation__.Notify(264774)

		switch flags.DesiredObjectKind {
		case tree.TableObject, tree.TypeObject:
			__antithesis_instrumentation__.Notify(264779)
		default:
			__antithesis_instrumentation__.Notify(264780)
			return prefix, nil, nil
		}
		__antithesis_instrumentation__.Notify(264775)
		tableDesc, err := tc.hydrateTypesInTableDesc(ctx, txn, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(264781)
			return prefix, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264782)
		}
		__antithesis_instrumentation__.Notify(264776)
		desc = tableDesc
		if flags.DesiredObjectKind == tree.TypeObject {
			__antithesis_instrumentation__.Notify(264783)

			if flags.RequireMutable {
				__antithesis_instrumentation__.Notify(264785)

				return prefix, nil, pgerror.Newf(pgcode.InsufficientPrivilege,
					"cannot modify table record type %q", objectName)
			} else {
				__antithesis_instrumentation__.Notify(264786)
			}
			__antithesis_instrumentation__.Notify(264784)
			desc, err = typedesc.CreateImplicitRecordTypeFromTableDesc(tableDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(264787)
				return prefix, nil, err
			} else {
				__antithesis_instrumentation__.Notify(264788)
			}
		} else {
			__antithesis_instrumentation__.Notify(264789)
		}
	case catalog.TypeDescriptor:
		__antithesis_instrumentation__.Notify(264777)
		if flags.DesiredObjectKind != tree.TypeObject {
			__antithesis_instrumentation__.Notify(264790)
			return prefix, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264791)
		}
	default:
		__antithesis_instrumentation__.Notify(264778)
		return prefix, nil, errors.AssertionFailedf(
			"unexpected object of type %T", t,
		)
	}
	__antithesis_instrumentation__.Notify(264751)
	return prefix, desc, nil
}

func (tc *Collection) getObjectByNameIgnoringRequiredAndType(
	ctx context.Context,
	txn *kv.Txn,
	catalogName, schemaName, objectName string,
	flags tree.ObjectLookupFlags,
	alwaysLookupLeasedPublicSchema bool,
) (prefix catalog.ResolvedObjectPrefix, _ catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264792)

	flags.Required = false

	avoidLeasedForParent := flags.AvoidLeased || func() bool {
		__antithesis_instrumentation__.Notify(264798)
		return flags.RequireMutable == true
	}() == true

	parentFlags := tree.DatabaseLookupFlags{
		Required:       flags.Required,
		AvoidLeased:    avoidLeasedForParent,
		IncludeDropped: flags.IncludeDropped,
		IncludeOffline: flags.IncludeOffline,
	}

	var db catalog.DatabaseDescriptor
	if catalogName != "" {
		__antithesis_instrumentation__.Notify(264799)
		db, err = tc.GetImmutableDatabaseByName(ctx, txn, catalogName, parentFlags)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(264800)
			return db == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(264801)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264802)
		}
	} else {
		__antithesis_instrumentation__.Notify(264803)
	}
	__antithesis_instrumentation__.Notify(264793)

	prefix.Database = db

	{
		__antithesis_instrumentation__.Notify(264804)
		isVirtual, virtualObject, err := tc.virtual.getObjectByName(
			schemaName, objectName, flags, catalogName,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(264806)
			return prefix, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264807)
		}
		__antithesis_instrumentation__.Notify(264805)
		if isVirtual {
			__antithesis_instrumentation__.Notify(264808)
			sc := tc.virtual.getSchemaByName(schemaName)
			return catalog.ResolvedObjectPrefix{
				Database: db,
				Schema:   sc,
			}, virtualObject, nil
		} else {
			__antithesis_instrumentation__.Notify(264809)
		}
	}
	__antithesis_instrumentation__.Notify(264794)

	if catalogName == "" {
		__antithesis_instrumentation__.Notify(264810)
		return catalog.ResolvedObjectPrefix{}, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264811)
	}
	__antithesis_instrumentation__.Notify(264795)

	sc, err := tc.getSchemaByNameMaybeLookingUpPublicSchema(
		ctx, txn, db, schemaName, parentFlags, alwaysLookupLeasedPublicSchema,
	)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264812)
		return sc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264813)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264814)
	}
	__antithesis_instrumentation__.Notify(264796)

	prefix.Schema = sc
	found, obj, err := tc.getByName(
		ctx, txn, db, sc, objectName, flags.AvoidLeased, flags.RequireMutable, flags.AvoidSynthetic,
		false,
	)
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(264815)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264816)

		if !alwaysLookupLeasedPublicSchema && func() bool {
			__antithesis_instrumentation__.Notify(264818)
			return sc.GetID() == keys.PublicSchemaID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(264819)
			return !tc.settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(264820)
			return tc.settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors-1) == true
		}() == true {
			__antithesis_instrumentation__.Notify(264821)
			const alwaysLookupLeasedPublicSchema = true
			return tc.getObjectByNameIgnoringRequiredAndType(
				ctx, txn, catalogName, schemaName, objectName, flags,
				alwaysLookupLeasedPublicSchema,
			)
		} else {
			__antithesis_instrumentation__.Notify(264822)
		}
		__antithesis_instrumentation__.Notify(264817)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264823)
	}
	__antithesis_instrumentation__.Notify(264797)
	return prefix, obj, nil
}
