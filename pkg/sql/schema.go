package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func schemaExists(
	ctx context.Context, txn *kv.Txn, col *descs.Collection, parentID descpb.ID, schema string,
) (bool, descpb.ID, error) {
	__antithesis_instrumentation__.Notify(576992)

	if schema == tree.PublicSchema {
		__antithesis_instrumentation__.Notify(576996)
		return true, descpb.InvalidID, nil
	} else {
		__antithesis_instrumentation__.Notify(576997)
	}
	__antithesis_instrumentation__.Notify(576993)
	for _, s := range virtualSchemas {
		__antithesis_instrumentation__.Notify(576998)
		if s.name == schema {
			__antithesis_instrumentation__.Notify(576999)
			return true, descpb.InvalidID, nil
		} else {
			__antithesis_instrumentation__.Notify(577000)
		}
	}
	__antithesis_instrumentation__.Notify(576994)

	schemaID, err := col.Direct().LookupSchemaID(ctx, txn, parentID, schema)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(577001)
		return schemaID == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(577002)
		return false, descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(577003)
	}
	__antithesis_instrumentation__.Notify(576995)
	return true, schemaID, nil
}

func (p *planner) writeSchemaDesc(ctx context.Context, desc *schemadesc.Mutable) error {
	__antithesis_instrumentation__.Notify(577004)
	b := p.txn.NewBatch()
	if err := p.Descriptors().WriteDescToBatch(
		ctx, p.extendedEvalCtx.Tracing.KVTracingEnabled(), desc, b,
	); err != nil {
		__antithesis_instrumentation__.Notify(577006)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577007)
	}
	__antithesis_instrumentation__.Notify(577005)
	return p.txn.Run(ctx, b)
}

func (p *planner) writeSchemaDescChange(
	ctx context.Context, desc *schemadesc.Mutable, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(577008)
	record, recordExists := p.extendedEvalCtx.SchemaChangeJobRecords[desc.ID]
	if recordExists {
		__antithesis_instrumentation__.Notify(577010)

		record.AppendDescription(jobDesc)
		log.Infof(ctx, "job %d: updated job's specification for change on schema %d", record.JobID, desc.ID)
	} else {
		__antithesis_instrumentation__.Notify(577011)

		jobRecord := jobs.Record{
			JobID:         p.extendedEvalCtx.ExecCfg.JobRegistry.MakeJobID(),
			Description:   jobDesc,
			Username:      p.User(),
			DescriptorIDs: descpb.IDs{desc.ID},
			Details: jobspb.SchemaChangeDetails{
				DescID: desc.ID,

				FormatVersion: jobspb.DatabaseJobFormatVersion,
			},
			Progress:      jobspb.SchemaChangeProgress{},
			NonCancelable: true,
		}
		p.extendedEvalCtx.SchemaChangeJobRecords[desc.ID] = &jobRecord
		log.Infof(ctx, "queued new schema change job %d for schema %d", jobRecord.JobID, desc.ID)
	}
	__antithesis_instrumentation__.Notify(577009)

	return p.writeSchemaDesc(ctx, desc)
}
