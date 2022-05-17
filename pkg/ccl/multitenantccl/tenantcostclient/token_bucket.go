package tenantcostclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
)

type tokenBucket struct {
	notifyCh chan struct{}

	notifyThreshold tenantcostmodel.RU

	rate tenantcostmodel.RU

	available tenantcostmodel.RU

	debt tenantcostmodel.RU

	debtRate tenantcostmodel.RU

	lastUpdated time.Time
}

const debtRepaymentSecs = 2

func (tb *tokenBucket) Init(
	now time.Time, notifyCh chan struct{}, rate, available tenantcostmodel.RU,
) {
	__antithesis_instrumentation__.Notify(20142)
	*tb = tokenBucket{
		notifyCh:        notifyCh,
		notifyThreshold: 0,
		rate:            rate,
		available:       available,
		lastUpdated:     now,
	}
}

func (tb *tokenBucket) update(now time.Time) {
	__antithesis_instrumentation__.Notify(20143)
	since := now.Sub(tb.lastUpdated)
	if since <= 0 {
		__antithesis_instrumentation__.Notify(20147)
		return
	} else {
		__antithesis_instrumentation__.Notify(20148)
	}
	__antithesis_instrumentation__.Notify(20144)
	tb.lastUpdated = now
	sinceSeconds := since.Seconds()
	refilled := tb.rate * tenantcostmodel.RU(sinceSeconds)

	if tb.debt == 0 {
		__antithesis_instrumentation__.Notify(20149)

		tb.available += refilled
		return
	} else {
		__antithesis_instrumentation__.Notify(20150)
	}
	__antithesis_instrumentation__.Notify(20145)

	debtPaid := tb.debtRate * tenantcostmodel.RU(sinceSeconds)
	if tb.debt >= debtPaid {
		__antithesis_instrumentation__.Notify(20151)
		tb.debt -= debtPaid
	} else {
		__antithesis_instrumentation__.Notify(20152)
		debtPaid = tb.debt
		tb.debt = 0
	}
	__antithesis_instrumentation__.Notify(20146)
	tb.available += refilled - debtPaid
}

func (tb *tokenBucket) notify() {
	__antithesis_instrumentation__.Notify(20153)
	tb.notifyThreshold = 0
	select {
	case tb.notifyCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(20154)
	default:
		__antithesis_instrumentation__.Notify(20155)
	}
}

func (tb *tokenBucket) maybeNotify(now time.Time) {
	__antithesis_instrumentation__.Notify(20156)
	if tb.notifyThreshold > 0 && func() bool {
		__antithesis_instrumentation__.Notify(20157)
		return tb.available < tb.notifyThreshold == true
	}() == true {
		__antithesis_instrumentation__.Notify(20158)
		tb.notify()
	} else {
		__antithesis_instrumentation__.Notify(20159)
	}
}

func (tb *tokenBucket) calculateDebtRate() {
	__antithesis_instrumentation__.Notify(20160)
	tb.debtRate = tb.debt / debtRepaymentSecs
	if tb.debtRate > tb.rate {
		__antithesis_instrumentation__.Notify(20161)
		tb.debtRate = tb.rate
	} else {
		__antithesis_instrumentation__.Notify(20162)
	}
}

func (tb *tokenBucket) RemoveTokens(now time.Time, amount tenantcostmodel.RU) {
	__antithesis_instrumentation__.Notify(20163)
	tb.update(now)
	if tb.available >= amount {
		__antithesis_instrumentation__.Notify(20165)
		tb.available -= amount
	} else {
		__antithesis_instrumentation__.Notify(20166)
		tb.debt += amount - tb.available
		tb.available = 0
		tb.calculateDebtRate()
	}
	__antithesis_instrumentation__.Notify(20164)
	tb.maybeNotify(now)
}

type tokenBucketReconfigureArgs struct {
	NewTokens tenantcostmodel.RU

	NewRate tenantcostmodel.RU

	NotifyThreshold tenantcostmodel.RU
}

