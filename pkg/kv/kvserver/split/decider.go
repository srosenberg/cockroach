package split

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const minSplitSuggestionInterval = time.Minute
const minQueriesPerSecondSampleDuration = time.Second

type Decider struct {
	intn         func(n int) int
	qpsThreshold func() float64
	qpsRetention func() time.Duration

	mu struct {
		syncutil.Mutex

		lastQPSRollover time.Time
		lastQPS         float64
		count           int64

		maxQPS maxQPSTracker

		splitFinder         *Finder
		lastSplitSuggestion time.Time
	}
}

func Init(
	lbs *Decider,
	intn func(n int) int,
	qpsThreshold func() float64,
	qpsRetention func() time.Duration,
) {
	__antithesis_instrumentation__.Notify(123354)
	lbs.intn = intn
	lbs.qpsThreshold = qpsThreshold
	lbs.qpsRetention = qpsRetention
}

func (d *Decider) Record(now time.Time, n int, span func() roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(123355)
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.recordLocked(now, n, span)
}

func (d *Decider) recordLocked(now time.Time, n int, span func() roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(123356)
	d.mu.count += int64(n)

	if d.mu.lastQPSRollover.IsZero() {
		__antithesis_instrumentation__.Notify(123360)
		d.mu.lastQPSRollover = now
	} else {
		__antithesis_instrumentation__.Notify(123361)
	}
	__antithesis_instrumentation__.Notify(123357)
	elapsedSinceLastQPS := now.Sub(d.mu.lastQPSRollover)
	if elapsedSinceLastQPS >= minQueriesPerSecondSampleDuration {
		__antithesis_instrumentation__.Notify(123362)

		d.mu.lastQPS = (float64(d.mu.count) / float64(elapsedSinceLastQPS)) * 1e9
		d.mu.lastQPSRollover = now
		d.mu.count = 0

		d.mu.maxQPS.record(now, d.qpsRetention(), d.mu.lastQPS)

		if d.mu.lastQPS >= d.qpsThreshold() {
			__antithesis_instrumentation__.Notify(123363)
			if d.mu.splitFinder == nil {
				__antithesis_instrumentation__.Notify(123364)
				d.mu.splitFinder = NewFinder(now)
			} else {
				__antithesis_instrumentation__.Notify(123365)
			}
		} else {
			__antithesis_instrumentation__.Notify(123366)
			d.mu.splitFinder = nil
		}
	} else {
		__antithesis_instrumentation__.Notify(123367)
	}
	__antithesis_instrumentation__.Notify(123358)

	if d.mu.splitFinder != nil && func() bool {
		__antithesis_instrumentation__.Notify(123368)
		return n != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(123369)
		s := span()
		if s.Key != nil {
			__antithesis_instrumentation__.Notify(123371)
			d.mu.splitFinder.Record(span(), d.intn)
		} else {
			__antithesis_instrumentation__.Notify(123372)
		}
		__antithesis_instrumentation__.Notify(123370)
		if now.Sub(d.mu.lastSplitSuggestion) > minSplitSuggestionInterval && func() bool {
			__antithesis_instrumentation__.Notify(123373)
			return d.mu.splitFinder.Ready(now) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(123374)
			return d.mu.splitFinder.Key() != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(123375)
			d.mu.lastSplitSuggestion = now
			return true
		} else {
			__antithesis_instrumentation__.Notify(123376)
		}
	} else {
		__antithesis_instrumentation__.Notify(123377)
	}
	__antithesis_instrumentation__.Notify(123359)
	return false
}

func (d *Decider) RecordMax(now time.Time, qps float64) {
	__antithesis_instrumentation__.Notify(123378)
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mu.maxQPS.record(now, d.qpsRetention(), qps)
}

func (d *Decider) LastQPS(now time.Time) float64 {
	__antithesis_instrumentation__.Notify(123379)
	d.mu.Lock()
	defer d.mu.Unlock()

	d.recordLocked(now, 0, nil)
	return d.mu.lastQPS
}

func (d *Decider) MaxQPS(now time.Time) (float64, bool) {
	__antithesis_instrumentation__.Notify(123380)
	d.mu.Lock()
	defer d.mu.Unlock()

	d.recordLocked(now, 0, nil)
	return d.mu.maxQPS.maxQPS(now, d.qpsRetention())
}

