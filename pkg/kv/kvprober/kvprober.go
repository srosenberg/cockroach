// Package kvprober sends queries to KV in a loop, with configurable sleep
// times, in order to generate data about the healthiness or unhealthiness of
// kvclient & below.
//
// Prober increments metrics that SRE & other operators can use as alerting
// signals. It also writes to logs to help narrow down the problem (e.g. which
// range(s) are acting up).
package kvprober

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/logtags"
)

const putValue = "thekvproberwrotethis"

type Prober struct {
	db       *kv.DB
	settings *cluster.Settings

	readPlanner  planner
	writePlanner planner

	metrics Metrics
	tracer  *tracing.Tracer
}

type Opts struct {
	DB       *kv.DB
	Settings *cluster.Settings
	Tracer   *tracing.Tracer

	HistogramWindowInterval time.Duration
}

var (
	metaReadProbeAttempts = metric.Metadata{
		Name:        "kv.prober.read.attempts",
		Help:        "Number of attempts made to read probe KV, regardless of outcome",
		Measurement: "Queries",
		Unit:        metric.Unit_COUNT,
	}
	metaReadProbeFailures = metric.Metadata{
		Name: "kv.prober.read.failures",
		Help: "Number of attempts made to read probe KV that failed, " +
			"whether due to error or timeout",
		Measurement: "Queries",
		Unit:        metric.Unit_COUNT,
	}
	metaReadProbeLatency = metric.Metadata{
		Name:        "kv.prober.read.latency",
		Help:        "Latency of successful KV read probes",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaWriteProbeAttempts = metric.Metadata{
		Name:        "kv.prober.write.attempts",
		Help:        "Number of attempts made to write probe KV, regardless of outcome",
		Measurement: "Queries",
		Unit:        metric.Unit_COUNT,
	}
	metaWriteProbeFailures = metric.Metadata{
		Name: "kv.prober.write.failures",
		Help: "Number of attempts made to write probe KV that failed, " +
			"whether due to error or timeout",
		Measurement: "Queries",
		Unit:        metric.Unit_COUNT,
	}
	metaWriteProbeLatency = metric.Metadata{
		Name:        "kv.prober.write.latency",
		Help:        "Latency of successful KV write probes",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaProbePlanAttempts = metric.Metadata{
		Name: "kv.prober.planning_attempts",
		Help: "Number of attempts at planning out probes made; " +
			"in order to probe KV we need to plan out which ranges to probe;",
		Measurement: "Runs",
		Unit:        metric.Unit_COUNT,
	}
	metaProbePlanFailures = metric.Metadata{
		Name: "kv.prober.planning_failures",
		Help: "Number of attempts at planning out probes that failed; " +
			"in order to probe KV we need to plan out which ranges to probe; " +
			"if planning fails, then kvprober is not able to send probes to " +
			"all ranges; consider alerting on this metric as a result",
		Measurement: "Runs",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	ReadProbeAttempts  *metric.Counter
	ReadProbeFailures  *metric.Counter
	ReadProbeLatency   *metric.Histogram
	WriteProbeAttempts *metric.Counter
	WriteProbeFailures *metric.Counter
	WriteProbeLatency  *metric.Histogram
	ProbePlanAttempts  *metric.Counter
	ProbePlanFailures  *metric.Counter
}

type proberOps interface {
	Read(key interface{}) func(context.Context, *kv.Txn) error
	Write(key interface{}) func(context.Context, *kv.Txn) error
}

type proberTxn interface {
	Txn(context.Context, func(context.Context, *kv.Txn) error) error

	TxnRootKV(context.Context, func(context.Context, *kv.Txn) error) error
}

type proberOpsImpl struct {
}

func (p *proberOpsImpl) Read(key interface{}) func(context.Context, *kv.Txn) error {
	__antithesis_instrumentation__.Notify(93940)
	return func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(93941)
		_, err := txn.Get(ctx, key)
		return err
	}
}

func (p *proberOpsImpl) Write(key interface{}) func(context.Context, *kv.Txn) error {
	__antithesis_instrumentation__.Notify(93942)
	return func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(93943)
		if err := txn.Put(ctx, key, putValue); err != nil {
			__antithesis_instrumentation__.Notify(93945)
			return err
		} else {
			__antithesis_instrumentation__.Notify(93946)
		}
		__antithesis_instrumentation__.Notify(93944)
		return txn.Del(ctx, key)
	}
}

type proberTxnImpl struct {
	db *kv.DB
}

func (p *proberTxnImpl) Txn(ctx context.Context, f func(context.Context, *kv.Txn) error) error {
	__antithesis_instrumentation__.Notify(93947)
	return p.db.Txn(ctx, f)
}

func (p *proberTxnImpl) TxnRootKV(
	ctx context.Context, f func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(93948)
	return p.db.TxnRootKV(ctx, f)
}

func NewProber(opts Opts) *Prober {
	__antithesis_instrumentation__.Notify(93949)
	return &Prober{
		db:       opts.DB,
		settings: opts.Settings,

		readPlanner: newMeta2Planner(opts.DB, opts.Settings, func() time.Duration {
			__antithesis_instrumentation__.Notify(93950)
			return readInterval.Get(&opts.Settings.SV)
		}),
		writePlanner: newMeta2Planner(opts.DB, opts.Settings, func() time.Duration {
			__antithesis_instrumentation__.Notify(93951)
			return writeInterval.Get(&opts.Settings.SV)
		}),

		metrics: Metrics{
			ReadProbeAttempts:  metric.NewCounter(metaReadProbeAttempts),
			ReadProbeFailures:  metric.NewCounter(metaReadProbeFailures),
			ReadProbeLatency:   metric.NewLatency(metaReadProbeLatency, opts.HistogramWindowInterval),
			WriteProbeAttempts: metric.NewCounter(metaWriteProbeAttempts),
			WriteProbeFailures: metric.NewCounter(metaWriteProbeFailures),
			WriteProbeLatency:  metric.NewLatency(metaWriteProbeLatency, opts.HistogramWindowInterval),
			ProbePlanAttempts:  metric.NewCounter(metaProbePlanAttempts),
			ProbePlanFailures:  metric.NewCounter(metaProbePlanFailures),
		},
		tracer: opts.Tracer,
	}
}

func (p *Prober) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(93952)
	return p.metrics
}

