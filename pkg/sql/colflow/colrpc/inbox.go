package colrpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"math"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/colserde"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type flowStreamServer interface {
	Send(*execinfrapb.ConsumerSignal) error
	Recv() (*execinfrapb.ProducerMessage, error)
}

type Inbox struct {
	colexecop.ZeroInputNode
	colexecop.InitHelper

	typs []*types.T

	allocator  *colmem.Allocator
	converter  *colserde.ArrowBatchConverter
	serializer *colserde.RecordBatchSerializer

	streamID execinfrapb.StreamID

	streamCh chan flowStreamServer

	contextCh chan context.Context

	timeoutCh chan error

	flowCtxDone <-chan struct{}

	errCh chan error

	ctxInterceptorFn func(context.Context)

	done bool

	bufferedMeta []execinfrapb.ProducerMetadata

	stream flowStreamServer

	admissionQ    *admission.WorkQueue
	admissionInfo admission.WorkInfo

	statsAtomics struct {
		rowsRead int64

		bytesRead int64

		numMessages int64
	}

	deserializationStopWatch *timeutil.StopWatch

	scratch struct {
		data []*array.Data
		b    coldata.Batch
	}
}

var _ colexecop.Operator = &Inbox{}

func NewInbox(
	allocator *colmem.Allocator, typs []*types.T, streamID execinfrapb.StreamID,
) (*Inbox, error) {
	__antithesis_instrumentation__.Notify(455804)
	c, err := colserde.NewArrowBatchConverter(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(455807)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455808)
	}
	__antithesis_instrumentation__.Notify(455805)
	s, err := colserde.NewRecordBatchSerializer(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(455809)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455810)
	}
	__antithesis_instrumentation__.Notify(455806)
	i := &Inbox{
		typs:                     typs,
		allocator:                allocator,
		converter:                c,
		serializer:               s,
		streamID:                 streamID,
		streamCh:                 make(chan flowStreamServer, 1),
		contextCh:                make(chan context.Context, 1),
		timeoutCh:                make(chan error, 1),
		errCh:                    make(chan error, 1),
		deserializationStopWatch: timeutil.NewStopWatch(),
	}
	i.scratch.data = make([]*array.Data, len(typs))
	return i, nil
}

func NewInboxWithFlowCtxDone(
	allocator *colmem.Allocator,
	typs []*types.T,
	streamID execinfrapb.StreamID,
	flowCtxDone <-chan struct{},
) (*Inbox, error) {
	__antithesis_instrumentation__.Notify(455811)
	i, err := NewInbox(allocator, typs, streamID)
	if err != nil {
		__antithesis_instrumentation__.Notify(455813)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455814)
	}
	__antithesis_instrumentation__.Notify(455812)
	i.flowCtxDone = flowCtxDone
	return i, nil
}

func NewInboxWithAdmissionControl(
	allocator *colmem.Allocator,
	typs []*types.T,
	streamID execinfrapb.StreamID,
	flowCtxDone <-chan struct{},
	admissionQ *admission.WorkQueue,
	admissionInfo admission.WorkInfo,
) (*Inbox, error) {
	__antithesis_instrumentation__.Notify(455815)
	i, err := NewInboxWithFlowCtxDone(allocator, typs, streamID, flowCtxDone)
	if err != nil {
		__antithesis_instrumentation__.Notify(455817)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455818)
	}
	__antithesis_instrumentation__.Notify(455816)
	i.admissionQ = admissionQ
	i.admissionInfo = admissionInfo
	return i, err
}

func (i *Inbox) close() {
	__antithesis_instrumentation__.Notify(455819)
	if !i.done {
		__antithesis_instrumentation__.Notify(455820)
		i.done = true
		close(i.errCh)
	} else {
		__antithesis_instrumentation__.Notify(455821)
	}
}

func (i *Inbox) checkFlowCtxCancellation() error {
	__antithesis_instrumentation__.Notify(455822)
	select {
	case <-i.flowCtxDone:
		__antithesis_instrumentation__.Notify(455823)
		return cancelchecker.QueryCanceledError
	default:
		__antithesis_instrumentation__.Notify(455824)
		return nil
	}
}

