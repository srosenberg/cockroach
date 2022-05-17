// Package tenanttokenbucket implements the tenant token bucket server-side
// algorithm described in the distributed token bucket RFC. It has minimal
// dependencies and is meant to be testable on its own.
package tenanttokenbucket

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type State struct {
	RUBurstLimit float64

	RURefillRate float64

	RUCurrent float64

	CurrentShareSum float64
}

const fallbackRateTimeFrame = time.Hour

func (s *State) Update(since time.Duration) {
	__antithesis_instrumentation__.Notify(20349)
	if since > 0 {
		__antithesis_instrumentation__.Notify(20350)
		s.RUCurrent += s.RURefillRate * since.Seconds()
	} else {
		__antithesis_instrumentation__.Notify(20351)
	}
}

func (s *State) Request(
	ctx context.Context, req *roachpb.TokenBucketRequest,
) roachpb.TokenBucketResponse {
	__antithesis_instrumentation__.Notify(20352)
	var res roachpb.TokenBucketResponse

	res.FallbackRate = s.RURefillRate
	if s.RUCurrent > 0 {
		__antithesis_instrumentation__.Notify(20361)
		res.FallbackRate += s.RUCurrent / fallbackRateTimeFrame.Seconds()
	} else {
		__antithesis_instrumentation__.Notify(20362)
	}
	__antithesis_instrumentation__.Notify(20353)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(20363)
		log.Infof(ctx, "token bucket request (tenant=%d requested=%g current=%g)", req.TenantID, req.RequestedRU, s.RUCurrent)
	} else {
		__antithesis_instrumentation__.Notify(20364)
	}
	__antithesis_instrumentation__.Notify(20354)

	needed := req.RequestedRU
	if needed <= 0 {
		__antithesis_instrumentation__.Notify(20365)
		return res
	} else {
		__antithesis_instrumentation__.Notify(20366)
	}
	__antithesis_instrumentation__.Notify(20355)

	if s.RUCurrent >= needed {
		__antithesis_instrumentation__.Notify(20367)
		s.RUCurrent -= needed
		res.GrantedRU = needed
		if log.V(1) {
			__antithesis_instrumentation__.Notify(20369)
			log.Infof(ctx, "request granted (tenant=%d remaining=%g)", req.TenantID, s.RUCurrent)
		} else {
			__antithesis_instrumentation__.Notify(20370)
		}
		__antithesis_instrumentation__.Notify(20368)
		return res
	} else {
		__antithesis_instrumentation__.Notify(20371)
	}
	__antithesis_instrumentation__.Notify(20356)

	var grantedTokens float64

	if s.RUCurrent > 0 {
		__antithesis_instrumentation__.Notify(20372)
		grantedTokens = s.RUCurrent
		needed -= s.RUCurrent
	} else {
		__antithesis_instrumentation__.Notify(20373)
	}
	__antithesis_instrumentation__.Notify(20357)

	availableRate := s.RURefillRate
	if debt := -s.RUCurrent; debt > 0 {
		__antithesis_instrumentation__.Notify(20374)

		debt -= req.TargetRequestPeriod.Seconds() * s.RURefillRate
		if debt > 0 {
			__antithesis_instrumentation__.Notify(20375)

			debtRate := debt / req.TargetRequestPeriod.Seconds()
			availableRate -= debtRate
			availableRate = math.Max(availableRate, 0.05*s.RURefillRate)
		} else {
			__antithesis_instrumentation__.Notify(20376)
		}
	} else {
		__antithesis_instrumentation__.Notify(20377)
	}
	__antithesis_instrumentation__.Notify(20358)

	allowedRate := availableRate
	duration := time.Duration(float64(time.Second) * (needed / allowedRate))
	if duration <= req.TargetRequestPeriod {
		__antithesis_instrumentation__.Notify(20378)
		grantedTokens += needed
	} else {
		__antithesis_instrumentation__.Notify(20379)

		duration = req.TargetRequestPeriod
		grantedTokens += allowedRate * duration.Seconds()
	}
	__antithesis_instrumentation__.Notify(20359)
	s.RUCurrent -= grantedTokens
	res.GrantedRU = grantedTokens
	res.TrickleDuration = duration
	if log.V(1) {
		__antithesis_instrumentation__.Notify(20380)
		log.Infof(ctx, "request granted over time (tenant=%d granted=%g trickle=%s)", req.TenantID, res.GrantedRU, res.TrickleDuration)
	} else {
		__antithesis_instrumentation__.Notify(20381)
	}
	__antithesis_instrumentation__.Notify(20360)
	return res
}

func (s *State) Reconfigure(
	ctx context.Context,
	tenantID roachpb.TenantID,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
	now time.Time,
	currentConsumedRequestUnits float64,
) {
	__antithesis_instrumentation__.Notify(20382)

	s.RUCurrent = availableRU
	s.RURefillRate = refillRate
	s.RUBurstLimit = maxBurstRU
	log.Infof(
		ctx, "token bucket for tenant %s reconfigured: available=%g refill-rate=%g burst-limit=%g",
		tenantID.String(), s.RUCurrent, s.RURefillRate, s.RUBurstLimit,
	)
}
