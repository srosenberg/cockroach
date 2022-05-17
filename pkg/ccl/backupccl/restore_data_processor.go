package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"runtime"

	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
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
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	gogotypes "github.com/gogo/protobuf/types"
)

var restoreDataOutputTypes = []*types.T{}

type restoreDataProcessor struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.RestoreDataSpec
	input   execinfra.RowSource
	output  execinfra.RowReceiver

	numWorkers int

	phaseGroup           ctxgroup.Group
	cancelWorkersAndWait func()

	sstCh chan mergedSST

	metaCh chan *execinfrapb.ProducerMetadata

	progCh chan RestoreProgress
}

var (
	_ execinfra.Processor = &restoreDataProcessor{}
	_ execinfra.RowSource = &restoreDataProcessor{}
)

const restoreDataProcName = "restoreDataProcessor"

const maxConcurrentRestoreWorkers = 32

func min(a, b int) int {
	__antithesis_instrumentation__.Notify(10359)
	if a < b {
		__antithesis_instrumentation__.Notify(10361)
		return a
	} else {
		__antithesis_instrumentation__.Notify(10362)
	}
	__antithesis_instrumentation__.Notify(10360)
	return b
}

var defaultNumWorkers = util.ConstantWithMetamorphicTestRange(
	"restore-worker-concurrency",
	func() int {
		__antithesis_instrumentation__.Notify(10363)

		restoreWorkerCores := runtime.GOMAXPROCS(0) - 1
		if restoreWorkerCores < 1 {
			__antithesis_instrumentation__.Notify(10365)
			restoreWorkerCores = 1
		} else {
			__antithesis_instrumentation__.Notify(10366)
		}
		__antithesis_instrumentation__.Notify(10364)
		return min(4, restoreWorkerCores)
	}(),
	1,
	8,
)

var numRestoreWorkers = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.restore_node_concurrency",
	fmt.Sprintf("the number of workers processing a restore per job per node; maximum %d",
		maxConcurrentRestoreWorkers),
	int64(defaultNumWorkers),
	settings.PositiveInt,
)

var restoreAtNow = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"bulkio.restore_at_current_time.enabled",
	"write restored data at the current timestamp",
	true,
)

func newRestoreDataProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.RestoreDataSpec,
	post *execinfrapb.PostProcessSpec,
	input execinfra.RowSource,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(10367)
	sv := &flowCtx.Cfg.Settings.SV

	if spec.Validation != jobspb.RestoreValidation_DefaultRestore {
		__antithesis_instrumentation__.Notify(10370)
		return nil, errors.New("Restore Data Processor does not support validation yet")
	} else {
		__antithesis_instrumentation__.Notify(10371)
	}
	__antithesis_instrumentation__.Notify(10368)

	rd := &restoreDataProcessor{
		flowCtx:    flowCtx,
		input:      input,
		spec:       spec,
		output:     output,
		progCh:     make(chan RestoreProgress, maxConcurrentRestoreWorkers),
		metaCh:     make(chan *execinfrapb.ProducerMetadata, 1),
		numWorkers: int(numRestoreWorkers.Get(sv)),
	}

	if err := rd.Init(rd, post, restoreDataOutputTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(10372)
				rd.ConsumerClosed()
				return nil
			},
		}); err != nil {
		__antithesis_instrumentation__.Notify(10373)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(10374)
	}
	__antithesis_instrumentation__.Notify(10369)
	return rd, nil
}

func (rd *restoreDataProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(10375)
	ctx = logtags.AddTag(ctx, "job", rd.spec.JobID)
	ctx = rd.StartInternal(ctx, restoreDataProcName)
	rd.input.Start(ctx)

	ctx, cancel := context.WithCancel(ctx)
	rd.cancelWorkersAndWait = func() {
		__antithesis_instrumentation__.Notify(10379)
		cancel()
		_ = rd.phaseGroup.Wait()
	}
	__antithesis_instrumentation__.Notify(10376)
	rd.phaseGroup = ctxgroup.WithContext(ctx)
	log.Infof(ctx, "starting restore data")

	entries := make(chan execinfrapb.RestoreSpanEntry, rd.numWorkers)
	rd.sstCh = make(chan mergedSST, rd.numWorkers)
	rd.phaseGroup.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(10380)
		defer close(entries)
		return inputReader(ctx, rd.input, entries, rd.metaCh)
	})
	__antithesis_instrumentation__.Notify(10377)

	rd.phaseGroup.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(10381)
		defer close(rd.sstCh)
		for entry := range entries {
			__antithesis_instrumentation__.Notify(10383)
			if err := rd.openSSTs(ctx, entry, rd.sstCh); err != nil {
				__antithesis_instrumentation__.Notify(10384)
				return err
			} else {
				__antithesis_instrumentation__.Notify(10385)
			}
		}
		__antithesis_instrumentation__.Notify(10382)

		return nil
	})
	__antithesis_instrumentation__.Notify(10378)

	rd.phaseGroup.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(10386)
		defer close(rd.progCh)
		return rd.runRestoreWorkers(ctx, rd.sstCh)
	})
}

