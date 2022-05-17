package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func insertMissingPublicSchemaNamespaceEntry(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128456)

	query := `
  SELECT id
    FROM system.namespace
   WHERE id
         NOT IN (
                SELECT ns_db.id
                  FROM system.namespace AS ns_db
                       INNER JOIN system.namespace
                            AS ns_sc ON (
                                        ns_db.id
                                        = ns_sc."parentID"
                                    )
                 WHERE ns_db."parentSchemaID" = 0
                   AND ns_db."parentID" = 0
                   AND ns_sc."parentSchemaID" = 0
                   AND ns_sc.name = 'public'
                   AND ns_sc.id = 29
            )
     AND "parentID" = 0
ORDER BY id ASC;
`
	rows, err := d.InternalExecutor.QueryIterator(
		ctx, "get_databases_without_public_schema_namespace_entry", nil, query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(128459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128460)
	}
	__antithesis_instrumentation__.Notify(128457)
	var databaseIDs []descpb.ID
	for ok, err := rows.Next(ctx); ok; ok, err = rows.Next(ctx) {
		__antithesis_instrumentation__.Notify(128461)
		if err != nil {
			__antithesis_instrumentation__.Notify(128463)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128464)
		}
		__antithesis_instrumentation__.Notify(128462)
		id := descpb.ID(tree.MustBeDInt(rows.Cur()[0]))
		databaseIDs = append(databaseIDs, id)
	}
	__antithesis_instrumentation__.Notify(128458)

	return d.CollectionFactory.Txn(ctx, d.InternalExecutor, d.DB, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(128465)
		b := txn.NewBatch()
		for _, dbID := range databaseIDs {
			__antithesis_instrumentation__.Notify(128467)
			publicSchemaKey := catalogkeys.MakeSchemaNameKey(d.Codec, dbID, tree.PublicSchema)
			b.Put(publicSchemaKey, keys.PublicSchemaID)
		}
		__antithesis_instrumentation__.Notify(128466)
		return txn.Run(ctx, b)
	})
}
