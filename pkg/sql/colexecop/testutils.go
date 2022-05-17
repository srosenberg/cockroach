package colexecop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type BatchBuffer struct {
	ZeroInputNode
	buffer []coldata.Batch
}

var _ Operator = &BatchBuffer{}

func NewBatchBuffer() *BatchBuffer {
	__antithesis_instrumentation__.Notify(455115)
	return &BatchBuffer{
		buffer: make([]coldata.Batch, 0, 2),
	}
}

func (b *BatchBuffer) Add(batch coldata.Batch, _ []*types.T) {
	__antithesis_instrumentation__.Notify(455116)
	b.buffer = append(b.buffer, batch)
}

func (b *BatchBuffer) Init(context.Context) { __antithesis_instrumentation__.Notify(455117) }

func (b *BatchBuffer) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455118)
	batch := b.buffer[0]
	b.buffer = b.buffer[1:]
	return batch
}

type RepeatableBatchSource struct {
	ZeroInputNode

	colVecs  []coldata.Vec
	typs     []*types.T
	sel      []int
	batchLen int

	numToCopy int
	output    coldata.Batch

	batchesToReturn int
	batchesReturned int
}

var _ Operator = &RepeatableBatchSource{}

func NewRepeatableBatchSource(
	allocator *colmem.Allocator, batch coldata.Batch, typs []*types.T,
) *RepeatableBatchSource {
	__antithesis_instrumentation__.Notify(455119)
	sel := batch.Selection()
	batchLen := batch.Length()
	numToCopy := batchLen
	if sel != nil {
		__antithesis_instrumentation__.Notify(455121)
		maxIdx := 0
		for _, selIdx := range sel[:batchLen] {
			__antithesis_instrumentation__.Notify(455123)
			if selIdx > maxIdx {
				__antithesis_instrumentation__.Notify(455124)
				maxIdx = selIdx
			} else {
				__antithesis_instrumentation__.Notify(455125)
			}
		}
		__antithesis_instrumentation__.Notify(455122)
		numToCopy = maxIdx + 1
	} else {
		__antithesis_instrumentation__.Notify(455126)
	}
	__antithesis_instrumentation__.Notify(455120)
	output := allocator.NewMemBatchWithFixedCapacity(typs, numToCopy)
	src := &RepeatableBatchSource{
		colVecs:   batch.ColVecs(),
		typs:      typs,
		sel:       sel,
		batchLen:  batchLen,
		numToCopy: numToCopy,
		output:    output,
	}
	return src
}

func (s *RepeatableBatchSource) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455127)
	s.batchesReturned++
	if s.batchesToReturn != 0 && func() bool {
		__antithesis_instrumentation__.Notify(455131)
		return s.batchesReturned > s.batchesToReturn == true
	}() == true {
		__antithesis_instrumentation__.Notify(455132)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(455133)
	}
	__antithesis_instrumentation__.Notify(455128)
	s.output.ResetInternalBatch()
	if s.sel != nil {
		__antithesis_instrumentation__.Notify(455134)
		s.output.SetSelection(true)
		copy(s.output.Selection()[:s.batchLen], s.sel[:s.batchLen])
	} else {
		__antithesis_instrumentation__.Notify(455135)
	}
	__antithesis_instrumentation__.Notify(455129)
	for i, colVec := range s.colVecs {
		__antithesis_instrumentation__.Notify(455136)

		s.output.ColVec(i).Copy(coldata.SliceArgs{
			Src:       colVec,
			SrcEndIdx: s.numToCopy,
		})
	}
	__antithesis_instrumentation__.Notify(455130)
	s.output.SetLength(s.batchLen)
	return s.output
}

func (s *RepeatableBatchSource) Init(context.Context) { __antithesis_instrumentation__.Notify(455137) }

func (s *RepeatableBatchSource) ResetBatchesToReturn(b int) {
	__antithesis_instrumentation__.Notify(455138)
	s.batchesToReturn = b
	s.batchesReturned = 0
}