func (i *Inbox) RunWithStream(streamCtx context.Context, stream flowStreamServer) error {
	__antithesis_instrumentation__.Notify(455825)
	streamCtx = logtags.AddTag(streamCtx, "streamID", i.streamID)
	log.VEvent(streamCtx, 2, "Inbox handling stream")
	defer log.VEvent(streamCtx, 2, "Inbox exited stream handler")

	i.streamCh <- stream
	var readerCtx context.Context
	select {
	case err := <-i.errCh:
		__antithesis_instrumentation__.Notify(455827)

		return err
	case readerCtx = <-i.contextCh:
		__antithesis_instrumentation__.Notify(455828)
		log.VEvent(streamCtx, 2, "Inbox reader arrived")
	case <-streamCtx.Done():
		__antithesis_instrumentation__.Notify(455829)
		return errors.Wrap(streamCtx.Err(), "streamCtx error while waiting for reader (remote client canceled)")
	case <-i.flowCtxDone:
		__antithesis_instrumentation__.Notify(455830)

		return cancelchecker.QueryCanceledError
	}
	__antithesis_instrumentation__.Notify(455826)

	select {
	case err := <-i.errCh:
		__antithesis_instrumentation__.Notify(455831)

		return err
	case <-i.flowCtxDone:
		__antithesis_instrumentation__.Notify(455832)

		return cancelchecker.QueryCanceledError
	case <-readerCtx.Done():
		__antithesis_instrumentation__.Notify(455833)

		return i.checkFlowCtxCancellation()
	case <-streamCtx.Done():
		__antithesis_instrumentation__.Notify(455834)

		return errors.Wrap(streamCtx.Err(), "streamCtx error in Inbox stream handler (remote client canceled)")
	}
}

func (i *Inbox) Timeout(err error) {
	__antithesis_instrumentation__.Notify(455835)
	i.timeoutCh <- err
}

func (i *Inbox) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455836)
	if !i.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(455838)
		return
	} else {
		__antithesis_instrumentation__.Notify(455839)
	}
	__antithesis_instrumentation__.Notify(455837)

	i.Ctx = logtags.AddTag(i.Ctx, "streamID", i.streamID)

	if err := func() error {
		__antithesis_instrumentation__.Notify(455840)

		select {
		case i.stream = <-i.streamCh:
			__antithesis_instrumentation__.Notify(455843)
		case err := <-i.timeoutCh:
			__antithesis_instrumentation__.Notify(455844)
			i.errCh <- errors.Wrap(err, "remote stream arrived too late")
			return err
		case <-i.flowCtxDone:
			__antithesis_instrumentation__.Notify(455845)
			i.errCh <- cancelchecker.QueryCanceledError
			return cancelchecker.QueryCanceledError
		case <-i.Ctx.Done():
			__antithesis_instrumentation__.Notify(455846)

			errToThrow := i.Ctx.Err()
			if err := i.checkFlowCtxCancellation(); err != nil {
				__antithesis_instrumentation__.Notify(455848)

				i.errCh <- err
				errToThrow = err
			} else {
				__antithesis_instrumentation__.Notify(455849)
			}
			__antithesis_instrumentation__.Notify(455847)
			return errToThrow
		}
		__antithesis_instrumentation__.Notify(455841)

		if i.ctxInterceptorFn != nil {
			__antithesis_instrumentation__.Notify(455850)
			i.ctxInterceptorFn(i.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(455851)
		}
		__antithesis_instrumentation__.Notify(455842)
		i.contextCh <- i.Ctx
		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(455852)

		i.close()
		log.VEventf(ctx, 1, "Inbox encountered an error in Init: %v", err)
		colexecerror.ExpectedError(err)
	} else {
		__antithesis_instrumentation__.Notify(455853)
	}
}

