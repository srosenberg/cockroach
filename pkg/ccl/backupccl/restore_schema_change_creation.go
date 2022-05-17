package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	descpb "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func jobDescriptionFromMutationID(
	tableDesc *descpb.TableDescriptor, id descpb.MutationID,
) (string, int, error) {
	__antithesis_instrumentation__.Notify(12463)
	var jobDescBuilder strings.Builder
	mutationCount := 0
	for _, m := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(12466)
		if m.MutationID == id {
			__antithesis_instrumentation__.Notify(12467)
			mutationCount++

			isPrimaryKeySwap := false
			switch m.Descriptor_.(type) {
			case *descpb.DescriptorMutation_PrimaryKeySwap:
				__antithesis_instrumentation__.Notify(12472)
				isPrimaryKeySwap = true
			}
			__antithesis_instrumentation__.Notify(12468)

			if isPrimaryKeySwap {
				__antithesis_instrumentation__.Notify(12473)

				jobDescBuilder.Reset()
			} else {
				__antithesis_instrumentation__.Notify(12474)
				if jobDescBuilder.Len() != 0 {
					__antithesis_instrumentation__.Notify(12475)
					jobDescBuilder.WriteString("; ")
				} else {
					__antithesis_instrumentation__.Notify(12476)
				}
			}
			__antithesis_instrumentation__.Notify(12469)

			if m.Rollback {
				__antithesis_instrumentation__.Notify(12477)
				jobDescBuilder.WriteString("rollback for ")
			} else {
				__antithesis_instrumentation__.Notify(12478)
			}
			__antithesis_instrumentation__.Notify(12470)

			jobDescBuilder.WriteString(fmt.Sprintf("schema change on %s ", tableDesc.Name))

			if !isPrimaryKeySwap {
				__antithesis_instrumentation__.Notify(12479)
				switch m.Direction {
				case descpb.DescriptorMutation_ADD:
					__antithesis_instrumentation__.Notify(12480)
					jobDescBuilder.WriteString("adding ")
				case descpb.DescriptorMutation_DROP:
					__antithesis_instrumentation__.Notify(12481)
					jobDescBuilder.WriteString("dropping ")
				default:
					__antithesis_instrumentation__.Notify(12482)
					return "", 0, errors.Newf("unsupported mutation %+v, while restoring table %+v", m, tableDesc)
				}
			} else {
				__antithesis_instrumentation__.Notify(12483)
			}
			__antithesis_instrumentation__.Notify(12471)

			switch t := m.Descriptor_.(type) {
			case *descpb.DescriptorMutation_Column:
				__antithesis_instrumentation__.Notify(12484)
				jobDescBuilder.WriteString("column ")
				jobDescBuilder.WriteString(t.Column.Name)
				if m.Direction == descpb.DescriptorMutation_ADD {
					__antithesis_instrumentation__.Notify(12490)
					jobDescBuilder.WriteString(" " + t.Column.Type.String())
				} else {
					__antithesis_instrumentation__.Notify(12491)
				}
			case *descpb.DescriptorMutation_Index:
				__antithesis_instrumentation__.Notify(12485)
				jobDescBuilder.WriteString("index ")
				jobDescBuilder.WriteString(t.Index.Name + " for " + tableDesc.Name + " (")
				jobDescBuilder.WriteString(strings.Join(t.Index.KeyColumnNames, ", "))
				jobDescBuilder.WriteString(")")
			case *descpb.DescriptorMutation_Constraint:
				__antithesis_instrumentation__.Notify(12486)
				jobDescBuilder.WriteString("constraint ")
				jobDescBuilder.WriteString(t.Constraint.Name)
			case *descpb.DescriptorMutation_PrimaryKeySwap:
				__antithesis_instrumentation__.Notify(12487)
				jobDescBuilder.WriteString("changing primary key to (")
				newIndexID := t.PrimaryKeySwap.NewPrimaryIndexId

				for _, otherMut := range tableDesc.Mutations {
					__antithesis_instrumentation__.Notify(12492)
					if indexMut, ok := otherMut.Descriptor_.(*descpb.DescriptorMutation_Index); ok && func() bool {
						__antithesis_instrumentation__.Notify(12493)
						return indexMut.Index.ID == newIndexID == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(12494)
						return otherMut.MutationID == m.MutationID == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(12495)
						return m.Direction == descpb.DescriptorMutation_ADD == true
					}() == true {
						__antithesis_instrumentation__.Notify(12496)
						jobDescBuilder.WriteString(strings.Join(indexMut.Index.KeyColumnNames, ", "))
					} else {
						__antithesis_instrumentation__.Notify(12497)
					}
				}
				__antithesis_instrumentation__.Notify(12488)
				jobDescBuilder.WriteString(")")
			default:
				__antithesis_instrumentation__.Notify(12489)
				return "", 0, errors.Newf("unsupported mutation %+v, while restoring table %+v", m, tableDesc)
			}
		} else {
			__antithesis_instrumentation__.Notify(12498)
		}
	}
	__antithesis_instrumentation__.Notify(12464)

	jobDesc := jobDescBuilder.String()
	if mutationCount == 0 {
		__antithesis_instrumentation__.Notify(12499)
		return "", 0, errors.Newf("could not find mutation %d on table %s (%d) while restoring", id, tableDesc.Name, tableDesc.ID)
	} else {
		__antithesis_instrumentation__.Notify(12500)
	}
	__antithesis_instrumentation__.Notify(12465)
	return jobDesc, mutationCount, nil
}

