package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

var (
	forwardClockJumpCheckEnabled = settings.RegisterBoolSetting(
		settings.TenantWritable,
		"server.clock.forward_jump_check_enabled",
		"if enabled, forward clock jumps > max_offset/2 will cause a panic",
		false,
	).WithPublic()

	persistHLCUpperBoundInterval = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.clock.persist_upper_bound_interval",
		"the interval between persisting the wall time upper bound of the clock. The clock "+
			"does not generate a wall time greater than the persisted timestamp and will panic if "+
			"it sees a wall time greater than this value. When cockroach starts, it waits for the "+
			"wall time to catch-up till this persisted timestamp. This guarantees monotonic wall "+
			"time across server restarts. Not setting this or setting a value of 0 disables this "+
			"feature.",
		0,
	).WithPublic()
)

func (s *Server) startMonitoringForwardClockJumps(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(189667)
	forwardJumpCheckEnabled := make(chan bool, 1)
	s.stopper.AddCloser(stop.CloserFn(func() { __antithesis_instrumentation__.Notify(189671); close(forwardJumpCheckEnabled) }))
	__antithesis_instrumentation__.Notify(189668)

	forwardClockJumpCheckEnabled.SetOnChange(&s.st.SV, func(context.Context) {
		__antithesis_instrumentation__.Notify(189672)
		forwardJumpCheckEnabled <- forwardClockJumpCheckEnabled.Get(&s.st.SV)
	})
	__antithesis_instrumentation__.Notify(189669)

	if err := s.clock.StartMonitoringForwardClockJumps(
		ctx,
		forwardJumpCheckEnabled,
		time.NewTicker,
		nil,
	); err != nil {
		__antithesis_instrumentation__.Notify(189673)
		return errors.Wrap(err, "monitoring forward clock jumps")
	} else {
		__antithesis_instrumentation__.Notify(189674)
	}
	__antithesis_instrumentation__.Notify(189670)

	log.Ops.Info(ctx, "monitoring forward clock jumps based on server.clock.forward_jump_check_enabled")
	return nil
}

func (s *Server) checkHLCUpperBoundExistsAndEnsureMonotonicity(
	ctx context.Context, initialStart bool,
) (hlcUpperBoundExists bool, err error) {
	__antithesis_instrumentation__.Notify(189675)
	if initialStart {
		__antithesis_instrumentation__.Notify(189678)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(189679)
	}
	__antithesis_instrumentation__.Notify(189676)

	hlcUpperBound, err := kvserver.ReadMaxHLCUpperBound(ctx, s.engines)
	if err != nil {
		__antithesis_instrumentation__.Notify(189680)
		return false, errors.Wrap(err, "reading max HLC upper bound")
	} else {
		__antithesis_instrumentation__.Notify(189681)
	}
	__antithesis_instrumentation__.Notify(189677)
	hlcUpperBoundExists = hlcUpperBound > 0

	ensureClockMonotonicity(
		ctx,
		s.clock,
		s.startTime,
		hlcUpperBound,
		s.clock.SleepUntil,
	)

	return hlcUpperBoundExists, nil
}

func ensureClockMonotonicity(
	ctx context.Context,
	clock *hlc.Clock,
	startTime time.Time,
	prevHLCUpperBound int64,
	sleepUntilFn func(context.Context, hlc.Timestamp) error,
) {
	__antithesis_instrumentation__.Notify(189682)
	var sleepUntil int64
	if prevHLCUpperBound != 0 {
		__antithesis_instrumentation__.Notify(189684)

		sleepUntil = prevHLCUpperBound + 1
	} else {
		__antithesis_instrumentation__.Notify(189685)

		sleepUntil = startTime.UnixNano() + int64(clock.MaxOffset()) + 1
	}
	__antithesis_instrumentation__.Notify(189683)

	currentWallTime := clock.Now().WallTime
	delta := time.Duration(sleepUntil - currentWallTime)
	if delta > 0 {
		__antithesis_instrumentation__.Notify(189686)
		log.Ops.Infof(
			ctx,
			"Sleeping till wall time %v to catches up to %v to ensure monotonicity. Delta: %v",
			currentWallTime,
			sleepUntil,
			delta,
		)
		_ = sleepUntilFn(ctx, hlc.Timestamp{WallTime: sleepUntil})
	} else {
		__antithesis_instrumentation__.Notify(189687)
	}
}

