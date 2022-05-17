package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (p *planner) getVirtualTabler() VirtualTabler {
	__antithesis_instrumentation__.Notify(627623)
	return p.extendedEvalCtx.VirtualSchemas
}

func (p *planner) createDropDatabaseJob(
	ctx context.Context,
	databaseID descpb.ID,
	schemasToDrop []descpb.ID,
	tableDropDetails []jobspb.DroppedTableDetails,
	typesToDrop []*typedesc.Mutable,
	jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(627624)

	tableIDs := make([]descpb.ID, 0, len(tableDropDetails))
	for _, d := range tableDropDetails {
		__antithesis_instrumentation__.Notify(627628)
		tableIDs = append(tableIDs, d.ID)
	}
	__antithesis_instrumentation__.Notify(627625)
	typeIDs := make([]descpb.ID, 0, len(typesToDrop))
	for _, t := range typesToDrop {
		__antithesis_instrumentation__.Notify(627629)
		typeIDs = append(typeIDs, t.ID)
	}
	__antithesis_instrumentation__.Notify(627626)
	jobRecord := jobs.Record{
		Description:   jobDesc,
		Username:      p.User(),
		DescriptorIDs: tableIDs,
		Details: jobspb.SchemaChangeDetails{
			DroppedSchemas:    schemasToDrop,
			DroppedTables:     tableDropDetails,
			DroppedTypes:      typeIDs,
			DroppedDatabaseID: databaseID,
			FormatVersion:     jobspb.DatabaseJobFormatVersion,
		},
		Progress:      jobspb.SchemaChangeProgress{},
		NonCancelable: true,
	}
	newJob, err := p.extendedEvalCtx.QueueJob(ctx, jobRecord)
	if err != nil {
		__antithesis_instrumentation__.Notify(627630)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627631)
	}
	__antithesis_instrumentation__.Notify(627627)
	log.Infof(ctx, "queued new drop database job %d for database %d", newJob.ID(), databaseID)
	return nil
}

func (p *planner) createNonDropDatabaseChangeJob(
	ctx context.Context, databaseID descpb.ID, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(627632)
	jobRecord := jobs.Record{
		Description: jobDesc,
		Username:    p.User(),
		Details: jobspb.SchemaChangeDetails{
			DescID:        databaseID,
			FormatVersion: jobspb.DatabaseJobFormatVersion,
		},
		Progress:      jobspb.SchemaChangeProgress{},
		NonCancelable: true,
	}
	newJob, err := p.extendedEvalCtx.QueueJob(ctx, jobRecord)
	if err != nil {
		__antithesis_instrumentation__.Notify(627634)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627635)
	}
	__antithesis_instrumentation__.Notify(627633)
	log.Infof(ctx, "queued new database schema change job %d for database %d", newJob.ID(), databaseID)
	return nil
}