func createTypeChangeJobFromDesc(
	ctx context.Context,
	jr *jobs.Registry,
	codec keys.SQLCodec,
	txn *kv.Txn,
	username security.SQLUsername,
	typ catalog.TypeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(12501)

	var transitioningMembers [][]byte
	for _, member := range typ.TypeDesc().EnumMembers {
		__antithesis_instrumentation__.Notify(12504)
		if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY {
			__antithesis_instrumentation__.Notify(12505)
			transitioningMembers = append(transitioningMembers, member.PhysicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(12506)
		}
	}
	__antithesis_instrumentation__.Notify(12502)
	record := jobs.Record{
		Description:   fmt.Sprintf("RESTORING: type %d", typ.GetID()),
		Username:      username,
		DescriptorIDs: descpb.IDs{typ.GetID()},
		Details: jobspb.TypeSchemaChangeDetails{
			TypeID:               typ.GetID(),
			TransitioningMembers: transitioningMembers,
		},
		Progress: jobspb.TypeSchemaChangeProgress{},

		NonCancelable: true,
	}
	jobID := jr.MakeJobID()
	if _, err := jr.CreateJobWithTxn(ctx, record, jobID, txn); err != nil {
		__antithesis_instrumentation__.Notify(12507)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12508)
	}
	__antithesis_instrumentation__.Notify(12503)
	log.Infof(ctx, "queued new type schema change job %d for type %d", jobID, typ.GetID())
	return nil
}

func createSchemaChangeJobsFromMutations(
	ctx context.Context,
	jr *jobs.Registry,
	codec keys.SQLCodec,
	txn *kv.Txn,
	username security.SQLUsername,
	tableDesc *tabledesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(12509)
	mutationJobs := make([]descpb.TableDescriptor_MutationJob, 0, len(tableDesc.Mutations))
	seenMutations := make(map[descpb.MutationID]bool)
	for idx, mutation := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(12511)
		if seenMutations[mutation.MutationID] {
			__antithesis_instrumentation__.Notify(12516)

			continue
		} else {
			__antithesis_instrumentation__.Notify(12517)
		}
		__antithesis_instrumentation__.Notify(12512)
		mutationID := mutation.MutationID
		seenMutations[mutationID] = true
		jobDesc, mutationCount, err := jobDescriptionFromMutationID(tableDesc.TableDesc(), mutationID)
		if err != nil {
			__antithesis_instrumentation__.Notify(12518)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12519)
		}
		__antithesis_instrumentation__.Notify(12513)
		spanList := make([]jobspb.ResumeSpanList, mutationCount)
		for i := range spanList {
			__antithesis_instrumentation__.Notify(12520)
			mut := tableDesc.Mutations[idx+i]

			if idx := mut.GetIndex(); idx != nil && func() bool {
				__antithesis_instrumentation__.Notify(12521)
				return idx.UseDeletePreservingEncoding == true
			}() == true {
				__antithesis_instrumentation__.Notify(12522)
				spanList[i] = jobspb.ResumeSpanList{ResumeSpans: []roachpb.Span{tableDesc.IndexSpan(codec, idx.ID)}}
			} else {
				__antithesis_instrumentation__.Notify(12523)
				spanList[i] = jobspb.ResumeSpanList{ResumeSpans: []roachpb.Span{tableDesc.PrimaryIndexSpan(codec)}}
			}
		}
		__antithesis_instrumentation__.Notify(12514)
		jobRecord := jobs.Record{

			Description:   "RESTORING: " + jobDesc,
			Username:      username,
			DescriptorIDs: descpb.IDs{tableDesc.GetID()},
			Details: jobspb.SchemaChangeDetails{
				DescID:          tableDesc.ID,
				TableMutationID: mutationID,
				ResumeSpanList:  spanList,

				FormatVersion: jobspb.DatabaseJobFormatVersion,
			},
			Progress: jobspb.SchemaChangeProgress{},
		}
		jobID := jr.MakeJobID()
		if _, err := jr.CreateJobWithTxn(ctx, jobRecord, jobID, txn); err != nil {
			__antithesis_instrumentation__.Notify(12524)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12525)
		}
		__antithesis_instrumentation__.Notify(12515)
		newMutationJob := descpb.TableDescriptor_MutationJob{
			MutationID: mutationID,
			JobID:      jobID,
		}
		mutationJobs = append(mutationJobs, newMutationJob)

		log.Infof(ctx, "queued new schema change job %d for table %d, mutation %d",
			jobID, tableDesc.ID, mutationID)
	}
	__antithesis_instrumentation__.Notify(12510)
	tableDesc.MutationJobs = mutationJobs
	return nil
}
