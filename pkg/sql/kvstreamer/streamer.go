package kvstreamer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const debug = false

type OperationMode int

const (
	_ OperationMode = iota

	InOrder

	OutOfOrder
)

type Result struct {
	GetResp *roachpb.GetResponse

	ScanResp struct {
		*roachpb.ScanResponse

		Complete bool
	}

	EnqueueKeysSatisfied []int

	memoryTok struct {
		streamer  *Streamer
		toRelease int64
	}

	position int
}

type Hints struct {
	UniqueRequests bool
}

func (r Result) Release(ctx context.Context) {
	__antithesis_instrumentation__.Notify(499122)
	if s := r.memoryTok.streamer; s != nil {
		__antithesis_instrumentation__.Notify(499123)
		s.results.releaseOne()
		s.budget.mu.Lock()
		defer s.budget.mu.Unlock()
		s.budget.releaseLocked(ctx, r.memoryTok.toRelease)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.signalBudgetIfNoRequestsInProgressLocked()
	} else {
		__antithesis_instrumentation__.Notify(499124)
	}
}

type Streamer struct {
	distSender *kvcoord.DistSender
	stopper    *stop.Stopper

	mode          OperationMode
	hints         Hints
	maxKeysPerRow int32
	budget        *budget

	coordinator          workerCoordinator
	coordinatorStarted   bool
	coordinatorCtxCancel context.CancelFunc

	waitGroup sync.WaitGroup

	enqueueKeys []int

	requestsToServe requestsProvider

	results resultsBuffer

	mu struct {
		syncutil.Mutex

		avgResponseEstimator avgResponseEstimator

		numRangesLeftPerScanRequest []int

		numRequestsInFlight int

		done bool
	}
}

var streamerConcurrencyLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.streamer.concurrency_limit",
	"maximum number of asynchronous requests by a single streamer",
	max(128, int64(8*runtime.GOMAXPROCS(0))),
	settings.PositiveInt,
)

func max(a, b int64) int64 {
	__antithesis_instrumentation__.Notify(499125)
	if a > b {
		__antithesis_instrumentation__.Notify(499127)
		return a
	} else {
		__antithesis_instrumentation__.Notify(499128)
	}
	__antithesis_instrumentation__.Notify(499126)
	return b
}

func NewStreamer(
	distSender *kvcoord.DistSender,
	stopper *stop.Stopper,
	txn *kv.Txn,
	st *cluster.Settings,
	lockWaitPolicy lock.WaitPolicy,
	limitBytes int64,
	acc *mon.BoundAccount,
) *Streamer {
	__antithesis_instrumentation__.Notify(499129)
	if txn.Type() != kv.LeafTxn {
		__antithesis_instrumentation__.Notify(499131)
		panic(errors.AssertionFailedf("RootTxn is given to the Streamer"))
	} else {
		__antithesis_instrumentation__.Notify(499132)
	}
	__antithesis_instrumentation__.Notify(499130)
	s := &Streamer{
		distSender: distSender,
		stopper:    stopper,
		budget:     newBudget(acc, limitBytes),
	}
	s.coordinator = workerCoordinator{
		s:                      s,
		txn:                    txn,
		lockWaitPolicy:         lockWaitPolicy,
		requestAdmissionHeader: txn.AdmissionHeader(),
		responseAdmissionQ:     txn.DB().SQLKVResponseAdmissionQ,
	}

	s.coordinator.asyncSem = quotapool.NewIntPool(
		"single Streamer async concurrency",
		uint64(streamerConcurrencyLimit.Get(&st.SV)),
	)
	return s
}

func (s *Streamer) Init(
	mode OperationMode,
	hints Hints,
	maxKeysPerRow int,
	engine diskmap.Factory,
	diskMonitor *mon.BytesMonitor,
) {
	__antithesis_instrumentation__.Notify(499133)
	s.mode = mode
	if mode == OutOfOrder {
		__antithesis_instrumentation__.Notify(499136)
		s.requestsToServe = newOutOfOrderRequestsProvider()
		s.results = newOutOfOrderResultsBuffer(s.budget)
	} else {
		__antithesis_instrumentation__.Notify(499137)
		s.requestsToServe = newInOrderRequestsProvider()
		s.results = newInOrderResultsBuffer(s.budget, engine, diskMonitor)
	}
	__antithesis_instrumentation__.Notify(499134)
	if !hints.UniqueRequests {
		__antithesis_instrumentation__.Notify(499138)
		panic(errors.AssertionFailedf("only unique requests are currently supported"))
	} else {
		__antithesis_instrumentation__.Notify(499139)
	}
	__antithesis_instrumentation__.Notify(499135)
	s.hints = hints
	s.maxKeysPerRow = int32(maxKeysPerRow)
}

