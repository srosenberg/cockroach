package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const RowChannelBufSize = 16

type ConsumerStatus uint32

const (
	NeedMoreRows ConsumerStatus = iota

	DrainRequested

	ConsumerClosed
)

type receiverBase interface {
	ProducerDone()
}

type RowReceiver interface {
	receiverBase

	Push(row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata) ConsumerStatus
}

type BatchReceiver interface {
	receiverBase

	PushBatch(batch coldata.Batch, meta *execinfrapb.ProducerMetadata) ConsumerStatus
}

type RowSource interface {
	OutputTypes() []*types.T

	Start(context.Context)

	Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)

	ConsumerDone()

	ConsumerClosed()
}

type RowSourcedProcessor interface {
	RowSource
	Processor
}

func Run(ctx context.Context, src RowSource, dst RowReceiver) {
	__antithesis_instrumentation__.Notify(470896)
	for {
		__antithesis_instrumentation__.Notify(470897)
		row, meta := src.Next()

		if row != nil || func() bool {
			__antithesis_instrumentation__.Notify(470899)
			return meta != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(470900)
			switch dst.Push(row, meta) {
			case NeedMoreRows:
				__antithesis_instrumentation__.Notify(470901)
				continue
			case DrainRequested:
				__antithesis_instrumentation__.Notify(470902)
				DrainAndForwardMetadata(ctx, src, dst)
				dst.ProducerDone()
				return
			case ConsumerClosed:
				__antithesis_instrumentation__.Notify(470903)
				src.ConsumerClosed()
				dst.ProducerDone()
				return
			default:
				__antithesis_instrumentation__.Notify(470904)
			}
		} else {
			__antithesis_instrumentation__.Notify(470905)
		}
		__antithesis_instrumentation__.Notify(470898)

		dst.ProducerDone()
		return
	}
}

type Releasable interface {
	Release()
}

func DrainAndForwardMetadata(ctx context.Context, src RowSource, dst RowReceiver) {
	__antithesis_instrumentation__.Notify(470906)
	src.ConsumerDone()
	for {
		__antithesis_instrumentation__.Notify(470907)
		row, meta := src.Next()
		if meta == nil {
			__antithesis_instrumentation__.Notify(470910)
			if row == nil {
				__antithesis_instrumentation__.Notify(470912)
				return
			} else {
				__antithesis_instrumentation__.Notify(470913)
			}
			__antithesis_instrumentation__.Notify(470911)
			continue
		} else {
			__antithesis_instrumentation__.Notify(470914)
		}
		__antithesis_instrumentation__.Notify(470908)
		if row != nil {
			__antithesis_instrumentation__.Notify(470915)
			log.Fatalf(
				ctx, "both row data and metadata in the same record. row: %s meta: %+v",
				row.String(src.OutputTypes()), meta,
			)
		} else {
			__antithesis_instrumentation__.Notify(470916)
		}
		__antithesis_instrumentation__.Notify(470909)

		switch dst.Push(nil, meta) {
		case ConsumerClosed:
			__antithesis_instrumentation__.Notify(470917)
			src.ConsumerClosed()
			return
		case NeedMoreRows:
			__antithesis_instrumentation__.Notify(470918)
		case DrainRequested:
			__antithesis_instrumentation__.Notify(470919)
		default:
			__antithesis_instrumentation__.Notify(470920)
		}
	}
}

func GetTraceData(ctx context.Context) []tracingpb.RecordedSpan {
	__antithesis_instrumentation__.Notify(470921)
	return tracing.SpanFromContext(ctx).GetConfiguredRecording()
}

func GetTraceDataAsMetadata(span *tracing.Span) *execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(470922)
	if trace := span.GetConfiguredRecording(); len(trace) > 0 {
		__antithesis_instrumentation__.Notify(470924)
		meta := execinfrapb.GetProducerMeta()
		meta.TraceData = trace
		return meta
	} else {
		__antithesis_instrumentation__.Notify(470925)
	}
	__antithesis_instrumentation__.Notify(470923)
	return nil
}

