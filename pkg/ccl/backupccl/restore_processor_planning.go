package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func distRestore(
	ctx context.Context,
	execCtx sql.JobExecContext,
	jobID int64,
	chunks [][]execinfrapb.RestoreSpanEntry,
	pkIDs map[uint64]bool,
	encryption *jobspb.BackupEncryptionOptions,
	tableRekeys []execinfrapb.TableRekey,
	tenantRekeys []execinfrapb.TenantRekey,
	restoreTime hlc.Timestamp,
	progCh chan *execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
) error {
	__antithesis_instrumentation__.Notify(12414)
	defer close(progCh)
	var noTxn *kv.Txn

	dsp := execCtx.DistSQLPlanner()
	evalCtx := execCtx.ExtendedEvalContext()

	if encryption != nil && func() bool {
		__antithesis_instrumentation__.Notify(12426)
		return encryption.Mode == jobspb.EncryptionMode_KMS == true
	}() == true {
		__antithesis_instrumentation__.Notify(12427)
		kms, err := cloud.KMSFromURI(encryption.KMSInfo.Uri, &backupKMSEnv{
			settings: execCtx.ExecCfg().Settings,
			conf:     &execCtx.ExecCfg().ExternalIODirConfig,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(12429)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12430)
		}
		__antithesis_instrumentation__.Notify(12428)

		encryption.Key, err = kms.Decrypt(ctx, encryption.KMSInfo.EncryptedDataKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(12431)
			return errors.Wrap(err,
				"failed to decrypt data key before starting BackupDataProcessor")
		} else {
			__antithesis_instrumentation__.Notify(12432)
		}
	} else {
		__antithesis_instrumentation__.Notify(12433)
	}
	__antithesis_instrumentation__.Notify(12415)

	var fileEncryption *roachpb.FileEncryptionOptions
	if encryption != nil {
		__antithesis_instrumentation__.Notify(12434)
		fileEncryption = &roachpb.FileEncryptionOptions{Key: encryption.Key}
	} else {
		__antithesis_instrumentation__.Notify(12435)
	}
	__antithesis_instrumentation__.Notify(12416)

	planCtx, sqlInstanceIDs, err := dsp.SetupAllNodesPlanning(ctx, evalCtx, execCtx.ExecCfg())
	if err != nil {
		__antithesis_instrumentation__.Notify(12436)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12437)
	}
	__antithesis_instrumentation__.Notify(12417)

	splitAndScatterSpecs, err := makeSplitAndScatterSpecs(sqlInstanceIDs, chunks, tableRekeys, tenantRekeys)
	if err != nil {
		__antithesis_instrumentation__.Notify(12438)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12439)
	}
	__antithesis_instrumentation__.Notify(12418)

	restoreDataSpec := execinfrapb.RestoreDataSpec{
		JobID:        jobID,
		RestoreTime:  restoreTime,
		Encryption:   fileEncryption,
		TableRekeys:  tableRekeys,
		TenantRekeys: tenantRekeys,
		PKIDs:        pkIDs,
		Validation:   jobspb.RestoreValidation_DefaultRestore,
	}

	if len(splitAndScatterSpecs) == 0 {
		__antithesis_instrumentation__.Notify(12440)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(12441)
	}
	__antithesis_instrumentation__.Notify(12419)

	p := planCtx.NewPhysicalPlan()

	splitAndScatterStageID := p.NewStageOnNodes(sqlInstanceIDs)
	splitAndScatterProcs := make(map[base.SQLInstanceID]physicalplan.ProcessorIdx)

	defaultStream := int32(0)
	rangeRouterSpec := execinfrapb.OutputRouterSpec_RangeRouterSpec{
		Spans:       nil,
		DefaultDest: &defaultStream,
		Encodings: []execinfrapb.OutputRouterSpec_RangeRouterSpec_ColumnEncoding{
			{
				Column:   0,
				Encoding: descpb.DatumEncoding_ASCENDING_KEY,
			},
		},
	}
	for stream, sqlInstanceID := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(12442)
		startBytes, endBytes, err := routingSpanForSQLInstance(sqlInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(12444)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12445)
		}
		__antithesis_instrumentation__.Notify(12443)

		span := execinfrapb.OutputRouterSpec_RangeRouterSpec_Span{
			Start:  startBytes,
			End:    endBytes,
			Stream: int32(stream),
		}
		rangeRouterSpec.Spans = append(rangeRouterSpec.Spans, span)
	}
	__antithesis_instrumentation__.Notify(12420)

	sort.Slice(rangeRouterSpec.Spans, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(12446)
		return bytes.Compare(rangeRouterSpec.Spans[i].Start, rangeRouterSpec.Spans[j].Start) == -1
	})
	__antithesis_instrumentation__.Notify(12421)

	for _, n := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(12447)
		spec := splitAndScatterSpecs[n]
		if spec == nil {
			__antithesis_instrumentation__.Notify(12449)

			continue
		} else {
			__antithesis_instrumentation__.Notify(12450)
		}
		__antithesis_instrumentation__.Notify(12448)
		proc := physicalplan.Processor{
			SQLInstanceID: n,
			Spec: execinfrapb.ProcessorSpec{
				Core: execinfrapb.ProcessorCoreUnion{SplitAndScatter: splitAndScatterSpecs[n]},
				Post: execinfrapb.PostProcessSpec{},
				Output: []execinfrapb.OutputRouterSpec{
					{
						Type:            execinfrapb.OutputRouterSpec_BY_RANGE,
						RangeRouterSpec: rangeRouterSpec,
					},
				},
				StageID:     splitAndScatterStageID,
				ResultTypes: splitAndScatterOutputTypes,
			},
		}
		pIdx := p.AddProcessor(proc)
		splitAndScatterProcs[n] = pIdx
	}
	__antithesis_instrumentation__.Notify(12422)

	restoreDataStageID := p.NewStageOnNodes(sqlInstanceIDs)
	restoreDataProcs := make(map[base.SQLInstanceID]physicalplan.ProcessorIdx)
	for _, sqlInstanceID := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(12451)
		proc := physicalplan.Processor{
			SQLInstanceID: sqlInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Input: []execinfrapb.InputSyncSpec{
					{ColumnTypes: splitAndScatterOutputTypes},
				},
				Core:        execinfrapb.ProcessorCoreUnion{RestoreData: &restoreDataSpec},
				Post:        execinfrapb.PostProcessSpec{},
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				StageID:     restoreDataStageID,
				ResultTypes: []*types.T{},
			},
		}
		pIdx := p.AddProcessor(proc)
		restoreDataProcs[sqlInstanceID] = pIdx
		p.ResultRouters = append(p.ResultRouters, pIdx)
	}
	__antithesis_instrumentation__.Notify(12423)

	for _, srcProc := range splitAndScatterProcs {
		__antithesis_instrumentation__.Notify(12452)
		slot := 0
		for _, destSQLInstanceID := range sqlInstanceIDs {
			__antithesis_instrumentation__.Notify(12453)

			destProc := restoreDataProcs[destSQLInstanceID]
			p.Streams = append(p.Streams, physicalplan.Stream{
				SourceProcessor:  srcProc,
				SourceRouterSlot: slot,
				DestProcessor:    destProc,
				DestInput:        0,
			})
			slot++
		}
	}
	__antithesis_instrumentation__.Notify(12424)

	dsp.FinalizePlan(planCtx, p)

	metaFn := func(_ context.Context, meta *execinfrapb.ProducerMetadata) error {
		__antithesis_instrumentation__.Notify(12454)
		if meta.BulkProcessorProgress != nil {
			__antithesis_instrumentation__.Notify(12456)

			progCh <- meta.BulkProcessorProgress
		} else {
			__antithesis_instrumentation__.Notify(12457)
		}
		__antithesis_instrumentation__.Notify(12455)
		return nil
	}
	__antithesis_instrumentation__.Notify(12425)

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

	evalCtxCopy := *evalCtx
	dsp.Run(ctx, planCtx, noTxn, p, recv, &evalCtxCopy, nil)()
	return rowResultWriter.Err()
}

