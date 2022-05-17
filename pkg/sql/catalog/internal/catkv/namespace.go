package catkv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
)

func LookupIDs(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, nameInfos []descpb.NameInfo,
) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(265812)
	return lookupIDs(ctx, txn, catalogQuerier{codec: codec}, nameInfos)
}

func LookupID(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(265813)
	nameInfo := descpb.NameInfo{ParentID: parentID, ParentSchemaID: parentSchemaID, Name: name}
	ids, err := LookupIDs(ctx, txn, codec, []descpb.NameInfo{nameInfo})
	if err != nil {
		__antithesis_instrumentation__.Notify(265815)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(265816)
	}
	__antithesis_instrumentation__.Notify(265814)
	return ids[0], nil
}

func GetAllDatabaseDescriptorIDs(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec,
) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(265817)
	cq := catalogQuerier{codec: codec}
	return cq.query(ctx, txn, func(codec keys.SQLCodec, b *kv.Batch) {
		__antithesis_instrumentation__.Notify(265818)
		b.Header.MaxSpanRequestKeys = 0
		prefix := catalogkeys.MakeDatabaseNameKey(codec, "")
		b.Scan(prefix, prefix.PrefixEnd())
	})
}
