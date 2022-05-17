package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"go.opentelemetry.io/otel/attribute"
)

const OutboxBufRows = 16
const outboxFlushPeriod = 100 * time.Microsecond

type flowStream interface {
	Send(*execinfrapb.ProducerMessage) error
	Recv() (*execinfrapb.ConsumerSignal, error)
}

type Outbox struct {
	execinfra.RowChannel

	flowCtx       *execinfra.FlowCtx
	streamID      execinfrapb.StreamID
	sqlInstanceID base.SQLInstanceID

	stream flowStream

	encoder StreamEncoder

	numRows int

	flowCtxCancel context.CancelFunc

	err error

	statsCollectionEnabled bool
	stats                  execinfrapb.ComponentStats

	numOutboxes *int32

	isGatewayNode bool
}

var _ execinfra.RowReceiver = &Outbox{}
var _ Startable = &Outbox{}

func NewOutbox(
	flowCtx *execinfra.FlowCtx,
	sqlInstanceID base.SQLInstanceID,
	streamID execinfrapb.StreamID,
	numOutboxes *int32,
	isGatewayNode bool,
) *Outbox {
	__antithesis_instrumentation__.Notify(491925)
	m := &Outbox{flowCtx: flowCtx, sqlInstanceID: sqlInstanceID}
	m.encoder.SetHeaderFields(flowCtx.ID, streamID)
	m.streamID = streamID
	m.numOutboxes = numOutboxes
	m.isGatewayNode = isGatewayNode
	m.stats.Component = flowCtx.StreamComponentID(streamID)
	return m
}

func (m *Outbox) Init(typs []*types.T) {
	__antithesis_instrumentation__.Notify(491926)
	if typs == nil {
		__antithesis_instrumentation__.Notify(491928)

		typs = make([]*types.T, 0)
	} else {
		__antithesis_instrumentation__.Notify(491929)
	}
	__antithesis_instrumentation__.Notify(491927)
	m.RowChannel.InitWithNumSenders(typs, 1)
	m.encoder.Init(typs)
}

func (m *Outbox) AddRow(
	ctx context.Context, row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) error {
	__antithesis_instrumentation__.Notify(491930)
	mustFlush := false
	var encodingErr error
	if meta != nil {
		__antithesis_instrumentation__.Notify(491934)
		m.encoder.AddMetadata(ctx, *meta)

		mustFlush = meta.Err != nil
	} else {
		__antithesis_instrumentation__.Notify(491935)
		encodingErr = m.encoder.AddRow(row)
		if encodingErr != nil {
			__antithesis_instrumentation__.Notify(491937)
			m.encoder.AddMetadata(ctx, execinfrapb.ProducerMetadata{Err: encodingErr})
			mustFlush = true
		} else {
			__antithesis_instrumentation__.Notify(491938)
		}
		__antithesis_instrumentation__.Notify(491936)
		if m.statsCollectionEnabled {
			__antithesis_instrumentation__.Notify(491939)
			m.stats.NetTx.TuplesSent.Add(1)
		} else {
			__antithesis_instrumentation__.Notify(491940)
		}
	}
	__antithesis_instrumentation__.Notify(491931)
	m.numRows++
	var flushErr error
	if m.numRows >= OutboxBufRows || func() bool {
		__antithesis_instrumentation__.Notify(491941)
		return mustFlush == true
	}() == true {
		__antithesis_instrumentation__.Notify(491942)
		flushErr = m.flush(ctx)
	} else {
		__antithesis_instrumentation__.Notify(491943)
	}
	__antithesis_instrumentation__.Notify(491932)
	if encodingErr != nil {
		__antithesis_instrumentation__.Notify(491944)
		return encodingErr
	} else {
		__antithesis_instrumentation__.Notify(491945)
	}
	__antithesis_instrumentation__.Notify(491933)
	return flushErr
}

