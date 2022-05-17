package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type ResourceLimiterOptions struct {
	MaxRunTime time.Duration
}

type ResourceLimitReached int

const (
	ResourceLimitNotReached  ResourceLimitReached = 0
	ResourceLimitReachedSoft ResourceLimitReached = 1
	ResourceLimitReachedHard ResourceLimitReached = 2
)

type ResourceLimiter interface {
	IsExhausted() ResourceLimitReached
}

type TimeResourceLimiter struct {
	softMaxRunTime time.Duration
	hardMaxRunTime time.Duration
	startTime      time.Time
	ts             timeutil.TimeSource
}

var _ ResourceLimiter = &TimeResourceLimiter{}

func NewResourceLimiter(opts ResourceLimiterOptions, ts timeutil.TimeSource) ResourceLimiter {
	__antithesis_instrumentation__.Notify(643651)
	if opts.MaxRunTime == 0 {
		__antithesis_instrumentation__.Notify(643653)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643654)
	}
	__antithesis_instrumentation__.Notify(643652)
	softTimeLimit := time.Duration(float64(opts.MaxRunTime) * 0.9)
	return &TimeResourceLimiter{hardMaxRunTime: opts.MaxRunTime, softMaxRunTime: softTimeLimit, startTime: ts.Now(), ts: ts}
}

func (l *TimeResourceLimiter) IsExhausted() ResourceLimitReached {
	__antithesis_instrumentation__.Notify(643655)
	timePassed := l.ts.Since(l.startTime)
	if timePassed < l.softMaxRunTime {
		__antithesis_instrumentation__.Notify(643658)
		return ResourceLimitNotReached
	} else {
		__antithesis_instrumentation__.Notify(643659)
	}
	__antithesis_instrumentation__.Notify(643656)
	if timePassed < l.hardMaxRunTime {
		__antithesis_instrumentation__.Notify(643660)
		return ResourceLimitReachedSoft
	} else {
		__antithesis_instrumentation__.Notify(643661)
	}
	__antithesis_instrumentation__.Notify(643657)
	return ResourceLimitReachedHard
}
