package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const rangeIDChunkSize = 1000

type rangeIDChunk struct {
	buf    [rangeIDChunkSize]roachpb.RangeID
	rd, wr int
}

func (c *rangeIDChunk) PushBack(id roachpb.RangeID) bool {
	__antithesis_instrumentation__.Notify(122237)
	if c.WriteCap() == 0 {
		__antithesis_instrumentation__.Notify(122239)
		return false
	} else {
		__antithesis_instrumentation__.Notify(122240)
	}
	__antithesis_instrumentation__.Notify(122238)
	c.buf[c.wr] = id
	c.wr++
	return true
}

func (c *rangeIDChunk) PopFront() (roachpb.RangeID, bool) {
	__antithesis_instrumentation__.Notify(122241)
	if c.Len() == 0 {
		__antithesis_instrumentation__.Notify(122243)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(122244)
	}
	__antithesis_instrumentation__.Notify(122242)
	id := c.buf[c.rd]
	c.rd++
	return id, true
}

func (c *rangeIDChunk) WriteCap() int {
	__antithesis_instrumentation__.Notify(122245)
	return len(c.buf) - c.wr
}

func (c *rangeIDChunk) Len() int {
	__antithesis_instrumentation__.Notify(122246)
	return c.wr - c.rd
}

type rangeIDQueue struct {
	len int

	chunks list.List

	priorityID     roachpb.RangeID
	priorityQueued bool
}

func (q *rangeIDQueue) Push(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122247)
	q.len++
	if q.priorityID == id {
		__antithesis_instrumentation__.Notify(122250)
		q.priorityQueued = true
		return
	} else {
		__antithesis_instrumentation__.Notify(122251)
	}
	__antithesis_instrumentation__.Notify(122248)
	if q.chunks.Len() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(122252)
		return q.back().WriteCap() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(122253)
		q.chunks.PushBack(&rangeIDChunk{})
	} else {
		__antithesis_instrumentation__.Notify(122254)
	}
	__antithesis_instrumentation__.Notify(122249)
	if !q.back().PushBack(id) {
		__antithesis_instrumentation__.Notify(122255)
		panic(fmt.Sprintf(
			"unable to push rangeID to chunk: len=%d, cap=%d",
			q.back().Len(), q.back().WriteCap()))
	} else {
		__antithesis_instrumentation__.Notify(122256)
	}
}

func (q *rangeIDQueue) PopFront() (roachpb.RangeID, bool) {
	__antithesis_instrumentation__.Notify(122257)
	if q.len == 0 {
		__antithesis_instrumentation__.Notify(122262)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(122263)
	}
	__antithesis_instrumentation__.Notify(122258)
	q.len--
	if q.priorityQueued {
		__antithesis_instrumentation__.Notify(122264)
		q.priorityQueued = false
		return q.priorityID, true
	} else {
		__antithesis_instrumentation__.Notify(122265)
	}
	__antithesis_instrumentation__.Notify(122259)
	frontElem := q.chunks.Front()
	front := frontElem.Value.(*rangeIDChunk)
	id, ok := front.PopFront()
	if !ok {
		__antithesis_instrumentation__.Notify(122266)
		panic("encountered empty chunk")
	} else {
		__antithesis_instrumentation__.Notify(122267)
	}
	__antithesis_instrumentation__.Notify(122260)
	if front.Len() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(122268)
		return front.WriteCap() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(122269)
		q.chunks.Remove(frontElem)
	} else {
		__antithesis_instrumentation__.Notify(122270)
	}
	__antithesis_instrumentation__.Notify(122261)
	return id, true
}

func (q *rangeIDQueue) Len() int {
	__antithesis_instrumentation__.Notify(122271)
	return q.len
}

func (q *rangeIDQueue) SetPriorityID(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122272)
	if q.priorityID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(122274)
		return q.priorityID != id == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(122275)
		return false == true
	}() == true {
		__antithesis_instrumentation__.Notify(122276)
		panic(fmt.Sprintf(
			"priority range ID already set: old=%d, new=%d",
			q.priorityID, id))
	} else {
		__antithesis_instrumentation__.Notify(122277)
	}
	__antithesis_instrumentation__.Notify(122273)
	q.priorityID = id
}

func (q *rangeIDQueue) back() *rangeIDChunk {
	__antithesis_instrumentation__.Notify(122278)
	return q.chunks.Back().Value.(*rangeIDChunk)
}

type raftProcessor interface {
	processReady(roachpb.RangeID)

	processRequestQueue(context.Context, roachpb.RangeID) bool

	processTick(context.Context, roachpb.RangeID) bool
}

