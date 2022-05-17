package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/joberror"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/ingesting"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/rewrite"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbackup"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

var restoreStatsInsertBatchSize = 10

func rewriteBackupSpanKey(
	codec keys.SQLCodec, kr *KeyRewriter, key roachpb.Key,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(10544)
	newKey, rewritten, err := kr.RewriteKey(append([]byte(nil), key...))
	if err != nil {
		__antithesis_instrumentation__.Notify(10549)
		return nil, errors.NewAssertionErrorWithWrappedErrf(err,
			"could not rewrite span start key: %s", key)
	} else {
		__antithesis_instrumentation__.Notify(10550)
	}
	__antithesis_instrumentation__.Notify(10545)
	if !rewritten && func() bool {
		__antithesis_instrumentation__.Notify(10551)
		return bytes.Equal(newKey, key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(10552)

		return nil, errors.AssertionFailedf(
			"no rewrite for span start key: %s", key)
	} else {
		__antithesis_instrumentation__.Notify(10553)
	}
	__antithesis_instrumentation__.Notify(10546)
	if bytes.HasPrefix(newKey, keys.TenantPrefix) {
		__antithesis_instrumentation__.Notify(10554)
		return newKey, nil
	} else {
		__antithesis_instrumentation__.Notify(10555)
	}
	__antithesis_instrumentation__.Notify(10547)

	if b, id, idx, err := codec.DecodeIndexPrefix(newKey); err != nil {
		__antithesis_instrumentation__.Notify(10556)
		return nil, errors.NewAssertionErrorWithWrappedErrf(err,
			"could not rewrite span start key: %s", key)
	} else {
		__antithesis_instrumentation__.Notify(10557)
		if idx == 1 && func() bool {
			__antithesis_instrumentation__.Notify(10558)
			return len(b) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(10559)
			newKey = codec.TablePrefix(id)
		} else {
			__antithesis_instrumentation__.Notify(10560)
		}
	}
	__antithesis_instrumentation__.Notify(10548)
	return newKey, nil
}

func restoreWithRetry(
	restoreCtx context.Context,
	execCtx sql.JobExecContext,
	numNodes int,
	backupManifests []BackupManifest,
	backupLocalityInfo []jobspb.RestoreDetails_BackupLocalityInfo,
	endTime hlc.Timestamp,
	dataToRestore restorationData,
	job *jobs.Job,
	encryption *jobspb.BackupEncryptionOptions,
) (roachpb.RowCount, error) {
	__antithesis_instrumentation__.Notify(10561)

	retryOpts := retry.Options{
		MaxBackoff: 1 * time.Second,
		MaxRetries: 5,
	}

	var res roachpb.RowCount
	var err error
	for r := retry.StartWithCtx(restoreCtx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(10564)
		res, err = restore(
			restoreCtx,
			execCtx,
			numNodes,
			backupManifests,
			backupLocalityInfo,
			endTime,
			dataToRestore,
			job,
			encryption,
		)
		if err == nil {
			__antithesis_instrumentation__.Notify(10568)
			break
		} else {
			__antithesis_instrumentation__.Notify(10569)
		}
		__antithesis_instrumentation__.Notify(10565)

		if errors.HasType(err, &roachpb.InsufficientSpaceError{}) {
			__antithesis_instrumentation__.Notify(10570)
			return roachpb.RowCount{}, jobs.MarkPauseRequestError(errors.UnwrapAll(err))
		} else {
			__antithesis_instrumentation__.Notify(10571)
		}
		__antithesis_instrumentation__.Notify(10566)

		if joberror.IsPermanentBulkJobError(err) {
			__antithesis_instrumentation__.Notify(10572)
			return roachpb.RowCount{}, err
		} else {
			__antithesis_instrumentation__.Notify(10573)
		}
		__antithesis_instrumentation__.Notify(10567)

		log.Warningf(restoreCtx, `encountered retryable error: %+v`, err)
	}
	__antithesis_instrumentation__.Notify(10562)

	if err != nil {
		__antithesis_instrumentation__.Notify(10574)
		return roachpb.RowCount{}, errors.Wrap(err, "exhausted retries")
	} else {
		__antithesis_instrumentation__.Notify(10575)
	}
	__antithesis_instrumentation__.Notify(10563)
	return res, nil
}

type storeByLocalityKV map[string]roachpb.ExternalStorage

func makeBackupLocalityMap(
	backupLocalityInfos []jobspb.RestoreDetails_BackupLocalityInfo, user security.SQLUsername,
) (map[int]storeByLocalityKV, error) {
	__antithesis_instrumentation__.Notify(10576)

	backupLocalityMap := make(map[int]storeByLocalityKV)
	for i, localityInfo := range backupLocalityInfos {
		__antithesis_instrumentation__.Notify(10578)
		storesByLocalityKV := make(storeByLocalityKV)
		if localityInfo.URIsByOriginalLocalityKV != nil {
			__antithesis_instrumentation__.Notify(10580)
			for kv, uri := range localityInfo.URIsByOriginalLocalityKV {
				__antithesis_instrumentation__.Notify(10581)
				conf, err := cloud.ExternalStorageConfFromURI(uri, user)
				if err != nil {
					__antithesis_instrumentation__.Notify(10583)
					return nil, errors.Wrap(err,
						"creating locality external storage configuration")
				} else {
					__antithesis_instrumentation__.Notify(10584)
				}
				__antithesis_instrumentation__.Notify(10582)
				storesByLocalityKV[kv] = conf
			}
		} else {
			__antithesis_instrumentation__.Notify(10585)
		}
		__antithesis_instrumentation__.Notify(10579)
		backupLocalityMap[i] = storesByLocalityKV
	}
	__antithesis_instrumentation__.Notify(10577)

	return backupLocalityMap, nil
}

func restore(
	restoreCtx context.Context,
	execCtx sql.JobExecContext,
	numNodes int,
	backupManifests []BackupManifest,
	backupLocalityInfo []jobspb.RestoreDetails_BackupLocalityInfo,
	endTime hlc.Timestamp,
	dataToRestore restorationData,
	job *jobs.Job,
	encryption *jobspb.BackupEncryptionOptions,
) (roachpb.RowCount, error) {
	__antithesis_instrumentation__.Notify(10586)
	user := execCtx.User()

	emptyRowCount := roachpb.RowCount{}

	if dataToRestore.isEmpty() {
		__antithesis_instrumentation__.Notify(10599)
		return emptyRowCount, nil
	} else {
		__antithesis_instrumentation__.Notify(10600)
	}
	__antithesis_instrumentation__.Notify(10587)

	details := job.Details().(jobspb.RestoreDetails)
	if alreadyMigrated := checkForMigratedData(details, dataToRestore); alreadyMigrated {
		__antithesis_instrumentation__.Notify(10601)
		return emptyRowCount, nil
	} else {
		__antithesis_instrumentation__.Notify(10602)
	}
	__antithesis_instrumentation__.Notify(10588)

	mu := struct {
		syncutil.Mutex
		highWaterMark     int
		res               roachpb.RowCount
		requestsCompleted []bool
	}{
		highWaterMark: -1,
	}

	backupLocalityMap, err := makeBackupLocalityMap(backupLocalityInfo, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(10603)
		return emptyRowCount, errors.Wrap(err, "resolving locality locations")
	} else {
		__antithesis_instrumentation__.Notify(10604)
	}
	__antithesis_instrumentation__.Notify(10589)

	if err := checkCoverage(restoreCtx, dataToRestore.getSpans(), backupManifests); err != nil {
		__antithesis_instrumentation__.Notify(10605)
		return emptyRowCount, err
	} else {
		__antithesis_instrumentation__.Notify(10606)
	}
	__antithesis_instrumentation__.Notify(10590)

	highWaterMark := job.Progress().Details.(*jobspb.Progress_Restore).Restore.HighWater

	importSpans := makeSimpleImportSpans(dataToRestore.getSpans(), backupManifests, backupLocalityMap,
		highWaterMark)

	if len(importSpans) == 0 {
		__antithesis_instrumentation__.Notify(10607)

		return emptyRowCount, nil
	} else {
		__antithesis_instrumentation__.Notify(10608)
	}
	__antithesis_instrumentation__.Notify(10591)

	for i := range importSpans {
		__antithesis_instrumentation__.Notify(10609)
		importSpans[i].ProgressIdx = int64(i)
	}
	__antithesis_instrumentation__.Notify(10592)
	mu.requestsCompleted = make([]bool, len(importSpans))

	chunkSize := int(math.Sqrt(float64(len(importSpans)))) / numNodes
	if chunkSize == 0 {
		__antithesis_instrumentation__.Notify(10610)
		chunkSize = 1
	} else {
		__antithesis_instrumentation__.Notify(10611)
	}
	__antithesis_instrumentation__.Notify(10593)
	importSpanChunks := make([][]execinfrapb.RestoreSpanEntry, 0, len(importSpans)/chunkSize)
	for start := 0; start < len(importSpans); {
		__antithesis_instrumentation__.Notify(10612)
		importSpanChunk := importSpans[start:]
		end := start + chunkSize
		if end < len(importSpans) {
			__antithesis_instrumentation__.Notify(10614)
			importSpanChunk = importSpans[start:end]
		} else {
			__antithesis_instrumentation__.Notify(10615)
		}
		__antithesis_instrumentation__.Notify(10613)
		importSpanChunks = append(importSpanChunks, importSpanChunk)
		start = end
	}
	__antithesis_instrumentation__.Notify(10594)

	requestFinishedCh := make(chan struct{}, len(importSpans))
	progCh := make(chan *execinfrapb.RemoteProducerMetadata_BulkProcessorProgress)

	var tasks []func(ctx context.Context) error
	if dataToRestore.isMainBundle() {
		__antithesis_instrumentation__.Notify(10616)

		progressLogger := jobs.NewChunkProgressLogger(job, len(importSpans), job.FractionCompleted(),
			func(progressedCtx context.Context, details jobspb.ProgressDetails) {
				__antithesis_instrumentation__.Notify(10619)
				switch d := details.(type) {
				case *jobspb.Progress_Restore:
					__antithesis_instrumentation__.Notify(10620)
					mu.Lock()
					if mu.highWaterMark >= 0 {
						__antithesis_instrumentation__.Notify(10623)
						d.Restore.HighWater = importSpans[mu.highWaterMark].Span.Key
					} else {
						__antithesis_instrumentation__.Notify(10624)
					}
					__antithesis_instrumentation__.Notify(10621)
					mu.Unlock()
				default:
					__antithesis_instrumentation__.Notify(10622)
					log.Errorf(progressedCtx, "job payload had unexpected type %T", d)
				}
			})
		__antithesis_instrumentation__.Notify(10617)

		jobProgressLoop := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(10625)
			ctx, progressSpan := tracing.ChildSpan(ctx, "progress-log")
			defer progressSpan.Finish()
			return progressLogger.Loop(ctx, requestFinishedCh)
		}
		__antithesis_instrumentation__.Notify(10618)
		tasks = append(tasks, jobProgressLoop)
	} else {
		__antithesis_instrumentation__.Notify(10626)
	}
	__antithesis_instrumentation__.Notify(10595)

	jobCheckpointLoop := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(10627)
		defer close(requestFinishedCh)

		for progress := range progCh {
			__antithesis_instrumentation__.Notify(10629)
			mu.Lock()
			var progDetails RestoreProgress
			if err := pbtypes.UnmarshalAny(&progress.ProgressDetails, &progDetails); err != nil {
				__antithesis_instrumentation__.Notify(10633)
				log.Errorf(ctx, "unable to unmarshal restore progress details: %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(10634)
			}
			__antithesis_instrumentation__.Notify(10630)

			mu.res.Add(progDetails.Summary)
			idx := progDetails.ProgressIdx

			if !importSpans[progDetails.ProgressIdx].Span.Key.Equal(progDetails.DataSpan.Key) {
				__antithesis_instrumentation__.Notify(10635)
				mu.Unlock()
				return errors.Newf("request %d for span %v does not match import span for same idx: %v",
					idx, progDetails.DataSpan, importSpans[idx],
				)
			} else {
				__antithesis_instrumentation__.Notify(10636)
			}
			__antithesis_instrumentation__.Notify(10631)
			mu.requestsCompleted[idx] = true
			for j := mu.highWaterMark + 1; j < len(mu.requestsCompleted) && func() bool {
				__antithesis_instrumentation__.Notify(10637)
				return mu.requestsCompleted[j] == true
			}() == true; j++ {
				__antithesis_instrumentation__.Notify(10638)
				mu.highWaterMark = j
			}
			__antithesis_instrumentation__.Notify(10632)
			mu.Unlock()

			requestFinishedCh <- struct{}{}
		}
		__antithesis_instrumentation__.Notify(10628)
		return nil
	}
	__antithesis_instrumentation__.Notify(10596)
	tasks = append(tasks, jobCheckpointLoop)

	runRestore := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(10639)
		return distRestore(
			ctx,
			execCtx,
			int64(job.ID()),
			importSpanChunks,
			dataToRestore.getPKIDs(),
			encryption,
			dataToRestore.getRekeys(),
			dataToRestore.getTenantRekeys(),
			endTime,
			progCh,
		)
	}
	__antithesis_instrumentation__.Notify(10597)
	tasks = append(tasks, runRestore)

	if err := ctxgroup.GoAndWait(restoreCtx, tasks...); err != nil {
		__antithesis_instrumentation__.Notify(10640)

		return emptyRowCount, errors.Wrapf(err, "importing %d ranges", len(importSpans))
	} else {
		__antithesis_instrumentation__.Notify(10641)
	}
	__antithesis_instrumentation__.Notify(10598)

	return mu.res, nil
}

