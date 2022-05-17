package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func distBackupPlanSpecs(
	ctx context.Context,
	planCtx *sql.PlanningCtx,
	execCtx sql.JobExecContext,
	dsp *sql.DistSQLPlanner,
	jobID int64,
	spans roachpb.Spans,
	introducedSpans roachpb.Spans,
	pkIDs map[uint64]bool,
	defaultURI string,
	urisByLocalityKV map[string]string,
	encryption *jobspb.BackupEncryptionOptions,
	mvccFilter roachpb.MVCCFilter,
	startTime, endTime hlc.Timestamp,
) (map[base.SQLInstanceID]*execinfrapb.BackupDataSpec, error) {
	__antithesis_instrumentation__.Notify(8919)
	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, "backup-plan-specs")
	_ = ctx
	defer span.Finish()
	user := execCtx.User()
	execCfg := execCtx.ExecCfg()

	var spanPartitions []sql.SpanPartition
	var introducedSpanPartitions []sql.SpanPartition
	var err error
	if len(spans) > 0 {
		__antithesis_instrumentation__.Notify(8927)
		spanPartitions, err = dsp.PartitionSpans(ctx, planCtx, spans)
		if err != nil {
			__antithesis_instrumentation__.Notify(8928)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(8929)
		}
	} else {
		__antithesis_instrumentation__.Notify(8930)
	}
	__antithesis_instrumentation__.Notify(8920)
	if len(introducedSpans) > 0 {
		__antithesis_instrumentation__.Notify(8931)
		introducedSpanPartitions, err = dsp.PartitionSpans(ctx, planCtx, introducedSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(8932)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(8933)
		}
	} else {
		__antithesis_instrumentation__.Notify(8934)
	}
	__antithesis_instrumentation__.Notify(8921)

	if encryption != nil && func() bool {
		__antithesis_instrumentation__.Notify(8935)
		return encryption.Mode == jobspb.EncryptionMode_KMS == true
	}() == true {
		__antithesis_instrumentation__.Notify(8936)
		kms, err := cloud.KMSFromURI(encryption.KMSInfo.Uri, &backupKMSEnv{
			settings: execCfg.Settings,
			conf:     &execCfg.ExternalIODirConfig,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(8938)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(8939)
		}
		__antithesis_instrumentation__.Notify(8937)

		encryption.Key, err = kms.Decrypt(planCtx.EvalContext().Context,
			encryption.KMSInfo.EncryptedDataKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(8940)
			return nil, errors.Wrap(err,
				"failed to decrypt data key before starting BackupDataProcessor")
		} else {
			__antithesis_instrumentation__.Notify(8941)
		}
	} else {
		__antithesis_instrumentation__.Notify(8942)
	}
	__antithesis_instrumentation__.Notify(8922)

	var fileEncryption *roachpb.FileEncryptionOptions
	if encryption != nil {
		__antithesis_instrumentation__.Notify(8943)
		fileEncryption = &roachpb.FileEncryptionOptions{Key: encryption.Key}
	} else {
		__antithesis_instrumentation__.Notify(8944)
	}
	__antithesis_instrumentation__.Notify(8923)

	sqlInstanceIDToSpec := make(map[base.SQLInstanceID]*execinfrapb.BackupDataSpec)
	for _, partition := range spanPartitions {
		__antithesis_instrumentation__.Notify(8945)
		spec := &execinfrapb.BackupDataSpec{
			JobID:            jobID,
			Spans:            partition.Spans,
			DefaultURI:       defaultURI,
			URIsByLocalityKV: urisByLocalityKV,
			MVCCFilter:       mvccFilter,
			Encryption:       fileEncryption,
			PKIDs:            pkIDs,
			BackupStartTime:  startTime,
			BackupEndTime:    endTime,
			UserProto:        user.EncodeProto(),
		}
		sqlInstanceIDToSpec[partition.SQLInstanceID] = spec
	}
	__antithesis_instrumentation__.Notify(8924)

	for _, partition := range introducedSpanPartitions {
		__antithesis_instrumentation__.Notify(8946)
		if spec, ok := sqlInstanceIDToSpec[partition.SQLInstanceID]; ok {
			__antithesis_instrumentation__.Notify(8947)
			spec.IntroducedSpans = partition.Spans
		} else {
			__antithesis_instrumentation__.Notify(8948)

			spec := &execinfrapb.BackupDataSpec{
				JobID:            jobID,
				IntroducedSpans:  partition.Spans,
				DefaultURI:       defaultURI,
				URIsByLocalityKV: urisByLocalityKV,
				MVCCFilter:       mvccFilter,
				Encryption:       fileEncryption,
				PKIDs:            pkIDs,
				BackupStartTime:  startTime,
				BackupEndTime:    endTime,
				UserProto:        user.EncodeProto(),
			}
			sqlInstanceIDToSpec[partition.SQLInstanceID] = spec
		}
	}
	__antithesis_instrumentation__.Notify(8925)

	backupPlanningTraceEvent := BackupProcessorPlanningTraceEvent{
		NodeToNumSpans: make(map[int32]int64),
	}
	for node, spec := range sqlInstanceIDToSpec {
		__antithesis_instrumentation__.Notify(8949)
		numSpans := int64(len(spec.Spans) + len(spec.IntroducedSpans))
		backupPlanningTraceEvent.NodeToNumSpans[int32(node)] = numSpans
		backupPlanningTraceEvent.TotalNumSpans += numSpans
	}
	__antithesis_instrumentation__.Notify(8926)
	span.RecordStructured(&backupPlanningTraceEvent)

	return sqlInstanceIDToSpec, nil
}

