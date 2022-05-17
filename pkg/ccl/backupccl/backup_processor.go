package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/storageccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
)

var backupOutputTypes = []*types.T{}

var (
	useTBI = settings.RegisterBoolSetting(
		settings.TenantWritable,
		"kv.bulk_io_write.experimental_incremental_export_enabled",
		"use experimental time-bound file filter when exporting in BACKUP",
		true,
	)
	priorityAfter = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"bulkio.backup.read_with_priority_after",
		"amount of time since the read-as-of time above which a BACKUP should use priority when retrying reads",
		time.Minute,
		settings.NonNegativeDuration,
	).WithPublic()
	delayPerAttmpt = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"bulkio.backup.read_retry_delay",
		"amount of time since the read-as-of time, per-prior attempt, to wait before making another attempt",
		time.Second*5,
		settings.NonNegativeDuration,
	)
	timeoutPerAttempt = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"bulkio.backup.read_timeout",
		"amount of time after which a read attempt is considered timed out, which causes the backup to fail",
		time.Minute*5,
		settings.NonNegativeDuration,
	).WithPublic()
	targetFileSize = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"bulkio.backup.file_size",
		"target size for individual data files produced during BACKUP",
		128<<20,
	).WithPublic()

	defaultSmallFileBuffer = util.ConstantWithMetamorphicTestRange(
		"backup-merge-file-buffer-size",
		128<<20,
		1<<20,
		16<<20,
	)
	smallFileBuffer = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"bulkio.backup.merge_file_buffer_size",
		"size limit used when buffering backup files before merging them",
		int64(defaultSmallFileBuffer),
		settings.NonNegativeInt,
	)

	splitKeysOnTimestamps = settings.RegisterBoolSetting(
		settings.TenantWritable,
		"bulkio.backup.split_keys_on_timestamps",
		"split backup data on timestamps when writing revision history",
		true,
	)
)

const backupProcessorName = "backupDataProcessor"

type backupDataProcessor struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.BackupDataSpec
	output  execinfra.RowReceiver

	cancelAndWaitForWorker func()
	progCh                 chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
	backupErr              error

	memAcc *mon.BoundAccount
}

var (
	_ execinfra.Processor = &backupDataProcessor{}
	_ execinfra.RowSource = &backupDataProcessor{}
)

func newBackupDataProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.BackupDataSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(8663)
	memMonitor := flowCtx.Cfg.BackupMonitor
	if knobs, ok := flowCtx.TestingKnobs().BackupRestoreTestingKnobs.(*sql.BackupRestoreTestingKnobs); ok {
		__antithesis_instrumentation__.Notify(8666)
		if knobs.BackupMemMonitor != nil {
			__antithesis_instrumentation__.Notify(8667)
			memMonitor = knobs.BackupMemMonitor
		} else {
			__antithesis_instrumentation__.Notify(8668)
		}
	} else {
		__antithesis_instrumentation__.Notify(8669)
	}
	__antithesis_instrumentation__.Notify(8664)
	ba := memMonitor.MakeBoundAccount()
	bp := &backupDataProcessor{
		flowCtx: flowCtx,
		spec:    spec,
		output:  output,
		progCh:  make(chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress),
		memAcc:  &ba,
	}
	if err := bp.Init(bp, post, backupOutputTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{

			InputsToDrain: nil,
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(8670)
				bp.close()
				return nil
			},
		}); err != nil {
		__antithesis_instrumentation__.Notify(8671)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(8672)
	}
	__antithesis_instrumentation__.Notify(8665)
	return bp, nil
}

func (bp *backupDataProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(8673)
	ctx = logtags.AddTag(ctx, "job", bp.spec.JobID)
	ctx = bp.StartInternal(ctx, backupProcessorName)
	ctx, cancel := context.WithCancel(ctx)
	bp.cancelAndWaitForWorker = func() {
		__antithesis_instrumentation__.Notify(8675)
		cancel()
		for range bp.progCh {
			__antithesis_instrumentation__.Notify(8676)
		}
	}
	__antithesis_instrumentation__.Notify(8674)
	log.Infof(ctx, "starting backup data")
	if err := bp.flowCtx.Stopper().RunAsyncTaskEx(ctx, stop.TaskOpts{
		TaskName: "backup-worker",
		SpanOpt:  stop.ChildSpan,
	}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(8677)
		bp.backupErr = runBackupProcessor(ctx, bp.flowCtx, &bp.spec, bp.progCh, bp.memAcc)
		cancel()
		close(bp.progCh)
	}); err != nil {
		__antithesis_instrumentation__.Notify(8678)

		bp.backupErr = err
		cancel()
		close(bp.progCh)
	} else {
		__antithesis_instrumentation__.Notify(8679)
	}
}

