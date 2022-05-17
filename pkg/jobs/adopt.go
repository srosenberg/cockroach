package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

const (
	claimableStatusList = `'` + string(StatusRunning) + `', ` +
		`'` + string(StatusPending) + `', ` +
		`'` + string(StatusCancelRequested) + `', ` +
		`'` + string(StatusPauseRequested) + `', ` +
		`'` + string(StatusReverting) + `'`

	claimableStatusTupleString = `(` + claimableStatusList + `)`

	nonTerminalStatusList = claimableStatusList + `, ` +
		`'` + string(StatusPaused) + `'`

	NonTerminalStatusTupleString = `(` + nonTerminalStatusList + `)`

	claimQuery = `
   UPDATE system.jobs
      SET claim_session_id = $1, claim_instance_id = $2
    WHERE ((claim_session_id IS NULL)
      AND (status IN ` + claimableStatusTupleString + `))
 ORDER BY created DESC
    LIMIT $3
RETURNING id;`
)

func (r *Registry) maybeDumpTrace(
	resumerCtx context.Context, resumer Resumer, jobID, traceID int64, jobErr error,
) {
	__antithesis_instrumentation__.Notify(70062)
	if _, ok := resumer.(TraceableJob); !ok || func() bool {
		__antithesis_instrumentation__.Notify(70066)
		return r.td == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70067)
		return
	} else {
		__antithesis_instrumentation__.Notify(70068)
	}
	__antithesis_instrumentation__.Notify(70063)
	dumpMode := traceableJobDumpTraceMode.Get(&r.settings.SV)
	if dumpMode == int64(noDump) {
		__antithesis_instrumentation__.Notify(70069)
		return
	} else {
		__antithesis_instrumentation__.Notify(70070)
	}
	__antithesis_instrumentation__.Notify(70064)

	dumpCtx, _ := r.makeCtx()

	if jobErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(70071)
		return !HasErrJobCanceled(jobErr) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70072)
		return resumerCtx.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70073)
		r.td.Dump(dumpCtx, strconv.Itoa(int(jobID)), traceID, r.ex)
		return
	} else {
		__antithesis_instrumentation__.Notify(70074)
	}
	__antithesis_instrumentation__.Notify(70065)

	if dumpMode == int64(dumpOnStop) {
		__antithesis_instrumentation__.Notify(70075)
		r.td.Dump(dumpCtx, strconv.Itoa(int(jobID)), traceID, r.ex)
	} else {
		__antithesis_instrumentation__.Notify(70076)
	}
}

func (r *Registry) claimJobs(ctx context.Context, s sqlliveness.Session) error {
	__antithesis_instrumentation__.Notify(70077)
	return r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(70078)

		if err := txn.SetUserPriority(roachpb.MinUserPriority); err != nil {
			__antithesis_instrumentation__.Notify(70082)
			return errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(70083)
		}
		__antithesis_instrumentation__.Notify(70079)
		numRows, err := r.ex.Exec(
			ctx, "claim-jobs", txn, claimQuery,
			s.ID().UnsafeBytes(), r.ID(), maxAdoptionsPerLoop)
		if err != nil {
			__antithesis_instrumentation__.Notify(70084)
			return errors.Wrap(err, "could not query jobs table")
		} else {
			__antithesis_instrumentation__.Notify(70085)
		}
		__antithesis_instrumentation__.Notify(70080)
		r.metrics.ClaimedJobs.Inc(int64(numRows))
		if log.ExpensiveLogEnabled(ctx, 1) || func() bool {
			__antithesis_instrumentation__.Notify(70086)
			return numRows > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(70087)
			log.Infof(ctx, "claimed %d jobs", numRows)
		} else {
			__antithesis_instrumentation__.Notify(70088)
		}
		__antithesis_instrumentation__.Notify(70081)
		return nil
	})
}

