package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/facebookgo/clock"
)

const maxBackoff = time.Second

type breakerClock struct {
	clock *hlc.Clock
}

func (c *breakerClock) After(d time.Duration) <-chan time.Time {
	__antithesis_instrumentation__.Notify(184181)
	return time.After(d)
}

func (c *breakerClock) AfterFunc(d time.Duration, f func()) *clock.Timer {
	__antithesis_instrumentation__.Notify(184182)
	panic("unimplemented")
}

func (c *breakerClock) Now() time.Time {
	__antithesis_instrumentation__.Notify(184183)
	return c.clock.PhysicalTime()
}

func (c *breakerClock) Sleep(d time.Duration) {
	__antithesis_instrumentation__.Notify(184184)
	panic("unimplemented")
}

func (c *breakerClock) Tick(d time.Duration) <-chan time.Time {
	__antithesis_instrumentation__.Notify(184185)
	panic("unimplemented")
}

func (c *breakerClock) Ticker(d time.Duration) *clock.Ticker {
	__antithesis_instrumentation__.Notify(184186)
	panic("unimplemented")
}

func (c *breakerClock) Timer(d time.Duration) *clock.Timer {
	__antithesis_instrumentation__.Notify(184187)
	panic("unimplemented")
}

var _ clock.Clock = &breakerClock{}

func newBackOff(clock backoff.Clock) backoff.BackOff {
	__antithesis_instrumentation__.Notify(184188)

	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         maxBackoff,
		MaxElapsedTime:      0,
		Clock:               clock,
	}
	b.Reset()
	return b
}

func newBreaker(ctx context.Context, name string, clock clock.Clock) *circuit.Breaker {
	__antithesis_instrumentation__.Notify(184189)
	return circuit.NewBreakerWithOptions(&circuit.Options{
		Name:       name,
		BackOff:    newBackOff(clock),
		Clock:      clock,
		ShouldTrip: circuit.ThresholdTripFunc(1),
		Logger:     breakerLogger{ctx},
	})
}

type breakerLogger struct {
	ctx context.Context
}

func (r breakerLogger) Debugf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(184190)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(184191)
		log.Dev.InfofDepth(r.ctx, 1, format, v...)
	} else {
		__antithesis_instrumentation__.Notify(184192)
	}
}

func (r breakerLogger) Infof(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(184193)
	log.Ops.InfofDepth(r.ctx, 1, format, v...)
}
