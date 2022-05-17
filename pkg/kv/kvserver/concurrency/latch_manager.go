package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/poison"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanlatch"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type latchManagerImpl struct {
	m spanlatch.Manager
}

func (m *latchManagerImpl) Acquire(ctx context.Context, req Request) (latchGuard, *Error) {
	__antithesis_instrumentation__.Notify(99148)
	lg, err := m.m.Acquire(ctx, req.LatchSpans, req.PoisonPolicy)
	if err != nil {
		__antithesis_instrumentation__.Notify(99150)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(99151)
	}
	__antithesis_instrumentation__.Notify(99149)
	return lg, nil
}

func (m *latchManagerImpl) AcquireOptimistic(req Request) latchGuard {
	__antithesis_instrumentation__.Notify(99152)
	lg := m.m.AcquireOptimistic(req.LatchSpans, req.PoisonPolicy)
	return lg
}

func (m *latchManagerImpl) CheckOptimisticNoConflicts(lg latchGuard, spans *spanset.SpanSet) bool {
	__antithesis_instrumentation__.Notify(99153)
	return m.m.CheckOptimisticNoConflicts(lg.(*spanlatch.Guard), spans)
}

func (m *latchManagerImpl) WaitUntilAcquired(
	ctx context.Context, lg latchGuard,
) (latchGuard, *Error) {
	__antithesis_instrumentation__.Notify(99154)
	lg, err := m.m.WaitUntilAcquired(ctx, lg.(*spanlatch.Guard))
	if err != nil {
		__antithesis_instrumentation__.Notify(99156)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(99157)
	}
	__antithesis_instrumentation__.Notify(99155)
	return lg, nil
}

func (m *latchManagerImpl) WaitFor(
	ctx context.Context, ss *spanset.SpanSet, pp poison.Policy,
) *Error {
	__antithesis_instrumentation__.Notify(99158)
	err := m.m.WaitFor(ctx, ss, pp)
	if err != nil {
		__antithesis_instrumentation__.Notify(99160)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(99161)
	}
	__antithesis_instrumentation__.Notify(99159)
	return nil
}

func (m *latchManagerImpl) Poison(lg latchGuard) {
	__antithesis_instrumentation__.Notify(99162)
	m.m.Poison(lg.(*spanlatch.Guard))
}

func (m *latchManagerImpl) Release(lg latchGuard) {
	__antithesis_instrumentation__.Notify(99163)
	m.m.Release(lg.(*spanlatch.Guard))
}

func (m *latchManagerImpl) Metrics() LatchMetrics {
	__antithesis_instrumentation__.Notify(99164)
	return m.m.Metrics()
}
