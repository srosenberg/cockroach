package intentresolver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/internal/client/requestbatcher"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const (
	defaultTaskLimit = 1000

	asyncIntentResolutionTimeout = 30 * time.Second

	gcBatchSize = 1024

	intentResolverBatchSize = 100

	intentResolverRangeBatchSize = 10

	intentResolverRangeRequestSize = 200

	MaxTxnsPerIntentCleanupBatch = 100

	defaultGCBatchIdle = -1

	defaultGCBatchWait = time.Second

	defaultIntentResolutionBatchWait = 10 * time.Millisecond

	defaultIntentResolutionBatchIdle = 5 * time.Millisecond

	gcTxnRecordTimeout = 20 * time.Second
)

type Config struct {
	Clock                *hlc.Clock
	DB                   *kv.DB
	Stopper              *stop.Stopper
	AmbientCtx           log.AmbientContext
	TestingKnobs         kvserverbase.IntentResolverTestingKnobs
	RangeDescriptorCache RangeCache

	TaskLimit                    int
	MaxGCBatchWait               time.Duration
	MaxGCBatchIdle               time.Duration
	MaxIntentResolutionBatchWait time.Duration
	MaxIntentResolutionBatchIdle time.Duration
}

type RangeCache interface {
	Lookup(ctx context.Context, key roachpb.RKey) (rangecache.CacheEntry, error)
}

type IntentResolver struct {
	Metrics Metrics

	clock        *hlc.Clock
	db           *kv.DB
	stopper      *stop.Stopper
	testingKnobs kvserverbase.IntentResolverTestingKnobs
	ambientCtx   log.AmbientContext
	sem          *quotapool.IntPool

	rdc RangeCache

	gcBatcher      *requestbatcher.RequestBatcher
	irBatcher      *requestbatcher.RequestBatcher
	irRangeBatcher *requestbatcher.RequestBatcher

	mu struct {
		syncutil.Mutex

		inFlightPushes map[uuid.UUID]int

		inFlightTxnCleanups map[uuid.UUID]struct{}
	}
	every log.EveryN
}

func setConfigDefaults(c *Config) {
	__antithesis_instrumentation__.Notify(101551)
	if c.TaskLimit == 0 {
		__antithesis_instrumentation__.Notify(101558)
		c.TaskLimit = defaultTaskLimit
	} else {
		__antithesis_instrumentation__.Notify(101559)
	}
	__antithesis_instrumentation__.Notify(101552)
	if c.TaskLimit == -1 || func() bool {
		__antithesis_instrumentation__.Notify(101560)
		return c.TestingKnobs.ForceSyncIntentResolution == true
	}() == true {
		__antithesis_instrumentation__.Notify(101561)
		c.TaskLimit = 0
	} else {
		__antithesis_instrumentation__.Notify(101562)
	}
	__antithesis_instrumentation__.Notify(101553)
	if c.MaxGCBatchIdle == 0 {
		__antithesis_instrumentation__.Notify(101563)
		c.MaxGCBatchIdle = defaultGCBatchIdle
	} else {
		__antithesis_instrumentation__.Notify(101564)
	}
	__antithesis_instrumentation__.Notify(101554)
	if c.MaxGCBatchWait == 0 {
		__antithesis_instrumentation__.Notify(101565)
		c.MaxGCBatchWait = defaultGCBatchWait
	} else {
		__antithesis_instrumentation__.Notify(101566)
	}
	__antithesis_instrumentation__.Notify(101555)
	if c.MaxIntentResolutionBatchIdle == 0 {
		__antithesis_instrumentation__.Notify(101567)
		c.MaxIntentResolutionBatchIdle = defaultIntentResolutionBatchIdle
	} else {
		__antithesis_instrumentation__.Notify(101568)
	}
	__antithesis_instrumentation__.Notify(101556)
	if c.MaxIntentResolutionBatchWait == 0 {
		__antithesis_instrumentation__.Notify(101569)
		c.MaxIntentResolutionBatchWait = defaultIntentResolutionBatchWait
	} else {
		__antithesis_instrumentation__.Notify(101570)
	}
	__antithesis_instrumentation__.Notify(101557)
	if c.RangeDescriptorCache == nil {
		__antithesis_instrumentation__.Notify(101571)
		c.RangeDescriptorCache = nopRangeDescriptorCache{}
	} else {
		__antithesis_instrumentation__.Notify(101572)
	}
}