func loadBackupSQLDescs(
	ctx context.Context,
	mem *mon.BoundAccount,
	p sql.JobExecContext,
	details jobspb.RestoreDetails,
	encryption *jobspb.BackupEncryptionOptions,
) ([]BackupManifest, BackupManifest, []catalog.Descriptor, int64, error) {
	__antithesis_instrumentation__.Notify(10642)
	backupManifests, sz, err := loadBackupManifests(ctx, mem, details.URIs,
		p.User(), p.ExecCfg().DistSQLSrv.ExternalStorageFromURI, encryption)
	if err != nil {
		__antithesis_instrumentation__.Notify(10647)
		return nil, BackupManifest{}, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10648)
	}
	__antithesis_instrumentation__.Notify(10643)

	allDescs, latestBackupManifest := loadSQLDescsFromBackupsAtTime(backupManifests, details.EndTime)

	for _, m := range details.DatabaseModifiers {
		__antithesis_instrumentation__.Notify(10649)
		for _, typ := range m.ExtraTypeDescs {
			__antithesis_instrumentation__.Notify(10650)
			allDescs = append(allDescs, typedesc.NewBuilder(typ).BuildCreatedMutableType())
		}
	}
	__antithesis_instrumentation__.Notify(10644)

	var sqlDescs []catalog.Descriptor
	for _, desc := range allDescs {
		__antithesis_instrumentation__.Notify(10651)
		id := desc.GetID()
		switch desc := desc.(type) {
		case *dbdesc.Mutable:
			__antithesis_instrumentation__.Notify(10653)
			if m, ok := details.DatabaseModifiers[id]; ok {
				__antithesis_instrumentation__.Notify(10654)
				desc.SetRegionConfig(m.RegionConfig)
			} else {
				__antithesis_instrumentation__.Notify(10655)
			}
		}
		__antithesis_instrumentation__.Notify(10652)
		if _, ok := details.DescriptorRewrites[id]; ok {
			__antithesis_instrumentation__.Notify(10656)
			sqlDescs = append(sqlDescs, desc)
		} else {
			__antithesis_instrumentation__.Notify(10657)
		}
	}
	__antithesis_instrumentation__.Notify(10645)

	if err := maybeUpgradeDescriptors(sqlDescs, true); err != nil {
		__antithesis_instrumentation__.Notify(10658)
		mem.Shrink(ctx, sz)
		return nil, BackupManifest{}, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(10659)
	}
	__antithesis_instrumentation__.Notify(10646)

	return backupManifests, latestBackupManifest, sqlDescs, sz, nil
}

type restoreResumer struct {
	job *jobs.Job

	settings     *cluster.Settings
	execCfg      *sql.ExecutorConfig
	restoreStats roachpb.RowCount

	testingKnobs struct {
		beforePublishingDescriptors func() error

		afterPublishingDescriptors func() error

		duringSystemTableRestoration func(systemTableName string) error

		afterOfflineTableCreation func() error

		afterPreRestore func() error
	}
}

func getStatisticsFromBackup(
	ctx context.Context,
	exportStore cloud.ExternalStorage,
	encryption *jobspb.BackupEncryptionOptions,
	backup BackupManifest,
) ([]*stats.TableStatisticProto, error) {
	__antithesis_instrumentation__.Notify(10660)

	if backup.DeprecatedStatistics != nil {
		__antithesis_instrumentation__.Notify(10663)
		return backup.DeprecatedStatistics, nil
	} else {
		__antithesis_instrumentation__.Notify(10664)
	}
	__antithesis_instrumentation__.Notify(10661)
	tableStatistics := make([]*stats.TableStatisticProto, 0, len(backup.StatisticsFilenames))
	uniqueFileNames := make(map[string]struct{})
	for _, fname := range backup.StatisticsFilenames {
		__antithesis_instrumentation__.Notify(10665)
		if _, exists := uniqueFileNames[fname]; !exists {
			__antithesis_instrumentation__.Notify(10666)
			uniqueFileNames[fname] = struct{}{}
			myStatsTable, err := readTableStatistics(ctx, exportStore, fname, encryption)
			if err != nil {
				__antithesis_instrumentation__.Notify(10668)
				return tableStatistics, err
			} else {
				__antithesis_instrumentation__.Notify(10669)
			}
			__antithesis_instrumentation__.Notify(10667)
			tableStatistics = append(tableStatistics, myStatsTable.Statistics...)
		} else {
			__antithesis_instrumentation__.Notify(10670)
		}
	}
	__antithesis_instrumentation__.Notify(10662)

	return tableStatistics, nil
}

func remapRelevantStatistics(
	ctx context.Context,
	tableStatistics []*stats.TableStatisticProto,
	descriptorRewrites jobspb.DescRewriteMap,
	tableDescs []*descpb.TableDescriptor,
) []*stats.TableStatisticProto {
	__antithesis_instrumentation__.Notify(10671)
	relevantTableStatistics := make([]*stats.TableStatisticProto, 0, len(tableStatistics))

	tableHasStatsInBackup := make(map[descpb.ID]struct{})
	for _, stat := range tableStatistics {
		__antithesis_instrumentation__.Notify(10674)
		tableHasStatsInBackup[stat.TableID] = struct{}{}
		if tableRewrite, ok := descriptorRewrites[stat.TableID]; ok {
			__antithesis_instrumentation__.Notify(10675)

			stat.TableID = tableRewrite.ID
			relevantTableStatistics = append(relevantTableStatistics, stat)
		} else {
			__antithesis_instrumentation__.Notify(10676)
		}
	}
	__antithesis_instrumentation__.Notify(10672)

	for _, desc := range tableDescs {
		__antithesis_instrumentation__.Notify(10677)
		if _, ok := tableHasStatsInBackup[desc.GetID()]; !ok {
			__antithesis_instrumentation__.Notify(10678)
			log.Warningf(ctx, "statistics for table: %s, table ID: %d not found in the backup. "+
				"Query performance on this table could suffer until statistics are recomputed.",
				desc.GetName(), desc.GetID())
		} else {
			__antithesis_instrumentation__.Notify(10679)
		}
	}
	__antithesis_instrumentation__.Notify(10673)

	return relevantTableStatistics
}

func isDatabaseEmpty(
	ctx context.Context,
	txn *kv.Txn,
	dbID descpb.ID,
	allDescs []catalog.Descriptor,
	ignoredChildren map[descpb.ID]struct{},
) (bool, error) {
	__antithesis_instrumentation__.Notify(10680)
	for _, desc := range allDescs {
		__antithesis_instrumentation__.Notify(10682)
		if _, ok := ignoredChildren[desc.GetID()]; ok {
			__antithesis_instrumentation__.Notify(10684)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10685)
		}
		__antithesis_instrumentation__.Notify(10683)
		if desc.GetParentID() == dbID {
			__antithesis_instrumentation__.Notify(10686)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(10687)
		}
	}
	__antithesis_instrumentation__.Notify(10681)
	return true, nil
}

func isSchemaEmpty(
	ctx context.Context,
	txn *kv.Txn,
	schemaID descpb.ID,
	allDescs []catalog.Descriptor,
	ignoredChildren map[descpb.ID]struct{},
) (bool, error) {
	__antithesis_instrumentation__.Notify(10688)
	for _, desc := range allDescs {
		__antithesis_instrumentation__.Notify(10690)
		if _, ok := ignoredChildren[desc.GetID()]; ok {
			__antithesis_instrumentation__.Notify(10692)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10693)
		}
		__antithesis_instrumentation__.Notify(10691)
		if desc.GetParentSchemaID() == schemaID {
			__antithesis_instrumentation__.Notify(10694)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(10695)
		}
	}
	__antithesis_instrumentation__.Notify(10689)
	return true, nil
}

func spansForAllRestoreTableIndexes(
	codec keys.SQLCodec, tables []catalog.TableDescriptor, revs []BackupManifest_DescriptorRevision,
) []roachpb.Span {
	__antithesis_instrumentation__.Notify(10696)

	added := make(map[tableAndIndex]bool, len(tables))
	sstIntervalTree := interval.NewTree(interval.ExclusiveOverlapper)
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(10700)

		if !table.IsPhysicalTable() {
			__antithesis_instrumentation__.Notify(10702)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10703)
		}
		__antithesis_instrumentation__.Notify(10701)
		for _, index := range table.ActiveIndexes() {
			__antithesis_instrumentation__.Notify(10704)
			if err := sstIntervalTree.Insert(intervalSpan(table.IndexSpan(codec, index.GetID())), false); err != nil {
				__antithesis_instrumentation__.Notify(10706)
				panic(errors.NewAssertionErrorWithWrappedErrf(err, "IndexSpan"))
			} else {
				__antithesis_instrumentation__.Notify(10707)
			}
			__antithesis_instrumentation__.Notify(10705)
			added[tableAndIndex{tableID: table.GetID(), indexID: index.GetID()}] = true
		}
	}
	__antithesis_instrumentation__.Notify(10697)

	for _, rev := range revs {
		__antithesis_instrumentation__.Notify(10708)

		rawTbl, _, _, _ := descpb.FromDescriptor(rev.Desc)
		if rawTbl != nil && func() bool {
			__antithesis_instrumentation__.Notify(10709)
			return !rawTbl.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(10710)
			tbl := tabledesc.NewBuilder(rawTbl).BuildImmutableTable()

			if !tbl.IsPhysicalTable() {
				__antithesis_instrumentation__.Notify(10712)
				continue
			} else {
				__antithesis_instrumentation__.Notify(10713)
			}
			__antithesis_instrumentation__.Notify(10711)
			for _, idx := range tbl.ActiveIndexes() {
				__antithesis_instrumentation__.Notify(10714)
				key := tableAndIndex{tableID: tbl.GetID(), indexID: idx.GetID()}
				if !added[key] {
					__antithesis_instrumentation__.Notify(10715)
					if err := sstIntervalTree.Insert(intervalSpan(tbl.IndexSpan(codec, idx.GetID())), false); err != nil {
						__antithesis_instrumentation__.Notify(10717)
						panic(errors.NewAssertionErrorWithWrappedErrf(err, "IndexSpan"))
					} else {
						__antithesis_instrumentation__.Notify(10718)
					}
					__antithesis_instrumentation__.Notify(10716)
					added[key] = true
				} else {
					__antithesis_instrumentation__.Notify(10719)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(10720)
		}
	}
	__antithesis_instrumentation__.Notify(10698)

	var spans []roachpb.Span
	_ = sstIntervalTree.Do(func(r interval.Interface) bool {
		__antithesis_instrumentation__.Notify(10721)
		spans = append(spans, roachpb.Span{
			Key:    roachpb.Key(r.Range().Start),
			EndKey: roachpb.Key(r.Range().End),
		})
		return false
	})
	__antithesis_instrumentation__.Notify(10699)
	return spans
}

func shouldPreRestore(table *tabledesc.Mutable) bool {
	__antithesis_instrumentation__.Notify(10722)
	if table.GetParentID() != keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(10724)
		return false
	} else {
		__antithesis_instrumentation__.Notify(10725)
	}
	__antithesis_instrumentation__.Notify(10723)
	tablesToPreRestore := getSystemTablesToRestoreBeforeData()
	_, ok := tablesToPreRestore[table.GetName()]
	return ok
}