func (i *Inbox) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455854)
	if i.done {
		__antithesis_instrumentation__.Notify(455857)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(455858)
	}
	__antithesis_instrumentation__.Notify(455855)

	defer func() {
		__antithesis_instrumentation__.Notify(455859)

		if panicObj := recover(); panicObj != nil {
			__antithesis_instrumentation__.Notify(455860)

			i.close()
			err := logcrash.PanicAsError(0, panicObj)
			log.VEventf(i.Ctx, 1, "Inbox encountered an error in Next: %v", err)

			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(455861)
		}
	}()
	__antithesis_instrumentation__.Notify(455856)

	i.deserializationStopWatch.Start()
	defer i.deserializationStopWatch.Stop()
	for {
		__antithesis_instrumentation__.Notify(455862)
		i.deserializationStopWatch.Stop()
		m, err := i.stream.Recv()
		i.deserializationStopWatch.Start()
		atomic.AddInt64(&i.statsAtomics.numMessages, 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(455869)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(455871)

				i.close()
				return coldata.ZeroBatch
			} else {
				__antithesis_instrumentation__.Notify(455872)
			}
			__antithesis_instrumentation__.Notify(455870)

			err = pgerror.Wrap(err, pgcode.InternalConnectionFailure, "inbox communication error")
			i.errCh <- err
			colexecerror.ExpectedError(err)
		} else {
			__antithesis_instrumentation__.Notify(455873)
		}
		__antithesis_instrumentation__.Notify(455863)
		if len(m.Data.Metadata) != 0 {
			__antithesis_instrumentation__.Notify(455874)
			for _, rpm := range m.Data.Metadata {
				__antithesis_instrumentation__.Notify(455876)
				meta, ok := execinfrapb.RemoteProducerMetaToLocalMeta(i.Ctx, rpm)
				if !ok {
					__antithesis_instrumentation__.Notify(455879)
					continue
				} else {
					__antithesis_instrumentation__.Notify(455880)
				}
				__antithesis_instrumentation__.Notify(455877)
				if meta.Err != nil {
					__antithesis_instrumentation__.Notify(455881)

					colexecerror.ExpectedError(meta.Err)
				} else {
					__antithesis_instrumentation__.Notify(455882)
				}
				__antithesis_instrumentation__.Notify(455878)
				i.bufferedMeta = append(i.bufferedMeta, meta)
			}
			__antithesis_instrumentation__.Notify(455875)

			continue
		} else {
			__antithesis_instrumentation__.Notify(455883)
		}
		__antithesis_instrumentation__.Notify(455864)
		if len(m.Data.RawBytes) == 0 {
			__antithesis_instrumentation__.Notify(455884)

			continue
		} else {
			__antithesis_instrumentation__.Notify(455885)
		}
		__antithesis_instrumentation__.Notify(455865)
		numSerializedBytes := int64(len(m.Data.RawBytes))
		atomic.AddInt64(&i.statsAtomics.bytesRead, numSerializedBytes)

		i.allocator.AdjustMemoryUsage(numSerializedBytes)

		if i.admissionQ != nil {
			__antithesis_instrumentation__.Notify(455886)
			if _, err := i.admissionQ.Admit(i.Ctx, i.admissionInfo); err != nil {
				__antithesis_instrumentation__.Notify(455887)

				colexecerror.ExpectedError(err)
			} else {
				__antithesis_instrumentation__.Notify(455888)
			}
		} else {
			__antithesis_instrumentation__.Notify(455889)
		}
		__antithesis_instrumentation__.Notify(455866)
		i.scratch.data = i.scratch.data[:0]
		batchLength, err := i.serializer.Deserialize(&i.scratch.data, m.Data.RawBytes)

		m.Data.RawBytes = nil
		if err != nil {
			__antithesis_instrumentation__.Notify(455890)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(455891)
		}
		__antithesis_instrumentation__.Notify(455867)

		const maxBatchMemSize = math.MaxInt64
		i.scratch.b, _ = i.allocator.ResetMaybeReallocate(i.typs, i.scratch.b, batchLength, maxBatchMemSize)
		i.allocator.PerformOperation(i.scratch.b.ColVecs(), func() {
			__antithesis_instrumentation__.Notify(455892)
			if err := i.converter.ArrowToBatch(i.scratch.data, batchLength, i.scratch.b); err != nil {
				__antithesis_instrumentation__.Notify(455893)
				colexecerror.InternalError(err)
			} else {
				__antithesis_instrumentation__.Notify(455894)
			}
		})
		__antithesis_instrumentation__.Notify(455868)

		i.allocator.AdjustMemoryUsage(-numSerializedBytes)
		atomic.AddInt64(&i.statsAtomics.rowsRead, int64(i.scratch.b.Length()))
		return i.scratch.b
	}
}

