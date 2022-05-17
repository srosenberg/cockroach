package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.Increment, DefaultDeclareIsolatedKeys, Increment)
}

func Increment(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96962)
	args := cArgs.Args.(*roachpb.IncrementRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.IncrementResponse)

	newVal, err := storage.MVCCIncrement(ctx, readWriter, cArgs.Stats, args.Key, h.Timestamp, h.Txn, args.Increment)
	reply.NewValue = newVal

	return result.FromAcquiredLocks(h.Txn, args.Key), err
}