type raftScheduleFlags int

const (
	stateQueued raftScheduleFlags = 1 << iota
	stateRaftReady
	stateRaftRequest
	stateRaftTick
)

type raftScheduleState struct {
	flags raftScheduleFlags
	begin int64
}

type raftScheduler struct {
	ambientContext log.AmbientContext
	processor      raftProcessor
	latency        *metric.Histogram
	numWorkers     int

	mu struct {
		syncutil.Mutex
		cond    *sync.Cond
		queue   rangeIDQueue
		state   map[roachpb.RangeID]raftScheduleState
		stopped bool
	}

	done sync.WaitGroup
}

func newRaftScheduler(
	ambient log.AmbientContext, metrics *StoreMetrics, processor raftProcessor, numWorkers int,
) *raftScheduler {
	__antithesis_instrumentation__.Notify(122279)
	s := &raftScheduler{
		ambientContext: ambient,
		processor:      processor,
		latency:        metrics.RaftSchedulerLatency,
		numWorkers:     numWorkers,
	}
	s.mu.cond = sync.NewCond(&s.mu.Mutex)
	s.mu.state = make(map[roachpb.RangeID]raftScheduleState)
	return s
}

func (s *raftScheduler) Start(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(122280)
	ctx := s.ambientContext.AnnotateCtx(context.Background())
	waitQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(122283)
		<-stopper.ShouldQuiesce()
		s.mu.Lock()
		s.mu.stopped = true
		s.mu.Unlock()
		s.mu.cond.Broadcast()
	}
	__antithesis_instrumentation__.Notify(122281)
	if err := stopper.RunAsyncTaskEx(ctx,
		stop.TaskOpts{
			TaskName: "raftsched-wait-quiesce",

			SpanOpt: stop.SterileRootSpan,
		},
		waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(122284)
		waitQuiesce(ctx)
	} else {
		__antithesis_instrumentation__.Notify(122285)
	}
	__antithesis_instrumentation__.Notify(122282)

	s.done.Add(s.numWorkers)
	for i := 0; i < s.numWorkers; i++ {
		__antithesis_instrumentation__.Notify(122286)
		if err := stopper.RunAsyncTaskEx(ctx,
			stop.TaskOpts{
				TaskName: "raft-worker",

				SpanOpt: stop.SterileRootSpan,
			},
			s.worker); err != nil {
			__antithesis_instrumentation__.Notify(122287)
			s.done.Done()
		} else {
			__antithesis_instrumentation__.Notify(122288)
		}
	}
}

func (s *raftScheduler) Wait(context.Context) {
	__antithesis_instrumentation__.Notify(122289)
	s.done.Wait()
}

func (s *raftScheduler) SetPriorityID(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122290)
	s.mu.Lock()
	s.mu.queue.SetPriorityID(id)
	s.mu.Unlock()
}

func (s *raftScheduler) PriorityID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(122291)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.queue.priorityID
}

func (s *raftScheduler) worker(ctx context.Context) {
	__antithesis_instrumentation__.Notify(122292)
	defer s.done.Done()

	s.mu.Lock()
	for {
		__antithesis_instrumentation__.Notify(122293)
		var id roachpb.RangeID
		for {
			__antithesis_instrumentation__.Notify(122298)
			if s.mu.stopped {
				__antithesis_instrumentation__.Notify(122301)
				s.mu.Unlock()
				return
			} else {
				__antithesis_instrumentation__.Notify(122302)
			}
			__antithesis_instrumentation__.Notify(122299)
			var ok bool
			if id, ok = s.mu.queue.PopFront(); ok {
				__antithesis_instrumentation__.Notify(122303)
				break
			} else {
				__antithesis_instrumentation__.Notify(122304)
			}
			__antithesis_instrumentation__.Notify(122300)
			s.mu.cond.Wait()
		}
		__antithesis_instrumentation__.Notify(122294)

		state := s.mu.state[id]
		s.mu.state[id] = raftScheduleState{flags: stateQueued}
		s.mu.Unlock()

		lat := nowNanos() - state.begin
		s.latency.RecordValue(lat)

		if state.flags&stateRaftRequest != 0 {
			__antithesis_instrumentation__.Notify(122305)

			if s.processor.processRequestQueue(ctx, id) {
				__antithesis_instrumentation__.Notify(122306)
				state.flags |= stateRaftReady
			} else {
				__antithesis_instrumentation__.Notify(122307)
			}
		} else {
			__antithesis_instrumentation__.Notify(122308)
		}
		__antithesis_instrumentation__.Notify(122295)
		if state.flags&stateRaftTick != 0 {
			__antithesis_instrumentation__.Notify(122309)

			if s.processor.processTick(ctx, id) {
				__antithesis_instrumentation__.Notify(122310)
				state.flags |= stateRaftReady
			} else {
				__antithesis_instrumentation__.Notify(122311)
			}
		} else {
			__antithesis_instrumentation__.Notify(122312)
		}
		__antithesis_instrumentation__.Notify(122296)
		if state.flags&stateRaftReady != 0 {
			__antithesis_instrumentation__.Notify(122313)
			s.processor.processReady(id)
		} else {
			__antithesis_instrumentation__.Notify(122314)
		}
		__antithesis_instrumentation__.Notify(122297)

		s.mu.Lock()
		state = s.mu.state[id]
		if state.flags == stateQueued {
			__antithesis_instrumentation__.Notify(122315)

			delete(s.mu.state, id)
		} else {
			__antithesis_instrumentation__.Notify(122316)

			s.mu.queue.Push(id)
		}
	}
}

