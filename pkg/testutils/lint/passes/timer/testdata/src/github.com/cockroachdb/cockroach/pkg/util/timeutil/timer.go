package timeutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "time"

func NewTimer() *Timer {
	__antithesis_instrumentation__.Notify(645351)
	return &Timer{}
}

type Timer struct {
	C    <-chan time.Time
	Read bool
}

func (t *Timer) Reset(d time.Duration) {
	__antithesis_instrumentation__.Notify(645352)
}
