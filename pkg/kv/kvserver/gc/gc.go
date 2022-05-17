// Package gc contains the logic to run scan a range for garbage and issue
// GC requests to remove that garbage.
//
// The Run function is the primary entry point and is called underneath the
// gcQueue in the storage package. It can also be run for debugging.
package gc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const (
	KeyVersionChunkBytes = base.ChunkRaftCommandThresholdBytes
)

var IntentAgeThreshold = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.gc.intent_age_threshold",
	"intents older than this threshold will be resolved when encountered by the MVCC GC queue",
	2*time.Hour,
	func(d time.Duration) error {
		__antithesis_instrumentation__.Notify(101256)
		if d < 2*time.Minute {
			__antithesis_instrumentation__.Notify(101258)
			return errors.New("intent age threshold must be >= 2 minutes")
		} else {
			__antithesis_instrumentation__.Notify(101259)
		}
		__antithesis_instrumentation__.Notify(101257)
		return nil
	},
)

var MaxIntentsPerCleanupBatch = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.gc.intent_cleanup_batch_size",
	"if non zero, gc will split found intents into batches of this size when trying to resolve them",
	5000,
	settings.NonNegativeInt,
)

var MaxIntentKeyBytesPerCleanupBatch = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.gc.intent_cleanup_batch_byte_size",
	"if non zero, gc will split found intents into batches of this size when trying to resolve them",
	1e6,
	settings.NonNegativeInt,
)

func CalculateThreshold(now hlc.Timestamp, gcttl time.Duration) (threshold hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(101260)
	ttlNanos := gcttl.Nanoseconds()
	return now.Add(-ttlNanos, 0)
}

func TimestampForThreshold(threshold hlc.Timestamp, gcttl time.Duration) (ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(101261)
	ttlNanos := gcttl.Nanoseconds()
	return threshold.Add(ttlNanos, 0)
}

type Thresholder interface {
	SetGCThreshold(context.Context, Threshold) error
}

type PureGCer interface {
	GC(context.Context, []roachpb.GCRequest_GCKey) error
}

type GCer interface {
	Thresholder
	PureGCer
}

type NoopGCer struct{}

var _ GCer = NoopGCer{}

func (NoopGCer) SetGCThreshold(context.Context, Threshold) error {
	__antithesis_instrumentation__.Notify(101262)
	return nil
}

func (NoopGCer) GC(context.Context, []roachpb.GCRequest_GCKey) error {
	__antithesis_instrumentation__.Notify(101263)
	return nil
}

type Threshold struct {
	Key hlc.Timestamp
	Txn hlc.Timestamp
}

type Info struct {
	Now hlc.Timestamp

	GCTTL time.Duration

	NumKeysAffected, IntentsConsidered, IntentTxns int

	TransactionSpanTotal int

	TransactionSpanGCAborted, TransactionSpanGCCommitted int
	TransactionSpanGCStaging, TransactionSpanGCPending   int

	AbortSpanTotal int

	AbortSpanConsidered int

	AbortSpanGCNum int

	PushTxn int

	ResolveTotal int

	Threshold hlc.Timestamp

	AffectedVersionsKeyBytes int64

	AffectedVersionsValBytes int64
}

type RunOptions struct {
	IntentAgeThreshold time.Duration

	MaxIntentsPerIntentCleanupBatch int64

	MaxIntentKeyBytesPerIntentCleanupBatch int64

	MaxTxnsPerIntentCleanupBatch int64

	IntentCleanupBatchTimeout time.Duration
}

type CleanupIntentsFunc func(context.Context, []roachpb.Intent) error

type CleanupTxnIntentsAsyncFunc func(context.Context, *roachpb.Transaction) error