const (
	processQueryStatusTupleString = `(` +
		`'` + string(StatusRunning) + `', ` +
		`'` + string(StatusReverting) + `'` +
		`)`

	canRunArgs = `(SELECT $3::TIMESTAMP AS ts, $4::FLOAT AS initial_delay, $5::FLOAT AS max_delay) args`

	NextRunClause = `
COALESCE(last_run, created) + least(
	IF(
		args.initial_delay * (power(2, least(62, COALESCE(num_runs, 0))) - 1)::FLOAT >= 0.0,
		args.initial_delay * (power(2, least(62, COALESCE(num_runs, 0))) - 1)::FLOAT,
		args.max_delay
	),
	args.max_delay
)::INTERVAL`
	canRunClause = `args.ts >= ` + NextRunClause

	processQueryBase      = `SELECT id FROM system.jobs`
	processQueryWhereBase = ` status IN ` + processQueryStatusTupleString + ` AND (claim_session_id = $1 AND claim_instance_id = $2)`

	processQueryWithBackoff = processQueryBase + ", " + canRunArgs +
		" WHERE " + processQueryWhereBase + " AND " + canRunClause

	resumeQueryBaseCols    = "status, payload, progress, crdb_internal.sql_liveness_is_alive(claim_session_id)"
	resumeQueryWhereBase   = `id = $1 AND claim_session_id = $2`
	resumeQueryWithBackoff = `SELECT ` + resumeQueryBaseCols + `, ` + canRunClause + ` AS can_run,` +
		` created_by_type, created_by_id  FROM system.jobs, ` + canRunArgs + " WHERE " + resumeQueryWhereBase
)

func getProcessQuery(
	ctx context.Context, s sqlliveness.Session, r *Registry,
) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(70089)

	query := processQueryWithBackoff
	args := []interface{}{s.ID().UnsafeBytes(), r.ID(),
		r.clock.Now().GoTime(), r.RetryInitialDelay(), r.RetryMaxDelay()}
	return query, args
}

func (r *Registry) processClaimedJobs(ctx context.Context, s sqlliveness.Session) error {
	__antithesis_instrumentation__.Notify(70090)
	query, args := getProcessQuery(ctx, s, r)

	it, err := r.ex.QueryIteratorEx(
		ctx, "select-running/get-claimed-jobs", nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()}, query, args...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(70094)
		return errors.Wrapf(err, "could not query for claimed jobs")
	} else {
		__antithesis_instrumentation__.Notify(70095)
	}
	__antithesis_instrumentation__.Notify(70091)

	claimedToResume := make(map[jobspb.JobID]struct{})

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(70096)
		row := it.Cur()
		id := jobspb.JobID(*row[0].(*tree.DInt))
		claimedToResume[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(70092)
	if err != nil {
		__antithesis_instrumentation__.Notify(70097)
		return errors.Wrapf(err, "could not query for claimed jobs")
	} else {
		__antithesis_instrumentation__.Notify(70098)
	}
	__antithesis_instrumentation__.Notify(70093)
	r.filterAlreadyRunningAndCancelFromPreviousSessions(ctx, s, claimedToResume)
	r.resumeClaimedJobs(ctx, s, claimedToResume)
	return nil
}

func (r *Registry) resumeClaimedJobs(
	ctx context.Context, s sqlliveness.Session, claimedToResume map[jobspb.JobID]struct{},
) {
	__antithesis_instrumentation__.Notify(70099)
	const resumeConcurrency = 64
	sem := make(chan struct{}, resumeConcurrency)
	var wg sync.WaitGroup
	add := func() { __antithesis_instrumentation__.Notify(70103); sem <- struct{}{}; wg.Add(1) }
	__antithesis_instrumentation__.Notify(70100)
	done := func() { __antithesis_instrumentation__.Notify(70104); <-sem; wg.Done() }
	__antithesis_instrumentation__.Notify(70101)
	for id := range claimedToResume {
		__antithesis_instrumentation__.Notify(70105)
		add()
		go func(id jobspb.JobID) {
			__antithesis_instrumentation__.Notify(70106)
			defer done()
			if err := r.resumeJob(ctx, id, s); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(70107)
				return ctx.Err() == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(70108)
				log.Errorf(ctx, "could not run claimed job %d: %v", id, err)
			} else {
				__antithesis_instrumentation__.Notify(70109)
			}
		}(id)
	}
	__antithesis_instrumentation__.Notify(70102)
	wg.Wait()
}