func (bp *backupDataProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(8680)
	if bp.State != execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(8684)
		return nil, bp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(8685)
	}
	__antithesis_instrumentation__.Notify(8681)

	for prog := range bp.progCh {
		__antithesis_instrumentation__.Notify(8686)

		p := prog
		return nil, &execinfrapb.ProducerMetadata{BulkProcessorProgress: &p}
	}
	__antithesis_instrumentation__.Notify(8682)

	if bp.backupErr != nil {
		__antithesis_instrumentation__.Notify(8687)
		bp.MoveToDraining(bp.backupErr)
		return nil, bp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(8688)
	}
	__antithesis_instrumentation__.Notify(8683)

	bp.MoveToDraining(nil)
	return nil, bp.DrainHelper()
}

func (bp *backupDataProcessor) close() {
	__antithesis_instrumentation__.Notify(8689)
	bp.cancelAndWaitForWorker()
	if bp.InternalClose() {
		__antithesis_instrumentation__.Notify(8690)
		bp.memAcc.Close(bp.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(8691)
	}
}

func (bp *backupDataProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(8692)
	bp.close()
}

type spanAndTime struct {
	spanIdx    int
	span       roachpb.Span
	firstKeyTS hlc.Timestamp
	start, end hlc.Timestamp
	attempts   int
	lastTried  time.Time
}

type returnedSST struct {
	f              BackupManifest_File
	sst            []byte
	revStart       hlc.Timestamp
	completedSpans int32
	atKeyBoundary  bool
}

func runBackupProcessor(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.BackupDataSpec,
	progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
	memAcc *mon.BoundAccount,
) error {
	__antithesis_instrumentation__.Notify(8693)
	backupProcessorSpan := tracing.SpanFromContext(ctx)
	clusterSettings := flowCtx.Cfg.Settings

	totalSpans := len(spec.Spans) + len(spec.IntroducedSpans)
	todo := make(chan spanAndTime, totalSpans)
	var spanIdx int
	for _, s := range spec.IntroducedSpans {
		__antithesis_instrumentation__.Notify(8700)
		todo <- spanAndTime{
			spanIdx: spanIdx, span: s, firstKeyTS: hlc.Timestamp{}, start: hlc.Timestamp{},
			end: spec.BackupStartTime,
		}
		spanIdx++
	}
	__antithesis_instrumentation__.Notify(8694)
	for _, s := range spec.Spans {
		__antithesis_instrumentation__.Notify(8701)
		todo <- spanAndTime{
			spanIdx: spanIdx, span: s, firstKeyTS: hlc.Timestamp{}, start: spec.BackupStartTime,
			end: spec.BackupEndTime,
		}
		spanIdx++
	}
	__antithesis_instrumentation__.Notify(8695)

	destURI := spec.DefaultURI
	var destLocalityKV string

	if len(spec.URIsByLocalityKV) > 0 {
		__antithesis_instrumentation__.Notify(8702)
		var localitySinkURI string

		for i := len(flowCtx.EvalCtx.Locality.Tiers) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(8704)
			tier := flowCtx.EvalCtx.Locality.Tiers[i].String()
			if dest, ok := spec.URIsByLocalityKV[tier]; ok {
				__antithesis_instrumentation__.Notify(8705)
				localitySinkURI = dest
				destLocalityKV = tier
				break
			} else {
				__antithesis_instrumentation__.Notify(8706)
			}
		}
		__antithesis_instrumentation__.Notify(8703)
		if localitySinkURI != "" {
			__antithesis_instrumentation__.Notify(8707)
			log.Infof(ctx, "backing up %d spans to destination specified by locality %s", totalSpans, destLocalityKV)
			destURI = localitySinkURI
		} else {
			__antithesis_instrumentation__.Notify(8708)
			nodeLocalities := make([]string, 0, len(flowCtx.EvalCtx.Locality.Tiers))
			for _, i := range flowCtx.EvalCtx.Locality.Tiers {
				__antithesis_instrumentation__.Notify(8711)
				nodeLocalities = append(nodeLocalities, i.String())
			}
			__antithesis_instrumentation__.Notify(8709)
			backupLocalities := make([]string, 0, len(spec.URIsByLocalityKV))
			for i := range spec.URIsByLocalityKV {
				__antithesis_instrumentation__.Notify(8712)
				backupLocalities = append(backupLocalities, i)
			}
			__antithesis_instrumentation__.Notify(8710)
			log.Infof(ctx, "backing up %d spans to default locality because backup localities %s have no match in node's localities %s", totalSpans, backupLocalities, nodeLocalities)
		}
	} else {
		__antithesis_instrumentation__.Notify(8713)
	}
	__antithesis_instrumentation__.Notify(8696)
	dest, err := cloud.ExternalStorageConfFromURI(destURI, spec.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(8714)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8715)
	}
	__antithesis_instrumentation__.Notify(8697)

	returnedSSTs := make(chan returnedSST, 1)

	grp := ctxgroup.WithContext(ctx)

	grp.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(8716)
		defer close(returnedSSTs)

		numSenders := int(kvserver.ExportRequestsLimit.Get(&clusterSettings.SV)) * 2

		return ctxgroup.GroupWorkers(ctx, numSenders, func(ctx context.Context, _ int) error {
			__antithesis_instrumentation__.Notify(8717)
			readTime := spec.BackupEndTime.GoTime()

			var priority bool
			timer := timeutil.NewTimer()
			defer timer.Stop()

			ctxDone := ctx.Done()
			for {
				__antithesis_instrumentation__.Notify(8718)
				select {
				case <-ctxDone:
					__antithesis_instrumentation__.Notify(8719)
					return ctx.Err()
				case span := <-todo:
					__antithesis_instrumentation__.Notify(8720)
					header := roachpb.Header{Timestamp: span.end}

					splitMidKey := splitKeysOnTimestamps.Get(&clusterSettings.SV)

					if !span.firstKeyTS.IsEmpty() {
						__antithesis_instrumentation__.Notify(8732)
						splitMidKey = true
					} else {
						__antithesis_instrumentation__.Notify(8733)
					}
					__antithesis_instrumentation__.Notify(8721)

					req := &roachpb.ExportRequest{
						RequestHeader:                       roachpb.RequestHeaderFromSpan(span.span),
						ResumeKeyTS:                         span.firstKeyTS,
						StartTime:                           span.start,
						EnableTimeBoundIteratorOptimization: useTBI.Get(&clusterSettings.SV),
						MVCCFilter:                          spec.MVCCFilter,
						TargetFileSize:                      batcheval.ExportRequestTargetFileSize.Get(&clusterSettings.SV),
						ReturnSST:                           true,
						SplitMidKey:                         splitMidKey,
					}

					if !priority && func() bool {
						__antithesis_instrumentation__.Notify(8734)
						return span.attempts > 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(8735)

						if delay := delayPerAttmpt.Get(&clusterSettings.SV) - timeutil.Since(span.lastTried); delay > 0 {
							__antithesis_instrumentation__.Notify(8737)
							timer.Reset(delay)
							log.Infof(ctx, "waiting %s to start attempt %d of remaining spans", delay, span.attempts+1)
							select {
							case <-ctxDone:
								__antithesis_instrumentation__.Notify(8738)
								return ctx.Err()
							case <-timer.C:
								__antithesis_instrumentation__.Notify(8739)
								timer.Read = true
							}
						} else {
							__antithesis_instrumentation__.Notify(8740)
						}
						__antithesis_instrumentation__.Notify(8736)

						priority = timeutil.Since(readTime) > priorityAfter.Get(&clusterSettings.SV)
					} else {
						__antithesis_instrumentation__.Notify(8741)
					}
					__antithesis_instrumentation__.Notify(8722)

					if priority {
						__antithesis_instrumentation__.Notify(8742)

						header.UserPriority = roachpb.MaxUserPriority
					} else {
						__antithesis_instrumentation__.Notify(8743)

						header.WaitPolicy = lock.WaitPolicy_Error
					}
					__antithesis_instrumentation__.Notify(8723)

					header.TargetBytes = 1
					admissionHeader := roachpb.AdmissionHeader{

						Priority:                 int32(admission.BulkNormalPri),
						CreateTime:               timeutil.Now().UnixNano(),
						Source:                   roachpb.AdmissionHeader_FROM_SQL,
						NoMemoryReservedAtSource: true,
					}
					log.Infof(ctx, "sending ExportRequest for span %s (attempt %d, priority %s)",
						span.span, span.attempts+1, header.UserPriority.String())
					var rawRes roachpb.Response
					var pErr *roachpb.Error
					var reqSentTime time.Time
					var respReceivedTime time.Time
					exportRequestErr := contextutil.RunWithTimeout(ctx,
						fmt.Sprintf("ExportRequest for span %s", span.span),
						timeoutPerAttempt.Get(&clusterSettings.SV), func(ctx context.Context) error {
							__antithesis_instrumentation__.Notify(8744)
							reqSentTime = timeutil.Now()
							backupProcessorSpan.RecordStructured(&BackupExportTraceRequestEvent{
								Span:        span.span.String(),
								Attempt:     int32(span.attempts + 1),
								Priority:    header.UserPriority.String(),
								ReqSentTime: reqSentTime.String(),
							})

							rawRes, pErr = kv.SendWrappedWithAdmission(
								ctx, flowCtx.Cfg.DB.NonTransactionalSender(), header, admissionHeader, req)
							respReceivedTime = timeutil.Now()
							if pErr != nil {
								__antithesis_instrumentation__.Notify(8746)
								return pErr.GoError()
							} else {
								__antithesis_instrumentation__.Notify(8747)
							}
							__antithesis_instrumentation__.Notify(8745)
							return nil
						})
					__antithesis_instrumentation__.Notify(8724)
					if exportRequestErr != nil {
						__antithesis_instrumentation__.Notify(8748)
						if intentErr, ok := pErr.GetDetail().(*roachpb.WriteIntentError); ok {
							__antithesis_instrumentation__.Notify(8752)
							span.lastTried = timeutil.Now()
							span.attempts++
							todo <- span

							backupProcessorSpan.RecordStructured(&BackupExportTraceResponseEvent{
								RetryableError: tracing.RedactAndTruncateError(intentErr),
							})
							continue
						} else {
							__antithesis_instrumentation__.Notify(8753)
						}
						__antithesis_instrumentation__.Notify(8749)

						if errors.HasType(exportRequestErr, (*contextutil.TimeoutError)(nil)) {
							__antithesis_instrumentation__.Notify(8754)
							return errors.Wrap(exportRequestErr, "export request timeout")
						} else {
							__antithesis_instrumentation__.Notify(8755)
						}
						__antithesis_instrumentation__.Notify(8750)

						if batchTimestampBeforeGCError, ok := pErr.GetDetail().(*roachpb.BatchTimestampBeforeGCError); ok {
							__antithesis_instrumentation__.Notify(8756)

							if batchTimestampBeforeGCError.DataExcludedFromBackup {
								__antithesis_instrumentation__.Notify(8757)
								continue
							} else {
								__antithesis_instrumentation__.Notify(8758)
							}
						} else {
							__antithesis_instrumentation__.Notify(8759)
						}
						__antithesis_instrumentation__.Notify(8751)
						return errors.Wrapf(exportRequestErr, "exporting %s", span.span)
					} else {
						__antithesis_instrumentation__.Notify(8760)
					}
					__antithesis_instrumentation__.Notify(8725)

					res := rawRes.(*roachpb.ExportResponse)

					if res.ResumeSpan != nil {
						__antithesis_instrumentation__.Notify(8761)
						if !res.ResumeSpan.Valid() {
							__antithesis_instrumentation__.Notify(8764)
							return errors.Errorf("invalid resume span: %s", res.ResumeSpan)
						} else {
							__antithesis_instrumentation__.Notify(8765)
						}
						__antithesis_instrumentation__.Notify(8762)

						resumeTS := hlc.Timestamp{}

						if fileCount := len(res.Files); fileCount > 0 {
							__antithesis_instrumentation__.Notify(8766)
							resumeTS = res.Files[fileCount-1].EndKeyTS
						} else {
							__antithesis_instrumentation__.Notify(8767)
						}
						__antithesis_instrumentation__.Notify(8763)
						resumeSpan := spanAndTime{
							span:       *res.ResumeSpan,
							firstKeyTS: resumeTS,
							start:      span.start,
							end:        span.end,
							attempts:   span.attempts,
							lastTried:  span.lastTried,
						}
						todo <- resumeSpan
					} else {
						__antithesis_instrumentation__.Notify(8768)
					}
					__antithesis_instrumentation__.Notify(8726)

					if backupKnobs, ok := flowCtx.TestingKnobs().BackupRestoreTestingKnobs.(*sql.BackupRestoreTestingKnobs); ok {
						__antithesis_instrumentation__.Notify(8769)
						if backupKnobs.RunAfterExportingSpanEntry != nil {
							__antithesis_instrumentation__.Notify(8770)
							backupKnobs.RunAfterExportingSpanEntry(ctx, res)
						} else {
							__antithesis_instrumentation__.Notify(8771)
						}
					} else {
						__antithesis_instrumentation__.Notify(8772)
					}
					__antithesis_instrumentation__.Notify(8727)

					var completedSpans int32
					if res.ResumeSpan == nil {
						__antithesis_instrumentation__.Notify(8773)
						completedSpans = 1
					} else {
						__antithesis_instrumentation__.Notify(8774)
					}
					__antithesis_instrumentation__.Notify(8728)

					duration := respReceivedTime.Sub(reqSentTime)
					exportResponseTraceEvent := &BackupExportTraceResponseEvent{
						Duration:      duration.String(),
						FileSummaries: make([]roachpb.RowCount, 0),
					}

					if len(res.Files) > 1 {
						__antithesis_instrumentation__.Notify(8775)
						log.Warning(ctx, "unexpected multi-file response using header.TargetBytes = 1")
					} else {
						__antithesis_instrumentation__.Notify(8776)
					}
					__antithesis_instrumentation__.Notify(8729)

					for i, file := range res.Files {
						__antithesis_instrumentation__.Notify(8777)
						f := BackupManifest_File{
							Span:        file.Span,
							Path:        file.Path,
							EntryCounts: countRows(file.Exported, spec.PKIDs),
						}
						exportResponseTraceEvent.FileSummaries = append(exportResponseTraceEvent.FileSummaries, f.EntryCounts)
						if span.start != spec.BackupStartTime {
							__antithesis_instrumentation__.Notify(8780)
							f.StartTime = span.start
							f.EndTime = span.end
						} else {
							__antithesis_instrumentation__.Notify(8781)
						}
						__antithesis_instrumentation__.Notify(8778)
						ret := returnedSST{f: f, sst: file.SST, revStart: res.StartTime, atKeyBoundary: file.EndKeyTS.IsEmpty()}

						if i == len(res.Files)-1 {
							__antithesis_instrumentation__.Notify(8782)
							ret.completedSpans = completedSpans
						} else {
							__antithesis_instrumentation__.Notify(8783)
						}
						__antithesis_instrumentation__.Notify(8779)
						select {
						case returnedSSTs <- ret:
							__antithesis_instrumentation__.Notify(8784)
						case <-ctxDone:
							__antithesis_instrumentation__.Notify(8785)
							return ctx.Err()
						}
					}
					__antithesis_instrumentation__.Notify(8730)
					exportResponseTraceEvent.NumFiles = int32(len(res.Files))
					backupProcessorSpan.RecordStructured(exportResponseTraceEvent)

				default:
					__antithesis_instrumentation__.Notify(8731)

					return nil
				}
			}
		})
	})
	__antithesis_instrumentation__.Notify(8698)

	grp.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(8786)
		sinkConf := sstSinkConf{
			id:       flowCtx.NodeID.SQLInstanceID(),
			enc:      spec.Encryption,
			progCh:   progCh,
			settings: &flowCtx.Cfg.Settings.SV,
		}

		storage, err := flowCtx.Cfg.ExternalStorage(ctx, dest)
		if err != nil {
			__antithesis_instrumentation__.Notify(8791)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8792)
		}
		__antithesis_instrumentation__.Notify(8787)

		sink, err := makeSSTSink(ctx, sinkConf, storage, memAcc)
		if err != nil {
			__antithesis_instrumentation__.Notify(8793)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8794)
		}
		__antithesis_instrumentation__.Notify(8788)

		defer func() {
			__antithesis_instrumentation__.Notify(8795)
			err := sink.Close()
			err = errors.CombineErrors(storage.Close(), err)
			if err != nil {
				__antithesis_instrumentation__.Notify(8796)
				log.Warningf(ctx, "failed to close backup sink(s): % #v", pretty.Formatter(err))
			} else {
				__antithesis_instrumentation__.Notify(8797)
			}
		}()
		__antithesis_instrumentation__.Notify(8789)

		for res := range returnedSSTs {
			__antithesis_instrumentation__.Notify(8798)
			res.f.LocalityKV = destLocalityKV
			if err := sink.push(ctx, res); err != nil {
				__antithesis_instrumentation__.Notify(8799)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8800)
			}
		}
		__antithesis_instrumentation__.Notify(8790)
		return sink.flush(ctx)
	})
	__antithesis_instrumentation__.Notify(8699)

	return grp.Wait()
}

