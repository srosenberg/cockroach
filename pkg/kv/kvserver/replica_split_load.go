package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var SplitByLoadEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.range_split.by_load_enabled",
	"allow automatic splits of ranges based on where load is concentrated",
	true,
).WithPublic()

var SplitByLoadQPSThreshold = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.range_split.load_qps_threshold",
	"the QPS over which, the range becomes a candidate for load based splitting",
	2500,
).WithPublic()

func (r *Replica) SplitByLoadQPSThreshold() float64 {
	__antithesis_instrumentation__.Notify(120705)
	return float64(SplitByLoadQPSThreshold.Get(&r.store.cfg.Settings.SV))
}

func (r *Replica) SplitByLoadEnabled() bool {
	__antithesis_instrumentation__.Notify(120706)
	return SplitByLoadEnabled.Get(&r.store.cfg.Settings.SV) && func() bool {
		__antithesis_instrumentation__.Notify(120707)
		return !r.store.TestingKnobs().DisableLoadBasedSplitting == true
	}() == true
}

func (r *Replica) recordBatchForLoadBasedSplitting(
	ctx context.Context, ba *roachpb.BatchRequest, spans *spanset.SpanSet,
) {
	__antithesis_instrumentation__.Notify(120708)
	if !r.SplitByLoadEnabled() {
		__antithesis_instrumentation__.Notify(120711)
		return
	} else {
		__antithesis_instrumentation__.Notify(120712)
	}
	__antithesis_instrumentation__.Notify(120709)
	shouldInitSplit := r.loadBasedSplitter.Record(timeutil.Now(), len(ba.Requests), func() roachpb.Span {
		__antithesis_instrumentation__.Notify(120713)
		return spans.BoundarySpan(spanset.SpanGlobal)
	})
	__antithesis_instrumentation__.Notify(120710)
	if shouldInitSplit {
		__antithesis_instrumentation__.Notify(120714)
		r.store.splitQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(120715)
	}
}
