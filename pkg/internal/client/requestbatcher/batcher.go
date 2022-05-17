// Package requestbatcher is a library to enable easy batching of roachpb
// requests.
//
// Batching in general represents a tradeoff between throughput and latency. The
// underlying assumption being that batched operations are cheaper than an
// individual operation. If this is not the case for your workload, don't use
// this library.
//
// Batching assumes that data with the same key can be sent in a single batch.
// The initial implementation uses rangeID as the key explicitly to avoid
// creating an overly general solution without motivation but interested readers
// should recognize that it would be easy to extend this package to accept an
// arbitrary comparable key.
package requestbatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type Config struct {
	AmbientCtx log.AmbientContext

	Name string

	Sender kv.Sender

	Stopper *stop.Stopper

	MaxSizePerBatch int

	MaxMsgsPerBatch int

	MaxKeysPerBatchReq int

	MaxWait time.Duration

	MaxIdle time.Duration

	InFlightBackpressureLimit int

	NowFunc func() time.Time
}

const (
	DefaultInFlightBackpressureLimit = 1000

	backpressureRecoveryFraction = .8
)

func backpressureRecoveryThreshold(limit int) int {
	__antithesis_instrumentation__.Notify(68239)
	if l := int(float64(limit) * backpressureRecoveryFraction); l > 0 {
		__antithesis_instrumentation__.Notify(68241)
		return l
	} else {
		__antithesis_instrumentation__.Notify(68242)
	}
	__antithesis_instrumentation__.Notify(68240)
	return 1
}

type RequestBatcher struct {
	pool pool
	cfg  Config

	sendBatchOpName string

	batches batchQueue

	requestChan  chan *request
	sendDoneChan chan struct{}
}

type Response struct {
	Resp roachpb.Response
	Err  error
}

func New(cfg Config) *RequestBatcher {
	__antithesis_instrumentation__.Notify(68243)
	validateConfig(&cfg)
	b := &RequestBatcher{
		cfg:          cfg,
		pool:         makePool(),
		batches:      makeBatchQueue(),
		requestChan:  make(chan *request),
		sendDoneChan: make(chan struct{}),
	}
	b.sendBatchOpName = b.cfg.Name + ".sendBatch"
	bgCtx := cfg.AmbientCtx.AnnotateCtx(context.Background())
	if err := cfg.Stopper.RunAsyncTask(bgCtx, b.cfg.Name, b.run); err != nil {
		__antithesis_instrumentation__.Notify(68245)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(68246)
	}
	__antithesis_instrumentation__.Notify(68244)
	return b
}

func validateConfig(cfg *Config) {
	__antithesis_instrumentation__.Notify(68247)
	if cfg.Stopper == nil {
		__antithesis_instrumentation__.Notify(68250)
		panic("cannot construct a Batcher with a nil Stopper")
	} else {
		__antithesis_instrumentation__.Notify(68251)
		if cfg.Sender == nil {
			__antithesis_instrumentation__.Notify(68252)
			panic("cannot construct a Batcher with a nil Sender")
		} else {
			__antithesis_instrumentation__.Notify(68253)
		}
	}
	__antithesis_instrumentation__.Notify(68248)
	if cfg.InFlightBackpressureLimit <= 0 {
		__antithesis_instrumentation__.Notify(68254)
		cfg.InFlightBackpressureLimit = DefaultInFlightBackpressureLimit
	} else {
		__antithesis_instrumentation__.Notify(68255)
	}
	__antithesis_instrumentation__.Notify(68249)
	if cfg.NowFunc == nil {
		__antithesis_instrumentation__.Notify(68256)
		cfg.NowFunc = timeutil.Now
	} else {
		__antithesis_instrumentation__.Notify(68257)
	}
}

func (b *RequestBatcher) SendWithChan(
	ctx context.Context, respChan chan<- Response, rangeID roachpb.RangeID, req roachpb.Request,
) error {
	__antithesis_instrumentation__.Notify(68258)
	select {
	case b.requestChan <- b.pool.newRequest(ctx, rangeID, req, respChan):
		__antithesis_instrumentation__.Notify(68259)
		return nil
	case <-b.cfg.Stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(68260)
		return stop.ErrUnavailable
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(68261)
		return ctx.Err()
	}
}