func createImportingDescriptors(
	ctx context.Context,
	p sql.JobExecContext,
	backupCodec keys.SQLCodec,
	sqlDescs []catalog.Descriptor,
	r *restoreResumer,
) (*restorationDataBase, *mainRestorationData, error) {
	__antithesis_instrumentation__.Notify(10726)
	details := r.job.Details().(jobspb.RestoreDetails)

	var databases []catalog.DatabaseDescriptor
	var writtenTypes []catalog.TypeDescriptor
	var schemas []*schemadesc.Mutable
	var types []*typedesc.Mutable

	var mutableTables []*tabledesc.Mutable
	var mutableDatabases []*dbdesc.Mutable

	oldTableIDs := make([]descpb.ID, 0)

	tables := make([]catalog.TableDescriptor, 0)
	postRestoreTables := make([]catalog.TableDescriptor, 0)

	preRestoreTables := make([]catalog.TableDescriptor, 0)

	for _, desc := range sqlDescs {
		__antithesis_instrumentation__.Notify(10749)
		switch desc := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(10750)
			mut := tabledesc.NewBuilder(desc.TableDesc()).BuildCreatedMutableTable()
			if shouldPreRestore(mut) {
				__antithesis_instrumentation__.Notify(10755)
				preRestoreTables = append(preRestoreTables, mut)
			} else {
				__antithesis_instrumentation__.Notify(10756)
				postRestoreTables = append(postRestoreTables, mut)
			}
			__antithesis_instrumentation__.Notify(10751)
			tables = append(tables, mut)
			mutableTables = append(mutableTables, mut)
			oldTableIDs = append(oldTableIDs, mut.GetID())
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(10752)
			if _, ok := details.DescriptorRewrites[desc.GetID()]; ok {
				__antithesis_instrumentation__.Notify(10757)
				mut := dbdesc.NewBuilder(desc.DatabaseDesc()).BuildCreatedMutableDatabase()
				databases = append(databases, mut)
				mutableDatabases = append(mutableDatabases, mut)
			} else {
				__antithesis_instrumentation__.Notify(10758)
			}
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(10753)
			mut := schemadesc.NewBuilder(desc.SchemaDesc()).BuildCreatedMutableSchema()
			schemas = append(schemas, mut)
		case catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(10754)
			mut := typedesc.NewBuilder(desc.TypeDesc()).BuildCreatedMutableType()
			types = append(types, mut)
		}
	}
	__antithesis_instrumentation__.Notify(10727)

	tempSystemDBID := tempSystemDatabaseID(details)
	if tempSystemDBID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(10759)
		tempSystemDB := dbdesc.NewInitial(tempSystemDBID, restoreTempSystemDB,
			security.AdminRoleName(), dbdesc.WithPublicSchemaID(keys.SystemPublicSchemaID))
		databases = append(databases, tempSystemDB)
	} else {
		__antithesis_instrumentation__.Notify(10760)
	}
	__antithesis_instrumentation__.Notify(10728)

	preRestoreSpans := spansForAllRestoreTableIndexes(backupCodec, preRestoreTables, nil)
	postRestoreSpans := spansForAllRestoreTableIndexes(backupCodec, postRestoreTables, nil)

	log.Eventf(ctx, "starting restore for %d tables", len(mutableTables))

	if err := rewrite.DatabaseDescs(mutableDatabases, details.DescriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(10761)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(10762)
	}
	__antithesis_instrumentation__.Notify(10729)

	databaseDescs := make([]*descpb.DatabaseDescriptor, len(mutableDatabases))
	for i, database := range mutableDatabases {
		__antithesis_instrumentation__.Notify(10763)
		databaseDescs[i] = database.DatabaseDesc()
	}
	__antithesis_instrumentation__.Notify(10730)

	var schemasToWrite []*schemadesc.Mutable
	var writtenSchemas []catalog.SchemaDescriptor
	for i := range schemas {
		__antithesis_instrumentation__.Notify(10764)
		sc := schemas[i]
		rw, ok := details.DescriptorRewrites[sc.ID]
		if ok {
			__antithesis_instrumentation__.Notify(10765)
			if !rw.ToExisting {
				__antithesis_instrumentation__.Notify(10766)
				schemasToWrite = append(schemasToWrite, sc)
				writtenSchemas = append(writtenSchemas, sc)
			} else {
				__antithesis_instrumentation__.Notify(10767)
			}
		} else {
			__antithesis_instrumentation__.Notify(10768)
		}
	}
	__antithesis_instrumentation__.Notify(10731)

	if err := rewrite.SchemaDescs(schemasToWrite, details.DescriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(10769)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(10770)
	}
	__antithesis_instrumentation__.Notify(10732)

	if err := remapPublicSchemas(ctx, p, mutableDatabases, &schemasToWrite, &writtenSchemas, &details); err != nil {
		__antithesis_instrumentation__.Notify(10771)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(10772)
	}
	__antithesis_instrumentation__.Notify(10733)

	if err := rewrite.TableDescs(
		mutableTables, details.DescriptorRewrites, details.OverrideDB,
	); err != nil {
		__antithesis_instrumentation__.Notify(10773)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(10774)
	}
	__antithesis_instrumentation__.Notify(10734)
	tableDescs := make([]*descpb.TableDescriptor, len(mutableTables))
	for i, table := range mutableTables {
		__antithesis_instrumentation__.Notify(10775)
		tableDescs[i] = table.TableDesc()
	}
	__antithesis_instrumentation__.Notify(10735)

	var typesToWrite []*typedesc.Mutable
	existingTypeIDs := make(map[descpb.ID]struct{})
	for i := range types {
		__antithesis_instrumentation__.Notify(10776)
		typ := types[i]
		rewrite := details.DescriptorRewrites[typ.GetID()]
		if rewrite.ToExisting {
			__antithesis_instrumentation__.Notify(10777)
			existingTypeIDs[rewrite.ID] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(10778)
			typesToWrite = append(typesToWrite, typ)
			writtenTypes = append(writtenTypes, typ)
		}
	}
	__antithesis_instrumentation__.Notify(10736)

	if err := rewrite.TypeDescs(types, details.DescriptorRewrites); err != nil {
		__antithesis_instrumentation__.Notify(10779)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(10780)
	}
	__antithesis_instrumentation__.Notify(10737)

	for _, desc := range mutableTables {
		__antithesis_instrumentation__.Notify(10781)
		desc.SetOffline("restoring")
	}
	__antithesis_instrumentation__.Notify(10738)
	for _, desc := range typesToWrite {
		__antithesis_instrumentation__.Notify(10782)
		desc.SetOffline("restoring")
	}
	__antithesis_instrumentation__.Notify(10739)
	for _, desc := range schemasToWrite {
		__antithesis_instrumentation__.Notify(10783)
		desc.SetOffline("restoring")
	}
	__antithesis_instrumentation__.Notify(10740)
	for _, desc := range mutableDatabases {
		__antithesis_instrumentation__.Notify(10784)
		desc.SetOffline("restoring")
	}
	__antithesis_instrumentation__.Notify(10741)

	if tempSystemDBID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(10785)
		for _, desc := range mutableTables {
			__antithesis_instrumentation__.Notify(10786)
			if desc.GetParentID() == tempSystemDBID {
				__antithesis_instrumentation__.Notify(10787)
				desc.SetPublic()
			} else {
				__antithesis_instrumentation__.Notify(10788)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(10789)
	}
	__antithesis_instrumentation__.Notify(10742)

	typesByID := make(map[descpb.ID]catalog.TypeDescriptor)
	for i := range types {
		__antithesis_instrumentation__.Notify(10790)
		typesByID[types[i].GetID()] = types[i]
	}
	__antithesis_instrumentation__.Notify(10743)

	dbsByID := make(map[descpb.ID]catalog.DatabaseDescriptor)
	for i := range databases {
		__antithesis_instrumentation__.Notify(10791)
		dbsByID[databases[i].GetID()] = databases[i]
	}
	__antithesis_instrumentation__.Notify(10744)

	if !details.PrepareCompleted {
		__antithesis_instrumentation__.Notify(10792)
		err := sql.DescsTxn(ctx, p.ExecCfg(), func(
			ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(10794)

			mrEnumsFound := make(map[descpb.ID]descpb.ID)
			for _, t := range typesByID {
				__antithesis_instrumentation__.Notify(10806)
				typeDesc := typedesc.NewBuilder(t.TypeDesc()).BuildImmutableType()
				if typeDesc.GetKind() == descpb.TypeDescriptor_MULTIREGION_ENUM {
					__antithesis_instrumentation__.Notify(10807)

					if id, ok := mrEnumsFound[typeDesc.GetParentID()]; ok {
						__antithesis_instrumentation__.Notify(10809)
						return errors.AssertionFailedf(
							"unexpectedly found more than one MULTIREGION_ENUM (IDs = %d, %d) "+
								"on database %d during restore", id, typeDesc.GetID(), typeDesc.GetParentID())
					} else {
						__antithesis_instrumentation__.Notify(10810)
					}
					__antithesis_instrumentation__.Notify(10808)
					mrEnumsFound[typeDesc.GetParentID()] = typeDesc.GetID()

					if db, ok := dbsByID[typeDesc.GetParentID()]; ok {
						__antithesis_instrumentation__.Notify(10811)
						desc := db.DatabaseDesc()
						if desc.RegionConfig == nil {
							__antithesis_instrumentation__.Notify(10813)
							return errors.AssertionFailedf(
								"found MULTIREGION_ENUM on non-multi-region database %s", desc.Name)
						} else {
							__antithesis_instrumentation__.Notify(10814)
						}
						__antithesis_instrumentation__.Notify(10812)

						desc.RegionConfig.RegionEnumID = t.GetID()

						if details.DescriptorCoverage != tree.AllDescriptors {
							__antithesis_instrumentation__.Notify(10815)
							log.Infof(ctx, "restoring zone configuration for database %d", desc.ID)
							regionNames, err := typeDesc.RegionNames()
							if err != nil {
								__antithesis_instrumentation__.Notify(10819)
								return err
							} else {
								__antithesis_instrumentation__.Notify(10820)
							}
							__antithesis_instrumentation__.Notify(10816)
							superRegions, err := typeDesc.SuperRegions()
							if err != nil {
								__antithesis_instrumentation__.Notify(10821)
								return err
							} else {
								__antithesis_instrumentation__.Notify(10822)
							}
							__antithesis_instrumentation__.Notify(10817)
							zoneCfgExtensions, err := typeDesc.ZoneConfigExtensions()
							if err != nil {
								__antithesis_instrumentation__.Notify(10823)
								return err
							} else {
								__antithesis_instrumentation__.Notify(10824)
							}
							__antithesis_instrumentation__.Notify(10818)
							regionConfig := multiregion.MakeRegionConfig(
								regionNames,
								desc.RegionConfig.PrimaryRegion,
								desc.RegionConfig.SurvivalGoal,
								desc.RegionConfig.RegionEnumID,
								desc.RegionConfig.Placement,
								superRegions,
								zoneCfgExtensions,
							)
							if err := sql.ApplyZoneConfigFromDatabaseRegionConfig(
								ctx,
								desc.GetID(),
								regionConfig,
								txn,
								p.ExecCfg(),
							); err != nil {
								__antithesis_instrumentation__.Notify(10825)
								return err
							} else {
								__antithesis_instrumentation__.Notify(10826)
							}
						} else {
							__antithesis_instrumentation__.Notify(10827)
						}
					} else {
						__antithesis_instrumentation__.Notify(10828)
					}
				} else {
					__antithesis_instrumentation__.Notify(10829)
				}
			}
			__antithesis_instrumentation__.Notify(10795)

			if details.DescriptorCoverage != tree.AllDescriptors {
				__antithesis_instrumentation__.Notify(10830)
				for _, table := range mutableTables {
					__antithesis_instrumentation__.Notify(10831)
					if table.HasRowLevelTTL() {
						__antithesis_instrumentation__.Notify(10832)
						table.RowLevelTTL.ScheduleID = 0
					} else {
						__antithesis_instrumentation__.Notify(10833)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(10834)
			}
			__antithesis_instrumentation__.Notify(10796)

			if err := ingesting.WriteDescriptors(
				ctx, p.ExecCfg().Codec, txn, p.User(), descsCol, databases, writtenSchemas, tables, writtenTypes,
				details.DescriptorCoverage, nil, restoreTempSystemDB,
			); err != nil {
				__antithesis_instrumentation__.Notify(10835)
				return errors.Wrapf(err, "restoring %d TableDescriptors from %d databases", len(tables), len(databases))
			} else {
				__antithesis_instrumentation__.Notify(10836)
			}
			__antithesis_instrumentation__.Notify(10797)

			b := txn.NewBatch()

			existingDBsWithNewSchemas := make(map[descpb.ID][]catalog.SchemaDescriptor)
			for _, sc := range writtenSchemas {
				__antithesis_instrumentation__.Notify(10837)
				parentID := sc.GetParentID()
				if _, ok := dbsByID[parentID]; !ok {
					__antithesis_instrumentation__.Notify(10838)
					existingDBsWithNewSchemas[parentID] = append(existingDBsWithNewSchemas[parentID], sc)
				} else {
					__antithesis_instrumentation__.Notify(10839)
				}
			}
			__antithesis_instrumentation__.Notify(10798)

			for dbID, schemas := range existingDBsWithNewSchemas {
				__antithesis_instrumentation__.Notify(10840)
				log.Infof(ctx, "writing %d schema entries to database %d", len(schemas), dbID)
				desc, err := descsCol.GetMutableDescriptorByID(ctx, txn, dbID)
				if err != nil {
					__antithesis_instrumentation__.Notify(10843)
					return err
				} else {
					__antithesis_instrumentation__.Notify(10844)
				}
				__antithesis_instrumentation__.Notify(10841)
				db := desc.(*dbdesc.Mutable)
				for _, sc := range schemas {
					__antithesis_instrumentation__.Notify(10845)
					db.AddSchemaToDatabase(sc.GetName(), descpb.DatabaseDescriptor_SchemaInfo{ID: sc.GetID()})
				}
				__antithesis_instrumentation__.Notify(10842)
				if err := descsCol.WriteDescToBatch(
					ctx, false, db, b,
				); err != nil {
					__antithesis_instrumentation__.Notify(10846)
					return err
				} else {
					__antithesis_instrumentation__.Notify(10847)
				}
			}
			__antithesis_instrumentation__.Notify(10799)

			for _, table := range mutableTables {
				__antithesis_instrumentation__.Notify(10848)

				_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
					ctx, txn, table.GetParentID(), tree.DatabaseLookupFlags{
						Required:       true,
						AvoidLeased:    true,
						IncludeOffline: true,
					})
				if err != nil {
					__antithesis_instrumentation__.Notify(10852)
					return err
				} else {
					__antithesis_instrumentation__.Notify(10853)
				}
				__antithesis_instrumentation__.Notify(10849)
				typeIDs, _, err := table.GetAllReferencedTypeIDs(dbDesc, func(id descpb.ID) (catalog.TypeDescriptor, error) {
					__antithesis_instrumentation__.Notify(10854)
					t, ok := typesByID[id]
					if !ok {
						__antithesis_instrumentation__.Notify(10856)
						return nil, errors.AssertionFailedf("type with id %d was not found in rewritten type mapping", id)
					} else {
						__antithesis_instrumentation__.Notify(10857)
					}
					__antithesis_instrumentation__.Notify(10855)
					return t, nil
				})
				__antithesis_instrumentation__.Notify(10850)
				if err != nil {
					__antithesis_instrumentation__.Notify(10858)
					return err
				} else {
					__antithesis_instrumentation__.Notify(10859)
				}
				__antithesis_instrumentation__.Notify(10851)
				for _, id := range typeIDs {
					__antithesis_instrumentation__.Notify(10860)

					_, ok := existingTypeIDs[id]
					if !ok {
						__antithesis_instrumentation__.Notify(10863)
						continue
					} else {
						__antithesis_instrumentation__.Notify(10864)
					}
					__antithesis_instrumentation__.Notify(10861)

					typDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, id)
					if err != nil {
						__antithesis_instrumentation__.Notify(10865)
						return err
					} else {
						__antithesis_instrumentation__.Notify(10866)
					}
					__antithesis_instrumentation__.Notify(10862)
					typDesc.AddReferencingDescriptorID(table.GetID())
					if err := descsCol.WriteDescToBatch(
						ctx, false, typDesc, b,
					); err != nil {
						__antithesis_instrumentation__.Notify(10867)
						return err
					} else {
						__antithesis_instrumentation__.Notify(10868)
					}
				}
			}
			__antithesis_instrumentation__.Notify(10800)
			if err := txn.Run(ctx, b); err != nil {
				__antithesis_instrumentation__.Notify(10869)
				return err
			} else {
				__antithesis_instrumentation__.Notify(10870)
			}
			__antithesis_instrumentation__.Notify(10801)

			if details.DescriptorCoverage != tree.AllDescriptors {
				__antithesis_instrumentation__.Notify(10871)
				for _, table := range tableDescs {
					__antithesis_instrumentation__.Notify(10872)
					if lc := table.GetLocalityConfig(); lc != nil {
						__antithesis_instrumentation__.Notify(10873)
						_, desc, err := descsCol.GetImmutableDatabaseByID(
							ctx,
							txn,
							table.ParentID,
							tree.DatabaseLookupFlags{
								Required:       true,
								AvoidLeased:    true,
								IncludeOffline: true,
							},
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(10878)
							return err
						} else {
							__antithesis_instrumentation__.Notify(10879)
						}
						__antithesis_instrumentation__.Notify(10874)
						if desc.GetRegionConfig() == nil {
							__antithesis_instrumentation__.Notify(10880)
							return errors.AssertionFailedf(
								"found multi-region table %d in non-multi-region database %d",
								table.ID, table.ParentID)
						} else {
							__antithesis_instrumentation__.Notify(10881)
						}
						__antithesis_instrumentation__.Notify(10875)

						mutTable, err := descsCol.GetMutableTableVersionByID(ctx, table.GetID(), txn)
						if err != nil {
							__antithesis_instrumentation__.Notify(10882)
							return err
						} else {
							__antithesis_instrumentation__.Notify(10883)
						}
						__antithesis_instrumentation__.Notify(10876)

						regionConfig, err := sql.SynthesizeRegionConfig(
							ctx,
							txn,
							desc.GetID(),
							descsCol,
							sql.SynthesizeRegionConfigOptionIncludeOffline,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(10884)
							return err
						} else {
							__antithesis_instrumentation__.Notify(10885)
						}
						__antithesis_instrumentation__.Notify(10877)
						if err := sql.ApplyZoneConfigForMultiRegionTable(
							ctx,
							txn,
							p.ExecCfg(),
							regionConfig,
							mutTable,
							sql.ApplyZoneConfigForMultiRegionTableOptionTableAndIndexes,
						); err != nil {
							__antithesis_instrumentation__.Notify(10886)
							return err
						} else {
							__antithesis_instrumentation__.Notify(10887)
						}
					} else {
						__antithesis_instrumentation__.Notify(10888)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(10889)
			}
			__antithesis_instrumentation__.Notify(10802)

			for _, tenant := range details.Tenants {
				__antithesis_instrumentation__.Notify(10890)

				tenant.State = descpb.TenantInfo_ADD
				if err := sql.CreateTenantRecord(ctx, p.ExecCfg(), txn, &tenant); err != nil {
					__antithesis_instrumentation__.Notify(10891)
					return err
				} else {
					__antithesis_instrumentation__.Notify(10892)
				}
			}
			__antithesis_instrumentation__.Notify(10803)

			details.PrepareCompleted = true
			details.DatabaseDescs = databaseDescs
			details.TableDescs = tableDescs
			details.TypeDescs = make([]*descpb.TypeDescriptor, len(typesToWrite))
			for i := range typesToWrite {
				__antithesis_instrumentation__.Notify(10893)
				details.TypeDescs[i] = typesToWrite[i].TypeDesc()
			}
			__antithesis_instrumentation__.Notify(10804)
			details.SchemaDescs = make([]*descpb.SchemaDescriptor, len(schemasToWrite))
			for i := range schemasToWrite {
				__antithesis_instrumentation__.Notify(10894)
				details.SchemaDescs[i] = schemasToWrite[i].SchemaDesc()
			}
			__antithesis_instrumentation__.Notify(10805)

			err := r.job.SetDetails(ctx, txn, details)

			emitRestoreJobEvent(ctx, p, jobs.StatusRunning, r.job)

			return err
		})
		__antithesis_instrumentation__.Notify(10793)
		if err != nil {
			__antithesis_instrumentation__.Notify(10895)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(10896)
		}
	} else {
		__antithesis_instrumentation__.Notify(10897)
	}
	__antithesis_instrumentation__.Notify(10745)

	var rekeys []execinfrapb.TableRekey
	for i := range tables {
		__antithesis_instrumentation__.Notify(10898)
		tableToSerialize := tables[i]
		newDescBytes, err := protoutil.Marshal(tableToSerialize.DescriptorProto())
		if err != nil {
			__antithesis_instrumentation__.Notify(10900)
			return nil, nil, errors.NewAssertionErrorWithWrappedErrf(err,
				"marshaling descriptor")
		} else {
			__antithesis_instrumentation__.Notify(10901)
		}
		__antithesis_instrumentation__.Notify(10899)
		rekeys = append(rekeys, execinfrapb.TableRekey{
			OldID:   uint32(oldTableIDs[i]),
			NewDesc: newDescBytes,
		})
	}
	__antithesis_instrumentation__.Notify(10746)

	pkIDs := make(map[uint64]bool)
	for _, tbl := range tables {
		__antithesis_instrumentation__.Notify(10902)
		pkIDs[roachpb.BulkOpSummaryID(uint64(tbl.GetID()), uint64(tbl.GetPrimaryIndexID()))] = true
	}
	__antithesis_instrumentation__.Notify(10747)

	dataToPreRestore := &restorationDataBase{
		spans:       preRestoreSpans,
		tableRekeys: rekeys,
		pkIDs:       pkIDs,
	}

	dataToRestore := &mainRestorationData{
		restorationDataBase{
			spans:       postRestoreSpans,
			tableRekeys: rekeys,
			pkIDs:       pkIDs,
		},
	}

	if tempSystemDBID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(10903)
		for _, table := range preRestoreTables {
			__antithesis_instrumentation__.Notify(10905)
			if table.GetParentID() == tempSystemDBID {
				__antithesis_instrumentation__.Notify(10906)
				dataToPreRestore.systemTables = append(dataToPreRestore.systemTables, table)
			} else {
				__antithesis_instrumentation__.Notify(10907)
			}
		}
		__antithesis_instrumentation__.Notify(10904)
		for _, table := range postRestoreTables {
			__antithesis_instrumentation__.Notify(10908)
			if table.GetParentID() == tempSystemDBID {
				__antithesis_instrumentation__.Notify(10909)
				dataToRestore.systemTables = append(dataToRestore.systemTables, table)
			} else {
				__antithesis_instrumentation__.Notify(10910)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(10911)
	}
	__antithesis_instrumentation__.Notify(10748)
	return dataToPreRestore, dataToRestore, nil
}

func remapPublicSchemas(
	ctx context.Context,
	p sql.JobExecContext,
	mutableDatabases []*dbdesc.Mutable,
	schemasToWrite *[]*schemadesc.Mutable,
	writtenSchemas *[]catalog.SchemaDescriptor,
	details *jobspb.RestoreDetails,
) error {
	__antithesis_instrumentation__.Notify(10912)
	if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(10916)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(10917)
	}
	__antithesis_instrumentation__.Notify(10913)
	databaseToPublicSchemaID := make(map[descpb.ID]descpb.ID)
	for _, db := range mutableDatabases {
		__antithesis_instrumentation__.Notify(10918)
		if db.HasPublicSchemaWithDescriptor() {
			__antithesis_instrumentation__.Notify(10921)
			continue
		} else {
			__antithesis_instrumentation__.Notify(10922)
		}
		__antithesis_instrumentation__.Notify(10919)

		id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(10923)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10924)
		}
		__antithesis_instrumentation__.Notify(10920)

		db.AddSchemaToDatabase(tree.PublicSchema, descpb.DatabaseDescriptor_SchemaInfo{ID: id})

		publicSchemaPrivileges := catpb.NewPublicSchemaPrivilegeDescriptor()
		publicSchemaDesc := schemadesc.NewBuilder(&descpb.SchemaDescriptor{
			ParentID:   db.GetID(),
			Name:       tree.PublicSchema,
			ID:         id,
			Privileges: publicSchemaPrivileges,
			Version:    1,
		}).BuildCreatedMutableSchema()

		*schemasToWrite = append(*schemasToWrite, publicSchemaDesc)
		*writtenSchemas = append(*writtenSchemas, publicSchemaDesc)
		databaseToPublicSchemaID[db.GetID()] = id
	}
	__antithesis_instrumentation__.Notify(10914)

	for id, rw := range details.DescriptorRewrites {
		__antithesis_instrumentation__.Notify(10925)
		if publicSchemaID, ok := databaseToPublicSchemaID[rw.ParentID]; ok {
			__antithesis_instrumentation__.Notify(10926)

			if details.DescriptorRewrites[id].ParentSchemaID == keys.PublicSchemaIDForBackup || func() bool {
				__antithesis_instrumentation__.Notify(10927)
				return details.DescriptorRewrites[id].ParentSchemaID == descpb.InvalidID == true
			}() == true {
				__antithesis_instrumentation__.Notify(10928)
				details.DescriptorRewrites[id].ParentSchemaID = publicSchemaID
			} else {
				__antithesis_instrumentation__.Notify(10929)
			}
		} else {
			__antithesis_instrumentation__.Notify(10930)
		}
	}
	__antithesis_instrumentation__.Notify(10915)

	return nil
}