func (s *Streamer) Enqueue(
	ctx context.Context, reqs []roachpb.RequestUnion, enqueueKeys []int,
) (retErr error) {
	__antithesis_instrumentation__.Notify(499140)
	if !s.coordinatorStarted {
		__antithesis_instrumentation__.Notify(499152)
		var coordinatorCtx context.Context
		coordinatorCtx, s.coordinatorCtxCancel = s.stopper.WithCancelOnQuiesce(ctx)
		s.waitGroup.Add(1)
		if err := s.stopper.RunAsyncTaskEx(
			coordinatorCtx,
			stop.TaskOpts{
				TaskName: "streamer-coordinator",
				SpanOpt:  stop.ChildSpan,
			},
			s.coordinator.mainLoop,
		); err != nil {
			__antithesis_instrumentation__.Notify(499154)

			s.waitGroup.Done()
			return err
		} else {
			__antithesis_instrumentation__.Notify(499155)
		}
		__antithesis_instrumentation__.Notify(499153)
		s.coordinatorStarted = true
	} else {
		__antithesis_instrumentation__.Notify(499156)
	}
	__antithesis_instrumentation__.Notify(499141)

	defer func() {
		__antithesis_instrumentation__.Notify(499157)

		if retErr != nil {
			__antithesis_instrumentation__.Notify(499158)
			s.results.setError(retErr)
		} else {
			__antithesis_instrumentation__.Notify(499159)
		}
	}()
	__antithesis_instrumentation__.Notify(499142)

	if enqueueKeys != nil && func() bool {
		__antithesis_instrumentation__.Notify(499160)
		return len(enqueueKeys) != len(reqs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(499161)
		return errors.AssertionFailedf("invalid enqueueKeys: len(reqs) = %d, len(enqueueKeys) = %d", len(reqs), len(enqueueKeys))
	} else {
		__antithesis_instrumentation__.Notify(499162)
	}
	__antithesis_instrumentation__.Notify(499143)
	s.enqueueKeys = enqueueKeys

	if err := s.results.init(ctx, len(reqs)); err != nil {
		__antithesis_instrumentation__.Notify(499163)
		return err
	} else {
		__antithesis_instrumentation__.Notify(499164)
	}
	__antithesis_instrumentation__.Notify(499144)

	rs, err := keys.Range(reqs)
	if err != nil {
		__antithesis_instrumentation__.Notify(499165)
		return err
	} else {
		__antithesis_instrumentation__.Notify(499166)
	}
	__antithesis_instrumentation__.Notify(499145)

	var totalReqsMemUsage int64

	var requestsToServe []singleRangeBatch
	seekKey := rs.Key
	const scanDir = kvcoord.Ascending
	ri := kvcoord.MakeRangeIterator(s.distSender)
	ri.Seek(ctx, seekKey, scanDir)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(499167)
		return ri.Error()
	} else {
		__antithesis_instrumentation__.Notify(499168)
	}
	__antithesis_instrumentation__.Notify(499146)
	firstScanRequest := true
	streamerLocked := false
	defer func() {
		__antithesis_instrumentation__.Notify(499169)
		if streamerLocked {
			__antithesis_instrumentation__.Notify(499170)
			s.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(499171)
		}
	}()
	__antithesis_instrumentation__.Notify(499147)
	for ; ri.Valid(); ri.Seek(ctx, seekKey, scanDir) {
		__antithesis_instrumentation__.Notify(499172)

		singleRangeSpan, err := rs.Intersect(ri.Token().Desc())
		if err != nil {
			__antithesis_instrumentation__.Notify(499177)
			return err
		} else {
			__antithesis_instrumentation__.Notify(499178)
		}
		__antithesis_instrumentation__.Notify(499173)

		singleRangeReqs, positions, err := kvcoord.Truncate(reqs, singleRangeSpan)
		if err != nil {
			__antithesis_instrumentation__.Notify(499179)
			return err
		} else {
			__antithesis_instrumentation__.Notify(499180)
		}
		__antithesis_instrumentation__.Notify(499174)
		for _, pos := range positions {
			__antithesis_instrumentation__.Notify(499181)
			if _, isScan := reqs[pos].GetInner().(*roachpb.ScanRequest); isScan {
				__antithesis_instrumentation__.Notify(499182)
				if firstScanRequest {
					__antithesis_instrumentation__.Notify(499184)

					streamerLocked = true
					s.mu.Lock()
					if cap(s.mu.numRangesLeftPerScanRequest) < len(reqs) {
						__antithesis_instrumentation__.Notify(499185)
						s.mu.numRangesLeftPerScanRequest = make([]int, len(reqs))
					} else {
						__antithesis_instrumentation__.Notify(499186)

						s.mu.numRangesLeftPerScanRequest = s.mu.numRangesLeftPerScanRequest[:len(reqs)]
						for n := 0; n < len(s.mu.numRangesLeftPerScanRequest); {
							__antithesis_instrumentation__.Notify(499187)
							n += copy(s.mu.numRangesLeftPerScanRequest[n:], zeroIntSlice)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(499188)
				}
				__antithesis_instrumentation__.Notify(499183)
				s.mu.numRangesLeftPerScanRequest[pos]++
				firstScanRequest = false
			} else {
				__antithesis_instrumentation__.Notify(499189)
			}
		}
		__antithesis_instrumentation__.Notify(499175)

		r := singleRangeBatch{
			reqs:              singleRangeReqs,
			positions:         positions,
			reqsReservedBytes: requestsMemUsage(singleRangeReqs),
			priority:          positions[0],
		}
		totalReqsMemUsage += r.reqsReservedBytes

		if s.mode == OutOfOrder {
			__antithesis_instrumentation__.Notify(499190)

			sort.Sort(&r)
		} else {
			__antithesis_instrumentation__.Notify(499191)
		}
		__antithesis_instrumentation__.Notify(499176)

		requestsToServe = append(requestsToServe, r)

		seekKey, err = kvcoord.Next(reqs, ri.Desc().EndKey)
		rs.Key = seekKey
		if err != nil {
			__antithesis_instrumentation__.Notify(499192)
			return err
		} else {
			__antithesis_instrumentation__.Notify(499193)
		}
	}
	__antithesis_instrumentation__.Notify(499148)

	if streamerLocked {
		__antithesis_instrumentation__.Notify(499194)

		s.mu.Unlock()
		streamerLocked = false
	} else {
		__antithesis_instrumentation__.Notify(499195)
	}
	__antithesis_instrumentation__.Notify(499149)

	allowDebt := len(reqs) == 1
	if err = s.budget.consume(ctx, totalReqsMemUsage, allowDebt); err != nil {
		__antithesis_instrumentation__.Notify(499196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(499197)
	}
	__antithesis_instrumentation__.Notify(499150)

	if debug {
		__antithesis_instrumentation__.Notify(499198)
		fmt.Printf("enqueuing %s to serve\n", reqsToString(requestsToServe))
	} else {
		__antithesis_instrumentation__.Notify(499199)
	}
	__antithesis_instrumentation__.Notify(499151)
	s.requestsToServe.enqueue(requestsToServe)
	return nil
}

func (s *Streamer) GetResults(ctx context.Context) ([]Result, error) {
	__antithesis_instrumentation__.Notify(499200)
	for {
		__antithesis_instrumentation__.Notify(499201)
		results, allComplete, err := s.results.get(ctx)
		if len(results) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(499204)
			return allComplete == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(499205)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(499206)
			if debug {
				__antithesis_instrumentation__.Notify(499208)
				if len(results) > 0 {
					__antithesis_instrumentation__.Notify(499209)
					fmt.Printf("returning %s to the client\n", resultsToString(results))
				} else {
					__antithesis_instrumentation__.Notify(499210)
					suffix := "all requests have been responded to"
					if !allComplete {
						__antithesis_instrumentation__.Notify(499212)
						suffix = fmt.Sprintf("%v", err)
					} else {
						__antithesis_instrumentation__.Notify(499213)
					}
					__antithesis_instrumentation__.Notify(499211)
					fmt.Printf("returning no results to the client because %s\n", suffix)
				}
			} else {
				__antithesis_instrumentation__.Notify(499214)
			}
			__antithesis_instrumentation__.Notify(499207)
			return results, err
		} else {
			__antithesis_instrumentation__.Notify(499215)
		}
		__antithesis_instrumentation__.Notify(499202)
		if debug {
			__antithesis_instrumentation__.Notify(499216)
			fmt.Println("client blocking to wait for results")
		} else {
			__antithesis_instrumentation__.Notify(499217)
		}
		__antithesis_instrumentation__.Notify(499203)
		s.results.wait()

		if err = ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(499218)
			s.results.setError(err)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(499219)
		}
	}
}

func (s *Streamer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(499220)
	if s.coordinatorStarted {
		__antithesis_instrumentation__.Notify(499222)
		s.coordinatorCtxCancel()
		s.mu.Lock()
		s.mu.done = true
		s.mu.Unlock()
		s.requestsToServe.close()
		s.results.close(ctx)

		s.budget.mu.waitForBudget.Signal()
	} else {
		__antithesis_instrumentation__.Notify(499223)
	}
	__antithesis_instrumentation__.Notify(499221)
	s.waitGroup.Wait()
	*s = Streamer{}
}

func (s *Streamer) getNumRequestsInProgress() int {
	__antithesis_instrumentation__.Notify(499224)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.numRequestsInFlight + s.results.numUnreleased()
}

func (s *Streamer) signalBudgetIfNoRequestsInProgressLocked() {
	__antithesis_instrumentation__.Notify(499225)
	s.mu.AssertHeld()
	if s.mu.numRequestsInFlight == 0 && func() bool {
		__antithesis_instrumentation__.Notify(499226)
		return s.results.numUnreleased() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(499227)
		s.budget.mu.waitForBudget.Signal()
	} else {
		__antithesis_instrumentation__.Notify(499228)
	}
}

func (s *Streamer) adjustNumRequestsInFlight(delta int) {
	__antithesis_instrumentation__.Notify(499229)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.numRequestsInFlight += delta
	s.signalBudgetIfNoRequestsInProgressLocked()
}

type workerCoordinator struct {
	s              *Streamer
	txn            *kv.Txn
	lockWaitPolicy lock.WaitPolicy

	asyncSem *quotapool.IntPool

	requestAdmissionHeader roachpb.AdmissionHeader
	responseAdmissionQ     *admission.WorkQueue
}

func (w *workerCoordinator) mainLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(499230)
	defer w.s.waitGroup.Done()
	for {
		__antithesis_instrumentation__.Notify(499231)
		if err := w.waitForRequests(ctx); err != nil {
			__antithesis_instrumentation__.Notify(499238)
			w.s.results.setError(err)
			return
		} else {
			__antithesis_instrumentation__.Notify(499239)
		}
		__antithesis_instrumentation__.Notify(499232)

		var atLeastBytes int64

		spillingPriority := math.MaxInt64
		w.s.requestsToServe.Lock()
		if !w.s.requestsToServe.emptyLocked() {
			__antithesis_instrumentation__.Notify(499240)

			atLeastBytes = w.s.requestsToServe.firstLocked().minTargetBytes

			spillingPriority = w.s.requestsToServe.firstLocked().priority
		} else {
			__antithesis_instrumentation__.Notify(499241)
		}
		__antithesis_instrumentation__.Notify(499233)
		w.s.requestsToServe.Unlock()

		avgResponseSize, shouldExit := w.getAvgResponseSize()
		if shouldExit {
			__antithesis_instrumentation__.Notify(499242)
			return
		} else {
			__antithesis_instrumentation__.Notify(499243)
		}
		__antithesis_instrumentation__.Notify(499234)

		if atLeastBytes == 0 {
			__antithesis_instrumentation__.Notify(499244)
			atLeastBytes = avgResponseSize
		} else {
			__antithesis_instrumentation__.Notify(499245)
		}
		__antithesis_instrumentation__.Notify(499235)

		shouldExit = w.waitUntilEnoughBudget(ctx, atLeastBytes, spillingPriority)
		if shouldExit {
			__antithesis_instrumentation__.Notify(499246)
			return
		} else {
			__antithesis_instrumentation__.Notify(499247)
		}
		__antithesis_instrumentation__.Notify(499236)

		maxNumRequestsToIssue, shouldExit := w.getMaxNumRequestsToIssue(ctx)
		if shouldExit {
			__antithesis_instrumentation__.Notify(499248)
			return
		} else {
			__antithesis_instrumentation__.Notify(499249)
		}
		__antithesis_instrumentation__.Notify(499237)

		err := w.issueRequestsForAsyncProcessing(ctx, maxNumRequestsToIssue, avgResponseSize)
		if err != nil {
			__antithesis_instrumentation__.Notify(499250)
			w.s.results.setError(err)
			return
		} else {
			__antithesis_instrumentation__.Notify(499251)
		}
	}
}