func (p *planner) createOrUpdateSchemaChangeJob(
	ctx context.Context, tableDesc *tabledesc.Mutable, jobDesc string, mutationID descpb.MutationID,
) error {
	__antithesis_instrumentation__.Notify(627636)

	if tableDesc.GetDeclarativeSchemaChangerState() != nil {
		__antithesis_instrumentation__.Notify(627643)
		return scerrors.ConcurrentSchemaChangeError(tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(627644)
	}
	__antithesis_instrumentation__.Notify(627637)

	record, recordExists := p.extendedEvalCtx.SchemaChangeJobRecords[tableDesc.ID]
	if p.extendedEvalCtx.ExecCfg.TestingKnobs.RunAfterSCJobsCacheLookup != nil {
		__antithesis_instrumentation__.Notify(627645)
		p.extendedEvalCtx.ExecCfg.TestingKnobs.RunAfterSCJobsCacheLookup(record)
	} else {
		__antithesis_instrumentation__.Notify(627646)
	}
	__antithesis_instrumentation__.Notify(627638)

	var spanList []jobspb.ResumeSpanList
	if recordExists {
		__antithesis_instrumentation__.Notify(627647)
		spanList = record.Details.(jobspb.SchemaChangeDetails).ResumeSpanList
		prefix := p.ExecCfg().Codec.TenantPrefix()
		for i := range spanList {
			__antithesis_instrumentation__.Notify(627648)
			for j := range spanList[i].ResumeSpans {
				__antithesis_instrumentation__.Notify(627649)
				sp, err := keys.RewriteSpanToTenantPrefix(spanList[i].ResumeSpans[j], prefix)
				if err != nil {
					__antithesis_instrumentation__.Notify(627651)
					return err
				} else {
					__antithesis_instrumentation__.Notify(627652)
				}
				__antithesis_instrumentation__.Notify(627650)
				spanList[i].ResumeSpans[j] = sp
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(627653)
	}
	__antithesis_instrumentation__.Notify(627639)
	span := tableDesc.PrimaryIndexSpan(p.ExecCfg().Codec)
	for i := len(tableDesc.ClusterVersion().Mutations) + len(spanList); i < len(tableDesc.Mutations); i++ {
		__antithesis_instrumentation__.Notify(627654)
		var resumeSpans []roachpb.Span
		mut := tableDesc.Mutations[i]
		if mut.GetIndex() != nil && func() bool {
			__antithesis_instrumentation__.Notify(627656)
			return mut.GetIndex().UseDeletePreservingEncoding == true
		}() == true {
			__antithesis_instrumentation__.Notify(627657)

			resumeSpans = []roachpb.Span{tableDesc.IndexSpan(p.ExecCfg().Codec, mut.GetIndex().ID)}
		} else {
			__antithesis_instrumentation__.Notify(627658)
			resumeSpans = []roachpb.Span{span}
		}
		__antithesis_instrumentation__.Notify(627655)
		spanList = append(spanList, jobspb.ResumeSpanList{
			ResumeSpans: resumeSpans,
		})
	}
	__antithesis_instrumentation__.Notify(627640)

	if !recordExists {
		__antithesis_instrumentation__.Notify(627659)

		newRecord := jobs.Record{
			JobID:         p.extendedEvalCtx.ExecCfg.JobRegistry.MakeJobID(),
			Description:   jobDesc,
			Username:      p.User(),
			DescriptorIDs: descpb.IDs{tableDesc.GetID()},
			Details: jobspb.SchemaChangeDetails{
				DescID:          tableDesc.ID,
				TableMutationID: mutationID,
				ResumeSpanList:  spanList,

				FormatVersion: jobspb.DatabaseJobFormatVersion,
			},
			Progress: jobspb.SchemaChangeProgress{},

			NonCancelable: mutationID == descpb.InvalidMutationID && func() bool {
				__antithesis_instrumentation__.Notify(627661)
				return !tableDesc.Adding() == true
			}() == true,
		}
		p.extendedEvalCtx.SchemaChangeJobRecords[tableDesc.ID] = &newRecord

		if mutationID != descpb.InvalidMutationID {
			__antithesis_instrumentation__.Notify(627662)
			tableDesc.MutationJobs = append(tableDesc.MutationJobs, descpb.TableDescriptor_MutationJob{
				MutationID: mutationID, JobID: newRecord.JobID})
		} else {
			__antithesis_instrumentation__.Notify(627663)
		}
		__antithesis_instrumentation__.Notify(627660)
		log.Infof(ctx, "queued new schema-change job %d for table %d, mutation %d",
			newRecord.JobID, tableDesc.ID, mutationID)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(627664)
	}
	__antithesis_instrumentation__.Notify(627641)

	oldDetails := record.Details.(jobspb.SchemaChangeDetails)
	newDetails := jobspb.SchemaChangeDetails{
		DescID:          tableDesc.ID,
		TableMutationID: oldDetails.TableMutationID,
		ResumeSpanList:  spanList,

		FormatVersion: jobspb.DatabaseJobFormatVersion,
	}
	if oldDetails.TableMutationID != descpb.InvalidMutationID {
		__antithesis_instrumentation__.Notify(627665)

		if mutationID != descpb.InvalidMutationID && func() bool {
			__antithesis_instrumentation__.Notify(627666)
			return mutationID != oldDetails.TableMutationID == true
		}() == true {
			__antithesis_instrumentation__.Notify(627667)
			return errors.AssertionFailedf(
				"attempted to update job for mutation %d, but job already exists with mutation %d",
				mutationID, oldDetails.TableMutationID)
		} else {
			__antithesis_instrumentation__.Notify(627668)
		}
	} else {
		__antithesis_instrumentation__.Notify(627669)

		if mutationID != descpb.InvalidMutationID {
			__antithesis_instrumentation__.Notify(627670)
			newDetails.TableMutationID = mutationID

			tableDesc.MutationJobs = append(tableDesc.MutationJobs, descpb.TableDescriptor_MutationJob{
				MutationID: mutationID, JobID: record.JobID})

			record.NonCancelable = false
		} else {
			__antithesis_instrumentation__.Notify(627671)
		}
	}
	__antithesis_instrumentation__.Notify(627642)
	record.Details = newDetails
	record.AppendDescription(jobDesc)
	log.Infof(ctx, "job %d: updated with schema change for table %d, mutation %d",
		record.JobID, tableDesc.ID, mutationID)
	return nil
}

func (p *planner) writeSchemaChange(
	ctx context.Context, tableDesc *tabledesc.Mutable, mutationID descpb.MutationID, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(627672)
	if !p.EvalContext().TxnImplicit {
		__antithesis_instrumentation__.Notify(627676)
		telemetry.Inc(sqltelemetry.SchemaChangeInExplicitTxnCounter)
	} else {
		__antithesis_instrumentation__.Notify(627677)
	}
	__antithesis_instrumentation__.Notify(627673)
	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(627678)

		return errors.Errorf("no schema changes allowed on table %q as it is being dropped",
			tableDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(627679)
	}
	__antithesis_instrumentation__.Notify(627674)
	if !tableDesc.IsNew() {
		__antithesis_instrumentation__.Notify(627680)
		if err := p.createOrUpdateSchemaChangeJob(ctx, tableDesc, jobDesc, mutationID); err != nil {
			__antithesis_instrumentation__.Notify(627681)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627682)
		}
	} else {
		__antithesis_instrumentation__.Notify(627683)
	}
	__antithesis_instrumentation__.Notify(627675)
	return p.writeTableDesc(ctx, tableDesc)
}

func (p *planner) writeSchemaChangeToBatch(
	ctx context.Context, tableDesc *tabledesc.Mutable, b *kv.Batch,
) error {
	__antithesis_instrumentation__.Notify(627684)
	if !p.EvalContext().TxnImplicit {
		__antithesis_instrumentation__.Notify(627687)
		telemetry.Inc(sqltelemetry.SchemaChangeInExplicitTxnCounter)
	} else {
		__antithesis_instrumentation__.Notify(627688)
	}
	__antithesis_instrumentation__.Notify(627685)
	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(627689)

		return errors.Errorf("no schema changes allowed on table %q as it is being dropped",
			tableDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(627690)
	}
	__antithesis_instrumentation__.Notify(627686)
	return p.writeTableDescToBatch(ctx, tableDesc, b)
}

func (p *planner) writeDropTable(
	ctx context.Context, tableDesc *tabledesc.Mutable, queueJob bool, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(627691)
	if queueJob {
		__antithesis_instrumentation__.Notify(627693)
		if err := p.createOrUpdateSchemaChangeJob(ctx, tableDesc, jobDesc, descpb.InvalidMutationID); err != nil {
			__antithesis_instrumentation__.Notify(627694)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627695)
		}
	} else {
		__antithesis_instrumentation__.Notify(627696)
	}
	__antithesis_instrumentation__.Notify(627692)
	return p.writeTableDesc(ctx, tableDesc)
}

func (p *planner) writeTableDesc(ctx context.Context, tableDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(627697)
	b := p.txn.NewBatch()
	if err := p.writeTableDescToBatch(ctx, tableDesc, b); err != nil {
		__antithesis_instrumentation__.Notify(627699)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627700)
	}
	__antithesis_instrumentation__.Notify(627698)
	return p.txn.Run(ctx, b)
}

