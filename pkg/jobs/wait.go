package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (r *Registry) NotifyToAdoptJobs() {
	__antithesis_instrumentation__.Notify(85129)
	select {
	case r.adoptionCh <- resumeClaimedJobs:
		__antithesis_instrumentation__.Notify(85130)
	default:
		__antithesis_instrumentation__.Notify(85131)
	}
}

func (r *Registry) NotifyToResume(ctx context.Context, jobs ...jobspb.JobID) {
	__antithesis_instrumentation__.Notify(85132)
	m := newJobIDSet(jobs...)
	_ = r.stopper.RunAsyncTask(ctx, "resume-jobs", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(85133)
		r.withSession(ctx, func(ctx context.Context, s sqlliveness.Session) {
			__antithesis_instrumentation__.Notify(85134)
			r.filterAlreadyRunningAndCancelFromPreviousSessions(ctx, s, m)
			r.resumeClaimedJobs(ctx, s, m)
		})
	})
}

func (r *Registry) WaitForJobs(
	ctx context.Context, ex sqlutil.InternalExecutor, jobs []jobspb.JobID,
) error {
	__antithesis_instrumentation__.Notify(85135)
	log.Infof(ctx, "waiting for %d %v queued jobs to complete", len(jobs), jobs)
	jobFinishedLocally, cleanup := r.installWaitingSet(jobs...)
	defer cleanup()
	return r.waitForJobs(ctx, ex, jobs, jobFinishedLocally)
}

func (r *Registry) waitForJobs(
	ctx context.Context,
	ex sqlutil.InternalExecutor,
	jobs []jobspb.JobID,
	jobFinishedLocally <-chan struct{},
) error {
	__antithesis_instrumentation__.Notify(85136)

	if len(jobs) == 0 {
		__antithesis_instrumentation__.Notify(85141)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(85142)
	}
	__antithesis_instrumentation__.Notify(85137)

	query := makeWaitForJobsQuery(jobs)
	start := timeutil.Now()

	ret := retry.StartWithCtx(ctx, retry.Options{
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     3 * time.Second,
		Multiplier:     1.5,

		Closer: jobFinishedLocally,
	})
	ret.Next()
	for ret.Next() {
		__antithesis_instrumentation__.Notify(85143)

		row, err := ex.QueryRowEx(
			ctx,
			"poll-show-jobs",
			nil,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			query,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(85147)
			return errors.Wrap(err, "polling for queued jobs to complete")
		} else {
			__antithesis_instrumentation__.Notify(85148)
		}
		__antithesis_instrumentation__.Notify(85144)
		if row == nil {
			__antithesis_instrumentation__.Notify(85149)
			return errors.New("polling for queued jobs failed")
		} else {
			__antithesis_instrumentation__.Notify(85150)
		}
		__antithesis_instrumentation__.Notify(85145)
		count := int64(tree.MustBeDInt(row[0]))
		if log.V(3) {
			__antithesis_instrumentation__.Notify(85151)
			log.Infof(ctx, "waiting for %d queued jobs to complete", count)
		} else {
			__antithesis_instrumentation__.Notify(85152)
		}
		__antithesis_instrumentation__.Notify(85146)
		if count == 0 {
			__antithesis_instrumentation__.Notify(85153)
			break
		} else {
			__antithesis_instrumentation__.Notify(85154)
		}
	}
	__antithesis_instrumentation__.Notify(85138)
	defer func() {
		__antithesis_instrumentation__.Notify(85155)
		log.Infof(ctx, "waited for %d %v queued jobs to complete %v",
			len(jobs), jobs, timeutil.Since(start))
	}()
	__antithesis_instrumentation__.Notify(85139)
	for i, id := range jobs {
		__antithesis_instrumentation__.Notify(85156)
		j, err := r.LoadJob(ctx, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(85160)
			return errors.WithHint(
				errors.Wrapf(err, "job %d could not be loaded", jobs[i]),
				"The job may not have succeeded.")
		} else {
			__antithesis_instrumentation__.Notify(85161)
		}
		__antithesis_instrumentation__.Notify(85157)
		if j.Payload().FinalResumeError != nil {
			__antithesis_instrumentation__.Notify(85162)
			decodedErr := errors.DecodeError(ctx, *j.Payload().FinalResumeError)
			return decodedErr
		} else {
			__antithesis_instrumentation__.Notify(85163)
		}
		__antithesis_instrumentation__.Notify(85158)
		if j.Status() == StatusPaused {
			__antithesis_instrumentation__.Notify(85164)
			if reason := j.Payload().PauseReason; reason != "" {
				__antithesis_instrumentation__.Notify(85166)
				return errors.Newf("job %d was paused before it completed with reason: %s", jobs[i], reason)
			} else {
				__antithesis_instrumentation__.Notify(85167)
			}
			__antithesis_instrumentation__.Notify(85165)
			return errors.Newf("job %d was paused before it completed", jobs[i])
		} else {
			__antithesis_instrumentation__.Notify(85168)
		}
		__antithesis_instrumentation__.Notify(85159)
		if j.Payload().Error != "" {
			__antithesis_instrumentation__.Notify(85169)
			return errors.Newf("job %d failed with error: %s", jobs[i], j.Payload().Error)
		} else {
			__antithesis_instrumentation__.Notify(85170)
		}
	}
	__antithesis_instrumentation__.Notify(85140)
	return nil
}