func (tb *tokenBucket) Reconfigure(now time.Time, args tokenBucketReconfigureArgs) {
	__antithesis_instrumentation__.Notify(20167)
	tb.update(now)

	select {
	case <-tb.notifyCh:
		__antithesis_instrumentation__.Notify(20170)
	default:
		__antithesis_instrumentation__.Notify(20171)
	}
	__antithesis_instrumentation__.Notify(20168)
	tb.rate = args.NewRate
	tb.notifyThreshold = args.NotifyThreshold
	if args.NewTokens > 0 {
		__antithesis_instrumentation__.Notify(20172)
		if tb.debt > 0 {
			__antithesis_instrumentation__.Notify(20173)
			if tb.debt >= args.NewTokens {
				__antithesis_instrumentation__.Notify(20175)
				tb.debt -= args.NewTokens
			} else {
				__antithesis_instrumentation__.Notify(20176)
				tb.available += args.NewTokens - tb.debt
				tb.debt = 0
			}
			__antithesis_instrumentation__.Notify(20174)
			tb.calculateDebtRate()
		} else {
			__antithesis_instrumentation__.Notify(20177)
			tb.available += args.NewTokens
		}
	} else {
		__antithesis_instrumentation__.Notify(20178)
	}
	__antithesis_instrumentation__.Notify(20169)
	tb.maybeNotify(now)
}

func (tb *tokenBucket) SetupNotification(now time.Time, threshold tenantcostmodel.RU) {
	__antithesis_instrumentation__.Notify(20179)
	tb.update(now)
	tb.notifyThreshold = threshold
}

const maxTryAgainAfterSeconds = 1000

func (tb *tokenBucket) TryToFulfill(
	now time.Time, amount tenantcostmodel.RU,
) (fulfilled bool, tryAgainAfter time.Duration) {
	__antithesis_instrumentation__.Notify(20180)
	tb.update(now)

	if amount <= tb.available {
		__antithesis_instrumentation__.Notify(20186)
		tb.available -= amount
		tb.maybeNotify(now)
		return true, 0
	} else {
		__antithesis_instrumentation__.Notify(20187)
	}
	__antithesis_instrumentation__.Notify(20181)

	if tb.notifyThreshold > 0 {
		__antithesis_instrumentation__.Notify(20188)
		tb.notify()
	} else {
		__antithesis_instrumentation__.Notify(20189)
	}
	__antithesis_instrumentation__.Notify(20182)

	needed := amount - tb.available

	var timeSeconds float64

	if tb.debt == 0 {
		__antithesis_instrumentation__.Notify(20190)
		timeSeconds = float64(needed / tb.rate)
	} else {
		__antithesis_instrumentation__.Notify(20191)
		remainingRate := tb.rate - tb.debtRate

		if needed*tb.debtRate <= tb.debt*remainingRate {
			__antithesis_instrumentation__.Notify(20192)

			timeSeconds = float64(needed / remainingRate)
		} else {
			__antithesis_instrumentation__.Notify(20193)

			debtPaySeconds := tb.debt / tb.debtRate
			timeSeconds = float64(debtPaySeconds + (needed-debtPaySeconds*remainingRate)/tb.rate)
		}
	}
	__antithesis_instrumentation__.Notify(20183)

	if timeSeconds > maxTryAgainAfterSeconds {
		__antithesis_instrumentation__.Notify(20194)
		return false, maxTryAgainAfterSeconds * time.Second
	} else {
		__antithesis_instrumentation__.Notify(20195)
	}
	__antithesis_instrumentation__.Notify(20184)

	timeDelta := time.Duration(timeSeconds * float64(time.Second))
	if timeDelta < time.Nanosecond {
		__antithesis_instrumentation__.Notify(20196)
		timeDelta = time.Nanosecond
	} else {
		__antithesis_instrumentation__.Notify(20197)
	}
	__antithesis_instrumentation__.Notify(20185)
	return false, timeDelta
}

func (tb *tokenBucket) AvailableTokens(now time.Time) tenantcostmodel.RU {
	__antithesis_instrumentation__.Notify(20198)
	tb.update(now)
	return tb.available - tb.debt
}
