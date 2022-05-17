package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streamclient"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv/bulk"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

var minimumFlushInterval = settings.RegisterPublicDurationSettingWithExplicitUnit(
	settings.TenantWritable,
	"bulkio.stream_ingestion.minimum_flush_interval",
	"the minimum timestamp between flushes; flushes may still occur if internal buffers fill up",
	5*time.Second,
	nil,
)

var cutoverSignalPollInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"bulkio.stream_ingestion.cutover_signal_poll_interval",
	"the interval at which the stream ingestion job checks if it has been signaled to cutover",
	30*time.Second,
	settings.NonNegativeDuration,
)

var streamIngestionResultTypes = []*types.T{
	types.Bytes,
}

type mvccKeyValues []storage.MVCCKeyValue

func (s mvccKeyValues) Len() int { __antithesis_instrumentation__.Notify(25426); return len(s) }
func (s mvccKeyValues) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(25427)
	s[i], s[j] = s[j], s[i]
}
func (s mvccKeyValues) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(25428)
	return s[i].Key.Less(s[j].Key)
}

type streamIngestionProcessor struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.StreamIngestionDataSpec
	output  execinfra.RowReceiver

	curBatch mvccKeyValues

	batcher           *bulk.SSTBatcher
	maxFlushRateTimer *timeutil.Timer

	forceClientForTests streamclient.Client

	streamPartitionClients []streamclient.Client

	bufferedCheckpoints map[string]hlc.Timestamp

	lastFlushTime time.Time

	internalDrained bool

	pollingWaitGroup sync.WaitGroup

	eventCh chan partitionEvent

	cutoverCh chan struct{}

	cg ctxgroup.Group

	closePoller chan struct{}

	cancelMergeAndWait func()

	mu struct {
		syncutil.Mutex

		ingestionErr error

		pollingErr error
	}

	metrics *Metrics
}

type partitionEvent struct {
	streamingccl.Event
	partition string
}

var (
	_ execinfra.Processor = &streamIngestionProcessor{}
	_ execinfra.RowSource = &streamIngestionProcessor{}
)

const streamIngestionProcessorName = "stream-ingestion-processor"

func newStreamIngestionDataProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.StreamIngestionDataSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(25429)

	sip := &streamIngestionProcessor{
		flowCtx:             flowCtx,
		spec:                spec,
		output:              output,
		curBatch:            make([]storage.MVCCKeyValue, 0),
		bufferedCheckpoints: make(map[string]hlc.Timestamp),
		maxFlushRateTimer:   timeutil.NewTimer(),
		cutoverCh:           make(chan struct{}),
		closePoller:         make(chan struct{}),
	}

	if err := sip.Init(sip, post, streamIngestionResultTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(25431)
				sip.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(25432)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25433)
	}
	__antithesis_instrumentation__.Notify(25430)

	return sip, nil
}

