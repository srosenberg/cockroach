package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func publicSchemaMigration(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128495)
	query := `
  SELECT ns_db.id
    FROM system.namespace AS ns_db
         INNER JOIN system.namespace
                AS ns_sc ON (
                            ns_db.id
                            = ns_sc."parentID"
                        )
   WHERE ns_db.id != 1
     AND ns_db."parentSchemaID" = 0
     AND ns_db."parentID" = 0
     AND ns_sc."parentSchemaID" = 0
     AND ns_sc.name = 'public'
     AND ns_sc.id = 29
ORDER BY ns_db.id ASC;
`
	rows, err := d.InternalExecutor.QueryIterator(
		ctx, "get_databases_with_synthetic_public_schemas", nil, query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(128499)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128500)
	}
	__antithesis_instrumentation__.Notify(128496)
	var databaseIDs []descpb.ID
	for ok, err := rows.Next(ctx); ok; ok, err = rows.Next(ctx) {
		__antithesis_instrumentation__.Notify(128501)
		if err != nil {
			__antithesis_instrumentation__.Notify(128503)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128504)
		}
		__antithesis_instrumentation__.Notify(128502)
		parentID := descpb.ID(tree.MustBeDInt(rows.Cur()[0]))
		databaseIDs = append(databaseIDs, parentID)
	}
	__antithesis_instrumentation__.Notify(128497)

	for _, dbID := range databaseIDs {
		__antithesis_instrumentation__.Notify(128505)
		if err := createPublicSchemaForDatabase(ctx, dbID, d); err != nil {
			__antithesis_instrumentation__.Notify(128506)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128507)
		}
	}
	__antithesis_instrumentation__.Notify(128498)

	return nil
}

func createPublicSchemaForDatabase(
	ctx context.Context, dbID descpb.ID, d migration.TenantDeps,
) error {
	__antithesis_instrumentation__.Notify(128508)
	return d.CollectionFactory.Txn(ctx, d.InternalExecutor, d.DB,
		func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
			__antithesis_instrumentation__.Notify(128509)
			return createPublicSchemaDescriptor(ctx, txn, descriptors, dbID, d)
		})
}

func createPublicSchemaDescriptor(
	ctx context.Context,
	txn *kv.Txn,
	descriptors *descs.Collection,
	dbID descpb.ID,
	d migration.TenantDeps,
) error {
	__antithesis_instrumentation__.Notify(128510)
	_, desc, err := descriptors.GetImmutableDatabaseByID(ctx, txn, dbID, tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(128519)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128520)
	}
	__antithesis_instrumentation__.Notify(128511)
	if desc.HasPublicSchemaWithDescriptor() {
		__antithesis_instrumentation__.Notify(128521)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(128522)
	}
	__antithesis_instrumentation__.Notify(128512)
	dbDescBuilder := dbdesc.NewBuilder(desc.DatabaseDesc())
	dbDesc := dbDescBuilder.BuildExistingMutableDatabase()

	b := txn.NewBatch()

	publicSchemaDesc, _, err := sql.CreateSchemaDescriptorWithPrivileges(
		ctx, d.DB, d.Codec, desc, tree.PublicSchema, security.AdminRoleName(), security.AdminRoleName(), true,
	)

	publicSchemaDesc.Privileges.Grant(
		security.PublicRoleName(),
		privilege.List{privilege.CREATE, privilege.USAGE},
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(128523)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128524)
	}
	__antithesis_instrumentation__.Notify(128513)
	publicSchemaID := publicSchemaDesc.GetID()
	newKey := catalogkeys.MakeSchemaNameKey(d.Codec, dbID, publicSchemaDesc.GetName())
	oldKey := catalogkeys.EncodeNameKey(d.Codec, catalogkeys.NewNameKeyComponents(dbID, keys.RootNamespaceID, tree.PublicSchema))

	b.Del(oldKey)
	b.CPut(newKey, publicSchemaID, nil)
	if err := descriptors.Direct().WriteNewDescToBatch(
		ctx,
		false,
		b,
		publicSchemaDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(128525)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128526)
	}
	__antithesis_instrumentation__.Notify(128514)

	if dbDesc.Schemas == nil {
		__antithesis_instrumentation__.Notify(128527)
		dbDesc.Schemas = map[string]descpb.DatabaseDescriptor_SchemaInfo{
			tree.PublicSchema: {
				ID: publicSchemaID,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(128528)
		dbDesc.Schemas[tree.PublicSchema] = descpb.DatabaseDescriptor_SchemaInfo{
			ID: publicSchemaID,
		}
	}
	__antithesis_instrumentation__.Notify(128515)
	if err := descriptors.WriteDescToBatch(ctx, false, dbDesc, b); err != nil {
		__antithesis_instrumentation__.Notify(128529)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128530)
	}
	__antithesis_instrumentation__.Notify(128516)
	all, err := descriptors.GetAllDescriptors(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(128531)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128532)
	}
	__antithesis_instrumentation__.Notify(128517)
	allDescriptors := all.OrderedDescriptors()
	if err := migrateObjectsInDatabase(ctx, dbID, d, txn, publicSchemaID, descriptors, allDescriptors); err != nil {
		__antithesis_instrumentation__.Notify(128533)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128534)
	}
	__antithesis_instrumentation__.Notify(128518)

	return txn.Run(ctx, b)
}

