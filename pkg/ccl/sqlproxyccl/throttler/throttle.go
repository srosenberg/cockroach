package throttler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "time"

const (
	throttleDisabled = time.Duration(0)
)

type throttle struct {
	nextTime time.Time

	nextBackoff time.Duration
}

func newThrottle(initialBackoff time.Duration) *throttle {
	__antithesis_instrumentation__.Notify(23335)
	return &throttle{
		nextTime:    time.Time{},
		nextBackoff: initialBackoff,
	}
}

func (l *throttle) triggerThrottle(now time.Time, maxBackoff time.Duration) {
	__antithesis_instrumentation__.Notify(23336)
	l.nextTime = now.Add(l.nextBackoff)
	l.nextBackoff *= 2
	if maxBackoff < l.nextBackoff {
		__antithesis_instrumentation__.Notify(23337)
		l.nextBackoff = maxBackoff
	} else {
		__antithesis_instrumentation__.Notify(23338)
	}
}

func (l *throttle) isThrottled(throttleTime time.Time) bool {
	__antithesis_instrumentation__.Notify(23339)
	return l.nextBackoff != throttleDisabled && func() bool {
		__antithesis_instrumentation__.Notify(23340)
		return throttleTime.Before(l.nextTime) == true
	}() == true
}

func (l *throttle) disable() {
	__antithesis_instrumentation__.Notify(23341)
	l.nextBackoff = throttleDisabled
}
