package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

var (
	sizeOfTimeSeriesData = int64(unsafe.Sizeof(roachpb.InternalTimeSeriesData{}))
	sizeOfSample         = int64(unsafe.Sizeof(roachpb.InternalTimeSeriesSample{}))
	sizeOfDataPoint      = int64(unsafe.Sizeof(tspb.TimeSeriesDatapoint{}))
	sizeOfInt32          = int64(unsafe.Sizeof(int32(0)))
	sizeOfUint32         = int64(unsafe.Sizeof(uint32(0)))
	sizeOfFloat64        = int64(unsafe.Sizeof(float64(0)))
	sizeOfTimestamp      = int64(unsafe.Sizeof(hlc.Timestamp{}))
)

type QueryMemoryOptions struct {
	BudgetBytes int64

	EstimatedSources int64

	InterpolationLimitNanos int64

	Columnar bool
}

type QueryMemoryContext struct {
	workerMonitor *mon.BytesMonitor
	resultAccount *mon.BoundAccount
	QueryMemoryOptions
}

func MakeQueryMemoryContext(
	workerMonitor, resultMonitor *mon.BytesMonitor, opts QueryMemoryOptions,
) QueryMemoryContext {
	__antithesis_instrumentation__.Notify(648064)
	resultAccount := resultMonitor.MakeBoundAccount()
	return QueryMemoryContext{
		workerMonitor:      workerMonitor,
		resultAccount:      &resultAccount,
		QueryMemoryOptions: opts,
	}
}

func (qmc QueryMemoryContext) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(648065)
	if qmc.resultAccount != nil {
		__antithesis_instrumentation__.Notify(648066)
		qmc.resultAccount.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(648067)
	}
}

func overflowSafeMultiply64(a, b int64) (int64, bool) {
	__antithesis_instrumentation__.Notify(648068)
	if a == 0 || func() bool {
		__antithesis_instrumentation__.Notify(648071)
		return b == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(648072)
		return 0, true
	} else {
		__antithesis_instrumentation__.Notify(648073)
	}
	__antithesis_instrumentation__.Notify(648069)
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		__antithesis_instrumentation__.Notify(648074)
		if c/b == a {
			__antithesis_instrumentation__.Notify(648075)
			return c, true
		} else {
			__antithesis_instrumentation__.Notify(648076)
		}
	} else {
		__antithesis_instrumentation__.Notify(648077)
	}
	__antithesis_instrumentation__.Notify(648070)
	return c, false
}

func (qmc QueryMemoryContext) GetMaxTimespan(r Resolution) (int64, error) {
	__antithesis_instrumentation__.Notify(648078)
	slabDuration := r.SlabDuration()

	sizeOfSlab := qmc.computeSizeOfSlab(r)

	interpolationBufferOneSide :=
		int64(math.Ceil(float64(qmc.InterpolationLimitNanos) / float64(slabDuration)))

	interpolationBuffer := interpolationBufferOneSide * 2

	if (interpolationBufferOneSide*slabDuration)-qmc.InterpolationLimitNanos < slabDuration/2 {
		__antithesis_instrumentation__.Notify(648082)
		interpolationBuffer++
	} else {
		__antithesis_instrumentation__.Notify(648083)
	}
	__antithesis_instrumentation__.Notify(648079)

	perSourceMem := qmc.BudgetBytes / qmc.EstimatedSources
	numSlabs := perSourceMem/sizeOfSlab - interpolationBuffer
	if numSlabs <= 0 {
		__antithesis_instrumentation__.Notify(648084)
		return 0, fmt.Errorf("insufficient memory budget to attempt query")
	} else {
		__antithesis_instrumentation__.Notify(648085)
	}
	__antithesis_instrumentation__.Notify(648080)

	maxDuration, valid := overflowSafeMultiply64(numSlabs, slabDuration)
	if valid {
		__antithesis_instrumentation__.Notify(648086)
		return maxDuration, nil
	} else {
		__antithesis_instrumentation__.Notify(648087)
	}
	__antithesis_instrumentation__.Notify(648081)
	return math.MaxInt64, nil
}

func (qmc QueryMemoryContext) GetMaxRollupSlabs(r Resolution) int64 {
	__antithesis_instrumentation__.Notify(648088)

	return qmc.BudgetBytes / qmc.computeSizeOfSlab(r)
}

func (qmc QueryMemoryContext) computeSizeOfSlab(r Resolution) int64 {
	__antithesis_instrumentation__.Notify(648089)
	slabDuration := r.SlabDuration()

	var sizeOfSlab int64
	if qmc.Columnar {
		__antithesis_instrumentation__.Notify(648091)

		sizeOfColumns := (sizeOfInt32 + sizeOfFloat64)
		if r.IsRollup() {
			__antithesis_instrumentation__.Notify(648093)

			sizeOfColumns += 5*sizeOfFloat64 + sizeOfUint32
		} else {
			__antithesis_instrumentation__.Notify(648094)
		}
		__antithesis_instrumentation__.Notify(648092)
		sizeOfSlab = sizeOfTimeSeriesData + (slabDuration/r.SampleDuration())*sizeOfColumns
	} else {
		__antithesis_instrumentation__.Notify(648095)

		sizeOfSlab = sizeOfTimeSeriesData + (slabDuration/r.SampleDuration())*sizeOfSample
	}
	__antithesis_instrumentation__.Notify(648090)
	return sizeOfSlab
}