func makeWaitForJobsQuery(jobs []jobspb.JobID) string {
	__antithesis_instrumentation__.Notify(85171)
	var buf strings.Builder
	buf.WriteString(`SELECT count(*) FROM system.jobs WHERE status NOT IN ( ` +
		`'` + string(StatusSucceeded) + `', ` +
		`'` + string(StatusFailed) + `',` +
		`'` + string(StatusCanceled) + `',` +
		`'` + string(StatusRevertFailed) + `',` +
		`'` + string(StatusPaused) + `'` +
		` ) AND id IN (`)
	for i, id := range jobs {
		__antithesis_instrumentation__.Notify(85173)
		if i > 0 {
			__antithesis_instrumentation__.Notify(85175)
			buf.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(85176)
		}
		__antithesis_instrumentation__.Notify(85174)
		_, _ = fmt.Fprintf(&buf, " %d", id)
	}
	__antithesis_instrumentation__.Notify(85172)
	buf.WriteString(")")
	return buf.String()
}

func (r *Registry) Run(
	ctx context.Context, ex sqlutil.InternalExecutor, jobs []jobspb.JobID,
) error {
	__antithesis_instrumentation__.Notify(85177)
	if len(jobs) == 0 {
		__antithesis_instrumentation__.Notify(85179)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(85180)
	}
	__antithesis_instrumentation__.Notify(85178)
	done, cleanup := r.installWaitingSet(jobs...)
	defer cleanup()
	r.NotifyToResume(ctx, jobs...)
	return r.waitForJobs(ctx, ex, jobs, done)
}

type jobWaitingSets map[jobspb.JobID]map[*waitingSet]struct{}

type waitingSet struct {
	jobDoneCh chan struct{}

	set jobIDSet
}

type jobIDSet map[jobspb.JobID]struct{}

func newJobIDSet(ids ...jobspb.JobID) jobIDSet {
	__antithesis_instrumentation__.Notify(85181)
	m := make(map[jobspb.JobID]struct{}, len(ids))
	for _, j := range ids {
		__antithesis_instrumentation__.Notify(85183)
		m[j] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(85182)
	return m
}

func (r *Registry) installWaitingSet(
	ids ...jobspb.JobID,
) (jobDoneCh <-chan struct{}, cleanup func()) {
	__antithesis_instrumentation__.Notify(85184)
	ws := &waitingSet{
		jobDoneCh: make(chan struct{}),
		set:       newJobIDSet(ids...),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(85187)
		sets, ok := r.mu.waiting[id]
		if !ok {
			__antithesis_instrumentation__.Notify(85189)
			sets = make(map[*waitingSet]struct{}, 1)
			r.mu.waiting[id] = sets
		} else {
			__antithesis_instrumentation__.Notify(85190)
		}
		__antithesis_instrumentation__.Notify(85188)
		sets[ws] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(85185)
	cleanup = func() {
		__antithesis_instrumentation__.Notify(85191)
		r.mu.Lock()
		defer r.mu.Unlock()
		for id := range ws.set {
			__antithesis_instrumentation__.Notify(85192)
			set, ok := r.mu.waiting[id]
			if !ok {
				__antithesis_instrumentation__.Notify(85194)

				log.Errorf(
					r.ac.AnnotateCtx(context.Background()),
					"corruption detected in waiting set for id %d", id,
				)
				continue
			} else {
				__antithesis_instrumentation__.Notify(85195)
			}
			__antithesis_instrumentation__.Notify(85193)
			delete(set, ws)
			delete(ws.set, id)
			if len(set) == 0 {
				__antithesis_instrumentation__.Notify(85196)
				delete(r.mu.waiting, id)
			} else {
				__antithesis_instrumentation__.Notify(85197)
			}
		}
	}
	__antithesis_instrumentation__.Notify(85186)
	return ws.jobDoneCh, cleanup
}

func (r *Registry) removeFromWaitingSets(id jobspb.JobID) {
	__antithesis_instrumentation__.Notify(85198)
	r.mu.Lock()
	defer r.mu.Unlock()
	sets, ok := r.mu.waiting[id]
	if !ok {
		__antithesis_instrumentation__.Notify(85201)
		return
	} else {
		__antithesis_instrumentation__.Notify(85202)
	}
	__antithesis_instrumentation__.Notify(85199)
	for ws := range sets {
		__antithesis_instrumentation__.Notify(85203)

		delete(ws.set, id)
		if len(ws.set) == 0 {
			__antithesis_instrumentation__.Notify(85204)
			close(ws.jobDoneCh)
		} else {
			__antithesis_instrumentation__.Notify(85205)
		}
	}
	__antithesis_instrumentation__.Notify(85200)
	delete(r.mu.waiting, id)
}