func periodicallyPersistHLCUpperBound(
	clock *hlc.Clock,
	persistHLCUpperBoundIntervalCh chan time.Duration,
	persistHLCUpperBoundFn func(int64) error,
	tickerFn func(d time.Duration) *time.Ticker,
	stopCh <-chan struct{},
	tickCallback func(),
) {
	__antithesis_instrumentation__.Notify(189688)

	ticker := tickerFn(time.Hour)
	ticker.Stop()

	var persistInterval time.Duration
	var ok bool

	persistHLCUpperBound := func() {
		__antithesis_instrumentation__.Notify(189690)
		if err := clock.RefreshHLCUpperBound(
			persistHLCUpperBoundFn,
			int64(persistInterval*3),
		); err != nil {
			__antithesis_instrumentation__.Notify(189691)
			log.Ops.Fatalf(
				context.Background(),
				"error persisting HLC upper bound: %v",
				err,
			)
		} else {
			__antithesis_instrumentation__.Notify(189692)
		}
	}
	__antithesis_instrumentation__.Notify(189689)

	for {
		__antithesis_instrumentation__.Notify(189693)
		select {
		case persistInterval, ok = <-persistHLCUpperBoundIntervalCh:
			__antithesis_instrumentation__.Notify(189695)
			ticker.Stop()
			if !ok {
				__antithesis_instrumentation__.Notify(189699)
				return
			} else {
				__antithesis_instrumentation__.Notify(189700)
			}
			__antithesis_instrumentation__.Notify(189696)

			if persistInterval > 0 {
				__antithesis_instrumentation__.Notify(189701)
				ticker = tickerFn(persistInterval)
				persistHLCUpperBound()
				log.Ops.Info(context.Background(), "persisting HLC upper bound is enabled")
			} else {
				__antithesis_instrumentation__.Notify(189702)
				if err := clock.ResetHLCUpperBound(persistHLCUpperBoundFn); err != nil {
					__antithesis_instrumentation__.Notify(189704)
					log.Ops.Fatalf(
						context.Background(),
						"error resetting hlc upper bound: %v",
						err,
					)
				} else {
					__antithesis_instrumentation__.Notify(189705)
				}
				__antithesis_instrumentation__.Notify(189703)
				log.Ops.Info(context.Background(), "persisting HLC upper bound is disabled")
			}

		case <-ticker.C:
			__antithesis_instrumentation__.Notify(189697)
			if persistInterval > 0 {
				__antithesis_instrumentation__.Notify(189706)
				persistHLCUpperBound()
			} else {
				__antithesis_instrumentation__.Notify(189707)
			}

		case <-stopCh:
			__antithesis_instrumentation__.Notify(189698)
			ticker.Stop()
			return
		}
		__antithesis_instrumentation__.Notify(189694)

		if tickCallback != nil {
			__antithesis_instrumentation__.Notify(189708)
			tickCallback()
		} else {
			__antithesis_instrumentation__.Notify(189709)
		}
	}
}

func (s *Server) startPersistingHLCUpperBound(ctx context.Context, hlcUpperBoundExists bool) error {
	__antithesis_instrumentation__.Notify(189710)
	tickerFn := time.NewTicker
	persistHLCUpperBoundFn := func(t int64) error {
		__antithesis_instrumentation__.Notify(189715)
		return s.node.SetHLCUpperBound(context.Background(), t)
	}
	__antithesis_instrumentation__.Notify(189711)
	persistHLCUpperBoundIntervalCh := make(chan time.Duration, 1)
	persistHLCUpperBoundInterval.SetOnChange(&s.st.SV, func(context.Context) {
		__antithesis_instrumentation__.Notify(189716)
		persistHLCUpperBoundIntervalCh <- persistHLCUpperBoundInterval.Get(&s.st.SV)
	})
	__antithesis_instrumentation__.Notify(189712)

	if hlcUpperBoundExists {
		__antithesis_instrumentation__.Notify(189717)

		if err := s.clock.RefreshHLCUpperBound(
			persistHLCUpperBoundFn,
			int64(5*time.Second),
		); err != nil {
			__antithesis_instrumentation__.Notify(189718)
			return errors.Wrap(err, "refreshing HLC upper bound")
		} else {
			__antithesis_instrumentation__.Notify(189719)
		}
	} else {
		__antithesis_instrumentation__.Notify(189720)
	}
	__antithesis_instrumentation__.Notify(189713)

	_ = s.stopper.RunAsyncTask(
		ctx,
		"persist-hlc-upper-bound",
		func(context.Context) {
			__antithesis_instrumentation__.Notify(189721)
			periodicallyPersistHLCUpperBound(
				s.clock,
				persistHLCUpperBoundIntervalCh,
				persistHLCUpperBoundFn,
				tickerFn,
				s.stopper.ShouldQuiesce(),
				nil,
			)
		},
	)
	__antithesis_instrumentation__.Notify(189714)
	return nil
}
