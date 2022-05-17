package descidgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

func GenerateUniqueDescID(ctx context.Context, db *kv.DB, codec keys.SQLCodec) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(251470)

	newVal, err := kv.IncrementValRetryable(ctx, db, codec.DescIDSequenceKey(), 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(251472)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(251473)
	}
	__antithesis_instrumentation__.Notify(251471)
	return descpb.ID(newVal - 1), nil
}

func PeekNextUniqueDescID(ctx context.Context, db *kv.DB, codec keys.SQLCodec) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(251474)
	v, err := db.Get(ctx, codec.DescIDSequenceKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(251476)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(251477)
	}
	__antithesis_instrumentation__.Notify(251475)
	return descpb.ID(v.ValueInt()), nil
}