func (r *restoreResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(10931)
	if err := r.doResume(ctx, execCtx); err != nil {
		__antithesis_instrumentation__.Notify(10933)
		details := r.job.Details().(jobspb.RestoreDetails)
		if details.DebugPauseOn == "error" {
			__antithesis_instrumentation__.Notify(10935)
			const errorFmt = "job failed with error (%v) but is being paused due to the %s=%s setting"
			log.Warningf(ctx, errorFmt, err, restoreOptDebugPauseOn, details.DebugPauseOn)

			return jobs.MarkPauseRequestError(errors.Wrapf(err,
				"pausing job due to the %s=%s setting",
				restoreOptDebugPauseOn, details.DebugPauseOn))
		} else {
			__antithesis_instrumentation__.Notify(10936)
		}
		__antithesis_instrumentation__.Notify(10934)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10937)
	}
	__antithesis_instrumentation__.Notify(10932)

	return nil
}

func (r *restoreResumer) doResume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(10938)
	details := r.job.Details().(jobspb.RestoreDetails)
	p := execCtx.(sql.JobExecContext)
	r.execCfg = p.ExecCfg()

	if details.Validation != jobspb.RestoreValidation_DefaultRestore {
		__antithesis_instrumentation__.Notify(10963)
		return errors.Errorf("No restore validation tools are supported")
	} else {
		__antithesis_instrumentation__.Notify(10964)
	}
	__antithesis_instrumentation__.Notify(10939)

	mem := p.ExecCfg().RootMemoryMonitor.MakeBoundAccount()
	defer mem.Close(ctx)

	if err := p.ExecCfg().JobRegistry.CheckPausepoint("restore.before_load_descriptors_from_backup"); err != nil {
		__antithesis_instrumentation__.Notify(10965)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10966)
	}
	__antithesis_instrumentation__.Notify(10940)

	backupManifests, latestBackupManifest, sqlDescs, memSize, err := loadBackupSQLDescs(
		ctx, &mem, p, details, details.Encryption,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(10967)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10968)
	}
	__antithesis_instrumentation__.Notify(10941)
	defer func() {
		__antithesis_instrumentation__.Notify(10969)
		mem.Shrink(ctx, memSize)
	}()
	__antithesis_instrumentation__.Notify(10942)

	backupCodec := keys.SystemSQLCodec
	backupTenantID := roachpb.SystemTenantID

	if len(sqlDescs) != 0 {
		__antithesis_instrumentation__.Notify(10970)
		if len(latestBackupManifest.Spans) != 0 && func() bool {
			__antithesis_instrumentation__.Notify(10971)
			return !latestBackupManifest.HasTenants() == true
		}() == true {
			__antithesis_instrumentation__.Notify(10972)

			_, backupTenantID, err = keys.DecodeTenantPrefix(latestBackupManifest.Spans[0].Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(10974)
				return err
			} else {
				__antithesis_instrumentation__.Notify(10975)
			}
			__antithesis_instrumentation__.Notify(10973)
			backupCodec = keys.MakeSQLCodec(backupTenantID)
		} else {
			__antithesis_instrumentation__.Notify(10976)
		}
	} else {
		__antithesis_instrumentation__.Notify(10977)
	}
	__antithesis_instrumentation__.Notify(10943)

	lastBackupIndex, err := getBackupIndexAtTime(backupManifests, details.EndTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(10978)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10979)
	}
	__antithesis_instrumentation__.Notify(10944)
	defaultConf, err := cloud.ExternalStorageConfFromURI(details.URIs[lastBackupIndex], p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(10980)
		return errors.Wrapf(err, "creating external store configuration")
	} else {
		__antithesis_instrumentation__.Notify(10981)
	}
	__antithesis_instrumentation__.Notify(10945)
	defaultStore, err := p.ExecCfg().DistSQLSrv.ExternalStorage(ctx, defaultConf)
	if err != nil {
		__antithesis_instrumentation__.Notify(10982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10983)
	}
	__antithesis_instrumentation__.Notify(10946)

	preData, mainData, err := createImportingDescriptors(ctx, p, backupCodec, sqlDescs, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(10984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(10985)
	}
	__antithesis_instrumentation__.Notify(10947)

	if !backupCodec.TenantPrefix().Equal(p.ExecCfg().Codec.TenantPrefix()) {
		__antithesis_instrumentation__.Notify(10986)

		if backupTenantID != roachpb.SystemTenantID && func() bool {
			__antithesis_instrumentation__.Notify(10987)
			return p.ExecCfg().Codec.ForSystemTenant() == true
		}() == true {
			__antithesis_instrumentation__.Notify(10988)

			preData.tableRekeys = append(preData.tableRekeys, execinfrapb.TableRekey{})
			mainData.tableRekeys = append(preData.tableRekeys, execinfrapb.TableRekey{})
		} else {
			__antithesis_instrumentation__.Notify(10989)
		}
	} else {
		__antithesis_instrumentation__.Notify(10990)
	}
	__antithesis_instrumentation__.Notify(10948)

	if backupTenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(10991)
		preData.tenantRekeys = append(preData.tenantRekeys, isBackupFromSystemTenantRekey)
		mainData.tenantRekeys = append(preData.tenantRekeys, isBackupFromSystemTenantRekey)
	} else {
		__antithesis_instrumentation__.Notify(10992)
	}
	__antithesis_instrumentation__.Notify(10949)

	details = r.job.Details().(jobspb.RestoreDetails)

	if fn := r.testingKnobs.afterOfflineTableCreation; fn != nil {
		__antithesis_instrumentation__.Notify(10993)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(10994)
			return err
		} else {
			__antithesis_instrumentation__.Notify(10995)
		}
	} else {
		__antithesis_instrumentation__.Notify(10996)
	}
	__antithesis_instrumentation__.Notify(10950)
	var remappedStats []*stats.TableStatisticProto
	backupStats, err := getStatisticsFromBackup(ctx, defaultStore, details.Encryption,
		latestBackupManifest)
	if err == nil {
		__antithesis_instrumentation__.Notify(10997)
		remappedStats = remapRelevantStatistics(ctx, backupStats, details.DescriptorRewrites,
			details.TableDescs)
	} else {
		__antithesis_instrumentation__.Notify(10998)

		log.Warningf(ctx, "failed to resolve table statistics from backup during restore: %+v",
			err.Error())
	}
	__antithesis_instrumentation__.Notify(10951)

	if len(details.TableDescs) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(10999)
		return len(details.Tenants) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(11000)
		return len(details.TypeDescs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(11001)

		log.Warning(ctx, "nothing to restore")

		publishDescriptors := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) (err error) {
			__antithesis_instrumentation__.Notify(11005)
			return r.publishDescriptors(ctx, txn, p.ExecCfg(), p.User(), descsCol, details, nil)
		}
		__antithesis_instrumentation__.Notify(11002)
		if err := sql.DescsTxn(ctx, r.execCfg, publishDescriptors); err != nil {
			__antithesis_instrumentation__.Notify(11006)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11007)
		}
		__antithesis_instrumentation__.Notify(11003)

		p.ExecCfg().JobRegistry.NotifyToAdoptJobs()
		if fn := r.testingKnobs.afterPublishingDescriptors; fn != nil {
			__antithesis_instrumentation__.Notify(11008)
			if err := fn(); err != nil {
				__antithesis_instrumentation__.Notify(11009)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11010)
			}
		} else {
			__antithesis_instrumentation__.Notify(11011)
		}
		__antithesis_instrumentation__.Notify(11004)
		emitRestoreJobEvent(ctx, p, jobs.StatusSucceeded, r.job)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(11012)
	}
	__antithesis_instrumentation__.Notify(10952)

	for _, tenant := range details.Tenants {
		__antithesis_instrumentation__.Notify(11013)
		to := roachpb.MakeTenantID(tenant.ID)
		from := to
		if details.PreRewriteTenantId != nil {
			__antithesis_instrumentation__.Notify(11015)
			from = *details.PreRewriteTenantId
		} else {
			__antithesis_instrumentation__.Notify(11016)
		}
		__antithesis_instrumentation__.Notify(11014)
		mainData.addTenant(from, to)
	}
	__antithesis_instrumentation__.Notify(10953)

	numNodes, err := clusterNodeCount(p.ExecCfg().Gossip)
	if err != nil {
		__antithesis_instrumentation__.Notify(11017)
		if !build.IsRelease() && func() bool {
			__antithesis_instrumentation__.Notify(11019)
			return p.ExecCfg().Codec.ForSystemTenant() == true
		}() == true {
			__antithesis_instrumentation__.Notify(11020)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11021)
		}
		__antithesis_instrumentation__.Notify(11018)
		log.Warningf(ctx, "unable to determine cluster node count: %v", err)
		numNodes = 1
	} else {
		__antithesis_instrumentation__.Notify(11022)
	}
	__antithesis_instrumentation__.Notify(10954)

	var resTotal roachpb.RowCount
	if !preData.isEmpty() {
		__antithesis_instrumentation__.Notify(11023)
		res, err := restoreWithRetry(
			ctx,
			p,
			numNodes,
			backupManifests,
			details.BackupLocalityInfo,
			details.EndTime,
			preData,
			r.job,
			details.Encryption,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(11026)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11027)
		}
		__antithesis_instrumentation__.Notify(11024)

		resTotal.Add(res)

		if details.DescriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11028)
			if err := r.restoreSystemTables(ctx, p.ExecCfg().DB, preData.systemTables); err != nil {
				__antithesis_instrumentation__.Notify(11030)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11031)
			}
			__antithesis_instrumentation__.Notify(11029)

			details = r.job.Details().(jobspb.RestoreDetails)
		} else {
			__antithesis_instrumentation__.Notify(11032)
		}
		__antithesis_instrumentation__.Notify(11025)

		if fn := r.testingKnobs.afterPreRestore; fn != nil {
			__antithesis_instrumentation__.Notify(11033)
			if err := fn(); err != nil {
				__antithesis_instrumentation__.Notify(11034)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11035)
			}
		} else {
			__antithesis_instrumentation__.Notify(11036)
		}
	} else {
		__antithesis_instrumentation__.Notify(11037)
	}

	{
		__antithesis_instrumentation__.Notify(11038)

		res, err := restoreWithRetry(
			ctx,
			p,
			numNodes,
			backupManifests,
			details.BackupLocalityInfo,
			details.EndTime,
			mainData,
			r.job,
			details.Encryption,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(11040)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11041)
		}
		__antithesis_instrumentation__.Notify(11039)

		resTotal.Add(res)
	}
	__antithesis_instrumentation__.Notify(10955)

	if err := insertStats(ctx, r.job, p.ExecCfg(), remappedStats); err != nil {
		__antithesis_instrumentation__.Notify(11042)
		return errors.Wrap(err, "inserting table statistics")
	} else {
		__antithesis_instrumentation__.Notify(11043)
	}
	__antithesis_instrumentation__.Notify(10956)

	var devalidateIndexes map[descpb.ID][]descpb.IndexID
	if toValidate := len(details.RevalidateIndexes); toValidate > 0 {
		__antithesis_instrumentation__.Notify(11044)
		if err := r.job.RunningStatus(ctx, nil, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(11047)
			return jobs.RunningStatus(fmt.Sprintf("re-validating %d indexes", toValidate)), nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(11048)
			return errors.Wrapf(err, "failed to update running status of job %d", errors.Safe(r.job.ID()))
		} else {
			__antithesis_instrumentation__.Notify(11049)
		}
		__antithesis_instrumentation__.Notify(11045)
		bad, err := revalidateIndexes(ctx, p.ExecCfg(), r.job, details.TableDescs, details.RevalidateIndexes)
		if err != nil {
			__antithesis_instrumentation__.Notify(11050)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11051)
		}
		__antithesis_instrumentation__.Notify(11046)
		devalidateIndexes = bad
	} else {
		__antithesis_instrumentation__.Notify(11052)
	}
	__antithesis_instrumentation__.Notify(10957)

	publishDescriptors := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) (err error) {
		__antithesis_instrumentation__.Notify(11053)
		err = r.publishDescriptors(ctx, txn, p.ExecCfg(), p.User(), descsCol, details, devalidateIndexes)
		return err
	}
	__antithesis_instrumentation__.Notify(10958)
	if err := sql.DescsTxn(ctx, p.ExecCfg(), publishDescriptors); err != nil {
		__antithesis_instrumentation__.Notify(11054)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11055)
	}
	__antithesis_instrumentation__.Notify(10959)

	details = r.job.Details().(jobspb.RestoreDetails)
	p.ExecCfg().JobRegistry.NotifyToAdoptJobs()

	if details.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(11056)

		if err := r.restoreSystemTables(ctx, p.ExecCfg().DB, mainData.systemTables); err != nil {
			__antithesis_instrumentation__.Notify(11058)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11059)
		}
		__antithesis_instrumentation__.Notify(11057)

		details = r.job.Details().(jobspb.RestoreDetails)

		if err := r.cleanupTempSystemTables(ctx, nil); err != nil {
			__antithesis_instrumentation__.Notify(11060)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11061)
		}
	} else {
		__antithesis_instrumentation__.Notify(11062)
		if details.RestoreSystemUsers {
			__antithesis_instrumentation__.Notify(11063)
			if err := r.restoreSystemUsers(ctx, p.ExecCfg().DB, mainData.systemTables); err != nil {
				__antithesis_instrumentation__.Notify(11065)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11066)
			}
			__antithesis_instrumentation__.Notify(11064)
			details = r.job.Details().(jobspb.RestoreDetails)

			if err := r.cleanupTempSystemTables(ctx, nil); err != nil {
				__antithesis_instrumentation__.Notify(11067)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11068)
			}
		} else {
			__antithesis_instrumentation__.Notify(11069)
		}
	}
	__antithesis_instrumentation__.Notify(10960)

	if fn := r.testingKnobs.afterPublishingDescriptors; fn != nil {
		__antithesis_instrumentation__.Notify(11070)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(11071)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11072)
		}
	} else {
		__antithesis_instrumentation__.Notify(11073)
	}
	__antithesis_instrumentation__.Notify(10961)

	r.notifyStatsRefresherOfNewTables()

	r.restoreStats = resTotal

	emitRestoreJobEvent(ctx, p, jobs.StatusSucceeded, r.job)

	{
		__antithesis_instrumentation__.Notify(11074)
		telemetry.Count("restore.total.succeeded")
		const mb = 1 << 20
		sizeMb := resTotal.DataSize / mb
		sec := int64(timeutil.Since(timeutil.FromUnixMicros(r.job.Payload().StartedMicros)).Seconds())
		var mbps int64
		if sec > 0 {
			__antithesis_instrumentation__.Notify(11076)
			mbps = mb / sec
		} else {
			__antithesis_instrumentation__.Notify(11077)
		}
		__antithesis_instrumentation__.Notify(11075)
		telemetry.CountBucketed("restore.duration-sec.succeeded", sec)
		telemetry.CountBucketed("restore.size-mb.full", sizeMb)
		telemetry.CountBucketed("restore.speed-mbps.total", mbps)
		telemetry.CountBucketed("restore.speed-mbps.per-node", mbps/int64(numNodes))

		if sizeMb > 10 {
			__antithesis_instrumentation__.Notify(11078)
			telemetry.CountBucketed("restore.speed-mbps.over10mb", mbps)
			telemetry.CountBucketed("restore.speed-mbps.over10mb.per-node", mbps/int64(numNodes))
		} else {
			__antithesis_instrumentation__.Notify(11079)
		}
	}
	__antithesis_instrumentation__.Notify(10962)
	return nil
}

