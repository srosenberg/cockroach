package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/prometheus"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/errors"
	"github.com/prometheus/common/model"
)

type tpccChaosEventProcessor struct {
	workloadInstances []workloadInstance
	workloadNodeIP    string
	ops               []string
	ch                chan ChaosEvent
	promClient        prometheus.Client
	errs              []error

	allowZeroSuccessDuringUptime bool

	maxErrorsDuringUptime int
}

func (ep *tpccChaosEventProcessor) checkUptime(
	ctx context.Context,
	l *logger.Logger,
	op string,
	w workloadInstance,
	from time.Time,
	to time.Time,
) error {
	__antithesis_instrumentation__.Notify(47481)
	return ep.checkMetrics(
		ctx,
		l,
		op,
		w,
		from,
		to,
		func(from, to model.SampleValue) error {
			__antithesis_instrumentation__.Notify(47482)
			if to > from {
				__antithesis_instrumentation__.Notify(47485)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47486)
			}
			__antithesis_instrumentation__.Notify(47483)

			if ep.allowZeroSuccessDuringUptime && func() bool {
				__antithesis_instrumentation__.Notify(47487)
				return to == from == true
			}() == true {
				__antithesis_instrumentation__.Notify(47488)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47489)
			}
			__antithesis_instrumentation__.Notify(47484)
			return errors.Newf("expected successes to be increasing, found from %f, to %f", from, to)
		},
		func(from, to model.SampleValue) error {
			__antithesis_instrumentation__.Notify(47490)

			if to <= from+model.SampleValue(ep.maxErrorsDuringUptime) {
				__antithesis_instrumentation__.Notify(47492)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47493)
			}
			__antithesis_instrumentation__.Notify(47491)
			return errors.Newf("expected <=%d errors, found from %f, to %f", ep.maxErrorsDuringUptime, from, to)
		},
	)
}

func (ep *tpccChaosEventProcessor) checkDowntime(
	ctx context.Context,
	l *logger.Logger,
	op string,
	w workloadInstance,
	from time.Time,
	to time.Time,
) error {
	__antithesis_instrumentation__.Notify(47494)
	return ep.checkMetrics(
		ctx,
		l,
		op,
		w,
		from,
		to,
		func(from, to model.SampleValue) error {
			__antithesis_instrumentation__.Notify(47495)
			if to == from {
				__antithesis_instrumentation__.Notify(47497)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47498)
			}
			__antithesis_instrumentation__.Notify(47496)
			return errors.Newf("expected successes to not increase, found from %f, to %f", from, to)
		},
		func(from, to model.SampleValue) error {
			__antithesis_instrumentation__.Notify(47499)
			if to > from {
				__antithesis_instrumentation__.Notify(47501)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47502)
			}
			__antithesis_instrumentation__.Notify(47500)
			return errors.Newf("expected errors, found from %f, to %f", from, to)
		},
	)
}

func (ep *tpccChaosEventProcessor) checkMetrics(
	ctx context.Context,
	l *logger.Logger,
	op string,
	w workloadInstance,
	fromTime time.Time,
	toTime time.Time,
	successCheckFn func(from, to model.SampleValue) error,
	errorCheckFn func(from, to model.SampleValue) error,
) error {
	__antithesis_instrumentation__.Notify(47503)

	fromTime = fromTime.Add(prometheus.DefaultScrapeInterval)

	toTime = toTime.Add(-prometheus.DefaultScrapeInterval)
	if !toTime.After(fromTime) {
		__antithesis_instrumentation__.Notify(47506)
		l.PrintfCtx(
			ctx,
			"to %s < from %s, skipping",
			toTime.Format(time.RFC3339),
			fromTime.Format(time.RFC3339),
		)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(47507)
	}
	__antithesis_instrumentation__.Notify(47504)

	for _, check := range []struct {
		metricType string
		checkFn    func(from, to model.SampleValue) error
	}{
		{
			metricType: "success",
			checkFn:    successCheckFn,
		},
		{
			metricType: "error",
			checkFn:    errorCheckFn,
		},
	} {
		__antithesis_instrumentation__.Notify(47508)

		q := fmt.Sprintf(
			`workload_tpcc_%s_%s_total{instance="%s:%d"}`,
			op,
			check.metricType,
			ep.workloadNodeIP,
			w.prometheusPort,
		)
		fromVal, warnings, err := ep.promClient.Query(
			ctx,
			q,
			fromTime,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47515)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47516)
		}
		__antithesis_instrumentation__.Notify(47509)
		if len(warnings) > 0 {
			__antithesis_instrumentation__.Notify(47517)
			return errors.Newf("found warnings querying prometheus: %s", warnings)
		} else {
			__antithesis_instrumentation__.Notify(47518)
		}
		__antithesis_instrumentation__.Notify(47510)
		toVal, warnings, err := ep.promClient.Query(
			ctx,
			q,
			toTime,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47519)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47520)
		}
		__antithesis_instrumentation__.Notify(47511)
		if len(warnings) > 0 {
			__antithesis_instrumentation__.Notify(47521)
			return errors.Newf("found warnings querying prometheus: %s", warnings)
		} else {
			__antithesis_instrumentation__.Notify(47522)
		}
		__antithesis_instrumentation__.Notify(47512)

		fromVec := fromVal.(model.Vector)
		if len(fromVec) == 0 {
			__antithesis_instrumentation__.Notify(47523)
			return errors.Newf("unexpected empty fromVec for %s @ %s", q, fromTime.Format(time.RFC3339))
		} else {
			__antithesis_instrumentation__.Notify(47524)
		}
		__antithesis_instrumentation__.Notify(47513)
		from := fromVec[0].Value
		toVec := toVal.(model.Vector)
		if len(toVec) == 0 {
			__antithesis_instrumentation__.Notify(47525)
			return errors.Newf("unexpected empty toVec for %s @ %s", q, toTime.Format(time.RFC3339))
		} else {
			__antithesis_instrumentation__.Notify(47526)
		}
		__antithesis_instrumentation__.Notify(47514)
		to := toVec[0].Value

		l.PrintfCtx(
			ctx,
			"metric %s, from %f @ %s, to %f @ %s\n",
			q,
			from,
			fromTime.Format(time.RFC3339),
			to,
			toTime.Format(time.RFC3339),
		)

		if err := check.checkFn(from, to); err != nil {
			__antithesis_instrumentation__.Notify(47527)
			return errors.Wrapf(
				err,
				"error at from %s, to %s on metric %s",
				fromTime.Format(time.RFC3339),
				toTime.Format(time.RFC3339),
				q,
			)
		} else {
			__antithesis_instrumentation__.Notify(47528)
		}
	}
	__antithesis_instrumentation__.Notify(47505)
	return nil
}