type sstSinkConf struct {
	progCh   chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
	enc      *roachpb.FileEncryptionOptions
	id       base.SQLInstanceID
	settings *settings.Values
}

type sstSink struct {
	dest cloud.ExternalStorage
	conf sstSinkConf

	queue []returnedSST

	queueCap int64

	queueSize int

	sst     storage.SSTWriter
	ctx     context.Context
	cancel  func()
	out     io.WriteCloser
	outName string

	flushedFiles    []BackupManifest_File
	flushedSize     int64
	flushedRevStart hlc.Timestamp
	completedSpans  int32

	stats struct {
		files       int
		flushes     int
		oooFlushes  int
		sizeFlushes int
		spanGrows   int
	}

	memAcc struct {
		ba            *mon.BoundAccount
		reservedBytes int64
	}
}

func makeSSTSink(
	ctx context.Context, conf sstSinkConf, dest cloud.ExternalStorage, backupMem *mon.BoundAccount,
) (*sstSink, error) {
	__antithesis_instrumentation__.Notify(8801)
	s := &sstSink{conf: conf, dest: dest}
	s.memAcc.ba = backupMem

	incrementSize := int64(32 << 20)
	maxSize := smallFileBuffer.Get(s.conf.settings)
	for {
		__antithesis_instrumentation__.Notify(8804)
		if s.queueCap >= maxSize {
			__antithesis_instrumentation__.Notify(8808)
			break
		} else {
			__antithesis_instrumentation__.Notify(8809)
		}
		__antithesis_instrumentation__.Notify(8805)

		if incrementSize > maxSize-s.queueCap {
			__antithesis_instrumentation__.Notify(8810)
			incrementSize = maxSize - s.queueCap
		} else {
			__antithesis_instrumentation__.Notify(8811)
		}
		__antithesis_instrumentation__.Notify(8806)

		if err := s.memAcc.ba.Grow(ctx, incrementSize); err != nil {
			__antithesis_instrumentation__.Notify(8812)
			log.Infof(ctx, "failed to grow file queue by %d bytes, running backup with queue size %d bytes: %+v", incrementSize, s.queueCap, err)
			break
		} else {
			__antithesis_instrumentation__.Notify(8813)
		}
		__antithesis_instrumentation__.Notify(8807)
		s.queueCap += incrementSize
	}
	__antithesis_instrumentation__.Notify(8802)
	if s.queueCap == 0 {
		__antithesis_instrumentation__.Notify(8814)
		return nil, errors.New("failed to reserve memory for sstSink queue")
	} else {
		__antithesis_instrumentation__.Notify(8815)
	}
	__antithesis_instrumentation__.Notify(8803)

	s.memAcc.reservedBytes += s.queueCap
	return s, nil
}