func inputReader(
	ctx context.Context,
	input execinfra.RowSource,
	entries chan execinfrapb.RestoreSpanEntry,
	metaCh chan *execinfrapb.ProducerMetadata,
) error {
	__antithesis_instrumentation__.Notify(10387)
	var alloc tree.DatumAlloc

	for {
		__antithesis_instrumentation__.Notify(10388)

		row, meta := input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(10395)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(10398)
				return meta.Err
			} else {
				__antithesis_instrumentation__.Notify(10399)
			}
			__antithesis_instrumentation__.Notify(10396)

			select {
			case metaCh <- meta:
				__antithesis_instrumentation__.Notify(10400)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(10401)
				return ctx.Err()
			}
			__antithesis_instrumentation__.Notify(10397)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10402)
		}
		__antithesis_instrumentation__.Notify(10389)

		if row == nil {
			__antithesis_instrumentation__.Notify(10403)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(10404)
		}
		__antithesis_instrumentation__.Notify(10390)

		if len(row) != 2 {
			__antithesis_instrumentation__.Notify(10405)
			return errors.New("expected input rows to have exactly 2 columns")
		} else {
			__antithesis_instrumentation__.Notify(10406)
		}
		__antithesis_instrumentation__.Notify(10391)
		if err := row[1].EnsureDecoded(types.Bytes, &alloc); err != nil {
			__antithesis_instrumentation__.Notify(10407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10408)
		}
		__antithesis_instrumentation__.Notify(10392)
		datum := row[1].Datum
		entryDatumBytes, ok := datum.(*tree.DBytes)
		if !ok {
			__antithesis_instrumentation__.Notify(10409)
			return errors.AssertionFailedf(`unexpected datum type %T: %+v`, datum, row)
		} else {
			__antithesis_instrumentation__.Notify(10410)
		}
		__antithesis_instrumentation__.Notify(10393)

		var entry execinfrapb.RestoreSpanEntry
		if err := protoutil.Unmarshal([]byte(*entryDatumBytes), &entry); err != nil {
			__antithesis_instrumentation__.Notify(10411)
			return errors.Wrap(err, "un-marshaling restore span entry")
		} else {
			__antithesis_instrumentation__.Notify(10412)
		}
		__antithesis_instrumentation__.Notify(10394)

		select {
		case entries <- entry:
			__antithesis_instrumentation__.Notify(10413)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(10414)
			return ctx.Err()
		}
	}
}

type mergedSST struct {
	entry   execinfrapb.RestoreSpanEntry
	iter    storage.SimpleMVCCIterator
	cleanup func()
}

