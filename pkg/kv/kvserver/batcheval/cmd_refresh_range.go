package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var refreshRangeTBIEnabled = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.refresh_range.time_bound_iterators.enabled",
	"use time-bound iterators when performing ranged transaction refreshes",
	util.ConstantWithMetamorphicTestBool("kv.refresh_range.time_bound_iterators.enabled", true),
)

func init() {
	RegisterReadOnlyCommand(roachpb.RefreshRange, DefaultDeclareKeys, RefreshRange)
}

func RefreshRange(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97336)
	args := cArgs.Args.(*roachpb.RefreshRangeRequest)
	h := cArgs.Header

	if h.Txn == nil {
		__antithesis_instrumentation__.Notify(97340)
		return result.Result{}, errors.AssertionFailedf("no transaction specified to %s", args.Method())
	} else {
		__antithesis_instrumentation__.Notify(97341)
	}
	__antithesis_instrumentation__.Notify(97337)

	if h.Timestamp != h.Txn.WriteTimestamp {
		__antithesis_instrumentation__.Notify(97342)

		log.Fatalf(ctx, "expected provisional commit ts %s == read ts %s. txn: %s", h.Timestamp,
			h.Txn.WriteTimestamp, h.Txn)
	} else {
		__antithesis_instrumentation__.Notify(97343)
	}
	__antithesis_instrumentation__.Notify(97338)
	refreshTo := h.Timestamp

	refreshFrom := args.RefreshFrom
	if refreshFrom.IsEmpty() {
		__antithesis_instrumentation__.Notify(97344)
		return result.Result{}, errors.AssertionFailedf("empty RefreshFrom: %s", args)
	} else {
		__antithesis_instrumentation__.Notify(97345)
	}
	__antithesis_instrumentation__.Notify(97339)

	log.VEventf(ctx, 2, "refresh %s @[%s-%s]", args.Span(), refreshFrom, refreshTo)
	tbi := refreshRangeTBIEnabled.Get(&cArgs.EvalCtx.ClusterSettings().SV)
	return result.Result{}, refreshRange(reader, tbi, args.Span(), refreshFrom, refreshTo, h.Txn.ID)
}

func refreshRange(
	reader storage.Reader,
	timeBoundIterator bool,
	span roachpb.Span,
	refreshFrom, refreshTo hlc.Timestamp,
	txnID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(97346)

	iter := storage.NewMVCCIncrementalIterator(reader, storage.MVCCIncrementalIterOptions{
		EnableTimeBoundIteratorOptimization: timeBoundIterator,
		EndKey:                              span.EndKey,
		StartTime:                           refreshFrom,
		EndTime:                             refreshTo,
		IntentPolicy:                        storage.MVCCIncrementalIterIntentPolicyEmit,
	})
	defer iter.Close()

	var meta enginepb.MVCCMetadata
	iter.SeekGE(storage.MakeMVCCMetadataKey(span.Key))
	for {
		__antithesis_instrumentation__.Notify(97348)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(97351)
			return err
		} else {
			__antithesis_instrumentation__.Notify(97352)
			if !ok {
				__antithesis_instrumentation__.Notify(97353)
				break
			} else {
				__antithesis_instrumentation__.Notify(97354)
			}
		}
		__antithesis_instrumentation__.Notify(97349)

		key := iter.Key()
		if !key.IsValue() {
			__antithesis_instrumentation__.Notify(97355)

			if err := protoutil.Unmarshal(iter.UnsafeValue(), &meta); err != nil {
				__antithesis_instrumentation__.Notify(97359)
				return errors.Wrapf(err, "unmarshaling mvcc meta: %v", key)
			} else {
				__antithesis_instrumentation__.Notify(97360)
			}
			__antithesis_instrumentation__.Notify(97356)
			if meta.IsInline() {
				__antithesis_instrumentation__.Notify(97361)

				iter.Next()
				continue
			} else {
				__antithesis_instrumentation__.Notify(97362)
			}
			__antithesis_instrumentation__.Notify(97357)
			if meta.Txn.ID == txnID {
				__antithesis_instrumentation__.Notify(97363)

				iter.Next()
				if ok, err := iter.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(97366)
					return errors.Wrap(err, "iterating to provisional value for intent")
				} else {
					__antithesis_instrumentation__.Notify(97367)
					if !ok {
						__antithesis_instrumentation__.Notify(97368)
						return errors.Errorf("expected provisional value for intent")
					} else {
						__antithesis_instrumentation__.Notify(97369)
					}
				}
				__antithesis_instrumentation__.Notify(97364)
				if !meta.Timestamp.ToTimestamp().EqOrdering(iter.UnsafeKey().Timestamp) {
					__antithesis_instrumentation__.Notify(97370)
					return errors.Errorf("expected provisional value for intent with ts %s, found %s",
						meta.Timestamp, iter.UnsafeKey().Timestamp)
				} else {
					__antithesis_instrumentation__.Notify(97371)
				}
				__antithesis_instrumentation__.Notify(97365)
				iter.Next()
				continue
			} else {
				__antithesis_instrumentation__.Notify(97372)
			}
			__antithesis_instrumentation__.Notify(97358)
			return roachpb.NewRefreshFailedError(roachpb.RefreshFailedError_REASON_INTENT,
				key.Key, meta.Txn.WriteTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(97373)
		}
		__antithesis_instrumentation__.Notify(97350)

		return roachpb.NewRefreshFailedError(roachpb.RefreshFailedError_REASON_COMMITTED_VALUE,
			key.Key, key.Timestamp)
	}
	__antithesis_instrumentation__.Notify(97347)
	return nil
}
