package kvstreamer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"fmt"
	"sort"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type singleRangeBatch struct {
	reqs []roachpb.RequestUnion

	positions []int

	reqsReservedBytes int64

	minTargetBytes int64

	priority int
}

var _ sort.Interface = &singleRangeBatch{}

func (r *singleRangeBatch) Len() int {
	__antithesis_instrumentation__.Notify(498877)
	return len(r.reqs)
}

func (r *singleRangeBatch) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(498878)
	r.reqs[i], r.reqs[j] = r.reqs[j], r.reqs[i]
	r.positions[i], r.positions[j] = r.positions[j], r.positions[i]
}

func (r *singleRangeBatch) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(498879)

	return r.reqs[i].GetInner().Header().Key.Compare(r.reqs[j].GetInner().Header().Key) < 0
}

func reqsToString(reqs []singleRangeBatch) string {
	__antithesis_instrumentation__.Notify(498880)
	result := "requests for positions "
	for i, r := range reqs {
		__antithesis_instrumentation__.Notify(498882)
		if i > 0 {
			__antithesis_instrumentation__.Notify(498884)
			result += ", "
		} else {
			__antithesis_instrumentation__.Notify(498885)
		}
		__antithesis_instrumentation__.Notify(498883)
		result += fmt.Sprintf("%v", r.positions)
	}
	__antithesis_instrumentation__.Notify(498881)
	return result
}

type requestsProvider interface {
	enqueue([]singleRangeBatch)

	close()

	add(singleRangeBatch)

	Lock()
	Unlock()

	waitLocked()

	emptyLocked() bool

	firstLocked() singleRangeBatch

	removeFirstLocked()
}

type requestsProviderBase struct {
	syncutil.Mutex

	hasWork *sync.Cond

	requests []singleRangeBatch

	done bool
}

func (b *requestsProviderBase) init() {
	b.hasWork = sync.NewCond(&b.Mutex)
}

func (b *requestsProviderBase) waitLocked() {
	__antithesis_instrumentation__.Notify(498886)
	b.Mutex.AssertHeld()
	if b.done {
		__antithesis_instrumentation__.Notify(498888)

		return
	} else {
		__antithesis_instrumentation__.Notify(498889)
	}
	__antithesis_instrumentation__.Notify(498887)
	b.hasWork.Wait()
}

func (b *requestsProviderBase) emptyLocked() bool {
	__antithesis_instrumentation__.Notify(498890)
	b.Mutex.AssertHeld()
	return len(b.requests) == 0
}

func (b *requestsProviderBase) close() {
	__antithesis_instrumentation__.Notify(498891)
	b.Lock()
	defer b.Unlock()
	b.done = true
	b.hasWork.Signal()
}

type outOfOrderRequestsProvider struct {
	*requestsProviderBase
}

var _ requestsProvider = &outOfOrderRequestsProvider{}

func newOutOfOrderRequestsProvider() requestsProvider {
	__antithesis_instrumentation__.Notify(498892)
	p := outOfOrderRequestsProvider{requestsProviderBase: &requestsProviderBase{}}
	p.init()
	return &p
}

func (p *outOfOrderRequestsProvider) enqueue(requests []singleRangeBatch) {
	__antithesis_instrumentation__.Notify(498893)
	p.Lock()
	defer p.Unlock()
	if len(p.requests) > 0 {
		__antithesis_instrumentation__.Notify(498895)
		panic(errors.AssertionFailedf("outOfOrderRequestsProvider has old requests in enqueue"))
	} else {
		__antithesis_instrumentation__.Notify(498896)
	}
	__antithesis_instrumentation__.Notify(498894)
	p.requests = requests
	p.hasWork.Signal()
}

func (p *outOfOrderRequestsProvider) add(request singleRangeBatch) {
	__antithesis_instrumentation__.Notify(498897)
	p.Lock()
	defer p.Unlock()
	p.requests = append(p.requests, request)
	p.hasWork.Signal()
}

func (p *outOfOrderRequestsProvider) firstLocked() singleRangeBatch {
	__antithesis_instrumentation__.Notify(498898)
	p.Mutex.AssertHeld()
	if len(p.requests) == 0 {
		__antithesis_instrumentation__.Notify(498900)
		panic(errors.AssertionFailedf("firstLocked called when requestsProvider is empty"))
	} else {
		__antithesis_instrumentation__.Notify(498901)
	}
	__antithesis_instrumentation__.Notify(498899)
	return p.requests[0]
}