func (p *Prober) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(93953)
	ctx = logtags.AddTag(ctx, "kvprober", nil)
	startLoop := func(ctx context.Context, opName string, probe func(context.Context, *kv.DB, planner), pl planner, interval *settings.DurationSetting) error {
		__antithesis_instrumentation__.Notify(93956)
		return stopper.RunAsyncTaskEx(ctx, stop.TaskOpts{TaskName: opName, SpanOpt: stop.SterileRootSpan}, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(93957)
			defer logcrash.RecoverAndReportNonfatalPanic(ctx, &p.settings.SV)

			rnd, _ := randutil.NewPseudoRand()
			d := func() time.Duration {
				__antithesis_instrumentation__.Notify(93959)
				return withJitter(interval.Get(&p.settings.SV), rnd)
			}
			__antithesis_instrumentation__.Notify(93958)
			t := timeutil.NewTimer()
			defer t.Stop()
			t.Reset(d())

			ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
			defer cancel()

			for {
				__antithesis_instrumentation__.Notify(93960)
				select {
				case <-t.C:
					__antithesis_instrumentation__.Notify(93962)
					t.Read = true

					t.Reset(d())
				case <-stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(93963)
					return
				}
				__antithesis_instrumentation__.Notify(93961)

				probeCtx, sp := tracing.EnsureChildSpan(ctx, p.tracer, opName+" - probe")
				probe(probeCtx, p.db, pl)
				sp.Finish()
			}
		})
	}
	__antithesis_instrumentation__.Notify(93954)

	if err := startLoop(ctx, "read probe loop", p.readProbe, p.readPlanner, readInterval); err != nil {
		__antithesis_instrumentation__.Notify(93964)
		return err
	} else {
		__antithesis_instrumentation__.Notify(93965)
	}
	__antithesis_instrumentation__.Notify(93955)
	return startLoop(ctx, "write probe loop", p.writeProbe, p.writePlanner, writeInterval)
}

func (p *Prober) readProbe(ctx context.Context, db *kv.DB, pl planner) {
	__antithesis_instrumentation__.Notify(93966)
	p.readProbeImpl(ctx, &proberOpsImpl{}, &proberTxnImpl{db: p.db}, pl)
}

