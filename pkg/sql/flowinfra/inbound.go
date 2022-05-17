package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type InboundStreamHandler interface {
	Run(
		ctx context.Context, stream execinfrapb.DistSQL_FlowStreamServer, firstMsg *execinfrapb.ProducerMessage, f *FlowBase,
	) error

	Timeout(err error)
}

type RowInboundStreamHandler struct {
	execinfra.RowReceiver
}

var _ InboundStreamHandler = RowInboundStreamHandler{}

func (s RowInboundStreamHandler) Run(
	ctx context.Context,
	stream execinfrapb.DistSQL_FlowStreamServer,
	firstMsg *execinfrapb.ProducerMessage,
	f *FlowBase,
) error {
	__antithesis_instrumentation__.Notify(491853)
	return processInboundStream(ctx, stream, firstMsg, s.RowReceiver, f)
}

func (s RowInboundStreamHandler) Timeout(err error) {
	__antithesis_instrumentation__.Notify(491854)
	s.Push(
		nil,
		&execinfrapb.ProducerMetadata{Err: err},
	)
	s.ProducerDone()
}

func processInboundStream(
	ctx context.Context,
	stream execinfrapb.DistSQL_FlowStreamServer,
	firstMsg *execinfrapb.ProducerMessage,
	dst execinfra.RowReceiver,
	f *FlowBase,
) error {
	__antithesis_instrumentation__.Notify(491855)

	err := processInboundStreamHelper(ctx, stream, firstMsg, dst, f)

	if err != nil {
		__antithesis_instrumentation__.Notify(491857)
		log.VEventf(ctx, 1, "inbound stream error: %s", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491858)
	}
	__antithesis_instrumentation__.Notify(491856)
	log.VEventf(ctx, 1, "inbound stream done")

	return nil
}