func (w *workerCoordinator) waitForRequests(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(499252)
	w.s.requestsToServe.Lock()
	defer w.s.requestsToServe.Unlock()
	if w.s.requestsToServe.emptyLocked() {
		__antithesis_instrumentation__.Notify(499254)
		w.s.requestsToServe.waitLocked()

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(499257)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(499258)
		}
		__antithesis_instrumentation__.Notify(499255)
		w.s.mu.Lock()
		shouldExit := w.s.results.error() != nil || func() bool {
			__antithesis_instrumentation__.Notify(499259)
			return w.s.mu.done == true
		}() == true
		w.s.mu.Unlock()
		if shouldExit {
			__antithesis_instrumentation__.Notify(499260)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(499261)
		}
		__antithesis_instrumentation__.Notify(499256)
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(499262)
			if w.s.requestsToServe.emptyLocked() {
				__antithesis_instrumentation__.Notify(499263)
				panic(errors.AssertionFailedf("unexpectedly zero requests to serve after waiting "))
			} else {
				__antithesis_instrumentation__.Notify(499264)
			}
		} else {
			__antithesis_instrumentation__.Notify(499265)
		}
	} else {
		__antithesis_instrumentation__.Notify(499266)
	}
	__antithesis_instrumentation__.Notify(499253)
	return nil
}