func revalidateIndexes(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	job *jobs.Job,
	tables []*descpb.TableDescriptor,
	indexIDs []jobspb.RestoreDetails_RevalidateIndex,
) (map[descpb.ID][]descpb.IndexID, error) {
	__antithesis_instrumentation__.Notify(11080)
	indexIDsByTable := make(map[descpb.ID]map[descpb.IndexID]struct{})
	for _, idx := range indexIDs {
		__antithesis_instrumentation__.Notify(11084)
		if indexIDsByTable[idx.TableID] == nil {
			__antithesis_instrumentation__.Notify(11086)
			indexIDsByTable[idx.TableID] = make(map[descpb.IndexID]struct{})
		} else {
			__antithesis_instrumentation__.Notify(11087)
		}
		__antithesis_instrumentation__.Notify(11085)
		indexIDsByTable[idx.TableID][idx.IndexID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(11081)

	var runner sqlutil.HistoricalInternalExecTxnRunner = func(ctx context.Context, fn sqlutil.InternalExecFn) error {
		__antithesis_instrumentation__.Notify(11088)
		return execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(11089)
			ie := job.MakeSessionBoundInternalExecutor(ctx, sql.NewFakeSessionData(execCfg.SV())).(*sql.InternalExecutor)
			return fn(ctx, txn, ie)
		})
	}
	__antithesis_instrumentation__.Notify(11082)

	invalidIndexes := make(map[descpb.ID][]descpb.IndexID)

	for _, tbl := range tables {
		__antithesis_instrumentation__.Notify(11090)
		indexes := indexIDsByTable[tbl.ID]
		if len(indexes) == 0 {
			__antithesis_instrumentation__.Notify(11094)
			continue
		} else {
			__antithesis_instrumentation__.Notify(11095)
		}
		__antithesis_instrumentation__.Notify(11091)
		tableDesc := tabledesc.NewBuilder(tbl).BuildExistingMutableTable()

		var forward, inverted []catalog.Index
		for _, idx := range tableDesc.AllIndexes() {
			__antithesis_instrumentation__.Notify(11096)
			if _, ok := indexes[idx.GetID()]; ok {
				__antithesis_instrumentation__.Notify(11097)
				switch idx.GetType() {
				case descpb.IndexDescriptor_FORWARD:
					__antithesis_instrumentation__.Notify(11098)
					forward = append(forward, idx)
				case descpb.IndexDescriptor_INVERTED:
					__antithesis_instrumentation__.Notify(11099)
					inverted = append(inverted, idx)
				default:
					__antithesis_instrumentation__.Notify(11100)
				}
			} else {
				__antithesis_instrumentation__.Notify(11101)
			}
		}
		__antithesis_instrumentation__.Notify(11092)
		if len(forward) > 0 {
			__antithesis_instrumentation__.Notify(11102)
			if err := sql.ValidateForwardIndexes(
				ctx,
				tableDesc.MakePublic(),
				forward,
				runner,
				false,
				true,
				sessiondata.InternalExecutorOverride{},
			); err != nil {
				__antithesis_instrumentation__.Notify(11103)
				if invalid := (sql.InvalidIndexesError{}); errors.As(err, &invalid) {
					__antithesis_instrumentation__.Notify(11104)
					invalidIndexes[tableDesc.ID] = invalid.Indexes
				} else {
					__antithesis_instrumentation__.Notify(11105)
					return nil, err
				}
			} else {
				__antithesis_instrumentation__.Notify(11106)
			}
		} else {
			__antithesis_instrumentation__.Notify(11107)
		}
		__antithesis_instrumentation__.Notify(11093)
		if len(inverted) > 0 {
			__antithesis_instrumentation__.Notify(11108)
			if err := sql.ValidateInvertedIndexes(
				ctx,
				execCfg.Codec,
				tableDesc.MakePublic(),
				inverted,
				runner,
				false,
				true,
				sessiondata.InternalExecutorOverride{},
			); err != nil {
				__antithesis_instrumentation__.Notify(11109)
				if invalid := (sql.InvalidIndexesError{}); errors.As(err, &invalid) {
					__antithesis_instrumentation__.Notify(11110)
					invalidIndexes[tableDesc.ID] = append(invalidIndexes[tableDesc.ID], invalid.Indexes...)
				} else {
					__antithesis_instrumentation__.Notify(11111)
					return nil, err
				}
			} else {
				__antithesis_instrumentation__.Notify(11112)
			}
		} else {
			__antithesis_instrumentation__.Notify(11113)
		}
	}
	__antithesis_instrumentation__.Notify(11083)
	return invalidIndexes, nil
}