func (s *sstSink) Close() error {
	__antithesis_instrumentation__.Notify(8816)
	if log.V(1) && func() bool {
		__antithesis_instrumentation__.Notify(8820)
		return s.ctx != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(8821)
		log.Infof(s.ctx, "backup sst sink recv'd %d files, wrote %d (%d due to size, %d due to re-ordering), %d recv files extended prior span",
			s.stats.files, s.stats.flushes, s.stats.sizeFlushes, s.stats.oooFlushes, s.stats.spanGrows)
	} else {
		__antithesis_instrumentation__.Notify(8822)
	}
	__antithesis_instrumentation__.Notify(8817)
	if s.cancel != nil {
		__antithesis_instrumentation__.Notify(8823)
		s.cancel()
	} else {
		__antithesis_instrumentation__.Notify(8824)
	}
	__antithesis_instrumentation__.Notify(8818)

	s.memAcc.ba.Shrink(s.ctx, s.memAcc.reservedBytes)
	s.memAcc.reservedBytes = 0
	if s.out != nil {
		__antithesis_instrumentation__.Notify(8825)
		return s.out.Close()
	} else {
		__antithesis_instrumentation__.Notify(8826)
	}
	__antithesis_instrumentation__.Notify(8819)
	return nil
}

func (s *sstSink) push(ctx context.Context, resp returnedSST) error {
	__antithesis_instrumentation__.Notify(8827)
	s.queue = append(s.queue, resp)
	s.queueSize += len(resp.sst)

	if s.queueSize >= int(s.queueCap) {
		__antithesis_instrumentation__.Notify(8829)
		sort.Slice(s.queue, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(8833)
			return s.queue[i].f.Span.Key.Compare(s.queue[j].f.Span.Key) < 0
		})
		__antithesis_instrumentation__.Notify(8830)

		drain := len(s.queue) / 2
		if drain < 1 {
			__antithesis_instrumentation__.Notify(8834)
			drain = 1
		} else {
			__antithesis_instrumentation__.Notify(8835)
		}
		__antithesis_instrumentation__.Notify(8831)
		for i := range s.queue[:drain] {
			__antithesis_instrumentation__.Notify(8836)
			if err := s.write(ctx, s.queue[i]); err != nil {
				__antithesis_instrumentation__.Notify(8838)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8839)
			}
			__antithesis_instrumentation__.Notify(8837)
			s.queueSize -= len(s.queue[i].sst)
		}
		__antithesis_instrumentation__.Notify(8832)

		copy(s.queue, s.queue[drain:])
		s.queue = s.queue[:len(s.queue)-drain]
	} else {
		__antithesis_instrumentation__.Notify(8840)
	}
	__antithesis_instrumentation__.Notify(8828)
	return nil
}

