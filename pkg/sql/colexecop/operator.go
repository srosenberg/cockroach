package colexecop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type Operator interface {
	Init(ctx context.Context)

	Next() coldata.Batch

	execinfra.OpNode
}

type DrainableOperator interface {
	Operator
	MetadataSource
}

type KVReader interface {
	GetBytesRead() int64

	GetRowsRead() int64

	GetCumulativeContentionTime() time.Duration

	GetScanStats() execinfra.ScanStats
}

type ZeroInputNode struct{}

func (ZeroInputNode) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(455044)
	return 0
}

func (ZeroInputNode) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(455045)
	colexecerror.InternalError(errors.AssertionFailedf("invalid index %d", nth))

	return nil
}

func NewOneInputNode(input Operator) OneInputNode {
	__antithesis_instrumentation__.Notify(455046)
	return OneInputNode{Input: input}
}

type OneInputNode struct {
	Input Operator
}

func (OneInputNode) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(455047)
	return 1
}

func (n OneInputNode) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(455048)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(455050)
		return n.Input
	} else {
		__antithesis_instrumentation__.Notify(455051)
	}
	__antithesis_instrumentation__.Notify(455049)
	colexecerror.InternalError(errors.AssertionFailedf("invalid index %d", nth))

	return nil
}

type BufferingInMemoryOperator interface {
	Operator

	ExportBuffered(input Operator) coldata.Batch
}

type Closer interface {
	Close(context.Context) error
}

type Closers []Closer

func (c Closers) CloseAndLogOnErr(ctx context.Context, prefix string) {
	__antithesis_instrumentation__.Notify(455052)
	if err := colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(455053)
		for _, closer := range c {
			__antithesis_instrumentation__.Notify(455054)
			if err := closer.Close(ctx); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(455055)
				return log.V(1) == true
			}() == true {
				__antithesis_instrumentation__.Notify(455056)
				log.Infof(ctx, "%s: error closing Closer: %v", prefix, err)
			} else {
				__antithesis_instrumentation__.Notify(455057)
			}
		}
	}); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(455058)
		return log.V(1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(455059)
		log.Infof(ctx, "%s: runtime error closing the closers: %v", prefix, err)
	} else {
		__antithesis_instrumentation__.Notify(455060)
	}
}

func (c Closers) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(455061)
	var lastErr error
	for _, closer := range c {
		__antithesis_instrumentation__.Notify(455063)
		if err := closer.Close(ctx); err != nil {
			__antithesis_instrumentation__.Notify(455064)
			lastErr = err
		} else {
			__antithesis_instrumentation__.Notify(455065)
		}
	}
	__antithesis_instrumentation__.Notify(455062)
	return lastErr
}

type Resetter interface {
	Reset(ctx context.Context)
}

type ResettableOperator interface {
	Operator
	Resetter
}

type FeedOperator struct {
	ZeroInputNode
	NonExplainable
	batch coldata.Batch
}

func NewFeedOperator() *FeedOperator {
	__antithesis_instrumentation__.Notify(455066)
	return &FeedOperator{}
}

func (FeedOperator) Init(context.Context) { __antithesis_instrumentation__.Notify(455067) }

func (o *FeedOperator) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455068)
	return o.batch
}

func (o *FeedOperator) SetBatch(batch coldata.Batch) {
	__antithesis_instrumentation__.Notify(455069)
	o.batch = batch
}

var _ Operator = &FeedOperator{}

type NonExplainable interface {
	nonExplainableMarker()
}

type InitHelper struct {
	Ctx context.Context
}

func (h *InitHelper) Init(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(455070)
	if h.Ctx != nil {
		__antithesis_instrumentation__.Notify(455073)
		return false
	} else {
		__antithesis_instrumentation__.Notify(455074)
	}
	__antithesis_instrumentation__.Notify(455071)
	if ctx == nil {
		__antithesis_instrumentation__.Notify(455075)
		colexecerror.InternalError(errors.AssertionFailedf("nil context is passed"))
	} else {
		__antithesis_instrumentation__.Notify(455076)
	}
	__antithesis_instrumentation__.Notify(455072)
	h.Ctx = ctx
	return true
}

