package rowflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type rowBasedFlow struct {
	*flowinfra.FlowBase

	localStreams map[execinfrapb.StreamID]execinfra.RowReceiver

	numOutboxes int32
}

var _ flowinfra.Flow = &rowBasedFlow{}

var rowBasedFlowPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(576172)
		return &rowBasedFlow{}
	},
}

func NewRowBasedFlow(base *flowinfra.FlowBase) flowinfra.Flow {
	__antithesis_instrumentation__.Notify(576173)
	rbf := rowBasedFlowPool.Get().(*rowBasedFlow)
	rbf.FlowBase = base
	return rbf
}

func (f *rowBasedFlow) Setup(
	ctx context.Context, spec *execinfrapb.FlowSpec, opt flowinfra.FuseOpt,
) (context.Context, execinfra.OpChains, error) {
	__antithesis_instrumentation__.Notify(576174)
	var err error
	ctx, _, err = f.FlowBase.Setup(ctx, spec, opt)
	if err != nil {
		__antithesis_instrumentation__.Notify(576177)
		return ctx, nil, err
	} else {
		__antithesis_instrumentation__.Notify(576178)
	}
	__antithesis_instrumentation__.Notify(576175)

	inputSyncs, err := f.setupInputSyncs(ctx, spec, opt)
	if err != nil {
		__antithesis_instrumentation__.Notify(576179)
		return ctx, nil, err
	} else {
		__antithesis_instrumentation__.Notify(576180)
	}
	__antithesis_instrumentation__.Notify(576176)

	return ctx, nil, f.setupProcessors(ctx, spec, inputSyncs)
}