func (r *restoreResumer) ReportResults(ctx context.Context, resultsCh chan<- tree.Datums) error {
	__antithesis_instrumentation__.Notify(11114)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(11115)
		return ctx.Err()
	case resultsCh <- tree.Datums{
		tree.NewDInt(tree.DInt(r.job.ID())),
		tree.NewDString(string(jobs.StatusSucceeded)),
		tree.NewDFloat(tree.DFloat(1.0)),
		tree.NewDInt(tree.DInt(r.restoreStats.Rows)),
		tree.NewDInt(tree.DInt(r.restoreStats.IndexEntries)),
		tree.NewDInt(tree.DInt(r.restoreStats.DataSize)),
	}:
		__antithesis_instrumentation__.Notify(11116)
		return nil
	}
}

func (r *restoreResumer) notifyStatsRefresherOfNewTables() {
	__antithesis_instrumentation__.Notify(11117)
	details := r.job.Details().(jobspb.RestoreDetails)
	for i := range details.TableDescs {
		__antithesis_instrumentation__.Notify(11118)
		desc := tabledesc.NewBuilder(details.TableDescs[i]).BuildImmutableTable()
		r.execCfg.StatsRefresher.NotifyMutation(desc, math.MaxInt32)
	}
}

func tempSystemDatabaseID(details jobspb.RestoreDetails) descpb.ID {
	__antithesis_instrumentation__.Notify(11119)
	if details.DescriptorCoverage != tree.AllDescriptors && func() bool {
		__antithesis_instrumentation__.Notify(11122)
		return !details.RestoreSystemUsers == true
	}() == true {
		__antithesis_instrumentation__.Notify(11123)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(11124)
	}
	__antithesis_instrumentation__.Notify(11120)
	var maxPreAllocatedID descpb.ID
	for id := range details.DescriptorRewrites {
		__antithesis_instrumentation__.Notify(11125)

		if id > maxPreAllocatedID {
			__antithesis_instrumentation__.Notify(11126)
			maxPreAllocatedID = id
		} else {
			__antithesis_instrumentation__.Notify(11127)
		}
	}
	__antithesis_instrumentation__.Notify(11121)
	return maxPreAllocatedID
}

func insertStats(
	ctx context.Context,
	job *jobs.Job,
	execCfg *sql.ExecutorConfig,
	latestStats []*stats.TableStatisticProto,
) error {
	__antithesis_instrumentation__.Notify(11128)
	details := job.Details().(jobspb.RestoreDetails)
	if details.StatsInserted {
		__antithesis_instrumentation__.Notify(11130)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(11131)
	}
	__antithesis_instrumentation__.Notify(11129)

	for {
		__antithesis_instrumentation__.Notify(11132)
		if len(latestStats) == 0 {
			__antithesis_instrumentation__.Notify(11136)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(11137)
		}
		__antithesis_instrumentation__.Notify(11133)

		if len(latestStats) < restoreStatsInsertBatchSize {
			__antithesis_instrumentation__.Notify(11138)
			restoreStatsInsertBatchSize = len(latestStats)
		} else {
			__antithesis_instrumentation__.Notify(11139)
		}
		__antithesis_instrumentation__.Notify(11134)

		if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(11140)
			if err := stats.InsertNewStats(ctx, execCfg.Settings, execCfg.InternalExecutor, txn,
				latestStats[:restoreStatsInsertBatchSize]); err != nil {
				__antithesis_instrumentation__.Notify(11143)
				return errors.Wrapf(err, "inserting stats from backup")
			} else {
				__antithesis_instrumentation__.Notify(11144)
			}
			__antithesis_instrumentation__.Notify(11141)

			if restoreStatsInsertBatchSize == len(latestStats) {
				__antithesis_instrumentation__.Notify(11145)
				details.StatsInserted = true
				if err := job.SetDetails(ctx, txn, details); err != nil {
					__antithesis_instrumentation__.Notify(11146)
					return errors.Wrapf(err, "updating job marking stats insertion complete")
				} else {
					__antithesis_instrumentation__.Notify(11147)
				}
			} else {
				__antithesis_instrumentation__.Notify(11148)
			}
			__antithesis_instrumentation__.Notify(11142)

			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(11149)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11150)
		}
		__antithesis_instrumentation__.Notify(11135)

		latestStats = latestStats[restoreStatsInsertBatchSize:]
	}
}

func (r *restoreResumer) publishDescriptors(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *sql.ExecutorConfig,
	user security.SQLUsername,
	descsCol *descs.Collection,
	details jobspb.RestoreDetails,
	devalidateIndexes map[descpb.ID][]descpb.IndexID,
) (err error) {
	__antithesis_instrumentation__.Notify(11151)
	if details.DescriptorsPublished {
		__antithesis_instrumentation__.Notify(11164)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(11165)
	}
	__antithesis_instrumentation__.Notify(11152)
	if fn := r.testingKnobs.beforePublishingDescriptors; fn != nil {
		__antithesis_instrumentation__.Notify(11166)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(11167)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11168)
		}
	} else {
		__antithesis_instrumentation__.Notify(11169)
	}
	__antithesis_instrumentation__.Notify(11153)
	log.VEventf(ctx, 1, "making tables live")

	all, err := prefetchDescriptors(ctx, txn, descsCol, details)
	if err != nil {
		__antithesis_instrumentation__.Notify(11170)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11171)
	}
	__antithesis_instrumentation__.Notify(11154)

	newTables := make([]*descpb.TableDescriptor, 0, len(details.TableDescs))
	newTypes := make([]*descpb.TypeDescriptor, 0, len(details.TypeDescs))
	newSchemas := make([]*descpb.SchemaDescriptor, 0, len(details.SchemaDescs))
	newDBs := make([]*descpb.DatabaseDescriptor, 0, len(details.DatabaseDescs))

	if details.DescriptorCoverage != tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(11172)
		if err := scbackup.CreateDeclarativeSchemaChangeJobs(
			ctx, r.execCfg.JobRegistry, txn, all,
		); err != nil {
			__antithesis_instrumentation__.Notify(11173)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11174)
		}
	} else {
		__antithesis_instrumentation__.Notify(11175)
	}
	__antithesis_instrumentation__.Notify(11155)

	for i := range details.TableDescs {
		__antithesis_instrumentation__.Notify(11176)
		mutTable := all.LookupDescriptorEntry(details.TableDescs[i].GetID()).(*tabledesc.Mutable)

		if mutTable.GetDeclarativeSchemaChangerState() != nil {
			__antithesis_instrumentation__.Notify(11181)
			newTables = append(newTables, mutTable.TableDesc())
			continue
		} else {
			__antithesis_instrumentation__.Notify(11182)
		}
		__antithesis_instrumentation__.Notify(11177)

		badIndexes := devalidateIndexes[mutTable.ID]
		for _, badIdx := range badIndexes {
			__antithesis_instrumentation__.Notify(11183)
			found, err := mutTable.FindIndexWithID(badIdx)
			if err != nil {
				__antithesis_instrumentation__.Notify(11185)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11186)
			}
			__antithesis_instrumentation__.Notify(11184)
			newIdx := found.IndexDescDeepCopy()
			mutTable.RemovePublicNonPrimaryIndex(found.Ordinal())
			if err := mutTable.AddIndexMutation(ctx, &newIdx, descpb.DescriptorMutation_ADD, r.settings); err != nil {
				__antithesis_instrumentation__.Notify(11187)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11188)
			}
		}
		__antithesis_instrumentation__.Notify(11178)
		version := r.settings.Version.ActiveVersion(ctx)
		if err := mutTable.AllocateIDs(ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(11189)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11190)
		}
		__antithesis_instrumentation__.Notify(11179)

		if details.DescriptorCoverage != tree.AllDescriptors && func() bool {
			__antithesis_instrumentation__.Notify(11191)
			return mutTable.HasRowLevelTTL() == true
		}() == true {
			__antithesis_instrumentation__.Notify(11192)
			j, err := sql.CreateRowLevelTTLScheduledJob(
				ctx,
				execCfg,
				txn,
				user,
				mutTable.GetID(),
				mutTable.GetRowLevelTTL(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(11194)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11195)
			}
			__antithesis_instrumentation__.Notify(11193)
			mutTable.RowLevelTTL.ScheduleID = j.ScheduleID()
		} else {
			__antithesis_instrumentation__.Notify(11196)
		}
		__antithesis_instrumentation__.Notify(11180)
		newTables = append(newTables, mutTable.TableDesc())

		if details.DescriptorCoverage != tree.AllDescriptors || func() bool {
			__antithesis_instrumentation__.Notify(11197)
			return len(badIndexes) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(11198)

			if err := createSchemaChangeJobsFromMutations(ctx,
				r.execCfg.JobRegistry, r.execCfg.Codec, txn, r.job.Payload().UsernameProto.Decode(), mutTable,
			); err != nil {
				__antithesis_instrumentation__.Notify(11199)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11200)
			}
		} else {
			__antithesis_instrumentation__.Notify(11201)
		}
	}
	__antithesis_instrumentation__.Notify(11156)

	for i := range details.TypeDescs {
		__antithesis_instrumentation__.Notify(11202)
		typ := all.LookupDescriptorEntry(details.TypeDescs[i].GetID()).(catalog.TypeDescriptor)
		newTypes = append(newTypes, typ.TypeDesc())
		if typ.GetDeclarativeSchemaChangerState() == nil && func() bool {
			__antithesis_instrumentation__.Notify(11203)
			return typ.HasPendingSchemaChanges() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(11204)
			return details.DescriptorCoverage != tree.AllDescriptors == true
		}() == true {
			__antithesis_instrumentation__.Notify(11205)
			if err := createTypeChangeJobFromDesc(
				ctx, r.execCfg.JobRegistry, r.execCfg.Codec, txn, r.job.Payload().UsernameProto.Decode(), typ,
			); err != nil {
				__antithesis_instrumentation__.Notify(11206)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11207)
			}
		} else {
			__antithesis_instrumentation__.Notify(11208)
		}
	}
	__antithesis_instrumentation__.Notify(11157)
	for i := range details.SchemaDescs {
		__antithesis_instrumentation__.Notify(11209)
		sc := all.LookupDescriptorEntry(details.SchemaDescs[i].GetID()).(catalog.SchemaDescriptor)
		newSchemas = append(newSchemas, sc.SchemaDesc())
	}
	__antithesis_instrumentation__.Notify(11158)
	for i := range details.DatabaseDescs {
		__antithesis_instrumentation__.Notify(11210)
		db := all.LookupDescriptorEntry(details.DatabaseDescs[i].GetID()).(catalog.DatabaseDescriptor)
		newDBs = append(newDBs, db.DatabaseDesc())
	}
	__antithesis_instrumentation__.Notify(11159)
	b := txn.NewBatch()
	if err := all.ForEachDescriptorEntry(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(11211)
		d := desc.(catalog.MutableDescriptor)
		d.SetPublic()
		return descsCol.WriteDescToBatch(
			ctx, false, d, b,
		)
	}); err != nil {
		__antithesis_instrumentation__.Notify(11212)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11213)
	}
	__antithesis_instrumentation__.Notify(11160)
	if err := txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(11214)
		return errors.Wrap(err, "publishing tables")
	} else {
		__antithesis_instrumentation__.Notify(11215)
	}
	__antithesis_instrumentation__.Notify(11161)

	for _, tenant := range details.Tenants {
		__antithesis_instrumentation__.Notify(11216)
		if err := sql.ActivateTenant(ctx, r.execCfg, txn, tenant.ID); err != nil {
			__antithesis_instrumentation__.Notify(11217)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11218)
		}
	}
	__antithesis_instrumentation__.Notify(11162)

	details.DescriptorsPublished = true
	details.TableDescs = newTables
	details.TypeDescs = newTypes
	details.SchemaDescs = newSchemas
	details.DatabaseDescs = newDBs
	if err := r.job.SetDetails(ctx, txn, details); err != nil {
		__antithesis_instrumentation__.Notify(11219)
		return errors.Wrap(err,
			"updating job details after publishing tables")
	} else {
		__antithesis_instrumentation__.Notify(11220)
	}
	__antithesis_instrumentation__.Notify(11163)
	return nil
}

func prefetchDescriptors(
	ctx context.Context, txn *kv.Txn, descsCol *descs.Collection, details jobspb.RestoreDetails,
) (_ nstree.Catalog, _ error) {
	__antithesis_instrumentation__.Notify(11221)
	var all nstree.MutableCatalog
	var allDescIDs catalog.DescriptorIDSet
	expVersion := map[descpb.ID]descpb.DescriptorVersion{}
	for i := range details.TableDescs {
		__antithesis_instrumentation__.Notify(11228)
		expVersion[details.TableDescs[i].GetID()] = details.TableDescs[i].GetVersion()
		allDescIDs.Add(details.TableDescs[i].GetID())
	}
	__antithesis_instrumentation__.Notify(11222)
	for i := range details.TypeDescs {
		__antithesis_instrumentation__.Notify(11229)
		expVersion[details.TypeDescs[i].GetID()] = details.TypeDescs[i].GetVersion()
		allDescIDs.Add(details.TypeDescs[i].GetID())
	}
	__antithesis_instrumentation__.Notify(11223)
	for i := range details.SchemaDescs {
		__antithesis_instrumentation__.Notify(11230)
		expVersion[details.SchemaDescs[i].GetID()] = details.SchemaDescs[i].GetVersion()
		allDescIDs.Add(details.SchemaDescs[i].GetID())
	}
	__antithesis_instrumentation__.Notify(11224)
	for i := range details.DatabaseDescs {
		__antithesis_instrumentation__.Notify(11231)
		expVersion[details.DatabaseDescs[i].GetID()] = details.DatabaseDescs[i].GetVersion()
		allDescIDs.Add(details.DatabaseDescs[i].GetID())
	}
	__antithesis_instrumentation__.Notify(11225)

	ids := allDescIDs.Ordered()
	got, err := descsCol.GetMutableDescriptorsByID(ctx, txn, ids...)
	if err != nil {
		__antithesis_instrumentation__.Notify(11232)
		return nstree.Catalog{}, errors.Wrap(err, "prefetch descriptors")
	} else {
		__antithesis_instrumentation__.Notify(11233)
	}
	__antithesis_instrumentation__.Notify(11226)
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(11234)
		if got[i].GetVersion() != expVersion[id] {
			__antithesis_instrumentation__.Notify(11236)
			return nstree.Catalog{}, errors.Errorf(
				"version mismatch for descriptor %d, expected version %d, got %v",
				got[i].GetID(), got[i].GetVersion(), expVersion[id],
			)
		} else {
			__antithesis_instrumentation__.Notify(11237)
		}
		__antithesis_instrumentation__.Notify(11235)
		all.UpsertDescriptorEntry(got[i])
	}
	__antithesis_instrumentation__.Notify(11227)
	return all.Catalog, nil
}

