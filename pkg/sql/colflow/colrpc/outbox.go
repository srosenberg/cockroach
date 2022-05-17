package colrpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/col/colserde"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecutils"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type flowStreamClient interface {
	Send(*execinfrapb.ProducerMessage) error
	Recv() (*execinfrapb.ConsumerSignal, error)
	CloseSend() error
}

type Outbox struct {
	colexecop.OneInputNode

	inputMetaInfo    colexecargs.OpWithMetaInfo
	inputInitialized bool

	typs []*types.T

	unlimitedAllocator *colmem.Allocator
	converter          *colserde.ArrowBatchConverter
	serializer         *colserde.RecordBatchSerializer

	draining uint32

	scratch struct {
		buf *bytes.Buffer
		msg *execinfrapb.ProducerMessage
	}

	span *tracing.Span

	getStats func() []*execinfrapb.ComponentStats

	runnerCtx context.Context
}

func NewOutbox(
	unlimitedAllocator *colmem.Allocator,
	input colexecargs.OpWithMetaInfo,
	typs []*types.T,
	getStats func() []*execinfrapb.ComponentStats,
) (*Outbox, error) {
	__antithesis_instrumentation__.Notify(455928)
	c, err := colserde.NewArrowBatchConverter(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(455931)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455932)
	}
	__antithesis_instrumentation__.Notify(455929)
	s, err := colserde.NewRecordBatchSerializer(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(455933)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455934)
	}
	__antithesis_instrumentation__.Notify(455930)
	o := &Outbox{

		OneInputNode:       colexecop.NewOneInputNode(colexecutils.NewDeselectorOp(unlimitedAllocator, input.Root, typs)),
		inputMetaInfo:      input,
		typs:               typs,
		unlimitedAllocator: unlimitedAllocator,
		converter:          c,
		serializer:         s,
		getStats:           getStats,
	}
	o.scratch.buf = &bytes.Buffer{}
	o.scratch.msg = &execinfrapb.ProducerMessage{}
	return o, nil
}

func (o *Outbox) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455935)
	o.scratch.buf = nil
	o.scratch.msg = nil

	o.Input = nil
	o.unlimitedAllocator.ReleaseMemory(o.unlimitedAllocator.Used())
	o.inputMetaInfo.ToClose.CloseAndLogOnErr(ctx, "outbox")
}