func (p *planner) writeTableDescToBatch(
	ctx context.Context, tableDesc *tabledesc.Mutable, b *kv.Batch,
) error {
	__antithesis_instrumentation__.Notify(627701)
	if tableDesc.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(627705)
		return errors.AssertionFailedf("virtual descriptors cannot be stored, found: %v", tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(627706)
	}
	__antithesis_instrumentation__.Notify(627702)

	if tableDesc.IsNew() {
		__antithesis_instrumentation__.Notify(627707)
		if err := runSchemaChangesInTxn(
			ctx, p, tableDesc, p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		); err != nil {
			__antithesis_instrumentation__.Notify(627708)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627709)
		}
	} else {
		__antithesis_instrumentation__.Notify(627710)
	}
	__antithesis_instrumentation__.Notify(627703)

	version := p.ExecCfg().Settings.Version.ActiveVersion(ctx)
	if err := descbuilder.ValidateSelf(tableDesc, version); err != nil {
		__antithesis_instrumentation__.Notify(627711)
		return errors.NewAssertionErrorWithWrappedErrf(err, "table descriptor is not valid\n%v\n", tableDesc)
	} else {
		__antithesis_instrumentation__.Notify(627712)
	}
	__antithesis_instrumentation__.Notify(627704)

	return p.Descriptors().WriteDescToBatch(
		ctx, p.extendedEvalCtx.Tracing.KVTracingEnabled(), tableDesc, b,
	)
}
