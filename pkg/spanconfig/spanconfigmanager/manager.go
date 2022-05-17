package spanconfigmanager

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var checkReconciliationJobInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"spanconfig.reconciliation_job.check_interval",
	"the frequency at which to check if the span config reconciliation job exists (and to start it if not)",
	10*time.Minute,
	settings.NonNegativeDuration,
)

var jobEnabledSetting = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"spanconfig.reconciliation_job.enabled",
	"enable the use of the kv accessor", true)

type Manager struct {
	db       *kv.DB
	jr       *jobs.Registry
	ie       sqlutil.InternalExecutor
	stopper  *stop.Stopper
	settings *cluster.Settings
	knobs    *spanconfig.TestingKnobs

	spanconfig.Reconciler
}

func New(
	db *kv.DB,
	jr *jobs.Registry,
	ie sqlutil.InternalExecutor,
	stopper *stop.Stopper,
	settings *cluster.Settings,
	reconciler spanconfig.Reconciler,
	knobs *spanconfig.TestingKnobs,
) *Manager {
	__antithesis_instrumentation__.Notify(240707)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240709)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240710)
	}
	__antithesis_instrumentation__.Notify(240708)
	return &Manager{
		db:         db,
		jr:         jr,
		ie:         ie,
		stopper:    stopper,
		settings:   settings,
		Reconciler: reconciler,
		knobs:      knobs,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(240711)
	return m.stopper.RunAsyncTask(ctx, "span-config-mgr", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(240712)
		m.run(ctx)
	})
}

func (m *Manager) run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(240713)
	jobCheckCh := make(chan struct{}, 1)
	triggerJobCheck := func() {
		__antithesis_instrumentation__.Notify(240719)
		select {
		case jobCheckCh <- struct{}{}:
			__antithesis_instrumentation__.Notify(240720)
		default:
			__antithesis_instrumentation__.Notify(240721)
		}
	}
	__antithesis_instrumentation__.Notify(240714)

	jobEnabledSetting.SetOnChange(&m.settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(240722)
		triggerJobCheck()
	})
	__antithesis_instrumentation__.Notify(240715)
	checkReconciliationJobInterval.SetOnChange(&m.settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(240723)
		triggerJobCheck()
	})
	__antithesis_instrumentation__.Notify(240716)
	m.settings.Version.SetOnChange(func(_ context.Context, _ clusterversion.ClusterVersion) {
		__antithesis_instrumentation__.Notify(240724)
		triggerJobCheck()
	})
	__antithesis_instrumentation__.Notify(240717)

	checkJob := func() {
		__antithesis_instrumentation__.Notify(240725)
		if fn := m.knobs.ManagerCheckJobInterceptor; fn != nil {
			__antithesis_instrumentation__.Notify(240729)
			fn()
		} else {
			__antithesis_instrumentation__.Notify(240730)
		}
		__antithesis_instrumentation__.Notify(240726)

		if !jobEnabledSetting.Get(&m.settings.SV) {
			__antithesis_instrumentation__.Notify(240731)
			return
		} else {
			__antithesis_instrumentation__.Notify(240732)
		}
		__antithesis_instrumentation__.Notify(240727)

		started, err := m.createAndStartJobIfNoneExists(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(240733)
			log.Errorf(ctx, "error starting auto span config reconciliation job: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(240734)
		}
		__antithesis_instrumentation__.Notify(240728)
		if started {
			__antithesis_instrumentation__.Notify(240735)
			log.Infof(ctx, "started auto span config reconciliation job")
		} else {
			__antithesis_instrumentation__.Notify(240736)
		}
	}
	__antithesis_instrumentation__.Notify(240718)

	timer := timeutil.NewTimer()
	defer timer.Stop()

	triggerJobCheck()
	for {
		__antithesis_instrumentation__.Notify(240737)
		timer.Reset(checkReconciliationJobInterval.Get(&m.settings.SV))
		select {
		case <-jobCheckCh:
			__antithesis_instrumentation__.Notify(240738)
			checkJob()
		case <-timer.C:
			__antithesis_instrumentation__.Notify(240739)
			timer.Read = true
			checkJob()
		case <-m.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(240740)
			return
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(240741)
			return
		}
	}
}

func (m *Manager) createAndStartJobIfNoneExists(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(240742)
	if m.knobs.ManagerDisableJobCreation {
		__antithesis_instrumentation__.Notify(240747)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(240748)
	}
	__antithesis_instrumentation__.Notify(240743)
	record := jobs.Record{
		JobID:         m.jr.MakeJobID(),
		Description:   "reconciling span configurations",
		Username:      security.NodeUserName(),
		Details:       jobspb.AutoSpanConfigReconciliationDetails{},
		Progress:      jobspb.AutoSpanConfigReconciliationProgress{},
		NonCancelable: true,
	}

	var job *jobs.Job
	if err := m.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(240749)
		exists, err := jobs.RunningJobExists(ctx, jobspb.InvalidJobID, m.ie, txn,
			func(payload *jobspb.Payload) bool {
				__antithesis_instrumentation__.Notify(240754)
				return payload.Type() == jobspb.TypeAutoSpanConfigReconciliation
			},
		)
		__antithesis_instrumentation__.Notify(240750)
		if err != nil {
			__antithesis_instrumentation__.Notify(240755)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240756)
		}
		__antithesis_instrumentation__.Notify(240751)

		if fn := m.knobs.ManagerAfterCheckedReconciliationJobExistsInterceptor; fn != nil {
			__antithesis_instrumentation__.Notify(240757)
			fn(exists)
		} else {
			__antithesis_instrumentation__.Notify(240758)
		}
		__antithesis_instrumentation__.Notify(240752)

		if exists {
			__antithesis_instrumentation__.Notify(240759)

			job = nil
			return nil
		} else {
			__antithesis_instrumentation__.Notify(240760)
		}
		__antithesis_instrumentation__.Notify(240753)
		job, err = m.jr.CreateJobWithTxn(ctx, record, record.JobID, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(240761)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(240762)
	}
	__antithesis_instrumentation__.Notify(240744)

	if job == nil {
		__antithesis_instrumentation__.Notify(240763)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(240764)
	}
	__antithesis_instrumentation__.Notify(240745)

	if fn := m.knobs.ManagerCreatedJobInterceptor; fn != nil {
		__antithesis_instrumentation__.Notify(240765)
		fn(job)
	} else {
		__antithesis_instrumentation__.Notify(240766)
	}
	__antithesis_instrumentation__.Notify(240746)
	m.jr.NotifyToResume(ctx, job.ID())
	return true, nil
}