func emitRestoreJobEvent(
	ctx context.Context, p sql.JobExecContext, status jobs.Status, job *jobs.Job,
) {
	__antithesis_instrumentation__.Notify(11238)

	var restoreEvent eventpb.Restore
	if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(11239)
		return sql.LogEventForJobs(ctx, p.ExecCfg(), txn, &restoreEvent, int64(job.ID()),
			job.Payload(), p.User(), status)
	}); err != nil {
		__antithesis_instrumentation__.Notify(11240)
		log.Warningf(ctx, "failed to log event: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(11241)
	}
}

func (r *restoreResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(11242)
	p := execCtx.(sql.JobExecContext)
	r.execCfg = p.ExecCfg()

	emitRestoreJobEvent(ctx, p, jobs.StatusReverting, r.job)

	telemetry.Count("restore.total.failed")
	telemetry.CountBucketed("restore.duration-sec.failed",
		int64(timeutil.Since(timeutil.FromUnixMicros(r.job.Payload().StartedMicros)).Seconds()))

	details := r.job.Details().(jobspb.RestoreDetails)

	execCfg := execCtx.(sql.JobExecContext).ExecCfg()
	if err := sql.DescsTxn(ctx, execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(11245)
		for _, tenant := range details.Tenants {
			__antithesis_instrumentation__.Notify(11249)
			tenant.State = descpb.TenantInfo_DROP

			if err := sql.GCTenantSync(ctx, execCfg, &tenant.TenantInfo); err != nil {
				__antithesis_instrumentation__.Notify(11250)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11251)
			}
		}
		__antithesis_instrumentation__.Notify(11246)

		if err := r.dropDescriptors(ctx, execCfg.JobRegistry, execCfg.Codec, txn, descsCol); err != nil {
			__antithesis_instrumentation__.Notify(11252)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11253)
		}
		__antithesis_instrumentation__.Notify(11247)

		if details.DescriptorCoverage == tree.AllDescriptors {
			__antithesis_instrumentation__.Notify(11254)

			ie := p.ExecCfg().InternalExecutor
			_, err := ie.Exec(ctx, "recreate-defaultdb", txn, "CREATE DATABASE IF NOT EXISTS defaultdb")
			if err != nil {
				__antithesis_instrumentation__.Notify(11256)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11257)
			}
			__antithesis_instrumentation__.Notify(11255)

			_, err = ie.Exec(ctx, "recreate-postgres", txn, "CREATE DATABASE IF NOT EXISTS postgres")
			if err != nil {
				__antithesis_instrumentation__.Notify(11258)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11259)
			}
		} else {
			__antithesis_instrumentation__.Notify(11260)
		}
		__antithesis_instrumentation__.Notify(11248)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(11261)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11262)
	}
	__antithesis_instrumentation__.Notify(11243)

	if details.DescriptorCoverage == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(11263)

		if err := execCfg.DB.Txn(ctx, r.cleanupTempSystemTables); err != nil {
			__antithesis_instrumentation__.Notify(11264)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11265)
		}
	} else {
		__antithesis_instrumentation__.Notify(11266)
	}
	__antithesis_instrumentation__.Notify(11244)

	emitRestoreJobEvent(ctx, p, jobs.StatusFailed, r.job)
	return nil
}

func (r *restoreResumer) dropDescriptors(
	ctx context.Context,
	jr *jobs.Registry,
	codec keys.SQLCodec,
	txn *kv.Txn,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(11267)
	details := r.job.Details().(jobspb.RestoreDetails)

	if !details.PrepareCompleted {
		__antithesis_instrumentation__.Notify(11284)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(11285)
	}
	__antithesis_instrumentation__.Notify(11268)

	b := txn.NewBatch()

	mutableTables := make([]*tabledesc.Mutable, len(details.TableDescs))
	for i := range details.TableDescs {
		__antithesis_instrumentation__.Notify(11286)
		var err error
		mutableTables[i], err = descsCol.GetMutableTableVersionByID(ctx, details.TableDescs[i].ID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(11288)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11289)
		}
		__antithesis_instrumentation__.Notify(11287)

		if !details.DescriptorsPublished {
			__antithesis_instrumentation__.Notify(11290)
			if got, exp := mutableTables[i].Version, details.TableDescs[i].Version; got != exp {
				__antithesis_instrumentation__.Notify(11291)
				log.Errorf(ctx, "version changed for restored descriptor %d before "+
					"drop: got %d, expected %d", mutableTables[i].GetVersion(), got, exp)
			} else {
				__antithesis_instrumentation__.Notify(11292)
			}
		} else {
			__antithesis_instrumentation__.Notify(11293)
		}

	}
	__antithesis_instrumentation__.Notify(11269)

	if err := r.removeExistingTypeBackReferences(
		ctx, txn, descsCol, b, mutableTables, &details,
	); err != nil {
		__antithesis_instrumentation__.Notify(11294)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11295)
	}
	__antithesis_instrumentation__.Notify(11270)

	tablesToGC := make([]descpb.ID, 0, len(details.TableDescs))

	dropTime := int64(1)
	for i := range mutableTables {
		__antithesis_instrumentation__.Notify(11296)
		tableToDrop := mutableTables[i]
		tablesToGC = append(tablesToGC, tableToDrop.ID)
		tableToDrop.SetDropped()

		if tableToDrop.HasRowLevelTTL() {
			__antithesis_instrumentation__.Notify(11298)
			scheduleID := tableToDrop.RowLevelTTL.ScheduleID
			if scheduleID != 0 {
				__antithesis_instrumentation__.Notify(11299)
				if err := sql.DeleteSchedule(
					ctx, r.execCfg, txn, scheduleID,
				); err != nil {
					__antithesis_instrumentation__.Notify(11300)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11301)
				}
			} else {
				__antithesis_instrumentation__.Notify(11302)
			}
		} else {
			__antithesis_instrumentation__.Notify(11303)
		}
		__antithesis_instrumentation__.Notify(11297)

		tableToDrop.DropTime = dropTime
		b.Del(catalogkeys.EncodeNameKey(codec, tableToDrop))
		descsCol.AddDeletedDescriptor(tableToDrop.GetID())
	}
	__antithesis_instrumentation__.Notify(11271)

	for i := range details.TypeDescs {
		__antithesis_instrumentation__.Notify(11304)

		typDesc := details.TypeDescs[i]
		mutType, err := descsCol.GetMutableTypeByID(ctx, txn, typDesc.ID, tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				AvoidLeased:    true,
				IncludeOffline: true,
			},
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(11306)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11307)
		}
		__antithesis_instrumentation__.Notify(11305)

		b.Del(catalogkeys.EncodeNameKey(codec, typDesc))
		mutType.SetDropped()

		b.Del(catalogkeys.MakeDescMetadataKey(codec, typDesc.ID))
		descsCol.AddDeletedDescriptor(mutType.GetID())
	}
	__antithesis_instrumentation__.Notify(11272)

	gcDetails := jobspb.SchemaChangeGCDetails{}
	for _, tableID := range tablesToGC {
		__antithesis_instrumentation__.Notify(11308)
		gcDetails.Tables = append(gcDetails.Tables, jobspb.SchemaChangeGCDetails_DroppedID{
			ID:       tableID,
			DropTime: dropTime,
		})
	}
	__antithesis_instrumentation__.Notify(11273)
	gcJobRecord := jobs.Record{
		Description:   fmt.Sprintf("GC for %s", r.job.Payload().Description),
		Username:      r.job.Payload().UsernameProto.Decode(),
		DescriptorIDs: tablesToGC,
		Details:       gcDetails,
		Progress:      jobspb.SchemaChangeGCProgress{},
		NonCancelable: true,
	}
	if _, err := jr.CreateJobWithTxn(ctx, gcJobRecord, jr.MakeJobID(), txn); err != nil {
		__antithesis_instrumentation__.Notify(11309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11310)
	}
	__antithesis_instrumentation__.Notify(11274)

	ignoredChildDescIDs := make(map[descpb.ID]struct{})
	for _, table := range details.TableDescs {
		__antithesis_instrumentation__.Notify(11311)
		ignoredChildDescIDs[table.ID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(11275)
	for _, typ := range details.TypeDescs {
		__antithesis_instrumentation__.Notify(11312)
		ignoredChildDescIDs[typ.ID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(11276)
	for _, schema := range details.SchemaDescs {
		__antithesis_instrumentation__.Notify(11313)
		ignoredChildDescIDs[schema.ID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(11277)
	all, err := descsCol.GetAllDescriptors(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(11314)
		return err
	} else {
		__antithesis_instrumentation__.Notify(11315)
	}
	__antithesis_instrumentation__.Notify(11278)
	allDescs := all.OrderedDescriptors()

	dbsWithDeletedSchemas := make(map[descpb.ID][]catalog.Descriptor)
	for _, schemaDesc := range details.SchemaDescs {
		__antithesis_instrumentation__.Notify(11316)

		isSchemaEmpty, err := isSchemaEmpty(ctx, txn, schemaDesc.GetID(), allDescs, ignoredChildDescIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(11321)
			return errors.Wrapf(err, "checking if schema %s is empty during restore cleanup", schemaDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(11322)
		}
		__antithesis_instrumentation__.Notify(11317)

		if !isSchemaEmpty {
			__antithesis_instrumentation__.Notify(11323)
			log.Warningf(ctx, "preserving schema %s on restore failure because it contains new child objects", schemaDesc.GetName())
			continue
		} else {
			__antithesis_instrumentation__.Notify(11324)
		}
		__antithesis_instrumentation__.Notify(11318)

		mutSchema, err := descsCol.GetMutableDescriptorByID(ctx, txn, schemaDesc.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(11325)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11326)
		}
		__antithesis_instrumentation__.Notify(11319)

		mutSchema.SetDropped()
		mutSchema.MaybeIncrementVersion()
		if err := descsCol.AddUncommittedDescriptor(mutSchema); err != nil {
			__antithesis_instrumentation__.Notify(11327)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11328)
		}
		__antithesis_instrumentation__.Notify(11320)

		b.Del(catalogkeys.EncodeNameKey(codec, mutSchema))
		b.Del(catalogkeys.MakeDescMetadataKey(codec, mutSchema.GetID()))
		descsCol.AddDeletedDescriptor(mutSchema.GetID())
		dbsWithDeletedSchemas[mutSchema.GetParentID()] = append(dbsWithDeletedSchemas[mutSchema.GetParentID()], mutSchema)
	}
	__antithesis_instrumentation__.Notify(11279)

	for dbID, schemas := range dbsWithDeletedSchemas {
		__antithesis_instrumentation__.Notify(11329)
		log.Infof(ctx, "deleting %d schema entries from database %d", len(schemas), dbID)
		desc, err := descsCol.GetMutableDescriptorByID(ctx, txn, dbID)
		if err != nil {
			__antithesis_instrumentation__.Notify(11332)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11333)
		}
		__antithesis_instrumentation__.Notify(11330)
		db := desc.(*dbdesc.Mutable)
		for _, sc := range schemas {
			__antithesis_instrumentation__.Notify(11334)
			if schemaInfo, ok := db.Schemas[sc.GetName()]; !ok {
				__antithesis_instrumentation__.Notify(11335)
				log.Warningf(ctx, "unexpected missing schema entry for %s from db %d; skipping deletion",
					sc.GetName(), dbID)
			} else {
				__antithesis_instrumentation__.Notify(11336)
				if schemaInfo.ID != sc.GetID() {
					__antithesis_instrumentation__.Notify(11337)
					log.Warningf(ctx, "unexpected schema entry %d for %s from db %d, expecting %d; skipping deletion",
						schemaInfo.ID, sc.GetName(), dbID, sc.GetID())
				} else {
					__antithesis_instrumentation__.Notify(11338)
					delete(db.Schemas, sc.GetName())
				}
			}
		}
		__antithesis_instrumentation__.Notify(11331)

		if err := descsCol.WriteDescToBatch(
			ctx, false, db, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(11339)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11340)
		}
	}
	__antithesis_instrumentation__.Notify(11280)

	deletedDBs := make(map[descpb.ID]struct{})
	for _, dbDesc := range details.DatabaseDescs {
		__antithesis_instrumentation__.Notify(11341)

		isDBEmpty, err := isDatabaseEmpty(ctx, txn, dbDesc.GetID(), allDescs, ignoredChildDescIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(11347)
			return errors.Wrapf(err, "checking if database %s is empty during restore cleanup", dbDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(11348)
		}
		__antithesis_instrumentation__.Notify(11342)
		if !isDBEmpty {
			__antithesis_instrumentation__.Notify(11349)
			log.Warningf(ctx, "preserving database %s on restore failure because it contains new child objects or schemas", dbDesc.GetName())
			continue
		} else {
			__antithesis_instrumentation__.Notify(11350)
		}
		__antithesis_instrumentation__.Notify(11343)

		db, err := descsCol.GetMutableDescriptorByID(ctx, txn, dbDesc.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(11351)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11352)
		}
		__antithesis_instrumentation__.Notify(11344)

		db.SetDropped()
		db.MaybeIncrementVersion()
		if err := descsCol.AddUncommittedDescriptor(db); err != nil {
			__antithesis_instrumentation__.Notify(11353)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11354)
		}
		__antithesis_instrumentation__.Notify(11345)

		descKey := catalogkeys.MakeDescMetadataKey(codec, db.GetID())
		b.Del(descKey)

		if !db.(catalog.DatabaseDescriptor).HasPublicSchemaWithDescriptor() {
			__antithesis_instrumentation__.Notify(11355)
			b.Del(catalogkeys.MakeSchemaNameKey(codec, db.GetID(), tree.PublicSchema))
		} else {
			__antithesis_instrumentation__.Notify(11356)
		}
		__antithesis_instrumentation__.Notify(11346)

		nameKey := catalogkeys.MakeDatabaseNameKey(codec, db.GetName())
		b.Del(nameKey)
		descsCol.AddDeletedDescriptor(db.GetID())
		deletedDBs[db.GetID()] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(11281)

	const kvTrace = false
	for _, t := range mutableTables {
		__antithesis_instrumentation__.Notify(11357)
		if err := descsCol.WriteDescToBatch(ctx, kvTrace, t, b); err != nil {
			__antithesis_instrumentation__.Notify(11358)
			return errors.Wrap(err, "writing dropping table to batch")
		} else {
			__antithesis_instrumentation__.Notify(11359)
		}
	}
	__antithesis_instrumentation__.Notify(11282)

	if err := txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(11360)
		return errors.Wrap(err, "dropping tables created at the start of restore caused by fail/cancel")
	} else {
		__antithesis_instrumentation__.Notify(11361)
	}
	__antithesis_instrumentation__.Notify(11283)

	return nil
}

func (r *restoreResumer) removeExistingTypeBackReferences(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	b *kv.Batch,
	restoredTables []*tabledesc.Mutable,
	details *jobspb.RestoreDetails,
) error {
	__antithesis_instrumentation__.Notify(11362)

	restoredTypes := make(map[descpb.ID]catalog.TypeDescriptor)
	existingTypes := make(map[descpb.ID]*typedesc.Mutable)
	for i := range details.TypeDescs {
		__antithesis_instrumentation__.Notify(11366)
		typ := details.TypeDescs[i]
		restoredTypes[typ.ID] = typedesc.NewBuilder(typ).BuildImmutableType()
	}
	__antithesis_instrumentation__.Notify(11363)
	for _, tbl := range restoredTables {
		__antithesis_instrumentation__.Notify(11367)
		lookup := func(id descpb.ID) (catalog.TypeDescriptor, error) {
			__antithesis_instrumentation__.Notify(11371)

			restored, ok := restoredTypes[id]
			if ok {
				__antithesis_instrumentation__.Notify(11374)
				return restored, nil
			} else {
				__antithesis_instrumentation__.Notify(11375)
			}
			__antithesis_instrumentation__.Notify(11372)

			typ, err := descsCol.GetMutableTypeVersionByID(ctx, txn, id)
			if err != nil {
				__antithesis_instrumentation__.Notify(11376)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(11377)
			}
			__antithesis_instrumentation__.Notify(11373)
			existingTypes[typ.GetID()] = typ
			return typ, nil
		}
		__antithesis_instrumentation__.Notify(11368)

		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx, txn, tbl.GetParentID(), tree.DatabaseLookupFlags{
				Required:       true,
				AvoidLeased:    true,
				IncludeOffline: true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(11378)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11379)
		}
		__antithesis_instrumentation__.Notify(11369)

		referencedTypes, _, err := tbl.GetAllReferencedTypeIDs(dbDesc, lookup)
		if err != nil {
			__antithesis_instrumentation__.Notify(11380)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11381)
		}
		__antithesis_instrumentation__.Notify(11370)

		for _, id := range referencedTypes {
			__antithesis_instrumentation__.Notify(11382)
			_, restored := restoredTypes[id]
			if !restored {
				__antithesis_instrumentation__.Notify(11383)
				desc, err := lookup(id)
				if err != nil {
					__antithesis_instrumentation__.Notify(11385)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11386)
				}
				__antithesis_instrumentation__.Notify(11384)
				existing := desc.(*typedesc.Mutable)
				existing.MaybeIncrementVersion()
				existing.RemoveReferencingDescriptorID(tbl.ID)
			} else {
				__antithesis_instrumentation__.Notify(11387)
			}
		}
	}
	__antithesis_instrumentation__.Notify(11364)

	for _, typ := range existingTypes {
		__antithesis_instrumentation__.Notify(11388)
		if typ.IsUncommittedVersion() {
			__antithesis_instrumentation__.Notify(11389)
			if err := descsCol.WriteDescToBatch(
				ctx, false, typ, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(11390)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11391)
			}
		} else {
			__antithesis_instrumentation__.Notify(11392)
		}
	}
	__antithesis_instrumentation__.Notify(11365)

	return nil
}

