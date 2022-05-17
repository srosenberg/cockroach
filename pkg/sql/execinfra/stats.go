package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	pbtypes "github.com/gogo/protobuf/types"
)

func ShouldCollectStats(ctx context.Context, flowCtx *FlowCtx) bool {
	__antithesis_instrumentation__.Notify(471459)
	return tracing.SpanFromContext(ctx) != nil && func() bool {
		__antithesis_instrumentation__.Notify(471460)
		return flowCtx.CollectStats == true
	}() == true
}

func GetCumulativeContentionTime(ctx context.Context) time.Duration {
	__antithesis_instrumentation__.Notify(471461)
	var cumulativeContentionTime time.Duration
	recording := GetTraceData(ctx)
	if recording == nil {
		__antithesis_instrumentation__.Notify(471464)
		return cumulativeContentionTime
	} else {
		__antithesis_instrumentation__.Notify(471465)
	}
	__antithesis_instrumentation__.Notify(471462)
	var ev roachpb.ContentionEvent
	for i := range recording {
		__antithesis_instrumentation__.Notify(471466)
		recording[i].Structured(func(any *pbtypes.Any, _ time.Time) {
			__antithesis_instrumentation__.Notify(471467)
			if !pbtypes.Is(any, &ev) {
				__antithesis_instrumentation__.Notify(471470)
				return
			} else {
				__antithesis_instrumentation__.Notify(471471)
			}
			__antithesis_instrumentation__.Notify(471468)
			if err := pbtypes.UnmarshalAny(any, &ev); err != nil {
				__antithesis_instrumentation__.Notify(471472)
				return
			} else {
				__antithesis_instrumentation__.Notify(471473)
			}
			__antithesis_instrumentation__.Notify(471469)
			cumulativeContentionTime += ev.Duration
		})
	}
	__antithesis_instrumentation__.Notify(471463)
	return cumulativeContentionTime
}

type ScanStats struct {
	NumInterfaceSteps uint64

	NumInternalSteps uint64

	NumInterfaceSeeks uint64

	NumInternalSeeks uint64
}

func PopulateKVMVCCStats(kvStats *execinfrapb.KVStats, ss *ScanStats) {
	__antithesis_instrumentation__.Notify(471474)
	kvStats.NumInterfaceSteps = optional.MakeUint(ss.NumInterfaceSteps)
	kvStats.NumInternalSteps = optional.MakeUint(ss.NumInternalSteps)
	kvStats.NumInterfaceSeeks = optional.MakeUint(ss.NumInterfaceSeeks)
	kvStats.NumInternalSeeks = optional.MakeUint(ss.NumInternalSeeks)
}

func GetScanStats(ctx context.Context) (ss ScanStats) {
	__antithesis_instrumentation__.Notify(471475)
	recording := GetTraceData(ctx)
	if recording == nil {
		__antithesis_instrumentation__.Notify(471478)
		return ScanStats{}
	} else {
		__antithesis_instrumentation__.Notify(471479)
	}
	__antithesis_instrumentation__.Notify(471476)
	var ev roachpb.ScanStats
	for i := range recording {
		__antithesis_instrumentation__.Notify(471480)
		recording[i].Structured(func(any *pbtypes.Any, _ time.Time) {
			__antithesis_instrumentation__.Notify(471481)
			if !pbtypes.Is(any, &ev) {
				__antithesis_instrumentation__.Notify(471484)
				return
			} else {
				__antithesis_instrumentation__.Notify(471485)
			}
			__antithesis_instrumentation__.Notify(471482)
			if err := pbtypes.UnmarshalAny(any, &ev); err != nil {
				__antithesis_instrumentation__.Notify(471486)
				return
			} else {
				__antithesis_instrumentation__.Notify(471487)
			}
			__antithesis_instrumentation__.Notify(471483)

			ss.NumInterfaceSteps += ev.NumInterfaceSteps
			ss.NumInternalSteps += ev.NumInternalSteps
			ss.NumInterfaceSeeks += ev.NumInterfaceSeeks
			ss.NumInternalSeeks += ev.NumInternalSeeks
		})
	}
	__antithesis_instrumentation__.Notify(471477)
	return ss
}