func (w *workerCoordinator) getAvgResponseSize() (avgResponseSize int64, shouldExit bool) {
	__antithesis_instrumentation__.Notify(499267)
	w.s.mu.Lock()
	defer w.s.mu.Unlock()
	avgResponseSize = w.s.mu.avgResponseEstimator.getAvgResponseSize()
	shouldExit = w.s.results.error() != nil || func() bool {
		__antithesis_instrumentation__.Notify(499268)
		return w.s.mu.done == true
	}() == true
	return avgResponseSize, shouldExit
}

func (w *workerCoordinator) waitUntilEnoughBudget(
	ctx context.Context, atLeastBytes int64, spillingPriority int,
) (shouldExit bool) {
	__antithesis_instrumentation__.Notify(499269)
	w.s.budget.mu.Lock()
	defer w.s.budget.mu.Unlock()
	for w.s.budget.limitBytes-w.s.budget.mu.acc.Used() < atLeastBytes {
		__antithesis_instrumentation__.Notify(499271)

		if ok, err := w.s.results.spill(
			ctx, atLeastBytes-(w.s.budget.limitBytes-w.s.budget.mu.acc.Used()), spillingPriority,
		); err != nil {
			__antithesis_instrumentation__.Notify(499275)
			w.s.results.setError(err)
			return true
		} else {
			__antithesis_instrumentation__.Notify(499276)
			if ok {
				__antithesis_instrumentation__.Notify(499277)

				return false
			} else {
				__antithesis_instrumentation__.Notify(499278)
			}
		}
		__antithesis_instrumentation__.Notify(499272)

		if w.s.getNumRequestsInProgress() == 0 {
			__antithesis_instrumentation__.Notify(499279)

			return false
		} else {
			__antithesis_instrumentation__.Notify(499280)
		}
		__antithesis_instrumentation__.Notify(499273)
		if debug {
			__antithesis_instrumentation__.Notify(499281)
			fmt.Printf(
				"waiting for budget to free up: atLeastBytes %d, available %d\n",
				atLeastBytes, w.s.budget.limitBytes-w.s.budget.mu.acc.Used(),
			)
		} else {
			__antithesis_instrumentation__.Notify(499282)
		}
		__antithesis_instrumentation__.Notify(499274)

		w.s.budget.mu.waitForBudget.Wait()

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(499283)
			w.s.results.setError(ctx.Err())
			return true
		} else {
			__antithesis_instrumentation__.Notify(499284)
		}
	}
	__antithesis_instrumentation__.Notify(499270)
	return false
}

func (w *workerCoordinator) getMaxNumRequestsToIssue(ctx context.Context) (_ int, shouldExit bool) {
	__antithesis_instrumentation__.Notify(499285)

	q := w.asyncSem.ApproximateQuota()
	if q > 0 {
		__antithesis_instrumentation__.Notify(499288)
		return int(q), false
	} else {
		__antithesis_instrumentation__.Notify(499289)
	}
	__antithesis_instrumentation__.Notify(499286)

	alloc, err := w.asyncSem.Acquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(499290)
		w.s.results.setError(err)
		return 0, true
	} else {
		__antithesis_instrumentation__.Notify(499291)
	}
	__antithesis_instrumentation__.Notify(499287)
	alloc.Release()
	return 1, false
}