func (m *Outbox) flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(491946)
	if m.numRows == 0 && func() bool {
		__antithesis_instrumentation__.Notify(491953)
		return m.encoder.HasHeaderBeenSent() == true
	}() == true {
		__antithesis_instrumentation__.Notify(491954)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(491955)
	}
	__antithesis_instrumentation__.Notify(491947)
	msg := m.encoder.FormMessage(ctx)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(491956)
		log.Infof(ctx, "flushing outbox")
	} else {
		__antithesis_instrumentation__.Notify(491957)
	}
	__antithesis_instrumentation__.Notify(491948)
	sendErr := m.stream.Send(msg)
	if m.statsCollectionEnabled {
		__antithesis_instrumentation__.Notify(491958)
		m.stats.NetTx.BytesSent.Add(int64(msg.Size()))
		m.stats.NetTx.MessagesSent.Add(1)
	} else {
		__antithesis_instrumentation__.Notify(491959)
	}
	__antithesis_instrumentation__.Notify(491949)
	for _, rpm := range msg.Data.Metadata {
		__antithesis_instrumentation__.Notify(491960)
		if metricsMeta, ok := rpm.Value.(*execinfrapb.RemoteProducerMetadata_Metrics_); ok {
			__antithesis_instrumentation__.Notify(491961)
			metricsMeta.Metrics.Release()
		} else {
			__antithesis_instrumentation__.Notify(491962)
		}
	}
	__antithesis_instrumentation__.Notify(491950)
	if sendErr != nil {
		__antithesis_instrumentation__.Notify(491963)

		m.stream = nil
		if log.V(1) {
			__antithesis_instrumentation__.Notify(491964)
			log.Errorf(ctx, "outbox flush error: %s", sendErr)
		} else {
			__antithesis_instrumentation__.Notify(491965)
		}
	} else {
		__antithesis_instrumentation__.Notify(491966)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(491967)
			log.Infof(ctx, "outbox flushed")
		} else {
			__antithesis_instrumentation__.Notify(491968)
		}
	}
	__antithesis_instrumentation__.Notify(491951)
	if sendErr != nil {
		__antithesis_instrumentation__.Notify(491969)
		return sendErr
	} else {
		__antithesis_instrumentation__.Notify(491970)
	}
	__antithesis_instrumentation__.Notify(491952)

	m.numRows = 0
	return nil
}