func (s *sstSink) flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(8841)
	for i := range s.queue {
		__antithesis_instrumentation__.Notify(8843)
		if err := s.write(ctx, s.queue[i]); err != nil {
			__antithesis_instrumentation__.Notify(8844)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8845)
		}
	}
	__antithesis_instrumentation__.Notify(8842)
	s.queue = nil
	return s.flushFile(ctx)
}

func (s *sstSink) flushFile(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(8846)
	if s.out == nil {
		__antithesis_instrumentation__.Notify(8852)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8853)
	}
	__antithesis_instrumentation__.Notify(8847)
	s.stats.flushes++

	if err := s.sst.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(8854)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8855)
	}
	__antithesis_instrumentation__.Notify(8848)
	if err := s.out.Close(); err != nil {
		__antithesis_instrumentation__.Notify(8856)
		log.Warningf(ctx, "failed to close write in sstSink: % #v", pretty.Formatter(err))
		return errors.Wrap(err, "writing SST")
	} else {
		__antithesis_instrumentation__.Notify(8857)
	}
	__antithesis_instrumentation__.Notify(8849)
	s.outName = ""
	s.out = nil

	progDetails := BackupManifest_Progress{
		RevStartTime:   s.flushedRevStart,
		Files:          s.flushedFiles,
		CompletedSpans: s.completedSpans,
	}
	var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
	details, err := gogotypes.MarshalAny(&progDetails)
	if err != nil {
		__antithesis_instrumentation__.Notify(8858)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8859)
	}
	__antithesis_instrumentation__.Notify(8850)
	prog.ProgressDetails = *details
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(8860)
		return ctx.Err()
	case s.conf.progCh <- prog:
		__antithesis_instrumentation__.Notify(8861)
	}
	__antithesis_instrumentation__.Notify(8851)

	s.flushedFiles = nil
	s.flushedSize = 0
	s.flushedRevStart.Reset()
	s.completedSpans = 0

	return nil
}