func (sip *streamIngestionProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(25434)
	ctx = logtags.AddTag(ctx, "job", sip.spec.JobID)
	log.Infof(ctx, "starting ingest proc")
	ctx = sip.StartInternal(ctx, streamIngestionProcessorName)

	sip.metrics = sip.flowCtx.Cfg.JobRegistry.MetricsStruct().StreamIngest.(*Metrics)

	evalCtx := sip.FlowCtx.EvalCtx
	db := sip.FlowCtx.Cfg.DB
	var err error
	sip.batcher, err = bulk.MakeStreamSSTBatcher(ctx, db, evalCtx.Settings, sip.flowCtx.Cfg.BackupMonitor.MakeBoundAccount())
	if err != nil {
		__antithesis_instrumentation__.Notify(25438)
		sip.MoveToDraining(errors.Wrap(err, "creating stream sst batcher"))
		return
	} else {
		__antithesis_instrumentation__.Notify(25439)
	}
	__antithesis_instrumentation__.Notify(25435)

	sip.pollingWaitGroup.Add(1)
	go func() {
		__antithesis_instrumentation__.Notify(25440)
		defer sip.pollingWaitGroup.Done()
		err := sip.checkForCutoverSignal(ctx, sip.closePoller)
		if err != nil {
			__antithesis_instrumentation__.Notify(25441)
			sip.mu.Lock()
			sip.mu.pollingErr = errors.Wrap(err, "error while polling job for cutover signal")
			sip.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(25442)
		}
	}()
	__antithesis_instrumentation__.Notify(25436)

	log.Infof(ctx, "starting %d stream partitions", len(sip.spec.PartitionIds))

	subscriptions := make(map[string]streamclient.Subscription)
	sip.cg = ctxgroup.WithContext(ctx)
	sip.streamPartitionClients = make([]streamclient.Client, 0)
	for i := range sip.spec.PartitionIds {
		__antithesis_instrumentation__.Notify(25443)
		id := sip.spec.PartitionIds[i]
		spec := streamclient.SubscriptionToken(sip.spec.PartitionSpecs[i])
		addr := sip.spec.PartitionAddresses[i]
		var streamClient streamclient.Client
		if sip.forceClientForTests != nil {
			__antithesis_instrumentation__.Notify(25446)
			streamClient = sip.forceClientForTests
			log.Infof(ctx, "using testing client")
		} else {
			__antithesis_instrumentation__.Notify(25447)
			if err != nil {
				__antithesis_instrumentation__.Notify(25450)
				sip.MoveToDraining(errors.Wrapf(err, "creating client for partition spec %q from %q", spec, addr))
				return
			} else {
				__antithesis_instrumentation__.Notify(25451)
			}
			__antithesis_instrumentation__.Notify(25448)
			streamClient, err = streamclient.NewStreamClient(streamingccl.StreamAddress(addr))
			if err != nil {
				__antithesis_instrumentation__.Notify(25452)
				sip.MoveToDraining(errors.Wrapf(err, "creating client for partition spec %q from %q", spec, addr))
				return
			} else {
				__antithesis_instrumentation__.Notify(25453)
			}
			__antithesis_instrumentation__.Notify(25449)
			sip.streamPartitionClients = append(sip.streamPartitionClients, streamClient)
		}
		__antithesis_instrumentation__.Notify(25444)

		sub, err := streamClient.Subscribe(ctx, streaming.StreamID(sip.spec.StreamID), spec, sip.spec.StartTime)
		subscriptions[id] = sub
		if err != nil {
			__antithesis_instrumentation__.Notify(25454)
			sip.MoveToDraining(errors.Wrapf(err, "consuming partition %v", addr))
			return
		} else {
			__antithesis_instrumentation__.Notify(25455)
		}
		__antithesis_instrumentation__.Notify(25445)
		sip.cg.GoCtx(sub.Subscribe)
	}
	__antithesis_instrumentation__.Notify(25437)
	sip.eventCh = sip.merge(ctx, subscriptions)
}

func (sip *streamIngestionProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(25456)
	if sip.State != execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(25462)
		return nil, sip.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(25463)
	}
	__antithesis_instrumentation__.Notify(25457)

	sip.mu.Lock()
	err := sip.mu.pollingErr
	sip.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(25464)
		sip.MoveToDraining(err)
		return nil, sip.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(25465)
	}
	__antithesis_instrumentation__.Notify(25458)

	progressUpdate, err := sip.consumeEvents()
	if err != nil {
		__antithesis_instrumentation__.Notify(25466)
		sip.MoveToDraining(err)
		return nil, sip.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(25467)
	}
	__antithesis_instrumentation__.Notify(25459)

	if progressUpdate != nil {
		__antithesis_instrumentation__.Notify(25468)
		progressBytes, err := protoutil.Marshal(progressUpdate)
		if err != nil {
			__antithesis_instrumentation__.Notify(25470)
			sip.MoveToDraining(err)
			return nil, sip.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(25471)
		}
		__antithesis_instrumentation__.Notify(25469)
		row := rowenc.EncDatumRow{
			rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(progressBytes))),
		}
		return row, nil
	} else {
		__antithesis_instrumentation__.Notify(25472)
	}
	__antithesis_instrumentation__.Notify(25460)

	sip.mu.Lock()
	err = sip.mu.ingestionErr
	sip.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(25473)
		sip.MoveToDraining(err)
		return nil, sip.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(25474)
	}
	__antithesis_instrumentation__.Notify(25461)

	sip.MoveToDraining(nil)
	return nil, sip.DrainHelper()
}

func (sip *streamIngestionProcessor) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(25475)
	return true
}

func (sip *streamIngestionProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(25476)
	sip.close()
}

