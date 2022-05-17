package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.Merge, DefaultDeclareKeys, Merge)
}

func Merge(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97043)
	args := cArgs.Args.(*roachpb.MergeRequest)
	h := cArgs.Header

	return result.Result{}, storage.MVCCMerge(ctx, readWriter, cArgs.Stats, args.Key, h.Timestamp, args.Value)
}