func (rd *restoreDataProcessor) openSSTs(
	ctx context.Context, entry execinfrapb.RestoreSpanEntry, sstCh chan mergedSST,
) error {
	__antithesis_instrumentation__.Notify(10415)
	ctxDone := ctx.Done()

	var iters []storage.SimpleMVCCIterator
	var dirs []cloud.ExternalStorage

	defer func() {
		__antithesis_instrumentation__.Notify(10419)
		for _, iter := range iters {
			__antithesis_instrumentation__.Notify(10421)
			iter.Close()
		}
		__antithesis_instrumentation__.Notify(10420)

		for _, dir := range dirs {
			__antithesis_instrumentation__.Notify(10422)
			if err := dir.Close(); err != nil {
				__antithesis_instrumentation__.Notify(10423)
				log.Warningf(ctx, "close export storage failed %v", err)
			} else {
				__antithesis_instrumentation__.Notify(10424)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(10416)

	sendIters := func(itersToSend []storage.SimpleMVCCIterator, dirsToSend []cloud.ExternalStorage) error {
		__antithesis_instrumentation__.Notify(10425)
		multiIter := storage.MakeMultiIterator(itersToSend)

		cleanup := func() {
			__antithesis_instrumentation__.Notify(10428)
			multiIter.Close()
			for _, iter := range itersToSend {
				__antithesis_instrumentation__.Notify(10430)
				iter.Close()
			}
			__antithesis_instrumentation__.Notify(10429)

			for _, dir := range dirsToSend {
				__antithesis_instrumentation__.Notify(10431)
				if err := dir.Close(); err != nil {
					__antithesis_instrumentation__.Notify(10432)
					log.Warningf(ctx, "close export storage failed %v", err)
				} else {
					__antithesis_instrumentation__.Notify(10433)
				}
			}
		}
		__antithesis_instrumentation__.Notify(10426)

		mSST := mergedSST{
			entry:   entry,
			iter:    multiIter,
			cleanup: cleanup,
		}

		select {
		case sstCh <- mSST:
			__antithesis_instrumentation__.Notify(10434)
		case <-ctxDone:
			__antithesis_instrumentation__.Notify(10435)
			return ctx.Err()
		}
		__antithesis_instrumentation__.Notify(10427)

		iters = make([]storage.SimpleMVCCIterator, 0)
		dirs = make([]cloud.ExternalStorage, 0)
		return nil
	}
	__antithesis_instrumentation__.Notify(10417)

	log.VEventf(ctx, 1, "ingesting span [%s-%s)", entry.Span.Key, entry.Span.EndKey)

	for _, file := range entry.Files {
		__antithesis_instrumentation__.Notify(10436)
		log.VEventf(ctx, 2, "import file %s which starts at %s", file.Path, entry.Span.Key)

		dir, err := rd.flowCtx.Cfg.ExternalStorage(ctx, file.Dir)
		if err != nil {
			__antithesis_instrumentation__.Notify(10439)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10440)
		}
		__antithesis_instrumentation__.Notify(10437)
		dirs = append(dirs, dir)

		iter, err := storageccl.ExternalSSTReader(ctx, dir, file.Path, rd.spec.Encryption)
		if err != nil {
			__antithesis_instrumentation__.Notify(10441)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10442)
		}
		__antithesis_instrumentation__.Notify(10438)
		iters = append(iters, iter)
	}
	__antithesis_instrumentation__.Notify(10418)

	return sendIters(iters, dirs)
}

func (rd *restoreDataProcessor) runRestoreWorkers(ctx context.Context, ssts chan mergedSST) error {
	__antithesis_instrumentation__.Notify(10443)
	return ctxgroup.GroupWorkers(ctx, rd.numWorkers, func(ctx context.Context, _ int) error {
		__antithesis_instrumentation__.Notify(10444)
		kr, err := makeKeyRewriterFromRekeys(rd.FlowCtx.Codec(), rd.spec.TableRekeys, rd.spec.TenantRekeys)
		if err != nil {
			__antithesis_instrumentation__.Notify(10446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10447)
		}
		__antithesis_instrumentation__.Notify(10445)

		for {
			__antithesis_instrumentation__.Notify(10448)
			done, err := func() (done bool, _ error) {
				__antithesis_instrumentation__.Notify(10451)
				sstIter, ok := <-ssts
				if !ok {
					__antithesis_instrumentation__.Notify(10455)
					done = true
					return done, nil
				} else {
					__antithesis_instrumentation__.Notify(10456)
				}
				__antithesis_instrumentation__.Notify(10452)

				summary, err := rd.processRestoreSpanEntry(ctx, kr, sstIter)
				if err != nil {
					__antithesis_instrumentation__.Notify(10457)
					return done, err
				} else {
					__antithesis_instrumentation__.Notify(10458)
				}
				__antithesis_instrumentation__.Notify(10453)

				select {
				case rd.progCh <- makeProgressUpdate(summary, sstIter.entry, rd.spec.PKIDs):
					__antithesis_instrumentation__.Notify(10459)
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(10460)
					return done, ctx.Err()
				}
				__antithesis_instrumentation__.Notify(10454)

				return done, nil
			}()
			__antithesis_instrumentation__.Notify(10449)
			if err != nil {
				__antithesis_instrumentation__.Notify(10461)
				return err
			} else {
				__antithesis_instrumentation__.Notify(10462)
			}
			__antithesis_instrumentation__.Notify(10450)

			if done {
				__antithesis_instrumentation__.Notify(10463)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(10464)
			}
		}
	})
}

func (rd *restoreDataProcessor) processRestoreSpanEntry(
	ctx context.Context, kr *KeyRewriter, sst mergedSST,
) (roachpb.BulkOpSummary, error) {
	__antithesis_instrumentation__.Notify(10465)
	db := rd.flowCtx.Cfg.DB
	evalCtx := rd.EvalCtx
	var summary roachpb.BulkOpSummary

	entry := sst.entry
	iter := sst.iter
	defer sst.cleanup()

	writeAtBatchTS := restoreAtNow.Get(&evalCtx.Settings.SV)
	if writeAtBatchTS && func() bool {
		__antithesis_instrumentation__.Notify(10472)
		return !evalCtx.Settings.Version.IsActive(ctx, clusterversion.MVCCAddSSTable) == true
	}() == true {
		__antithesis_instrumentation__.Notify(10473)
		return roachpb.BulkOpSummary{}, errors.Newf(
			"cannot use %s until version %s", restoreAtNow.Key(), clusterversion.MVCCAddSSTable.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(10474)
	}
	__antithesis_instrumentation__.Notify(10466)

	if writeAtBatchTS && func() bool {
		__antithesis_instrumentation__.Notify(10475)
		return kr.fromSystemTenant == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(10476)
		return (bytes.HasPrefix(entry.Span.Key, keys.TenantPrefix) || func() bool {
			__antithesis_instrumentation__.Notify(10477)
			return bytes.HasPrefix(entry.Span.EndKey, keys.TenantPrefix) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(10478)
		log.Warningf(ctx, "restoring span %s at its original timestamps because it is a tenant span", entry.Span)
		writeAtBatchTS = false
	} else {
		__antithesis_instrumentation__.Notify(10479)
	}
	__antithesis_instrumentation__.Notify(10467)

	disallowShadowingBelow := hlc.Timestamp{Logical: 1}
	batcher, err := bulk.MakeSSTBatcher(ctx,
		"restore",
		db,
		evalCtx.Settings,
		disallowShadowingBelow,
		writeAtBatchTS,
		false,
		rd.flowCtx.Cfg.BackupMonitor.MakeBoundAccount(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(10480)
		return summary, err
	} else {
		__antithesis_instrumentation__.Notify(10481)
	}
	__antithesis_instrumentation__.Notify(10468)
	defer batcher.Close(ctx)

	var keyScratch, valueScratch []byte

	startKeyMVCC, endKeyMVCC := storage.MVCCKey{Key: entry.Span.Key},
		storage.MVCCKey{Key: entry.Span.EndKey}

	for iter.SeekGE(startKeyMVCC); ; {
		__antithesis_instrumentation__.Notify(10482)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(10491)
			return summary, err
		} else {
			__antithesis_instrumentation__.Notify(10492)
		}
		__antithesis_instrumentation__.Notify(10483)
		if !ok {
			__antithesis_instrumentation__.Notify(10493)
			break
		} else {
			__antithesis_instrumentation__.Notify(10494)
		}
		__antithesis_instrumentation__.Notify(10484)

		if !rd.spec.RestoreTime.IsEmpty() {
			__antithesis_instrumentation__.Notify(10495)

			if rd.spec.RestoreTime.Less(iter.UnsafeKey().Timestamp) {
				__antithesis_instrumentation__.Notify(10496)
				iter.Next()
				continue
			} else {
				__antithesis_instrumentation__.Notify(10497)
			}
		} else {
			__antithesis_instrumentation__.Notify(10498)
		}
		__antithesis_instrumentation__.Notify(10485)

		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(10499)
			return !iter.UnsafeKey().Less(endKeyMVCC) == true
		}() == true {
			__antithesis_instrumentation__.Notify(10500)
			break
		} else {
			__antithesis_instrumentation__.Notify(10501)
		}
		__antithesis_instrumentation__.Notify(10486)
		if len(iter.UnsafeValue()) == 0 {
			__antithesis_instrumentation__.Notify(10502)

			iter.NextKey()
			continue
		} else {
			__antithesis_instrumentation__.Notify(10503)
		}
		__antithesis_instrumentation__.Notify(10487)

		keyScratch = append(keyScratch[:0], iter.UnsafeKey().Key...)
		valueScratch = append(valueScratch[:0], iter.UnsafeValue()...)
		key := storage.MVCCKey{Key: keyScratch, Timestamp: iter.UnsafeKey().Timestamp}
		value := roachpb.Value{RawBytes: valueScratch}
		iter.NextKey()

		key.Key, ok, err = kr.RewriteKey(key.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(10504)
			return summary, err
		} else {
			__antithesis_instrumentation__.Notify(10505)
		}
		__antithesis_instrumentation__.Notify(10488)
		if !ok {
			__antithesis_instrumentation__.Notify(10506)

			if log.V(5) {
				__antithesis_instrumentation__.Notify(10508)
				log.Infof(ctx, "skipping %s %s", key.Key, value.PrettyPrint())
			} else {
				__antithesis_instrumentation__.Notify(10509)
			}
			__antithesis_instrumentation__.Notify(10507)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10510)
		}
		__antithesis_instrumentation__.Notify(10489)

		value.ClearChecksum()
		value.InitChecksum(key.Key)

		if log.V(5) {
			__antithesis_instrumentation__.Notify(10511)
			log.Infof(ctx, "Put %s -> %s", key.Key, value.PrettyPrint())
		} else {
			__antithesis_instrumentation__.Notify(10512)
		}
		__antithesis_instrumentation__.Notify(10490)
		if err := batcher.AddMVCCKey(ctx, key, value.RawBytes); err != nil {
			__antithesis_instrumentation__.Notify(10513)
			return summary, errors.Wrapf(err, "adding to batch: %s -> %s", key, value.PrettyPrint())
		} else {
			__antithesis_instrumentation__.Notify(10514)
		}
	}
	__antithesis_instrumentation__.Notify(10469)

	if err := batcher.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(10515)
		return summary, err
	} else {
		__antithesis_instrumentation__.Notify(10516)
	}
	__antithesis_instrumentation__.Notify(10470)

	if restoreKnobs, ok := rd.flowCtx.TestingKnobs().BackupRestoreTestingKnobs.(*sql.BackupRestoreTestingKnobs); ok {
		__antithesis_instrumentation__.Notify(10517)
		if restoreKnobs.RunAfterProcessingRestoreSpanEntry != nil {
			__antithesis_instrumentation__.Notify(10518)
			restoreKnobs.RunAfterProcessingRestoreSpanEntry(ctx)
		} else {
			__antithesis_instrumentation__.Notify(10519)
		}
	} else {
		__antithesis_instrumentation__.Notify(10520)
	}
	__antithesis_instrumentation__.Notify(10471)

	return batcher.GetSummary(), nil
}

func makeProgressUpdate(
	summary roachpb.BulkOpSummary, entry execinfrapb.RestoreSpanEntry, pkIDs map[uint64]bool,
) (progDetails RestoreProgress) {
	__antithesis_instrumentation__.Notify(10521)
	progDetails.Summary = countRows(summary, pkIDs)
	progDetails.ProgressIdx = entry.ProgressIdx
	progDetails.DataSpan = entry.Span
	return
}

func (rd *restoreDataProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(10522)
	if rd.State != execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(10525)
		return nil, rd.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(10526)
	}
	__antithesis_instrumentation__.Notify(10523)

	var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress

	select {
	case progDetails, ok := <-rd.progCh:
		__antithesis_instrumentation__.Notify(10527)
		if !ok {
			__antithesis_instrumentation__.Notify(10532)

			err := rd.phaseGroup.Wait()
			rd.MoveToDraining(err)
			return nil, rd.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(10533)
		}
		__antithesis_instrumentation__.Notify(10528)

		details, err := gogotypes.MarshalAny(&progDetails)
		if err != nil {
			__antithesis_instrumentation__.Notify(10534)
			rd.MoveToDraining(err)
			return nil, rd.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(10535)
		}
		__antithesis_instrumentation__.Notify(10529)
		prog.ProgressDetails = *details
	case meta := <-rd.metaCh:
		__antithesis_instrumentation__.Notify(10530)
		return nil, meta
	case <-rd.Ctx.Done():
		__antithesis_instrumentation__.Notify(10531)
		rd.MoveToDraining(rd.Ctx.Err())
		return nil, rd.DrainHelper()
	}
	__antithesis_instrumentation__.Notify(10524)

	return nil, &execinfrapb.ProducerMetadata{BulkProcessorProgress: &prog}
}

func (rd *restoreDataProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(10536)
	if rd.Closed {
		__antithesis_instrumentation__.Notify(10539)
		return
	} else {
		__antithesis_instrumentation__.Notify(10540)
	}
	__antithesis_instrumentation__.Notify(10537)
	rd.cancelWorkersAndWait()
	if rd.sstCh != nil {
		__antithesis_instrumentation__.Notify(10541)

		for sst := range rd.sstCh {
			__antithesis_instrumentation__.Notify(10542)
			sst.cleanup()
		}
	} else {
		__antithesis_instrumentation__.Notify(10543)
	}
	__antithesis_instrumentation__.Notify(10538)
	rd.InternalClose()
}

func init() {
	rowexec.NewRestoreDataProcessor = newRestoreDataProcessor
}
