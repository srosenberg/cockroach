package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.Delete, DefaultDeclareIsolatedKeys, Delete)
}

func Delete(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96496)
	args := cArgs.Args.(*roachpb.DeleteRequest)
	h := cArgs.Header

	err := storage.MVCCDelete(ctx, readWriter, cArgs.Stats, args.Key, h.Timestamp, h.Txn)

	return result.FromAcquiredLocks(h.Txn, args.Key), err
}