func Run(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	snap storage.Reader,
	now, newThreshold hlc.Timestamp,
	options RunOptions,
	gcTTL time.Duration,
	gcer GCer,
	cleanupIntentsFn CleanupIntentsFunc,
	cleanupTxnIntentsAsyncFn CleanupTxnIntentsAsyncFunc,
) (Info, error) {
	__antithesis_instrumentation__.Notify(101264)

	txnExp := now.Add(-kvserverbase.TxnCleanupThreshold.Nanoseconds(), 0)
	if err := gcer.SetGCThreshold(ctx, Threshold{
		Key: newThreshold,
		Txn: txnExp,
	}); err != nil {
		__antithesis_instrumentation__.Notify(101269)
		return Info{}, errors.Wrap(err, "failed to set GC thresholds")
	} else {
		__antithesis_instrumentation__.Notify(101270)
	}
	__antithesis_instrumentation__.Notify(101265)

	info := Info{
		GCTTL:     gcTTL,
		Now:       now,
		Threshold: newThreshold,
	}

	err := processReplicatedKeyRange(ctx, desc, snap, now, newThreshold, options.IntentAgeThreshold, gcer,
		intentBatcherOptions{
			maxIntentsPerIntentCleanupBatch:        options.MaxIntentsPerIntentCleanupBatch,
			maxIntentKeyBytesPerIntentCleanupBatch: options.MaxIntentKeyBytesPerIntentCleanupBatch,
			maxTxnsPerIntentCleanupBatch:           options.MaxTxnsPerIntentCleanupBatch,
			intentCleanupBatchTimeout:              options.IntentCleanupBatchTimeout,
		}, cleanupIntentsFn, &info)
	if err != nil {
		__antithesis_instrumentation__.Notify(101271)
		return Info{}, err
	} else {
		__antithesis_instrumentation__.Notify(101272)
	}
	__antithesis_instrumentation__.Notify(101266)

	if err := processLocalKeyRange(ctx, snap, desc, txnExp, &info, cleanupTxnIntentsAsyncFn, gcer); err != nil {
		__antithesis_instrumentation__.Notify(101273)
		if errors.Is(err, ctx.Err()) {
			__antithesis_instrumentation__.Notify(101275)
			return Info{}, err
		} else {
			__antithesis_instrumentation__.Notify(101276)
		}
		__antithesis_instrumentation__.Notify(101274)
		log.Warningf(ctx, "while gc'ing local key range: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(101277)
	}
	__antithesis_instrumentation__.Notify(101267)

	log.Event(ctx, "processing AbortSpan")
	if err := processAbortSpan(ctx, snap, desc.RangeID, txnExp, &info, gcer); err != nil {
		__antithesis_instrumentation__.Notify(101278)
		if errors.Is(err, ctx.Err()) {
			__antithesis_instrumentation__.Notify(101280)
			return Info{}, err
		} else {
			__antithesis_instrumentation__.Notify(101281)
		}
		__antithesis_instrumentation__.Notify(101279)
		log.Warningf(ctx, "while gc'ing abort span: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(101282)
	}
	__antithesis_instrumentation__.Notify(101268)

	log.Eventf(ctx, "GC'ed keys; stats %+v", info)

	return info, nil
}

func processReplicatedKeyRange(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	snap storage.Reader,
	now hlc.Timestamp,
	threshold hlc.Timestamp,
	intentAgeThreshold time.Duration,
	gcer GCer,
	options intentBatcherOptions,
	cleanupIntentsFn CleanupIntentsFunc,
	info *Info,
) error {
	__antithesis_instrumentation__.Notify(101283)
	var alloc bufalloc.ByteAllocator

	intentExp := now.Add(-intentAgeThreshold.Nanoseconds(), 0)

	batcher := newIntentBatcher(cleanupIntentsFn, options, info)

	handleIntent := func(keyValue *storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(101288)
		meta := &enginepb.MVCCMetadata{}
		if err := protoutil.Unmarshal(keyValue.Value, meta); err != nil {
			__antithesis_instrumentation__.Notify(101291)
			log.Errorf(ctx, "unable to unmarshal MVCC metadata for key %q: %+v", keyValue.Key, err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(101292)
		}
		__antithesis_instrumentation__.Notify(101289)
		if meta.Txn != nil {
			__antithesis_instrumentation__.Notify(101293)

			if meta.Timestamp.ToTimestamp().Less(intentExp) {
				__antithesis_instrumentation__.Notify(101294)
				info.IntentsConsidered++
				if err := batcher.addAndMaybeFlushIntents(ctx, keyValue.Key.Key, meta); err != nil {
					__antithesis_instrumentation__.Notify(101295)
					if errors.Is(err, ctx.Err()) {
						__antithesis_instrumentation__.Notify(101297)
						return err
					} else {
						__antithesis_instrumentation__.Notify(101298)
					}
					__antithesis_instrumentation__.Notify(101296)
					log.Warningf(ctx, "failed to cleanup intents batch: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(101299)
				}
			} else {
				__antithesis_instrumentation__.Notify(101300)
			}
		} else {
			__antithesis_instrumentation__.Notify(101301)
		}
		__antithesis_instrumentation__.Notify(101290)
		return nil
	}
	__antithesis_instrumentation__.Notify(101284)

	var (
		batchGCKeys           []roachpb.GCRequest_GCKey
		batchGCKeysBytes      int64
		haveGarbageForThisKey bool
		gcTimestampForThisKey hlc.Timestamp
		sentBatchForThisKey   bool
	)
	it := makeGCIterator(desc, snap)
	defer it.close()
	for ; ; it.step() {
		__antithesis_instrumentation__.Notify(101302)
		s, ok := it.state()
		if !ok {
			__antithesis_instrumentation__.Notify(101309)
			if it.err != nil {
				__antithesis_instrumentation__.Notify(101311)
				return it.err
			} else {
				__antithesis_instrumentation__.Notify(101312)
			}
			__antithesis_instrumentation__.Notify(101310)
			break
		} else {
			__antithesis_instrumentation__.Notify(101313)
		}
		__antithesis_instrumentation__.Notify(101303)
		if s.curIsNotValue() {
			__antithesis_instrumentation__.Notify(101314)
			continue
		} else {
			__antithesis_instrumentation__.Notify(101315)
		}
		__antithesis_instrumentation__.Notify(101304)
		if s.curIsIntent() {
			__antithesis_instrumentation__.Notify(101316)
			if err := handleIntent(s.next); err != nil {
				__antithesis_instrumentation__.Notify(101318)
				return err
			} else {
				__antithesis_instrumentation__.Notify(101319)
			}
			__antithesis_instrumentation__.Notify(101317)
			continue
		} else {
			__antithesis_instrumentation__.Notify(101320)
		}
		__antithesis_instrumentation__.Notify(101305)
		isNewest := s.curIsNewest()
		if isGarbage(threshold, s.cur, s.next, isNewest) {
			__antithesis_instrumentation__.Notify(101321)
			keyBytes := int64(s.cur.Key.EncodedSize())
			batchGCKeysBytes += keyBytes
			haveGarbageForThisKey = true
			gcTimestampForThisKey = s.cur.Key.Timestamp
			info.AffectedVersionsKeyBytes += keyBytes
			info.AffectedVersionsValBytes += int64(len(s.cur.Value))
		} else {
			__antithesis_instrumentation__.Notify(101322)
		}
		__antithesis_instrumentation__.Notify(101306)
		if affected := isNewest && func() bool {
			__antithesis_instrumentation__.Notify(101323)
			return (sentBatchForThisKey || func() bool {
				__antithesis_instrumentation__.Notify(101324)
				return haveGarbageForThisKey == true
			}() == true) == true
		}() == true; affected {
			__antithesis_instrumentation__.Notify(101325)
			info.NumKeysAffected++
		} else {
			__antithesis_instrumentation__.Notify(101326)
		}
		__antithesis_instrumentation__.Notify(101307)
		shouldSendBatch := batchGCKeysBytes >= KeyVersionChunkBytes
		if shouldSendBatch || func() bool {
			__antithesis_instrumentation__.Notify(101327)
			return (isNewest && func() bool {
				__antithesis_instrumentation__.Notify(101328)
				return haveGarbageForThisKey == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(101329)
			alloc, s.cur.Key.Key = alloc.Copy(s.cur.Key.Key, 0)
			batchGCKeys = append(batchGCKeys, roachpb.GCRequest_GCKey{
				Key:       s.cur.Key.Key,
				Timestamp: gcTimestampForThisKey,
			})
			haveGarbageForThisKey = false
			gcTimestampForThisKey = hlc.Timestamp{}

			sentBatchForThisKey = shouldSendBatch && func() bool {
				__antithesis_instrumentation__.Notify(101330)
				return !isNewest == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(101331)
		}
		__antithesis_instrumentation__.Notify(101308)
		if shouldSendBatch {
			__antithesis_instrumentation__.Notify(101332)
			if err := gcer.GC(ctx, batchGCKeys); err != nil {
				__antithesis_instrumentation__.Notify(101334)
				if errors.Is(err, ctx.Err()) {
					__antithesis_instrumentation__.Notify(101336)
					return err
				} else {
					__antithesis_instrumentation__.Notify(101337)
				}
				__antithesis_instrumentation__.Notify(101335)

				log.Warningf(ctx, "failed to GC a batch of keys: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(101338)
			}
			__antithesis_instrumentation__.Notify(101333)
			batchGCKeys = nil
			batchGCKeysBytes = 0
			alloc = bufalloc.ByteAllocator{}
		} else {
			__antithesis_instrumentation__.Notify(101339)
		}
	}
	__antithesis_instrumentation__.Notify(101285)

	if err := batcher.maybeFlushPendingIntents(ctx); err != nil {
		__antithesis_instrumentation__.Notify(101340)
		if errors.Is(err, ctx.Err()) {
			__antithesis_instrumentation__.Notify(101342)
			return err
		} else {
			__antithesis_instrumentation__.Notify(101343)
		}
		__antithesis_instrumentation__.Notify(101341)
		log.Warningf(ctx, "failed to cleanup intents batch: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(101344)
	}
	__antithesis_instrumentation__.Notify(101286)
	if len(batchGCKeys) > 0 {
		__antithesis_instrumentation__.Notify(101345)
		if err := gcer.GC(ctx, batchGCKeys); err != nil {
			__antithesis_instrumentation__.Notify(101346)
			return err
		} else {
			__antithesis_instrumentation__.Notify(101347)
		}
	} else {
		__antithesis_instrumentation__.Notify(101348)
	}
	__antithesis_instrumentation__.Notify(101287)
	return nil
}

type intentBatcher struct {
	cleanupIntentsFn CleanupIntentsFunc

	options intentBatcherOptions

	pendingTxns          map[uuid.UUID]bool
	pendingIntents       []roachpb.Intent
	collectedIntentBytes int64

	alloc bufalloc.ByteAllocator

	gcStats *Info
}

type intentBatcherOptions struct {
	maxIntentsPerIntentCleanupBatch        int64
	maxIntentKeyBytesPerIntentCleanupBatch int64
	maxTxnsPerIntentCleanupBatch           int64
	intentCleanupBatchTimeout              time.Duration
}

func newIntentBatcher(
	cleanupIntentsFunc CleanupIntentsFunc, options intentBatcherOptions, gcStats *Info,
) intentBatcher {
	__antithesis_instrumentation__.Notify(101349)
	if options.maxIntentsPerIntentCleanupBatch <= 0 {
		__antithesis_instrumentation__.Notify(101353)
		options.maxIntentsPerIntentCleanupBatch = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(101354)
	}
	__antithesis_instrumentation__.Notify(101350)
	if options.maxIntentKeyBytesPerIntentCleanupBatch <= 0 {
		__antithesis_instrumentation__.Notify(101355)
		options.maxIntentKeyBytesPerIntentCleanupBatch = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(101356)
	}
	__antithesis_instrumentation__.Notify(101351)
	if options.maxTxnsPerIntentCleanupBatch <= 0 {
		__antithesis_instrumentation__.Notify(101357)
		options.maxTxnsPerIntentCleanupBatch = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(101358)
	}
	__antithesis_instrumentation__.Notify(101352)
	return intentBatcher{
		cleanupIntentsFn: cleanupIntentsFunc,
		options:          options,
		pendingTxns:      make(map[uuid.UUID]bool),
		gcStats:          gcStats,
	}
}

func (b *intentBatcher) addAndMaybeFlushIntents(
	ctx context.Context, key roachpb.Key, meta *enginepb.MVCCMetadata,
) error {
	__antithesis_instrumentation__.Notify(101359)
	var err error = nil
	txnID := meta.Txn.ID
	_, existingTransaction := b.pendingTxns[txnID]

	if int64(len(b.pendingIntents)) >= b.options.maxIntentsPerIntentCleanupBatch || func() bool {
		__antithesis_instrumentation__.Notify(101361)
		return b.collectedIntentBytes >= b.options.maxIntentKeyBytesPerIntentCleanupBatch == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(101362)
		return (!existingTransaction && func() bool {
			__antithesis_instrumentation__.Notify(101363)
			return int64(len(b.pendingTxns)) >= b.options.maxTxnsPerIntentCleanupBatch == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(101364)
		err = b.maybeFlushPendingIntents(ctx)
	} else {
		__antithesis_instrumentation__.Notify(101365)
	}
	__antithesis_instrumentation__.Notify(101360)

	b.alloc, key = b.alloc.Copy(key, 0)
	b.pendingIntents = append(b.pendingIntents, roachpb.MakeIntent(meta.Txn, key))
	b.collectedIntentBytes += int64(len(key))
	b.pendingTxns[txnID] = true

	return err
}

func (b *intentBatcher) maybeFlushPendingIntents(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(101366)
	if len(b.pendingIntents) == 0 {
		__antithesis_instrumentation__.Notify(101372)

		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(101373)
	}
	__antithesis_instrumentation__.Notify(101367)

	var err error
	cleanupIntentsFn := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(101374)
		return b.cleanupIntentsFn(ctx, b.pendingIntents)
	}
	__antithesis_instrumentation__.Notify(101368)
	if b.options.intentCleanupBatchTimeout > 0 {
		__antithesis_instrumentation__.Notify(101375)
		err = contextutil.RunWithTimeout(
			ctx, "intent GC batch", b.options.intentCleanupBatchTimeout, cleanupIntentsFn)
	} else {
		__antithesis_instrumentation__.Notify(101376)
		err = cleanupIntentsFn(ctx)
	}
	__antithesis_instrumentation__.Notify(101369)
	if err == nil {
		__antithesis_instrumentation__.Notify(101377)

		b.gcStats.IntentTxns += len(b.pendingTxns)

		b.gcStats.PushTxn += len(b.pendingTxns)

		b.gcStats.ResolveTotal += len(b.pendingIntents)
	} else {
		__antithesis_instrumentation__.Notify(101378)
	}
	__antithesis_instrumentation__.Notify(101370)

	for k := range b.pendingTxns {
		__antithesis_instrumentation__.Notify(101379)
		delete(b.pendingTxns, k)
	}
	__antithesis_instrumentation__.Notify(101371)
	b.pendingIntents = b.pendingIntents[:0]
	b.collectedIntentBytes = 0
	return err
}

func isGarbage(threshold hlc.Timestamp, cur, next *storage.MVCCKeyValue, isNewest bool) bool {
	__antithesis_instrumentation__.Notify(101380)

	if belowThreshold := cur.Key.Timestamp.LessEq(threshold); !belowThreshold {
		__antithesis_instrumentation__.Notify(101384)
		return false
	} else {
		__antithesis_instrumentation__.Notify(101385)
	}
	__antithesis_instrumentation__.Notify(101381)
	isDelete := len(cur.Value) == 0
	if isNewest && func() bool {
		__antithesis_instrumentation__.Notify(101386)
		return !isDelete == true
	}() == true {
		__antithesis_instrumentation__.Notify(101387)
		return false
	} else {
		__antithesis_instrumentation__.Notify(101388)
	}
	__antithesis_instrumentation__.Notify(101382)

	if !isDelete && func() bool {
		__antithesis_instrumentation__.Notify(101389)
		return next == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(101390)
		panic("huh")
	} else {
		__antithesis_instrumentation__.Notify(101391)
	}
	__antithesis_instrumentation__.Notify(101383)
	return isDelete || func() bool {
		__antithesis_instrumentation__.Notify(101392)
		return next.Key.Timestamp.LessEq(threshold) == true
	}() == true
}

func processLocalKeyRange(
	ctx context.Context,
	snap storage.Reader,
	desc *roachpb.RangeDescriptor,
	cutoff hlc.Timestamp,
	info *Info,
	cleanupTxnIntentsAsyncFn CleanupTxnIntentsAsyncFunc,
	gcer PureGCer,
) error {
	__antithesis_instrumentation__.Notify(101393)
	b := makeBatchingInlineGCer(gcer, func(err error) {
		__antithesis_instrumentation__.Notify(101400)
		log.Warningf(ctx, "failed to GC from local key range: %s", err)
	})
	__antithesis_instrumentation__.Notify(101394)
	defer b.Flush(ctx)

	handleTxnIntents := func(key roachpb.Key, txn *roachpb.Transaction) error {
		__antithesis_instrumentation__.Notify(101401)

		if !txn.Status.IsFinalized() || func() bool {
			__antithesis_instrumentation__.Notify(101403)
			return len(txn.LockSpans) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(101404)
			return cleanupTxnIntentsAsyncFn(ctx, txn)
		} else {
			__antithesis_instrumentation__.Notify(101405)
		}
		__antithesis_instrumentation__.Notify(101402)
		b.FlushingAdd(ctx, key)
		return nil
	}
	__antithesis_instrumentation__.Notify(101395)

	handleOneTransaction := func(kv roachpb.KeyValue) error {
		__antithesis_instrumentation__.Notify(101406)
		var txn roachpb.Transaction
		if err := kv.Value.GetProto(&txn); err != nil {
			__antithesis_instrumentation__.Notify(101410)
			return err
		} else {
			__antithesis_instrumentation__.Notify(101411)
		}
		__antithesis_instrumentation__.Notify(101407)
		info.TransactionSpanTotal++
		if cutoff.LessEq(txn.LastActive()) {
			__antithesis_instrumentation__.Notify(101412)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(101413)
		}
		__antithesis_instrumentation__.Notify(101408)

		switch txn.Status {
		case roachpb.PENDING:
			__antithesis_instrumentation__.Notify(101414)
			info.TransactionSpanGCPending++
		case roachpb.STAGING:
			__antithesis_instrumentation__.Notify(101415)
			info.TransactionSpanGCStaging++
		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(101416)
			info.TransactionSpanGCAborted++
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(101417)
			info.TransactionSpanGCCommitted++
		default:
			__antithesis_instrumentation__.Notify(101418)
			panic(fmt.Sprintf("invalid transaction state: %s", txn))
		}
		__antithesis_instrumentation__.Notify(101409)
		return handleTxnIntents(kv.Key, &txn)
	}
	__antithesis_instrumentation__.Notify(101396)

	handleOneQueueLastProcessed := func(kv roachpb.KeyValue, rangeKey roachpb.RKey) error {
		__antithesis_instrumentation__.Notify(101419)
		if !rangeKey.Equal(desc.StartKey) {
			__antithesis_instrumentation__.Notify(101421)

			b.FlushingAdd(ctx, kv.Key)
		} else {
			__antithesis_instrumentation__.Notify(101422)
		}
		__antithesis_instrumentation__.Notify(101420)
		return nil
	}
	__antithesis_instrumentation__.Notify(101397)

	handleOne := func(kv roachpb.KeyValue) error {
		__antithesis_instrumentation__.Notify(101423)
		rangeKey, suffix, _, err := keys.DecodeRangeKey(kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(101426)
			return err
		} else {
			__antithesis_instrumentation__.Notify(101427)
		}
		__antithesis_instrumentation__.Notify(101424)
		if suffix.Equal(keys.LocalTransactionSuffix.AsRawKey()) {
			__antithesis_instrumentation__.Notify(101428)
			if err := handleOneTransaction(kv); err != nil {
				__antithesis_instrumentation__.Notify(101429)
				return err
			} else {
				__antithesis_instrumentation__.Notify(101430)
			}
		} else {
			__antithesis_instrumentation__.Notify(101431)
			if suffix.Equal(keys.LocalQueueLastProcessedSuffix.AsRawKey()) {
				__antithesis_instrumentation__.Notify(101432)
				if err := handleOneQueueLastProcessed(kv, roachpb.RKey(rangeKey)); err != nil {
					__antithesis_instrumentation__.Notify(101433)
					return err
				} else {
					__antithesis_instrumentation__.Notify(101434)
				}
			} else {
				__antithesis_instrumentation__.Notify(101435)
			}
		}
		__antithesis_instrumentation__.Notify(101425)
		return nil
	}
	__antithesis_instrumentation__.Notify(101398)

	startKey := keys.MakeRangeKeyPrefix(desc.StartKey)
	endKey := keys.MakeRangeKeyPrefix(desc.EndKey)

	_, err := storage.MVCCIterate(ctx, snap, startKey, endKey, hlc.Timestamp{}, storage.MVCCScanOptions{},
		func(kv roachpb.KeyValue) error {
			__antithesis_instrumentation__.Notify(101436)
			return handleOne(kv)
		})
	__antithesis_instrumentation__.Notify(101399)
	return err
}

func processAbortSpan(
	ctx context.Context,
	snap storage.Reader,
	rangeID roachpb.RangeID,
	threshold hlc.Timestamp,
	info *Info,
	gcer PureGCer,
) error {
	__antithesis_instrumentation__.Notify(101437)
	b := makeBatchingInlineGCer(gcer, func(err error) {
		__antithesis_instrumentation__.Notify(101439)
		log.Warningf(ctx, "unable to GC from abort span: %s", err)
	})
	__antithesis_instrumentation__.Notify(101438)
	defer b.Flush(ctx)
	abortSpan := abortspan.New(rangeID)
	return abortSpan.Iterate(ctx, snap, func(key roachpb.Key, v roachpb.AbortSpanEntry) error {
		__antithesis_instrumentation__.Notify(101440)
		info.AbortSpanTotal++
		if v.Timestamp.Less(threshold) {
			__antithesis_instrumentation__.Notify(101442)
			info.AbortSpanGCNum++
			b.FlushingAdd(ctx, key)
		} else {
			__antithesis_instrumentation__.Notify(101443)
		}
		__antithesis_instrumentation__.Notify(101441)
		return nil
	})
}

type batchingInlineGCer struct {
	gcer  PureGCer
	onErr func(error)

	size   int
	max    int
	gcKeys []roachpb.GCRequest_GCKey
}

func makeBatchingInlineGCer(gcer PureGCer, onErr func(error)) batchingInlineGCer {
	__antithesis_instrumentation__.Notify(101444)
	return batchingInlineGCer{gcer: gcer, onErr: onErr, max: KeyVersionChunkBytes}
}

func (b *batchingInlineGCer) FlushingAdd(ctx context.Context, key roachpb.Key) {
	__antithesis_instrumentation__.Notify(101445)
	b.gcKeys = append(b.gcKeys, roachpb.GCRequest_GCKey{Key: key})
	b.size += len(key)
	if b.size < b.max {
		__antithesis_instrumentation__.Notify(101447)
		return
	} else {
		__antithesis_instrumentation__.Notify(101448)
	}
	__antithesis_instrumentation__.Notify(101446)
	b.Flush(ctx)
}

func (b *batchingInlineGCer) Flush(ctx context.Context) {
	__antithesis_instrumentation__.Notify(101449)
	err := b.gcer.GC(ctx, b.gcKeys)
	b.gcKeys = nil
	b.size = 0
	if err != nil {
		__antithesis_instrumentation__.Notify(101450)
		b.onErr(err)
	} else {
		__antithesis_instrumentation__.Notify(101451)
	}
}
