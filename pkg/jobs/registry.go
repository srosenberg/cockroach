package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/server/tracedumper"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/logtags"
)

type adoptedJob struct {
	session sqlliveness.Session
	isIdle  bool

	cancel context.CancelFunc
}

type adoptionNotice bool

const (
	resumeClaimedJobs         adoptionNotice = false
	claimAndResumeClaimedJobs adoptionNotice = true
)

type Registry struct {
	serverCtx context.Context

	ac        log.AmbientContext
	stopper   *stop.Stopper
	db        *kv.DB
	ex        sqlutil.InternalExecutor
	clock     *hlc.Clock
	clusterID *base.ClusterIDContainer
	nodeID    *base.SQLIDContainer
	settings  *cluster.Settings
	execCtx   jobExecCtxMaker
	metrics   Metrics
	td        *tracedumper.TraceDumper
	knobs     TestingKnobs

	adoptionCh  chan adoptionNotice
	sqlInstance sqlliveness.Instance

	sessionBoundInternalExecutorFactory sqlutil.SessionBoundInternalExecutorFactory

	preventAdoptionFile string

	mu struct {
		syncutil.Mutex

		adoptedJobs map[jobspb.JobID]*adoptedJob

		waiting jobWaitingSets
	}

	withSessionEvery log.EveryN

	TestingResumerCreationKnobs map[jobspb.Type]func(Resumer) Resumer
}

type jobExecCtxMaker func(opName string, user security.SQLUsername) (interface{}, func())

const PreventAdoptionFile = "DISABLE_STARTING_BACKGROUND_JOBS"

func MakeRegistry(
	ctx context.Context,
	ac log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
	clusterID *base.ClusterIDContainer,
	nodeID *base.SQLIDContainer,
	sqlInstance sqlliveness.Instance,
	settings *cluster.Settings,
	histogramWindowInterval time.Duration,
	execCtxFn jobExecCtxMaker,
	preventAdoptionFile string,
	td *tracedumper.TraceDumper,
	knobs *TestingKnobs,
) *Registry {
	__antithesis_instrumentation__.Notify(84254)
	r := &Registry{
		serverCtx:           ctx,
		ac:                  ac,
		stopper:             stopper,
		clock:               clock,
		db:                  db,
		ex:                  ex,
		clusterID:           clusterID,
		nodeID:              nodeID,
		sqlInstance:         sqlInstance,
		settings:            settings,
		execCtx:             execCtxFn,
		preventAdoptionFile: preventAdoptionFile,
		td:                  td,

		adoptionCh:       make(chan adoptionNotice, 1),
		withSessionEvery: log.Every(time.Second),
	}
	if knobs != nil {
		__antithesis_instrumentation__.Notify(84256)
		r.knobs = *knobs
		if knobs.TimeSource != nil {
			__antithesis_instrumentation__.Notify(84257)
			r.clock = knobs.TimeSource
		} else {
			__antithesis_instrumentation__.Notify(84258)
		}
	} else {
		__antithesis_instrumentation__.Notify(84259)
	}
	__antithesis_instrumentation__.Notify(84255)
	r.mu.adoptedJobs = make(map[jobspb.JobID]*adoptedJob)
	r.mu.waiting = make(map[jobspb.JobID]map[*waitingSet]struct{})
	r.metrics.init(histogramWindowInterval)
	return r
}

func (r *Registry) SetSessionBoundInternalExecutorFactory(
	factory sqlutil.SessionBoundInternalExecutorFactory,
) {
	__antithesis_instrumentation__.Notify(84260)
	r.sessionBoundInternalExecutorFactory = factory
}

func (r *Registry) MetricsStruct() *Metrics {
	__antithesis_instrumentation__.Notify(84261)
	return &r.metrics
}

func (r *Registry) CurrentlyRunningJobs() []jobspb.JobID {
	__antithesis_instrumentation__.Notify(84262)
	r.mu.Lock()
	defer r.mu.Unlock()
	jobs := make([]jobspb.JobID, 0, len(r.mu.adoptedJobs))
	for jID := range r.mu.adoptedJobs {
		__antithesis_instrumentation__.Notify(84264)
		jobs = append(jobs, jID)
	}
	__antithesis_instrumentation__.Notify(84263)
	return jobs
}

func (r *Registry) ID() base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(84265)
	return r.nodeID.SQLInstanceID()
}

func (r *Registry) makeCtx() (context.Context, func()) {
	__antithesis_instrumentation__.Notify(84266)
	ctx := r.ac.AnnotateCtx(context.Background())

	ctx = logtags.AddTags(ctx, logtags.FromContext(r.serverCtx))
	return context.WithCancel(ctx)
}

func (r *Registry) MakeJobID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(84267)
	return jobspb.JobID(builtins.GenerateUniqueInt(r.nodeID.SQLInstanceID()))
}

func (r *Registry) newJob(ctx context.Context, record Record) *Job {
	__antithesis_instrumentation__.Notify(84268)
	job := &Job{
		id:        record.JobID,
		registry:  r,
		createdBy: record.CreatedBy,
	}
	job.mu.payload = r.makePayload(ctx, &record)
	job.mu.progress = r.makeProgress(&record)
	job.mu.status = StatusRunning
	return job
}

func (r *Registry) makePayload(ctx context.Context, record *Record) jobspb.Payload {
	__antithesis_instrumentation__.Notify(84269)
	return jobspb.Payload{
		Description:            record.Description,
		Statement:              record.Statements,
		UsernameProto:          record.Username.EncodeProto(),
		DescriptorIDs:          record.DescriptorIDs,
		Details:                jobspb.WrapPayloadDetails(record.Details),
		Noncancelable:          record.NonCancelable,
		CreationClusterVersion: r.settings.Version.ActiveVersion(ctx).Version,
		CreationClusterID:      r.clusterID.Get(),
	}
}

func (r *Registry) makeProgress(record *Record) jobspb.Progress {
	__antithesis_instrumentation__.Notify(84270)
	return jobspb.Progress{
		Details:       jobspb.WrapProgressDetails(record.Progress),
		RunningStatus: string(record.RunningStatus),
	}
}

func (r *Registry) CreateJobsWithTxn(
	ctx context.Context, txn *kv.Txn, records []*Record,
) ([]jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(84271)
	created := make([]jobspb.JobID, 0, len(records))
	for toCreate := records; len(toCreate) > 0; {
		__antithesis_instrumentation__.Notify(84273)
		const maxBatchSize = 100
		batchSize := len(toCreate)
		if batchSize > maxBatchSize {
			__antithesis_instrumentation__.Notify(84276)
			batchSize = maxBatchSize
		} else {
			__antithesis_instrumentation__.Notify(84277)
		}
		__antithesis_instrumentation__.Notify(84274)
		createdInBatch, err := r.createJobsInBatchWithTxn(ctx, txn, toCreate[:batchSize])
		if err != nil {
			__antithesis_instrumentation__.Notify(84278)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(84279)
		}
		__antithesis_instrumentation__.Notify(84275)
		created = append(created, createdInBatch...)
		toCreate = toCreate[batchSize:]
	}
	__antithesis_instrumentation__.Notify(84272)
	return created, nil
}

