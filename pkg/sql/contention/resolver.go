package contention

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sort"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type resolverQueue interface {
	enqueue([]contentionpb.ExtendedContentionEvent)

	dequeue(context.Context) ([]contentionpb.ExtendedContentionEvent, error)
}

const (
	retryBudgetForMissingResult = uint32(1)

	retryBudgetForRPCFailure = uint32(3)

	retryBudgetForTxnInProgress = uint32(math.MaxUint32)
)

type ResolverEndpoint func(context.Context, *serverpb.TxnIDResolutionRequest) (*serverpb.TxnIDResolutionResponse, error)

type resolverQueueImpl struct {
	mu struct {
		syncutil.RWMutex

		unresolvedEvents []contentionpb.ExtendedContentionEvent
		resolvedEvents   []contentionpb.ExtendedContentionEvent

		remainingRetries map[uint64]uint32
	}

	resolverEndpoint ResolverEndpoint

	metrics *Metrics
}

var _ resolverQueue = &resolverQueueImpl{}

func newResolver(endpoint ResolverEndpoint, metrics *Metrics, sizeHint int) *resolverQueueImpl {
	__antithesis_instrumentation__.Notify(459242)
	s := &resolverQueueImpl{
		resolverEndpoint: endpoint,
		metrics:          metrics,
	}

	s.mu.unresolvedEvents = make([]contentionpb.ExtendedContentionEvent, 0, sizeHint)
	s.mu.resolvedEvents = make([]contentionpb.ExtendedContentionEvent, 0, sizeHint)
	s.mu.remainingRetries = make(map[uint64]uint32, sizeHint)

	return s
}

func (q *resolverQueueImpl) enqueue(block []contentionpb.ExtendedContentionEvent) {
	__antithesis_instrumentation__.Notify(459243)
	q.mu.Lock()
	defer q.mu.Unlock()

	q.mu.unresolvedEvents = append(q.mu.unresolvedEvents, block...)
	q.metrics.ResolverQueueSize.Inc(int64(len(block)))
}

func (q *resolverQueueImpl) dequeue(
	ctx context.Context,
) ([]contentionpb.ExtendedContentionEvent, error) {
	__antithesis_instrumentation__.Notify(459244)
	q.mu.Lock()
	defer q.mu.Unlock()

	err := q.resolveLocked(ctx)
	result := q.mu.resolvedEvents
	q.mu.resolvedEvents = q.mu.resolvedEvents[:0]

	q.metrics.ResolverQueueSize.Dec(int64(len(result)))

	return result, err
}

func (q *resolverQueueImpl) resolveLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(459245)
	queueCpy := make([]contentionpb.ExtendedContentionEvent, len(q.mu.unresolvedEvents))
	copy(queueCpy, q.mu.unresolvedEvents)

	q.mu.unresolvedEvents = q.mu.unresolvedEvents[:0]

	sort.Slice(queueCpy, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(459248)
		return queueCpy[i].BlockingEvent.TxnMeta.CoordinatorNodeID <
			queueCpy[j].BlockingEvent.TxnMeta.CoordinatorNodeID
	})
	__antithesis_instrumentation__.Notify(459246)

	currentBatch, remaining := readUntilNextCoordinatorID(queueCpy)
	var allErrors error
	for len(currentBatch) > 0 {
		__antithesis_instrumentation__.Notify(459249)

		blockingTxnIDsReq, waitingTxnIDsReq := makeRPCRequestsFromBatch(currentBatch)

		blockingTxnIDsResp, err := q.resolverEndpoint(ctx, blockingTxnIDsReq)
		if err != nil {
			__antithesis_instrumentation__.Notify(459253)
			allErrors = errors.CombineErrors(allErrors, err)
		} else {
			__antithesis_instrumentation__.Notify(459254)
		}
		__antithesis_instrumentation__.Notify(459250)

		waitingTxnIDsResp, err := q.resolverEndpoint(ctx, waitingTxnIDsReq)
		if err != nil {
			__antithesis_instrumentation__.Notify(459255)
			allErrors = errors.CombineErrors(allErrors, err)
		} else {
			__antithesis_instrumentation__.Notify(459256)
		}
		__antithesis_instrumentation__.Notify(459251)

		resolvedBlockingTxnIDs, inProgressBlockingTxnIDs := extractResolvedAndInProgressTxnIDs(blockingTxnIDsResp)
		resolvedWaitingTxnIDs, inProgressWaitingTxnIDs := extractResolvedAndInProgressTxnIDs(waitingTxnIDsResp)

		for _, event := range currentBatch {
			__antithesis_instrumentation__.Notify(459257)
			needToRetryDueToBlockingTxnID, initialRetryBudgetDueToBlockingTxnID :=
				maybeUpdateTxnFingerprintID(
					event.BlockingEvent.TxnMeta.ID,
					&event.BlockingTxnFingerprintID,
					resolvedBlockingTxnIDs,
					inProgressBlockingTxnIDs,
				)

			needToRetryDueToWaitingTxnID, initialRetryBudgetDueToWaitingTxnID :=
				maybeUpdateTxnFingerprintID(
					event.WaitingTxnID,
					&event.WaitingTxnFingerprintID,
					resolvedWaitingTxnIDs,
					inProgressWaitingTxnIDs,
				)

			initialRetryBudget := initialRetryBudgetDueToBlockingTxnID
			if initialRetryBudget < initialRetryBudgetDueToWaitingTxnID {
				__antithesis_instrumentation__.Notify(459259)
				initialRetryBudget = initialRetryBudgetDueToWaitingTxnID
			} else {
				__antithesis_instrumentation__.Notify(459260)
			}
			__antithesis_instrumentation__.Notify(459258)

			if needToRetryDueToBlockingTxnID || func() bool {
				__antithesis_instrumentation__.Notify(459261)
				return needToRetryDueToWaitingTxnID == true
			}() == true {
				__antithesis_instrumentation__.Notify(459262)
				q.maybeRequeueEventForRetryLocked(event, initialRetryBudget)
			} else {
				__antithesis_instrumentation__.Notify(459263)
				q.mu.resolvedEvents = append(q.mu.resolvedEvents, event)
				delete(q.mu.remainingRetries, event.Hash())
			}
		}
		__antithesis_instrumentation__.Notify(459252)

		currentBatch, remaining = readUntilNextCoordinatorID(remaining)
	}
	__antithesis_instrumentation__.Notify(459247)

	return allErrors
}

