package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdcutils"
)

type throttlingBuffer struct {
	Buffer
	throttle *cdcutils.Throttler
}

func NewThrottlingBuffer(b Buffer, throttle *cdcutils.Throttler) Buffer {
	__antithesis_instrumentation__.Notify(17147)
	return &throttlingBuffer{Buffer: b, throttle: throttle}
}

func (b *throttlingBuffer) Get(ctx context.Context) (Event, error) {
	__antithesis_instrumentation__.Notify(17148)
	evt, err := b.Buffer.Get(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(17151)
		return evt, err
	} else {
		__antithesis_instrumentation__.Notify(17152)
	}
	__antithesis_instrumentation__.Notify(17149)

	if err := b.throttle.AcquireMessageQuota(ctx, evt.ApproximateSize()); err != nil {
		__antithesis_instrumentation__.Notify(17153)
		return Event{}, err
	} else {
		__antithesis_instrumentation__.Notify(17154)
	}
	__antithesis_instrumentation__.Notify(17150)

	return evt, nil
}