func SendTraceData(ctx context.Context, dst RowReceiver) {
	__antithesis_instrumentation__.Notify(470926)
	if rec := GetTraceData(ctx); rec != nil {
		__antithesis_instrumentation__.Notify(470927)
		dst.Push(nil, &execinfrapb.ProducerMetadata{TraceData: rec})
	} else {
		__antithesis_instrumentation__.Notify(470928)
	}
}

func GetLeafTxnFinalState(ctx context.Context, txn *kv.Txn) *roachpb.LeafTxnFinalState {
	__antithesis_instrumentation__.Notify(470929)
	if txn.Type() != kv.LeafTxn {
		__antithesis_instrumentation__.Notify(470933)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(470934)
	}
	__antithesis_instrumentation__.Notify(470930)
	txnMeta, err := txn.GetLeafTxnFinalState(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(470935)

		panic(errors.Wrap(err, "in execinfra.GetLeafTxnFinalState"))
	} else {
		__antithesis_instrumentation__.Notify(470936)
	}
	__antithesis_instrumentation__.Notify(470931)

	if txnMeta.Txn.ID == uuid.Nil {
		__antithesis_instrumentation__.Notify(470937)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(470938)
	}
	__antithesis_instrumentation__.Notify(470932)
	return txnMeta
}

func DrainAndClose(
	ctx context.Context,
	dst RowReceiver,
	cause error,
	pushTrailingMeta func(context.Context),
	srcs ...RowSource,
) {
	__antithesis_instrumentation__.Notify(470939)
	if cause != nil {
		__antithesis_instrumentation__.Notify(470942)

		_ = dst.Push(nil, &execinfrapb.ProducerMetadata{Err: cause})
	} else {
		__antithesis_instrumentation__.Notify(470943)
	}
	__antithesis_instrumentation__.Notify(470940)
	if len(srcs) > 0 {
		__antithesis_instrumentation__.Notify(470944)
		var wg sync.WaitGroup
		for _, input := range srcs[1:] {
			__antithesis_instrumentation__.Notify(470946)
			wg.Add(1)
			go func(input RowSource) {
				__antithesis_instrumentation__.Notify(470947)
				DrainAndForwardMetadata(ctx, input, dst)
				wg.Done()
			}(input)
		}
		__antithesis_instrumentation__.Notify(470945)
		DrainAndForwardMetadata(ctx, srcs[0], dst)
		wg.Wait()
	} else {
		__antithesis_instrumentation__.Notify(470948)
	}
	__antithesis_instrumentation__.Notify(470941)
	pushTrailingMeta(ctx)
	dst.ProducerDone()
}

type NoMetadataRowSource struct {
	src          RowSource
	metadataSink RowReceiver
}

func MakeNoMetadataRowSource(src RowSource, sink RowReceiver) NoMetadataRowSource {
	__antithesis_instrumentation__.Notify(470949)
	return NoMetadataRowSource{src: src, metadataSink: sink}
}

func (rs *NoMetadataRowSource) NextRow() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(470950)
	for {
		__antithesis_instrumentation__.Notify(470951)
		row, meta := rs.src.Next()
		if meta == nil {
			__antithesis_instrumentation__.Notify(470954)
			return row, nil
		} else {
			__antithesis_instrumentation__.Notify(470955)
		}
		__antithesis_instrumentation__.Notify(470952)
		if meta.Err != nil {
			__antithesis_instrumentation__.Notify(470956)
			return nil, meta.Err
		} else {
			__antithesis_instrumentation__.Notify(470957)
		}
		__antithesis_instrumentation__.Notify(470953)

		_ = rs.metadataSink.Push(nil, meta)
	}
}

type RowChannelMsg struct {
	Row  rowenc.EncDatumRow
	Meta *execinfrapb.ProducerMetadata
}

type rowSourceBase struct {
	ConsumerStatus ConsumerStatus
}

func (rb *rowSourceBase) consumerDone() {
	__antithesis_instrumentation__.Notify(470958)
	atomic.CompareAndSwapUint32((*uint32)(&rb.ConsumerStatus),
		uint32(NeedMoreRows), uint32(DrainRequested))
}

