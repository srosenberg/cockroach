package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(roachpb.Refresh, DefaultDeclareKeys, Refresh)
}

func Refresh(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97315)
	args := cArgs.Args.(*roachpb.RefreshRequest)
	h := cArgs.Header

	if h.Txn == nil {
		__antithesis_instrumentation__.Notify(97321)
		return result.Result{}, errors.AssertionFailedf("no transaction specified to %s", args.Method())
	} else {
		__antithesis_instrumentation__.Notify(97322)
	}
	__antithesis_instrumentation__.Notify(97316)

	if h.Timestamp != h.Txn.WriteTimestamp {
		__antithesis_instrumentation__.Notify(97323)

		log.Fatalf(ctx, "expected provisional commit ts %s == read ts %s. txn: %s", h.Timestamp,
			h.Txn.WriteTimestamp, h.Txn)
	} else {
		__antithesis_instrumentation__.Notify(97324)
	}
	__antithesis_instrumentation__.Notify(97317)
	refreshTo := h.Timestamp

	refreshFrom := args.RefreshFrom
	if refreshFrom.IsEmpty() {
		__antithesis_instrumentation__.Notify(97325)
		return result.Result{}, errors.AssertionFailedf("empty RefreshFrom: %s", args)
	} else {
		__antithesis_instrumentation__.Notify(97326)
	}
	__antithesis_instrumentation__.Notify(97318)

	log.VEventf(ctx, 2, "refresh %s @[%s-%s]", args.Span(), refreshFrom, refreshTo)
	val, intent, err := storage.MVCCGet(ctx, reader, args.Key, refreshTo, storage.MVCCGetOptions{
		Inconsistent: true,
		Tombstones:   true,
	})

	if err != nil {
		__antithesis_instrumentation__.Notify(97327)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97328)
		if val != nil {
			__antithesis_instrumentation__.Notify(97329)
			if ts := val.Timestamp; refreshFrom.Less(ts) {
				__antithesis_instrumentation__.Notify(97330)
				return result.Result{},
					roachpb.NewRefreshFailedError(roachpb.RefreshFailedError_REASON_COMMITTED_VALUE, args.Key, ts)
			} else {
				__antithesis_instrumentation__.Notify(97331)
			}
		} else {
			__antithesis_instrumentation__.Notify(97332)
		}
	}
	__antithesis_instrumentation__.Notify(97319)

	if intent != nil && func() bool {
		__antithesis_instrumentation__.Notify(97333)
		return intent.Txn.ID != h.Txn.ID == true
	}() == true {
		__antithesis_instrumentation__.Notify(97334)
		return result.Result{}, roachpb.NewRefreshFailedError(roachpb.RefreshFailedError_REASON_INTENT,
			intent.Key, intent.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97335)
	}
	__antithesis_instrumentation__.Notify(97320)

	return result.Result{}, nil
}
