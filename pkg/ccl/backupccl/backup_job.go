package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/joberror"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
)

var BackupCheckpointInterval = time.Minute

func TestingShortBackupCheckpointInterval(oldInterval time.Duration) func() {
	__antithesis_instrumentation__.Notify(7031)
	BackupCheckpointInterval = time.Millisecond * 10
	return func() {
		__antithesis_instrumentation__.Notify(7032)
		BackupCheckpointInterval = oldInterval
	}
}

var forceReadBackupManifest = util.ConstantWithMetamorphicTestBool("backup-read-manifest", false)

func countRows(raw roachpb.BulkOpSummary, pkIDs map[uint64]bool) roachpb.RowCount {
	__antithesis_instrumentation__.Notify(7033)
	res := roachpb.RowCount{DataSize: raw.DataSize}
	for id, count := range raw.EntryCounts {
		__antithesis_instrumentation__.Notify(7035)
		if _, ok := pkIDs[id]; ok {
			__antithesis_instrumentation__.Notify(7036)
			res.Rows += count
		} else {
			__antithesis_instrumentation__.Notify(7037)
			res.IndexEntries += count
		}
	}
	__antithesis_instrumentation__.Notify(7034)
	return res
}

func filterSpans(includes []roachpb.Span, excludes []roachpb.Span) []roachpb.Span {
	__antithesis_instrumentation__.Notify(7038)
	var cov roachpb.SpanGroup
	cov.Add(includes...)
	cov.Sub(excludes...)
	return cov.Slice()
}

