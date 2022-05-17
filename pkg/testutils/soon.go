package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	DefaultSucceedsSoonDuration = 45 * time.Second

	RaceSucceedsSoonDuration = DefaultSucceedsSoonDuration * 5
)

func SucceedsSoon(t TB, fn func() error) {
	__antithesis_instrumentation__.Notify(646021)
	t.Helper()
	SucceedsWithin(t, fn, succeedsSoonDuration())
}

func SucceedsSoonError(fn func() error) error {
	__antithesis_instrumentation__.Notify(646022)
	return SucceedsWithinError(fn, succeedsSoonDuration())
}

func SucceedsWithin(t TB, fn func() error, duration time.Duration) {
	__antithesis_instrumentation__.Notify(646023)
	t.Helper()
	if err := SucceedsWithinError(fn, duration); err != nil {
		__antithesis_instrumentation__.Notify(646024)
		t.Fatalf("condition failed to evaluate within %s: %s\n%s",
			duration, err, string(debug.Stack()))
	} else {
		__antithesis_instrumentation__.Notify(646025)
	}
}

func SucceedsWithinError(fn func() error, duration time.Duration) error {
	__antithesis_instrumentation__.Notify(646026)
	tBegin := timeutil.Now()
	wrappedFn := func() error {
		__antithesis_instrumentation__.Notify(646028)
		err := fn()
		if timeutil.Since(tBegin) > 3*time.Second && func() bool {
			__antithesis_instrumentation__.Notify(646030)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(646031)
			log.InfofDepth(context.Background(), 4, "SucceedsSoon: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(646032)
		}
		__antithesis_instrumentation__.Notify(646029)
		return err
	}
	__antithesis_instrumentation__.Notify(646027)
	return retry.ForDuration(duration, wrappedFn)
}

func succeedsSoonDuration() time.Duration {
	__antithesis_instrumentation__.Notify(646033)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(646035)
		return RaceSucceedsSoonDuration
	} else {
		__antithesis_instrumentation__.Notify(646036)
	}
	__antithesis_instrumentation__.Notify(646034)
	return DefaultSucceedsSoonDuration
}