func (w *workerCoordinator) issueRequestsForAsyncProcessing(
	ctx context.Context, maxNumRequestsToIssue int, avgResponseSize int64,
) error {
	__antithesis_instrumentation__.Notify(499292)
	w.s.requestsToServe.Lock()
	defer w.s.requestsToServe.Unlock()
	w.s.budget.mu.Lock()
	defer w.s.budget.mu.Unlock()

	headOfLine := w.s.getNumRequestsInProgress() == 0
	var budgetIsExhausted bool
	for !w.s.requestsToServe.emptyLocked() && func() bool {
		__antithesis_instrumentation__.Notify(499294)
		return maxNumRequestsToIssue > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(499295)
		return !budgetIsExhausted == true
	}() == true {
		__antithesis_instrumentation__.Notify(499296)
		singleRangeReqs := w.s.requestsToServe.firstLocked()
		availableBudget := w.s.budget.limitBytes - w.s.budget.mu.acc.Used()

		minAcceptableBudget := singleRangeReqs.minTargetBytes
		if minAcceptableBudget == 0 {
			__antithesis_instrumentation__.Notify(499303)
			minAcceptableBudget = avgResponseSize
		} else {
			__antithesis_instrumentation__.Notify(499304)
		}
		__antithesis_instrumentation__.Notify(499297)
		if availableBudget < minAcceptableBudget {
			__antithesis_instrumentation__.Notify(499305)
			if !headOfLine {
				__antithesis_instrumentation__.Notify(499307)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(499308)
			}
			__antithesis_instrumentation__.Notify(499306)
			budgetIsExhausted = true
			if availableBudget < 1 {
				__antithesis_instrumentation__.Notify(499309)

				availableBudget = 1
			} else {
				__antithesis_instrumentation__.Notify(499310)
			}
		} else {
			__antithesis_instrumentation__.Notify(499311)
		}
		__antithesis_instrumentation__.Notify(499298)

		targetBytes := int64(len(singleRangeReqs.reqs)) * avgResponseSize

		if targetBytes < singleRangeReqs.minTargetBytes {
			__antithesis_instrumentation__.Notify(499312)
			targetBytes = singleRangeReqs.minTargetBytes
		} else {
			__antithesis_instrumentation__.Notify(499313)
		}
		__antithesis_instrumentation__.Notify(499299)
		if targetBytes > availableBudget {
			__antithesis_instrumentation__.Notify(499314)

			targetBytes = availableBudget
		} else {
			__antithesis_instrumentation__.Notify(499315)
		}
		__antithesis_instrumentation__.Notify(499300)
		if err := w.s.budget.consumeLocked(ctx, targetBytes, headOfLine); err != nil {
			__antithesis_instrumentation__.Notify(499316)

			if !headOfLine {
				__antithesis_instrumentation__.Notify(499318)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(499319)
			}
			__antithesis_instrumentation__.Notify(499317)

			return err
		} else {
			__antithesis_instrumentation__.Notify(499320)
		}
		__antithesis_instrumentation__.Notify(499301)
		if debug {
			__antithesis_instrumentation__.Notify(499321)
			fmt.Printf(
				"issuing an async request for positions %v, targetBytes=%d, headOfLine=%t\n",
				singleRangeReqs.positions, targetBytes, headOfLine,
			)
		} else {
			__antithesis_instrumentation__.Notify(499322)
		}
		__antithesis_instrumentation__.Notify(499302)
		w.performRequestAsync(ctx, singleRangeReqs, targetBytes, headOfLine)
		w.s.requestsToServe.removeFirstLocked()
		maxNumRequestsToIssue--
		headOfLine = false
	}
	__antithesis_instrumentation__.Notify(499293)
	return nil
}

func (w *workerCoordinator) asyncRequestCleanup(budgetMuAlreadyLocked bool) {
	__antithesis_instrumentation__.Notify(499323)
	if !budgetMuAlreadyLocked {
		__antithesis_instrumentation__.Notify(499325)

		w.s.budget.mu.Lock()
		defer w.s.budget.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(499326)
		w.s.budget.mu.AssertHeld()
	}
	__antithesis_instrumentation__.Notify(499324)
	w.s.adjustNumRequestsInFlight(-1)
	w.s.waitGroup.Done()
}

const AsyncRequestOp = "streamer-lookup-async"