func (s *raftScheduler) enqueue1Locked(
	addFlags raftScheduleFlags, id roachpb.RangeID, now int64,
) int {
	__antithesis_instrumentation__.Notify(122317)
	prevState := s.mu.state[id]
	if prevState.flags&addFlags == addFlags {
		__antithesis_instrumentation__.Notify(122321)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(122322)
	}
	__antithesis_instrumentation__.Notify(122318)
	var queued int
	newState := prevState
	newState.flags = newState.flags | addFlags
	if newState.flags&stateQueued == 0 {
		__antithesis_instrumentation__.Notify(122323)
		newState.flags |= stateQueued
		queued++
		s.mu.queue.Push(id)
	} else {
		__antithesis_instrumentation__.Notify(122324)
	}
	__antithesis_instrumentation__.Notify(122319)
	if newState.begin == 0 {
		__antithesis_instrumentation__.Notify(122325)
		newState.begin = now
	} else {
		__antithesis_instrumentation__.Notify(122326)
	}
	__antithesis_instrumentation__.Notify(122320)
	s.mu.state[id] = newState
	return queued
}

func (s *raftScheduler) enqueue1(addFlags raftScheduleFlags, id roachpb.RangeID) int {
	__antithesis_instrumentation__.Notify(122327)
	now := nowNanos()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enqueue1Locked(addFlags, id, now)
}

func (s *raftScheduler) enqueueN(addFlags raftScheduleFlags, ids ...roachpb.RangeID) int {
	__antithesis_instrumentation__.Notify(122328)

	const enqueueChunkSize = 128

	if len(ids) == 0 {
		__antithesis_instrumentation__.Notify(122331)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(122332)
	}
	__antithesis_instrumentation__.Notify(122329)

	now := nowNanos()
	s.mu.Lock()
	var count int
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(122333)
		count += s.enqueue1Locked(addFlags, id, now)
		if (i+1)%enqueueChunkSize == 0 {
			__antithesis_instrumentation__.Notify(122334)
			s.mu.Unlock()
			now = nowNanos()
			s.mu.Lock()
		} else {
			__antithesis_instrumentation__.Notify(122335)
		}
	}
	__antithesis_instrumentation__.Notify(122330)
	s.mu.Unlock()
	return count
}

func (s *raftScheduler) signal(count int) {
	__antithesis_instrumentation__.Notify(122336)
	if count >= s.numWorkers {
		__antithesis_instrumentation__.Notify(122337)
		s.mu.cond.Broadcast()
	} else {
		__antithesis_instrumentation__.Notify(122338)
		for i := 0; i < count; i++ {
			__antithesis_instrumentation__.Notify(122339)
			s.mu.cond.Signal()
		}
	}
}

func (s *raftScheduler) EnqueueRaftReady(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122340)
	s.signal(s.enqueue1(stateRaftReady, id))
}

func (s *raftScheduler) EnqueueRaftRequest(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122341)
	s.signal(s.enqueue1(stateRaftRequest, id))
}

func (s *raftScheduler) EnqueueRaftRequests(ids ...roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122342)
	s.signal(s.enqueueN(stateRaftRequest, ids...))
}

func (s *raftScheduler) EnqueueRaftTicks(ids ...roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(122343)
	s.signal(s.enqueueN(stateRaftTick, ids...))
}

func nowNanos() int64 {
	__antithesis_instrumentation__.Notify(122344)
	return timeutil.Now().UnixNano()
}