type nopRangeDescriptorCache struct{}

func (nrdc nopRangeDescriptorCache) Lookup(
	ctx context.Context, key roachpb.RKey,
) (rangecache.CacheEntry, error) {
	__antithesis_instrumentation__.Notify(101573)
	return rangecache.CacheEntry{}, nil
}

func New(c Config) *IntentResolver {
	__antithesis_instrumentation__.Notify(101574)
	setConfigDefaults(&c)
	ir := &IntentResolver{
		clock:        c.Clock,
		db:           c.DB,
		stopper:      c.Stopper,
		sem:          quotapool.NewIntPool("intent resolver", uint64(c.TaskLimit)),
		every:        log.Every(time.Minute),
		Metrics:      makeMetrics(),
		rdc:          c.RangeDescriptorCache,
		testingKnobs: c.TestingKnobs,
	}
	c.Stopper.AddCloser(ir.sem.Closer("stopper"))
	ir.mu.inFlightPushes = map[uuid.UUID]int{}
	ir.mu.inFlightTxnCleanups = map[uuid.UUID]struct{}{}
	gcBatchSize := gcBatchSize
	if c.TestingKnobs.MaxIntentResolutionBatchSize > 0 {
		__antithesis_instrumentation__.Notify(101577)
		gcBatchSize = c.TestingKnobs.MaxGCBatchSize
	} else {
		__antithesis_instrumentation__.Notify(101578)
	}
	__antithesis_instrumentation__.Notify(101575)
	ir.gcBatcher = requestbatcher.New(requestbatcher.Config{
		AmbientCtx:      c.AmbientCtx,
		Name:            "intent_resolver_gc_batcher",
		MaxMsgsPerBatch: gcBatchSize,
		MaxWait:         c.MaxGCBatchWait,
		MaxIdle:         c.MaxGCBatchIdle,
		Stopper:         c.Stopper,
		Sender:          c.DB.NonTransactionalSender(),
	})
	intentResolutionBatchSize := intentResolverBatchSize
	intentResolutionRangeBatchSize := intentResolverRangeBatchSize
	if c.TestingKnobs.MaxIntentResolutionBatchSize > 0 {
		__antithesis_instrumentation__.Notify(101579)
		intentResolutionBatchSize = c.TestingKnobs.MaxIntentResolutionBatchSize
		intentResolutionRangeBatchSize = c.TestingKnobs.MaxIntentResolutionBatchSize
	} else {
		__antithesis_instrumentation__.Notify(101580)
	}
	__antithesis_instrumentation__.Notify(101576)
	ir.irBatcher = requestbatcher.New(requestbatcher.Config{
		AmbientCtx:      c.AmbientCtx,
		Name:            "intent_resolver_ir_batcher",
		MaxMsgsPerBatch: intentResolutionBatchSize,
		MaxWait:         c.MaxIntentResolutionBatchWait,
		MaxIdle:         c.MaxIntentResolutionBatchIdle,
		Stopper:         c.Stopper,
		Sender:          c.DB.NonTransactionalSender(),
	})
	ir.irRangeBatcher = requestbatcher.New(requestbatcher.Config{
		AmbientCtx:         c.AmbientCtx,
		Name:               "intent_resolver_ir_range_batcher",
		MaxMsgsPerBatch:    intentResolutionRangeBatchSize,
		MaxKeysPerBatchReq: intentResolverRangeRequestSize,
		MaxWait:            c.MaxIntentResolutionBatchWait,
		MaxIdle:            c.MaxIntentResolutionBatchIdle,
		Stopper:            c.Stopper,
		Sender:             c.DB.NonTransactionalSender(),
	})
	return ir
}