func (m *Outbox) mainLoop(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(491971)

	defer m.RowChannel.ConsumerClosed()

	var span *tracing.Span
	ctx, span = execinfra.ProcessorSpan(ctx, "outbox")
	defer span.Finish()
	if span != nil && func() bool {
		__antithesis_instrumentation__.Notify(491976)
		return span.IsVerbose() == true
	}() == true {
		__antithesis_instrumentation__.Notify(491977)
		m.statsCollectionEnabled = true
		span.SetTag(execinfrapb.FlowIDTagKey, attribute.StringValue(m.flowCtx.ID.String()))
		span.SetTag(execinfrapb.StreamIDTagKey, attribute.IntValue(int(m.streamID)))
	} else {
		__antithesis_instrumentation__.Notify(491978)
	}
	__antithesis_instrumentation__.Notify(491972)

	if m.stream == nil {
		__antithesis_instrumentation__.Notify(491979)
		conn, err := execinfra.GetConnForOutbox(
			ctx, m.flowCtx.Cfg.PodNodeDialer, m.sqlInstanceID, SettingFlowStreamTimeout.Get(&m.flowCtx.Cfg.Settings.SV),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(491983)

			log.Infof(ctx, "outbox: connection dial error: %+v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491984)
		}
		__antithesis_instrumentation__.Notify(491980)
		client := execinfrapb.NewDistSQLClient(conn)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(491985)
			log.Infof(ctx, "outbox: calling FlowStream")
		} else {
			__antithesis_instrumentation__.Notify(491986)
		}
		__antithesis_instrumentation__.Notify(491981)

		m.stream, err = client.FlowStream(context.TODO())
		if err != nil {
			__antithesis_instrumentation__.Notify(491987)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(491989)
				log.Infof(ctx, "FlowStream error: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(491990)
			}
			__antithesis_instrumentation__.Notify(491988)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491991)
		}
		__antithesis_instrumentation__.Notify(491982)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(491992)
			log.Infof(ctx, "outbox: FlowStream returned")
		} else {
			__antithesis_instrumentation__.Notify(491993)
		}
	} else {
		__antithesis_instrumentation__.Notify(491994)
	}
	__antithesis_instrumentation__.Notify(491973)

	var flushTimer timeutil.Timer
	defer flushTimer.Stop()

	draining := false

	listenToConsumerCtx, cancel := contextutil.WithCancel(ctx)
	drainCh, err := m.listenForDrainSignalFromConsumer(listenToConsumerCtx)
	defer cancel()
	if err != nil {
		__antithesis_instrumentation__.Notify(491995)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491996)
	}
	__antithesis_instrumentation__.Notify(491974)

	if err := m.flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(491997)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491998)
	}
	__antithesis_instrumentation__.Notify(491975)

	for {
		__antithesis_instrumentation__.Notify(491999)
		select {
		case msg, ok := <-m.RowChannel.C:
			__antithesis_instrumentation__.Notify(492000)
			if !ok {
				__antithesis_instrumentation__.Notify(492005)

				if m.statsCollectionEnabled {
					__antithesis_instrumentation__.Notify(492007)
					err := m.flush(ctx)
					if err != nil {
						__antithesis_instrumentation__.Notify(492010)
						return err
					} else {
						__antithesis_instrumentation__.Notify(492011)
					}
					__antithesis_instrumentation__.Notify(492008)
					if !m.isGatewayNode && func() bool {
						__antithesis_instrumentation__.Notify(492012)
						return m.numOutboxes != nil == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(492013)
						return atomic.AddInt32(m.numOutboxes, -1) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(492014)

						m.stats.FlowStats.MaxMemUsage.Set(uint64(m.flowCtx.EvalCtx.Mon.MaximumBytes()))
						m.stats.FlowStats.MaxDiskUsage.Set(uint64(m.flowCtx.DiskMonitor.MaximumBytes()))
					} else {
						__antithesis_instrumentation__.Notify(492015)
					}
					__antithesis_instrumentation__.Notify(492009)
					span.RecordStructured(&m.stats)
					if trace := execinfra.GetTraceData(ctx); trace != nil {
						__antithesis_instrumentation__.Notify(492016)
						err := m.AddRow(ctx, nil, &execinfrapb.ProducerMetadata{TraceData: trace})
						if err != nil {
							__antithesis_instrumentation__.Notify(492017)
							return err
						} else {
							__antithesis_instrumentation__.Notify(492018)
						}
					} else {
						__antithesis_instrumentation__.Notify(492019)
					}
				} else {
					__antithesis_instrumentation__.Notify(492020)
				}
				__antithesis_instrumentation__.Notify(492006)
				return m.flush(ctx)
			} else {
				__antithesis_instrumentation__.Notify(492021)
			}
			__antithesis_instrumentation__.Notify(492001)
			if !draining || func() bool {
				__antithesis_instrumentation__.Notify(492022)
				return msg.Meta != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(492023)

				err := m.AddRow(ctx, msg.Row, msg.Meta)
				if err != nil {
					__antithesis_instrumentation__.Notify(492026)
					return err
				} else {
					__antithesis_instrumentation__.Notify(492027)
				}
				__antithesis_instrumentation__.Notify(492024)
				if msg.Meta != nil {
					__antithesis_instrumentation__.Notify(492028)

					msg.Meta.Release()
				} else {
					__antithesis_instrumentation__.Notify(492029)
				}
				__antithesis_instrumentation__.Notify(492025)

				if m.numRows == 1 {
					__antithesis_instrumentation__.Notify(492030)
					flushTimer.Reset(outboxFlushPeriod)
				} else {
					__antithesis_instrumentation__.Notify(492031)
				}
			} else {
				__antithesis_instrumentation__.Notify(492032)
			}
		case <-flushTimer.C:
			__antithesis_instrumentation__.Notify(492002)
			flushTimer.Read = true
			err := m.flush(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(492033)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492034)
			}
		case drainSignal := <-drainCh:
			__antithesis_instrumentation__.Notify(492003)
			if drainSignal.err != nil {
				__antithesis_instrumentation__.Notify(492035)

				m.flowCtxCancel()

				m.stream = nil
				return drainSignal.err
			} else {
				__antithesis_instrumentation__.Notify(492036)
			}
			__antithesis_instrumentation__.Notify(492004)
			drainCh = nil
			if drainSignal.drainRequested {
				__antithesis_instrumentation__.Notify(492037)

				draining = true
				m.RowChannel.ConsumerDone()
			} else {
				__antithesis_instrumentation__.Notify(492038)

				return nil
			}
		}
	}
}

