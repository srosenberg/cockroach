package colexecargs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/marusama/semaphore"
	"github.com/stretchr/testify/require"
)

var TestNewColOperator func(ctx context.Context, flowCtx *execinfra.FlowCtx, args *NewColOperatorArgs,
) (r *NewColOperatorResult, err error)

type OpWithMetaInfo struct {
	Root colexecop.Operator

	StatsCollectors []colexecop.VectorizedStatsCollector

	MetadataSources colexecop.MetadataSources

	ToClose colexecop.Closers
}

type NewColOperatorArgs struct {
	Spec                 *execinfrapb.ProcessorSpec
	Inputs               []OpWithMetaInfo
	StreamingMemAccount  *mon.BoundAccount
	ProcessorConstructor execinfra.ProcessorConstructor
	LocalProcessors      []execinfra.LocalProcessor
	DiskQueueCfg         colcontainer.DiskQueueCfg
	FDSemaphore          semaphore.Semaphore
	ExprHelper           *ExprHelper
	Factory              coldata.ColumnFactory
	MonitorRegistry      *MonitorRegistry
	TestingKnobs         struct {
		SpillingCallbackFn func()

		NumForcedRepartitions int

		DiskSpillingDisabled bool

		DelegateFDAcquisitions bool
	}
}

type NewColOperatorResult struct {
	OpWithMetaInfo
	KVReader colexecop.KVReader

	Columnarizer colexecop.VectorizedStatsCollector
	ColumnTypes  []*types.T
	Releasables  []execinfra.Releasable
}

var _ execinfra.Releasable = &NewColOperatorResult{}

func (r *NewColOperatorResult) TestCleanupNoError(t testing.TB) {
	__antithesis_instrumentation__.Notify(286170)
	require.NoError(t, r.ToClose.Close(context.Background()))
}

var newColOperatorResultPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(286171)
		return &NewColOperatorResult{}
	},
}

func GetNewColOperatorResult() *NewColOperatorResult {
	__antithesis_instrumentation__.Notify(286172)
	return newColOperatorResultPool.Get().(*NewColOperatorResult)
}

func (r *NewColOperatorResult) Release() {
	__antithesis_instrumentation__.Notify(286173)
	for _, releasable := range r.Releasables {
		__antithesis_instrumentation__.Notify(286179)
		releasable.Release()
	}
	__antithesis_instrumentation__.Notify(286174)

	for i := range r.StatsCollectors {
		__antithesis_instrumentation__.Notify(286180)
		r.StatsCollectors[i] = nil
	}
	__antithesis_instrumentation__.Notify(286175)
	for i := range r.MetadataSources {
		__antithesis_instrumentation__.Notify(286181)
		r.MetadataSources[i] = nil
	}
	__antithesis_instrumentation__.Notify(286176)
	for i := range r.ToClose {
		__antithesis_instrumentation__.Notify(286182)
		r.ToClose[i] = nil
	}
	__antithesis_instrumentation__.Notify(286177)
	for i := range r.Releasables {
		__antithesis_instrumentation__.Notify(286183)
		r.Releasables[i] = nil
	}
	__antithesis_instrumentation__.Notify(286178)
	*r = NewColOperatorResult{
		OpWithMetaInfo: OpWithMetaInfo{
			StatsCollectors: r.StatsCollectors[:0],
			MetadataSources: r.MetadataSources[:0],
			ToClose:         r.ToClose[:0],
		},

		ColumnTypes: r.ColumnTypes[:0],
		Releasables: r.Releasables[:0],
	}
	newColOperatorResultPool.Put(r)
}