func (d *Decider) MaybeSplitKey(now time.Time) roachpb.Key {
	__antithesis_instrumentation__.Notify(123381)
	var key roachpb.Key

	d.mu.Lock()
	defer d.mu.Unlock()

	d.recordLocked(now, 0, nil)
	if d.mu.splitFinder != nil && func() bool {
		__antithesis_instrumentation__.Notify(123383)
		return d.mu.splitFinder.Ready(now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(123384)

		key = d.mu.splitFinder.Key()
		if safeKey, err := keys.EnsureSafeSplitKey(key); err == nil {
			__antithesis_instrumentation__.Notify(123385)
			key = safeKey
		} else {
			__antithesis_instrumentation__.Notify(123386)
		}
	} else {
		__antithesis_instrumentation__.Notify(123387)
	}
	__antithesis_instrumentation__.Notify(123382)
	return key
}

func (d *Decider) Reset(now time.Time) {
	__antithesis_instrumentation__.Notify(123388)
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mu.lastQPSRollover = time.Time{}
	d.mu.lastQPS = 0
	d.mu.count = 0
	d.mu.maxQPS.reset(now, d.qpsRetention())
	d.mu.splitFinder = nil
	d.mu.lastSplitSuggestion = time.Time{}
}

type maxQPSTracker struct {
	windows      [6]float64
	curIdx       int
	curStart     time.Time
	lastReset    time.Time
	minRetention time.Duration
}

func (t *maxQPSTracker) record(now time.Time, minRetention time.Duration, qps float64) {
	__antithesis_instrumentation__.Notify(123389)
	t.maybeReset(now, minRetention)
	t.maybeRotate(now)
	t.windows[t.curIdx] = max(t.windows[t.curIdx], qps)
}

func (t *maxQPSTracker) reset(now time.Time, minRetention time.Duration) {
	__antithesis_instrumentation__.Notify(123390)
	if minRetention <= 0 {
		__antithesis_instrumentation__.Notify(123392)
		panic("minRetention must be positive")
	} else {
		__antithesis_instrumentation__.Notify(123393)
	}
	__antithesis_instrumentation__.Notify(123391)
	t.windows = [6]float64{}
	t.curIdx = 0
	t.curStart = now
	t.lastReset = now
	t.minRetention = minRetention
}

func (t *maxQPSTracker) maybeReset(now time.Time, minRetention time.Duration) {
	__antithesis_instrumentation__.Notify(123394)

	if minRetention != t.minRetention {
		__antithesis_instrumentation__.Notify(123395)
		t.reset(now, minRetention)
	} else {
		__antithesis_instrumentation__.Notify(123396)
	}
}

func (t *maxQPSTracker) maybeRotate(now time.Time) {
	__antithesis_instrumentation__.Notify(123397)
	sinceLastRotate := now.Sub(t.curStart)
	windowWidth := t.windowWidth()
	if sinceLastRotate < windowWidth {
		__antithesis_instrumentation__.Notify(123400)

		return
	} else {
		__antithesis_instrumentation__.Notify(123401)
	}
	__antithesis_instrumentation__.Notify(123398)

	shift := int(sinceLastRotate / windowWidth)
	if shift >= len(t.windows) {
		__antithesis_instrumentation__.Notify(123402)

		t.windows = [6]float64{}
		t.curIdx = 0
		t.curStart = now
		return
	} else {
		__antithesis_instrumentation__.Notify(123403)
	}
	__antithesis_instrumentation__.Notify(123399)
	for i := 0; i < shift; i++ {
		__antithesis_instrumentation__.Notify(123404)
		t.curIdx = (t.curIdx + 1) % len(t.windows)
		t.curStart = t.curStart.Add(windowWidth)
		t.windows[t.curIdx] = 0
	}
}

func (t *maxQPSTracker) maxQPS(now time.Time, minRetention time.Duration) (float64, bool) {
	__antithesis_instrumentation__.Notify(123405)
	t.record(now, minRetention, 0)

	if now.Sub(t.lastReset) < t.minRetention {
		__antithesis_instrumentation__.Notify(123408)

		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(123409)
	}
	__antithesis_instrumentation__.Notify(123406)

	qps := 0.0
	for _, v := range t.windows {
		__antithesis_instrumentation__.Notify(123410)
		qps = max(qps, v)
	}
	__antithesis_instrumentation__.Notify(123407)
	return qps, true
}

func (t *maxQPSTracker) windowWidth() time.Duration {
	__antithesis_instrumentation__.Notify(123411)

	return t.minRetention / time.Duration(len(t.windows)-1)
}

func max(a, b float64) float64 {
	__antithesis_instrumentation__.Notify(123412)
	if a > b {
		__antithesis_instrumentation__.Notify(123414)
		return a
	} else {
		__antithesis_instrumentation__.Notify(123415)
	}
	__antithesis_instrumentation__.Notify(123413)
	return b
}