func (r *Registry) createJobsInBatchWithTxn(
	ctx context.Context, txn *kv.Txn, records []*Record,
) ([]jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(84280)
	s, err := r.sqlInstance.Session(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(84285)
		return nil, errors.Wrap(err, "error getting live session")
	} else {
		__antithesis_instrumentation__.Notify(84286)
	}
	__antithesis_instrumentation__.Notify(84281)
	start := timeutil.Now()
	if txn != nil {
		__antithesis_instrumentation__.Notify(84287)
		start = txn.ReadTimestamp().GoTime()
	} else {
		__antithesis_instrumentation__.Notify(84288)
	}
	__antithesis_instrumentation__.Notify(84282)
	modifiedMicros := timeutil.ToUnixMicros(start)
	stmt, args, jobIDs, err := r.batchJobInsertStmt(ctx, s.ID(), records, modifiedMicros)
	if err != nil {
		__antithesis_instrumentation__.Notify(84289)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84290)
	}
	__antithesis_instrumentation__.Notify(84283)
	if _, err = r.ex.Exec(
		ctx, "job-rows-batch-insert", txn, stmt, args...,
	); err != nil {
		__antithesis_instrumentation__.Notify(84291)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84292)
	}
	__antithesis_instrumentation__.Notify(84284)
	return jobIDs, nil
}

func (r *Registry) batchJobInsertStmt(
	ctx context.Context, sessionID sqlliveness.SessionID, records []*Record, modifiedMicros int64,
) (string, []interface{}, []jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(84293)
	instanceID := r.ID()
	const numColumns = 6
	columns := [numColumns]string{`id`, `status`, `payload`, `progress`, `claim_session_id`, `claim_instance_id`}
	marshalPanic := func(m protoutil.Message) []byte {
		__antithesis_instrumentation__.Notify(84298)
		data, err := protoutil.Marshal(m)
		if err != nil {
			__antithesis_instrumentation__.Notify(84300)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(84301)
		}
		__antithesis_instrumentation__.Notify(84299)
		return data
	}
	__antithesis_instrumentation__.Notify(84294)
	valueFns := map[string]func(*Record) interface{}{
		`id`:     func(rec *Record) interface{} { __antithesis_instrumentation__.Notify(84302); return rec.JobID },
		`status`: func(rec *Record) interface{} { __antithesis_instrumentation__.Notify(84303); return StatusRunning },
		`claim_session_id`: func(rec *Record) interface{} {
			__antithesis_instrumentation__.Notify(84304)
			return sessionID.UnsafeBytes()
		},
		`claim_instance_id`: func(rec *Record) interface{} { __antithesis_instrumentation__.Notify(84305); return instanceID },
		`payload`: func(rec *Record) interface{} {
			__antithesis_instrumentation__.Notify(84306)
			payload := r.makePayload(ctx, rec)
			return marshalPanic(&payload)
		},
		`progress`: func(rec *Record) interface{} {
			__antithesis_instrumentation__.Notify(84307)
			progress := r.makeProgress(rec)
			progress.ModifiedMicros = modifiedMicros
			return marshalPanic(&progress)
		},
	}
	__antithesis_instrumentation__.Notify(84295)
	appendValues := func(rec *Record, vals *[]interface{}) (err error) {
		__antithesis_instrumentation__.Notify(84308)
		defer func() {
			__antithesis_instrumentation__.Notify(84311)
			switch r := recover(); r.(type) {
			case nil:
				__antithesis_instrumentation__.Notify(84312)
			case error:
				__antithesis_instrumentation__.Notify(84313)
				err = errors.CombineErrors(err, errors.Wrapf(r.(error), "encoding job %d", rec.JobID))
			default:
				__antithesis_instrumentation__.Notify(84314)
				panic(r)
			}
		}()
		__antithesis_instrumentation__.Notify(84309)
		for _, c := range columns {
			__antithesis_instrumentation__.Notify(84315)
			*vals = append(*vals, valueFns[c](rec))
		}
		__antithesis_instrumentation__.Notify(84310)
		return nil
	}
	__antithesis_instrumentation__.Notify(84296)
	args := make([]interface{}, 0, len(records)*numColumns)
	jobIDs := make([]jobspb.JobID, 0, len(records))
	var buf strings.Builder
	buf.WriteString(`INSERT INTO system.jobs (`)
	buf.WriteString(strings.Join(columns[:numColumns], ", "))
	buf.WriteString(`) VALUES `)
	argIdx := 1
	for i, rec := range records {
		__antithesis_instrumentation__.Notify(84316)
		if i > 0 {
			__antithesis_instrumentation__.Notify(84320)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(84321)
		}
		__antithesis_instrumentation__.Notify(84317)
		buf.WriteString("(")
		for j := range columns {
			__antithesis_instrumentation__.Notify(84322)
			if j > 0 {
				__antithesis_instrumentation__.Notify(84324)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(84325)
			}
			__antithesis_instrumentation__.Notify(84323)
			buf.WriteString("$")
			buf.WriteString(strconv.Itoa(argIdx))
			argIdx++
		}
		__antithesis_instrumentation__.Notify(84318)
		buf.WriteString(")")
		if err := appendValues(rec, &args); err != nil {
			__antithesis_instrumentation__.Notify(84326)
			return "", nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(84327)
		}
		__antithesis_instrumentation__.Notify(84319)
		jobIDs = append(jobIDs, rec.JobID)
	}
	__antithesis_instrumentation__.Notify(84297)
	return buf.String(), args, jobIDs, nil
}

func (r *Registry) CreateJobWithTxn(
	ctx context.Context, record Record, jobID jobspb.JobID, txn *kv.Txn,
) (*Job, error) {
	__antithesis_instrumentation__.Notify(84328)

	record.JobID = jobID
	j := r.newJob(ctx, record)

	s, err := r.sqlInstance.Session(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(84334)
		return nil, errors.Wrap(err, "error getting live session")
	} else {
		__antithesis_instrumentation__.Notify(84335)
	}
	__antithesis_instrumentation__.Notify(84329)
	j.session = s
	start := timeutil.Now()
	if txn != nil {
		__antithesis_instrumentation__.Notify(84336)
		start = txn.ReadTimestamp().GoTime()
	} else {
		__antithesis_instrumentation__.Notify(84337)
	}
	__antithesis_instrumentation__.Notify(84330)
	j.mu.progress.ModifiedMicros = timeutil.ToUnixMicros(start)
	payloadBytes, err := protoutil.Marshal(&j.mu.payload)
	if err != nil {
		__antithesis_instrumentation__.Notify(84338)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84339)
	}
	__antithesis_instrumentation__.Notify(84331)
	progressBytes, err := protoutil.Marshal(&j.mu.progress)
	if err != nil {
		__antithesis_instrumentation__.Notify(84340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84341)
	}
	__antithesis_instrumentation__.Notify(84332)
	if _, err = j.registry.ex.Exec(ctx, "job-row-insert", txn, `
INSERT INTO system.jobs (id, status, payload, progress, claim_session_id, claim_instance_id)
VALUES ($1, $2, $3, $4, $5, $6)`, jobID, StatusRunning, payloadBytes, progressBytes, s.ID().UnsafeBytes(), r.ID(),
	); err != nil {
		__antithesis_instrumentation__.Notify(84342)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84343)
	}
	__antithesis_instrumentation__.Notify(84333)
	return j, nil
}

