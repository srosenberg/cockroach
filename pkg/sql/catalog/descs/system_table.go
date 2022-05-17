package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
)

type systemTableIDResolver struct {
	collectionFactory *CollectionFactory
	ie                sqlutil.InternalExecutor
	db                *kv.DB
}

var _ catalog.SystemTableIDResolver = (*systemTableIDResolver)(nil)

func MakeSystemTableIDResolver(
	collectionFactory *CollectionFactory, ie sqlutil.InternalExecutor, db *kv.DB,
) catalog.SystemTableIDResolver {
	__antithesis_instrumentation__.Notify(264907)
	return &systemTableIDResolver{
		collectionFactory: collectionFactory,
		ie:                ie,
		db:                db,
	}
}

func (r *systemTableIDResolver) LookupSystemTableID(
	ctx context.Context, tableName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(264908)

	var id descpb.ID
	if err := r.collectionFactory.Txn(ctx, r.ie, r.db, func(
		ctx context.Context, txn *kv.Txn, descriptors *Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(264910)
		id, err = descriptors.kv.lookupName(
			ctx, txn, nil, keys.SystemDatabaseID, keys.PublicSchemaID, tableName,
		)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(264911)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(264912)
	}
	__antithesis_instrumentation__.Notify(264909)
	return id, nil
}