func (b *RequestBatcher) Send(
	ctx context.Context, rangeID roachpb.RangeID, req roachpb.Request,
) (roachpb.Response, error) {
	__antithesis_instrumentation__.Notify(68262)
	responseChan := b.pool.getResponseChan()
	if err := b.SendWithChan(ctx, responseChan, rangeID, req); err != nil {
		__antithesis_instrumentation__.Notify(68264)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(68265)
	}
	__antithesis_instrumentation__.Notify(68263)
	select {
	case resp := <-responseChan:
		__antithesis_instrumentation__.Notify(68266)

		b.pool.putResponseChan(responseChan)
		return resp.Resp, resp.Err
	case <-b.cfg.Stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(68267)
		return nil, stop.ErrUnavailable
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(68268)
		return nil, ctx.Err()
	}
}

func (b *RequestBatcher) sendDone(ba *batch) {
	__antithesis_instrumentation__.Notify(68269)
	b.pool.putBatch(ba)
	select {
	case b.sendDoneChan <- struct{}{}:
		__antithesis_instrumentation__.Notify(68270)
	case <-b.cfg.Stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(68271)
	}
}

func (b *RequestBatcher) sendBatch(ctx context.Context, ba *batch) {
	__antithesis_instrumentation__.Notify(68272)
	if err := b.cfg.Stopper.RunAsyncTask(ctx, "send-batch", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(68273)
		defer b.sendDone(ba)
		var br *roachpb.BatchResponse
		send := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(68276)
			var pErr *roachpb.Error
			if br, pErr = b.cfg.Sender.Send(ctx, ba.batchRequest(&b.cfg)); pErr != nil {
				__antithesis_instrumentation__.Notify(68278)
				return pErr.GoError()
			} else {
				__antithesis_instrumentation__.Notify(68279)
			}
			__antithesis_instrumentation__.Notify(68277)
			return nil
		}
		__antithesis_instrumentation__.Notify(68274)
		if !ba.sendDeadline.IsZero() {
			__antithesis_instrumentation__.Notify(68280)
			actualSend := send
			send = func(context.Context) error {
				__antithesis_instrumentation__.Notify(68281)
				return contextutil.RunWithTimeout(
					ctx, b.sendBatchOpName, timeutil.Until(ba.sendDeadline), actualSend)
			}
		} else {
			__antithesis_instrumentation__.Notify(68282)
		}
		__antithesis_instrumentation__.Notify(68275)

		var prevResps []roachpb.Response
		for len(ba.reqs) > 0 {
			__antithesis_instrumentation__.Notify(68283)
			err := send(ctx)
			nextReqs, nextPrevResps := ba.reqs[:0], prevResps[:0]
			for i, r := range ba.reqs {
				__antithesis_instrumentation__.Notify(68285)
				var res Response
				if br != nil {
					__antithesis_instrumentation__.Notify(68288)
					resp := br.Responses[i].GetInner()
					if prevResps != nil {
						__antithesis_instrumentation__.Notify(68291)
						prevResp := prevResps[i]
						if cErr := roachpb.CombineResponses(prevResp, resp); cErr != nil {
							__antithesis_instrumentation__.Notify(68293)
							log.Fatalf(ctx, "%v", cErr)
						} else {
							__antithesis_instrumentation__.Notify(68294)
						}
						__antithesis_instrumentation__.Notify(68292)
						resp = prevResp
					} else {
						__antithesis_instrumentation__.Notify(68295)
					}
					__antithesis_instrumentation__.Notify(68289)
					if resume := resp.Header().ResumeSpan; resume != nil {
						__antithesis_instrumentation__.Notify(68296)

						h := r.req.Header()
						h.SetSpan(*resume)
						r.req = r.req.ShallowCopy()
						r.req.SetHeader(h)
						nextReqs = append(nextReqs, r)

						prevH := resp.Header()
						prevH.ResumeSpan = nil
						prevResp := resp
						prevResp.SetHeader(prevH)
						nextPrevResps = append(nextPrevResps, prevResp)
						continue
					} else {
						__antithesis_instrumentation__.Notify(68297)
					}
					__antithesis_instrumentation__.Notify(68290)
					res.Resp = resp
				} else {
					__antithesis_instrumentation__.Notify(68298)
				}
				__antithesis_instrumentation__.Notify(68286)
				if err != nil {
					__antithesis_instrumentation__.Notify(68299)
					res.Err = err
				} else {
					__antithesis_instrumentation__.Notify(68300)
				}
				__antithesis_instrumentation__.Notify(68287)
				b.sendResponse(r, res)
			}
			__antithesis_instrumentation__.Notify(68284)
			ba.reqs, prevResps = nextReqs, nextPrevResps
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(68301)
		b.sendDone(ba)
	} else {
		__antithesis_instrumentation__.Notify(68302)
	}
}