func (sip *streamIngestionProcessor) close() {
	__antithesis_instrumentation__.Notify(25477)
	if sip.Closed {
		__antithesis_instrumentation__.Notify(25483)
		return
	} else {
		__antithesis_instrumentation__.Notify(25484)
	}
	__antithesis_instrumentation__.Notify(25478)

	for _, client := range sip.streamPartitionClients {
		__antithesis_instrumentation__.Notify(25485)
		_ = client.Close()
	}
	__antithesis_instrumentation__.Notify(25479)
	if sip.batcher != nil {
		__antithesis_instrumentation__.Notify(25486)
		sip.batcher.Close(sip.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(25487)
	}
	__antithesis_instrumentation__.Notify(25480)
	if sip.maxFlushRateTimer != nil {
		__antithesis_instrumentation__.Notify(25488)
		sip.maxFlushRateTimer.Stop()
	} else {
		__antithesis_instrumentation__.Notify(25489)
	}
	__antithesis_instrumentation__.Notify(25481)
	close(sip.closePoller)

	sip.pollingWaitGroup.Wait()

	if sip.cancelMergeAndWait != nil {
		__antithesis_instrumentation__.Notify(25490)
		sip.cancelMergeAndWait()
	} else {
		__antithesis_instrumentation__.Notify(25491)
	}
	__antithesis_instrumentation__.Notify(25482)

	sip.InternalClose()
}

func (sip *streamIngestionProcessor) checkForCutoverSignal(
	ctx context.Context, stopPoller chan struct{},
) error {
	__antithesis_instrumentation__.Notify(25492)
	sv := &sip.flowCtx.Cfg.Settings.SV
	registry := sip.flowCtx.Cfg.JobRegistry
	tick := time.NewTicker(cutoverSignalPollInterval.Get(sv))
	jobID := sip.spec.JobID
	defer tick.Stop()
	for {
		__antithesis_instrumentation__.Notify(25493)
		select {
		case <-stopPoller:
			__antithesis_instrumentation__.Notify(25494)
			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(25495)
			return ctx.Err()
		case <-tick.C:
			__antithesis_instrumentation__.Notify(25496)
			j, err := registry.LoadJob(ctx, jobspb.JobID(jobID))
			if err != nil {
				__antithesis_instrumentation__.Notify(25499)
				return err
			} else {
				__antithesis_instrumentation__.Notify(25500)
			}
			__antithesis_instrumentation__.Notify(25497)
			progress := j.Progress()
			var sp *jobspb.Progress_StreamIngest
			var ok bool
			if sp, ok = progress.GetDetails().(*jobspb.Progress_StreamIngest); !ok {
				__antithesis_instrumentation__.Notify(25501)
				return errors.Newf("unknown progress type %T in stream ingestion job %d",
					j.Progress().Progress, jobID)
			} else {
				__antithesis_instrumentation__.Notify(25502)
			}
			__antithesis_instrumentation__.Notify(25498)

			if !sp.StreamIngest.CutoverTime.IsEmpty() {
				__antithesis_instrumentation__.Notify(25503)

				resolvedTimestamp := progress.GetHighWater()
				if resolvedTimestamp == nil {
					__antithesis_instrumentation__.Notify(25506)
					return errors.AssertionFailedf("cutover has been requested before job %d has had a chance to"+
						" record a resolved ts", jobID)
				} else {
					__antithesis_instrumentation__.Notify(25507)
				}
				__antithesis_instrumentation__.Notify(25504)
				if resolvedTimestamp.Less(sp.StreamIngest.CutoverTime) {
					__antithesis_instrumentation__.Notify(25508)
					return errors.AssertionFailedf("requested cutover time %s is before the resolved time %s recorded"+
						" in job %d", sp.StreamIngest.CutoverTime.String(), resolvedTimestamp.String(),
						jobID)
				} else {
					__antithesis_instrumentation__.Notify(25509)
				}
				__antithesis_instrumentation__.Notify(25505)
				sip.cutoverCh <- struct{}{}
				return nil
			} else {
				__antithesis_instrumentation__.Notify(25510)
			}
		}
	}
}

func (sip *streamIngestionProcessor) merge(
	ctx context.Context, subscriptions map[string]streamclient.Subscription,
) chan partitionEvent {
	__antithesis_instrumentation__.Notify(25511)
	merged := make(chan partitionEvent)

	ctx, cancel := context.WithCancel(ctx)
	g := ctxgroup.WithContext(ctx)

	sip.cancelMergeAndWait = func() {
		__antithesis_instrumentation__.Notify(25515)
		cancel()

		for range merged {
			__antithesis_instrumentation__.Notify(25516)
		}
	}
	__antithesis_instrumentation__.Notify(25512)

	for partition, sub := range subscriptions {
		__antithesis_instrumentation__.Notify(25517)
		partition := partition
		sub := sub
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(25518)
			ctxDone := ctx.Done()
			for {
				__antithesis_instrumentation__.Notify(25519)
				select {
				case event, ok := <-sub.Events():
					__antithesis_instrumentation__.Notify(25520)
					if !ok {
						__antithesis_instrumentation__.Notify(25523)
						return sub.Err()
					} else {
						__antithesis_instrumentation__.Notify(25524)
					}
					__antithesis_instrumentation__.Notify(25521)

					pe := partitionEvent{
						Event:     event,
						partition: partition,
					}

					select {
					case merged <- pe:
						__antithesis_instrumentation__.Notify(25525)
					case <-ctxDone:
						__antithesis_instrumentation__.Notify(25526)
						return ctx.Err()
					}
				case <-ctxDone:
					__antithesis_instrumentation__.Notify(25522)
					return ctx.Err()
				}
			}
		})
	}
	__antithesis_instrumentation__.Notify(25513)
	go func() {
		__antithesis_instrumentation__.Notify(25527)
		err := g.Wait()
		sip.mu.Lock()
		defer sip.mu.Unlock()
		sip.mu.ingestionErr = err
		close(merged)
	}()
	__antithesis_instrumentation__.Notify(25514)

	return merged
}