func clusterNodeCount(gw gossip.OptionalGossip) (int, error) {
	__antithesis_instrumentation__.Notify(7039)
	g, err := gw.OptionalErr(47970)
	if err != nil {
		__antithesis_instrumentation__.Notify(7044)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(7045)
	}
	__antithesis_instrumentation__.Notify(7040)
	var nodes int
	err = g.IterateInfos(
		gossip.KeyNodeIDPrefix, func(_ string, _ gossip.Info) error {
			__antithesis_instrumentation__.Notify(7046)
			nodes++
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(7041)
	if err != nil {
		__antithesis_instrumentation__.Notify(7047)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(7048)
	}
	__antithesis_instrumentation__.Notify(7042)

	if nodes == 0 {
		__antithesis_instrumentation__.Notify(7049)
		return 1, errors.New("failed to count nodes")
	} else {
		__antithesis_instrumentation__.Notify(7050)
	}
	__antithesis_instrumentation__.Notify(7043)
	return nodes, nil
}

func backup(
	ctx context.Context,
	execCtx sql.JobExecContext,
	defaultURI string,
	urisByLocalityKV map[string]string,
	db *kv.DB,
	settings *cluster.Settings,
	defaultStore cloud.ExternalStorage,
	storageByLocalityKV map[string]*roachpb.ExternalStorage,
	job *jobs.Job,
	backupManifest *BackupManifest,
	makeExternalStorage cloud.ExternalStorageFactory,
	encryption *jobspb.BackupEncryptionOptions,
	statsCache *stats.TableStatisticsCache,
) (roachpb.RowCount, error) {
	__antithesis_instrumentation__.Notify(7051)

	resumerSpan := tracing.SpanFromContext(ctx)
	var lastCheckpoint time.Time

	var completedSpans, completedIntroducedSpans []roachpb.Span

	for _, file := range backupManifest.Files {
		__antithesis_instrumentation__.Notify(7066)
		if file.StartTime.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(7067)
			return !file.EndTime.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(7068)
			completedIntroducedSpans = append(completedIntroducedSpans, file.Span)
		} else {
			__antithesis_instrumentation__.Notify(7069)
			completedSpans = append(completedSpans, file.Span)
		}
	}
	__antithesis_instrumentation__.Notify(7052)

	spans := filterSpans(backupManifest.Spans, completedSpans)
	introducedSpans := filterSpans(backupManifest.IntroducedSpans, completedIntroducedSpans)

	pkIDs := make(map[uint64]bool)
	for i := range backupManifest.Descriptors {
		__antithesis_instrumentation__.Notify(7070)
		if t, _, _, _ := descpb.FromDescriptor(&backupManifest.Descriptors[i]); t != nil {
			__antithesis_instrumentation__.Notify(7071)
			pkIDs[roachpb.BulkOpSummaryID(uint64(t.ID), uint64(t.PrimaryIndex.ID))] = true
		} else {
			__antithesis_instrumentation__.Notify(7072)
		}
	}
	__antithesis_instrumentation__.Notify(7053)

	evalCtx := execCtx.ExtendedEvalContext()
	dsp := execCtx.DistSQLPlanner()

	planCtx, _, err := dsp.SetupAllNodesPlanning(ctx, evalCtx, execCtx.ExecCfg())
	if err != nil {
		__antithesis_instrumentation__.Notify(7073)
		return roachpb.RowCount{}, errors.Wrap(err, "failed to determine nodes on which to run")
	} else {
		__antithesis_instrumentation__.Notify(7074)
	}
	__antithesis_instrumentation__.Notify(7054)

	backupSpecs, err := distBackupPlanSpecs(
		ctx,
		planCtx,
		execCtx,
		dsp,
		int64(job.ID()),
		spans,
		introducedSpans,
		pkIDs,
		defaultURI,
		urisByLocalityKV,
		encryption,
		roachpb.MVCCFilter(backupManifest.MVCCFilter),
		backupManifest.StartTime,
		backupManifest.EndTime,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(7075)
		return roachpb.RowCount{}, err
	} else {
		__antithesis_instrumentation__.Notify(7076)
	}
	__antithesis_instrumentation__.Notify(7055)

	numTotalSpans := 0
	for _, spec := range backupSpecs {
		__antithesis_instrumentation__.Notify(7077)
		numTotalSpans += len(spec.IntroducedSpans) + len(spec.Spans)
	}
	__antithesis_instrumentation__.Notify(7056)

	progressLogger := jobs.NewChunkProgressLogger(job, numTotalSpans, job.FractionCompleted(), jobs.ProgressUpdateOnly)

	requestFinishedCh := make(chan struct{}, numTotalSpans)
	var jobProgressLoop func(ctx context.Context) error
	if numTotalSpans > 0 {
		__antithesis_instrumentation__.Notify(7078)
		jobProgressLoop = func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(7079)

			return progressLogger.Loop(ctx, requestFinishedCh)
		}
	} else {
		__antithesis_instrumentation__.Notify(7080)
	}
	__antithesis_instrumentation__.Notify(7057)

	progCh := make(chan *execinfrapb.RemoteProducerMetadata_BulkProcessorProgress)
	checkpointLoop := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(7081)

		defer close(requestFinishedCh)
		var numBackedUpFiles int64
		for progress := range progCh {
			__antithesis_instrumentation__.Notify(7083)
			var progDetails BackupManifest_Progress
			if err := types.UnmarshalAny(&progress.ProgressDetails, &progDetails); err != nil {
				__antithesis_instrumentation__.Notify(7088)
				log.Errorf(ctx, "unable to unmarshal backup progress details: %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(7089)
			}
			__antithesis_instrumentation__.Notify(7084)
			if backupManifest.RevisionStartTime.Less(progDetails.RevStartTime) {
				__antithesis_instrumentation__.Notify(7090)
				backupManifest.RevisionStartTime = progDetails.RevStartTime
			} else {
				__antithesis_instrumentation__.Notify(7091)
			}
			__antithesis_instrumentation__.Notify(7085)
			for _, file := range progDetails.Files {
				__antithesis_instrumentation__.Notify(7092)
				backupManifest.Files = append(backupManifest.Files, file)
				backupManifest.EntryCounts.Add(file.EntryCounts)
				numBackedUpFiles++
			}
			__antithesis_instrumentation__.Notify(7086)

			for i := int32(0); i < progDetails.CompletedSpans; i++ {
				__antithesis_instrumentation__.Notify(7093)
				requestFinishedCh <- struct{}{}
			}
			__antithesis_instrumentation__.Notify(7087)
			if timeutil.Since(lastCheckpoint) > BackupCheckpointInterval {
				__antithesis_instrumentation__.Notify(7094)
				resumerSpan.RecordStructured(&BackupProgressTraceEvent{
					TotalNumFiles:     numBackedUpFiles,
					TotalEntryCounts:  backupManifest.EntryCounts,
					RevisionStartTime: backupManifest.RevisionStartTime,
				})

				lastCheckpoint = timeutil.Now()

				err := writeBackupManifestCheckpoint(
					ctx, defaultURI, encryption, backupManifest, execCtx.ExecCfg(), execCtx.User(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(7096)
					log.Errorf(ctx, "unable to checkpoint backup descriptor: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(7097)
				}
				__antithesis_instrumentation__.Notify(7095)

				if execCtx.ExecCfg().TestingKnobs.AfterBackupCheckpoint != nil {
					__antithesis_instrumentation__.Notify(7098)
					execCtx.ExecCfg().TestingKnobs.AfterBackupCheckpoint()
				} else {
					__antithesis_instrumentation__.Notify(7099)
				}
			} else {
				__antithesis_instrumentation__.Notify(7100)
			}
		}
		__antithesis_instrumentation__.Notify(7082)
		return nil
	}
	__antithesis_instrumentation__.Notify(7058)

	resumerSpan.RecordStructured(&types.StringValue{Value: "starting DistSQL backup execution"})
	runBackup := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(7101)
		return distBackup(
			ctx,
			execCtx,
			planCtx,
			dsp,
			progCh,
			backupSpecs,
		)
	}
	__antithesis_instrumentation__.Notify(7059)

	if err := ctxgroup.GoAndWait(ctx, jobProgressLoop, checkpointLoop, runBackup); err != nil {
		__antithesis_instrumentation__.Notify(7102)
		return roachpb.RowCount{}, errors.Wrapf(err, "exporting %d ranges", errors.Safe(numTotalSpans))
	} else {
		__antithesis_instrumentation__.Notify(7103)
	}
	__antithesis_instrumentation__.Notify(7060)

	backupID := uuid.MakeV4()
	backupManifest.ID = backupID

	if len(storageByLocalityKV) > 0 {
		__antithesis_instrumentation__.Notify(7104)
		resumerSpan.RecordStructured(&types.StringValue{Value: "writing partition descriptors for partitioned backup"})
		filesByLocalityKV := make(map[string][]BackupManifest_File)
		for _, file := range backupManifest.Files {
			__antithesis_instrumentation__.Notify(7106)
			filesByLocalityKV[file.LocalityKV] = append(filesByLocalityKV[file.LocalityKV], file)
		}
		__antithesis_instrumentation__.Notify(7105)

		nextPartitionedDescFilenameID := 1
		for kv, conf := range storageByLocalityKV {
			__antithesis_instrumentation__.Notify(7107)
			backupManifest.LocalityKVs = append(backupManifest.LocalityKVs, kv)

			filename := fmt.Sprintf("%s_%d_%s",
				backupPartitionDescriptorPrefix, nextPartitionedDescFilenameID, sanitizeLocalityKV(kv))
			nextPartitionedDescFilenameID++
			backupManifest.PartitionDescriptorFilenames = append(backupManifest.PartitionDescriptorFilenames, filename)
			desc := BackupPartitionDescriptor{
				LocalityKV: kv,
				Files:      filesByLocalityKV[kv],
				BackupID:   backupID,
			}

			if err := func() error {
				__antithesis_instrumentation__.Notify(7108)
				store, err := makeExternalStorage(ctx, *conf)
				if err != nil {
					__antithesis_instrumentation__.Notify(7110)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7111)
				}
				__antithesis_instrumentation__.Notify(7109)
				defer store.Close()
				return writeBackupPartitionDescriptor(ctx, store, filename, encryption, &desc)
			}(); err != nil {
				__antithesis_instrumentation__.Notify(7112)
				return roachpb.RowCount{}, err
			} else {
				__antithesis_instrumentation__.Notify(7113)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(7114)
	}
	__antithesis_instrumentation__.Notify(7061)

	resumerSpan.RecordStructured(&types.StringValue{Value: "writing backup manifest"})
	if err := writeBackupManifest(ctx, settings, defaultStore, backupManifestName, encryption, backupManifest); err != nil {
		__antithesis_instrumentation__.Notify(7115)
		return roachpb.RowCount{}, err
	} else {
		__antithesis_instrumentation__.Notify(7116)
	}
	__antithesis_instrumentation__.Notify(7062)
	var tableStatistics []*stats.TableStatisticProto
	for i := range backupManifest.Descriptors {
		__antithesis_instrumentation__.Notify(7117)
		if tbl, _, _, _ := descpb.FromDescriptor(&backupManifest.Descriptors[i]); tbl != nil {
			__antithesis_instrumentation__.Notify(7118)
			tableDesc := tabledesc.NewBuilder(tbl).BuildImmutableTable()

			tableStatisticsAcc, err := statsCache.GetTableStats(ctx, tableDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(7120)

				log.Warningf(ctx, "failed to collect stats for table: %s, "+
					"table ID: %d during a backup: %s", tableDesc.GetName(), tableDesc.GetID(),
					err.Error())
				continue
			} else {
				__antithesis_instrumentation__.Notify(7121)
			}
			__antithesis_instrumentation__.Notify(7119)
			for _, stat := range tableStatisticsAcc {
				__antithesis_instrumentation__.Notify(7122)
				tableStatistics = append(tableStatistics, &stat.TableStatisticProto)
			}
		} else {
			__antithesis_instrumentation__.Notify(7123)
		}
	}
	__antithesis_instrumentation__.Notify(7063)
	statsTable := StatsTable{
		Statistics: tableStatistics,
	}

	resumerSpan.RecordStructured(&types.StringValue{Value: "writing backup table statistics"})
	if err := writeTableStatistics(ctx, defaultStore, backupStatisticsFileName, encryption, &statsTable); err != nil {
		__antithesis_instrumentation__.Notify(7124)
		return roachpb.RowCount{}, err
	} else {
		__antithesis_instrumentation__.Notify(7125)
	}
	__antithesis_instrumentation__.Notify(7064)

	if writeMetadataSST.Get(&settings.SV) {
		__antithesis_instrumentation__.Notify(7126)
		if err := writeBackupMetadataSST(ctx, defaultStore, encryption, backupManifest, tableStatistics); err != nil {
			__antithesis_instrumentation__.Notify(7127)
			err = errors.Wrap(err, "writing forward-compat metadata sst")
			if !build.IsRelease() {
				__antithesis_instrumentation__.Notify(7129)
				return roachpb.RowCount{}, err
			} else {
				__antithesis_instrumentation__.Notify(7130)
			}
			__antithesis_instrumentation__.Notify(7128)
			log.Warningf(ctx, "%+v", err)
		} else {
			__antithesis_instrumentation__.Notify(7131)
		}
	} else {
		__antithesis_instrumentation__.Notify(7132)
	}
	__antithesis_instrumentation__.Notify(7065)

	return backupManifest.EntryCounts, nil
}

func releaseProtectedTimestamp(
	ctx context.Context, txn *kv.Txn, pts protectedts.Storage, ptsID *uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(7133)

	if ptsID == nil {
		__antithesis_instrumentation__.Notify(7136)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(7137)
	}
	__antithesis_instrumentation__.Notify(7134)
	err := pts.Release(ctx, txn, *ptsID)
	if errors.Is(err, protectedts.ErrNotExists) {
		__antithesis_instrumentation__.Notify(7138)

		log.Warningf(ctx, "failed to release protected which seems not to exist: %v", err)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(7139)
	}
	__antithesis_instrumentation__.Notify(7135)
	return err
}

type backupResumer struct {
	job         *jobs.Job
	backupStats roachpb.RowCount

	testingKnobs struct {
		ignoreProtectedTimestamps bool
	}
}

var _ jobs.TraceableJob = &backupResumer{}

func (b *backupResumer) ForceRealSpan() bool {
	__antithesis_instrumentation__.Notify(7140)
	return true
}

func (b *backupResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(7141)

	resumerSpan := tracing.SpanFromContext(ctx)
	details := b.job.Details().(jobspb.BackupDetails)
	p := execCtx.(sql.JobExecContext)

	var backupManifest *BackupManifest

	if details.URI == "" {
		__antithesis_instrumentation__.Notify(7157)
		initialDetails := details
		backupDetails, m, err := getBackupDetailAndManifest(
			ctx, p.ExecCfg(), p.ExtendedEvalContext().Txn, details, p.User(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(7164)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7165)
		}
		__antithesis_instrumentation__.Notify(7158)
		details = backupDetails

		backupDetails = jobspb.BackupDetails{}
		backupManifest = &m

		{
			__antithesis_instrumentation__.Notify(7166)
			if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.EnableProtectedTimestampsForTenant) {
				__antithesis_instrumentation__.Notify(7168)
				protectedtsID := uuid.MakeV4()
				details.ProtectedTimestampRecord = &protectedtsID
			} else {
				__antithesis_instrumentation__.Notify(7169)
				if len(backupManifest.Spans) > 0 && func() bool {
					__antithesis_instrumentation__.Notify(7170)
					return p.ExecCfg().Codec.ForSystemTenant() == true
				}() == true {
					__antithesis_instrumentation__.Notify(7171)

					protectedtsID := uuid.MakeV4()
					details.ProtectedTimestampRecord = &protectedtsID
				} else {
					__antithesis_instrumentation__.Notify(7172)
				}
			}
			__antithesis_instrumentation__.Notify(7167)

			if details.ProtectedTimestampRecord != nil {
				__antithesis_instrumentation__.Notify(7173)
				if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
					__antithesis_instrumentation__.Notify(7174)
					return protectTimestampForBackup(
						ctx, p.ExecCfg(), txn, b.job.ID(), m, details,
					)
				}); err != nil {
					__antithesis_instrumentation__.Notify(7175)
					return err
				} else {
					__antithesis_instrumentation__.Notify(7176)
				}
			} else {
				__antithesis_instrumentation__.Notify(7177)
			}
		}
		__antithesis_instrumentation__.Notify(7159)

		if err := writeBackupManifestCheckpoint(
			ctx, details.URI, details.EncryptionOptions, backupManifest, p.ExecCfg(), p.User(),
		); err != nil {
			__antithesis_instrumentation__.Notify(7178)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7179)
		}
		__antithesis_instrumentation__.Notify(7160)

		if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(7180)
			return planSchedulePTSChaining(ctx, p.ExecCfg(), txn, &details, b.job.CreatedBy())
		}); err != nil {
			__antithesis_instrumentation__.Notify(7181)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7182)
		}
		__antithesis_instrumentation__.Notify(7161)

		description := b.job.Payload().Description
		const unresolvedText = "INTO 'LATEST' IN"
		if initialDetails.Destination.Subdir == "LATEST" && func() bool {
			__antithesis_instrumentation__.Notify(7183)
			return strings.Count(description, unresolvedText) == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(7184)
			description = strings.ReplaceAll(description, unresolvedText, fmt.Sprintf("INTO '%s' IN", details.Destination.Subdir))
		} else {
			__antithesis_instrumentation__.Notify(7185)
		}
		__antithesis_instrumentation__.Notify(7162)

		if err := b.job.Update(ctx, nil, func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
			__antithesis_instrumentation__.Notify(7186)
			if err := md.CheckRunningOrReverting(); err != nil {
				__antithesis_instrumentation__.Notify(7188)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7189)
			}
			__antithesis_instrumentation__.Notify(7187)
			md.Payload.Details = jobspb.WrapPayloadDetails(details)
			md.Payload.Description = description
			ju.UpdatePayload(md.Payload)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(7190)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7191)
		}
		__antithesis_instrumentation__.Notify(7163)

		lic := utilccl.CheckEnterpriseEnabled(
			p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(), "",
		) != nil
		collectTelemetry(m, details, details, lic)
	} else {
		__antithesis_instrumentation__.Notify(7192)
	}
	__antithesis_instrumentation__.Notify(7142)

	defaultConf, err := cloud.ExternalStorageConfFromURI(details.URI, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(7193)
		return errors.Wrapf(err, "export configuration")
	} else {
		__antithesis_instrumentation__.Notify(7194)
	}
	__antithesis_instrumentation__.Notify(7143)
	defaultStore, err := p.ExecCfg().DistSQLSrv.ExternalStorage(ctx, defaultConf)
	if err != nil {
		__antithesis_instrumentation__.Notify(7195)
		return errors.Wrapf(err, "make storage")
	} else {
		__antithesis_instrumentation__.Notify(7196)
	}
	__antithesis_instrumentation__.Notify(7144)
	defer defaultStore.Close()

	redactedURI := RedactURIForErrorMessage(details.URI)
	if details.EncryptionInfo != nil {
		__antithesis_instrumentation__.Notify(7197)
		if err := writeEncryptionInfoIfNotExists(ctx, details.EncryptionInfo,
			defaultStore); err != nil {
			__antithesis_instrumentation__.Notify(7198)
			return errors.Wrapf(err, "creating encryption info file to %s", redactedURI)
		} else {
			__antithesis_instrumentation__.Notify(7199)
		}
	} else {
		__antithesis_instrumentation__.Notify(7200)
	}
	__antithesis_instrumentation__.Notify(7145)

	storageByLocalityKV := make(map[string]*roachpb.ExternalStorage)
	for kv, uri := range details.URIsByLocalityKV {
		__antithesis_instrumentation__.Notify(7201)
		conf, err := cloud.ExternalStorageConfFromURI(uri, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(7203)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7204)
		}
		__antithesis_instrumentation__.Notify(7202)
		storageByLocalityKV[kv] = &conf
	}
	__antithesis_instrumentation__.Notify(7146)

	mem := p.ExecCfg().RootMemoryMonitor.MakeBoundAccount()
	defer mem.Close(ctx)

	var memSize int64
	defer func() {
		__antithesis_instrumentation__.Notify(7205)
		if memSize != 0 {
			__antithesis_instrumentation__.Notify(7206)
			mem.Shrink(ctx, memSize)
		} else {
			__antithesis_instrumentation__.Notify(7207)
		}
	}()
	__antithesis_instrumentation__.Notify(7147)

	if backupManifest == nil || func() bool {
		__antithesis_instrumentation__.Notify(7208)
		return forceReadBackupManifest == true
	}() == true {
		__antithesis_instrumentation__.Notify(7209)
		backupManifest, memSize, err = b.readManifestOnResume(ctx, &mem, p.ExecCfg(), defaultStore, details, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(7210)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7211)
		}
	} else {
		__antithesis_instrumentation__.Notify(7212)
	}
	__antithesis_instrumentation__.Notify(7148)

	statsCache := p.ExecCfg().TableStatsCache

	retryOpts := retry.Options{
		MaxBackoff: 1 * time.Second,
		MaxRetries: 5,
	}

	if err := p.ExecCfg().JobRegistry.CheckPausepoint("backup.before_flow"); err != nil {
		__antithesis_instrumentation__.Notify(7213)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7214)
	}
	__antithesis_instrumentation__.Notify(7149)

	var res roachpb.RowCount
	var retryCount int32
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(7215)
		retryCount++
		resumerSpan.RecordStructured(&roachpb.RetryTracingEvent{
			Operation:     "backupResumer.Resume",
			AttemptNumber: retryCount,
			RetryError:    tracing.RedactAndTruncateError(err),
		})
		res, err = backup(
			ctx,
			p,
			details.URI,
			details.URIsByLocalityKV,
			p.ExecCfg().DB,
			p.ExecCfg().Settings,
			defaultStore,
			storageByLocalityKV,
			b.job,
			backupManifest,
			p.ExecCfg().DistSQLSrv.ExternalStorage,
			details.EncryptionOptions,
			statsCache,
		)
		if err == nil {
			__antithesis_instrumentation__.Notify(7218)
			break
		} else {
			__antithesis_instrumentation__.Notify(7219)
		}
		__antithesis_instrumentation__.Notify(7216)

		if joberror.IsPermanentBulkJobError(err) {
			__antithesis_instrumentation__.Notify(7220)
			return errors.Wrap(err, "failed to run backup")
		} else {
			__antithesis_instrumentation__.Notify(7221)
		}
		__antithesis_instrumentation__.Notify(7217)

		log.Warningf(ctx, `BACKUP job encountered retryable error: %+v`, err)

		var reloadBackupErr error
		mem.Shrink(ctx, memSize)
		memSize = 0
		backupManifest, memSize, reloadBackupErr = b.readManifestOnResume(ctx, &mem, p.ExecCfg(), defaultStore, details, p.User())
		if reloadBackupErr != nil {
			__antithesis_instrumentation__.Notify(7222)
			return errors.Wrap(reloadBackupErr, "could not reload backup manifest when retrying")
		} else {
			__antithesis_instrumentation__.Notify(7223)
		}
	}
	__antithesis_instrumentation__.Notify(7150)
	if err != nil {
		__antithesis_instrumentation__.Notify(7224)
		return errors.Wrap(err, "exhausted retries")
	} else {
		__antithesis_instrumentation__.Notify(7225)
	}
	__antithesis_instrumentation__.Notify(7151)

	var backupDetails jobspb.BackupDetails
	var ok bool
	if backupDetails, ok = b.job.Details().(jobspb.BackupDetails); !ok {
		__antithesis_instrumentation__.Notify(7226)
		return errors.Newf("unexpected job details type %T", b.job.Details())
	} else {
		__antithesis_instrumentation__.Notify(7227)
	}
	__antithesis_instrumentation__.Notify(7152)

	if err := maybeUpdateSchedulePTSRecord(ctx, p.ExecCfg(), backupDetails, b.job.ID()); err != nil {
		__antithesis_instrumentation__.Notify(7228)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7229)
	}
	__antithesis_instrumentation__.Notify(7153)

	if details.ProtectedTimestampRecord != nil && func() bool {
		__antithesis_instrumentation__.Notify(7230)
		return !b.testingKnobs.ignoreProtectedTimestamps == true
	}() == true {
		__antithesis_instrumentation__.Notify(7231)
		if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(7232)
			details := b.job.Details().(jobspb.BackupDetails)
			return releaseProtectedTimestamp(ctx, txn, p.ExecCfg().ProtectedTimestampProvider,
				details.ProtectedTimestampRecord)
		}); err != nil {
			__antithesis_instrumentation__.Notify(7233)
			log.Errorf(ctx, "failed to release protected timestamp: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(7234)
		}
	} else {
		__antithesis_instrumentation__.Notify(7235)
	}
	__antithesis_instrumentation__.Notify(7154)

	if backupManifest.StartTime.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(7236)
		return details.CollectionURI != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(7237)
		backupURI, err := url.Parse(details.URI)
		if err != nil {
			__antithesis_instrumentation__.Notify(7241)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7242)
		}
		__antithesis_instrumentation__.Notify(7238)
		collectionURI, err := url.Parse(details.CollectionURI)
		if err != nil {
			__antithesis_instrumentation__.Notify(7243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7244)
		}
		__antithesis_instrumentation__.Notify(7239)

		suffix := strings.TrimPrefix(path.Clean(backupURI.Path), path.Clean(collectionURI.Path))

		c, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, details.CollectionURI, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(7245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7246)
		}
		__antithesis_instrumentation__.Notify(7240)
		defer c.Close()

		if err := writeNewLatestFile(ctx, p.ExecCfg().Settings, c, suffix); err != nil {
			__antithesis_instrumentation__.Notify(7247)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7248)
		}
	} else {
		__antithesis_instrumentation__.Notify(7249)
	}
	__antithesis_instrumentation__.Notify(7155)

	b.backupStats = res

	{
		__antithesis_instrumentation__.Notify(7250)
		numClusterNodes, err := clusterNodeCount(p.ExecCfg().Gossip)
		if err != nil {
			__antithesis_instrumentation__.Notify(7253)
			if !build.IsRelease() && func() bool {
				__antithesis_instrumentation__.Notify(7255)
				return p.ExecCfg().Codec.ForSystemTenant() == true
			}() == true {
				__antithesis_instrumentation__.Notify(7256)
				return err
			} else {
				__antithesis_instrumentation__.Notify(7257)
			}
			__antithesis_instrumentation__.Notify(7254)
			log.Warningf(ctx, "unable to determine cluster node count: %v", err)
			numClusterNodes = 1
		} else {
			__antithesis_instrumentation__.Notify(7258)
		}
		__antithesis_instrumentation__.Notify(7251)

		telemetry.Count("backup.total.succeeded")
		const mb = 1 << 20
		sizeMb := res.DataSize / mb
		sec := int64(timeutil.Since(timeutil.FromUnixMicros(b.job.Payload().StartedMicros)).Seconds())
		var mbps int64
		if sec > 0 {
			__antithesis_instrumentation__.Notify(7259)
			mbps = mb / sec
		} else {
			__antithesis_instrumentation__.Notify(7260)
		}
		__antithesis_instrumentation__.Notify(7252)
		if details.StartTime.IsEmpty() {
			__antithesis_instrumentation__.Notify(7261)
			telemetry.CountBucketed("backup.duration-sec.full-succeeded", sec)
			telemetry.CountBucketed("backup.size-mb.full", sizeMb)
			telemetry.CountBucketed("backup.speed-mbps.full.total", mbps)
			telemetry.CountBucketed("backup.speed-mbps.full.per-node", mbps/int64(numClusterNodes))
		} else {
			__antithesis_instrumentation__.Notify(7262)
			telemetry.CountBucketed("backup.duration-sec.inc-succeeded", sec)
			telemetry.CountBucketed("backup.size-mb.inc", sizeMb)
			telemetry.CountBucketed("backup.speed-mbps.inc.total", mbps)
			telemetry.CountBucketed("backup.speed-mbps.inc.per-node", mbps/int64(numClusterNodes))
		}
	}
	__antithesis_instrumentation__.Notify(7156)

	return b.maybeNotifyScheduledJobCompletion(ctx, jobs.StatusSucceeded, p.ExecCfg())
}