func (r *Registry) CreateAdoptableJobWithTxn(
	ctx context.Context, record Record, jobID jobspb.JobID, txn *kv.Txn,
) (*Job, error) {
	__antithesis_instrumentation__.Notify(84344)

	record.JobID = jobID
	j := r.newJob(ctx, record)
	if err := j.runInTxn(ctx, txn, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(84346)

		j.mu.progress.ModifiedMicros = timeutil.ToUnixMicros(txn.ReadTimestamp().GoTime())
		payloadBytes, err := protoutil.Marshal(&j.mu.payload)
		if err != nil {
			__antithesis_instrumentation__.Notify(84350)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84351)
		}
		__antithesis_instrumentation__.Notify(84347)
		progressBytes, err := protoutil.Marshal(&j.mu.progress)
		if err != nil {
			__antithesis_instrumentation__.Notify(84352)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84353)
		}
		__antithesis_instrumentation__.Notify(84348)

		var createdByType, createdByID interface{}
		if j.createdBy != nil {
			__antithesis_instrumentation__.Notify(84354)
			createdByType = j.createdBy.Name
			createdByID = j.createdBy.ID
		} else {
			__antithesis_instrumentation__.Notify(84355)
		}
		__antithesis_instrumentation__.Notify(84349)

		const stmt = `INSERT
  INTO system.jobs (
                    id,
                    status,
                    payload,
                    progress,
                    created_by_type,
                    created_by_id
                   )
VALUES ($1, $2, $3, $4, $5, $6);`
		_, err = j.registry.ex.Exec(ctx, "job-insert", txn, stmt,
			jobID, StatusRunning, payloadBytes, progressBytes, createdByType, createdByID)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(84356)
		return nil, errors.Wrap(err, "CreateAdoptableJobInTxn")
	} else {
		__antithesis_instrumentation__.Notify(84357)
	}
	__antithesis_instrumentation__.Notify(84345)
	return j, nil
}

func (r *Registry) CreateStartableJobWithTxn(
	ctx context.Context, sj **StartableJob, jobID jobspb.JobID, txn *kv.Txn, record Record,
) error {
	__antithesis_instrumentation__.Notify(84358)
	alreadyInitialized := *sj != nil
	if alreadyInitialized {
		__antithesis_instrumentation__.Notify(84364)
		if jobID != (*sj).Job.ID() {
			__antithesis_instrumentation__.Notify(84365)
			log.Fatalf(ctx,
				"attempted to rewrite startable job for ID %d with unexpected ID %d",
				(*sj).Job.ID(), jobID,
			)
		} else {
			__antithesis_instrumentation__.Notify(84366)
		}
	} else {
		__antithesis_instrumentation__.Notify(84367)
	}
	__antithesis_instrumentation__.Notify(84359)

	j, err := r.CreateJobWithTxn(ctx, record, jobID, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(84368)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84369)
	}
	__antithesis_instrumentation__.Notify(84360)
	resumer, err := r.createResumer(j, r.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(84370)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84371)
	}
	__antithesis_instrumentation__.Notify(84361)

	var resumerCtx context.Context
	var cancel func()
	var execDone chan struct{}
	if !alreadyInitialized {
		__antithesis_instrumentation__.Notify(84372)

		resumerCtx, cancel = r.makeCtx()

		if alreadyAdopted := r.addAdoptedJob(jobID, j.session, cancel); alreadyAdopted {
			__antithesis_instrumentation__.Notify(84374)
			log.Fatalf(
				ctx,
				"job %d: was just created but found in registered adopted jobs",
				jobID,
			)
		} else {
			__antithesis_instrumentation__.Notify(84375)
		}
		__antithesis_instrumentation__.Notify(84373)
		execDone = make(chan struct{})
	} else {
		__antithesis_instrumentation__.Notify(84376)
	}
	__antithesis_instrumentation__.Notify(84362)

	if !alreadyInitialized {
		__antithesis_instrumentation__.Notify(84377)
		*sj = &StartableJob{}
		(*sj).resumerCtx = resumerCtx
		(*sj).cancel = cancel
		(*sj).execDone = execDone
	} else {
		__antithesis_instrumentation__.Notify(84378)
	}
	__antithesis_instrumentation__.Notify(84363)
	(*sj).Job = j
	(*sj).resumer = resumer
	(*sj).txn = txn
	return nil
}

func (r *Registry) LoadJob(ctx context.Context, jobID jobspb.JobID) (*Job, error) {
	__antithesis_instrumentation__.Notify(84379)
	return r.LoadJobWithTxn(ctx, jobID, nil)
}

func (r *Registry) LoadClaimedJob(ctx context.Context, jobID jobspb.JobID) (*Job, error) {
	__antithesis_instrumentation__.Notify(84380)
	j, err := r.getClaimedJob(jobID)
	if err != nil {
		__antithesis_instrumentation__.Notify(84383)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84384)
	}
	__antithesis_instrumentation__.Notify(84381)
	if err := j.load(ctx, nil); err != nil {
		__antithesis_instrumentation__.Notify(84385)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84386)
	}
	__antithesis_instrumentation__.Notify(84382)
	return j, nil
}

func (r *Registry) LoadJobWithTxn(
	ctx context.Context, jobID jobspb.JobID, txn *kv.Txn,
) (*Job, error) {
	__antithesis_instrumentation__.Notify(84387)
	j := &Job{
		id:       jobID,
		registry: r,
	}
	if err := j.load(ctx, txn); err != nil {
		__antithesis_instrumentation__.Notify(84389)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84390)
	}
	__antithesis_instrumentation__.Notify(84388)
	return j, nil
}

func (r *Registry) UpdateJobWithTxn(
	ctx context.Context, jobID jobspb.JobID, txn *kv.Txn, useReadLock bool, updateFunc UpdateFn,
) error {
	__antithesis_instrumentation__.Notify(84391)
	j := &Job{
		id:       jobID,
		registry: r,
	}
	return j.update(ctx, txn, useReadLock, updateFunc)
}

var maxAdoptionsPerLoop = envutil.EnvOrDefaultInt(`COCKROACH_JOB_ADOPTIONS_PER_PERIOD`, 10)

const removeClaimsForDeadSessionsQuery = `
UPDATE system.jobs
   SET claim_session_id = NULL
 WHERE claim_session_id in (
SELECT claim_session_id
 WHERE claim_session_id <> $1
   AND status IN ` + claimableStatusTupleString + `
   AND NOT crdb_internal.sql_liveness_is_alive(claim_session_id)
 FETCH FIRST $2 ROWS ONLY)
`
const removeClaimsForSessionQuery = `
UPDATE system.jobs
   SET claim_session_id = NULL
 WHERE claim_session_id in (
SELECT claim_session_id
 WHERE claim_session_id = $1
   AND status IN ` + claimableStatusTupleString + `
)`

type withSessionFunc func(ctx context.Context, s sqlliveness.Session)