func processInboundStreamHelper(
	ctx context.Context,
	stream execinfrapb.DistSQL_FlowStreamServer,
	firstMsg *execinfrapb.ProducerMessage,
	dst execinfra.RowReceiver,
	f *FlowBase,
) error {
	__antithesis_instrumentation__.Notify(491859)
	draining := false
	var sd StreamDecoder

	sendErrToConsumer := func(err error) {
		__antithesis_instrumentation__.Notify(491863)
		if err != nil {
			__antithesis_instrumentation__.Notify(491865)
			dst.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(491866)
		}
		__antithesis_instrumentation__.Notify(491864)
		dst.ProducerDone()
	}
	__antithesis_instrumentation__.Notify(491860)

	if firstMsg != nil {
		__antithesis_instrumentation__.Notify(491867)
		if res := processProducerMessage(
			ctx, f, stream, dst, &sd, &draining, firstMsg,
		); res.err != nil || func() bool {
			__antithesis_instrumentation__.Notify(491868)
			return res.consumerClosed == true
		}() == true {
			__antithesis_instrumentation__.Notify(491869)
			sendErrToConsumer(res.err)
			return res.err
		} else {
			__antithesis_instrumentation__.Notify(491870)
		}
	} else {
		__antithesis_instrumentation__.Notify(491871)
	}
	__antithesis_instrumentation__.Notify(491861)

	errChan := make(chan error, 1)

	f.GetWaitGroup().Add(1)
	go func() {
		__antithesis_instrumentation__.Notify(491872)
		defer f.GetWaitGroup().Done()
		for {
			__antithesis_instrumentation__.Notify(491873)
			msg, err := stream.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(491875)
				if err != io.EOF {
					__antithesis_instrumentation__.Notify(491877)

					err = pgerror.Wrap(err, pgcode.InternalConnectionFailure, "inbox communication error")
					sendErrToConsumer(err)
					errChan <- err
					return
				} else {
					__antithesis_instrumentation__.Notify(491878)
				}
				__antithesis_instrumentation__.Notify(491876)

				sendErrToConsumer(nil)
				errChan <- nil
				return
			} else {
				__antithesis_instrumentation__.Notify(491879)
			}
			__antithesis_instrumentation__.Notify(491874)

			if res := processProducerMessage(
				ctx, f, stream, dst, &sd, &draining, msg,
			); res.err != nil || func() bool {
				__antithesis_instrumentation__.Notify(491880)
				return res.consumerClosed == true
			}() == true {
				__antithesis_instrumentation__.Notify(491881)
				sendErrToConsumer(res.err)
				errChan <- res.err
				return
			} else {
				__antithesis_instrumentation__.Notify(491882)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(491862)

	select {
	case <-f.GetCtxDone():
		__antithesis_instrumentation__.Notify(491883)
		return cancelchecker.QueryCanceledError
	case err := <-errChan:
		__antithesis_instrumentation__.Notify(491884)
		return err
	}
}

func sendDrainSignalToStreamProducer(
	ctx context.Context, stream execinfrapb.DistSQL_FlowStreamServer,
) error {
	__antithesis_instrumentation__.Notify(491885)
	log.VEvent(ctx, 1, "sending drain signal to producer")
	sig := execinfrapb.ConsumerSignal{DrainRequest: &execinfrapb.DrainRequest{}}
	return stream.Send(&sig)
}

func processProducerMessage(
	ctx context.Context,
	flowBase *FlowBase,
	stream execinfrapb.DistSQL_FlowStreamServer,
	dst execinfra.RowReceiver,
	sd *StreamDecoder,
	draining *bool,
	msg *execinfrapb.ProducerMessage,
) processMessageResult {
	__antithesis_instrumentation__.Notify(491886)
	err := sd.AddMessage(ctx, msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(491890)
		return processMessageResult{
			err: errors.Wrapf(err, "%s",

				log.FormatWithContextTags(ctx, "decoding error")),
			consumerClosed: false,
		}
	} else {
		__antithesis_instrumentation__.Notify(491891)
	}
	__antithesis_instrumentation__.Notify(491887)
	var admissionQ *admission.WorkQueue
	if flowBase.Cfg != nil {
		__antithesis_instrumentation__.Notify(491892)
		admissionQ = flowBase.Cfg.SQLSQLResponseAdmissionQ
	} else {
		__antithesis_instrumentation__.Notify(491893)
	}
	__antithesis_instrumentation__.Notify(491888)
	if admissionQ != nil {
		__antithesis_instrumentation__.Notify(491894)
		if _, err := admissionQ.Admit(ctx, flowBase.admissionInfo); err != nil {
			__antithesis_instrumentation__.Notify(491895)
			return processMessageResult{err: err, consumerClosed: false}
		} else {
			__antithesis_instrumentation__.Notify(491896)
		}
	} else {
		__antithesis_instrumentation__.Notify(491897)
	}
	__antithesis_instrumentation__.Notify(491889)
	var types []*types.T
	for {
		__antithesis_instrumentation__.Notify(491898)
		row, meta, err := sd.GetRow(nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(491903)
			return processMessageResult{err: err, consumerClosed: false}
		} else {
			__antithesis_instrumentation__.Notify(491904)
		}
		__antithesis_instrumentation__.Notify(491899)
		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(491905)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(491906)

			return processMessageResult{err: nil, consumerClosed: false}
		} else {
			__antithesis_instrumentation__.Notify(491907)
		}
		__antithesis_instrumentation__.Notify(491900)

		if log.V(3) && func() bool {
			__antithesis_instrumentation__.Notify(491908)
			return row != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(491909)
			if types == nil {
				__antithesis_instrumentation__.Notify(491911)
				types = sd.Types()
			} else {
				__antithesis_instrumentation__.Notify(491912)
			}
			__antithesis_instrumentation__.Notify(491910)
			log.Infof(ctx, "inbound stream pushing row %s", row.String(types))
		} else {
			__antithesis_instrumentation__.Notify(491913)
		}
		__antithesis_instrumentation__.Notify(491901)
		if *draining && func() bool {
			__antithesis_instrumentation__.Notify(491914)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(491915)

			continue
		} else {
			__antithesis_instrumentation__.Notify(491916)
		}
		__antithesis_instrumentation__.Notify(491902)
		switch dst.Push(row, meta) {
		case execinfra.NeedMoreRows:
			__antithesis_instrumentation__.Notify(491917)
			continue
		case execinfra.DrainRequested:
			__antithesis_instrumentation__.Notify(491918)

			if !*draining {
				__antithesis_instrumentation__.Notify(491921)
				*draining = true
				if err := sendDrainSignalToStreamProducer(ctx, stream); err != nil {
					__antithesis_instrumentation__.Notify(491922)
					log.Errorf(ctx, "draining error: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(491923)
				}
			} else {
				__antithesis_instrumentation__.Notify(491924)
			}
		case execinfra.ConsumerClosed:
			__antithesis_instrumentation__.Notify(491919)
			return processMessageResult{err: nil, consumerClosed: true}
		default:
			__antithesis_instrumentation__.Notify(491920)
		}
	}
}

type processMessageResult struct {
	err            error
	consumerClosed bool
}
