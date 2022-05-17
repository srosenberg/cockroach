package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streamclient"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var PartitionProgressFrequency = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"streaming.partition_progress_frequency",
	"controls the frequency with which partitions update their progress; if 0, disabled.",
	10*time.Second,
	settings.NonNegativeDuration,
)

const streamIngestionFrontierProcName = `ingestfntr`

type streamIngestionFrontier struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.StreamIngestionFrontierSpec
	alloc   tree.DatumAlloc

	input execinfra.RowSource

	highWaterAtStart hlc.Timestamp

	frontier *span.Frontier

	heartbeatSender *heartbeatSender

	lastPartitionUpdate time.Time
	partitionProgress   map[string]jobspb.StreamIngestionProgress_PartitionProgress
}

var _ execinfra.Processor = &streamIngestionFrontier{}
var _ execinfra.RowSource = &streamIngestionFrontier{}

func init() {
	rowexec.NewStreamIngestionFrontierProcessor = newStreamIngestionFrontierProcessor
}

func newStreamIngestionFrontierProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.StreamIngestionFrontierSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(25184)
	frontier, err := span.MakeFrontier(spec.TrackedSpans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(25188)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25189)
	}
	__antithesis_instrumentation__.Notify(25185)
	heartbeatSender, err := newHeartbeatSender(flowCtx, spec)
	if err != nil {
		__antithesis_instrumentation__.Notify(25190)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25191)
	}
	__antithesis_instrumentation__.Notify(25186)
	sf := &streamIngestionFrontier{
		flowCtx:           flowCtx,
		spec:              spec,
		input:             input,
		highWaterAtStart:  spec.HighWaterAtStart,
		frontier:          frontier,
		partitionProgress: make(map[string]jobspb.StreamIngestionProgress_PartitionProgress),
		heartbeatSender:   heartbeatSender,
	}
	if err := sf.Init(
		sf,
		post,
		input.OutputTypes(),
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{sf.input},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(25192)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25193)
	}
	__antithesis_instrumentation__.Notify(25187)
	return sf, nil
}

func (sf *streamIngestionFrontier) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(25194)
	return true
}

type heartbeatSender struct {
	lastSent        time.Time
	client          streamclient.Client
	streamID        streaming.StreamID
	frontierUpdates chan hlc.Timestamp
	frontier        hlc.Timestamp
	flowCtx         *execinfra.FlowCtx

	cg ctxgroup.Group

	stopChan chan struct{}

	stoppedChan chan struct{}
}

func newHeartbeatSender(
	flowCtx *execinfra.FlowCtx, spec execinfrapb.StreamIngestionFrontierSpec,
) (*heartbeatSender, error) {
	__antithesis_instrumentation__.Notify(25195)
	streamClient, err := streamclient.NewStreamClient(streamingccl.StreamAddress(spec.StreamAddress))
	if err != nil {
		__antithesis_instrumentation__.Notify(25197)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25198)
	}
	__antithesis_instrumentation__.Notify(25196)
	return &heartbeatSender{
		client:          streamClient,
		streamID:        streaming.StreamID(spec.StreamID),
		flowCtx:         flowCtx,
		frontierUpdates: make(chan hlc.Timestamp),
		stopChan:        make(chan struct{}),
		stoppedChan:     make(chan struct{}),
	}, nil
}

func (h *heartbeatSender) maybeHeartbeat(ctx context.Context, frontier hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(25199)
	heartbeatFrequency := streamingccl.StreamReplicationConsumerHeartbeatFrequency.Get(&h.flowCtx.EvalCtx.Settings.SV)
	if h.lastSent.Add(heartbeatFrequency).After(timeutil.Now()) {
		__antithesis_instrumentation__.Notify(25201)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(25202)
	}
	__antithesis_instrumentation__.Notify(25200)
	h.lastSent = timeutil.Now()
	return h.client.Heartbeat(ctx, h.streamID, frontier)
}