func (r *Registry) filterAlreadyRunningAndCancelFromPreviousSessions(
	ctx context.Context, s sqlliveness.Session, claimedToResume map[jobspb.JobID]struct{},
) {
	__antithesis_instrumentation__.Notify(70110)
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, aj := range r.mu.adoptedJobs {
		__antithesis_instrumentation__.Notify(70111)
		if aj.session.ID() != s.ID() {
			__antithesis_instrumentation__.Notify(70112)
			log.Warningf(ctx, "job %d: running without having a live claim; canceling", id)
			aj.cancel()
			delete(r.mu.adoptedJobs, id)
		} else {
			__antithesis_instrumentation__.Notify(70113)
			if _, ok := claimedToResume[id]; ok {
				__antithesis_instrumentation__.Notify(70114)

				delete(claimedToResume, id)
				continue
			} else {
				__antithesis_instrumentation__.Notify(70115)
			}
		}
	}
}

func (r *Registry) resumeJob(ctx context.Context, jobID jobspb.JobID, s sqlliveness.Session) error {
	__antithesis_instrumentation__.Notify(70116)
	log.Infof(ctx, "job %d: resuming execution", jobID)
	resumeQuery := resumeQueryWithBackoff
	args := []interface{}{jobID, s.ID().UnsafeBytes(),
		r.clock.Now().GoTime(), r.RetryInitialDelay(), r.RetryMaxDelay()}
	row, err := r.ex.QueryRowEx(
		ctx, "get-job-row", nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()}, resumeQuery, args...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(70130)
		return errors.Wrapf(err, "job %d: could not query job table row", jobID)
	} else {
		__antithesis_instrumentation__.Notify(70131)
	}
	__antithesis_instrumentation__.Notify(70117)
	if row == nil {
		__antithesis_instrumentation__.Notify(70132)
		return errors.Errorf("job %d: claim with session id %s does not exist", jobID, s.ID())
	} else {
		__antithesis_instrumentation__.Notify(70133)
	}
	__antithesis_instrumentation__.Notify(70118)

	status := Status(*row[0].(*tree.DString))
	if status == StatusSucceeded {
		__antithesis_instrumentation__.Notify(70134)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(70135)
	}
	__antithesis_instrumentation__.Notify(70119)
	if status != StatusRunning && func() bool {
		__antithesis_instrumentation__.Notify(70136)
		return status != StatusReverting == true
	}() == true {
		__antithesis_instrumentation__.Notify(70137)

		return errors.Errorf("job %d: status changed to %s which is not resumable`", jobID, status)
	} else {
		__antithesis_instrumentation__.Notify(70138)
	}
	__antithesis_instrumentation__.Notify(70120)

	if isAlive := *row[3].(*tree.DBool); !isAlive {
		__antithesis_instrumentation__.Notify(70139)
		return errors.Errorf("job %d: claim with session id %s has expired", jobID, s.ID())
	} else {
		__antithesis_instrumentation__.Notify(70140)
	}
	__antithesis_instrumentation__.Notify(70121)

	if !(*row[4].(*tree.DBool)) {
		__antithesis_instrumentation__.Notify(70141)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(70142)
	}
	__antithesis_instrumentation__.Notify(70122)

	payload, err := UnmarshalPayload(row[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(70143)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70144)
	}
	__antithesis_instrumentation__.Notify(70123)

	if deprecatedIsOldSchemaChangeJob(payload) {
		__antithesis_instrumentation__.Notify(70145)
		log.VEventf(ctx, 2, "job %d: skipping adoption because schema change job has not been migrated", jobID)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(70146)
	}
	__antithesis_instrumentation__.Notify(70124)

	progress, err := UnmarshalProgress(row[2])
	if err != nil {
		__antithesis_instrumentation__.Notify(70147)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70148)
	}
	__antithesis_instrumentation__.Notify(70125)

	createdBy, err := unmarshalCreatedBy(row[5], row[6])
	if err != nil {
		__antithesis_instrumentation__.Notify(70149)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70150)
	}
	__antithesis_instrumentation__.Notify(70126)

	job := &Job{id: jobID, registry: r, createdBy: createdBy}
	job.mu.payload = *payload
	job.mu.progress = *progress
	job.session = s

	resumer, err := r.createResumer(job, r.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(70151)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70152)
	}
	__antithesis_instrumentation__.Notify(70127)
	resumeCtx, cancel := r.makeCtx()

	if alreadyAdopted := r.addAdoptedJob(jobID, s, cancel); alreadyAdopted {
		__antithesis_instrumentation__.Notify(70153)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(70154)
	}
	__antithesis_instrumentation__.Notify(70128)

	r.metrics.ResumedJobs.Inc(1)
	if err := r.stopper.RunAsyncTask(ctx, job.taskName(), func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(70155)

		_ = r.runJob(resumeCtx, resumer, job, status, job.taskName())
	}); err != nil {
		__antithesis_instrumentation__.Notify(70156)
		r.unregister(jobID)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70157)
	}
	__antithesis_instrumentation__.Notify(70129)
	return nil
}