func (rb *rowSourceBase) consumerClosed(name string) {
	__antithesis_instrumentation__.Notify(470959)
	status := ConsumerStatus(atomic.LoadUint32((*uint32)(&rb.ConsumerStatus)))
	if status == ConsumerClosed {
		__antithesis_instrumentation__.Notify(470961)
		logcrash.ReportOrPanic(context.Background(), nil, "%s already closed", redact.Safe(name))
	} else {
		__antithesis_instrumentation__.Notify(470962)
	}
	__antithesis_instrumentation__.Notify(470960)
	atomic.StoreUint32((*uint32)(&rb.ConsumerStatus), uint32(ConsumerClosed))
}

type RowChannel struct {
	types []*types.T

	C <-chan RowChannelMsg

	dataChan chan RowChannelMsg

	rowSourceBase

	numSenders int32
}

var _ RowReceiver = &RowChannel{}
var _ RowSource = &RowChannel{}

func (rc *RowChannel) InitWithNumSenders(types []*types.T, numSenders int) {
	__antithesis_instrumentation__.Notify(470963)
	rc.InitWithBufSizeAndNumSenders(types, RowChannelBufSize, numSenders)
}

func (rc *RowChannel) InitWithBufSizeAndNumSenders(types []*types.T, chanBufSize, numSenders int) {
	__antithesis_instrumentation__.Notify(470964)
	rc.types = types
	rc.dataChan = make(chan RowChannelMsg, chanBufSize)
	rc.C = rc.dataChan
	atomic.StoreInt32(&rc.numSenders, int32(numSenders))
}

func (rc *RowChannel) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) ConsumerStatus {
	__antithesis_instrumentation__.Notify(470965)
	consumerStatus := ConsumerStatus(
		atomic.LoadUint32((*uint32)(&rc.ConsumerStatus)))
	switch consumerStatus {
	case NeedMoreRows:
		__antithesis_instrumentation__.Notify(470967)
		rc.dataChan <- RowChannelMsg{Row: row, Meta: meta}
	case DrainRequested:
		__antithesis_instrumentation__.Notify(470968)

		if meta != nil {
			__antithesis_instrumentation__.Notify(470971)
			rc.dataChan <- RowChannelMsg{Meta: meta}
		} else {
			__antithesis_instrumentation__.Notify(470972)
		}
	case ConsumerClosed:
		__antithesis_instrumentation__.Notify(470969)
	default:
		__antithesis_instrumentation__.Notify(470970)

	}
	__antithesis_instrumentation__.Notify(470966)
	return consumerStatus
}

func (rc *RowChannel) ProducerDone() {
	__antithesis_instrumentation__.Notify(470973)
	newVal := atomic.AddInt32(&rc.numSenders, -1)
	if newVal < 0 {
		__antithesis_instrumentation__.Notify(470975)
		panic("too many ProducerDone() calls")
	} else {
		__antithesis_instrumentation__.Notify(470976)
	}
	__antithesis_instrumentation__.Notify(470974)
	if newVal == 0 {
		__antithesis_instrumentation__.Notify(470977)
		close(rc.dataChan)
	} else {
		__antithesis_instrumentation__.Notify(470978)
	}
}

func (rc *RowChannel) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(470979)
	return rc.types
}

func (rc *RowChannel) Start(ctx context.Context) { __antithesis_instrumentation__.Notify(470980) }

func (rc *RowChannel) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(470981)
	d, ok := <-rc.C
	if !ok {
		__antithesis_instrumentation__.Notify(470983)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(470984)
	}
	__antithesis_instrumentation__.Notify(470982)
	return d.Row, d.Meta
}

func (rc *RowChannel) ConsumerDone() {
	__antithesis_instrumentation__.Notify(470985)
	rc.consumerDone()
}

func (rc *RowChannel) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(470986)
	rc.consumerClosed("RowChannel")
	numSenders := atomic.LoadInt32(&rc.numSenders)

	for i := int32(0); i < numSenders; i++ {
		__antithesis_instrumentation__.Notify(470987)
		select {
		case <-rc.dataChan:
			__antithesis_instrumentation__.Notify(470988)
		default:
			__antithesis_instrumentation__.Notify(470989)
		}
	}
}

func (rc *RowChannel) DoesNotUseTxn() bool {
	__antithesis_instrumentation__.Notify(470990)
	return true
}