func (h *heartbeatSender) startHeartbeatLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(25203)
	h.cg = ctxgroup.WithContext(ctx)
	h.cg.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(25204)
		sendHeartbeats := func() error {
			__antithesis_instrumentation__.Notify(25206)

			timer := time.NewTimer(streamingccl.StreamReplicationConsumerHeartbeatFrequency.
				Get(&h.flowCtx.EvalCtx.Settings.SV))
			defer timer.Stop()
			unknownStatusErr := log.Every(1 * time.Minute)
			for {
				__antithesis_instrumentation__.Notify(25207)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(25212)
					return ctx.Err()
				case <-h.stopChan:
					__antithesis_instrumentation__.Notify(25213)
					return nil
				case <-timer.C:
					__antithesis_instrumentation__.Notify(25214)
					timer.Reset(streamingccl.StreamReplicationConsumerHeartbeatFrequency.
						Get(&h.flowCtx.EvalCtx.Settings.SV))
				case frontier := <-h.frontierUpdates:
					__antithesis_instrumentation__.Notify(25215)
					h.frontier.Forward(frontier)
				}
				__antithesis_instrumentation__.Notify(25208)
				err := h.maybeHeartbeat(ctx, h.frontier)
				if err == nil {
					__antithesis_instrumentation__.Notify(25216)
					continue
				} else {
					__antithesis_instrumentation__.Notify(25217)
				}
				__antithesis_instrumentation__.Notify(25209)

				var se streamingccl.StreamStatusErr
				if !errors.As(err, &se) {
					__antithesis_instrumentation__.Notify(25218)
					return errors.Wrap(err, "unknown stream status error")
				} else {
					__antithesis_instrumentation__.Notify(25219)
				}
				__antithesis_instrumentation__.Notify(25210)

				if se.StreamStatus == streampb.StreamReplicationStatus_UNKNOWN_STREAM_STATUS_RETRY {
					__antithesis_instrumentation__.Notify(25220)
					if unknownStatusErr.ShouldLog() {
						__antithesis_instrumentation__.Notify(25222)
						log.Warningf(ctx, "replication stream %d has unknown status error", se.StreamID)
					} else {
						__antithesis_instrumentation__.Notify(25223)
					}
					__antithesis_instrumentation__.Notify(25221)
					continue
				} else {
					__antithesis_instrumentation__.Notify(25224)
				}
				__antithesis_instrumentation__.Notify(25211)

				return err
			}
		}
		__antithesis_instrumentation__.Notify(25205)
		err := errors.CombineErrors(sendHeartbeats(), h.client.Close())
		close(h.stoppedChan)
		return err
	})
}

func (h *heartbeatSender) stop() error {
	__antithesis_instrumentation__.Notify(25225)
	close(h.stopChan)
	return h.cg.Wait()
}

func (h *heartbeatSender) err() error {
	__antithesis_instrumentation__.Notify(25226)
	return h.cg.Wait()
}

func (sf *streamIngestionFrontier) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(25227)
	ctx = sf.StartInternal(ctx, streamIngestionFrontierProcName)
	sf.input.Start(ctx)
	sf.heartbeatSender.startHeartbeatLoop(ctx)
}

func (sf *streamIngestionFrontier) Next() (
	row rowenc.EncDatumRow,
	meta *execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(25228)
	for sf.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(25230)
		row, meta := sf.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(25235)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(25237)
				sf.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(25238)
			}
			__antithesis_instrumentation__.Notify(25236)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(25239)
		}
		__antithesis_instrumentation__.Notify(25231)
		if row == nil {
			__antithesis_instrumentation__.Notify(25240)
			sf.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(25241)
		}
		__antithesis_instrumentation__.Notify(25232)

		if err := sf.maybeUpdatePartitionProgress(); err != nil {
			__antithesis_instrumentation__.Notify(25242)

			log.Errorf(sf.Ctx, "failed to update partition progress: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(25243)
		}
		__antithesis_instrumentation__.Notify(25233)

		var frontierChanged bool
		var err error
		if frontierChanged, err = sf.noteResolvedTimestamps(row[0]); err != nil {
			__antithesis_instrumentation__.Notify(25244)
			sf.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(25245)
		}
		__antithesis_instrumentation__.Notify(25234)

		newResolvedTS := sf.frontier.Frontier()
		select {
		case <-sf.Ctx.Done():
			__antithesis_instrumentation__.Notify(25246)
			sf.MoveToDraining(sf.Ctx.Err())
			return nil, sf.DrainHelper()

		case sf.heartbeatSender.frontierUpdates <- newResolvedTS:
			__antithesis_instrumentation__.Notify(25247)
			if !frontierChanged {
				__antithesis_instrumentation__.Notify(25251)
				break
			} else {
				__antithesis_instrumentation__.Notify(25252)
			}
			__antithesis_instrumentation__.Notify(25248)
			progressBytes, err := protoutil.Marshal(&newResolvedTS)
			if err != nil {
				__antithesis_instrumentation__.Notify(25253)
				sf.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(25254)
			}
			__antithesis_instrumentation__.Notify(25249)
			pushRow := rowenc.EncDatumRow{
				rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(progressBytes))),
			}
			if outRow := sf.ProcessRowHelper(pushRow); outRow != nil {
				__antithesis_instrumentation__.Notify(25255)
				return outRow, nil
			} else {
				__antithesis_instrumentation__.Notify(25256)
			}

		case <-sf.heartbeatSender.stoppedChan:
			__antithesis_instrumentation__.Notify(25250)
			sf.MoveToDraining(sf.heartbeatSender.err())
			return nil, sf.DrainHelper()
		}
	}
	__antithesis_instrumentation__.Notify(25229)
	return nil, sf.DrainHelper()
}

func (sf *streamIngestionFrontier) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(25257)
	if sf.InternalClose() {
		__antithesis_instrumentation__.Notify(25258)
		if err := sf.heartbeatSender.stop(); err != nil {
			__antithesis_instrumentation__.Notify(25259)
			log.Errorf(sf.Ctx, "heartbeatSender exited with error: %s", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(25260)
		}
	} else {
		__antithesis_instrumentation__.Notify(25261)
	}
}