func (r *Registry) addAdoptedJob(
	jobID jobspb.JobID, session sqlliveness.Session, cancel context.CancelFunc,
) (alreadyAdopted bool) {
	__antithesis_instrumentation__.Notify(70158)
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, alreadyAdopted = r.mu.adoptedJobs[jobID]; alreadyAdopted {
		__antithesis_instrumentation__.Notify(70160)
		return true
	} else {
		__antithesis_instrumentation__.Notify(70161)
	}
	__antithesis_instrumentation__.Notify(70159)

	r.mu.adoptedJobs[jobID] = &adoptedJob{
		session: session,
		cancel:  cancel,
		isIdle:  false,
	}
	return false
}

func (r *Registry) runJob(
	ctx context.Context, resumer Resumer, job *Job, status Status, taskName string,
) error {
	__antithesis_instrumentation__.Notify(70162)
	job.mu.Lock()
	var finalResumeError error
	if job.mu.payload.FinalResumeError != nil {
		__antithesis_instrumentation__.Notify(70168)
		finalResumeError = errors.DecodeError(ctx, *job.mu.payload.FinalResumeError)
	} else {
		__antithesis_instrumentation__.Notify(70169)
	}
	__antithesis_instrumentation__.Notify(70163)
	username := job.mu.payload.UsernameProto.Decode()
	typ := job.mu.payload.Type()
	job.mu.Unlock()

	defer r.unregister(job.ID())

	execCtx, cleanup := r.execCtx("resume-"+taskName, username)
	defer cleanup()

	var spanOptions []tracing.SpanOption
	if tj, ok := resumer.(TraceableJob); ok && func() bool {
		__antithesis_instrumentation__.Notify(70170)
		return tj.ForceRealSpan() == true
	}() == true {
		__antithesis_instrumentation__.Notify(70171)
		spanOptions = append(spanOptions, tracing.WithRecording(tracing.RecordingStructured))
	} else {
		__antithesis_instrumentation__.Notify(70172)
	}
	__antithesis_instrumentation__.Notify(70164)

	ctx, span := r.ac.Tracer.StartSpanCtx(ctx, typ.String(), spanOptions...)
	span.SetTag("job-id", attribute.Int64Value(int64(job.ID())))
	defer span.Finish()
	if span.TraceID() != 0 {
		__antithesis_instrumentation__.Notify(70173)
		if err := job.Update(ctx, nil, func(txn *kv.Txn, md JobMetadata,
			ju *JobUpdater) error {
			__antithesis_instrumentation__.Notify(70174)
			progress := *md.Progress
			progress.TraceID = span.TraceID()
			ju.UpdateProgress(&progress)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(70175)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70176)
		}
	} else {
		__antithesis_instrumentation__.Notify(70177)
	}
	__antithesis_instrumentation__.Notify(70165)

	err := r.stepThroughStateMachine(ctx, execCtx, resumer, job, status, finalResumeError)

	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(70178)
		return ctx.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70179)
		log.Errorf(ctx, "job %d: adoption completed with error %v", job.ID(), err)
	} else {
		__antithesis_instrumentation__.Notify(70180)
	}
	__antithesis_instrumentation__.Notify(70166)

	r.maybeDumpTrace(ctx, resumer, int64(job.ID()), int64(span.TraceID()), err)
	r.maybeRecordExecutionFailure(ctx, err, job)
	if r.knobs.AfterJobStateMachine != nil {
		__antithesis_instrumentation__.Notify(70181)
		r.knobs.AfterJobStateMachine()
	} else {
		__antithesis_instrumentation__.Notify(70182)
	}
	__antithesis_instrumentation__.Notify(70167)
	return err
}