func maybeUpdateTxnFingerprintID(
	txnID uuid.UUID,
	existingTxnFingerprintID *roachpb.TransactionFingerprintID,
	resolvedTxnIDs, inProgressTxnIDs map[uuid.UUID]roachpb.TransactionFingerprintID,
) (needToRetry bool, initialRetryBudget uint32) {
	__antithesis_instrumentation__.Notify(459264)

	if *existingTxnFingerprintID != roachpb.InvalidTransactionFingerprintID {
		__antithesis_instrumentation__.Notify(459271)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(459272)
	}
	__antithesis_instrumentation__.Notify(459265)

	if uuid.Nil.Equal(txnID) {
		__antithesis_instrumentation__.Notify(459273)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(459274)
	}
	__antithesis_instrumentation__.Notify(459266)

	if resolvedTxnIDs == nil {
		__antithesis_instrumentation__.Notify(459275)
		return true, retryBudgetForRPCFailure
	} else {
		__antithesis_instrumentation__.Notify(459276)
	}
	__antithesis_instrumentation__.Notify(459267)

	if _, ok := inProgressTxnIDs[txnID]; ok {
		__antithesis_instrumentation__.Notify(459277)
		return true, retryBudgetForTxnInProgress
	} else {
		__antithesis_instrumentation__.Notify(459278)
	}
	__antithesis_instrumentation__.Notify(459268)

	if inProgressTxnIDs == nil {
		__antithesis_instrumentation__.Notify(459279)
		return true, retryBudgetForRPCFailure
	} else {
		__antithesis_instrumentation__.Notify(459280)
	}
	__antithesis_instrumentation__.Notify(459269)

	if txnFingerprintID, ok := resolvedTxnIDs[txnID]; ok {
		__antithesis_instrumentation__.Notify(459281)
		*existingTxnFingerprintID = txnFingerprintID
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(459282)
	}
	__antithesis_instrumentation__.Notify(459270)

	return true, retryBudgetForMissingResult
}

func (q *resolverQueueImpl) maybeRequeueEventForRetryLocked(
	event contentionpb.ExtendedContentionEvent, initialBudget uint32,
) (requeued bool) {
	__antithesis_instrumentation__.Notify(459283)
	var remainingRetryBudget uint32
	var ok bool

	if initialBudget == retryBudgetForTxnInProgress {
		__antithesis_instrumentation__.Notify(459285)
		delete(q.mu.remainingRetries, event.Hash())
	} else {
		__antithesis_instrumentation__.Notify(459286)

		remainingRetryBudget, ok = q.mu.remainingRetries[event.Hash()]
		if !ok {
			__antithesis_instrumentation__.Notify(459288)
			remainingRetryBudget = initialBudget
		} else {
			__antithesis_instrumentation__.Notify(459289)
			remainingRetryBudget--
		}
		__antithesis_instrumentation__.Notify(459287)

		q.mu.remainingRetries[event.Hash()] = remainingRetryBudget

		if remainingRetryBudget == 0 {
			__antithesis_instrumentation__.Notify(459290)
			delete(q.mu.remainingRetries, event.Hash())
			q.metrics.ResolverFailed.Inc(1)
			q.metrics.ResolverQueueSize.Dec(1)
			return false
		} else {
			__antithesis_instrumentation__.Notify(459291)
		}
	}
	__antithesis_instrumentation__.Notify(459284)

	q.mu.unresolvedEvents = append(q.mu.unresolvedEvents, event)
	q.metrics.ResolverRetries.Inc(1)
	return true
}