func getPusherTxn(h roachpb.Header) roachpb.Transaction {
	__antithesis_instrumentation__.Notify(101581)

	txn := h.Txn
	if txn == nil {
		__antithesis_instrumentation__.Notify(101583)
		txn = &roachpb.Transaction{
			TxnMeta: enginepb.TxnMeta{
				Priority: roachpb.MakePriority(h.UserPriority),
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(101584)
	}
	__antithesis_instrumentation__.Notify(101582)
	return *txn
}

func updateIntentTxnStatus(
	ctx context.Context,
	pushedTxns map[uuid.UUID]*roachpb.Transaction,
	intents []roachpb.Intent,
	skipIfInFlight bool,
	results []roachpb.LockUpdate,
) []roachpb.LockUpdate {
	__antithesis_instrumentation__.Notify(101585)
	for _, intent := range intents {
		__antithesis_instrumentation__.Notify(101587)
		pushee, ok := pushedTxns[intent.Txn.ID]
		if !ok {
			__antithesis_instrumentation__.Notify(101589)

			if !skipIfInFlight {
				__antithesis_instrumentation__.Notify(101591)
				log.Fatalf(ctx, "no PushTxn response for intent %+v", intent)
			} else {
				__antithesis_instrumentation__.Notify(101592)
			}
			__antithesis_instrumentation__.Notify(101590)

			continue
		} else {
			__antithesis_instrumentation__.Notify(101593)
		}
		__antithesis_instrumentation__.Notify(101588)
		up := roachpb.MakeLockUpdate(pushee, roachpb.Span{Key: intent.Key})
		results = append(results, up)
	}
	__antithesis_instrumentation__.Notify(101586)
	return results
}

func (ir *IntentResolver) PushTransaction(
	ctx context.Context, pushTxn *enginepb.TxnMeta, h roachpb.Header, pushType roachpb.PushTxnType,
) (*roachpb.Transaction, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(101594)
	pushTxns := make(map[uuid.UUID]*enginepb.TxnMeta, 1)
	pushTxns[pushTxn.ID] = pushTxn
	pushedTxns, pErr := ir.MaybePushTransactions(ctx, pushTxns, h, pushType, false)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(101597)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(101598)
	}
	__antithesis_instrumentation__.Notify(101595)
	pushedTxn, ok := pushedTxns[pushTxn.ID]
	if !ok {
		__antithesis_instrumentation__.Notify(101599)
		log.Fatalf(ctx, "missing PushTxn responses for %s", pushTxn)
	} else {
		__antithesis_instrumentation__.Notify(101600)
	}
	__antithesis_instrumentation__.Notify(101596)
	return pushedTxn, nil
}

func (ir *IntentResolver) MaybePushTransactions(
	ctx context.Context,
	pushTxns map[uuid.UUID]*enginepb.TxnMeta,
	h roachpb.Header,
	pushType roachpb.PushTxnType,
	skipIfInFlight bool,
) (map[uuid.UUID]*roachpb.Transaction, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(101601)

	ir.mu.Lock()
	for txnID := range pushTxns {
		__antithesis_instrumentation__.Notify(101608)
		_, pushTxnInFlight := ir.mu.inFlightPushes[txnID]
		if pushTxnInFlight && func() bool {
			__antithesis_instrumentation__.Notify(101609)
			return skipIfInFlight == true
		}() == true {
			__antithesis_instrumentation__.Notify(101610)

			if log.V(1) {
				__antithesis_instrumentation__.Notify(101612)
				log.Infof(ctx, "skipping PushTxn for %s; attempt already in flight", txnID)
			} else {
				__antithesis_instrumentation__.Notify(101613)
			}
			__antithesis_instrumentation__.Notify(101611)
			delete(pushTxns, txnID)
		} else {
			__antithesis_instrumentation__.Notify(101614)
			ir.mu.inFlightPushes[txnID]++
		}
	}
	__antithesis_instrumentation__.Notify(101602)
	cleanupInFlightPushes := func() {
		__antithesis_instrumentation__.Notify(101615)
		ir.mu.Lock()
		for txnID := range pushTxns {
			__antithesis_instrumentation__.Notify(101617)
			ir.mu.inFlightPushes[txnID]--
			if ir.mu.inFlightPushes[txnID] == 0 {
				__antithesis_instrumentation__.Notify(101618)
				delete(ir.mu.inFlightPushes, txnID)
			} else {
				__antithesis_instrumentation__.Notify(101619)
			}
		}
		__antithesis_instrumentation__.Notify(101616)
		ir.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(101603)
	ir.mu.Unlock()
	if len(pushTxns) == 0 {
		__antithesis_instrumentation__.Notify(101620)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(101621)
	}
	__antithesis_instrumentation__.Notify(101604)

	pusherTxn := getPusherTxn(h)
	log.Eventf(ctx, "pushing %d transaction(s)", len(pushTxns))

	pushTo := h.Timestamp.Next()
	b := &kv.Batch{}
	b.Header.Timestamp = ir.clock.Now()
	b.Header.Timestamp.Forward(pushTo)
	for _, pushTxn := range pushTxns {
		__antithesis_instrumentation__.Notify(101622)
		b.AddRawRequest(&roachpb.PushTxnRequest{
			RequestHeader: roachpb.RequestHeader{
				Key: pushTxn.Key,
			},
			PusherTxn: pusherTxn,
			PusheeTxn: *pushTxn,
			PushTo:    pushTo,
			PushType:  pushType,
		})
	}
	__antithesis_instrumentation__.Notify(101605)
	err := ir.db.Run(ctx, b)
	cleanupInFlightPushes()
	if err != nil {
		__antithesis_instrumentation__.Notify(101623)
		return nil, b.MustPErr()
	} else {
		__antithesis_instrumentation__.Notify(101624)
	}
	__antithesis_instrumentation__.Notify(101606)

	br := b.RawResponse()
	pushedTxns := make(map[uuid.UUID]*roachpb.Transaction, len(br.Responses))
	for _, resp := range br.Responses {
		__antithesis_instrumentation__.Notify(101625)
		txn := &resp.GetInner().(*roachpb.PushTxnResponse).PusheeTxn
		if _, ok := pushedTxns[txn.ID]; ok {
			__antithesis_instrumentation__.Notify(101627)
			log.Fatalf(ctx, "have two PushTxn responses for %s", txn.ID)
		} else {
			__antithesis_instrumentation__.Notify(101628)
		}
		__antithesis_instrumentation__.Notify(101626)
		pushedTxns[txn.ID] = txn
		log.Eventf(ctx, "%s is now %s", txn.ID, txn.Status)
	}
	__antithesis_instrumentation__.Notify(101607)
	return pushedTxns, nil
}

func (ir *IntentResolver) runAsyncTask(
	ctx context.Context, allowSyncProcessing bool, taskFn func(context.Context),
) error {
	__antithesis_instrumentation__.Notify(101629)
	if ir.testingKnobs.DisableAsyncIntentResolution {
		__antithesis_instrumentation__.Notify(101632)
		return errors.New("intents not processed as async resolution is disabled")
	} else {
		__antithesis_instrumentation__.Notify(101633)
	}
	__antithesis_instrumentation__.Notify(101630)
	err := ir.stopper.RunAsyncTaskEx(

		ir.ambientCtx.AnnotateCtx(context.Background()),
		stop.TaskOpts{
			TaskName:   "storage.IntentResolver: processing intents",
			Sem:        ir.sem,
			WaitForSem: false,
		},
		taskFn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(101634)
		if errors.Is(err, stop.ErrThrottled) {
			__antithesis_instrumentation__.Notify(101636)
			ir.Metrics.IntentResolverAsyncThrottled.Inc(1)
			if allowSyncProcessing {
				__antithesis_instrumentation__.Notify(101637)

				taskFn(ctx)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(101638)
			}
		} else {
			__antithesis_instrumentation__.Notify(101639)
		}
		__antithesis_instrumentation__.Notify(101635)
		return errors.Wrapf(err, "during async intent resolution")
	} else {
		__antithesis_instrumentation__.Notify(101640)
	}
	__antithesis_instrumentation__.Notify(101631)
	return nil
}

func (ir *IntentResolver) CleanupIntentsAsync(
	ctx context.Context, intents []roachpb.Intent, allowSyncProcessing bool,
) error {
	__antithesis_instrumentation__.Notify(101641)
	if len(intents) == 0 {
		__antithesis_instrumentation__.Notify(101643)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(101644)
	}
	__antithesis_instrumentation__.Notify(101642)
	now := ir.clock.Now()
	return ir.runAsyncTask(ctx, allowSyncProcessing, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(101645)
		err := contextutil.RunWithTimeout(ctx, "async intent resolution",
			asyncIntentResolutionTimeout, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(101647)
				_, err := ir.CleanupIntents(ctx, intents, now, roachpb.PUSH_TOUCH)
				return err
			})
		__antithesis_instrumentation__.Notify(101646)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(101648)
			return ir.every.ShouldLog() == true
		}() == true {
			__antithesis_instrumentation__.Notify(101649)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(101650)
		}
	})
}