func (b *RequestBatcher) sendResponse(req *request, resp Response) {
	__antithesis_instrumentation__.Notify(68303)

	req.responseChan <- resp
	b.pool.putRequest(req)
}

func addRequestToBatch(cfg *Config, now time.Time, ba *batch, r *request) (shouldSend bool) {
	__antithesis_instrumentation__.Notify(68304)

	rDeadline, rHasDeadline := r.ctx.Deadline()

	if len(ba.reqs) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(68308)
		return (len(ba.reqs) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(68309)
			return !ba.sendDeadline.IsZero() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(68310)
			return (!rHasDeadline || func() bool {
				__antithesis_instrumentation__.Notify(68311)
				return rDeadline.After(ba.sendDeadline) == true
			}() == true) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(68312)

		ba.sendDeadline = rDeadline
	} else {
		__antithesis_instrumentation__.Notify(68313)
	}
	__antithesis_instrumentation__.Notify(68305)

	ba.reqs = append(ba.reqs, r)
	ba.size += r.req.Size()
	ba.lastUpdated = now

	if cfg.MaxIdle > 0 {
		__antithesis_instrumentation__.Notify(68314)
		ba.deadline = ba.lastUpdated.Add(cfg.MaxIdle)
	} else {
		__antithesis_instrumentation__.Notify(68315)
	}
	__antithesis_instrumentation__.Notify(68306)
	if cfg.MaxWait > 0 {
		__antithesis_instrumentation__.Notify(68316)
		waitDeadline := ba.startTime.Add(cfg.MaxWait)
		if cfg.MaxIdle <= 0 || func() bool {
			__antithesis_instrumentation__.Notify(68317)
			return waitDeadline.Before(ba.deadline) == true
		}() == true {
			__antithesis_instrumentation__.Notify(68318)
			ba.deadline = waitDeadline
		} else {
			__antithesis_instrumentation__.Notify(68319)
		}
	} else {
		__antithesis_instrumentation__.Notify(68320)
	}
	__antithesis_instrumentation__.Notify(68307)
	return (cfg.MaxMsgsPerBatch > 0 && func() bool {
		__antithesis_instrumentation__.Notify(68321)
		return len(ba.reqs) >= cfg.MaxMsgsPerBatch == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(68322)
		return (cfg.MaxSizePerBatch > 0 && func() bool {
			__antithesis_instrumentation__.Notify(68323)
			return ba.size >= cfg.MaxSizePerBatch == true
		}() == true) == true
	}() == true
}

func (b *RequestBatcher) cleanup(err error) {
	__antithesis_instrumentation__.Notify(68324)
	for ba := b.batches.popFront(); ba != nil; ba = b.batches.popFront() {
		__antithesis_instrumentation__.Notify(68325)
		for _, r := range ba.reqs {
			__antithesis_instrumentation__.Notify(68326)
			b.sendResponse(r, Response{Err: err})
		}
	}
}

func (b *RequestBatcher) run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(68327)

	sendCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		inFlight = 0

		inBackPressure = false

		recoveryThreshold = backpressureRecoveryThreshold(b.cfg.InFlightBackpressureLimit)

		reqChan = func() <-chan *request {
			__antithesis_instrumentation__.Notify(68329)
			if inBackPressure {
				__antithesis_instrumentation__.Notify(68331)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(68332)
			}
			__antithesis_instrumentation__.Notify(68330)
			return b.requestChan
		}
		sendBatch = func(ba *batch) {
			__antithesis_instrumentation__.Notify(68333)
			inFlight++
			if inFlight >= b.cfg.InFlightBackpressureLimit {
				__antithesis_instrumentation__.Notify(68335)
				inBackPressure = true
			} else {
				__antithesis_instrumentation__.Notify(68336)
			}
			__antithesis_instrumentation__.Notify(68334)
			b.sendBatch(sendCtx, ba)
		}
		handleSendDone = func() {
			__antithesis_instrumentation__.Notify(68337)
			inFlight--
			if inFlight < recoveryThreshold {
				__antithesis_instrumentation__.Notify(68338)
				inBackPressure = false
			} else {
				__antithesis_instrumentation__.Notify(68339)
			}
		}
		handleRequest = func(req *request) {
			__antithesis_instrumentation__.Notify(68340)
			now := b.cfg.NowFunc()
			ba, existsInQueue := b.batches.get(req.rangeID)
			if !existsInQueue {
				__antithesis_instrumentation__.Notify(68342)
				ba = b.pool.newBatch(now)
			} else {
				__antithesis_instrumentation__.Notify(68343)
			}
			__antithesis_instrumentation__.Notify(68341)
			if shouldSend := addRequestToBatch(&b.cfg, now, ba, req); shouldSend {
				__antithesis_instrumentation__.Notify(68344)
				if existsInQueue {
					__antithesis_instrumentation__.Notify(68346)
					b.batches.remove(ba)
				} else {
					__antithesis_instrumentation__.Notify(68347)
				}
				__antithesis_instrumentation__.Notify(68345)
				sendBatch(ba)
			} else {
				__antithesis_instrumentation__.Notify(68348)
				b.batches.upsert(ba)
			}
		}
		deadline      time.Time
		timer         = timeutil.NewTimer()
		maybeSetTimer = func() {
			__antithesis_instrumentation__.Notify(68349)
			var nextDeadline time.Time
			if next := b.batches.peekFront(); next != nil {
				__antithesis_instrumentation__.Notify(68351)
				nextDeadline = next.deadline
			} else {
				__antithesis_instrumentation__.Notify(68352)
			}
			__antithesis_instrumentation__.Notify(68350)
			if !deadline.Equal(nextDeadline) || func() bool {
				__antithesis_instrumentation__.Notify(68353)
				return timer.Read == true
			}() == true {
				__antithesis_instrumentation__.Notify(68354)
				deadline = nextDeadline
				if !deadline.IsZero() {
					__antithesis_instrumentation__.Notify(68355)
					timer.Reset(timeutil.Until(deadline))
				} else {
					__antithesis_instrumentation__.Notify(68356)

					timer.Stop()
					timer = timeutil.NewTimer()
				}
			} else {
				__antithesis_instrumentation__.Notify(68357)
			}
		}
	)
	__antithesis_instrumentation__.Notify(68328)
	for {
		__antithesis_instrumentation__.Notify(68358)
		select {
		case req := <-reqChan():
			__antithesis_instrumentation__.Notify(68359)
			handleRequest(req)
			maybeSetTimer()
		case <-timer.C:
			__antithesis_instrumentation__.Notify(68360)
			timer.Read = true
			sendBatch(b.batches.popFront())
			maybeSetTimer()
		case <-b.sendDoneChan:
			__antithesis_instrumentation__.Notify(68361)
			handleSendDone()
		case <-b.cfg.Stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(68362)
			b.cleanup(stop.ErrUnavailable)
			return
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(68363)
			b.cleanup(ctx.Err())
			return
		}
	}
}