func (p *Prober) readProbeImpl(ctx context.Context, ops proberOps, txns proberTxn, pl planner) {
	__antithesis_instrumentation__.Notify(93967)
	if !readEnabled.Get(&p.settings.SV) {
		__antithesis_instrumentation__.Notify(93972)
		return
	} else {
		__antithesis_instrumentation__.Notify(93973)
	}
	__antithesis_instrumentation__.Notify(93968)

	p.metrics.ProbePlanAttempts.Inc(1)

	step, err := pl.next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(93974)
		log.Health.Errorf(ctx, "can't make a plan: %v", err)
		p.metrics.ProbePlanFailures.Inc(1)
		return
	} else {
		__antithesis_instrumentation__.Notify(93975)
	}
	__antithesis_instrumentation__.Notify(93969)

	p.metrics.ReadProbeAttempts.Inc(1)

	start := timeutil.Now()

	timeout := readTimeout.Get(&p.settings.SV)
	err = contextutil.RunWithTimeout(ctx, "read probe", timeout, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(93976)

		f := ops.Read(step.Key)
		if bypassAdmissionControl.Get(&p.settings.SV) {
			__antithesis_instrumentation__.Notify(93978)
			return txns.Txn(ctx, f)
		} else {
			__antithesis_instrumentation__.Notify(93979)
		}
		__antithesis_instrumentation__.Notify(93977)
		return txns.TxnRootKV(ctx, f)
	})
	__antithesis_instrumentation__.Notify(93970)
	if err != nil {
		__antithesis_instrumentation__.Notify(93980)

		log.Health.Errorf(ctx, "kv.Get(%s), r=%v failed with: %v", step.Key, step.RangeID, err)
		p.metrics.ReadProbeFailures.Inc(1)
		return
	} else {
		__antithesis_instrumentation__.Notify(93981)
	}
	__antithesis_instrumentation__.Notify(93971)

	d := timeutil.Since(start)
	log.Health.Infof(ctx, "kv.Get(%s), r=%v returned success in %v", step.Key, step.RangeID, d)

	p.metrics.ReadProbeLatency.RecordValue(d.Nanoseconds())
}

func (p *Prober) writeProbe(ctx context.Context, db *kv.DB, pl planner) {
	__antithesis_instrumentation__.Notify(93982)
	p.writeProbeImpl(ctx, &proberOpsImpl{}, &proberTxnImpl{db: p.db}, pl)
}

func (p *Prober) writeProbeImpl(ctx context.Context, ops proberOps, txns proberTxn, pl planner) {
	__antithesis_instrumentation__.Notify(93983)
	if !writeEnabled.Get(&p.settings.SV) {
		__antithesis_instrumentation__.Notify(93988)
		return
	} else {
		__antithesis_instrumentation__.Notify(93989)
	}
	__antithesis_instrumentation__.Notify(93984)

	p.metrics.ProbePlanAttempts.Inc(1)

	step, err := pl.next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(93990)
		log.Health.Errorf(ctx, "can't make a plan: %v", err)
		p.metrics.ProbePlanFailures.Inc(1)
		return
	} else {
		__antithesis_instrumentation__.Notify(93991)
	}
	__antithesis_instrumentation__.Notify(93985)

	p.metrics.WriteProbeAttempts.Inc(1)

	start := timeutil.Now()

	timeout := writeTimeout.Get(&p.settings.SV)
	err = contextutil.RunWithTimeout(ctx, "write probe", timeout, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(93992)
		f := ops.Write(step.Key)
		if bypassAdmissionControl.Get(&p.settings.SV) {
			__antithesis_instrumentation__.Notify(93994)
			return txns.Txn(ctx, f)
		} else {
			__antithesis_instrumentation__.Notify(93995)
		}
		__antithesis_instrumentation__.Notify(93993)
		return txns.TxnRootKV(ctx, f)
	})
	__antithesis_instrumentation__.Notify(93986)
	if err != nil {
		__antithesis_instrumentation__.Notify(93996)
		log.Health.Errorf(ctx, "kv.Txn(Put(%s); Del(-)), r=%v failed with: %v", step.Key, step.RangeID, err)
		p.metrics.WriteProbeFailures.Inc(1)
		return
	} else {
		__antithesis_instrumentation__.Notify(93997)
	}
	__antithesis_instrumentation__.Notify(93987)

	d := timeutil.Since(start)
	log.Health.Infof(ctx, "kv.Txn(Put(%s); Del(-)), r=%v returned success in %v", step.Key, step.RangeID, d)

	p.metrics.WriteProbeLatency.RecordValue(d.Nanoseconds())
}

func withJitter(d time.Duration, rnd *rand.Rand) time.Duration {
	__antithesis_instrumentation__.Notify(93998)
	amplitudeNanos := d.Nanoseconds() / 4
	return d + time.Duration(randutil.RandInt63InRange(rnd, -amplitudeNanos, amplitudeNanos))
}