func (w *workerCoordinator) performRequestAsync(
	ctx context.Context, req singleRangeBatch, targetBytes int64, headOfLine bool,
) {
	__antithesis_instrumentation__.Notify(499327)
	w.s.waitGroup.Add(1)
	w.s.adjustNumRequestsInFlight(1)
	if err := w.s.stopper.RunAsyncTaskEx(
		ctx,
		stop.TaskOpts{
			TaskName: AsyncRequestOp,
			SpanOpt:  stop.ChildSpan,

			Sem: w.asyncSem,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(499328)
			defer w.asyncRequestCleanup(false)
			var ba roachpb.BatchRequest
			ba.Header.WaitPolicy = w.lockWaitPolicy
			ba.Header.TargetBytes = targetBytes
			ba.Header.AllowEmpty = !headOfLine
			ba.Header.WholeRowsOfSize = w.s.maxKeysPerRow

			ba.AdmissionHeader = w.requestAdmissionHeader

			ba.AdmissionHeader.NoMemoryReservedAtSource = false
			ba.Requests = req.reqs

			br, err := w.txn.Send(ctx, ba)
			if err != nil {
				__antithesis_instrumentation__.Notify(499332)

				w.s.results.setError(err.GoError())
				return
			} else {
				__antithesis_instrumentation__.Notify(499333)
			}
			__antithesis_instrumentation__.Notify(499329)

			memoryFootprintBytes, resumeReqsMemUsage, numIncompleteGets, numIncompleteScans := calculateFootprint(req, br)

			respOverestimate := targetBytes - memoryFootprintBytes
			reqOveraccounted := req.reqsReservedBytes - resumeReqsMemUsage
			overaccountedTotal := respOverestimate + reqOveraccounted
			if overaccountedTotal >= 0 {
				__antithesis_instrumentation__.Notify(499334)
				w.s.budget.release(ctx, overaccountedTotal)
			} else {
				__antithesis_instrumentation__.Notify(499335)

				toConsume := -overaccountedTotal
				if err := w.s.budget.consume(ctx, toConsume, headOfLine); err != nil {
					__antithesis_instrumentation__.Notify(499336)
					w.s.budget.release(ctx, targetBytes)
					if !headOfLine {
						__antithesis_instrumentation__.Notify(499338)

						w.s.requestsToServe.add(req)
						return
					} else {
						__antithesis_instrumentation__.Notify(499339)
					}
					__antithesis_instrumentation__.Notify(499337)

					w.s.results.setError(err)
					return
				} else {
					__antithesis_instrumentation__.Notify(499340)
				}
			}
			__antithesis_instrumentation__.Notify(499330)

			if br != nil && func() bool {
				__antithesis_instrumentation__.Notify(499341)
				return w.responseAdmissionQ != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(499342)
				responseAdmission := admission.WorkInfo{
					TenantID:   roachpb.SystemTenantID,
					Priority:   admission.WorkPriority(w.requestAdmissionHeader.Priority),
					CreateTime: w.requestAdmissionHeader.CreateTime,
				}
				if _, err := w.responseAdmissionQ.Admit(ctx, responseAdmission); err != nil {
					__antithesis_instrumentation__.Notify(499343)
					w.s.results.setError(err)
					return
				} else {
					__antithesis_instrumentation__.Notify(499344)
				}
			} else {
				__antithesis_instrumentation__.Notify(499345)
			}
			__antithesis_instrumentation__.Notify(499331)

			if err := w.processSingleRangeResults(
				req, br, memoryFootprintBytes, resumeReqsMemUsage,
				numIncompleteGets, numIncompleteScans,
			); err != nil {
				__antithesis_instrumentation__.Notify(499346)
				w.s.results.setError(err)
			} else {
				__antithesis_instrumentation__.Notify(499347)
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(499348)

		w.asyncRequestCleanup(true)
		w.s.results.setError(err)
	} else {
		__antithesis_instrumentation__.Notify(499349)
	}
}

func calculateFootprint(
	req singleRangeBatch, br *roachpb.BatchResponse,
) (
	memoryFootprintBytes int64,
	resumeReqsMemUsage int64,
	numIncompleteGets, numIncompleteScans int,
) {
	__antithesis_instrumentation__.Notify(499350)

	var getRequestScratch roachpb.GetRequest
	var scanRequestScratch roachpb.ScanRequest
	for i, resp := range br.Responses {
		__antithesis_instrumentation__.Notify(499352)
		reply := resp.GetInner()
		switch req.reqs[i].GetInner().(type) {
		case *roachpb.GetRequest:
			__antithesis_instrumentation__.Notify(499353)
			get := reply.(*roachpb.GetResponse)
			if get.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(499356)

				getRequestScratch.SetSpan(*get.ResumeSpan)
				resumeReqsMemUsage += int64(getRequestScratch.Size())
				numIncompleteGets++
			} else {
				__antithesis_instrumentation__.Notify(499357)

				memoryFootprintBytes += getResponseSize(get)
			}
		case *roachpb.ScanRequest:
			__antithesis_instrumentation__.Notify(499354)
			scan := reply.(*roachpb.ScanResponse)
			if len(scan.BatchResponses) > 0 {
				__antithesis_instrumentation__.Notify(499358)
				memoryFootprintBytes += scanResponseSize(scan)
			} else {
				__antithesis_instrumentation__.Notify(499359)
			}
			__antithesis_instrumentation__.Notify(499355)
			if scan.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(499360)

				scanRequestScratch.SetSpan(*scan.ResumeSpan)
				resumeReqsMemUsage += int64(scanRequestScratch.Size())
				numIncompleteScans++
			} else {
				__antithesis_instrumentation__.Notify(499361)
			}
		}
	}
	__antithesis_instrumentation__.Notify(499351)

	resumeReqsMemUsage += requestUnionOverhead * int64(numIncompleteGets+numIncompleteScans)
	return memoryFootprintBytes, resumeReqsMemUsage, numIncompleteGets, numIncompleteScans
}