func (ir *IntentResolver) CleanupIntents(
	ctx context.Context, intents []roachpb.Intent, now hlc.Timestamp, pushType roachpb.PushTxnType,
) (int, error) {
	__antithesis_instrumentation__.Notify(101651)
	h := roachpb.Header{Timestamp: now}

	sort.Sort(intentsByTxn(intents))
	resolved := 0
	const skipIfInFlight = true
	pushTxns := make(map[uuid.UUID]*enginepb.TxnMeta)
	var resolveIntents []roachpb.LockUpdate
	for unpushed := intents; len(unpushed) > 0; {
		__antithesis_instrumentation__.Notify(101653)
		for k := range pushTxns {
			__antithesis_instrumentation__.Notify(101658)
			delete(pushTxns, k)
		}
		__antithesis_instrumentation__.Notify(101654)
		var prevTxnID uuid.UUID
		var i int
		for i = 0; i < len(unpushed); i++ {
			__antithesis_instrumentation__.Notify(101659)
			if curTxn := &unpushed[i].Txn; curTxn.ID != prevTxnID {
				__antithesis_instrumentation__.Notify(101660)
				if len(pushTxns) >= MaxTxnsPerIntentCleanupBatch {
					__antithesis_instrumentation__.Notify(101662)
					break
				} else {
					__antithesis_instrumentation__.Notify(101663)
				}
				__antithesis_instrumentation__.Notify(101661)
				prevTxnID = curTxn.ID
				pushTxns[curTxn.ID] = curTxn
			} else {
				__antithesis_instrumentation__.Notify(101664)
			}
		}
		__antithesis_instrumentation__.Notify(101655)

		pushedTxns, pErr := ir.MaybePushTransactions(ctx, pushTxns, h, pushType, skipIfInFlight)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(101665)
			return 0, errors.Wrapf(pErr.GoError(), "failed to push during intent resolution")
		} else {
			__antithesis_instrumentation__.Notify(101666)
		}
		__antithesis_instrumentation__.Notify(101656)
		resolveIntents = updateIntentTxnStatus(ctx, pushedTxns, unpushed[:i],
			skipIfInFlight, resolveIntents[:0])

		opts := ResolveOptions{Poison: true}
		if pErr := ir.ResolveIntents(ctx, resolveIntents, opts); pErr != nil {
			__antithesis_instrumentation__.Notify(101667)
			return 0, errors.Wrapf(pErr.GoError(), "failed to resolve intents")
		} else {
			__antithesis_instrumentation__.Notify(101668)
		}
		__antithesis_instrumentation__.Notify(101657)
		resolved += len(resolveIntents)
		unpushed = unpushed[i:]
	}
	__antithesis_instrumentation__.Notify(101652)
	return resolved, nil
}