type request struct {
	ctx          context.Context
	req          roachpb.Request
	rangeID      roachpb.RangeID
	responseChan chan<- Response
}

type batch struct {
	reqs []*request
	size int

	sendDeadline time.Time

	idx int

	deadline time.Time

	startTime time.Time

	lastUpdated time.Time
}

func (b *batch) rangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(68364)
	if len(b.reqs) == 0 {
		__antithesis_instrumentation__.Notify(68366)
		panic("rangeID cannot be called on an empty batch")
	} else {
		__antithesis_instrumentation__.Notify(68367)
	}
	__antithesis_instrumentation__.Notify(68365)
	return b.reqs[0].rangeID
}

func (b *batch) batchRequest(cfg *Config) roachpb.BatchRequest {
	__antithesis_instrumentation__.Notify(68368)
	req := roachpb.BatchRequest{

		Requests: make([]roachpb.RequestUnion, 0, len(b.reqs)),
	}
	for _, r := range b.reqs {
		__antithesis_instrumentation__.Notify(68371)
		req.Add(r.req)
	}
	__antithesis_instrumentation__.Notify(68369)
	if cfg.MaxKeysPerBatchReq > 0 {
		__antithesis_instrumentation__.Notify(68372)
		req.MaxSpanRequestKeys = int64(cfg.MaxKeysPerBatchReq)
	} else {
		__antithesis_instrumentation__.Notify(68373)
	}
	__antithesis_instrumentation__.Notify(68370)
	return req
}

type pool struct {
	responseChanPool sync.Pool
	batchPool        sync.Pool
	requestPool      sync.Pool
}

func makePool() pool {
	__antithesis_instrumentation__.Notify(68374)
	return pool{
		responseChanPool: sync.Pool{
			New: func() interface{} { __antithesis_instrumentation__.Notify(68375); return make(chan Response, 1) },
		},
		batchPool: sync.Pool{
			New: func() interface{} { __antithesis_instrumentation__.Notify(68376); return &batch{} },
		},
		requestPool: sync.Pool{
			New: func() interface{} { __antithesis_instrumentation__.Notify(68377); return &request{} },
		},
	}
}