func (b *backupResumer) ReportResults(ctx context.Context, resultsCh chan<- tree.Datums) error {
	__antithesis_instrumentation__.Notify(7263)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(7264)
		return ctx.Err()
	case resultsCh <- tree.Datums{
		tree.NewDInt(tree.DInt(b.job.ID())),
		tree.NewDString(string(jobs.StatusSucceeded)),
		tree.NewDFloat(tree.DFloat(1.0)),
		tree.NewDInt(tree.DInt(b.backupStats.Rows)),
		tree.NewDInt(tree.DInt(b.backupStats.IndexEntries)),
		tree.NewDInt(tree.DInt(b.backupStats.DataSize)),
	}:
		__antithesis_instrumentation__.Notify(7265)
		return nil
	}
}

func (b *backupResumer) readManifestOnResume(
	ctx context.Context,
	mem *mon.BoundAccount,
	cfg *sql.ExecutorConfig,
	defaultStore cloud.ExternalStorage,
	details jobspb.BackupDetails,
	user security.SQLUsername,
) (*BackupManifest, int64, error) {
	__antithesis_instrumentation__.Notify(7266)

	desc, memSize, err := readBackupCheckpointManifest(ctx, mem, defaultStore, backupManifestCheckpointName,
		details.EncryptionOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(7269)
		if !errors.Is(err, cloud.ErrFileDoesNotExist) {
			__antithesis_instrumentation__.Notify(7274)
			return nil, 0, errors.Wrapf(err, "reading backup checkpoint")
		} else {
			__antithesis_instrumentation__.Notify(7275)
		}
		__antithesis_instrumentation__.Notify(7270)

		tmpCheckpoint := tempCheckpointFileNameForJob(b.job.ID())
		desc, memSize, err = readBackupCheckpointManifest(ctx, mem, defaultStore, tmpCheckpoint, details.EncryptionOptions)
		if err != nil {
			__antithesis_instrumentation__.Notify(7276)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(7277)
		}
		__antithesis_instrumentation__.Notify(7271)

		if err := writeBackupManifestCheckpoint(
			ctx, details.URI, details.EncryptionOptions, &desc, cfg, user,
		); err != nil {
			__antithesis_instrumentation__.Notify(7278)
			mem.Shrink(ctx, memSize)
			return nil, 0, errors.Wrapf(err, "renaming temp checkpoint file")
		} else {
			__antithesis_instrumentation__.Notify(7279)
		}
		__antithesis_instrumentation__.Notify(7272)

		if err := defaultStore.Delete(ctx, tmpCheckpoint); err != nil {
			__antithesis_instrumentation__.Notify(7280)
			log.Errorf(ctx, "error removing temporary checkpoint %s", tmpCheckpoint)
		} else {
			__antithesis_instrumentation__.Notify(7281)
		}
		__antithesis_instrumentation__.Notify(7273)
		if err := defaultStore.Delete(ctx, backupProgressDirectory+"/"+tmpCheckpoint); err != nil {
			__antithesis_instrumentation__.Notify(7282)
			log.Errorf(ctx, "error removing temporary checkpoint %s", backupProgressDirectory+"/"+tmpCheckpoint)
		} else {
			__antithesis_instrumentation__.Notify(7283)
		}
	} else {
		__antithesis_instrumentation__.Notify(7284)
	}
	__antithesis_instrumentation__.Notify(7267)

	if !desc.ClusterID.Equal(cfg.LogicalClusterID()) {
		__antithesis_instrumentation__.Notify(7285)
		mem.Shrink(ctx, memSize)
		return nil, 0, errors.Newf("cannot resume backup started on another cluster (%s != %s)",
			desc.ClusterID, cfg.LogicalClusterID())
	} else {
		__antithesis_instrumentation__.Notify(7286)
	}
	__antithesis_instrumentation__.Notify(7268)
	return &desc, memSize, nil
}