func (i *Inbox) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(455895)
	return atomic.LoadInt64(&i.statsAtomics.bytesRead)
}

func (i *Inbox) GetRowsRead() int64 {
	__antithesis_instrumentation__.Notify(455896)
	return atomic.LoadInt64(&i.statsAtomics.rowsRead)
}

func (i *Inbox) GetDeserializationTime() time.Duration {
	__antithesis_instrumentation__.Notify(455897)
	return i.deserializationStopWatch.Elapsed()
}

func (i *Inbox) GetNumMessages() int64 {
	__antithesis_instrumentation__.Notify(455898)
	return atomic.LoadInt64(&i.statsAtomics.numMessages)
}

func (i *Inbox) sendDrainSignal(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(455899)
	log.VEvent(ctx, 2, "Inbox sending drain signal to Outbox")
	if err := i.stream.Send(&execinfrapb.ConsumerSignal{DrainRequest: &execinfrapb.DrainRequest{}}); err != nil {
		__antithesis_instrumentation__.Notify(455901)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(455903)
			log.Warningf(ctx, "Inbox unable to send drain signal to Outbox: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(455904)
		}
		__antithesis_instrumentation__.Notify(455902)
		return err
	} else {
		__antithesis_instrumentation__.Notify(455905)
	}
	__antithesis_instrumentation__.Notify(455900)
	return nil
}

func (i *Inbox) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(455906)
	allMeta := i.bufferedMeta
	i.bufferedMeta = i.bufferedMeta[:0]

	if i.done {
		__antithesis_instrumentation__.Notify(455910)

		return allMeta
	} else {
		__antithesis_instrumentation__.Notify(455911)
	}
	__antithesis_instrumentation__.Notify(455907)
	defer i.close()

	if err := i.sendDrainSignal(i.Ctx); err != nil {
		__antithesis_instrumentation__.Notify(455912)
		return allMeta
	} else {
		__antithesis_instrumentation__.Notify(455913)
	}
	__antithesis_instrumentation__.Notify(455908)

	for {
		__antithesis_instrumentation__.Notify(455914)
		msg, err := i.stream.Recv()
		if err != nil {
			__antithesis_instrumentation__.Notify(455916)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(455919)
				break
			} else {
				__antithesis_instrumentation__.Notify(455920)
			}
			__antithesis_instrumentation__.Notify(455917)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(455921)
				log.Warningf(i.Ctx, "Inbox Recv connection error while draining metadata: %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(455922)
			}
			__antithesis_instrumentation__.Notify(455918)
			return allMeta
		} else {
			__antithesis_instrumentation__.Notify(455923)
		}
		__antithesis_instrumentation__.Notify(455915)
		for _, remoteMeta := range msg.Data.Metadata {
			__antithesis_instrumentation__.Notify(455924)
			meta, ok := execinfrapb.RemoteProducerMetaToLocalMeta(i.Ctx, remoteMeta)
			if !ok {
				__antithesis_instrumentation__.Notify(455926)
				continue
			} else {
				__antithesis_instrumentation__.Notify(455927)
			}
			__antithesis_instrumentation__.Notify(455925)
			allMeta = append(allMeta, meta)
		}
	}
	__antithesis_instrumentation__.Notify(455909)

	return allMeta
}
