package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

type JobMetadataGetter interface {
	GetJobMetadata(jobspb.JobID) (*JobMetadata, error)
}

func ValidateJobReferencesInDescriptor(
	desc catalog.Descriptor, jmg JobMetadataGetter, errorAccFn func(error),
) {
	__antithesis_instrumentation__.Notify(85080)

	tbl, isTable := desc.(catalog.TableDescriptor)
	if !isTable {
		__antithesis_instrumentation__.Notify(85082)
		return
	} else {
		__antithesis_instrumentation__.Notify(85083)
	}
	__antithesis_instrumentation__.Notify(85081)

	for _, m := range tbl.GetMutationJobs() {
		__antithesis_instrumentation__.Notify(85084)
		j, err := jmg.GetJobMetadata(m.JobID)
		if err != nil {
			__antithesis_instrumentation__.Notify(85088)
			errorAccFn(errors.WithAssertionFailure(errors.Wrapf(err, "mutation job %d", m.JobID)))
			continue
		} else {
			__antithesis_instrumentation__.Notify(85089)
		}
		__antithesis_instrumentation__.Notify(85085)
		if j == nil {
			__antithesis_instrumentation__.Notify(85090)
			errorAccFn(errors.AssertionFailedf("mutation job %d not found in system.jobs", m.JobID))
			continue
		} else {
			__antithesis_instrumentation__.Notify(85091)
		}
		__antithesis_instrumentation__.Notify(85086)
		if j.Payload.Type() != jobspb.TypeSchemaChange {
			__antithesis_instrumentation__.Notify(85092)
			errorAccFn(errors.AssertionFailedf("mutation job %d is of type %q, expected schema change job", m.JobID, j.Payload.Type()))
		} else {
			__antithesis_instrumentation__.Notify(85093)
		}
		__antithesis_instrumentation__.Notify(85087)
		if j.Status.Terminal() {
			__antithesis_instrumentation__.Notify(85094)
			errorAccFn(errors.AssertionFailedf("mutation job %d has terminal status (%s)", m.JobID, j.Status))
		} else {
			__antithesis_instrumentation__.Notify(85095)
		}
	}
}

func ValidateDescriptorReferencesInJob(
	j JobMetadata, descLookupFn func(id descpb.ID) catalog.Descriptor, errorAccFn func(error),
) {
	__antithesis_instrumentation__.Notify(85096)
	switch j.Status {
	case StatusRunning, StatusPaused, StatusPauseRequested:
		__antithesis_instrumentation__.Notify(85100)

	default:
		__antithesis_instrumentation__.Notify(85101)
		return
	}
	__antithesis_instrumentation__.Notify(85097)
	existing := catalog.MakeDescriptorIDSet()
	missing := catalog.MakeDescriptorIDSet()
	for _, id := range collectDescriptorReferences(j).Ordered() {
		__antithesis_instrumentation__.Notify(85102)
		if descLookupFn(id) != nil {
			__antithesis_instrumentation__.Notify(85103)
			existing.Add(id)
		} else {
			__antithesis_instrumentation__.Notify(85104)
			if id != descpb.InvalidID {
				__antithesis_instrumentation__.Notify(85105)
				missing.Add(id)
			} else {
				__antithesis_instrumentation__.Notify(85106)
			}
		}
	}
	__antithesis_instrumentation__.Notify(85098)
	if missing.Len() == 0 {
		__antithesis_instrumentation__.Notify(85107)
		return
	} else {
		__antithesis_instrumentation__.Notify(85108)
	}
	__antithesis_instrumentation__.Notify(85099)
	switch j.Payload.Type() {
	case jobspb.TypeSchemaChange:
		__antithesis_instrumentation__.Notify(85109)
		errorAccFn(errors.AssertionFailedf("%s schema change refers to missing descriptor(s) %+v",
			j.Status, missing.Ordered()))
	case jobspb.TypeSchemaChangeGC:
		__antithesis_instrumentation__.Notify(85110)
		isSafeToDelete := existing.Len() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(85113)
			return len(j.Progress.GetSchemaChangeGC().Indexes) == 0 == true
		}() == true
		errorAccFn(errors.AssertionFailedf("%s schema change GC refers to missing table descriptor(s) %+v; "+
			"existing descriptors that still need to be dropped %+v; job safe to delete: %v",
			j.Status, missing.Ordered(), existing.Ordered(), isSafeToDelete))
	case jobspb.TypeTypeSchemaChange:
		__antithesis_instrumentation__.Notify(85111)
		errorAccFn(errors.AssertionFailedf("%s type schema change refers to missing type descriptor %v",
			j.Status, missing.Ordered()))
	default:
		__antithesis_instrumentation__.Notify(85112)
	}
}

func collectDescriptorReferences(j JobMetadata) (ids catalog.DescriptorIDSet) {
	__antithesis_instrumentation__.Notify(85114)
	switch j.Payload.Type() {
	case jobspb.TypeSchemaChange:
		__antithesis_instrumentation__.Notify(85116)
		sc := j.Payload.GetSchemaChange()
		ids.Add(sc.DescID)
		ids.Add(sc.DroppedDatabaseID)
		for _, schemaID := range sc.DroppedSchemas {
			__antithesis_instrumentation__.Notify(85122)
			ids.Add(schemaID)
		}
		__antithesis_instrumentation__.Notify(85117)
		for _, typeID := range sc.DroppedTypes {
			__antithesis_instrumentation__.Notify(85123)
			ids.Add(typeID)
		}
		__antithesis_instrumentation__.Notify(85118)
		for _, table := range sc.DroppedTables {
			__antithesis_instrumentation__.Notify(85124)
			ids.Add(table.ID)
		}
	case jobspb.TypeSchemaChangeGC:
		__antithesis_instrumentation__.Notify(85119)
		for _, table := range j.Progress.GetSchemaChangeGC().Tables {
			__antithesis_instrumentation__.Notify(85125)
			if table.Status == jobspb.SchemaChangeGCProgress_DELETED {
				__antithesis_instrumentation__.Notify(85127)
				continue
			} else {
				__antithesis_instrumentation__.Notify(85128)
			}
			__antithesis_instrumentation__.Notify(85126)
			ids.Add(table.ID)
		}
	case jobspb.TypeTypeSchemaChange:
		__antithesis_instrumentation__.Notify(85120)
		sc := j.Payload.GetTypeSchemaChange()
		ids.Add(sc.TypeID)
	default:
		__antithesis_instrumentation__.Notify(85121)
	}
	__antithesis_instrumentation__.Notify(85115)
	return ids
}