func (w *workerCoordinator) processSingleRangeResults(
	req singleRangeBatch,
	br *roachpb.BatchResponse,
	memoryFootprintBytes int64,
	resumeReqsMemUsage int64,
	numIncompleteGets, numIncompleteScans int,
) error {
	__antithesis_instrumentation__.Notify(499362)
	numIncompleteRequests := numIncompleteGets + numIncompleteScans
	var resumeReq singleRangeBatch

	resumeReq.reqs = make([]roachpb.RequestUnion, numIncompleteRequests)

	resumeReq.positions = req.positions[:numIncompleteRequests]

	resumeReq.reqsReservedBytes = resumeReqsMemUsage
	resumeReq.priority = math.MaxInt
	gets := make([]struct {
		req   roachpb.GetRequest
		union roachpb.RequestUnion_Get
	}, numIncompleteGets)
	scans := make([]struct {
		req   roachpb.ScanRequest
		union roachpb.RequestUnion_Scan
	}, numIncompleteScans)
	var results []Result
	var hasNonEmptyScanResponse bool
	var resumeReqIdx int

	var memoryTokensBytes int64
	for i, resp := range br.Responses {
		__antithesis_instrumentation__.Notify(499367)
		position := req.positions[i]
		enqueueKey := position
		if w.s.enqueueKeys != nil {
			__antithesis_instrumentation__.Notify(499369)
			enqueueKey = w.s.enqueueKeys[position]
		} else {
			__antithesis_instrumentation__.Notify(499370)
		}
		__antithesis_instrumentation__.Notify(499368)
		reply := resp.GetInner()
		switch origRequest := req.reqs[i].GetInner().(type) {
		case *roachpb.GetRequest:
			__antithesis_instrumentation__.Notify(499371)
			get := reply.(*roachpb.GetResponse)
			if get.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(499376)

				newGet := gets[0]
				gets = gets[1:]
				newGet.req.SetSpan(*get.ResumeSpan)
				newGet.req.KeyLocking = origRequest.KeyLocking
				newGet.union.Get = &newGet.req
				resumeReq.reqs[resumeReqIdx].Value = &newGet.union
				resumeReq.positions[resumeReqIdx] = req.positions[i]
				if resumeReq.minTargetBytes == 0 {
					__antithesis_instrumentation__.Notify(499379)
					resumeReq.minTargetBytes = get.ResumeNextBytes
				} else {
					__antithesis_instrumentation__.Notify(499380)
				}
				__antithesis_instrumentation__.Notify(499377)
				if position < resumeReq.priority {
					__antithesis_instrumentation__.Notify(499381)
					resumeReq.priority = position
				} else {
					__antithesis_instrumentation__.Notify(499382)
				}
				__antithesis_instrumentation__.Notify(499378)
				resumeReqIdx++
			} else {
				__antithesis_instrumentation__.Notify(499383)

				if get.IntentValue != nil {
					__antithesis_instrumentation__.Notify(499385)
					return errors.AssertionFailedf(
						"unexpectedly got an IntentValue back from a SQL GetRequest %v", *get.IntentValue,
					)
				} else {
					__antithesis_instrumentation__.Notify(499386)
				}
				__antithesis_instrumentation__.Notify(499384)
				result := Result{
					GetResp: get,

					EnqueueKeysSatisfied: []int{enqueueKey},
					position:             position,
				}
				result.memoryTok.streamer = w.s
				result.memoryTok.toRelease = getResponseSize(get)
				memoryTokensBytes += result.memoryTok.toRelease
				results = append(results, result)
			}

		case *roachpb.ScanRequest:
			__antithesis_instrumentation__.Notify(499372)
			scan := reply.(*roachpb.ScanResponse)
			if len(scan.Rows) > 0 {
				__antithesis_instrumentation__.Notify(499387)
				return errors.AssertionFailedf(
					"unexpectedly got a ScanResponse using KEY_VALUES response format",
				)
			} else {
				__antithesis_instrumentation__.Notify(499388)
			}
			__antithesis_instrumentation__.Notify(499373)
			if len(scan.IntentRows) > 0 {
				__antithesis_instrumentation__.Notify(499389)
				return errors.AssertionFailedf(
					"unexpectedly got a ScanResponse with non-nil IntentRows",
				)
			} else {
				__antithesis_instrumentation__.Notify(499390)
			}
			__antithesis_instrumentation__.Notify(499374)
			if len(scan.BatchResponses) > 0 || func() bool {
				__antithesis_instrumentation__.Notify(499391)
				return scan.ResumeSpan == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(499392)

				result := Result{

					EnqueueKeysSatisfied: []int{enqueueKey},
					position:             position,
				}
				result.memoryTok.streamer = w.s
				result.memoryTok.toRelease = scanResponseSize(scan)
				memoryTokensBytes += result.memoryTok.toRelease
				result.ScanResp.ScanResponse = scan

				results = append(results, result)
				hasNonEmptyScanResponse = true
			} else {
				__antithesis_instrumentation__.Notify(499393)
			}
			__antithesis_instrumentation__.Notify(499375)
			if scan.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(499394)

				newScan := scans[0]
				scans = scans[1:]
				newScan.req.SetSpan(*scan.ResumeSpan)
				newScan.req.ScanFormat = roachpb.BATCH_RESPONSE
				newScan.req.KeyLocking = origRequest.KeyLocking
				newScan.union.Scan = &newScan.req
				resumeReq.reqs[resumeReqIdx].Value = &newScan.union
				resumeReq.positions[resumeReqIdx] = req.positions[i]
				if resumeReq.minTargetBytes == 0 {
					__antithesis_instrumentation__.Notify(499397)
					resumeReq.minTargetBytes = scan.ResumeNextBytes
				} else {
					__antithesis_instrumentation__.Notify(499398)
				}
				__antithesis_instrumentation__.Notify(499395)
				if position < resumeReq.priority {
					__antithesis_instrumentation__.Notify(499399)
					resumeReq.priority = position
				} else {
					__antithesis_instrumentation__.Notify(499400)
				}
				__antithesis_instrumentation__.Notify(499396)
				resumeReqIdx++
			} else {
				__antithesis_instrumentation__.Notify(499401)
			}
		}
	}
	__antithesis_instrumentation__.Notify(499363)

	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(499402)
		if memoryFootprintBytes != memoryTokensBytes {
			__antithesis_instrumentation__.Notify(499403)
			panic(errors.AssertionFailedf(
				"different calculation of memory footprint\ncalculateFootprint: %d bytes\n"+
					"processSingleRangeResults: %d bytes", memoryFootprintBytes, memoryTokensBytes,
			))
		} else {
			__antithesis_instrumentation__.Notify(499404)
		}
	} else {
		__antithesis_instrumentation__.Notify(499405)
	}
	__antithesis_instrumentation__.Notify(499364)

	if len(results) > 0 {
		__antithesis_instrumentation__.Notify(499406)
		w.finalizeSingleRangeResults(
			results, memoryFootprintBytes, hasNonEmptyScanResponse,
		)
	} else {
		__antithesis_instrumentation__.Notify(499407)

		if req.minTargetBytes != 0 {
			__antithesis_instrumentation__.Notify(499409)

			if resumeReq.minTargetBytes <= req.minTargetBytes {
				__antithesis_instrumentation__.Notify(499410)

				resumeReq.minTargetBytes = 2 * req.minTargetBytes
			} else {
				__antithesis_instrumentation__.Notify(499411)
			}
		} else {
			__antithesis_instrumentation__.Notify(499412)
		}
		__antithesis_instrumentation__.Notify(499408)
		if debug {
			__antithesis_instrumentation__.Notify(499413)
			fmt.Printf(
				"request for positions %v came back empty, original minTargetBytes=%d, "+
					"resumeReq.minTargetBytes=%d\n", req.positions, req.minTargetBytes, resumeReq.minTargetBytes,
			)
		} else {
			__antithesis_instrumentation__.Notify(499414)
		}
	}
	__antithesis_instrumentation__.Notify(499365)

	if len(resumeReq.reqs) > 0 {
		__antithesis_instrumentation__.Notify(499415)
		w.s.requestsToServe.add(resumeReq)
	} else {
		__antithesis_instrumentation__.Notify(499416)
	}
	__antithesis_instrumentation__.Notify(499366)

	return nil
}

