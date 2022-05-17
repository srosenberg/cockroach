package ingesting

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func WriteDescriptors(
	ctx context.Context,
	codec keys.SQLCodec,
	txn *kv.Txn,
	user security.SQLUsername,
	descsCol *descs.Collection,
	databases []catalog.DatabaseDescriptor,
	schemas []catalog.SchemaDescriptor,
	tables []catalog.TableDescriptor,
	types []catalog.TypeDescriptor,
	descCoverage tree.DescriptorCoverage,
	extra []roachpb.KeyValue,
	inheritParentName string,
) (err error) {
	__antithesis_instrumentation__.Notify(265582)
	ctx, span := tracing.ChildSpan(ctx, "WriteDescriptors")
	defer span.Finish()
	defer func() {
		__antithesis_instrumentation__.Notify(265590)
		err = errors.Wrapf(err, "restoring table desc and namespace entries")
	}()
	__antithesis_instrumentation__.Notify(265583)

	b := txn.NewBatch()

	wroteDBs := make(map[descpb.ID]catalog.DatabaseDescriptor)

	wroteSchemas := make(map[descpb.ID]catalog.SchemaDescriptor)
	for i := range databases {
		__antithesis_instrumentation__.Notify(265591)
		desc := databases[i]
		updatedPrivileges, err := GetIngestingDescriptorPrivileges(ctx, txn, descsCol, desc, user,
			wroteDBs, wroteSchemas, descCoverage)
		if err != nil {
			__antithesis_instrumentation__.Notify(265596)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265597)
		}
		__antithesis_instrumentation__.Notify(265592)
		if updatedPrivileges != nil {
			__antithesis_instrumentation__.Notify(265598)
			if mut, ok := desc.(*dbdesc.Mutable); ok {
				__antithesis_instrumentation__.Notify(265599)
				mut.Privileges = updatedPrivileges
			} else {
				__antithesis_instrumentation__.Notify(265600)
				log.Fatalf(ctx, "wrong type for database %d, %T, expected Mutable",
					desc.GetID(), desc)
			}
		} else {
			__antithesis_instrumentation__.Notify(265601)
		}
		__antithesis_instrumentation__.Notify(265593)
		privilegeDesc := desc.GetPrivileges()
		catprivilege.MaybeFixUsagePrivForTablesAndDBs(&privilegeDesc)
		if descCoverage == tree.RequestedDescriptors || func() bool {
			__antithesis_instrumentation__.Notify(265602)
			return desc.GetName() == inheritParentName == true
		}() == true {
			__antithesis_instrumentation__.Notify(265603)
			wroteDBs[desc.GetID()] = desc
		} else {
			__antithesis_instrumentation__.Notify(265604)
		}
		__antithesis_instrumentation__.Notify(265594)
		if err := descsCol.WriteDescToBatch(
			ctx, false, desc.(catalog.MutableDescriptor), b,
		); err != nil {
			__antithesis_instrumentation__.Notify(265605)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265606)
		}
		__antithesis_instrumentation__.Notify(265595)
		b.CPut(catalogkeys.EncodeNameKey(codec, desc), desc.GetID(), nil)

		if !desc.HasPublicSchemaWithDescriptor() {
			__antithesis_instrumentation__.Notify(265607)
			b.CPut(catalogkeys.MakeSchemaNameKey(codec, desc.GetID(), tree.PublicSchema), keys.PublicSchemaID, nil)
		} else {
			__antithesis_instrumentation__.Notify(265608)
		}
	}
	__antithesis_instrumentation__.Notify(265584)

	for i := range schemas {
		__antithesis_instrumentation__.Notify(265609)
		sc := schemas[i]
		updatedPrivileges, err := GetIngestingDescriptorPrivileges(ctx, txn, descsCol, sc, user,
			wroteDBs, wroteSchemas, descCoverage)
		if err != nil {
			__antithesis_instrumentation__.Notify(265614)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265615)
		}
		__antithesis_instrumentation__.Notify(265610)
		if updatedPrivileges != nil {
			__antithesis_instrumentation__.Notify(265616)
			if mut, ok := sc.(*schemadesc.Mutable); ok {
				__antithesis_instrumentation__.Notify(265617)
				mut.Privileges = updatedPrivileges
			} else {
				__antithesis_instrumentation__.Notify(265618)
				log.Fatalf(ctx, "wrong type for schema %d, %T, expected Mutable",
					sc.GetID(), sc)
			}
		} else {
			__antithesis_instrumentation__.Notify(265619)
		}
		__antithesis_instrumentation__.Notify(265611)
		if descCoverage == tree.RequestedDescriptors {
			__antithesis_instrumentation__.Notify(265620)
			wroteSchemas[sc.GetID()] = sc
		} else {
			__antithesis_instrumentation__.Notify(265621)
		}
		__antithesis_instrumentation__.Notify(265612)
		if err := descsCol.WriteDescToBatch(
			ctx, false, sc.(catalog.MutableDescriptor), b,
		); err != nil {
			__antithesis_instrumentation__.Notify(265622)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265623)
		}
		__antithesis_instrumentation__.Notify(265613)
		b.CPut(catalogkeys.EncodeNameKey(codec, sc), sc.GetID(), nil)
	}
	__antithesis_instrumentation__.Notify(265585)

	for i := range tables {
		__antithesis_instrumentation__.Notify(265624)
		table := tables[i]
		updatedPrivileges, err := GetIngestingDescriptorPrivileges(ctx, txn, descsCol, table, user,
			wroteDBs, wroteSchemas, descCoverage)
		if err != nil {
			__antithesis_instrumentation__.Notify(265629)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265630)
		}
		__antithesis_instrumentation__.Notify(265625)
		if updatedPrivileges != nil {
			__antithesis_instrumentation__.Notify(265631)
			if mut, ok := table.(*tabledesc.Mutable); ok {
				__antithesis_instrumentation__.Notify(265632)
				mut.Privileges = updatedPrivileges
			} else {
				__antithesis_instrumentation__.Notify(265633)
				log.Fatalf(ctx, "wrong type for table %d, %T, expected Mutable",
					table.GetID(), table)
			}
		} else {
			__antithesis_instrumentation__.Notify(265634)
		}
		__antithesis_instrumentation__.Notify(265626)
		privilegeDesc := table.GetPrivileges()
		catprivilege.MaybeFixUsagePrivForTablesAndDBs(&privilegeDesc)

		if err := processTableForMultiRegion(ctx, txn, descsCol, table); err != nil {
			__antithesis_instrumentation__.Notify(265635)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265636)
		}
		__antithesis_instrumentation__.Notify(265627)

		if err := descsCol.WriteDescToBatch(
			ctx, false, tables[i].(catalog.MutableDescriptor), b,
		); err != nil {
			__antithesis_instrumentation__.Notify(265637)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265638)
		}
		__antithesis_instrumentation__.Notify(265628)
		b.CPut(catalogkeys.EncodeNameKey(codec, table), table.GetID(), nil)
	}
	__antithesis_instrumentation__.Notify(265586)

	for i := range types {
		__antithesis_instrumentation__.Notify(265639)
		typ := types[i]
		updatedPrivileges, err := GetIngestingDescriptorPrivileges(ctx, txn, descsCol, typ, user,
			wroteDBs, wroteSchemas, descCoverage)
		if err != nil {
			__antithesis_instrumentation__.Notify(265643)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265644)
		}
		__antithesis_instrumentation__.Notify(265640)
		if updatedPrivileges != nil {
			__antithesis_instrumentation__.Notify(265645)
			if mut, ok := typ.(*typedesc.Mutable); ok {
				__antithesis_instrumentation__.Notify(265646)
				mut.Privileges = updatedPrivileges
			} else {
				__antithesis_instrumentation__.Notify(265647)
				log.Fatalf(ctx, "wrong type for type %d, %T, expected Mutable",
					typ.GetID(), typ)
			}
		} else {
			__antithesis_instrumentation__.Notify(265648)
		}
		__antithesis_instrumentation__.Notify(265641)
		if err := descsCol.WriteDescToBatch(
			ctx, false, typ.(catalog.MutableDescriptor), b,
		); err != nil {
			__antithesis_instrumentation__.Notify(265649)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265650)
		}
		__antithesis_instrumentation__.Notify(265642)
		b.CPut(catalogkeys.EncodeNameKey(codec, typ), typ.GetID(), nil)
	}
	__antithesis_instrumentation__.Notify(265587)

	for _, kv := range extra {
		__antithesis_instrumentation__.Notify(265651)
		b.InitPut(kv.Key, &kv.Value, false)
	}
	__antithesis_instrumentation__.Notify(265588)
	if err := txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(265652)
		if errors.HasType(err, (*roachpb.ConditionFailedError)(nil)) {
			__antithesis_instrumentation__.Notify(265654)
			return pgerror.Newf(pgcode.DuplicateObject, "table already exists")
		} else {
			__antithesis_instrumentation__.Notify(265655)
		}
		__antithesis_instrumentation__.Notify(265653)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265656)
	}
	__antithesis_instrumentation__.Notify(265589)
	return nil
}

func processTableForMultiRegion(
	ctx context.Context, txn *kv.Txn, descsCol *descs.Collection, table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(265657)
	_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
		ctx, txn, table.GetParentID(), tree.DatabaseLookupFlags{
			Required:       true,
			AvoidLeased:    true,
			IncludeOffline: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(265660)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265661)
	}
	__antithesis_instrumentation__.Notify(265658)

	if dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(265662)
		if table.GetLocalityConfig() == nil {
			__antithesis_instrumentation__.Notify(265663)
			table.(*tabledesc.Mutable).SetTableLocalityRegionalByTable(tree.PrimaryRegionNotSpecifiedName)
		} else {
			__antithesis_instrumentation__.Notify(265664)
		}
	} else {
		__antithesis_instrumentation__.Notify(265665)

		if table.GetLocalityConfig() != nil {
			__antithesis_instrumentation__.Notify(265666)
			return pgerror.Newf(pgcode.FeatureNotSupported,
				"cannot restore or create multi-region table %s into non-multi-region database %s",
				table.GetName(),
				dbDesc.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(265667)
		}
	}
	__antithesis_instrumentation__.Notify(265659)
	return nil
}