func (b *backupResumer) maybeNotifyScheduledJobCompletion(
	ctx context.Context, jobStatus jobs.Status, exec *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(7287)
	env := scheduledjobs.ProdJobSchedulerEnv
	if knobs, ok := exec.DistSQLSrv.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(7290)
		if knobs.JobSchedulerEnv != nil {
			__antithesis_instrumentation__.Notify(7291)
			env = knobs.JobSchedulerEnv
		} else {
			__antithesis_instrumentation__.Notify(7292)
		}
	} else {
		__antithesis_instrumentation__.Notify(7293)
	}
	__antithesis_instrumentation__.Notify(7288)

	err := exec.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(7294)

		datums, err := exec.InternalExecutor.QueryRowEx(
			ctx,
			"lookup-schedule-info",
			txn,
			sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
			fmt.Sprintf(
				"SELECT created_by_id FROM %s WHERE id=$1 AND created_by_type=$2",
				env.SystemJobsTableName()),
			b.job.ID(), jobs.CreatedByScheduledJobs)
		if err != nil {
			__antithesis_instrumentation__.Notify(7298)
			return errors.Wrap(err, "schedule info lookup")
		} else {
			__antithesis_instrumentation__.Notify(7299)
		}
		__antithesis_instrumentation__.Notify(7295)
		if datums == nil {
			__antithesis_instrumentation__.Notify(7300)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(7301)
		}
		__antithesis_instrumentation__.Notify(7296)

		scheduleID := int64(tree.MustBeDInt(datums[0]))
		if err := jobs.NotifyJobTermination(
			ctx, env, b.job.ID(), jobStatus, b.job.Details(), scheduleID, exec.InternalExecutor, txn); err != nil {
			__antithesis_instrumentation__.Notify(7302)
			return errors.Wrapf(err,
				"failed to notify schedule %d of completion of job %d", scheduleID, b.job.ID())
		} else {
			__antithesis_instrumentation__.Notify(7303)
		}
		__antithesis_instrumentation__.Notify(7297)
		return nil
	})
	__antithesis_instrumentation__.Notify(7289)
	return err
}