func (sip *streamIngestionProcessor) consumeEvents() (*jobspb.ResolvedSpans, error) {
	__antithesis_instrumentation__.Notify(25528)

	sv := &sip.FlowCtx.Cfg.Settings.SV

	if sip.internalDrained {
		__antithesis_instrumentation__.Notify(25531)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(25532)
	}
	__antithesis_instrumentation__.Notify(25529)

	for sip.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(25533)
		select {
		case event, ok := <-sip.eventCh:
			__antithesis_instrumentation__.Notify(25534)
			if !ok {
				__antithesis_instrumentation__.Notify(25539)
				sip.internalDrained = true
				return sip.flush()
			} else {
				__antithesis_instrumentation__.Notify(25540)
			}
			__antithesis_instrumentation__.Notify(25535)

			if streamingKnobs, ok := sip.FlowCtx.TestingKnobs().StreamingTestingKnobs.(*sql.StreamingTestingKnobs); ok {
				__antithesis_instrumentation__.Notify(25541)
				if streamingKnobs != nil {
					__antithesis_instrumentation__.Notify(25542)
					if streamingKnobs.RunAfterReceivingEvent != nil {
						__antithesis_instrumentation__.Notify(25543)
						streamingKnobs.RunAfterReceivingEvent(sip.Ctx)
					} else {
						__antithesis_instrumentation__.Notify(25544)
					}
				} else {
					__antithesis_instrumentation__.Notify(25545)
				}
			} else {
				__antithesis_instrumentation__.Notify(25546)
			}
			__antithesis_instrumentation__.Notify(25536)

			switch event.Type() {
			case streamingccl.KVEvent:
				__antithesis_instrumentation__.Notify(25547)
				if err := sip.bufferKV(event); err != nil {
					__antithesis_instrumentation__.Notify(25553)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(25554)
				}
			case streamingccl.CheckpointEvent:
				__antithesis_instrumentation__.Notify(25548)
				if err := sip.bufferCheckpoint(event); err != nil {
					__antithesis_instrumentation__.Notify(25555)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(25556)
				}
				__antithesis_instrumentation__.Notify(25549)

				minFlushInterval := minimumFlushInterval.Get(sv)
				if timeutil.Since(sip.lastFlushTime) < minFlushInterval {
					__antithesis_instrumentation__.Notify(25557)

					sip.maxFlushRateTimer.Reset(time.Until(sip.lastFlushTime.Add(minFlushInterval)))
					continue
				} else {
					__antithesis_instrumentation__.Notify(25558)
				}
				__antithesis_instrumentation__.Notify(25550)

				return sip.flush()
			case streamingccl.GenerationEvent:
				__antithesis_instrumentation__.Notify(25551)
				log.Info(sip.Ctx, "GenerationEvent received")
				select {
				case <-sip.cutoverCh:
					__antithesis_instrumentation__.Notify(25559)
					sip.internalDrained = true
					return nil, nil
				case <-sip.Ctx.Done():
					__antithesis_instrumentation__.Notify(25560)
					return nil, sip.Ctx.Err()
				}
			default:
				__antithesis_instrumentation__.Notify(25552)
				return nil, errors.Newf("unknown streaming event type %v", event.Type())
			}
		case <-sip.cutoverCh:
			__antithesis_instrumentation__.Notify(25537)

			sip.internalDrained = true
			return nil, nil

		case <-sip.maxFlushRateTimer.C:
			__antithesis_instrumentation__.Notify(25538)
			sip.maxFlushRateTimer.Read = true
			return sip.flush()
		}
	}
	__antithesis_instrumentation__.Notify(25530)

	return nil, nil
}

