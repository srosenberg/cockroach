package scbackup

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
)

func CreateDeclarativeSchemaChangeJobs(
	ctx context.Context, registry *jobs.Registry, txn *kv.Txn, allMut nstree.Catalog,
) error {
	__antithesis_instrumentation__.Notify(579276)
	byJobID := make(map[catpb.JobID][]catalog.MutableDescriptor)
	_ = allMut.ForEachDescriptorEntry(func(d catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(579279)
		if s := d.GetDeclarativeSchemaChangerState(); s != nil {
			__antithesis_instrumentation__.Notify(579281)
			byJobID[s.JobID] = append(byJobID[s.JobID], d.(catalog.MutableDescriptor))
		} else {
			__antithesis_instrumentation__.Notify(579282)
		}
		__antithesis_instrumentation__.Notify(579280)
		return nil
	})
	__antithesis_instrumentation__.Notify(579277)
	var records []*jobs.Record
	for _, descs := range byJobID {
		__antithesis_instrumentation__.Notify(579283)

		newID := registry.MakeJobID()
		var descriptorStates []*scpb.DescriptorState
		for _, d := range descs {
			__antithesis_instrumentation__.Notify(579286)
			ds := d.GetDeclarativeSchemaChangerState()
			ds.JobID = newID
			descriptorStates = append(descriptorStates, ds)
		}
		__antithesis_instrumentation__.Notify(579284)
		currentState, err := scpb.MakeCurrentStateFromDescriptors(
			descriptorStates,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(579287)
			return err
		} else {
			__antithesis_instrumentation__.Notify(579288)
		}
		__antithesis_instrumentation__.Notify(579285)
		const runningStatus = "restored from backup"
		records = append(records, scexec.MakeDeclarativeSchemaChangeJobRecord(
			newID,
			currentState.Statements,
			!currentState.Revertible,
			currentState.Authorization,
			screl.AllTargetDescIDs(currentState.TargetState).Ordered(),
			runningStatus,
		))
	}
	__antithesis_instrumentation__.Notify(579278)
	_, err := registry.CreateJobsWithTxn(ctx, txn, records)
	return err
}