func (b *backupResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(7304)
	telemetry.Count("backup.total.failed")
	telemetry.CountBucketed("backup.duration-sec.failed",
		int64(timeutil.Since(timeutil.FromUnixMicros(b.job.Payload().StartedMicros)).Seconds()))

	p := execCtx.(sql.JobExecContext)
	cfg := p.ExecCfg()
	b.deleteCheckpoint(ctx, cfg, p.User())
	if err := cfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(7307)
		details := b.job.Details().(jobspb.BackupDetails)
		return releaseProtectedTimestamp(ctx, txn, cfg.ProtectedTimestampProvider,
			details.ProtectedTimestampRecord)
	}); err != nil {
		__antithesis_instrumentation__.Notify(7308)
		return err
	} else {
		__antithesis_instrumentation__.Notify(7309)
	}
	__antithesis_instrumentation__.Notify(7305)

	if err := b.maybeNotifyScheduledJobCompletion(ctx, jobs.StatusFailed,
		execCtx.(sql.JobExecContext).ExecCfg()); err != nil {
		__antithesis_instrumentation__.Notify(7310)
		log.Errorf(ctx, "failed to notify job %d on completion of OnFailOrCancel: %+v",
			b.job.ID(), err)
	} else {
		__antithesis_instrumentation__.Notify(7311)
	}
	__antithesis_instrumentation__.Notify(7306)
	return nil
}