func (sip *streamIngestionProcessor) bufferKV(event partitionEvent) error {
	__antithesis_instrumentation__.Notify(25561)

	kv := event.GetKV()
	if kv == nil {
		__antithesis_instrumentation__.Notify(25563)
		return errors.New("kv event expected to have kv")
	} else {
		__antithesis_instrumentation__.Notify(25564)
	}
	__antithesis_instrumentation__.Notify(25562)
	mvccKey := storage.MVCCKey{
		Key:       kv.Key,
		Timestamp: kv.Value.Timestamp,
	}
	sip.curBatch = append(sip.curBatch, storage.MVCCKeyValue{Key: mvccKey, Value: kv.Value.RawBytes})
	return nil
}

func (sip *streamIngestionProcessor) bufferCheckpoint(event partitionEvent) error {
	__antithesis_instrumentation__.Notify(25565)
	log.Infof(sip.Ctx, "got checkpoint %v", event.GetResolved())
	resolvedTimePtr := event.GetResolved()
	if resolvedTimePtr == nil {
		__antithesis_instrumentation__.Notify(25568)
		return errors.New("checkpoint event expected to have a resolved timestamp")
	} else {
		__antithesis_instrumentation__.Notify(25569)
	}
	__antithesis_instrumentation__.Notify(25566)
	resolvedTime := *resolvedTimePtr

	if lastTimestamp, ok := sip.bufferedCheckpoints[event.partition]; !ok || func() bool {
		__antithesis_instrumentation__.Notify(25570)
		return lastTimestamp.Less(resolvedTime) == true
	}() == true {
		__antithesis_instrumentation__.Notify(25571)
		sip.bufferedCheckpoints[event.partition] = resolvedTime
	} else {
		__antithesis_instrumentation__.Notify(25572)
	}
	__antithesis_instrumentation__.Notify(25567)
	sip.metrics.ResolvedEvents.Inc(1)
	return nil
}

func (sip *streamIngestionProcessor) flush() (*jobspb.ResolvedSpans, error) {
	__antithesis_instrumentation__.Notify(25573)
	flushedCheckpoints := jobspb.ResolvedSpans{ResolvedSpans: make([]jobspb.ResolvedSpan, 0)}

	sort.Sort(sip.curBatch)

	totalSize := 0
	for _, kv := range sip.curBatch {
		__antithesis_instrumentation__.Notify(25577)
		if err := sip.batcher.AddMVCCKey(sip.Ctx, kv.Key, kv.Value); err != nil {
			__antithesis_instrumentation__.Notify(25579)
			return nil, errors.Wrapf(err, "adding key %+v", kv)
		} else {
			__antithesis_instrumentation__.Notify(25580)
		}
		__antithesis_instrumentation__.Notify(25578)
		totalSize += len(kv.Key.Key) + len(kv.Value)
	}
	__antithesis_instrumentation__.Notify(25574)

	if err := sip.batcher.Flush(sip.Ctx); err != nil {
		__antithesis_instrumentation__.Notify(25581)
		return nil, errors.Wrap(err, "flushing")
	} else {
		__antithesis_instrumentation__.Notify(25582)
	}
	__antithesis_instrumentation__.Notify(25575)
	sip.metrics.Flushes.Inc(1)
	sip.metrics.IngestedBytes.Inc(int64(totalSize))
	sip.metrics.IngestedEvents.Inc(int64(len(sip.curBatch)))

	for partition, timestamp := range sip.bufferedCheckpoints {
		__antithesis_instrumentation__.Notify(25583)

		spanStartKey := roachpb.Key(partition)
		resolvedSpan := jobspb.ResolvedSpan{
			Span:      roachpb.Span{Key: spanStartKey, EndKey: spanStartKey.Next()},
			Timestamp: timestamp,
		}
		flushedCheckpoints.ResolvedSpans = append(flushedCheckpoints.ResolvedSpans, resolvedSpan)
	}
	__antithesis_instrumentation__.Notify(25576)

	sip.curBatch = nil
	sip.lastFlushTime = timeutil.Now()
	sip.bufferedCheckpoints = make(map[string]hlc.Timestamp)

	return &flushedCheckpoints, sip.batcher.Reset(sip.Ctx)
}

func init() {
	rowexec.NewStreamIngestionDataProcessor = newStreamIngestionDataProcessor
}