type systemTableNameWithConfig struct {
	systemTableName  string
	stagingTableName string
	config           systemBackupConfiguration
}

func (r *restoreResumer) restoreSystemUsers(
	ctx context.Context, db *kv.DB, systemTables []catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(11393)
	executor := r.execCfg.InternalExecutor
	return db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(11394)
		selectNonExistentUsers := "SELECT * FROM crdb_temp_system.users temp " +
			"WHERE NOT EXISTS (SELECT * FROM system.users u WHERE temp.username = u.username)"
		users, err := executor.QueryBuffered(ctx, "get-users",
			txn, selectNonExistentUsers)
		if err != nil {
			__antithesis_instrumentation__.Notify(11400)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11401)
		}
		__antithesis_instrumentation__.Notify(11395)

		insertUser := `INSERT INTO system.users ("username", "hashedPassword", "isRole") VALUES ($1, $2, $3)`
		newUsernames := make(map[string]bool)
		for _, user := range users {
			__antithesis_instrumentation__.Notify(11402)
			newUsernames[user[0].String()] = true
			if _, err = executor.Exec(ctx, "insert-non-existent-users", txn, insertUser,
				user[0], user[1], user[2]); err != nil {
				__antithesis_instrumentation__.Notify(11403)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11404)
			}
		}
		__antithesis_instrumentation__.Notify(11396)

		if len(systemTables) == 1 {
			__antithesis_instrumentation__.Notify(11405)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(11406)
		}
		__antithesis_instrumentation__.Notify(11397)

		selectNonExistentRoleMembers := "SELECT * FROM crdb_temp_system.role_members temp_rm WHERE " +
			"NOT EXISTS (SELECT * FROM system.role_members rm WHERE temp_rm.role = rm.role AND temp_rm.member = rm.member)"
		roleMembers, err := executor.QueryBuffered(ctx, "get-role-members",
			txn, selectNonExistentRoleMembers)
		if err != nil {
			__antithesis_instrumentation__.Notify(11407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11408)
		}
		__antithesis_instrumentation__.Notify(11398)

		insertRoleMember := `INSERT INTO system.role_members ("role", "member", "isAdmin") VALUES ($1, $2, $3)`
		for _, roleMember := range roleMembers {
			__antithesis_instrumentation__.Notify(11409)

			if _, ok := newUsernames[roleMember[1].String()]; ok {
				__antithesis_instrumentation__.Notify(11410)
				if _, err = executor.Exec(ctx, "insert-non-existent-role-members", txn, insertRoleMember,
					roleMember[0], roleMember[1], roleMember[2]); err != nil {
					__antithesis_instrumentation__.Notify(11411)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11412)
				}
			} else {
				__antithesis_instrumentation__.Notify(11413)
			}
		}
		__antithesis_instrumentation__.Notify(11399)
		return nil
	})
}

func (r *restoreResumer) restoreSystemTables(
	ctx context.Context, db *kv.DB, tables []catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(11414)
	details := r.job.Details().(jobspb.RestoreDetails)
	if details.SystemTablesMigrated == nil {
		__antithesis_instrumentation__.Notify(11419)
		details.SystemTablesMigrated = make(map[string]bool)
	} else {
		__antithesis_instrumentation__.Notify(11420)
	}
	__antithesis_instrumentation__.Notify(11415)
	tempSystemDBID := tempSystemDatabaseID(details)

	systemTablesToRestore := make([]systemTableNameWithConfig, 0)
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(11421)
		if table.GetParentID() != tempSystemDBID {
			__antithesis_instrumentation__.Notify(11424)
			continue
		} else {
			__antithesis_instrumentation__.Notify(11425)
		}
		__antithesis_instrumentation__.Notify(11422)
		systemTableName := table.GetName()
		stagingTableName := restoreTempSystemDB + "." + systemTableName

		config, ok := systemTableBackupConfiguration[systemTableName]
		if !ok {
			__antithesis_instrumentation__.Notify(11426)
			log.Warningf(ctx, "no configuration specified for table %s... skipping restoration",
				systemTableName)
		} else {
			__antithesis_instrumentation__.Notify(11427)
		}
		__antithesis_instrumentation__.Notify(11423)
		systemTablesToRestore = append(systemTablesToRestore, systemTableNameWithConfig{
			systemTableName:  systemTableName,
			stagingTableName: stagingTableName,
			config:           config,
		})
	}
	__antithesis_instrumentation__.Notify(11416)

	sort.SliceStable(systemTablesToRestore, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(11428)
		return systemTablesToRestore[i].config.restoreInOrder < systemTablesToRestore[j].config.restoreInOrder
	})
	__antithesis_instrumentation__.Notify(11417)

	for _, systemTable := range systemTablesToRestore {
		__antithesis_instrumentation__.Notify(11429)
		if systemTable.config.migrationFunc != nil {
			__antithesis_instrumentation__.Notify(11432)
			if details.SystemTablesMigrated[systemTable.systemTableName] {
				__antithesis_instrumentation__.Notify(11434)
				continue
			} else {
				__antithesis_instrumentation__.Notify(11435)
			}
			__antithesis_instrumentation__.Notify(11433)

			if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(11436)
				if err := systemTable.config.migrationFunc(ctx, r.execCfg, txn,
					systemTable.stagingTableName); err != nil {
					__antithesis_instrumentation__.Notify(11438)
					return err
				} else {
					__antithesis_instrumentation__.Notify(11439)
				}
				__antithesis_instrumentation__.Notify(11437)

				details.SystemTablesMigrated[systemTable.systemTableName] = true
				return r.job.SetDetails(ctx, txn, details)
			}); err != nil {
				__antithesis_instrumentation__.Notify(11440)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11441)
			}
		} else {
			__antithesis_instrumentation__.Notify(11442)
		}
		__antithesis_instrumentation__.Notify(11430)

		if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(11443)
			txn.SetDebugName("system-restore-txn")

			restoreFunc := defaultSystemTableRestoreFunc
			if systemTable.config.customRestoreFunc != nil {
				__antithesis_instrumentation__.Notify(11446)
				restoreFunc = systemTable.config.customRestoreFunc
				log.Eventf(ctx, "using custom restore function for table %s", systemTable.systemTableName)
			} else {
				__antithesis_instrumentation__.Notify(11447)
			}
			__antithesis_instrumentation__.Notify(11444)

			log.Eventf(ctx, "restoring system table %s", systemTable.systemTableName)
			err := restoreFunc(ctx, r.execCfg, txn, systemTable.systemTableName, systemTable.stagingTableName)
			if err != nil {
				__antithesis_instrumentation__.Notify(11448)
				return errors.Wrapf(err, "restoring system table %s", systemTable.systemTableName)
			} else {
				__antithesis_instrumentation__.Notify(11449)
			}
			__antithesis_instrumentation__.Notify(11445)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(11450)
			return err
		} else {
			__antithesis_instrumentation__.Notify(11451)
		}
		__antithesis_instrumentation__.Notify(11431)

		if fn := r.testingKnobs.duringSystemTableRestoration; fn != nil {
			__antithesis_instrumentation__.Notify(11452)
			if err := fn(systemTable.systemTableName); err != nil {
				__antithesis_instrumentation__.Notify(11453)
				return err
			} else {
				__antithesis_instrumentation__.Notify(11454)
			}
		} else {
			__antithesis_instrumentation__.Notify(11455)
		}
	}
	__antithesis_instrumentation__.Notify(11418)

	return nil
}

func (r *restoreResumer) cleanupTempSystemTables(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(11456)
	executor := r.execCfg.InternalExecutor

	checkIfDatabaseExists := "SELECT database_name FROM [SHOW DATABASES] WHERE database_name=$1"
	if row, err := executor.QueryRow(ctx, "checking-for-temp-system-db", txn, checkIfDatabaseExists, restoreTempSystemDB); err != nil {
		__antithesis_instrumentation__.Notify(11460)
		return errors.Wrap(err, "checking for temporary system db")
	} else {
		__antithesis_instrumentation__.Notify(11461)
		if row == nil {
			__antithesis_instrumentation__.Notify(11462)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(11463)
		}
	}
	__antithesis_instrumentation__.Notify(11457)

	gcTTLQuery := fmt.Sprintf("ALTER DATABASE %s CONFIGURE ZONE USING gc.ttlseconds=1", restoreTempSystemDB)
	if _, err := executor.Exec(ctx, "altering-gc-ttl-temp-system", txn, gcTTLQuery); err != nil {
		__antithesis_instrumentation__.Notify(11464)
		log.Errorf(ctx, "failed to update the GC TTL of %q: %+v", restoreTempSystemDB, err)
	} else {
		__antithesis_instrumentation__.Notify(11465)
	}
	__antithesis_instrumentation__.Notify(11458)
	dropTableQuery := fmt.Sprintf("DROP DATABASE %s CASCADE", restoreTempSystemDB)
	if _, err := executor.Exec(ctx, "drop-temp-system-db", txn, dropTableQuery); err != nil {
		__antithesis_instrumentation__.Notify(11466)
		return errors.Wrap(err, "dropping temporary system db")
	} else {
		__antithesis_instrumentation__.Notify(11467)
	}
	__antithesis_instrumentation__.Notify(11459)
	return nil
}

var _ jobs.Resumer = &restoreResumer{}

func init() {
	jobs.RegisterConstructor(
		jobspb.TypeRestore,
		func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
			return &restoreResumer{
				job:      job,
				settings: settings,
			}
		},
	)
}