func (p *pool) getResponseChan() chan Response {
	__antithesis_instrumentation__.Notify(68378)
	return p.responseChanPool.Get().(chan Response)
}

func (p *pool) putResponseChan(r chan Response) {
	__antithesis_instrumentation__.Notify(68379)
	p.responseChanPool.Put(r)
}

func (p *pool) newRequest(
	ctx context.Context, rangeID roachpb.RangeID, req roachpb.Request, responseChan chan<- Response,
) *request {
	__antithesis_instrumentation__.Notify(68380)
	r := p.requestPool.Get().(*request)
	*r = request{
		ctx:          ctx,
		rangeID:      rangeID,
		req:          req,
		responseChan: responseChan,
	}
	return r
}

func (p *pool) putRequest(r *request) {
	__antithesis_instrumentation__.Notify(68381)
	*r = request{}
	p.requestPool.Put(r)
}

func (p *pool) newBatch(now time.Time) *batch {
	__antithesis_instrumentation__.Notify(68382)
	ba := p.batchPool.Get().(*batch)
	*ba = batch{
		startTime: now,
		idx:       -1,
	}
	return ba
}

func (p *pool) putBatch(b *batch) {
	__antithesis_instrumentation__.Notify(68383)
	*b = batch{}
	p.batchPool.Put(b)
}

type batchQueue struct {
	batches []*batch
	byRange map[roachpb.RangeID]*batch
}

var _ heap.Interface = (*batchQueue)(nil)

func makeBatchQueue() batchQueue {
	__antithesis_instrumentation__.Notify(68384)
	return batchQueue{
		byRange: map[roachpb.RangeID]*batch{},
	}
}

func (q *batchQueue) peekFront() *batch {
	__antithesis_instrumentation__.Notify(68385)
	if q.Len() == 0 {
		__antithesis_instrumentation__.Notify(68387)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(68388)
	}
	__antithesis_instrumentation__.Notify(68386)
	return q.batches[0]
}

func (q *batchQueue) popFront() *batch {
	__antithesis_instrumentation__.Notify(68389)
	if q.Len() == 0 {
		__antithesis_instrumentation__.Notify(68391)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(68392)
	}
	__antithesis_instrumentation__.Notify(68390)
	return heap.Pop(q).(*batch)
}

func (q *batchQueue) get(id roachpb.RangeID) (*batch, bool) {
	__antithesis_instrumentation__.Notify(68393)
	b, exists := q.byRange[id]
	return b, exists
}

func (q *batchQueue) remove(ba *batch) {
	__antithesis_instrumentation__.Notify(68394)
	delete(q.byRange, ba.rangeID())
	heap.Remove(q, ba.idx)
}

func (q *batchQueue) upsert(ba *batch) {
	__antithesis_instrumentation__.Notify(68395)
	if ba.idx >= 0 {
		__antithesis_instrumentation__.Notify(68396)
		heap.Fix(q, ba.idx)
	} else {
		__antithesis_instrumentation__.Notify(68397)
		heap.Push(q, ba)
	}
}

func (q *batchQueue) Len() int {
	__antithesis_instrumentation__.Notify(68398)
	return len(q.batches)
}

func (q *batchQueue) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(68399)
	q.batches[i], q.batches[j] = q.batches[j], q.batches[i]
	q.batches[i].idx = i
	q.batches[j].idx = j
}

func (q *batchQueue) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(68400)
	idl, jdl := q.batches[i].deadline, q.batches[j].deadline
	if before := idl.Before(jdl); before || func() bool {
		__antithesis_instrumentation__.Notify(68402)
		return !idl.Equal(jdl) == true
	}() == true {
		__antithesis_instrumentation__.Notify(68403)
		return before
	} else {
		__antithesis_instrumentation__.Notify(68404)
	}
	__antithesis_instrumentation__.Notify(68401)
	return q.batches[i].rangeID() < q.batches[j].rangeID()
}

func (q *batchQueue) Push(v interface{}) {
	__antithesis_instrumentation__.Notify(68405)
	ba := v.(*batch)
	ba.idx = len(q.batches)
	q.byRange[ba.rangeID()] = ba
	q.batches = append(q.batches, ba)
}

func (q *batchQueue) Pop() interface{} {
	__antithesis_instrumentation__.Notify(68406)
	ba := q.batches[len(q.batches)-1]
	q.batches = q.batches[:len(q.batches)-1]
	delete(q.byRange, ba.rangeID())
	ba.idx = -1
	return ba
}