const pauseAndCancelUpdate = `
   UPDATE system.jobs
      SET status = 
          CASE
						 WHEN status = '` + string(StatusPauseRequested) + `' THEN '` + string(StatusPaused) + `'
						 WHEN status = '` + string(StatusCancelRequested) + `' THEN '` + string(StatusReverting) + `'
						 ELSE status
          END,
					num_runs = 0,
					last_run = NULL
    WHERE (status IN ('` + string(StatusPauseRequested) + `', '` + string(StatusCancelRequested) + `'))
      AND ((claim_session_id = $1) AND (claim_instance_id = $2))
RETURNING id, status
`

func (r *Registry) servePauseAndCancelRequests(ctx context.Context, s sqlliveness.Session) error {
	__antithesis_instrumentation__.Notify(70183)
	return r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(70184)

		if err := txn.SetUserPriority(roachpb.MinUserPriority); err != nil {
			__antithesis_instrumentation__.Notify(70188)
			return errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(70189)
		}
		__antithesis_instrumentation__.Notify(70185)

		rows, err := r.ex.QueryBufferedEx(
			ctx, "cancel/pause-requested", txn, sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
			pauseAndCancelUpdate, s.ID().UnsafeBytes(), r.ID(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(70190)
			return errors.Wrap(err, "could not query jobs table")
		} else {
			__antithesis_instrumentation__.Notify(70191)
		}
		__antithesis_instrumentation__.Notify(70186)
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(70192)
			id := jobspb.JobID(*row[0].(*tree.DInt))
			job := &Job{id: id, registry: r}
			statusString := *row[1].(*tree.DString)
			switch Status(statusString) {
			case StatusPaused:
				__antithesis_instrumentation__.Notify(70193)
				r.cancelRegisteredJobContext(id)
				log.Infof(ctx, "job %d, session %s: paused", id, s.ID())
			case StatusReverting:
				__antithesis_instrumentation__.Notify(70194)
				if err := job.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
					__antithesis_instrumentation__.Notify(70197)
					r.cancelRegisteredJobContext(id)
					md.Payload.Error = errJobCanceled.Error()
					encodedErr := errors.EncodeError(ctx, errJobCanceled)
					md.Payload.FinalResumeError = &encodedErr
					ju.UpdatePayload(md.Payload)

					ju.UpdateRunStats(0, r.clock.Now().GoTime())
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(70198)
					return errors.Wrapf(err, "job %d: tried to cancel but could not mark as reverting", id)
				} else {
					__antithesis_instrumentation__.Notify(70199)
				}
				__antithesis_instrumentation__.Notify(70195)
				log.Infof(ctx, "job %d, session id: %s canceled: the job is now reverting",
					id, s.ID())
			default:
				__antithesis_instrumentation__.Notify(70196)
				return errors.AssertionFailedf("unexpected job status %s: %v", statusString, job)
			}
		}
		__antithesis_instrumentation__.Notify(70187)
		return nil
	})
}
