package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
)

func init() {

	RegisterReadWriteCommand(roachpb.AddSSTable, declareKeysAddSSTable, EvalAddSSTable)
}

func declareKeysAddSSTable(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(96311)
	DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

var AddSSTableRewriteConcurrency = settings.RegisterIntSetting(
	settings.SystemOnly,
	"kv.bulk_io_write.sst_rewrite_concurrency.per_call",
	"concurrency to use when rewriting sstable timestamps by block, or 0 to use a loop",
	int64(util.ConstantWithMetamorphicTestRange("addsst-rewrite-concurrency", 0, 0, 16)),
	settings.NonNegativeInt,
)

var AddSSTableRequireAtRequestTimestamp = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.bulk_io_write.sst_require_at_request_timestamp.enabled",
	"rejects addsstable requests that don't write at the request timestamp",
	false,
)

var addSSTableCapacityRemainingLimit = settings.RegisterFloatSetting(
	settings.SystemOnly,
	"kv.bulk_io_write.min_capacity_remaining_fraction",
	"remaining store capacity fraction below which an addsstable request is rejected",
	0.05,
)

var forceRewrite = util.ConstantWithMetamorphicTestBool("addsst-rewrite-forced", false)

func EvalAddSSTable(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96312)
	args := cArgs.Args.(*roachpb.AddSSTableRequest)
	h := cArgs.Header
	ms := cArgs.Stats
	start, end := storage.MVCCKey{Key: args.Key}, storage.MVCCKey{Key: args.EndKey}
	sst := args.Data
	sstToReqTS := args.SSTTimestampToRequestTimestamp

	var span *tracing.Span
	var err error
	ctx, span = tracing.ChildSpan(ctx, fmt.Sprintf("AddSSTable [%s,%s)", start.Key, end.Key))
	defer span.Finish()
	log.Eventf(ctx, "evaluating AddSSTable [%s,%s)", start.Key, end.Key)

	if min := addSSTableCapacityRemainingLimit.Get(&cArgs.EvalCtx.ClusterSettings().SV); min > 0 {
		__antithesis_instrumentation__.Notify(96326)
		cap, err := cArgs.EvalCtx.GetEngineCapacity()
		if err != nil {
			__antithesis_instrumentation__.Notify(96328)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96329)
		}
		__antithesis_instrumentation__.Notify(96327)
		if remaining := float64(cap.Available) / float64(cap.Capacity); remaining < min {
			__antithesis_instrumentation__.Notify(96330)
			return result.Result{}, &roachpb.InsufficientSpaceError{
				StoreID:   cArgs.EvalCtx.StoreID(),
				Op:        "ingest data",
				Available: cap.Available,
				Capacity:  cap.Capacity,
				Required:  min,
			}
		} else {
			__antithesis_instrumentation__.Notify(96331)
		}
	} else {
		__antithesis_instrumentation__.Notify(96332)
	}
	__antithesis_instrumentation__.Notify(96313)

	if cArgs.EvalCtx.ClusterSettings().Version.IsActive(ctx, clusterversion.MVCCAddSSTable) && func() bool {
		__antithesis_instrumentation__.Notify(96333)
		return AddSSTableRequireAtRequestTimestamp.Get(&cArgs.EvalCtx.ClusterSettings().SV) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(96334)
		return sstToReqTS.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(96335)
		return result.Result{}, errors.AssertionFailedf(
			"AddSSTable requests must set SSTTimestampToRequestTimestamp")
	} else {
		__antithesis_instrumentation__.Notify(96336)
	}
	__antithesis_instrumentation__.Notify(96314)

	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(96337)
		if err := assertSSTContents(sst, sstToReqTS, args.MVCCStats); err != nil {
			__antithesis_instrumentation__.Notify(96338)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96339)
		}
	} else {
		__antithesis_instrumentation__.Notify(96340)
	}
	__antithesis_instrumentation__.Notify(96315)

	if sstToReqTS.IsSet() && func() bool {
		__antithesis_instrumentation__.Notify(96341)
		return (h.Timestamp != sstToReqTS || func() bool {
			__antithesis_instrumentation__.Notify(96342)
			return forceRewrite == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(96343)
		st := cArgs.EvalCtx.ClusterSettings()

		conc := int(AddSSTableRewriteConcurrency.Get(&cArgs.EvalCtx.ClusterSettings().SV))
		sst, err = storage.UpdateSSTTimestamps(ctx, st, sst, sstToReqTS, h.Timestamp, conc)
		if err != nil {
			__antithesis_instrumentation__.Notify(96344)
			return result.Result{}, errors.Wrap(err, "updating SST timestamps")
		} else {
			__antithesis_instrumentation__.Notify(96345)
		}
	} else {
		__antithesis_instrumentation__.Notify(96346)
	}
	__antithesis_instrumentation__.Notify(96316)

	var statsDelta enginepb.MVCCStats
	maxIntents := storage.MaxIntentsPerWriteIntentError.Get(&cArgs.EvalCtx.ClusterSettings().SV)
	checkConflicts := args.DisallowConflicts || func() bool {
		__antithesis_instrumentation__.Notify(96347)
		return args.DisallowShadowing == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(96348)
		return !args.DisallowShadowingBelow.IsEmpty() == true
	}() == true
	if checkConflicts {
		__antithesis_instrumentation__.Notify(96349)

		statsDelta, err = storage.CheckSSTConflicts(ctx, sst, readWriter, start, end,
			args.DisallowShadowing, args.DisallowShadowingBelow, maxIntents)
		if err != nil {
			__antithesis_instrumentation__.Notify(96350)
			return result.Result{}, errors.Wrap(err, "checking for key collisions")
		} else {
			__antithesis_instrumentation__.Notify(96351)
		}

	} else {
		__antithesis_instrumentation__.Notify(96352)

		intents, err := storage.ScanIntents(ctx, readWriter, start.Key, end.Key, maxIntents, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(96353)
			return result.Result{}, errors.Wrap(err, "scanning intents")
		} else {
			__antithesis_instrumentation__.Notify(96354)
			if len(intents) > 0 {
				__antithesis_instrumentation__.Notify(96355)
				return result.Result{}, &roachpb.WriteIntentError{Intents: intents}
			} else {
				__antithesis_instrumentation__.Notify(96356)
			}
		}
	}
	__antithesis_instrumentation__.Notify(96317)

	sstIter, err := storage.NewMemSSTIterator(sst, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(96357)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96358)
	}
	__antithesis_instrumentation__.Notify(96318)
	defer sstIter.Close()

	sstIter.SeekGE(storage.MVCCKey{Key: keys.MinKey})
	ok, err := sstIter.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(96359)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96360)
		if ok {
			__antithesis_instrumentation__.Notify(96361)
			if unsafeKey := sstIter.UnsafeKey(); unsafeKey.Less(start) {
				__antithesis_instrumentation__.Notify(96362)
				return result.Result{}, errors.Errorf("first key %s not in request range [%s,%s)",
					unsafeKey.Key, start.Key, end.Key)
			} else {
				__antithesis_instrumentation__.Notify(96363)
			}
		} else {
			__antithesis_instrumentation__.Notify(96364)
		}
	}
	__antithesis_instrumentation__.Notify(96319)

	var stats enginepb.MVCCStats
	if args.MVCCStats != nil {
		__antithesis_instrumentation__.Notify(96365)
		stats = *args.MVCCStats
	} else {
		__antithesis_instrumentation__.Notify(96366)
		log.VEventf(ctx, 2, "computing MVCCStats for SSTable [%s,%s)", start.Key, end.Key)
		stats, err = storage.ComputeStatsForRange(sstIter, start.Key, end.Key, h.Timestamp.WallTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(96367)
			return result.Result{}, errors.Wrap(err, "computing SSTable MVCC stats")
		} else {
			__antithesis_instrumentation__.Notify(96368)
		}
	}
	__antithesis_instrumentation__.Notify(96320)

	sstIter.SeekGE(end)
	ok, err = sstIter.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(96369)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96370)
		if ok {
			__antithesis_instrumentation__.Notify(96371)
			return result.Result{}, errors.Errorf("last key %s not in request range [%s,%s)",
				sstIter.UnsafeKey(), start.Key, end.Key)
		} else {
			__antithesis_instrumentation__.Notify(96372)
		}
	}
	__antithesis_instrumentation__.Notify(96321)

	if checkConflicts {
		__antithesis_instrumentation__.Notify(96373)
		stats.Add(statsDelta)
		stats.ContainsEstimates = 0
	} else {
		__antithesis_instrumentation__.Notify(96374)
		stats.ContainsEstimates++
	}
	__antithesis_instrumentation__.Notify(96322)

	ms.Add(stats)

	var mvccHistoryMutation *kvserverpb.ReplicatedEvalResult_MVCCHistoryMutation
	if sstToReqTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(96375)
		mvccHistoryMutation = &kvserverpb.ReplicatedEvalResult_MVCCHistoryMutation{
			Spans: []roachpb.Span{{Key: start.Key, EndKey: end.Key}},
		}
	} else {
		__antithesis_instrumentation__.Notify(96376)
	}
	__antithesis_instrumentation__.Notify(96323)

	reply := resp.(*roachpb.AddSSTableResponse)
	reply.RangeSpan = cArgs.EvalCtx.Desc().KeySpan().AsRawSpanWithNoLocals()
	reply.AvailableBytes = cArgs.EvalCtx.GetMaxBytes() - cArgs.EvalCtx.GetMVCCStats().Total() - stats.Total()

	if args.ReturnFollowingLikelyNonEmptySpanStart {
		__antithesis_instrumentation__.Notify(96377)
		existingIter := spanset.DisableReaderAssertions(readWriter).NewMVCCIterator(
			storage.MVCCKeyIterKind,
			storage.IterOptions{UpperBound: reply.RangeSpan.EndKey},
		)
		defer existingIter.Close()
		existingIter.SeekGE(end)
		ok, err = existingIter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(96378)
			return result.Result{}, errors.Wrap(err, "error while searching for non-empty span start")
		} else {
			__antithesis_instrumentation__.Notify(96379)
			if ok {
				__antithesis_instrumentation__.Notify(96380)
				reply.FollowingLikelyNonEmptySpanStart = existingIter.Key().Key
			} else {
				__antithesis_instrumentation__.Notify(96381)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(96382)
	}
	__antithesis_instrumentation__.Notify(96324)

	if args.IngestAsWrites {
		__antithesis_instrumentation__.Notify(96383)
		span.RecordStructured(&types.StringValue{Value: fmt.Sprintf("ingesting SST (%d keys/%d bytes) via regular write batch", stats.KeyCount, len(sst))})
		log.VEventf(ctx, 2, "ingesting SST (%d keys/%d bytes) via regular write batch", stats.KeyCount, len(sst))
		sstIter.SeekGE(storage.MVCCKey{Key: keys.MinKey})
		for {
			__antithesis_instrumentation__.Notify(96385)
			ok, err := sstIter.Valid()
			if err != nil {
				__antithesis_instrumentation__.Notify(96389)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96390)
				if !ok {
					__antithesis_instrumentation__.Notify(96391)
					break
				} else {
					__antithesis_instrumentation__.Notify(96392)
				}
			}
			__antithesis_instrumentation__.Notify(96386)

			k := sstIter.UnsafeKey()
			if k.Timestamp.IsEmpty() {
				__antithesis_instrumentation__.Notify(96393)
				if err := readWriter.PutUnversioned(k.Key, sstIter.UnsafeValue()); err != nil {
					__antithesis_instrumentation__.Notify(96394)
					return result.Result{}, err
				} else {
					__antithesis_instrumentation__.Notify(96395)
				}
			} else {
				__antithesis_instrumentation__.Notify(96396)
				if err := readWriter.PutMVCC(k, sstIter.UnsafeValue()); err != nil {
					__antithesis_instrumentation__.Notify(96397)
					return result.Result{}, err
				} else {
					__antithesis_instrumentation__.Notify(96398)
				}
			}
			__antithesis_instrumentation__.Notify(96387)

			if sstToReqTS.IsSet() {
				__antithesis_instrumentation__.Notify(96399)
				readWriter.LogLogicalOp(storage.MVCCWriteValueOpType, storage.MVCCLogicalOpDetails{
					Key:       k.Key,
					Timestamp: k.Timestamp,
				})
			} else {
				__antithesis_instrumentation__.Notify(96400)
			}
			__antithesis_instrumentation__.Notify(96388)
			sstIter.Next()
		}
		__antithesis_instrumentation__.Notify(96384)
		return result.Result{
			Replicated: kvserverpb.ReplicatedEvalResult{
				MVCCHistoryMutation: mvccHistoryMutation,
			},
			Local: result.LocalResult{
				Metrics: &result.Metrics{
					AddSSTableAsWrites: 1,
				},
			},
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(96401)
	}
	__antithesis_instrumentation__.Notify(96325)

	return result.Result{
		Replicated: kvserverpb.ReplicatedEvalResult{
			AddSSTable: &kvserverpb.ReplicatedEvalResult_AddSSTable{
				Data:             sst,
				CRC32:            util.CRC32(sst),
				Span:             roachpb.Span{Key: start.Key, EndKey: end.Key},
				AtWriteTimestamp: sstToReqTS.IsSet(),
			},
			MVCCHistoryMutation: mvccHistoryMutation,
		},
	}, nil
}