func (p *outOfOrderRequestsProvider) removeFirstLocked() {
	__antithesis_instrumentation__.Notify(498902)
	p.Mutex.AssertHeld()
	if len(p.requests) == 0 {
		__antithesis_instrumentation__.Notify(498904)
		panic(errors.AssertionFailedf("removeFirstLocked called when requestsProvider is empty"))
	} else {
		__antithesis_instrumentation__.Notify(498905)
	}
	__antithesis_instrumentation__.Notify(498903)
	p.requests = p.requests[1:]
}

type inOrderRequestsProvider struct {
	*requestsProviderBase
}

var _ requestsProvider = &inOrderRequestsProvider{}
var _ heap.Interface = &inOrderRequestsProvider{}

func newInOrderRequestsProvider() requestsProvider {
	__antithesis_instrumentation__.Notify(498906)
	p := inOrderRequestsProvider{requestsProviderBase: &requestsProviderBase{}}
	p.init()
	return &p
}

func (p *inOrderRequestsProvider) Len() int {
	__antithesis_instrumentation__.Notify(498907)
	return len(p.requests)
}

func (p *inOrderRequestsProvider) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(498908)
	return p.requests[i].priority < p.requests[j].priority
}

func (p *inOrderRequestsProvider) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(498909)
	p.requests[i], p.requests[j] = p.requests[j], p.requests[i]
}

func (p *inOrderRequestsProvider) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(498910)
	p.requests = append(p.requests, x.(singleRangeBatch))
}

func (p *inOrderRequestsProvider) Pop() interface{} {
	__antithesis_instrumentation__.Notify(498911)
	x := p.requests[len(p.requests)-1]
	p.requests = p.requests[:len(p.requests)-1]
	return x
}

func (p *inOrderRequestsProvider) enqueue(requests []singleRangeBatch) {
	__antithesis_instrumentation__.Notify(498912)
	p.Lock()
	defer p.Unlock()
	if len(p.requests) > 0 {
		__antithesis_instrumentation__.Notify(498914)
		panic(errors.AssertionFailedf("inOrderRequestsProvider has old requests in enqueue"))
	} else {
		__antithesis_instrumentation__.Notify(498915)
	}
	__antithesis_instrumentation__.Notify(498913)
	p.requests = requests
	heap.Init(p)
	p.hasWork.Signal()
}

func (p *inOrderRequestsProvider) add(request singleRangeBatch) {
	__antithesis_instrumentation__.Notify(498916)
	p.Lock()
	defer p.Unlock()
	if debug {
		__antithesis_instrumentation__.Notify(498918)
		fmt.Printf("adding a request for positions %v to be served, minTargetBytes=%d\n", request.positions, request.minTargetBytes)
	} else {
		__antithesis_instrumentation__.Notify(498919)
	}
	__antithesis_instrumentation__.Notify(498917)
	heap.Push(p, request)
	p.hasWork.Signal()
}

func (p *inOrderRequestsProvider) firstLocked() singleRangeBatch {
	__antithesis_instrumentation__.Notify(498920)
	p.Mutex.AssertHeld()
	if len(p.requests) == 0 {
		__antithesis_instrumentation__.Notify(498922)
		panic(errors.AssertionFailedf("firstLocked called when requestsProvider is empty"))
	} else {
		__antithesis_instrumentation__.Notify(498923)
	}
	__antithesis_instrumentation__.Notify(498921)
	return p.requests[0]
}

func (p *inOrderRequestsProvider) removeFirstLocked() {
	__antithesis_instrumentation__.Notify(498924)
	p.Mutex.AssertHeld()
	if len(p.requests) == 0 {
		__antithesis_instrumentation__.Notify(498926)
		panic(errors.AssertionFailedf("removeFirstLocked called when requestsProvider is empty"))
	} else {
		__antithesis_instrumentation__.Notify(498927)
	}
	__antithesis_instrumentation__.Notify(498925)
	heap.Remove(p, 0)
}