func distBackup(
	ctx context.Context,
	execCtx sql.JobExecContext,
	planCtx *sql.PlanningCtx,
	dsp *sql.DistSQLPlanner,
	progCh chan *execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
	backupSpecs map[base.SQLInstanceID]*execinfrapb.BackupDataSpec,
) error {
	__antithesis_instrumentation__.Notify(8950)
	ctx, span := tracing.ChildSpan(ctx, "backup-distsql")
	defer span.Finish()
	evalCtx := execCtx.ExtendedEvalContext()
	var noTxn *kv.Txn

	if len(backupSpecs) == 0 {
		__antithesis_instrumentation__.Notify(8954)
		close(progCh)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(8955)
	}
	__antithesis_instrumentation__.Notify(8951)

	corePlacement := make([]physicalplan.ProcessorCorePlacement, len(backupSpecs))
	i := 0
	for sqlInstanceID, spec := range backupSpecs {
		__antithesis_instrumentation__.Notify(8956)
		corePlacement[i].SQLInstanceID = sqlInstanceID
		corePlacement[i].Core.BackupData = spec
		i++
	}
	__antithesis_instrumentation__.Notify(8952)

	p := planCtx.NewPhysicalPlan()

	p.AddNoInputStage(corePlacement, execinfrapb.PostProcessSpec{}, []*types.T{}, execinfrapb.Ordering{})
	p.PlanToStreamColMap = []int{}

	dsp.FinalizePlan(planCtx, p)

	metaFn := func(_ context.Context, meta *execinfrapb.ProducerMetadata) error {
		__antithesis_instrumentation__.Notify(8957)
		if meta.BulkProcessorProgress != nil {
			__antithesis_instrumentation__.Notify(8959)

			progCh <- meta.BulkProcessorProgress
		} else {
			__antithesis_instrumentation__.Notify(8960)
		}
		__antithesis_instrumentation__.Notify(8958)
		return nil
	}
	__antithesis_instrumentation__.Notify(8953)

	rowResultWriter := sql.NewRowResultWriter(nil)

	recv := sql.MakeDistSQLReceiver(
		ctx,
		sql.NewMetadataCallbackWriter(rowResultWriter, metaFn),
		tree.Rows,
		nil,
		noTxn,
		nil,
		evalCtx.Tracing,
		evalCtx.ExecCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	defer close(progCh)

	evalCtxCopy := *evalCtx
	dsp.Run(ctx, planCtx, noTxn, p, recv, &evalCtxCopy, nil)()
	return rowResultWriter.Err()
}