func (w *workerCoordinator) finalizeSingleRangeResults(
	results []Result, actualMemoryReservation int64, hasNonEmptyScanResponse bool,
) {
	__antithesis_instrumentation__.Notify(499417)
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(499421)
		if len(results) == 0 {
			__antithesis_instrumentation__.Notify(499422)
			panic(errors.AssertionFailedf("finalizeSingleRangeResults is called with no results"))
		} else {
			__antithesis_instrumentation__.Notify(499423)
		}
	} else {
		__antithesis_instrumentation__.Notify(499424)
	}
	__antithesis_instrumentation__.Notify(499418)
	w.s.mu.Lock()
	defer w.s.mu.Unlock()

	if hasNonEmptyScanResponse {
		__antithesis_instrumentation__.Notify(499425)
		for i := range results {
			__antithesis_instrumentation__.Notify(499426)
			if results[i].ScanResp.ScanResponse != nil {
				__antithesis_instrumentation__.Notify(499427)
				if results[i].ScanResp.ResumeSpan == nil {
					__antithesis_instrumentation__.Notify(499428)

					w.s.mu.numRangesLeftPerScanRequest[results[i].position]--
					if w.s.mu.numRangesLeftPerScanRequest[results[i].position] == 0 {
						__antithesis_instrumentation__.Notify(499429)

						results[i].ScanResp.Complete = true
					} else {
						__antithesis_instrumentation__.Notify(499430)
					}
				} else {
					__antithesis_instrumentation__.Notify(499431)

					results[i].ScanResp.ResumeSpan = nil
				}
			} else {
				__antithesis_instrumentation__.Notify(499432)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(499433)
	}
	__antithesis_instrumentation__.Notify(499419)

	w.s.mu.avgResponseEstimator.update(actualMemoryReservation, int64(len(results)))
	if debug {
		__antithesis_instrumentation__.Notify(499434)
		fmt.Printf("created %s with total size %d\n", resultsToString(results), actualMemoryReservation)
	} else {
		__antithesis_instrumentation__.Notify(499435)
	}
	__antithesis_instrumentation__.Notify(499420)
	w.s.results.add(results)
}

var zeroIntSlice []int

func init() {
	zeroIntSlice = make([]int, 1<<10)
}

const requestUnionOverhead = int64(unsafe.Sizeof(roachpb.RequestUnion{}))

func requestsMemUsage(reqs []roachpb.RequestUnion) int64 {
	__antithesis_instrumentation__.Notify(499436)

	memUsage := requestUnionOverhead * int64(cap(reqs))

	for _, r := range reqs {
		__antithesis_instrumentation__.Notify(499438)
		memUsage += int64(r.Size())
	}
	__antithesis_instrumentation__.Notify(499437)
	return memUsage
}

func getResponseSize(get *roachpb.GetResponse) int64 {
	__antithesis_instrumentation__.Notify(499439)
	if get.Value == nil {
		__antithesis_instrumentation__.Notify(499441)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(499442)
	}
	__antithesis_instrumentation__.Notify(499440)
	return int64(len(get.Value.RawBytes))
}

func scanResponseSize(scan *roachpb.ScanResponse) int64 {
	__antithesis_instrumentation__.Notify(499443)
	return scan.NumBytes
}