func (b *backupResumer) deleteCheckpoint(
	ctx context.Context, cfg *sql.ExecutorConfig, user security.SQLUsername,
) {
	__antithesis_instrumentation__.Notify(7312)

	if err := func() error {
		__antithesis_instrumentation__.Notify(7313)
		details := b.job.Details().(jobspb.BackupDetails)

		exportStore, err := cfg.DistSQLSrv.ExternalStorageFromURI(ctx, details.URI, user)
		if err != nil {
			__antithesis_instrumentation__.Notify(7317)
			return err
		} else {
			__antithesis_instrumentation__.Notify(7318)
		}
		__antithesis_instrumentation__.Notify(7314)
		defer exportStore.Close()

		err = exportStore.Delete(ctx, backupManifestCheckpointName)
		if err != nil {
			__antithesis_instrumentation__.Notify(7319)
			log.Warningf(ctx, "unable to delete checkpointed backup descriptor file in base directory: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(7320)
		}
		__antithesis_instrumentation__.Notify(7315)
		err = exportStore.Delete(ctx, backupManifestCheckpointName+backupManifestChecksumSuffix)
		if err != nil {
			__antithesis_instrumentation__.Notify(7321)
			log.Warningf(ctx, "unable to delete checkpoint checksum file in base directory: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(7322)
		}
		__antithesis_instrumentation__.Notify(7316)

		return exportStore.List(ctx, backupProgressDirectory, "", func(p string) error {
			__antithesis_instrumentation__.Notify(7323)
			return exportStore.Delete(ctx, backupProgressDirectory+p)
		})
	}(); err != nil {
		__antithesis_instrumentation__.Notify(7324)
		log.Warningf(ctx, "unable to delete checkpointed backup descriptor file in progress directory: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(7325)
	}
}

var _ jobs.Resumer = &backupResumer{}

func init() {
	jobs.RegisterConstructor(
		jobspb.TypeBackup,
		func(job *jobs.Job, _ *cluster.Settings) jobs.Resumer {
			return &backupResumer{
				job: job,
			}
		},
	)
}