func (o *Outbox) Run(
	ctx context.Context,
	dialer execinfra.Dialer,
	sqlInstanceID base.SQLInstanceID,
	flowID execinfrapb.FlowID,
	streamID execinfrapb.StreamID,
	flowCtxCancel context.CancelFunc,
	connectionTimeout time.Duration,
) {
	__antithesis_instrumentation__.Notify(455936)
	flowCtx := ctx

	var outboxCtxCancel context.CancelFunc
	ctx, outboxCtxCancel = context.WithCancel(ctx)

	defer outboxCtxCancel()

	ctx, o.span = execinfra.ProcessorSpan(ctx, "outbox")
	if o.span != nil {
		__antithesis_instrumentation__.Notify(455939)
		defer o.span.Finish()
	} else {
		__antithesis_instrumentation__.Notify(455940)
	}
	__antithesis_instrumentation__.Notify(455937)

	o.runnerCtx = ctx
	ctx = logtags.AddTag(ctx, "streamID", streamID)
	log.VEventf(ctx, 2, "Outbox Dialing %s", sqlInstanceID)

	var stream execinfrapb.DistSQL_FlowStreamClient
	if err := func() error {
		__antithesis_instrumentation__.Notify(455941)
		conn, err := execinfra.GetConnForOutbox(ctx, dialer, sqlInstanceID, connectionTimeout)
		if err != nil {
			__antithesis_instrumentation__.Notify(455945)
			log.Warningf(
				ctx,
				"Outbox Dial connection error, distributed query will fail: %+v",
				err,
			)
			return err
		} else {
			__antithesis_instrumentation__.Notify(455946)
		}
		__antithesis_instrumentation__.Notify(455942)

		client := execinfrapb.NewDistSQLClient(conn)

		stream, err = client.FlowStream(flowCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(455947)
			log.Warningf(
				ctx,
				"Outbox FlowStream connection error, distributed query will fail: %+v",
				err,
			)
			return err
		} else {
			__antithesis_instrumentation__.Notify(455948)
		}
		__antithesis_instrumentation__.Notify(455943)

		log.VEvent(ctx, 2, "Outbox sending header")

		if err := stream.Send(
			&execinfrapb.ProducerMessage{Header: &execinfrapb.ProducerHeader{FlowID: flowID, StreamID: streamID}},
		); err != nil {
			__antithesis_instrumentation__.Notify(455949)
			log.Warningf(
				ctx,
				"Outbox Send header error, distributed query will fail: %+v",
				err,
			)
			return err
		} else {
			__antithesis_instrumentation__.Notify(455950)
		}
		__antithesis_instrumentation__.Notify(455944)
		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(455951)

		o.close(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(455952)
	}
	__antithesis_instrumentation__.Notify(455938)

	log.VEvent(ctx, 2, "Outbox starting normal operation")
	o.runWithStream(ctx, stream, flowCtxCancel, outboxCtxCancel)
	log.VEvent(ctx, 2, "Outbox exiting")
}

func handleStreamErr(
	ctx context.Context,
	opName redact.SafeString,
	err error,
	flowCtxCancel, outboxCtxCancel context.CancelFunc,
) {
	__antithesis_instrumentation__.Notify(455953)
	if err == io.EOF {
		__antithesis_instrumentation__.Notify(455954)
		log.VEventf(ctx, 2, "Outbox calling outboxCtxCancel after %s EOF", opName)
		outboxCtxCancel()
	} else {
		__antithesis_instrumentation__.Notify(455955)
		log.VEventf(ctx, 1, "Outbox calling flowCtxCancel after %s connection error: %+v", opName, err)
		flowCtxCancel()
	}
}

func (o *Outbox) moveToDraining(ctx context.Context, reason redact.RedactableString) {
	__antithesis_instrumentation__.Notify(455956)
	if atomic.CompareAndSwapUint32(&o.draining, 0, 1) {
		__antithesis_instrumentation__.Notify(455957)
		log.VEventf(ctx, 2, "Outbox moved to draining (%s)", reason)
	} else {
		__antithesis_instrumentation__.Notify(455958)
	}
}

func (o *Outbox) sendBatches(
	ctx context.Context, stream flowStreamClient, flowCtxCancel, outboxCtxCancel context.CancelFunc,
) (terminatedGracefully bool, errToSend error) {
	__antithesis_instrumentation__.Notify(455959)
	if o.runnerCtx == nil {
		__antithesis_instrumentation__.Notify(455962)

		o.runnerCtx = ctx
	} else {
		__antithesis_instrumentation__.Notify(455963)
	}
	__antithesis_instrumentation__.Notify(455960)
	errToSend = colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(455964)
		o.Input.Init(o.runnerCtx)
		o.inputInitialized = true
		for {
			__antithesis_instrumentation__.Notify(455965)
			if atomic.LoadUint32(&o.draining) == 1 {
				__antithesis_instrumentation__.Notify(455970)
				terminatedGracefully = true
				return
			} else {
				__antithesis_instrumentation__.Notify(455971)
			}
			__antithesis_instrumentation__.Notify(455966)

			batch := o.Input.Next()
			n := batch.Length()
			if n == 0 {
				__antithesis_instrumentation__.Notify(455972)
				terminatedGracefully = true
				return
			} else {
				__antithesis_instrumentation__.Notify(455973)
			}
			__antithesis_instrumentation__.Notify(455967)

			d, err := o.converter.BatchToArrow(batch)
			if err != nil {
				__antithesis_instrumentation__.Notify(455974)
				colexecerror.InternalError(errors.Wrap(err, "Outbox BatchToArrow data serialization error"))
			} else {
				__antithesis_instrumentation__.Notify(455975)
			}
			__antithesis_instrumentation__.Notify(455968)

			oldBufCap := o.scratch.buf.Cap()
			o.scratch.buf.Reset()
			if _, _, err := o.serializer.Serialize(o.scratch.buf, d, n); err != nil {
				__antithesis_instrumentation__.Notify(455976)
				colexecerror.InternalError(errors.Wrap(err, "Outbox Serialize data error"))
			} else {
				__antithesis_instrumentation__.Notify(455977)
			}
			__antithesis_instrumentation__.Notify(455969)

			o.unlimitedAllocator.AdjustMemoryUsage(int64(o.scratch.buf.Cap() - oldBufCap))
			o.scratch.msg.Data.RawBytes = o.scratch.buf.Bytes()

			if err := stream.Send(o.scratch.msg); err != nil {
				__antithesis_instrumentation__.Notify(455978)
				handleStreamErr(ctx, "Send (batches)", err, flowCtxCancel, outboxCtxCancel)
				return
			} else {
				__antithesis_instrumentation__.Notify(455979)
			}
		}
	})
	__antithesis_instrumentation__.Notify(455961)
	return terminatedGracefully, errToSend
}