func makeSplitAndScatterSpecs(
	sqlInstanceIDs []base.SQLInstanceID,
	chunks [][]execinfrapb.RestoreSpanEntry,
	tableRekeys []execinfrapb.TableRekey,
	tenantRekeys []execinfrapb.TenantRekey,
) (map[base.SQLInstanceID]*execinfrapb.SplitAndScatterSpec, error) {
	__antithesis_instrumentation__.Notify(12458)
	specsBySQLInstanceID := make(map[base.SQLInstanceID]*execinfrapb.SplitAndScatterSpec)
	for i, chunk := range chunks {
		__antithesis_instrumentation__.Notify(12460)
		sqlInstanceID := sqlInstanceIDs[i%len(sqlInstanceIDs)]
		if spec, ok := specsBySQLInstanceID[sqlInstanceID]; ok {
			__antithesis_instrumentation__.Notify(12461)
			spec.Chunks = append(spec.Chunks, execinfrapb.SplitAndScatterSpec_RestoreEntryChunk{
				Entries: chunk,
			})
		} else {
			__antithesis_instrumentation__.Notify(12462)
			specsBySQLInstanceID[sqlInstanceID] = &execinfrapb.SplitAndScatterSpec{
				Chunks: []execinfrapb.SplitAndScatterSpec_RestoreEntryChunk{{
					Entries: chunk,
				}},
				TableRekeys:  tableRekeys,
				TenantRekeys: tenantRekeys,
				Validation:   jobspb.RestoreValidation_DefaultRestore,
			}
		}
	}
	__antithesis_instrumentation__.Notify(12459)
	return specsBySQLInstanceID, nil
}
