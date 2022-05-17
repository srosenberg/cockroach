package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/errors"
)

func (m *visitor) CreateSchemaChangerJob(
	ctx context.Context, job scop.CreateSchemaChangerJob,
) error {
	__antithesis_instrumentation__.Notify(582334)
	return m.s.AddNewSchemaChangerJob(
		job.JobID, job.Statements, job.NonCancelable, job.Authorization, job.DescriptorIDs, job.RunningStatus,
	)
}

func (m *visitor) UpdateSchemaChangerJob(
	ctx context.Context, op scop.UpdateSchemaChangerJob,
) error {
	__antithesis_instrumentation__.Notify(582335)
	return m.s.UpdateSchemaChangerJob(op.JobID, op.IsNonCancelable, op.RunningStatus)
}

func (m *visitor) SetJobStateOnDescriptor(
	ctx context.Context, op scop.SetJobStateOnDescriptor,
) error {
	__antithesis_instrumentation__.Notify(582336)
	mut, err := m.s.CheckOutDescriptor(ctx, op.DescriptorID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582340)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582341)
	}
	__antithesis_instrumentation__.Notify(582337)

	expected := op.State.JobID
	if op.Initialize {
		__antithesis_instrumentation__.Notify(582342)
		expected = jobspb.InvalidJobID
	} else {
		__antithesis_instrumentation__.Notify(582343)
	}
	__antithesis_instrumentation__.Notify(582338)
	scs := mut.GetDeclarativeSchemaChangerState()
	if scs != nil {
		__antithesis_instrumentation__.Notify(582344)
		if op.Initialize || func() bool {
			__antithesis_instrumentation__.Notify(582345)
			return scs.JobID != expected == true
		}() == true {
			__antithesis_instrumentation__.Notify(582346)
			return errors.AssertionFailedf(
				"unexpected schema change job ID %d on table %d, expected %d",
				scs.JobID, op.DescriptorID, expected,
			)
		} else {
			__antithesis_instrumentation__.Notify(582347)
		}
	} else {
		__antithesis_instrumentation__.Notify(582348)
	}
	__antithesis_instrumentation__.Notify(582339)
	mut.SetDeclarativeSchemaChangerState(op.State.Clone())
	return nil
}

func (m *visitor) RemoveJobStateFromDescriptor(
	ctx context.Context, op scop.RemoveJobStateFromDescriptor,
) error {
	__antithesis_instrumentation__.Notify(582349)
	{
		__antithesis_instrumentation__.Notify(582353)
		_, err := m.s.GetDescriptor(ctx, op.DescriptorID)

		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(582355)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(582356)
		}
		__antithesis_instrumentation__.Notify(582354)
		if err != nil {
			__antithesis_instrumentation__.Notify(582357)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582358)
		}
	}
	__antithesis_instrumentation__.Notify(582350)

	mut, err := m.s.CheckOutDescriptor(ctx, op.DescriptorID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582359)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582360)
	}
	__antithesis_instrumentation__.Notify(582351)
	existing := mut.GetDeclarativeSchemaChangerState()
	if existing == nil {
		__antithesis_instrumentation__.Notify(582361)
		return errors.AssertionFailedf(
			"expected schema change state with job ID %d on table %d, found none",
			op.JobID, op.DescriptorID,
		)
	} else {
		__antithesis_instrumentation__.Notify(582362)
		if existing.JobID != op.JobID {
			__antithesis_instrumentation__.Notify(582363)
			return errors.AssertionFailedf(
				"unexpected schema change job ID %d on table %d, expected %d",
				existing.JobID, op.DescriptorID, op.JobID,
			)
		} else {
			__antithesis_instrumentation__.Notify(582364)
		}
	}
	__antithesis_instrumentation__.Notify(582352)
	mut.SetDeclarativeSchemaChangerState(nil)
	return nil
}