type drainSignal struct {
	drainRequested bool

	err error
}

func (m *Outbox) listenForDrainSignalFromConsumer(ctx context.Context) (<-chan drainSignal, error) {
	__antithesis_instrumentation__.Notify(492039)
	ch := make(chan drainSignal, 1)

	stream := m.stream
	if err := m.flowCtx.Cfg.Stopper.RunAsyncTask(ctx, "drain", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(492041)
		sendDrainSignal := func(drainRequested bool, err error) (shouldExit bool) {
			__antithesis_instrumentation__.Notify(492043)
			select {
			case ch <- drainSignal{drainRequested: drainRequested, err: err}:
				__antithesis_instrumentation__.Notify(492044)
				return false
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(492045)

				return true
			}
		}
		__antithesis_instrumentation__.Notify(492042)

		for {
			__antithesis_instrumentation__.Notify(492046)
			signal, err := stream.Recv()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(492049)

				sendDrainSignal(false, nil)
				return
			} else {
				__antithesis_instrumentation__.Notify(492050)
			}
			__antithesis_instrumentation__.Notify(492047)
			if err != nil {
				__antithesis_instrumentation__.Notify(492051)
				sendDrainSignal(false, err)
				return
			} else {
				__antithesis_instrumentation__.Notify(492052)
			}
			__antithesis_instrumentation__.Notify(492048)
			switch {
			case signal.DrainRequest != nil:
				__antithesis_instrumentation__.Notify(492053)
				if shouldExit := sendDrainSignal(true, nil); shouldExit {
					__antithesis_instrumentation__.Notify(492056)
					return
				} else {
					__antithesis_instrumentation__.Notify(492057)
				}
			case signal.Handshake != nil:
				__antithesis_instrumentation__.Notify(492054)
				log.Eventf(ctx, "consumer sent handshake.\nConsuming flow scheduled: %t",
					signal.Handshake.ConsumerScheduled)
			default:
				__antithesis_instrumentation__.Notify(492055)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(492058)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492059)
	}
	__antithesis_instrumentation__.Notify(492040)
	return ch, nil
}

func (m *Outbox) run(ctx context.Context, wg *sync.WaitGroup) {
	__antithesis_instrumentation__.Notify(492060)
	err := m.mainLoop(ctx)
	if stream, ok := m.stream.(execinfrapb.DistSQL_FlowStreamClient); ok {
		__antithesis_instrumentation__.Notify(492062)
		closeErr := stream.CloseSend()
		if err == nil {
			__antithesis_instrumentation__.Notify(492063)
			err = closeErr
		} else {
			__antithesis_instrumentation__.Notify(492064)
		}
	} else {
		__antithesis_instrumentation__.Notify(492065)
	}
	__antithesis_instrumentation__.Notify(492061)
	m.err = err
	if wg != nil {
		__antithesis_instrumentation__.Notify(492066)
		wg.Done()
	} else {
		__antithesis_instrumentation__.Notify(492067)
	}
}

func (m *Outbox) Start(ctx context.Context, wg *sync.WaitGroup, flowCtxCancel context.CancelFunc) {
	__antithesis_instrumentation__.Notify(492068)
	if m.OutputTypes() == nil {
		__antithesis_instrumentation__.Notify(492071)
		panic("outbox not initialized")
	} else {
		__antithesis_instrumentation__.Notify(492072)
	}
	__antithesis_instrumentation__.Notify(492069)
	if wg != nil {
		__antithesis_instrumentation__.Notify(492073)
		wg.Add(1)
	} else {
		__antithesis_instrumentation__.Notify(492074)
	}
	__antithesis_instrumentation__.Notify(492070)
	m.flowCtxCancel = flowCtxCancel
	go m.run(ctx, wg)
}

func (m *Outbox) Err() error {
	__antithesis_instrumentation__.Notify(492075)
	return m.err
}