func (ir *IntentResolver) CleanupTxnIntentsAsync(
	ctx context.Context,
	rangeID roachpb.RangeID,
	endTxns []result.EndTxnIntents,
	allowSyncProcessing bool,
) error {
	__antithesis_instrumentation__.Notify(101669)
	for i := range endTxns {
		__antithesis_instrumentation__.Notify(101671)
		onComplete := func(err error) {
			__antithesis_instrumentation__.Notify(101673)
			if err != nil {
				__antithesis_instrumentation__.Notify(101674)
				ir.Metrics.FinalizedTxnCleanupFailed.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(101675)
			}
		}
		__antithesis_instrumentation__.Notify(101672)
		et := &endTxns[i]
		if err := ir.runAsyncTask(ctx, allowSyncProcessing, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(101676)
			locked, release := ir.lockInFlightTxnCleanup(ctx, et.Txn.ID)
			if !locked {
				__antithesis_instrumentation__.Notify(101678)
				return
			} else {
				__antithesis_instrumentation__.Notify(101679)
			}
			__antithesis_instrumentation__.Notify(101677)
			defer release()
			if err := ir.cleanupFinishedTxnIntents(
				ctx, rangeID, et.Txn, et.Poison, onComplete,
			); err != nil {
				__antithesis_instrumentation__.Notify(101680)
				if ir.every.ShouldLog() {
					__antithesis_instrumentation__.Notify(101681)
					log.Warningf(ctx, "failed to cleanup transaction intents: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(101682)
				}
			} else {
				__antithesis_instrumentation__.Notify(101683)
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(101684)
			ir.Metrics.FinalizedTxnCleanupFailed.Inc(int64(len(endTxns) - i))
			return err
		} else {
			__antithesis_instrumentation__.Notify(101685)
		}
	}
	__antithesis_instrumentation__.Notify(101670)
	return nil
}

func (ir *IntentResolver) lockInFlightTxnCleanup(
	ctx context.Context, txnID uuid.UUID,
) (locked bool, release func()) {
	__antithesis_instrumentation__.Notify(101686)
	ir.mu.Lock()
	defer ir.mu.Unlock()
	_, inFlight := ir.mu.inFlightTxnCleanups[txnID]
	if inFlight {
		__antithesis_instrumentation__.Notify(101688)
		log.Eventf(ctx, "skipping txn resolved; already in flight")
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(101689)
	}
	__antithesis_instrumentation__.Notify(101687)
	ir.mu.inFlightTxnCleanups[txnID] = struct{}{}
	return true, func() {
		__antithesis_instrumentation__.Notify(101690)
		ir.mu.Lock()
		delete(ir.mu.inFlightTxnCleanups, txnID)
		ir.mu.Unlock()
	}
}

func (ir *IntentResolver) CleanupTxnIntentsOnGCAsync(
	ctx context.Context,
	rangeID roachpb.RangeID,
	txn *roachpb.Transaction,
	now hlc.Timestamp,
	onComplete func(pushed, succeeded bool),
) error {
	__antithesis_instrumentation__.Notify(101691)
	return ir.stopper.RunAsyncTaskEx(

		ir.ambientCtx.AnnotateCtx(context.Background()),
		stop.TaskOpts{
			TaskName: "processing txn intents",
			Sem:      ir.sem,

			WaitForSem: false,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(101692)
			var pushed, succeeded bool
			defer func() {
				__antithesis_instrumentation__.Notify(101697)
				if onComplete != nil {
					__antithesis_instrumentation__.Notify(101698)
					onComplete(pushed, succeeded)
				} else {
					__antithesis_instrumentation__.Notify(101699)
				}
			}()
			__antithesis_instrumentation__.Notify(101693)
			locked, release := ir.lockInFlightTxnCleanup(ctx, txn.ID)
			if !locked {
				__antithesis_instrumentation__.Notify(101700)
				return
			} else {
				__antithesis_instrumentation__.Notify(101701)
			}
			__antithesis_instrumentation__.Notify(101694)
			defer release()

			if !txn.Status.IsFinalized() {
				__antithesis_instrumentation__.Notify(101702)
				if !txnwait.IsExpired(now, txn) {
					__antithesis_instrumentation__.Notify(101705)
					log.VErrEventf(ctx, 3, "cannot push a %s transaction which is not expired: %s", txn.Status, txn)
					return
				} else {
					__antithesis_instrumentation__.Notify(101706)
				}
				__antithesis_instrumentation__.Notify(101703)
				b := &kv.Batch{}
				b.Header.Timestamp = now
				b.AddRawRequest(&roachpb.PushTxnRequest{
					RequestHeader: roachpb.RequestHeader{Key: txn.Key},
					PusherTxn: roachpb.Transaction{
						TxnMeta: enginepb.TxnMeta{Priority: enginepb.MaxTxnPriority},
					},
					PusheeTxn: txn.TxnMeta,
					PushType:  roachpb.PUSH_ABORT,
				})
				pushed = true
				if err := ir.db.Run(ctx, b); err != nil {
					__antithesis_instrumentation__.Notify(101707)
					log.VErrEventf(ctx, 2, "failed to push %s, expired txn (%s): %s", txn.Status, txn, err)
					return
				} else {
					__antithesis_instrumentation__.Notify(101708)
				}
				__antithesis_instrumentation__.Notify(101704)

				finalizedTxn := &b.RawResponse().Responses[0].GetInner().(*roachpb.PushTxnResponse).PusheeTxn
				txn = txn.Clone()
				txn.Update(finalizedTxn)
			} else {
				__antithesis_instrumentation__.Notify(101709)
			}
			__antithesis_instrumentation__.Notify(101695)
			var onCleanupComplete func(error)
			if onComplete != nil {
				__antithesis_instrumentation__.Notify(101710)
				onCompleteCopy := onComplete
				onCleanupComplete = func(err error) {
					__antithesis_instrumentation__.Notify(101711)
					onCompleteCopy(pushed, err == nil)
				}
			} else {
				__antithesis_instrumentation__.Notify(101712)
			}
			__antithesis_instrumentation__.Notify(101696)

			onComplete = nil
			err := ir.cleanupFinishedTxnIntents(ctx, rangeID, txn, false, onCleanupComplete)
			if err != nil {
				__antithesis_instrumentation__.Notify(101713)
				if ir.every.ShouldLog() {
					__antithesis_instrumentation__.Notify(101714)
					log.Warningf(ctx, "failed to cleanup transaction intents: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(101715)
				}
			} else {
				__antithesis_instrumentation__.Notify(101716)
			}
		},
	)
}

func (ir *IntentResolver) gcTxnRecord(
	ctx context.Context, rangeID roachpb.RangeID, txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(101717)

	txnKey := keys.TransactionKey(txn.Key, txn.ID)

	var gcArgs roachpb.GCRequest
	{
		__antithesis_instrumentation__.Notify(101720)
		key := keys.MustAddr(txn.Key)
		if localMax := keys.MustAddr(keys.LocalMax); key.Less(localMax) {
			__antithesis_instrumentation__.Notify(101722)
			key = localMax
		} else {
			__antithesis_instrumentation__.Notify(101723)
		}
		__antithesis_instrumentation__.Notify(101721)
		endKey := key.Next()

		gcArgs.RequestHeader = roachpb.RequestHeader{
			Key:    key.AsRawKey(),
			EndKey: endKey.AsRawKey(),
		}
	}
	__antithesis_instrumentation__.Notify(101718)
	gcArgs.Keys = append(gcArgs.Keys, roachpb.GCRequest_GCKey{
		Key: txnKey,
	})

	_, err := ir.gcBatcher.Send(ctx, rangeID, &gcArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(101724)
		return errors.Wrapf(err, "could not GC completed transaction anchored at %s",
			roachpb.Key(txn.Key))
	} else {
		__antithesis_instrumentation__.Notify(101725)
	}
	__antithesis_instrumentation__.Notify(101719)
	return nil
}

func (ir *IntentResolver) cleanupFinishedTxnIntents(
	ctx context.Context,
	rangeID roachpb.RangeID,
	txn *roachpb.Transaction,
	poison bool,
	onComplete func(error),
) (err error) {
	__antithesis_instrumentation__.Notify(101726)
	defer func() {
		__antithesis_instrumentation__.Notify(101729)

		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(101730)
			return onComplete != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(101731)
			onComplete(err)
		} else {
			__antithesis_instrumentation__.Notify(101732)
		}
	}()
	__antithesis_instrumentation__.Notify(101727)

	opts := ResolveOptions{Poison: poison, MinTimestamp: txn.MinTimestamp}
	if pErr := ir.ResolveIntents(ctx, txn.LocksAsLockUpdates(), opts); pErr != nil {
		__antithesis_instrumentation__.Notify(101733)
		return errors.Wrapf(pErr.GoError(), "failed to resolve intents")
	} else {
		__antithesis_instrumentation__.Notify(101734)
	}
	__antithesis_instrumentation__.Notify(101728)

	return ir.stopper.RunAsyncTask(
		ir.ambientCtx.AnnotateCtx(context.Background()),
		"storage.IntentResolver: cleanup txn records",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(101735)
			err := contextutil.RunWithTimeout(ctx, "cleanup txn record",
				gcTxnRecordTimeout, func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(101738)
					return ir.gcTxnRecord(ctx, rangeID, txn)
				})
			__antithesis_instrumentation__.Notify(101736)
			if onComplete != nil {
				__antithesis_instrumentation__.Notify(101739)
				onComplete(err)
			} else {
				__antithesis_instrumentation__.Notify(101740)
			}
			__antithesis_instrumentation__.Notify(101737)
			if err != nil {
				__antithesis_instrumentation__.Notify(101741)
				if ir.every.ShouldLog() {
					__antithesis_instrumentation__.Notify(101742)
					log.Warningf(ctx, "failed to gc transaction record: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(101743)
				}
			} else {
				__antithesis_instrumentation__.Notify(101744)
			}
		})
}

type ResolveOptions struct {
	Poison bool

	MinTimestamp hlc.Timestamp
}

func (ir *IntentResolver) lookupRangeID(ctx context.Context, key roachpb.Key) roachpb.RangeID {
	__antithesis_instrumentation__.Notify(101745)
	rKey, err := keys.Addr(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(101748)
		if ir.every.ShouldLog() {
			__antithesis_instrumentation__.Notify(101750)
			log.Warningf(ctx, "failed to resolve addr for key %q: %+v", key, err)
		} else {
			__antithesis_instrumentation__.Notify(101751)
		}
		__antithesis_instrumentation__.Notify(101749)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(101752)
	}
	__antithesis_instrumentation__.Notify(101746)
	rInfo, err := ir.rdc.Lookup(ctx, rKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(101753)
		if ir.every.ShouldLog() {
			__antithesis_instrumentation__.Notify(101755)
			log.Warningf(ctx, "failed to look up range descriptor for key %q: %+v", key, err)
		} else {
			__antithesis_instrumentation__.Notify(101756)
		}
		__antithesis_instrumentation__.Notify(101754)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(101757)
	}
	__antithesis_instrumentation__.Notify(101747)
	return rInfo.Desc().RangeID
}

func (ir *IntentResolver) ResolveIntent(
	ctx context.Context, intent roachpb.LockUpdate, opts ResolveOptions,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(101758)
	return ir.ResolveIntents(ctx, []roachpb.LockUpdate{intent}, opts)
}

func (ir *IntentResolver) ResolveIntents(
	ctx context.Context, intents []roachpb.LockUpdate, opts ResolveOptions,
) (pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(101759)
	if len(intents) == 0 {
		__antithesis_instrumentation__.Notify(101765)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(101766)
	}
	__antithesis_instrumentation__.Notify(101760)
	defer func() {
		__antithesis_instrumentation__.Notify(101767)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(101768)
			ir.Metrics.IntentResolutionFailed.Inc(int64(len(intents)))
		} else {
			__antithesis_instrumentation__.Notify(101769)
		}
	}()
	__antithesis_instrumentation__.Notify(101761)

	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(101770)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(101771)
	}
	__antithesis_instrumentation__.Notify(101762)
	log.Eventf(ctx, "resolving intents")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	respChan := make(chan requestbatcher.Response, len(intents))
	for _, intent := range intents {
		__antithesis_instrumentation__.Notify(101772)
		rangeID := ir.lookupRangeID(ctx, intent.Key)
		var req roachpb.Request
		var batcher *requestbatcher.RequestBatcher
		if len(intent.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(101774)
			req = &roachpb.ResolveIntentRequest{
				RequestHeader:  roachpb.RequestHeaderFromSpan(intent.Span),
				IntentTxn:      intent.Txn,
				Status:         intent.Status,
				Poison:         opts.Poison,
				IgnoredSeqNums: intent.IgnoredSeqNums,
			}
			batcher = ir.irBatcher
		} else {
			__antithesis_instrumentation__.Notify(101775)
			req = &roachpb.ResolveIntentRangeRequest{
				RequestHeader:  roachpb.RequestHeaderFromSpan(intent.Span),
				IntentTxn:      intent.Txn,
				Status:         intent.Status,
				Poison:         opts.Poison,
				MinTimestamp:   opts.MinTimestamp,
				IgnoredSeqNums: intent.IgnoredSeqNums,
			}
			batcher = ir.irRangeBatcher
		}
		__antithesis_instrumentation__.Notify(101773)
		if err := batcher.SendWithChan(ctx, respChan, rangeID, req); err != nil {
			__antithesis_instrumentation__.Notify(101776)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(101777)
		}
	}
	__antithesis_instrumentation__.Notify(101763)
	for seen := 0; seen < len(intents); seen++ {
		__antithesis_instrumentation__.Notify(101778)
		select {
		case resp := <-respChan:
			__antithesis_instrumentation__.Notify(101779)
			if resp.Err != nil {
				__antithesis_instrumentation__.Notify(101783)
				return roachpb.NewError(resp.Err)
			} else {
				__antithesis_instrumentation__.Notify(101784)
			}
			__antithesis_instrumentation__.Notify(101780)
			_ = resp.Resp
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(101781)
			return roachpb.NewError(ctx.Err())
		case <-ir.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(101782)
			return roachpb.NewErrorf("stopping")
		}
	}
	__antithesis_instrumentation__.Notify(101764)
	return nil
}

type intentsByTxn []roachpb.Intent

var _ sort.Interface = intentsByTxn(nil)

func (s intentsByTxn) Len() int { __antithesis_instrumentation__.Notify(101785); return len(s) }
func (s intentsByTxn) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(101786)
	s[i], s[j] = s[j], s[i]
}
func (s intentsByTxn) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(101787)
	return bytes.Compare(s[i].Txn.ID[:], s[j].Txn.ID[:]) < 0
}