func (ep *tpccChaosEventProcessor) writeErr(ctx context.Context, l *logger.Logger, err error) {
	__antithesis_instrumentation__.Notify(47529)
	l.PrintfCtx(ctx, "error during chaos: %v", err)
	ep.errs = append(ep.errs, err)
}

func (ep *tpccChaosEventProcessor) err() error {
	__antithesis_instrumentation__.Notify(47530)
	var err error
	for i := range ep.errs {
		__antithesis_instrumentation__.Notify(47532)
		if i == 0 {
			__antithesis_instrumentation__.Notify(47533)
			err = ep.errs[i]
		} else {
			__antithesis_instrumentation__.Notify(47534)
			err = errors.CombineErrors(err, ep.errs[i])
		}
	}
	__antithesis_instrumentation__.Notify(47531)
	return err
}

func (ep *tpccChaosEventProcessor) listen(ctx context.Context, l *logger.Logger) {
	__antithesis_instrumentation__.Notify(47535)
	go func() {
		__antithesis_instrumentation__.Notify(47536)
		var prevTime time.Time
		started := false
		for ev := range ep.ch {
			__antithesis_instrumentation__.Notify(47537)
			switch ev.Type {
			case ChaosEventTypeStart,
				ChaosEventTypeEnd,
				ChaosEventTypeShutdownComplete,
				ChaosEventTypeStartupComplete:
				__antithesis_instrumentation__.Notify(47539)

			case ChaosEventTypePreShutdown:
				__antithesis_instrumentation__.Notify(47540)

				if !started {
					__antithesis_instrumentation__.Notify(47544)
					started = true
					break
				} else {
					__antithesis_instrumentation__.Notify(47545)
				}
				__antithesis_instrumentation__.Notify(47541)
				for _, op := range ep.ops {
					__antithesis_instrumentation__.Notify(47546)
					for _, w := range ep.workloadInstances {
						__antithesis_instrumentation__.Notify(47547)
						if err := ep.checkUptime(
							ctx,
							l,
							op,
							w,
							prevTime,
							ev.Time,
						); err != nil {
							__antithesis_instrumentation__.Notify(47548)
							ep.writeErr(ctx, l, err)
						} else {
							__antithesis_instrumentation__.Notify(47549)
						}
					}
				}
			case ChaosEventTypePreStartup:
				__antithesis_instrumentation__.Notify(47542)
				for _, op := range ep.ops {
					__antithesis_instrumentation__.Notify(47550)
					for _, w := range ep.workloadInstances {
						__antithesis_instrumentation__.Notify(47551)
						if w.nodes.Equals(ev.Target) {
							__antithesis_instrumentation__.Notify(47552)

							if err := ep.checkDowntime(
								ctx,
								l,
								op,
								w,
								prevTime,
								ev.Time,
							); err != nil {
								__antithesis_instrumentation__.Notify(47553)
								ep.writeErr(ctx, l, err)
							} else {
								__antithesis_instrumentation__.Notify(47554)
							}
						} else {
							__antithesis_instrumentation__.Notify(47555)
							if err := ep.checkUptime(
								ctx,
								l,
								op,
								w,
								prevTime,
								ev.Time,
							); err != nil {
								__antithesis_instrumentation__.Notify(47556)
								ep.writeErr(ctx, l, err)
							} else {
								__antithesis_instrumentation__.Notify(47557)
							}
						}
					}
				}
			default:
				__antithesis_instrumentation__.Notify(47543)
			}
			__antithesis_instrumentation__.Notify(47538)
			prevTime = ev.Time
		}
	}()
}
