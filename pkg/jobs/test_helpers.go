package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/tracedumper"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func (r *Registry) TestingNudgeAdoptionQueue() {
	__antithesis_instrumentation__.Notify(84924)
	r.adoptionCh <- claimAndResumeClaimedJobs
}

type config struct {
	jobID jobspb.JobID
}

type TestCreateAndStartJobOption func(*config)

func WithJobID(jobID jobspb.JobID) TestCreateAndStartJobOption {
	__antithesis_instrumentation__.Notify(84925)
	return func(c *config) {
		__antithesis_instrumentation__.Notify(84926)
		c.jobID = jobID
	}
}

func TestingCreateAndStartJob(
	ctx context.Context, r *Registry, db *kv.DB, record Record, opts ...TestCreateAndStartJobOption,
) (*StartableJob, error) {
	__antithesis_instrumentation__.Notify(84927)
	var rj *StartableJob
	c := config{
		jobID: r.MakeJobID(),
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(84931)
		opt(&c)
	}
	__antithesis_instrumentation__.Notify(84928)
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(84932)
		return r.CreateStartableJobWithTxn(ctx, &rj, c.jobID, txn, record)
	}); err != nil {
		__antithesis_instrumentation__.Notify(84933)
		if rj != nil {
			__antithesis_instrumentation__.Notify(84935)
			if cleanupErr := rj.CleanupOnRollback(ctx); cleanupErr != nil {
				__antithesis_instrumentation__.Notify(84936)
				log.Warningf(ctx, "failed to cleanup StartableJob: %v", cleanupErr)
			} else {
				__antithesis_instrumentation__.Notify(84937)
			}
		} else {
			__antithesis_instrumentation__.Notify(84938)
		}
		__antithesis_instrumentation__.Notify(84934)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84939)
	}
	__antithesis_instrumentation__.Notify(84929)
	err := rj.Start(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(84940)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84941)
	}
	__antithesis_instrumentation__.Notify(84930)
	return rj, nil
}

func TestingGetTraceDumpDir(r *Registry) string {
	__antithesis_instrumentation__.Notify(84942)
	if r.td == nil {
		__antithesis_instrumentation__.Notify(84944)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(84945)
	}
	__antithesis_instrumentation__.Notify(84943)
	return tracedumper.TestingGetTraceDumpDir(r.td)
}