func (r *Registry) withSession(ctx context.Context, f withSessionFunc) {
	__antithesis_instrumentation__.Notify(84392)
	s, err := r.sqlInstance.Session(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(84394)
		if log.ExpensiveLogEnabled(ctx, 2) || func() bool {
			__antithesis_instrumentation__.Notify(84396)
			return (ctx.Err() == nil && func() bool {
				__antithesis_instrumentation__.Notify(84397)
				return r.withSessionEvery.ShouldLog() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(84398)
			log.Errorf(ctx, "error getting live session: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(84399)
		}
		__antithesis_instrumentation__.Notify(84395)
		return
	} else {
		__antithesis_instrumentation__.Notify(84400)
	}
	__antithesis_instrumentation__.Notify(84393)

	log.VEventf(ctx, 1, "registry live claim (instance_id: %s, sid: %s)", r.ID(), s.ID())
	f(ctx, s)
}

func (r *Registry) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(84401)
	wrapWithSession := func(f withSessionFunc) func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(84411)
		return func(ctx context.Context) { __antithesis_instrumentation__.Notify(84412); r.withSession(ctx, f) }
	}
	__antithesis_instrumentation__.Notify(84402)

	removeClaimsFromDeadSessions := func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84413)
		if err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(84414)

			if err := txn.SetUserPriority(roachpb.MinUserPriority); err != nil {
				__antithesis_instrumentation__.Notify(84416)
				return errors.WithAssertionFailure(err)
			} else {
				__antithesis_instrumentation__.Notify(84417)
			}
			__antithesis_instrumentation__.Notify(84415)
			_, err := r.ex.ExecEx(
				ctx, "expire-sessions", nil,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				removeClaimsForDeadSessionsQuery,
				s.ID().UnsafeBytes(),
				cancellationsUpdateLimitSetting.Get(&r.settings.SV),
			)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(84418)
			log.Errorf(ctx, "error expiring job sessions: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(84419)
		}
	}
	__antithesis_instrumentation__.Notify(84403)

	servePauseAndCancelRequests := func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84420)
		if err := r.servePauseAndCancelRequests(ctx, s); err != nil {
			__antithesis_instrumentation__.Notify(84421)
			log.Errorf(ctx, "failed to serve pause and cancel requests: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(84422)
		}
	}
	__antithesis_instrumentation__.Notify(84404)
	cancelLoopTask := wrapWithSession(func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84423)
		removeClaimsFromDeadSessions(ctx, s)
		r.maybeCancelJobs(ctx, s)
		servePauseAndCancelRequests(ctx, s)
	})
	__antithesis_instrumentation__.Notify(84405)

	claimJobs := wrapWithSession(func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84424)
		if r.adoptionDisabled(ctx) {
			__antithesis_instrumentation__.Notify(84426)
			log.Warningf(ctx, "job adoption is disabled, registry will not claim any jobs")
			return
		} else {
			__antithesis_instrumentation__.Notify(84427)
		}
		__antithesis_instrumentation__.Notify(84425)
		r.metrics.AdoptIterations.Inc(1)
		if err := r.claimJobs(ctx, s); err != nil {
			__antithesis_instrumentation__.Notify(84428)
			log.Errorf(ctx, "error claiming jobs: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(84429)
		}
	})
	__antithesis_instrumentation__.Notify(84406)

	removeClaimsFromSession := func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84430)
		if err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(84431)

			if err := txn.SetUserPriority(roachpb.MinUserPriority); err != nil {
				__antithesis_instrumentation__.Notify(84433)
				return errors.WithAssertionFailure(err)
			} else {
				__antithesis_instrumentation__.Notify(84434)
			}
			__antithesis_instrumentation__.Notify(84432)
			_, err := r.ex.ExecEx(
				ctx, "remove-claims-for-session", nil,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				removeClaimsForSessionQuery, s.ID().UnsafeBytes(),
			)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(84435)
			log.Errorf(ctx, "error expiring job sessions: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(84436)
		}
	}
	__antithesis_instrumentation__.Notify(84407)

	processClaimedJobs := wrapWithSession(func(ctx context.Context, s sqlliveness.Session) {
		__antithesis_instrumentation__.Notify(84437)

		if r.adoptionDisabled(ctx) {
			__antithesis_instrumentation__.Notify(84439)
			log.Warningf(ctx, "job adoptions is disabled, canceling all adopted jobs due to liveness failure")
			removeClaimsFromSession(ctx, s)
			r.cancelAllAdoptedJobs()
			return
		} else {
			__antithesis_instrumentation__.Notify(84440)
		}
		__antithesis_instrumentation__.Notify(84438)
		if err := r.processClaimedJobs(ctx, s); err != nil {
			__antithesis_instrumentation__.Notify(84441)
			log.Errorf(ctx, "error processing claimed jobs: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(84442)
		}
	})
	__antithesis_instrumentation__.Notify(84408)

	if err := stopper.RunAsyncTask(ctx, "jobs/cancel", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(84443)
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()

		cancelLoopTask(ctx)
		lc, cleanup := makeLoopController(r.settings, cancelIntervalSetting, r.knobs.IntervalOverrides.Cancel)
		defer cleanup()
		for {
			__antithesis_instrumentation__.Notify(84444)
			select {
			case <-lc.updated:
				__antithesis_instrumentation__.Notify(84445)
				lc.onUpdate()
			case <-r.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(84446)
				log.Warningf(ctx, "canceling all adopted jobs due to stopper quiescing")
				r.cancelAllAdoptedJobs()
				return
			case <-lc.timer.C:
				__antithesis_instrumentation__.Notify(84447)
				lc.timer.Read = true
				cancelLoopTask(ctx)
				lc.onExecute()
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(84448)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84449)
	}
	__antithesis_instrumentation__.Notify(84409)
	if err := stopper.RunAsyncTask(ctx, "jobs/gc", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(84450)
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()

		lc, cleanup := makeLoopController(r.settings, gcIntervalSetting, r.knobs.IntervalOverrides.Gc)
		defer cleanup()

		retentionDuration := func() time.Duration {
			__antithesis_instrumentation__.Notify(84452)
			if r.knobs.IntervalOverrides.RetentionTime != nil {
				__antithesis_instrumentation__.Notify(84454)
				return *r.knobs.IntervalOverrides.RetentionTime
			} else {
				__antithesis_instrumentation__.Notify(84455)
			}
			__antithesis_instrumentation__.Notify(84453)
			return retentionTimeSetting.Get(&r.settings.SV)
		}
		__antithesis_instrumentation__.Notify(84451)

		for {
			__antithesis_instrumentation__.Notify(84456)
			select {
			case <-lc.updated:
				__antithesis_instrumentation__.Notify(84457)
				lc.onUpdate()
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(84458)
				return
			case <-lc.timer.C:
				__antithesis_instrumentation__.Notify(84459)
				lc.timer.Read = true
				old := timeutil.Now().Add(-1 * retentionDuration())
				if err := r.cleanupOldJobs(ctx, old); err != nil {
					__antithesis_instrumentation__.Notify(84461)
					log.Warningf(ctx, "error cleaning up old job records: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(84462)
				}
				__antithesis_instrumentation__.Notify(84460)
				lc.onExecute()
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(84463)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84464)
	}
	__antithesis_instrumentation__.Notify(84410)
	return stopper.RunAsyncTask(ctx, "jobs/adopt", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(84465)
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()
		lc, cleanup := makeLoopController(r.settings, adoptIntervalSetting, r.knobs.IntervalOverrides.Adopt)
		defer cleanup()
		for {
			__antithesis_instrumentation__.Notify(84466)
			select {
			case <-lc.updated:
				__antithesis_instrumentation__.Notify(84467)
				lc.onUpdate()
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(84468)
				return
			case shouldClaim := <-r.adoptionCh:
				__antithesis_instrumentation__.Notify(84469)

				if shouldClaim {
					__antithesis_instrumentation__.Notify(84472)
					claimJobs(ctx)
				} else {
					__antithesis_instrumentation__.Notify(84473)
				}
				__antithesis_instrumentation__.Notify(84470)
				processClaimedJobs(ctx)
			case <-lc.timer.C:
				__antithesis_instrumentation__.Notify(84471)
				lc.timer.Read = true
				claimJobs(ctx)
				processClaimedJobs(ctx)
				lc.onExecute()
			}
		}
	})
}

func (r *Registry) maybeCancelJobs(ctx context.Context, s sqlliveness.Session) {
	__antithesis_instrumentation__.Notify(84474)
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, aj := range r.mu.adoptedJobs {
		__antithesis_instrumentation__.Notify(84475)
		if aj.session.ID() != s.ID() {
			__antithesis_instrumentation__.Notify(84476)
			log.Warningf(ctx, "job %d: running without having a live claim; killed.", id)
			aj.cancel()
			delete(r.mu.adoptedJobs, id)
		} else {
			__antithesis_instrumentation__.Notify(84477)
		}
	}
}

const cleanupPageSize = 100

func (r *Registry) cleanupOldJobs(ctx context.Context, olderThan time.Time) error {
	__antithesis_instrumentation__.Notify(84478)
	var maxID jobspb.JobID
	for {
		__antithesis_instrumentation__.Notify(84479)
		var done bool
		var err error
		done, maxID, err = r.cleanupOldJobsPage(ctx, olderThan, maxID, cleanupPageSize)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(84480)
			return done == true
		}() == true {
			__antithesis_instrumentation__.Notify(84481)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84482)
		}
	}
}

