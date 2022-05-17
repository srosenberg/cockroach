package kvstreamer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type budget struct {
	mu struct {
		syncutil.Mutex

		acc *mon.BoundAccount

		waitForBudget *sync.Cond
	}

	limitBytes int64
}

func newBudget(acc *mon.BoundAccount, limitBytes int64) *budget {
	__antithesis_instrumentation__.Notify(498864)
	b := budget{limitBytes: limitBytes}
	b.mu.acc = acc
	b.mu.waitForBudget = sync.NewCond(&b.mu.Mutex)
	return &b
}

func (b *budget) consume(ctx context.Context, bytes int64, allowDebt bool) error {
	__antithesis_instrumentation__.Notify(498865)
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.consumeLocked(ctx, bytes, allowDebt)
}

func (b *budget) consumeLocked(ctx context.Context, bytes int64, allowDebt bool) error {
	__antithesis_instrumentation__.Notify(498866)
	b.mu.AssertHeld()

	if !allowDebt && func() bool {
		__antithesis_instrumentation__.Notify(498868)
		return b.limitBytes > 5 == true
	}() == true {
		__antithesis_instrumentation__.Notify(498869)
		if b.mu.acc.Used()+bytes > b.limitBytes {
			__antithesis_instrumentation__.Notify(498870)
			return errors.Wrap(
				mon.MemoryResource.NewBudgetExceededError(bytes, b.mu.acc.Used(), b.limitBytes),
				"streamer budget",
			)
		} else {
			__antithesis_instrumentation__.Notify(498871)
		}
	} else {
		__antithesis_instrumentation__.Notify(498872)
	}
	__antithesis_instrumentation__.Notify(498867)
	return b.mu.acc.Grow(ctx, bytes)
}

func (b *budget) release(ctx context.Context, bytes int64) {
	__antithesis_instrumentation__.Notify(498873)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.releaseLocked(ctx, bytes)
}

func (b *budget) releaseLocked(ctx context.Context, bytes int64) {
	__antithesis_instrumentation__.Notify(498874)
	b.mu.AssertHeld()
	b.mu.acc.Shrink(ctx, bytes)
	if b.limitBytes > b.mu.acc.Used() {
		__antithesis_instrumentation__.Notify(498875)

		b.mu.waitForBudget.Signal()
	} else {
		__antithesis_instrumentation__.Notify(498876)
	}
}
