package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.InitPut, DefaultDeclareIsolatedKeys, InitPut)
}

func InitPut(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96963)
	args := cArgs.Args.(*roachpb.InitPutRequest)
	h := cArgs.Header

	var err error
	if args.Blind {
		__antithesis_instrumentation__.Notify(96965)
		err = storage.MVCCBlindInitPut(ctx, readWriter, cArgs.Stats, args.Key, h.Timestamp, args.Value, args.FailOnTombstones, h.Txn)
	} else {
		__antithesis_instrumentation__.Notify(96966)
		err = storage.MVCCInitPut(ctx, readWriter, cArgs.Stats, args.Key, h.Timestamp, args.Value, args.FailOnTombstones, h.Txn)
	}
	__antithesis_instrumentation__.Notify(96964)

	return result.FromAcquiredLocks(h.Txn, args.Key), err
}