const expiredJobsQuery = "SELECT id, payload, status, created FROM system.jobs " +
	"WHERE (created < $1) AND (id > $2) " +
	"ORDER BY id " +
	"LIMIT $3"

func (r *Registry) cleanupOldJobsPage(
	ctx context.Context, olderThan time.Time, minID jobspb.JobID, pageSize int,
) (done bool, maxID jobspb.JobID, retErr error) {
	__antithesis_instrumentation__.Notify(84483)
	it, err := r.ex.QueryIterator(ctx, "gc-jobs", nil, expiredJobsQuery, olderThan, minID, pageSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(84490)
		return false, 0, err
	} else {
		__antithesis_instrumentation__.Notify(84491)
	}
	__antithesis_instrumentation__.Notify(84484)

	defer func() {
		__antithesis_instrumentation__.Notify(84492)
		retErr = errors.CombineErrors(retErr, it.Close())
	}()
	__antithesis_instrumentation__.Notify(84485)
	toDelete := tree.NewDArray(types.Int)
	oldMicros := timeutil.ToUnixMicros(olderThan)

	var ok bool
	var numRows int
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(84493)
		numRows++
		row := it.Cur()
		payload, err := UnmarshalPayload(row[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(84496)
			return false, 0, err
		} else {
			__antithesis_instrumentation__.Notify(84497)
		}
		__antithesis_instrumentation__.Notify(84494)
		remove := false
		switch Status(*row[2].(*tree.DString)) {
		case StatusSucceeded, StatusCanceled, StatusFailed:
			__antithesis_instrumentation__.Notify(84498)
			remove = payload.FinishedMicros < oldMicros
		default:
			__antithesis_instrumentation__.Notify(84499)
		}
		__antithesis_instrumentation__.Notify(84495)
		if remove {
			__antithesis_instrumentation__.Notify(84500)
			toDelete.Array = append(toDelete.Array, row[0])
		} else {
			__antithesis_instrumentation__.Notify(84501)
		}
	}
	__antithesis_instrumentation__.Notify(84486)
	if err != nil {
		__antithesis_instrumentation__.Notify(84502)
		return false, 0, err
	} else {
		__antithesis_instrumentation__.Notify(84503)
	}
	__antithesis_instrumentation__.Notify(84487)
	if numRows == 0 {
		__antithesis_instrumentation__.Notify(84504)
		return true, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(84505)
	}
	__antithesis_instrumentation__.Notify(84488)

	log.VEventf(ctx, 2, "read potentially expired jobs: %d", numRows)
	if len(toDelete.Array) > 0 {
		__antithesis_instrumentation__.Notify(84506)
		log.Infof(ctx, "attempting to clean up %d expired job records", len(toDelete.Array))
		const stmt = `DELETE FROM system.jobs WHERE id = ANY($1)`
		var nDeleted int
		if nDeleted, err = r.ex.Exec(
			ctx, "gc-jobs", nil, stmt, toDelete,
		); err != nil {
			__antithesis_instrumentation__.Notify(84508)
			return false, 0, errors.Wrap(err, "deleting old jobs")
		} else {
			__antithesis_instrumentation__.Notify(84509)
		}
		__antithesis_instrumentation__.Notify(84507)
		log.Infof(ctx, "cleaned up %d expired job records", nDeleted)
	} else {
		__antithesis_instrumentation__.Notify(84510)
	}
	__antithesis_instrumentation__.Notify(84489)

	morePages := numRows == pageSize

	lastRow := it.Cur()
	maxID = jobspb.JobID(*(lastRow[0].(*tree.DInt)))
	return !morePages, maxID, nil
}

func (r *Registry) getJobFn(
	ctx context.Context, txn *kv.Txn, id jobspb.JobID,
) (*Job, Resumer, error) {
	__antithesis_instrumentation__.Notify(84511)
	job, err := r.LoadJobWithTxn(ctx, id, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(84514)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(84515)
	}
	__antithesis_instrumentation__.Notify(84512)
	resumer, err := r.createResumer(job, r.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(84516)
		return job, nil, errors.Errorf("job %d is not controllable", id)
	} else {
		__antithesis_instrumentation__.Notify(84517)
	}
	__antithesis_instrumentation__.Notify(84513)
	return job, resumer, nil
}

func (r *Registry) CancelRequested(ctx context.Context, txn *kv.Txn, id jobspb.JobID) error {
	__antithesis_instrumentation__.Notify(84518)
	job, _, err := r.getJobFn(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(84520)

		if job != nil {
			__antithesis_instrumentation__.Notify(84522)
			payload := job.Payload()

			if payload.Type() == jobspb.TypeSchemaChange && func() bool {
				__antithesis_instrumentation__.Notify(84523)
				return !strings.HasPrefix(payload.Description, "ROLL BACK") == true
			}() == true {
				__antithesis_instrumentation__.Notify(84524)
				return job.cancelRequested(ctx, txn, nil)
			} else {
				__antithesis_instrumentation__.Notify(84525)
			}
		} else {
			__antithesis_instrumentation__.Notify(84526)
		}
		__antithesis_instrumentation__.Notify(84521)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84527)
	}
	__antithesis_instrumentation__.Notify(84519)
	return job.cancelRequested(ctx, txn, nil)
}

func (r *Registry) PauseRequested(
	ctx context.Context, txn *kv.Txn, id jobspb.JobID, reason string,
) error {
	__antithesis_instrumentation__.Notify(84528)
	job, resumer, err := r.getJobFn(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(84531)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84532)
	}
	__antithesis_instrumentation__.Notify(84529)
	var onPauseRequested onPauseRequestFunc
	if pr, ok := resumer.(PauseRequester); ok {
		__antithesis_instrumentation__.Notify(84533)
		onPauseRequested = pr.OnPauseRequest
	} else {
		__antithesis_instrumentation__.Notify(84534)
	}
	__antithesis_instrumentation__.Notify(84530)
	return job.PauseRequested(ctx, txn, onPauseRequested, reason)
}

func (r *Registry) Succeeded(ctx context.Context, txn *kv.Txn, id jobspb.JobID) error {
	__antithesis_instrumentation__.Notify(84535)
	job, _, err := r.getJobFn(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(84537)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84538)
	}
	__antithesis_instrumentation__.Notify(84536)
	return job.succeeded(ctx, txn, nil)
}

