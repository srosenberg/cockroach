package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var resetHighWaterMarkInterval = func() time.Duration {
	__antithesis_instrumentation__.Notify(193576)
	dur := envutil.EnvOrDefaultDuration("COCKROACH_MEMPROF_INTERVAL", time.Hour)
	if dur <= 0 {
		__antithesis_instrumentation__.Notify(193578)

		return 0
	} else {
		__antithesis_instrumentation__.Notify(193579)
	}
	__antithesis_instrumentation__.Notify(193577)
	return dur
}()

const timestampFormat = "2006-01-02T15_04_05.000"

type testingKnobs struct {
	dontWriteProfiles    bool
	maybeTakeProfileHook func(willTakeProfile bool)
	now                  func() time.Time
}

type profiler struct {
	store *profileStore

	lastProfileTime time.Time

	highwaterMarkBytes int64

	knobs testingKnobs
}

func (o *profiler) now() time.Time {
	__antithesis_instrumentation__.Notify(193580)
	if o.knobs.now != nil {
		__antithesis_instrumentation__.Notify(193582)
		return o.knobs.now()
	} else {
		__antithesis_instrumentation__.Notify(193583)
	}
	__antithesis_instrumentation__.Notify(193581)
	return timeutil.Now()
}

func (o *profiler) maybeTakeProfile(
	ctx context.Context,
	thresholdValue int64,
	takeProfileFn func(ctx context.Context, path string) bool,
) {
	__antithesis_instrumentation__.Notify(193584)
	if resetHighWaterMarkInterval == 0 {
		__antithesis_instrumentation__.Notify(193590)

		return
	} else {
		__antithesis_instrumentation__.Notify(193591)
	}
	__antithesis_instrumentation__.Notify(193585)

	now := o.now()

	if now.Sub(o.lastProfileTime) >= resetHighWaterMarkInterval {
		__antithesis_instrumentation__.Notify(193592)
		o.highwaterMarkBytes = 0
	} else {
		__antithesis_instrumentation__.Notify(193593)
	}
	__antithesis_instrumentation__.Notify(193586)

	takeProfile := thresholdValue > o.highwaterMarkBytes
	if hook := o.knobs.maybeTakeProfileHook; hook != nil {
		__antithesis_instrumentation__.Notify(193594)
		hook(takeProfile)
	} else {
		__antithesis_instrumentation__.Notify(193595)
	}
	__antithesis_instrumentation__.Notify(193587)
	if !takeProfile {
		__antithesis_instrumentation__.Notify(193596)
		return
	} else {
		__antithesis_instrumentation__.Notify(193597)
	}
	__antithesis_instrumentation__.Notify(193588)

	o.highwaterMarkBytes = thresholdValue
	o.lastProfileTime = now

	if o.knobs.dontWriteProfiles {
		__antithesis_instrumentation__.Notify(193598)
		return
	} else {
		__antithesis_instrumentation__.Notify(193599)
	}
	__antithesis_instrumentation__.Notify(193589)
	success := takeProfileFn(ctx, o.store.makeNewFileName(now, thresholdValue))
	if success {
		__antithesis_instrumentation__.Notify(193600)

		o.store.gcProfiles(ctx, now)
	} else {
		__antithesis_instrumentation__.Notify(193601)
	}
}