func migrateObjectsInDatabase(
	ctx context.Context,
	dbID descpb.ID,
	d migration.TenantDeps,
	txn *kv.Txn,
	newPublicSchemaID descpb.ID,
	descriptors *descs.Collection,
	allDescriptors []catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(128535)
	const minBatchSizeInBytes = 1 << 20
	currSize := 0
	var modifiedDescs []catalog.MutableDescriptor
	batch := txn.NewBatch()
	for _, desc := range allDescriptors {
		__antithesis_instrumentation__.Notify(128538)

		if desc.Dropped() || func() bool {
			__antithesis_instrumentation__.Notify(128542)
			return desc.GetParentID() != dbID == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(128543)
			return (desc.GetParentSchemaID() != keys.PublicSchemaID && func() bool {
				__antithesis_instrumentation__.Notify(128544)
				return desc.GetParentSchemaID() != descpb.InvalidID == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(128545)
			continue
		} else {
			__antithesis_instrumentation__.Notify(128546)
		}
		__antithesis_instrumentation__.Notify(128539)
		b := desc.NewBuilder()
		updateDesc := func(mut catalog.MutableDescriptor, newPublicSchemaID descpb.ID) {
			__antithesis_instrumentation__.Notify(128547)
			oldKey := catalogkeys.MakeObjectNameKey(d.Codec, mut.GetParentID(), mut.GetParentSchemaID(), mut.GetName())
			batch.Del(oldKey)
			newKey := catalogkeys.MakeObjectNameKey(d.Codec, mut.GetParentID(), newPublicSchemaID, mut.GetName())
			batch.Put(newKey, mut.GetID())
			modifiedDescs = append(modifiedDescs, mut)
		}
		__antithesis_instrumentation__.Notify(128540)
		switch mut := b.BuildExistingMutable().(type) {
		case *dbdesc.Mutable, *schemadesc.Mutable:
			__antithesis_instrumentation__.Notify(128548)

		case *tabledesc.Mutable:
			__antithesis_instrumentation__.Notify(128549)
			updateDesc(mut, newPublicSchemaID)
			mut.UnexposedParentSchemaID = newPublicSchemaID
			currSize += mut.Size()
		case *typedesc.Mutable:
			__antithesis_instrumentation__.Notify(128550)
			updateDesc(mut, newPublicSchemaID)
			mut.ParentSchemaID = newPublicSchemaID
			currSize += mut.Size()
		}
		__antithesis_instrumentation__.Notify(128541)

		if currSize >= minBatchSizeInBytes {
			__antithesis_instrumentation__.Notify(128551)
			for _, modified := range modifiedDescs {
				__antithesis_instrumentation__.Notify(128554)
				err := descriptors.WriteDescToBatch(
					ctx, false, modified, batch,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(128555)
					return err
				} else {
					__antithesis_instrumentation__.Notify(128556)
				}
			}
			__antithesis_instrumentation__.Notify(128552)
			if err := txn.Run(ctx, batch); err != nil {
				__antithesis_instrumentation__.Notify(128557)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128558)
			}
			__antithesis_instrumentation__.Notify(128553)
			currSize = 0
			batch = txn.NewBatch()
			modifiedDescs = make([]catalog.MutableDescriptor, 0)
		} else {
			__antithesis_instrumentation__.Notify(128559)
		}
	}
	__antithesis_instrumentation__.Notify(128536)
	for _, modified := range modifiedDescs {
		__antithesis_instrumentation__.Notify(128560)
		err := descriptors.WriteDescToBatch(
			ctx, false, modified, batch,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(128561)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128562)
		}
	}
	__antithesis_instrumentation__.Notify(128537)
	return txn.Run(ctx, batch)
}