func readUntilNextCoordinatorID(
	sortedEvents []contentionpb.ExtendedContentionEvent,
) (eventsForFirstCoordinator, remaining []contentionpb.ExtendedContentionEvent) {
	__antithesis_instrumentation__.Notify(459292)
	if len(sortedEvents) == 0 {
		__antithesis_instrumentation__.Notify(459295)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(459296)
	}
	__antithesis_instrumentation__.Notify(459293)

	currentCoordinatorID := sortedEvents[0].BlockingEvent.TxnMeta.CoordinatorNodeID
	for idx, event := range sortedEvents {
		__antithesis_instrumentation__.Notify(459297)
		if event.BlockingEvent.TxnMeta.CoordinatorNodeID != currentCoordinatorID {
			__antithesis_instrumentation__.Notify(459298)
			return sortedEvents[:idx], sortedEvents[idx:]
		} else {
			__antithesis_instrumentation__.Notify(459299)
		}
	}
	__antithesis_instrumentation__.Notify(459294)

	return sortedEvents, nil
}

func extractResolvedAndInProgressTxnIDs(
	resp *serverpb.TxnIDResolutionResponse,
) (resolvedTxnIDs, inProgressTxnIDs map[uuid.UUID]roachpb.TransactionFingerprintID) {
	__antithesis_instrumentation__.Notify(459300)
	if resp == nil {
		__antithesis_instrumentation__.Notify(459303)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(459304)
	}
	__antithesis_instrumentation__.Notify(459301)

	resolvedTxnIDs = make(map[uuid.UUID]roachpb.TransactionFingerprintID, len(resp.ResolvedTxnIDs))
	inProgressTxnIDs = make(map[uuid.UUID]roachpb.TransactionFingerprintID, len(resp.ResolvedTxnIDs))

	for _, event := range resp.ResolvedTxnIDs {
		__antithesis_instrumentation__.Notify(459305)
		if event.TxnFingerprintID == roachpb.InvalidTransactionFingerprintID {
			__antithesis_instrumentation__.Notify(459306)
			inProgressTxnIDs[event.TxnID] = roachpb.InvalidTransactionFingerprintID
		} else {
			__antithesis_instrumentation__.Notify(459307)
			resolvedTxnIDs[event.TxnID] = event.TxnFingerprintID
		}
	}
	__antithesis_instrumentation__.Notify(459302)

	return resolvedTxnIDs, inProgressTxnIDs
}

func makeRPCRequestsFromBatch(
	batch []contentionpb.ExtendedContentionEvent,
) (blockingTxnIDReq, waitingTxnIDReq *serverpb.TxnIDResolutionRequest) {
	__antithesis_instrumentation__.Notify(459308)
	blockingTxnIDReq = &serverpb.TxnIDResolutionRequest{
		CoordinatorID: strconv.Itoa(int(batch[0].BlockingEvent.TxnMeta.CoordinatorNodeID)),
		TxnIDs:        make([]uuid.UUID, 0, len(batch)),
	}
	waitingTxnIDReq = &serverpb.TxnIDResolutionRequest{
		CoordinatorID: "local",
		TxnIDs:        make([]uuid.UUID, 0, len(batch)),
	}

	for i := range batch {
		__antithesis_instrumentation__.Notify(459310)
		if batch[i].BlockingTxnFingerprintID == roachpb.InvalidTransactionFingerprintID {
			__antithesis_instrumentation__.Notify(459312)
			blockingTxnIDReq.TxnIDs = append(blockingTxnIDReq.TxnIDs, batch[i].BlockingEvent.TxnMeta.ID)
		} else {
			__antithesis_instrumentation__.Notify(459313)
		}
		__antithesis_instrumentation__.Notify(459311)
		if batch[i].WaitingTxnFingerprintID == roachpb.InvalidTransactionFingerprintID {
			__antithesis_instrumentation__.Notify(459314)
			waitingTxnIDReq.TxnIDs = append(waitingTxnIDReq.TxnIDs, batch[i].WaitingTxnID)
		} else {
			__antithesis_instrumentation__.Notify(459315)
		}
	}
	__antithesis_instrumentation__.Notify(459309)

	return blockingTxnIDReq, waitingTxnIDReq
}