func (r *Registry) Failed(
	ctx context.Context, txn *kv.Txn, id jobspb.JobID, causingError error,
) error {
	__antithesis_instrumentation__.Notify(84539)
	job, _, err := r.getJobFn(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(84541)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84542)
	}
	__antithesis_instrumentation__.Notify(84540)
	return job.failed(ctx, txn, causingError, nil)
}

func (r *Registry) Unpause(ctx context.Context, txn *kv.Txn, id jobspb.JobID) error {
	__antithesis_instrumentation__.Notify(84543)
	job, _, err := r.getJobFn(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(84545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84546)
	}
	__antithesis_instrumentation__.Notify(84544)
	return job.unpaused(ctx, txn)
}

type Resumer interface {
	Resume(ctx context.Context, execCtx interface{}) error

	OnFailOrCancel(ctx context.Context, execCtx interface{}) error
}

type PauseRequester interface {
	Resumer

	OnPauseRequest(ctx context.Context, execCtx interface{}, txn *kv.Txn, details *jobspb.Progress) error
}

type JobResultsReporter interface {
	ReportResults(ctx context.Context, resultsCh chan<- tree.Datums) error
}

type Constructor func(job *Job, settings *cluster.Settings) Resumer

var constructors = make(map[jobspb.Type]Constructor)

func RegisterConstructor(typ jobspb.Type, fn Constructor) {
	__antithesis_instrumentation__.Notify(84547)
	constructors[typ] = fn
}

func (r *Registry) createResumer(job *Job, settings *cluster.Settings) (Resumer, error) {
	__antithesis_instrumentation__.Notify(84548)
	payload := job.Payload()
	fn := constructors[payload.Type()]
	if fn == nil {
		__antithesis_instrumentation__.Notify(84551)
		return nil, errors.Errorf("no resumer is available for %s", payload.Type())
	} else {
		__antithesis_instrumentation__.Notify(84552)
	}
	__antithesis_instrumentation__.Notify(84549)
	if wrapper := r.TestingResumerCreationKnobs[payload.Type()]; wrapper != nil {
		__antithesis_instrumentation__.Notify(84553)
		return wrapper(fn(job, settings)), nil
	} else {
		__antithesis_instrumentation__.Notify(84554)
	}
	__antithesis_instrumentation__.Notify(84550)
	return fn(job, settings), nil
}