func (f *rowBasedFlow) setupProcessors(
	ctx context.Context, spec *execinfrapb.FlowSpec, inputSyncs [][]execinfra.RowSource,
) error {
	__antithesis_instrumentation__.Notify(576181)
	processors := make([]execinfra.Processor, 0, len(spec.Processors))

	for i := range spec.Processors {
		__antithesis_instrumentation__.Notify(576183)
		pspec := &spec.Processors[i]
		p, err := f.makeProcessor(ctx, pspec, inputSyncs[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(576186)
			return err
		} else {
			__antithesis_instrumentation__.Notify(576187)
		}
		__antithesis_instrumentation__.Notify(576184)

		fuse := func() bool {
			__antithesis_instrumentation__.Notify(576188)

			source, ok := p.(execinfra.RowSource)
			if !ok {
				__antithesis_instrumentation__.Notify(576194)
				return false
			} else {
				__antithesis_instrumentation__.Notify(576195)
			}
			__antithesis_instrumentation__.Notify(576189)
			if len(pspec.Output) != 1 {
				__antithesis_instrumentation__.Notify(576196)

				return false
			} else {
				__antithesis_instrumentation__.Notify(576197)
			}
			__antithesis_instrumentation__.Notify(576190)
			ospec := &pspec.Output[0]
			if ospec.Type != execinfrapb.OutputRouterSpec_PASS_THROUGH {
				__antithesis_instrumentation__.Notify(576198)

				return false
			} else {
				__antithesis_instrumentation__.Notify(576199)
			}
			__antithesis_instrumentation__.Notify(576191)
			if len(ospec.Streams) != 1 {
				__antithesis_instrumentation__.Notify(576200)

				return false
			} else {
				__antithesis_instrumentation__.Notify(576201)
			}
			__antithesis_instrumentation__.Notify(576192)

			for pIdx, ps := range spec.Processors {
				__antithesis_instrumentation__.Notify(576202)
				if pIdx <= i {
					__antithesis_instrumentation__.Notify(576204)

					continue
				} else {
					__antithesis_instrumentation__.Notify(576205)
				}
				__antithesis_instrumentation__.Notify(576203)
				for inIdx, in := range ps.Input {
					__antithesis_instrumentation__.Notify(576206)
					if len(in.Streams) == 1 {
						__antithesis_instrumentation__.Notify(576209)
						if in.Streams[0].StreamID != ospec.Streams[0].StreamID {
							__antithesis_instrumentation__.Notify(576211)
							continue
						} else {
							__antithesis_instrumentation__.Notify(576212)
						}
						__antithesis_instrumentation__.Notify(576210)

						inputSyncs[pIdx][inIdx] = source
						return true
					} else {
						__antithesis_instrumentation__.Notify(576213)
					}
					__antithesis_instrumentation__.Notify(576207)

					sync, ok := inputSyncs[pIdx][inIdx].(serialSynchronizer)

					if !ok {
						__antithesis_instrumentation__.Notify(576214)
						continue
					} else {
						__antithesis_instrumentation__.Notify(576215)
					}
					__antithesis_instrumentation__.Notify(576208)

					for sIdx, sspec := range in.Streams {
						__antithesis_instrumentation__.Notify(576216)
						input := findProcByOutputStreamID(spec, sspec.StreamID)
						if input == nil {
							__antithesis_instrumentation__.Notify(576219)
							continue
						} else {
							__antithesis_instrumentation__.Notify(576220)
						}
						__antithesis_instrumentation__.Notify(576217)
						if input.ProcessorID != pspec.ProcessorID {
							__antithesis_instrumentation__.Notify(576221)
							continue
						} else {
							__antithesis_instrumentation__.Notify(576222)
						}
						__antithesis_instrumentation__.Notify(576218)

						sync.getSources()[sIdx].src = source
						return true
					}
				}
			}
			__antithesis_instrumentation__.Notify(576193)
			return false
		}
		__antithesis_instrumentation__.Notify(576185)
		if !fuse() {
			__antithesis_instrumentation__.Notify(576223)
			processors = append(processors, p)
		} else {
			__antithesis_instrumentation__.Notify(576224)
		}
	}
	__antithesis_instrumentation__.Notify(576182)
	f.SetProcessors(processors)
	return nil
}

func findProcByOutputStreamID(
	spec *execinfrapb.FlowSpec, streamID execinfrapb.StreamID,
) *execinfrapb.ProcessorSpec {
	__antithesis_instrumentation__.Notify(576225)
	for i := range spec.Processors {
		__antithesis_instrumentation__.Notify(576227)
		pspec := &spec.Processors[i]
		if len(pspec.Output) > 1 {
			__antithesis_instrumentation__.Notify(576231)

			continue
		} else {
			__antithesis_instrumentation__.Notify(576232)
		}
		__antithesis_instrumentation__.Notify(576228)
		ospec := &pspec.Output[0]
		if ospec.Type != execinfrapb.OutputRouterSpec_PASS_THROUGH {
			__antithesis_instrumentation__.Notify(576233)

			continue
		} else {
			__antithesis_instrumentation__.Notify(576234)
		}
		__antithesis_instrumentation__.Notify(576229)
		if len(ospec.Streams) != 1 {
			__antithesis_instrumentation__.Notify(576235)
			panic(errors.AssertionFailedf("pass-through router with %d streams", len(ospec.Streams)))
		} else {
			__antithesis_instrumentation__.Notify(576236)
		}
		__antithesis_instrumentation__.Notify(576230)
		if ospec.Streams[0].StreamID == streamID {
			__antithesis_instrumentation__.Notify(576237)
			return pspec
		} else {
			__antithesis_instrumentation__.Notify(576238)
		}
	}
	__antithesis_instrumentation__.Notify(576226)
	return nil
}

func (f *rowBasedFlow) makeProcessor(
	ctx context.Context, ps *execinfrapb.ProcessorSpec, inputs []execinfra.RowSource,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(576239)
	if len(ps.Output) != 1 {
		__antithesis_instrumentation__.Notify(576244)
		return nil, errors.Errorf("only single-output processors supported")
	} else {
		__antithesis_instrumentation__.Notify(576245)
	}
	__antithesis_instrumentation__.Notify(576240)
	var output execinfra.RowReceiver
	spec := &ps.Output[0]
	if spec.Type == execinfrapb.OutputRouterSpec_PASS_THROUGH {
		__antithesis_instrumentation__.Notify(576246)

		if len(spec.Streams) != 1 {
			__antithesis_instrumentation__.Notify(576248)
			return nil, errors.Errorf("expected one stream for passthrough router")
		} else {
			__antithesis_instrumentation__.Notify(576249)
		}
		__antithesis_instrumentation__.Notify(576247)
		var err error
		output, err = f.setupOutboundStream(spec.Streams[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(576250)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(576251)
		}
	} else {
		__antithesis_instrumentation__.Notify(576252)
		r, err := f.setupRouter(spec)
		if err != nil {
			__antithesis_instrumentation__.Notify(576254)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(576255)
		}
		__antithesis_instrumentation__.Notify(576253)
		output = r
		f.AddStartable(r)
	}
	__antithesis_instrumentation__.Notify(576241)

	output = &copyingRowReceiver{RowReceiver: output}

	outputs := []execinfra.RowReceiver{output}
	proc, err := rowexec.NewProcessor(
		ctx,
		&f.FlowCtx,
		ps.ProcessorID,
		&ps.Core,
		&ps.Post,
		inputs,
		outputs,
		f.GetLocalProcessors(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(576256)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(576257)
	}
	__antithesis_instrumentation__.Notify(576242)

	types := proc.OutputTypes()
	rowRecv := output.(*copyingRowReceiver).RowReceiver
	switch o := rowRecv.(type) {
	case router:
		__antithesis_instrumentation__.Notify(576258)
		o.init(ctx, &f.FlowCtx, types)
	case *flowinfra.Outbox:
		__antithesis_instrumentation__.Notify(576259)
		o.Init(types)
	}
	__antithesis_instrumentation__.Notify(576243)
	return proc, nil
}

func (f *rowBasedFlow) setupInputSyncs(
	ctx context.Context, spec *execinfrapb.FlowSpec, opt flowinfra.FuseOpt,
) ([][]execinfra.RowSource, error) {
	__antithesis_instrumentation__.Notify(576260)
	inputSyncs := make([][]execinfra.RowSource, len(spec.Processors))
	for pIdx, ps := range spec.Processors {
		__antithesis_instrumentation__.Notify(576262)
		for _, is := range ps.Input {
			__antithesis_instrumentation__.Notify(576263)
			if len(is.Streams) == 0 {
				__antithesis_instrumentation__.Notify(576269)
				return nil, errors.Errorf("input sync with no streams")
			} else {
				__antithesis_instrumentation__.Notify(576270)
			}
			__antithesis_instrumentation__.Notify(576264)
			var sync execinfra.RowSource
			if is.Type != execinfrapb.InputSyncSpec_PARALLEL_UNORDERED && func() bool {
				__antithesis_instrumentation__.Notify(576271)
				return is.Type != execinfrapb.InputSyncSpec_ORDERED == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(576272)
				return is.Type != execinfrapb.InputSyncSpec_SERIAL_UNORDERED == true
			}() == true {
				__antithesis_instrumentation__.Notify(576273)
				return nil, errors.Errorf("unsupported input sync type %s", is.Type)
			} else {
				__antithesis_instrumentation__.Notify(576274)
			}
			__antithesis_instrumentation__.Notify(576265)

			resolver := f.NewTypeResolver(f.EvalCtx.Txn)
			if err := resolver.HydrateTypeSlice(ctx, is.ColumnTypes); err != nil {
				__antithesis_instrumentation__.Notify(576275)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(576276)
			}
			__antithesis_instrumentation__.Notify(576266)

			if is.Type == execinfrapb.InputSyncSpec_PARALLEL_UNORDERED {
				__antithesis_instrumentation__.Notify(576277)
				if opt == flowinfra.FuseNormally || func() bool {
					__antithesis_instrumentation__.Notify(576278)
					return len(is.Streams) == 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(576279)

					mrc := &execinfra.RowChannel{}
					mrc.InitWithNumSenders(is.ColumnTypes, len(is.Streams))
					for _, s := range is.Streams {
						__antithesis_instrumentation__.Notify(576281)
						if err := f.setupInboundStream(ctx, s, mrc); err != nil {
							__antithesis_instrumentation__.Notify(576282)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(576283)
						}
					}
					__antithesis_instrumentation__.Notify(576280)
					sync = mrc
				} else {
					__antithesis_instrumentation__.Notify(576284)
				}
			} else {
				__antithesis_instrumentation__.Notify(576285)
			}
			__antithesis_instrumentation__.Notify(576267)
			if sync == nil {
				__antithesis_instrumentation__.Notify(576286)

				streams := make([]execinfra.RowSource, len(is.Streams))
				for i, s := range is.Streams {
					__antithesis_instrumentation__.Notify(576289)
					rowChan := &execinfra.RowChannel{}
					rowChan.InitWithNumSenders(is.ColumnTypes, 1)
					if err := f.setupInboundStream(ctx, s, rowChan); err != nil {
						__antithesis_instrumentation__.Notify(576291)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(576292)
					}
					__antithesis_instrumentation__.Notify(576290)
					streams[i] = rowChan
				}
				__antithesis_instrumentation__.Notify(576287)
				var err error
				ordering := colinfo.NoOrdering
				if is.Type == execinfrapb.InputSyncSpec_ORDERED {
					__antithesis_instrumentation__.Notify(576293)
					ordering = execinfrapb.ConvertToColumnOrdering(is.Ordering)
				} else {
					__antithesis_instrumentation__.Notify(576294)
				}
				__antithesis_instrumentation__.Notify(576288)
				sync, err = makeSerialSync(ordering, f.EvalCtx, streams)
				if err != nil {
					__antithesis_instrumentation__.Notify(576295)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(576296)
				}
			} else {
				__antithesis_instrumentation__.Notify(576297)
			}
			__antithesis_instrumentation__.Notify(576268)
			inputSyncs[pIdx] = append(inputSyncs[pIdx], sync)
		}
	}
	__antithesis_instrumentation__.Notify(576261)
	return inputSyncs, nil
}

func (f *rowBasedFlow) setupInboundStream(
	ctx context.Context, spec execinfrapb.StreamEndpointSpec, receiver execinfra.RowReceiver,
) error {
	__antithesis_instrumentation__.Notify(576298)
	sid := spec.StreamID
	switch spec.Type {
	case execinfrapb.StreamEndpointSpec_SYNC_RESPONSE:
		__antithesis_instrumentation__.Notify(576300)
		return errors.Errorf("inbound stream of type SYNC_RESPONSE")

	case execinfrapb.StreamEndpointSpec_REMOTE:
		__antithesis_instrumentation__.Notify(576301)
		if err := f.CheckInboundStreamID(sid); err != nil {
			__antithesis_instrumentation__.Notify(576308)
			return err
		} else {
			__antithesis_instrumentation__.Notify(576309)
		}
		__antithesis_instrumentation__.Notify(576302)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(576310)
			log.Infof(ctx, "set up inbound stream %d", sid)
		} else {
			__antithesis_instrumentation__.Notify(576311)
		}
		__antithesis_instrumentation__.Notify(576303)
		f.AddRemoteStream(sid, flowinfra.NewInboundStreamInfo(
			flowinfra.RowInboundStreamHandler{RowReceiver: receiver},
			f.GetWaitGroup(),
		))

	case execinfrapb.StreamEndpointSpec_LOCAL:
		__antithesis_instrumentation__.Notify(576304)
		if _, found := f.localStreams[sid]; found {
			__antithesis_instrumentation__.Notify(576312)
			return errors.Errorf("local stream %d has multiple consumers", sid)
		} else {
			__antithesis_instrumentation__.Notify(576313)
		}
		__antithesis_instrumentation__.Notify(576305)
		if f.localStreams == nil {
			__antithesis_instrumentation__.Notify(576314)
			f.localStreams = make(map[execinfrapb.StreamID]execinfra.RowReceiver)
		} else {
			__antithesis_instrumentation__.Notify(576315)
		}
		__antithesis_instrumentation__.Notify(576306)
		f.localStreams[sid] = receiver

	default:
		__antithesis_instrumentation__.Notify(576307)
		return errors.Errorf("invalid stream type %d", spec.Type)
	}
	__antithesis_instrumentation__.Notify(576299)

	return nil
}

func (f *rowBasedFlow) setupOutboundStream(
	spec execinfrapb.StreamEndpointSpec,
) (execinfra.RowReceiver, error) {
	__antithesis_instrumentation__.Notify(576316)
	sid := spec.StreamID
	switch spec.Type {
	case execinfrapb.StreamEndpointSpec_SYNC_RESPONSE:
		__antithesis_instrumentation__.Notify(576317)
		return f.GetRowSyncFlowConsumer(), nil

	case execinfrapb.StreamEndpointSpec_REMOTE:
		__antithesis_instrumentation__.Notify(576318)
		atomic.AddInt32(&f.numOutboxes, 1)
		outbox := flowinfra.NewOutbox(&f.FlowCtx, spec.TargetNodeID, sid, &f.numOutboxes, f.FlowCtx.Gateway)
		f.AddStartable(outbox)
		return outbox, nil

	case execinfrapb.StreamEndpointSpec_LOCAL:
		__antithesis_instrumentation__.Notify(576319)
		rowChan, found := f.localStreams[sid]
		if !found {
			__antithesis_instrumentation__.Notify(576323)
			return nil, errors.Errorf("unconnected inbound stream %d", sid)
		} else {
			__antithesis_instrumentation__.Notify(576324)
		}
		__antithesis_instrumentation__.Notify(576320)

		if rowChan == nil {
			__antithesis_instrumentation__.Notify(576325)
			return nil, errors.Errorf("stream %d has multiple connections", sid)
		} else {
			__antithesis_instrumentation__.Notify(576326)
		}
		__antithesis_instrumentation__.Notify(576321)
		f.localStreams[sid] = nil
		return rowChan, nil
	default:
		__antithesis_instrumentation__.Notify(576322)
		return nil, errors.Errorf("invalid stream type %d", spec.Type)
	}
}

func (f *rowBasedFlow) setupRouter(spec *execinfrapb.OutputRouterSpec) (router, error) {
	__antithesis_instrumentation__.Notify(576327)
	streams := make([]execinfra.RowReceiver, len(spec.Streams))
	for i := range spec.Streams {
		__antithesis_instrumentation__.Notify(576329)
		var err error
		streams[i], err = f.setupOutboundStream(spec.Streams[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(576330)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(576331)
		}
	}
	__antithesis_instrumentation__.Notify(576328)
	return makeRouter(spec, streams)
}

func (f *rowBasedFlow) IsVectorized() bool {
	__antithesis_instrumentation__.Notify(576332)
	return false
}

func (f *rowBasedFlow) Release() {
	__antithesis_instrumentation__.Notify(576333)
	*f = rowBasedFlow{}
	rowBasedFlowPool.Put(f)
}

func (f *rowBasedFlow) Cleanup(ctx context.Context) {
	__antithesis_instrumentation__.Notify(576334)
	f.FlowBase.Cleanup(ctx)
	f.Release()
}

type copyingRowReceiver struct {
	execinfra.RowReceiver
	alloc rowenc.EncDatumRowAlloc
}

func (r *copyingRowReceiver) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(576335)
	if row != nil {
		__antithesis_instrumentation__.Notify(576337)
		row = r.alloc.CopyRow(row)
	} else {
		__antithesis_instrumentation__.Notify(576338)
	}
	__antithesis_instrumentation__.Notify(576336)
	return r.RowReceiver.Push(row, meta)
}