func (h *InitHelper) EnsureCtx() context.Context {
	__antithesis_instrumentation__.Notify(455077)
	if h.Ctx == nil {
		__antithesis_instrumentation__.Notify(455079)
		return context.Background()
	} else {
		__antithesis_instrumentation__.Notify(455080)
	}
	__antithesis_instrumentation__.Notify(455078)
	return h.Ctx
}

func MakeOneInputHelper(input Operator) OneInputHelper {
	__antithesis_instrumentation__.Notify(455081)
	return OneInputHelper{
		OneInputNode: NewOneInputNode(input),
	}
}

type OneInputHelper struct {
	OneInputNode
	InitHelper
}

func (h *OneInputHelper) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455082)
	if !h.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(455084)
		return
	} else {
		__antithesis_instrumentation__.Notify(455085)
	}
	__antithesis_instrumentation__.Notify(455083)
	h.Input.Init(h.Ctx)
}

type CloserHelper struct {
	closed bool
}

func (c *CloserHelper) Close() bool {
	__antithesis_instrumentation__.Notify(455086)
	if c.closed {
		__antithesis_instrumentation__.Notify(455088)
		return false
	} else {
		__antithesis_instrumentation__.Notify(455089)
	}
	__antithesis_instrumentation__.Notify(455087)
	c.closed = true
	return true
}

func (c *CloserHelper) Reset() {
	__antithesis_instrumentation__.Notify(455090)
	c.closed = false
}

type ClosableOperator interface {
	Operator
	Closer
}

func MakeOneInputCloserHelper(input Operator) OneInputCloserHelper {
	__antithesis_instrumentation__.Notify(455091)
	return OneInputCloserHelper{
		OneInputNode: NewOneInputNode(input),
	}
}

type OneInputCloserHelper struct {
	OneInputNode
	CloserHelper
}

var _ Closer = &OneInputCloserHelper{}

func (c *OneInputCloserHelper) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(455092)
	if !c.CloserHelper.Close() {
		__antithesis_instrumentation__.Notify(455095)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(455096)
	}
	__antithesis_instrumentation__.Notify(455093)
	if closer, ok := c.Input.(Closer); ok {
		__antithesis_instrumentation__.Notify(455097)
		return closer.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(455098)
	}
	__antithesis_instrumentation__.Notify(455094)
	return nil
}

func MakeOneInputInitCloserHelper(input Operator) OneInputInitCloserHelper {
	__antithesis_instrumentation__.Notify(455099)
	return OneInputInitCloserHelper{
		OneInputCloserHelper: MakeOneInputCloserHelper(input),
	}
}

type OneInputInitCloserHelper struct {
	InitHelper
	OneInputCloserHelper
}

func (h *OneInputInitCloserHelper) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455100)
	if !h.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(455102)
		return
	} else {
		__antithesis_instrumentation__.Notify(455103)
	}
	__antithesis_instrumentation__.Notify(455101)
	h.Input.Init(h.Ctx)
}

type noopOperator struct {
	OneInputInitCloserHelper
	NonExplainable
}

var _ ResettableOperator = &noopOperator{}

func NewNoop(input Operator) ResettableOperator {
	__antithesis_instrumentation__.Notify(455104)
	return &noopOperator{OneInputInitCloserHelper: MakeOneInputInitCloserHelper(input)}
}

func (n *noopOperator) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455105)
	return n.Input.Next()
}

func (n *noopOperator) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455106)
	if r, ok := n.Input.(Resetter); ok {
		__antithesis_instrumentation__.Notify(455107)
		r.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(455108)
	}
}

type MetadataSource interface {
	DrainMeta() []execinfrapb.ProducerMetadata
}

type MetadataSources []MetadataSource

func (s MetadataSources) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(455109)
	var result []execinfrapb.ProducerMetadata
	if err := colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(455111)
		for _, src := range s {
			__antithesis_instrumentation__.Notify(455112)
			result = append(result, src.DrainMeta()...)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(455113)
		meta := execinfrapb.GetProducerMeta()
		meta.Err = err
		result = append(result, *meta)
	} else {
		__antithesis_instrumentation__.Notify(455114)
	}
	__antithesis_instrumentation__.Notify(455110)
	return result
}

type VectorizedStatsCollector interface {
	Operator

	GetStats() *execinfrapb.ComponentStats
}
