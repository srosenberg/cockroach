package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func NewJobRunDependencies(
	collectionFactory *descs.CollectionFactory,
	db *kv.DB,
	internalExecutor sqlutil.InternalExecutor,
	backfiller scexec.Backfiller,
	rangeCounter RangeCounter,
	eventLoggerFactory EventLoggerFactory,
	jobRegistry *jobs.Registry,
	job *jobs.Job,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	indexValidator scexec.IndexValidator,
	commentUpdaterFactory scexec.DescriptorMetadataUpdaterFactory,
	testingKnobs *scrun.TestingKnobs,
	statements []string,
	sessionData *sessiondata.SessionData,
	kvTrace bool,
) scrun.JobRunDependencies {
	__antithesis_instrumentation__.Notify(580753)
	return &jobExecutionDeps{
		collectionFactory:     collectionFactory,
		db:                    db,
		internalExecutor:      internalExecutor,
		backfiller:            backfiller,
		rangeCounter:          rangeCounter,
		eventLoggerFactory:    eventLoggerFactory,
		jobRegistry:           jobRegistry,
		job:                   job,
		codec:                 codec,
		settings:              settings,
		testingKnobs:          testingKnobs,
		statements:            statements,
		indexValidator:        indexValidator,
		commentUpdaterFactory: commentUpdaterFactory,
		sessionData:           sessionData,
		kvTrace:               kvTrace,
	}
}

type jobExecutionDeps struct {
	collectionFactory     *descs.CollectionFactory
	db                    *kv.DB
	internalExecutor      sqlutil.InternalExecutor
	eventLoggerFactory    func(txn *kv.Txn) scexec.EventLogger
	backfiller            scexec.Backfiller
	commentUpdaterFactory scexec.DescriptorMetadataUpdaterFactory
	rangeCounter          RangeCounter
	jobRegistry           *jobs.Registry
	job                   *jobs.Job
	kvTrace               bool

	indexValidator scexec.IndexValidator

	codec        keys.SQLCodec
	settings     *cluster.Settings
	testingKnobs *scrun.TestingKnobs
	statements   []string
	sessionData  *sessiondata.SessionData
}

var _ scrun.JobRunDependencies = (*jobExecutionDeps)(nil)

func (d *jobExecutionDeps) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(580754)
	return d.settings
}

func (d *jobExecutionDeps) WithTxnInJob(ctx context.Context, fn scrun.JobTxnFunc) error {
	__antithesis_instrumentation__.Notify(580755)
	var createdJobs []jobspb.JobID
	err := d.collectionFactory.Txn(ctx, d.internalExecutor, d.db, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(580759)
		pl := d.job.Payload()
		ed := &execDeps{
			txnDeps: txnDeps{
				txn:                txn,
				codec:              d.codec,
				descsCollection:    descriptors,
				jobRegistry:        d.jobRegistry,
				indexValidator:     d.indexValidator,
				eventLogger:        d.eventLoggerFactory(txn),
				schemaChangerJobID: d.job.ID(),
				kvTrace:            d.kvTrace,
			},
			backfiller: d.backfiller,
			backfillTracker: newBackfillTracker(d.codec,
				newBackfillTrackerConfig(ctx, d.codec, d.db, d.rangeCounter, d.job),
				convertFromJobBackfillProgress(
					d.codec, pl.GetNewSchemaChange().BackfillProgress,
				),
			),
			periodicProgressFlusher: newPeriodicProgressFlusherForIndexBackfill(d.settings),
			statements:              d.statements,
			user:                    pl.UsernameProto.Decode(),
			clock:                   NewConstantClock(timeutil.FromUnixMicros(pl.StartedMicros)),
			commentUpdaterFactory:   d.commentUpdaterFactory,
			sessionData:             d.sessionData,
		}
		if err := fn(ctx, ed); err != nil {
			__antithesis_instrumentation__.Notify(580761)
			return err
		} else {
			__antithesis_instrumentation__.Notify(580762)
		}
		__antithesis_instrumentation__.Notify(580760)
		createdJobs = ed.CreatedJobs()
		return nil
	})
	__antithesis_instrumentation__.Notify(580756)
	if err != nil {
		__antithesis_instrumentation__.Notify(580763)
		return err
	} else {
		__antithesis_instrumentation__.Notify(580764)
	}
	__antithesis_instrumentation__.Notify(580757)
	if len(createdJobs) > 0 {
		__antithesis_instrumentation__.Notify(580765)
		d.jobRegistry.NotifyToResume(ctx, createdJobs...)
	} else {
		__antithesis_instrumentation__.Notify(580766)
	}
	__antithesis_instrumentation__.Notify(580758)
	return nil
}
