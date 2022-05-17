package kvprober

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Step struct {
	RangeID roachpb.RangeID
	Key     roachpb.Key
}

type planner interface {
	next(ctx context.Context) (Step, error)
}

type meta2Planner struct {
	db       *kv.DB
	settings *cluster.Settings

	cursor roachpb.Key

	plan []Step

	lastPlanTime time.Time

	happyInterval func() time.Duration

	now          func() time.Time
	getRateLimit func(interval time.Duration, settings *cluster.Settings) time.Duration
	getNMeta2KVs func(
		ctx context.Context,
		db dbScan,
		n int64,
		cursor roachpb.Key,
		timeout time.Duration) ([]kv.KeyValue, roachpb.Key, error)
	meta2KVsToPlan func(kvs []kv.KeyValue) ([]Step, error)
}

func newMeta2Planner(
	db *kv.DB, settings *cluster.Settings, happyInterval func() time.Duration,
) *meta2Planner {
	__antithesis_instrumentation__.Notify(93999)
	return &meta2Planner{
		db:       db,
		settings: settings,
		cursor:   keys.Meta2Prefix,

		lastPlanTime:   timeutil.Unix(0, 0),
		happyInterval:  happyInterval,
		now:            timeutil.Now,
		getRateLimit:   getRateLimitImpl,
		getNMeta2KVs:   getNMeta2KVsImpl,
		meta2KVsToPlan: meta2KVsToPlanImpl,
	}
}

func (p *meta2Planner) next(ctx context.Context) (Step, error) {
	__antithesis_instrumentation__.Notify(94000)
	if len(p.plan) == 0 {
		__antithesis_instrumentation__.Notify(94002)

		timeSinceLastPlan := p.now().Sub(p.lastPlanTime)
		happyInterval := p.happyInterval()
		if limit := p.getRateLimit(happyInterval, p.settings); timeSinceLastPlan < limit {
			__antithesis_instrumentation__.Notify(94007)
			return Step{}, errors.Newf("planner rate limit hit: "+
				"timSinceLastPlan=%v, happyInterval=%v, limit=%v", timeSinceLastPlan, happyInterval, limit)
		} else {
			__antithesis_instrumentation__.Notify(94008)
		}
		__antithesis_instrumentation__.Notify(94003)
		p.lastPlanTime = p.now()

		timeout := scanMeta2Timeout.Get(&p.settings.SV)
		kvs, cursor, err := p.getNMeta2KVs(
			ctx, p.db, numStepsToPlanAtOnce.Get(&p.settings.SV), p.cursor, timeout)
		if err != nil {
			__antithesis_instrumentation__.Notify(94009)
			return Step{}, errors.Wrapf(err, "failed to get meta2 rows")
		} else {
			__antithesis_instrumentation__.Notify(94010)
		}
		__antithesis_instrumentation__.Notify(94004)
		p.cursor = cursor

		plan, err := p.meta2KVsToPlan(kvs)
		if err != nil {
			__antithesis_instrumentation__.Notify(94011)
			return Step{}, errors.Wrapf(err, "failed to make plan from meta2 rows")
		} else {
			__antithesis_instrumentation__.Notify(94012)
		}
		__antithesis_instrumentation__.Notify(94005)

		rand.Shuffle(len(plan), func(i, j int) {
			__antithesis_instrumentation__.Notify(94013)
			plan[i], plan[j] = plan[j], plan[i]
		})
		__antithesis_instrumentation__.Notify(94006)

		p.plan = plan
	} else {
		__antithesis_instrumentation__.Notify(94014)
	}
	__antithesis_instrumentation__.Notify(94001)

	step := p.plan[0]
	p.plan = p.plan[1:]
	return step, nil
}

func getRateLimitImpl(interval time.Duration, settings *cluster.Settings) time.Duration {
	__antithesis_instrumentation__.Notify(94015)
	sv := &settings.SV
	const happyPathIntervalToRateLimitIntervalRatio = 2
	return time.Duration(
		(interval.Nanoseconds())*
			numStepsToPlanAtOnce.Get(sv)/happyPathIntervalToRateLimitIntervalRatio) * time.Nanosecond
}

type dbScan interface {
	Scan(ctx context.Context, begin, end interface{}, maxRows int64) ([]kv.KeyValue, error)
}

func getNMeta2KVsImpl(
	ctx context.Context, db dbScan, n int64, cursor roachpb.Key, timeout time.Duration,
) ([]kv.KeyValue, roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(94016)
	var kvs []kv.KeyValue

	for n > 0 {
		__antithesis_instrumentation__.Notify(94018)
		var newkvs []kv.KeyValue
		if err := contextutil.RunWithTimeout(ctx, "db.Scan", timeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(94021)

			var err error

			newkvs, err = db.Scan(ctx, cursor, keys.Meta2KeyMax.Next(), n)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(94022)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(94023)
		}
		__antithesis_instrumentation__.Notify(94019)

		if len(newkvs) == 0 {
			__antithesis_instrumentation__.Notify(94024)
			return nil, nil, errors.New("scanning meta2 returned no KV pairs")
		} else {
			__antithesis_instrumentation__.Notify(94025)
		}
		__antithesis_instrumentation__.Notify(94020)

		n = n - int64(len(newkvs))
		kvs = append(kvs, newkvs...)
		cursor = kvs[len(kvs)-1].Key.Next()

		if cursor.Equal(keys.Meta2KeyMax.Next()) {
			__antithesis_instrumentation__.Notify(94026)
			cursor = keys.Meta2Prefix
		} else {
			__antithesis_instrumentation__.Notify(94027)
		}
	}
	__antithesis_instrumentation__.Notify(94017)

	return kvs, cursor, nil
}

func meta2KVsToPlanImpl(kvs []kv.KeyValue) ([]Step, error) {
	__antithesis_instrumentation__.Notify(94028)
	plans := make([]Step, len(kvs))

	var rangeDesc roachpb.RangeDescriptor
	for i, kv := range kvs {
		__antithesis_instrumentation__.Notify(94030)
		if err := kv.ValueProto(&rangeDesc); err != nil {
			__antithesis_instrumentation__.Notify(94032)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(94033)
		}
		__antithesis_instrumentation__.Notify(94031)
		plans[i] = Step{
			RangeID: rangeDesc.RangeID,
		}

		if rangeDesc.RangeID == 1 {
			__antithesis_instrumentation__.Notify(94034)
			plans[i].Key = keys.RangeProbeKey(keys.MustAddr(keys.LocalMax))
		} else {
			__antithesis_instrumentation__.Notify(94035)
			plans[i].Key = keys.RangeProbeKey(rangeDesc.StartKey)
		}
	}
	__antithesis_instrumentation__.Notify(94029)

	return plans, nil
}