type CallbackOperator struct {
	ZeroInputNode
	InitCb  func(context.Context)
	NextCb  func() coldata.Batch
	CloseCb func(ctx context.Context) error
}

var _ ClosableOperator = &CallbackOperator{}

func (o *CallbackOperator) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455139)
	if o.InitCb == nil {
		__antithesis_instrumentation__.Notify(455141)
		return
	} else {
		__antithesis_instrumentation__.Notify(455142)
	}
	__antithesis_instrumentation__.Notify(455140)
	o.InitCb(ctx)
}

func (o *CallbackOperator) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455143)
	if o.NextCb == nil {
		__antithesis_instrumentation__.Notify(455145)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(455146)
	}
	__antithesis_instrumentation__.Notify(455144)
	return o.NextCb()
}

func (o *CallbackOperator) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(455147)
	if o.CloseCb == nil {
		__antithesis_instrumentation__.Notify(455149)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(455150)
	}
	__antithesis_instrumentation__.Notify(455148)
	return o.CloseCb(ctx)
}

type TestingSemaphore struct {
	count int
	limit int
}

func NewTestingSemaphore(limit int) *TestingSemaphore {
	__antithesis_instrumentation__.Notify(455151)
	return &TestingSemaphore{limit: limit}
}

func (s *TestingSemaphore) Acquire(_ context.Context, n int) error {
	__antithesis_instrumentation__.Notify(455152)
	if n < 0 {
		__antithesis_instrumentation__.Notify(455155)
		return errors.New("acquiring a negative amount")
	} else {
		__antithesis_instrumentation__.Notify(455156)
	}
	__antithesis_instrumentation__.Notify(455153)
	if s.limit != 0 && func() bool {
		__antithesis_instrumentation__.Notify(455157)
		return s.count+n > s.limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(455158)
		return errors.Errorf("testing semaphore limit exceeded: tried acquiring %d but already have a count of %d from a total limit of %d", n, s.count, s.limit)
	} else {
		__antithesis_instrumentation__.Notify(455159)
	}
	__antithesis_instrumentation__.Notify(455154)
	s.count += n
	return nil
}

func (s *TestingSemaphore) TryAcquire(n int) bool {
	__antithesis_instrumentation__.Notify(455160)
	if s.limit != 0 && func() bool {
		__antithesis_instrumentation__.Notify(455162)
		return s.count+n > s.limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(455163)
		return false
	} else {
		__antithesis_instrumentation__.Notify(455164)
	}
	__antithesis_instrumentation__.Notify(455161)
	s.count += n
	return true
}

func (s *TestingSemaphore) Release(n int) int {
	__antithesis_instrumentation__.Notify(455165)
	if n < 0 {
		__antithesis_instrumentation__.Notify(455168)
		colexecerror.InternalError(errors.AssertionFailedf("releasing a negative amount"))
	} else {
		__antithesis_instrumentation__.Notify(455169)
	}
	__antithesis_instrumentation__.Notify(455166)
	if s.count-n < 0 {
		__antithesis_instrumentation__.Notify(455170)
		colexecerror.InternalError(errors.AssertionFailedf("testing semaphore too many resources released, releasing %d, have %d", n, s.count))
	} else {
		__antithesis_instrumentation__.Notify(455171)
	}
	__antithesis_instrumentation__.Notify(455167)
	pre := s.count
	s.count -= n
	return pre
}

func (s *TestingSemaphore) SetLimit(n int) {
	__antithesis_instrumentation__.Notify(455172)
	s.limit = n
}

func (s *TestingSemaphore) GetLimit() int {
	__antithesis_instrumentation__.Notify(455173)
	return s.limit
}

func (s *TestingSemaphore) GetCount() int {
	__antithesis_instrumentation__.Notify(455174)
	return s.count
}
