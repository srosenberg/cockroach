package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func createSystemTable(
	ctx context.Context, db *kv.DB, codec keys.SQLCodec, desc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(128393)

	return db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(128394)
		tKey := catalogkeys.EncodeNameKey(codec, desc)

		if desc.GetID() == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(128396)

			got, err := txn.Get(ctx, tKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(128400)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128401)
			}
			__antithesis_instrumentation__.Notify(128397)
			if got.Value.IsPresent() {
				__antithesis_instrumentation__.Notify(128402)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(128403)
			}
			__antithesis_instrumentation__.Notify(128398)
			id, err := descidgen.GenerateUniqueDescID(ctx, db, codec)
			if err != nil {
				__antithesis_instrumentation__.Notify(128404)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128405)
			}
			__antithesis_instrumentation__.Notify(128399)
			mut := desc.NewBuilder().BuildCreatedMutable().(*tabledesc.Mutable)
			mut.ID = id
			desc = mut
		} else {
			__antithesis_instrumentation__.Notify(128406)
		}
		__antithesis_instrumentation__.Notify(128395)

		b := txn.NewBatch()
		b.CPut(tKey, desc.GetID(), nil)
		b.CPut(catalogkeys.MakeDescMetadataKey(codec, desc.GetID()), desc.DescriptorProto(), nil)
		return txn.Run(ctx, b)
	})
}

func runPostDeserializationChangesOnAllDescriptors(
	ctx context.Context, d migration.TenantDeps,
) error {
	__antithesis_instrumentation__.Notify(128407)

	maybeUpgradeDescriptors := func(
		ctx context.Context, d migration.TenantDeps, toUpgrade []descpb.ID,
	) error {
		__antithesis_instrumentation__.Notify(128415)
		return d.CollectionFactory.Txn(ctx, d.InternalExecutor, d.DB, func(
			ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(128416)
			descs, err := descriptors.GetMutableDescriptorsByID(ctx, txn, toUpgrade...)
			if err != nil {
				__antithesis_instrumentation__.Notify(128419)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128420)
			}
			__antithesis_instrumentation__.Notify(128417)
			batch := txn.NewBatch()
			for _, desc := range descs {
				__antithesis_instrumentation__.Notify(128421)
				if !desc.GetPostDeserializationChanges().HasChanges() {
					__antithesis_instrumentation__.Notify(128423)
					continue
				} else {
					__antithesis_instrumentation__.Notify(128424)
				}
				__antithesis_instrumentation__.Notify(128422)
				if err := descriptors.WriteDescToBatch(
					ctx, false, desc, batch,
				); err != nil {
					__antithesis_instrumentation__.Notify(128425)
					return err
				} else {
					__antithesis_instrumentation__.Notify(128426)
				}
			}
			__antithesis_instrumentation__.Notify(128418)
			return txn.Run(ctx, batch)
		})
	}
	__antithesis_instrumentation__.Notify(128408)

	query := `SELECT id, length(descriptor) FROM system.descriptor ORDER BY id DESC`
	rows, err := d.InternalExecutor.QueryIterator(
		ctx, "retrieve-descriptors-for-upgrade", nil, query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(128427)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128428)
	}
	__antithesis_instrumentation__.Notify(128409)
	defer func() { __antithesis_instrumentation__.Notify(128429); _ = rows.Close() }()
	__antithesis_instrumentation__.Notify(128410)
	var toUpgrade []descpb.ID
	var curBatchBytes int
	const maxBatchSize = 1 << 19
	ok, err := rows.Next(ctx)
	for ; ok && func() bool {
		__antithesis_instrumentation__.Notify(128430)
		return err == nil == true
	}() == true; ok, err = rows.Next(ctx) {
		__antithesis_instrumentation__.Notify(128431)
		datums := rows.Cur()
		id := tree.MustBeDInt(datums[0])
		size := tree.MustBeDInt(datums[1])
		if curBatchBytes+int(size) > maxBatchSize && func() bool {
			__antithesis_instrumentation__.Notify(128433)
			return curBatchBytes > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(128434)
			if err := maybeUpgradeDescriptors(ctx, d, toUpgrade); err != nil {
				__antithesis_instrumentation__.Notify(128436)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128437)
			}
			__antithesis_instrumentation__.Notify(128435)
			toUpgrade = toUpgrade[:0]
		} else {
			__antithesis_instrumentation__.Notify(128438)
		}
		__antithesis_instrumentation__.Notify(128432)
		curBatchBytes += int(size)
		toUpgrade = append(toUpgrade, descpb.ID(id))
	}
	__antithesis_instrumentation__.Notify(128411)
	if err != nil {
		__antithesis_instrumentation__.Notify(128439)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128440)
	}
	__antithesis_instrumentation__.Notify(128412)
	if err := rows.Close(); err != nil {
		__antithesis_instrumentation__.Notify(128441)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128442)
	}
	__antithesis_instrumentation__.Notify(128413)
	if len(toUpgrade) == 0 {
		__antithesis_instrumentation__.Notify(128443)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128444)
	}
	__antithesis_instrumentation__.Notify(128414)
	return maybeUpgradeDescriptors(ctx, d, toUpgrade)
}