func assertSSTContents(sst []byte, sstTimestamp hlc.Timestamp, stats *enginepb.MVCCStats) error {
	__antithesis_instrumentation__.Notify(96402)
	iter, err := storage.NewMemSSTIterator(sst, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(96406)
		return err
	} else {
		__antithesis_instrumentation__.Notify(96407)
	}
	__antithesis_instrumentation__.Notify(96403)
	defer iter.Close()

	iter.SeekGE(storage.MVCCKey{Key: keys.MinKey})
	for {
		__antithesis_instrumentation__.Notify(96408)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(96414)
			return err
		} else {
			__antithesis_instrumentation__.Notify(96415)
		}
		__antithesis_instrumentation__.Notify(96409)
		if !ok {
			__antithesis_instrumentation__.Notify(96416)
			break
		} else {
			__antithesis_instrumentation__.Notify(96417)
		}
		__antithesis_instrumentation__.Notify(96410)

		key, value := iter.UnsafeKey(), iter.UnsafeValue()
		if key.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(96418)
			return errors.AssertionFailedf("SST contains inline value or intent for key %s", key)
		} else {
			__antithesis_instrumentation__.Notify(96419)
		}
		__antithesis_instrumentation__.Notify(96411)
		if len(value) == 0 {
			__antithesis_instrumentation__.Notify(96420)
			return errors.AssertionFailedf("SST contains tombstone for key %s", key)
		} else {
			__antithesis_instrumentation__.Notify(96421)
		}
		__antithesis_instrumentation__.Notify(96412)
		if sstTimestamp.IsSet() && func() bool {
			__antithesis_instrumentation__.Notify(96422)
			return key.Timestamp != sstTimestamp == true
		}() == true {
			__antithesis_instrumentation__.Notify(96423)
			return errors.AssertionFailedf("SST has unexpected timestamp %s (expected %s) for key %s",
				key.Timestamp, sstTimestamp, key.Key)
		} else {
			__antithesis_instrumentation__.Notify(96424)
		}
		__antithesis_instrumentation__.Notify(96413)
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(96404)

	if stats != nil {
		__antithesis_instrumentation__.Notify(96425)
		given := *stats
		actual, err := storage.ComputeStatsForRange(
			iter, keys.MinKey, keys.MaxKey, given.LastUpdateNanos)
		if err != nil {
			__antithesis_instrumentation__.Notify(96427)
			return errors.Wrap(err, "failed to compare stats: %w")
		} else {
			__antithesis_instrumentation__.Notify(96428)
		}
		__antithesis_instrumentation__.Notify(96426)
		if !given.Equal(actual) {
			__antithesis_instrumentation__.Notify(96429)
			return errors.AssertionFailedf("SST stats are incorrect: diff(given, actual) = %s",
				pretty.Diff(given, actual))
		} else {
			__antithesis_instrumentation__.Notify(96430)
		}
	} else {
		__antithesis_instrumentation__.Notify(96431)
	}
	__antithesis_instrumentation__.Notify(96405)

	return nil
}