func (s *sstSink) open(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(8862)
	s.outName = generateUniqueSSTName(s.conf.id)
	if s.ctx == nil {
		__antithesis_instrumentation__.Notify(8866)
		s.ctx, s.cancel = context.WithCancel(ctx)
	} else {
		__antithesis_instrumentation__.Notify(8867)
	}
	__antithesis_instrumentation__.Notify(8863)
	w, err := s.dest.Writer(s.ctx, s.outName)
	if err != nil {
		__antithesis_instrumentation__.Notify(8868)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8869)
	}
	__antithesis_instrumentation__.Notify(8864)
	if s.conf.enc != nil {
		__antithesis_instrumentation__.Notify(8870)
		var err error
		w, err = storageccl.EncryptingWriter(w, s.conf.enc.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(8871)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8872)
		}
	} else {
		__antithesis_instrumentation__.Notify(8873)
	}
	__antithesis_instrumentation__.Notify(8865)
	s.out = w
	s.sst = storage.MakeBackupSSTWriter(ctx, s.dest.Settings(), s.out)

	return nil
}

func (s *sstSink) write(ctx context.Context, resp returnedSST) error {
	__antithesis_instrumentation__.Notify(8874)
	s.stats.files++

	span := resp.f.Span

	if len(s.flushedFiles) > 0 {
		__antithesis_instrumentation__.Notify(8881)
		last := s.flushedFiles[len(s.flushedFiles)-1].Span.EndKey
		if span.Key.Compare(last) < 0 {
			__antithesis_instrumentation__.Notify(8882)
			log.VEventf(ctx, 1, "flushing backup file %s of size %d because span %s cannot append before %s",
				s.outName, s.flushedSize, span, last,
			)
			s.stats.oooFlushes++
			if err := s.flushFile(ctx); err != nil {
				__antithesis_instrumentation__.Notify(8883)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8884)
			}
		} else {
			__antithesis_instrumentation__.Notify(8885)
		}
	} else {
		__antithesis_instrumentation__.Notify(8886)
	}
	__antithesis_instrumentation__.Notify(8875)

	if s.out == nil {
		__antithesis_instrumentation__.Notify(8887)
		if err := s.open(ctx); err != nil {
			__antithesis_instrumentation__.Notify(8888)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8889)
		}
	} else {
		__antithesis_instrumentation__.Notify(8890)
	}
	__antithesis_instrumentation__.Notify(8876)

	log.VEventf(ctx, 2, "writing %s to backup file %s", span, s.outName)

	sst, err := storage.NewMemSSTIterator(resp.sst, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(8891)
		return err
	} else {
		__antithesis_instrumentation__.Notify(8892)
	}
	__antithesis_instrumentation__.Notify(8877)
	defer sst.Close()

	sst.SeekGE(storage.MVCCKey{Key: keys.MinKey})
	for {
		__antithesis_instrumentation__.Notify(8893)
		if valid, err := sst.Valid(); !valid || func() bool {
			__antithesis_instrumentation__.Notify(8896)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(8897)
			if err != nil {
				__antithesis_instrumentation__.Notify(8899)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8900)
			}
			__antithesis_instrumentation__.Notify(8898)
			break
		} else {
			__antithesis_instrumentation__.Notify(8901)
		}
		__antithesis_instrumentation__.Notify(8894)
		k := sst.UnsafeKey()
		if k.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(8902)
			if err := s.sst.PutUnversioned(k.Key, sst.UnsafeValue()); err != nil {
				__antithesis_instrumentation__.Notify(8903)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8904)
			}
		} else {
			__antithesis_instrumentation__.Notify(8905)
			if err := s.sst.PutMVCC(sst.UnsafeKey(), sst.UnsafeValue()); err != nil {
				__antithesis_instrumentation__.Notify(8906)
				return err
			} else {
				__antithesis_instrumentation__.Notify(8907)
			}
		}
		__antithesis_instrumentation__.Notify(8895)
		sst.Next()
	}
	__antithesis_instrumentation__.Notify(8878)

	if l := len(s.flushedFiles) - 1; l > 0 && func() bool {
		__antithesis_instrumentation__.Notify(8908)
		return s.flushedFiles[l].Span.EndKey.Equal(span.Key) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(8909)
		return s.flushedFiles[l].EndTime.EqOrdering(resp.f.EndTime) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(8910)
		return s.flushedFiles[l].StartTime.EqOrdering(resp.f.StartTime) == true
	}() == true {
		__antithesis_instrumentation__.Notify(8911)
		s.flushedFiles[l].Span.EndKey = span.EndKey
		s.flushedFiles[l].EntryCounts.Add(resp.f.EntryCounts)
		s.stats.spanGrows++
	} else {
		__antithesis_instrumentation__.Notify(8912)
		f := resp.f
		f.Path = s.outName
		s.flushedFiles = append(s.flushedFiles, f)
	}
	__antithesis_instrumentation__.Notify(8879)
	s.flushedRevStart.Forward(resp.revStart)
	s.completedSpans += resp.completedSpans
	s.flushedSize += int64(len(resp.sst))

	if s.flushedSize > targetFileSize.Get(s.conf.settings) && func() bool {
		__antithesis_instrumentation__.Notify(8913)
		return resp.atKeyBoundary == true
	}() == true {
		__antithesis_instrumentation__.Notify(8914)
		s.stats.sizeFlushes++
		log.VEventf(ctx, 2, "flushing backup file %s with size %d", s.outName, s.flushedSize)
		if err := s.flushFile(ctx); err != nil {
			__antithesis_instrumentation__.Notify(8915)
			return err
		} else {
			__antithesis_instrumentation__.Notify(8916)
		}
	} else {
		__antithesis_instrumentation__.Notify(8917)
		log.VEventf(ctx, 3, "continuing to write to backup file %s of size %d", s.outName, s.flushedSize)
	}
	__antithesis_instrumentation__.Notify(8880)
	return nil
}

func generateUniqueSSTName(nodeID base.SQLInstanceID) string {
	__antithesis_instrumentation__.Notify(8918)

	return fmt.Sprintf("data/%d.sst", builtins.GenerateUniqueInt(nodeID))
}

func init() {
	rowexec.NewBackupDataProcessor = newBackupDataProcessor
}