func (sf *streamIngestionFrontier) noteResolvedTimestamps(
	resolvedSpanDatums rowenc.EncDatum,
) (bool, error) {
	__antithesis_instrumentation__.Notify(25262)
	var frontierChanged bool
	if err := resolvedSpanDatums.EnsureDecoded(streamIngestionResultTypes[0], &sf.alloc); err != nil {
		__antithesis_instrumentation__.Notify(25267)
		return frontierChanged, err
	} else {
		__antithesis_instrumentation__.Notify(25268)
	}
	__antithesis_instrumentation__.Notify(25263)
	raw, ok := resolvedSpanDatums.Datum.(*tree.DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(25269)
		return frontierChanged, errors.AssertionFailedf(`unexpected datum type %T: %s`,
			resolvedSpanDatums.Datum, resolvedSpanDatums.Datum)
	} else {
		__antithesis_instrumentation__.Notify(25270)
	}
	__antithesis_instrumentation__.Notify(25264)
	var resolvedSpans jobspb.ResolvedSpans
	if err := protoutil.Unmarshal([]byte(*raw), &resolvedSpans); err != nil {
		__antithesis_instrumentation__.Notify(25271)
		return frontierChanged, errors.NewAssertionErrorWithWrappedErrf(err,
			`unmarshalling resolved timestamp: %x`, raw)
	} else {
		__antithesis_instrumentation__.Notify(25272)
	}
	__antithesis_instrumentation__.Notify(25265)

	for _, resolved := range resolvedSpans.ResolvedSpans {
		__antithesis_instrumentation__.Notify(25273)

		if !resolved.Timestamp.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(25275)
			return resolved.Timestamp.Less(sf.highWaterAtStart) == true
		}() == true {
			__antithesis_instrumentation__.Notify(25276)
			return frontierChanged, errors.AssertionFailedf(
				`got a resolved timestamp %s that is less than the frontier processor start time %s`,
				redact.Safe(resolved.Timestamp), redact.Safe(sf.highWaterAtStart))
		} else {
			__antithesis_instrumentation__.Notify(25277)
		}
		__antithesis_instrumentation__.Notify(25274)

		if changed, err := sf.frontier.Forward(resolved.Span, resolved.Timestamp); err == nil {
			__antithesis_instrumentation__.Notify(25278)
			frontierChanged = frontierChanged || func() bool {
				__antithesis_instrumentation__.Notify(25279)
				return changed == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(25280)
			return false, err
		}
	}
	__antithesis_instrumentation__.Notify(25266)

	return frontierChanged, nil
}

func (sf *streamIngestionFrontier) maybeUpdatePartitionProgress() error {
	__antithesis_instrumentation__.Notify(25281)
	ctx := sf.Ctx
	updateFreq := PartitionProgressFrequency.Get(&sf.flowCtx.Cfg.Settings.SV)
	if updateFreq == 0 || func() bool {
		__antithesis_instrumentation__.Notify(25285)
		return timeutil.Since(sf.lastPartitionUpdate) < updateFreq == true
	}() == true {
		__antithesis_instrumentation__.Notify(25286)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(25287)
	}
	__antithesis_instrumentation__.Notify(25282)
	registry := sf.flowCtx.Cfg.JobRegistry
	jobID := jobspb.JobID(sf.spec.JobID)
	f := sf.frontier
	allSpans := roachpb.Span{
		Key:    keys.MinKey,
		EndKey: keys.MaxKey,
	}
	partitionFrontiers := sf.partitionProgress
	job, err := registry.LoadJob(ctx, jobID)
	if err != nil {
		__antithesis_instrumentation__.Notify(25288)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25289)
	}
	__antithesis_instrumentation__.Notify(25283)

	f.SpanEntries(allSpans, func(span roachpb.Span, timestamp hlc.Timestamp) (done span.OpResult) {
		__antithesis_instrumentation__.Notify(25290)
		partitionKey := span.Key
		partition := string(partitionKey)
		if curFrontier, ok := partitionFrontiers[partition]; !ok {
			__antithesis_instrumentation__.Notify(25292)
			partitionFrontiers[partition] = jobspb.StreamIngestionProgress_PartitionProgress{
				IngestedTimestamp: timestamp,
			}
		} else {
			__antithesis_instrumentation__.Notify(25293)
			if curFrontier.IngestedTimestamp.Less(timestamp) {
				__antithesis_instrumentation__.Notify(25294)
				curFrontier.IngestedTimestamp = timestamp
			} else {
				__antithesis_instrumentation__.Notify(25295)
			}
		}
		__antithesis_instrumentation__.Notify(25291)
		return true
	})
	__antithesis_instrumentation__.Notify(25284)

	sf.lastPartitionUpdate = timeutil.Now()

	return job.FractionProgressed(ctx, nil,
		func(ctx context.Context, details jobspb.ProgressDetails) float32 {
			__antithesis_instrumentation__.Notify(25296)
			prog := details.(*jobspb.Progress_StreamIngest).StreamIngest
			prog.PartitionProgress = partitionFrontiers

			return 0.0
		},
	)
}