func (r *Registry) stepThroughStateMachine(
	ctx context.Context, execCtx interface{}, resumer Resumer, job *Job, status Status, jobErr error,
) error {
	__antithesis_instrumentation__.Notify(84555)
	payload := job.Payload()
	jobType := payload.Type()
	log.Infof(ctx, "%s job %d: stepping through state %s with error: %+v", jobType, job.ID(), status, jobErr)
	jm := r.metrics.JobMetrics[jobType]
	onExecutionFailed := func(cause error) error {
		__antithesis_instrumentation__.Notify(84557)
		log.InfofDepth(
			ctx, 1,
			"job %d: %s execution encountered retriable error: %+v",
			job.ID(), status, cause,
		)
		start := job.getRunStats().LastRun
		end := r.clock.Now().GoTime()
		return newRetriableExecutionError(
			r.nodeID.SQLInstanceID(), status, start, end, cause,
		)
	}
	__antithesis_instrumentation__.Notify(84556)
	switch status {
	case StatusRunning:
		__antithesis_instrumentation__.Notify(84558)
		if jobErr != nil {
			__antithesis_instrumentation__.Notify(84587)
			return errors.NewAssertionErrorWithWrappedErrf(jobErr,
				"job %d: resuming with non-nil error", job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84588)
		}
		__antithesis_instrumentation__.Notify(84559)
		resumeCtx := logtags.AddTag(ctx, "job", job.ID())

		if err := job.started(ctx, nil); err != nil {
			__antithesis_instrumentation__.Notify(84589)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84590)
		}
		__antithesis_instrumentation__.Notify(84560)

		var err error
		func() {
			__antithesis_instrumentation__.Notify(84591)
			jm.CurrentlyRunning.Inc(1)
			r.metrics.RunningNonIdleJobs.Inc(1)
			defer func() {
				__antithesis_instrumentation__.Notify(84593)
				jm.CurrentlyRunning.Dec(1)
				r.metrics.RunningNonIdleJobs.Dec(1)
			}()
			__antithesis_instrumentation__.Notify(84592)
			err = resumer.Resume(resumeCtx, execCtx)
		}()
		__antithesis_instrumentation__.Notify(84561)

		r.MarkIdle(job, false)

		if err == nil {
			__antithesis_instrumentation__.Notify(84594)
			jm.ResumeCompleted.Inc(1)
			return r.stepThroughStateMachine(ctx, execCtx, resumer, job, StatusSucceeded, nil)
		} else {
			__antithesis_instrumentation__.Notify(84595)
		}
		__antithesis_instrumentation__.Notify(84562)
		if resumeCtx.Err() != nil {
			__antithesis_instrumentation__.Notify(84596)

			jm.ResumeRetryError.Inc(1)
			return errors.Errorf("job %d: node liveness error: restarting in background", job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84597)
		}
		__antithesis_instrumentation__.Notify(84563)

		if errors.Is(err, errPauseSelfSentinel) {
			__antithesis_instrumentation__.Notify(84598)
			if err := r.PauseRequested(ctx, nil, job.ID(), err.Error()); err != nil {
				__antithesis_instrumentation__.Notify(84600)
				return err
			} else {
				__antithesis_instrumentation__.Notify(84601)
			}
			__antithesis_instrumentation__.Notify(84599)
			return errors.Wrap(err, PauseRequestExplained)
		} else {
			__antithesis_instrumentation__.Notify(84602)
		}
		__antithesis_instrumentation__.Notify(84564)

		if nonCancelableRetry := job.Payload().Noncancelable && func() bool {
			__antithesis_instrumentation__.Notify(84603)
			return !IsPermanentJobError(err) == true
		}() == true; nonCancelableRetry || func() bool {
			__antithesis_instrumentation__.Notify(84604)
			return errors.Is(err, errRetryJobSentinel) == true
		}() == true {
			__antithesis_instrumentation__.Notify(84605)
			jm.ResumeRetryError.Inc(1)
			if nonCancelableRetry {
				__antithesis_instrumentation__.Notify(84607)
				err = errors.Wrapf(err, "non-cancelable")
			} else {
				__antithesis_instrumentation__.Notify(84608)
			}
			__antithesis_instrumentation__.Notify(84606)
			return onExecutionFailed(err)
		} else {
			__antithesis_instrumentation__.Notify(84609)
		}
		__antithesis_instrumentation__.Notify(84565)

		jm.ResumeFailed.Inc(1)
		if sErr := (*InvalidStatusError)(nil); errors.As(err, &sErr) {
			__antithesis_instrumentation__.Notify(84610)
			if sErr.status != StatusCancelRequested && func() bool {
				__antithesis_instrumentation__.Notify(84612)
				return sErr.status != StatusPauseRequested == true
			}() == true {
				__antithesis_instrumentation__.Notify(84613)
				return errors.NewAssertionErrorWithWrappedErrf(sErr,
					"job %d: unexpected status %s provided for a running job", job.ID(), sErr.status)
			} else {
				__antithesis_instrumentation__.Notify(84614)
			}
			__antithesis_instrumentation__.Notify(84611)
			return sErr
		} else {
			__antithesis_instrumentation__.Notify(84615)
		}
		__antithesis_instrumentation__.Notify(84566)
		return r.stepThroughStateMachine(ctx, execCtx, resumer, job, StatusReverting, err)
	case StatusPauseRequested:
		__antithesis_instrumentation__.Notify(84567)
		return errors.Errorf("job %s", status)
	case StatusCancelRequested:
		__antithesis_instrumentation__.Notify(84568)
		return errors.Errorf("job %s", status)
	case StatusPaused:
		__antithesis_instrumentation__.Notify(84569)
		return errors.NewAssertionErrorWithWrappedErrf(jobErr,
			"job %d: unexpected status %s provided to state machine", job.ID(), status)
	case StatusCanceled:
		__antithesis_instrumentation__.Notify(84570)
		if err := job.canceled(ctx, nil, nil); err != nil {
			__antithesis_instrumentation__.Notify(84616)

			return errors.WithSecondaryError(
				errors.Wrapf(err, "job %d: could not mark as canceled", job.ID()),
				jobErr,
			)
		} else {
			__antithesis_instrumentation__.Notify(84617)
		}
		__antithesis_instrumentation__.Notify(84571)
		telemetry.Inc(TelemetryMetrics[jobType].Canceled)
		r.removeFromWaitingSets(job.ID())
		return errors.WithSecondaryError(errors.Errorf("job %s", status), jobErr)
	case StatusSucceeded:
		__antithesis_instrumentation__.Notify(84572)
		if jobErr != nil {
			__antithesis_instrumentation__.Notify(84618)
			return errors.NewAssertionErrorWithWrappedErrf(jobErr,
				"job %d: successful but unexpected error provided", job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84619)
		}
		__antithesis_instrumentation__.Notify(84573)
		err := job.succeeded(ctx, nil, nil)
		switch {
		case err == nil:
			__antithesis_instrumentation__.Notify(84620)
			telemetry.Inc(TelemetryMetrics[jobType].Successful)
			r.removeFromWaitingSets(job.ID())
		default:
			__antithesis_instrumentation__.Notify(84621)

			err = errors.Wrapf(err, "job %d: could not mark as succeeded", job.ID())
		}
		__antithesis_instrumentation__.Notify(84574)
		return err
	case StatusReverting:
		__antithesis_instrumentation__.Notify(84575)
		if err := job.reverted(ctx, nil, jobErr, nil); err != nil {
			__antithesis_instrumentation__.Notify(84622)

			return errors.WithSecondaryError(
				errors.Wrapf(err, "job %d: could not mark as reverting", job.ID()),
				jobErr,
			)
		} else {
			__antithesis_instrumentation__.Notify(84623)
		}
		__antithesis_instrumentation__.Notify(84576)
		onFailOrCancelCtx := logtags.AddTag(ctx, "job", job.ID())
		var err error
		func() {
			__antithesis_instrumentation__.Notify(84624)
			jm.CurrentlyRunning.Inc(1)
			r.metrics.RunningNonIdleJobs.Inc(1)
			defer func() {
				__antithesis_instrumentation__.Notify(84626)
				jm.CurrentlyRunning.Dec(1)
				r.metrics.RunningNonIdleJobs.Dec(1)
			}()
			__antithesis_instrumentation__.Notify(84625)
			err = resumer.OnFailOrCancel(onFailOrCancelCtx, execCtx)
		}()
		__antithesis_instrumentation__.Notify(84577)
		if successOnFailOrCancel := err == nil; successOnFailOrCancel {
			__antithesis_instrumentation__.Notify(84627)
			jm.FailOrCancelCompleted.Inc(1)

			nextStatus := StatusFailed
			if HasErrJobCanceled(jobErr) {
				__antithesis_instrumentation__.Notify(84629)
				nextStatus = StatusCanceled
			} else {
				__antithesis_instrumentation__.Notify(84630)
			}
			__antithesis_instrumentation__.Notify(84628)
			return r.stepThroughStateMachine(ctx, execCtx, resumer, job, nextStatus, jobErr)
		} else {
			__antithesis_instrumentation__.Notify(84631)
		}
		__antithesis_instrumentation__.Notify(84578)
		jm.FailOrCancelRetryError.Inc(1)
		if onFailOrCancelCtx.Err() != nil {
			__antithesis_instrumentation__.Notify(84632)

			return errors.Errorf("job %d: node liveness error: restarting in background", job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84633)
		}
		__antithesis_instrumentation__.Notify(84579)
		return onExecutionFailed(err)
	case StatusFailed:
		__antithesis_instrumentation__.Notify(84580)
		if jobErr == nil {
			__antithesis_instrumentation__.Notify(84634)
			return errors.AssertionFailedf("job %d: has StatusFailed but no error was provided", job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84635)
		}
		__antithesis_instrumentation__.Notify(84581)
		if err := job.failed(ctx, nil, jobErr, nil); err != nil {
			__antithesis_instrumentation__.Notify(84636)

			return errors.WithSecondaryError(
				errors.Wrapf(err, "job %d: could not mark as failed", job.ID()),
				jobErr,
			)
		} else {
			__antithesis_instrumentation__.Notify(84637)
		}
		__antithesis_instrumentation__.Notify(84582)
		telemetry.Inc(TelemetryMetrics[jobType].Failed)
		r.removeFromWaitingSets(job.ID())
		return jobErr
	case StatusRevertFailed:
		__antithesis_instrumentation__.Notify(84583)

		if jobErr == nil {
			__antithesis_instrumentation__.Notify(84638)
			return errors.AssertionFailedf("job %d: has StatusRevertFailed but no error was provided",
				job.ID())
		} else {
			__antithesis_instrumentation__.Notify(84639)
		}
		__antithesis_instrumentation__.Notify(84584)
		if err := job.revertFailed(ctx, nil, jobErr, nil); err != nil {
			__antithesis_instrumentation__.Notify(84640)

			return errors.WithSecondaryError(
				errors.Wrapf(err, "job %d: could not mark as revert field", job.ID()),
				jobErr,
			)
		} else {
			__antithesis_instrumentation__.Notify(84641)
		}
		__antithesis_instrumentation__.Notify(84585)
		return jobErr
	default:
		__antithesis_instrumentation__.Notify(84586)
		return errors.NewAssertionErrorWithWrappedErrf(jobErr,
			"job %d: has unsupported status %s", job.ID(), status)
	}
}

func (r *Registry) adoptionDisabled(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(84642)
	if r.knobs.DisableAdoptions {
		__antithesis_instrumentation__.Notify(84645)
		return true
	} else {
		__antithesis_instrumentation__.Notify(84646)
	}
	__antithesis_instrumentation__.Notify(84643)
	if r.preventAdoptionFile != "" {
		__antithesis_instrumentation__.Notify(84647)
		if _, err := os.Stat(r.preventAdoptionFile); err != nil {
			__antithesis_instrumentation__.Notify(84649)
			if !oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(84651)
				log.Warningf(ctx, "error checking if job adoption is currently disabled: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(84652)
			}
			__antithesis_instrumentation__.Notify(84650)
			return false
		} else {
			__antithesis_instrumentation__.Notify(84653)
		}
		__antithesis_instrumentation__.Notify(84648)
		log.Warningf(ctx, "job adoption is currently disabled by existence of %s", r.preventAdoptionFile)
		return true
	} else {
		__antithesis_instrumentation__.Notify(84654)
	}
	__antithesis_instrumentation__.Notify(84644)
	return false
}