func (o *Outbox) sendMetadata(ctx context.Context, stream flowStreamClient, errToSend error) error {
	__antithesis_instrumentation__.Notify(455980)
	msg := &execinfrapb.ProducerMessage{}
	if errToSend != nil {
		__antithesis_instrumentation__.Notify(455985)
		log.VEventf(ctx, 1, "Outbox sending an error as metadata: %v", errToSend)
		msg.Data.Metadata = append(
			msg.Data.Metadata, execinfrapb.LocalMetaToRemoteProducerMeta(ctx, execinfrapb.ProducerMetadata{Err: errToSend}),
		)
	} else {
		__antithesis_instrumentation__.Notify(455986)
	}
	__antithesis_instrumentation__.Notify(455981)
	if o.inputInitialized {
		__antithesis_instrumentation__.Notify(455987)

		if o.span != nil && func() bool {
			__antithesis_instrumentation__.Notify(455989)
			return o.getStats != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(455990)
			for _, s := range o.getStats() {
				__antithesis_instrumentation__.Notify(455991)
				o.span.RecordStructured(s)
			}
		} else {
			__antithesis_instrumentation__.Notify(455992)
		}
		__antithesis_instrumentation__.Notify(455988)
		for _, meta := range o.inputMetaInfo.MetadataSources.DrainMeta() {
			__antithesis_instrumentation__.Notify(455993)
			msg.Data.Metadata = append(msg.Data.Metadata, execinfrapb.LocalMetaToRemoteProducerMeta(ctx, meta))
		}
	} else {
		__antithesis_instrumentation__.Notify(455994)
	}
	__antithesis_instrumentation__.Notify(455982)
	if trace := execinfra.GetTraceData(ctx); trace != nil {
		__antithesis_instrumentation__.Notify(455995)
		msg.Data.Metadata = append(msg.Data.Metadata, execinfrapb.RemoteProducerMetadata{
			Value: &execinfrapb.RemoteProducerMetadata_TraceData_{
				TraceData: &execinfrapb.RemoteProducerMetadata_TraceData{
					CollectedSpans: trace,
				},
			},
		})
	} else {
		__antithesis_instrumentation__.Notify(455996)
	}
	__antithesis_instrumentation__.Notify(455983)
	if len(msg.Data.Metadata) == 0 {
		__antithesis_instrumentation__.Notify(455997)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(455998)
	}
	__antithesis_instrumentation__.Notify(455984)
	return stream.Send(msg)
}

func (o *Outbox) runWithStream(
	ctx context.Context, stream flowStreamClient, flowCtxCancel, outboxCtxCancel context.CancelFunc,
) {
	__antithesis_instrumentation__.Notify(455999)

	if flowCtxCancel == nil {
		__antithesis_instrumentation__.Notify(456004)
		flowCtxCancel = func() { __antithesis_instrumentation__.Notify(456005) }
	} else {
		__antithesis_instrumentation__.Notify(456006)
	}
	__antithesis_instrumentation__.Notify(456000)
	if outboxCtxCancel == nil {
		__antithesis_instrumentation__.Notify(456007)
		outboxCtxCancel = func() { __antithesis_instrumentation__.Notify(456008) }
	} else {
		__antithesis_instrumentation__.Notify(456009)
	}
	__antithesis_instrumentation__.Notify(456001)
	waitCh := make(chan struct{})
	go func() {
		__antithesis_instrumentation__.Notify(456010)

		for {
			__antithesis_instrumentation__.Notify(456012)
			msg, err := stream.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(456014)
				handleStreamErr(ctx, "watchdog Recv", err, flowCtxCancel, outboxCtxCancel)
				break
			} else {
				__antithesis_instrumentation__.Notify(456015)
			}
			__antithesis_instrumentation__.Notify(456013)
			switch {
			case msg.Handshake != nil:
				__antithesis_instrumentation__.Notify(456016)
				log.VEventf(ctx, 2, "Outbox received handshake: %v", msg.Handshake)
			case msg.DrainRequest != nil:
				__antithesis_instrumentation__.Notify(456017)
				log.VEventf(ctx, 2, "Outbox received drain request")
				o.moveToDraining(ctx, "consumer requested draining")
			default:
				__antithesis_instrumentation__.Notify(456018)
			}
		}
		__antithesis_instrumentation__.Notify(456011)
		close(waitCh)
	}()
	__antithesis_instrumentation__.Notify(456002)

	terminatedGracefully, errToSend := o.sendBatches(ctx, stream, flowCtxCancel, outboxCtxCancel)
	if terminatedGracefully || func() bool {
		__antithesis_instrumentation__.Notify(456019)
		return errToSend != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(456020)
		var reason redact.RedactableString
		if errToSend != nil {
			__antithesis_instrumentation__.Notify(456022)
			reason = redact.Sprintf("encountered error when sending batches: %v", errToSend)
		} else {
			__antithesis_instrumentation__.Notify(456023)
			reason = redact.Sprint(redact.SafeString("terminated gracefully"))
		}
		__antithesis_instrumentation__.Notify(456021)
		o.moveToDraining(ctx, reason)
		if err := o.sendMetadata(ctx, stream, errToSend); err != nil {
			__antithesis_instrumentation__.Notify(456024)
			handleStreamErr(ctx, "Send (metadata)", err, flowCtxCancel, outboxCtxCancel)
		} else {
			__antithesis_instrumentation__.Notify(456025)

			if err := stream.CloseSend(); err != nil {
				__antithesis_instrumentation__.Notify(456026)
				handleStreamErr(ctx, "CloseSend", err, flowCtxCancel, outboxCtxCancel)
			} else {
				__antithesis_instrumentation__.Notify(456027)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(456028)
	}
	__antithesis_instrumentation__.Notify(456003)

	o.close(ctx)
	<-waitCh
}
