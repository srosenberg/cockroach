package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

type controlJobsNode struct {
	rows          planNode
	desiredStatus jobs.Status
	numRows       int
	reason        string
}

var jobCommandToDesiredStatus = map[tree.JobCommand]jobs.Status{
	tree.CancelJob: jobs.StatusCanceled,
	tree.ResumeJob: jobs.StatusRunning,
	tree.PauseJob:  jobs.StatusPaused,
}

func (n *controlJobsNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(460494)
	return n.numRows, true
}

func (n *controlJobsNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(460495)
	userIsAdmin, err := params.p.HasAdminRole(params.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(460501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460502)
	}
	__antithesis_instrumentation__.Notify(460496)

	if !userIsAdmin {
		__antithesis_instrumentation__.Notify(460503)
		hasControlJob, err := params.p.HasRoleOption(params.ctx, roleoption.CONTROLJOB)
		if err != nil {
			__antithesis_instrumentation__.Notify(460505)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460506)
		}
		__antithesis_instrumentation__.Notify(460504)

		if !hasControlJob {
			__antithesis_instrumentation__.Notify(460507)
			return pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s does not have %s privilege",
				params.p.User(), roleoption.CONTROLJOB)
		} else {
			__antithesis_instrumentation__.Notify(460508)
		}
	} else {
		__antithesis_instrumentation__.Notify(460509)
	}
	__antithesis_instrumentation__.Notify(460497)

	if n.desiredStatus != jobs.StatusPaused && func() bool {
		__antithesis_instrumentation__.Notify(460510)
		return len(n.reason) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(460511)
		return errors.AssertionFailedf("status %v is not %v and thus does not support a reason %v",
			n.desiredStatus, jobs.StatusPaused, n.reason)
	} else {
		__antithesis_instrumentation__.Notify(460512)
	}
	__antithesis_instrumentation__.Notify(460498)

	reg := params.p.ExecCfg().JobRegistry
	for {
		__antithesis_instrumentation__.Notify(460513)
		ok, err := n.rows.Next(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(460522)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460523)
		}
		__antithesis_instrumentation__.Notify(460514)
		if !ok {
			__antithesis_instrumentation__.Notify(460524)
			break
		} else {
			__antithesis_instrumentation__.Notify(460525)
		}
		__antithesis_instrumentation__.Notify(460515)

		jobIDDatum := n.rows.Values()[0]
		if jobIDDatum == tree.DNull {
			__antithesis_instrumentation__.Notify(460526)
			continue
		} else {
			__antithesis_instrumentation__.Notify(460527)
		}
		__antithesis_instrumentation__.Notify(460516)

		jobID, ok := tree.AsDInt(jobIDDatum)
		if !ok {
			__antithesis_instrumentation__.Notify(460528)
			return errors.AssertionFailedf("%q: expected *DInt, found %T", jobIDDatum, jobIDDatum)
		} else {
			__antithesis_instrumentation__.Notify(460529)
		}
		__antithesis_instrumentation__.Notify(460517)

		job, err := reg.LoadJobWithTxn(params.ctx, jobspb.JobID(jobID), params.p.Txn())
		if err != nil {
			__antithesis_instrumentation__.Notify(460530)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460531)
		}
		__antithesis_instrumentation__.Notify(460518)

		if job != nil {
			__antithesis_instrumentation__.Notify(460532)
			owner := job.Payload().UsernameProto.Decode()

			if !userIsAdmin {
				__antithesis_instrumentation__.Notify(460533)
				ok, err := params.p.UserHasAdminRole(params.ctx, owner)
				if err != nil {
					__antithesis_instrumentation__.Notify(460535)
					return err
				} else {
					__antithesis_instrumentation__.Notify(460536)
				}
				__antithesis_instrumentation__.Notify(460534)

				if ok {
					__antithesis_instrumentation__.Notify(460537)
					return pgerror.Newf(pgcode.InsufficientPrivilege,
						"only admins can control jobs owned by other admins")
				} else {
					__antithesis_instrumentation__.Notify(460538)
				}
			} else {
				__antithesis_instrumentation__.Notify(460539)
			}
		} else {
			__antithesis_instrumentation__.Notify(460540)
		}
		__antithesis_instrumentation__.Notify(460519)

		switch n.desiredStatus {
		case jobs.StatusPaused:
			__antithesis_instrumentation__.Notify(460541)
			err = reg.PauseRequested(params.ctx, params.p.txn, jobspb.JobID(jobID), n.reason)
		case jobs.StatusRunning:
			__antithesis_instrumentation__.Notify(460542)
			err = reg.Unpause(params.ctx, params.p.txn, jobspb.JobID(jobID))
		case jobs.StatusCanceled:
			__antithesis_instrumentation__.Notify(460543)
			err = reg.CancelRequested(params.ctx, params.p.txn, jobspb.JobID(jobID))
		default:
			__antithesis_instrumentation__.Notify(460544)
			err = errors.AssertionFailedf("unhandled status %v", n.desiredStatus)
		}
		__antithesis_instrumentation__.Notify(460520)
		if err != nil {
			__antithesis_instrumentation__.Notify(460545)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460546)
		}
		__antithesis_instrumentation__.Notify(460521)
		n.numRows++
	}
	__antithesis_instrumentation__.Notify(460499)
	switch n.desiredStatus {
	case jobs.StatusPaused:
		__antithesis_instrumentation__.Notify(460547)
		telemetry.Inc(sqltelemetry.SchemaJobControlCounter("pause"))
	case jobs.StatusRunning:
		__antithesis_instrumentation__.Notify(460548)
		telemetry.Inc(sqltelemetry.SchemaJobControlCounter("resume"))
	case jobs.StatusCanceled:
		__antithesis_instrumentation__.Notify(460549)
		telemetry.Inc(sqltelemetry.SchemaJobControlCounter("cancel"))
	default:
		__antithesis_instrumentation__.Notify(460550)
	}
	__antithesis_instrumentation__.Notify(460500)
	return nil
}

func (*controlJobsNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(460551)
	return false, nil
}

func (*controlJobsNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(460552)
	return nil
}

func (n *controlJobsNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(460553)
	n.rows.Close(ctx)
}