func (r *Registry) MarkIdle(job *Job, isIdle bool) {
	__antithesis_instrumentation__.Notify(84655)
	r.mu.Lock()
	defer r.mu.Unlock()
	if aj, ok := r.mu.adoptedJobs[job.ID()]; ok {
		__antithesis_instrumentation__.Notify(84656)
		payload := job.Payload()
		jobType := payload.Type()
		jm := r.metrics.JobMetrics[jobType]
		if aj.isIdle != isIdle {
			__antithesis_instrumentation__.Notify(84657)
			log.Infof(r.serverCtx, "%s job %d: toggling idleness to %+v", jobType, job.ID(), isIdle)
			if isIdle {
				__antithesis_instrumentation__.Notify(84659)
				r.metrics.RunningNonIdleJobs.Dec(1)
				jm.CurrentlyIdle.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(84660)
				r.metrics.RunningNonIdleJobs.Inc(1)
				jm.CurrentlyIdle.Dec(1)
			}
			__antithesis_instrumentation__.Notify(84658)
			aj.isIdle = isIdle
		} else {
			__antithesis_instrumentation__.Notify(84661)
		}
	} else {
		__antithesis_instrumentation__.Notify(84662)
	}
}

func (r *Registry) cancelAllAdoptedJobs() {
	__antithesis_instrumentation__.Notify(84663)
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, aj := range r.mu.adoptedJobs {
		__antithesis_instrumentation__.Notify(84665)
		aj.cancel()
	}
	__antithesis_instrumentation__.Notify(84664)
	r.mu.adoptedJobs = make(map[jobspb.JobID]*adoptedJob)
}

func (r *Registry) unregister(jobID jobspb.JobID) {
	__antithesis_instrumentation__.Notify(84666)
	r.mu.Lock()
	defer r.mu.Unlock()
	if aj, ok := r.mu.adoptedJobs[jobID]; ok {
		__antithesis_instrumentation__.Notify(84667)
		aj.cancel()
		delete(r.mu.adoptedJobs, jobID)
	} else {
		__antithesis_instrumentation__.Notify(84668)
	}
}

func (r *Registry) cancelRegisteredJobContext(jobID jobspb.JobID) {
	__antithesis_instrumentation__.Notify(84669)
	r.mu.Lock()
	defer r.mu.Unlock()
	if aj, ok := r.mu.adoptedJobs[jobID]; ok {
		__antithesis_instrumentation__.Notify(84670)
		aj.cancel()
	} else {
		__antithesis_instrumentation__.Notify(84671)
	}
}

func (r *Registry) getClaimedJob(jobID jobspb.JobID) (*Job, error) {
	__antithesis_instrumentation__.Notify(84672)
	r.mu.Lock()
	defer r.mu.Unlock()

	aj, ok := r.mu.adoptedJobs[jobID]
	if !ok {
		__antithesis_instrumentation__.Notify(84674)
		return nil, &JobNotFoundError{jobID: jobID}
	} else {
		__antithesis_instrumentation__.Notify(84675)
	}
	__antithesis_instrumentation__.Notify(84673)
	return &Job{
		id:       jobID,
		session:  aj.session,
		registry: r,
	}, nil
}

func (r *Registry) RetryInitialDelay() float64 {
	__antithesis_instrumentation__.Notify(84676)
	if r.knobs.IntervalOverrides.RetryInitialDelay != nil {
		__antithesis_instrumentation__.Notify(84678)
		return r.knobs.IntervalOverrides.RetryInitialDelay.Seconds()
	} else {
		__antithesis_instrumentation__.Notify(84679)
	}
	__antithesis_instrumentation__.Notify(84677)
	return retryInitialDelaySetting.Get(&r.settings.SV).Seconds()
}

func (r *Registry) RetryMaxDelay() float64 {
	__antithesis_instrumentation__.Notify(84680)
	if r.knobs.IntervalOverrides.RetryMaxDelay != nil {
		__antithesis_instrumentation__.Notify(84682)
		return r.knobs.IntervalOverrides.RetryMaxDelay.Seconds()
	} else {
		__antithesis_instrumentation__.Notify(84683)
	}
	__antithesis_instrumentation__.Notify(84681)
	return retryMaxDelaySetting.Get(&r.settings.SV).Seconds()
}

func (r *Registry) maybeRecordExecutionFailure(ctx context.Context, err error, j *Job) {
	__antithesis_instrumentation__.Notify(84684)
	var efe *retriableExecutionError
	if !errors.As(err, &efe) {
		__antithesis_instrumentation__.Notify(84688)
		return
	} else {
		__antithesis_instrumentation__.Notify(84689)
	}
	__antithesis_instrumentation__.Notify(84685)

	updateErr := j.Update(ctx, nil, func(
		txn *kv.Txn, md JobMetadata, ju *JobUpdater,
	) error {
		__antithesis_instrumentation__.Notify(84690)
		pl := md.Payload
		{
			__antithesis_instrumentation__.Notify(84692)
			maxSize := int(executionErrorsMaxEntrySize.Get(&r.settings.SV))
			pl.RetriableExecutionFailureLog = append(pl.RetriableExecutionFailureLog,
				efe.toRetriableExecutionFailure(ctx, maxSize))
		}
		{
			__antithesis_instrumentation__.Notify(84693)
			maxEntries := int(executionErrorsMaxEntriesSetting.Get(&r.settings.SV))
			log := &pl.RetriableExecutionFailureLog
			if len(*log) > maxEntries {
				__antithesis_instrumentation__.Notify(84694)
				*log = (*log)[len(*log)-maxEntries:]
			} else {
				__antithesis_instrumentation__.Notify(84695)
			}
		}
		__antithesis_instrumentation__.Notify(84691)
		ju.UpdatePayload(pl)
		return nil
	})
	__antithesis_instrumentation__.Notify(84686)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(84696)
		return
	} else {
		__antithesis_instrumentation__.Notify(84697)
	}
	__antithesis_instrumentation__.Notify(84687)
	if updateErr != nil {
		__antithesis_instrumentation__.Notify(84698)
		log.Warningf(ctx, "failed to record error for job %d: %v: %v", j.ID(), err, err)
	} else {
		__antithesis_instrumentation__.Notify(84699)
	}
}

func (r *Registry) CheckPausepoint(name string) error {
	__antithesis_instrumentation__.Notify(84700)
	s := debugPausepoints.Get(&r.settings.SV)
	if s == "" {
		__antithesis_instrumentation__.Notify(84703)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(84704)
	}
	__antithesis_instrumentation__.Notify(84701)
	for _, point := range strings.Split(s, ",") {
		__antithesis_instrumentation__.Notify(84705)
		if name == point {
			__antithesis_instrumentation__.Notify(84706)
			return MarkPauseRequestError(errors.Newf("pause point %q hit", name))
		} else {
			__antithesis_instrumentation__.Notify(84707)
		}
	}
	__antithesis_instrumentation__.Notify(84702)
	return nil
}

func (r *Registry) TestingIsJobIdle(jobID jobspb.JobID) bool {
	__antithesis_instrumentation__.Notify(84708)
	r.mu.Lock()
	defer r.mu.Unlock()
	adoptedJob := r.mu.adoptedJobs[jobID]
	return adoptedJob != nil && func() bool {
		__antithesis_instrumentation__.Notify(84709)
		return adoptedJob.isIdle == true
	}() == true
}
