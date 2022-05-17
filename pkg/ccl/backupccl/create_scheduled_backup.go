package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/jsonpb"
	pbtypes "github.com/gogo/protobuf/types"
	cron "github.com/robfig/cron/v3"
)

const (
	optFirstRun                = "first_run"
	optOnExecFailure           = "on_execution_failure"
	optOnPreviousRunning       = "on_previous_running"
	optIgnoreExistingBackups   = "ignore_existing_backups"
	optUpdatesLastBackupMetric = "updates_cluster_last_backup_time_metric"
)

var scheduledBackupOptionExpectValues = map[string]sql.KVStringOptValidate{
	optFirstRun:                sql.KVStringOptRequireValue,
	optOnExecFailure:           sql.KVStringOptRequireValue,
	optOnPreviousRunning:       sql.KVStringOptRequireValue,
	optIgnoreExistingBackups:   sql.KVStringOptRequireNoValue,
	optUpdatesLastBackupMetric: sql.KVStringOptRequireNoValue,
}

var scheduledBackupGCProtectionEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"schedules.backup.gc_protection.enabled",
	"enable chaining of GC protection across backups run as part of a schedule; default is false",
	false,
).WithPublic()

type scheduledBackupEval struct {
	*tree.ScheduledBackup

	isEnterpriseUser bool

	scheduleLabel        func() (string, error)
	recurrence           func() (string, error)
	fullBackupRecurrence func() (string, error)
	scheduleOpts         func() (map[string]string, error)

	destination          func() ([]string, error)
	encryptionPassphrase func() (string, error)
	kmsURIs              func() ([]string, error)
	incrementalStorage   func() ([]string, error)
}

func parseOnError(onError string, details *jobspb.ScheduleDetails) error {
	__antithesis_instrumentation__.Notify(9276)
	switch strings.ToLower(onError) {
	case "retry":
		__antithesis_instrumentation__.Notify(9278)
		details.OnError = jobspb.ScheduleDetails_RETRY_SOON
	case "reschedule":
		__antithesis_instrumentation__.Notify(9279)
		details.OnError = jobspb.ScheduleDetails_RETRY_SCHED
	case "pause":
		__antithesis_instrumentation__.Notify(9280)
		details.OnError = jobspb.ScheduleDetails_PAUSE_SCHED
	default:
		__antithesis_instrumentation__.Notify(9281)
		return errors.Newf(
			"%q is not a valid on_execution_error; valid values are [retry|reschedule|pause]",
			onError)
	}
	__antithesis_instrumentation__.Notify(9277)
	return nil
}

func parseWaitBehavior(wait string, details *jobspb.ScheduleDetails) error {
	__antithesis_instrumentation__.Notify(9282)
	switch strings.ToLower(wait) {
	case "start":
		__antithesis_instrumentation__.Notify(9284)
		details.Wait = jobspb.ScheduleDetails_NO_WAIT
	case "skip":
		__antithesis_instrumentation__.Notify(9285)
		details.Wait = jobspb.ScheduleDetails_SKIP
	case "wait":
		__antithesis_instrumentation__.Notify(9286)
		details.Wait = jobspb.ScheduleDetails_WAIT
	default:
		__antithesis_instrumentation__.Notify(9287)
		return errors.Newf(
			"%q is not a valid on_previous_running; valid values are [start|skip|wait]",
			wait)
	}
	__antithesis_instrumentation__.Notify(9283)
	return nil
}

func parseOnPreviousRunningOption(
	onPreviousRunning jobspb.ScheduleDetails_WaitBehavior,
) (string, error) {
	__antithesis_instrumentation__.Notify(9288)
	var onPreviousRunningOption string
	switch onPreviousRunning {
	case jobspb.ScheduleDetails_WAIT:
		__antithesis_instrumentation__.Notify(9290)
		onPreviousRunningOption = "WAIT"
	case jobspb.ScheduleDetails_NO_WAIT:
		__antithesis_instrumentation__.Notify(9291)
		onPreviousRunningOption = "START"
	case jobspb.ScheduleDetails_SKIP:
		__antithesis_instrumentation__.Notify(9292)
		onPreviousRunningOption = "SKIP"
	default:
		__antithesis_instrumentation__.Notify(9293)
		return onPreviousRunningOption, errors.Newf("%s is an invalid onPreviousRunning option", onPreviousRunning.String())
	}
	__antithesis_instrumentation__.Notify(9289)
	return onPreviousRunningOption, nil
}

func parseOnErrorOption(onError jobspb.ScheduleDetails_ErrorHandlingBehavior) (string, error) {
	__antithesis_instrumentation__.Notify(9294)
	var onErrorOption string
	switch onError {
	case jobspb.ScheduleDetails_RETRY_SCHED:
		__antithesis_instrumentation__.Notify(9296)
		onErrorOption = "RESCHEDULE"
	case jobspb.ScheduleDetails_RETRY_SOON:
		__antithesis_instrumentation__.Notify(9297)
		onErrorOption = "RETRY"
	case jobspb.ScheduleDetails_PAUSE_SCHED:
		__antithesis_instrumentation__.Notify(9298)
		onErrorOption = "PAUSE"
	default:
		__antithesis_instrumentation__.Notify(9299)
		return onErrorOption, errors.Newf("%s is an invalid onError option", onError.String())
	}
	__antithesis_instrumentation__.Notify(9295)
	return onErrorOption, nil
}

func makeScheduleDetails(opts map[string]string) (jobspb.ScheduleDetails, error) {
	__antithesis_instrumentation__.Notify(9300)
	var details jobspb.ScheduleDetails
	if v, ok := opts[optOnExecFailure]; ok {
		__antithesis_instrumentation__.Notify(9303)
		if err := parseOnError(v, &details); err != nil {
			__antithesis_instrumentation__.Notify(9304)
			return details, err
		} else {
			__antithesis_instrumentation__.Notify(9305)
		}
	} else {
		__antithesis_instrumentation__.Notify(9306)
	}
	__antithesis_instrumentation__.Notify(9301)

	if v, ok := opts[optOnPreviousRunning]; ok {
		__antithesis_instrumentation__.Notify(9307)
		if err := parseWaitBehavior(v, &details); err != nil {
			__antithesis_instrumentation__.Notify(9308)
			return details, err
		} else {
			__antithesis_instrumentation__.Notify(9309)
		}
	} else {
		__antithesis_instrumentation__.Notify(9310)
	}
	__antithesis_instrumentation__.Notify(9302)
	return details, nil
}

func scheduleFirstRun(evalCtx *tree.EvalContext, opts map[string]string) (*time.Time, error) {
	__antithesis_instrumentation__.Notify(9311)
	if v, ok := opts[optFirstRun]; ok {
		__antithesis_instrumentation__.Notify(9313)
		firstRun, _, err := tree.ParseDTimestampTZ(evalCtx, v, time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(9315)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9316)
		}
		__antithesis_instrumentation__.Notify(9314)
		return &firstRun.Time, nil
	} else {
		__antithesis_instrumentation__.Notify(9317)
	}
	__antithesis_instrumentation__.Notify(9312)
	return nil, nil
}

type scheduleRecurrence struct {
	cron      string
	frequency time.Duration
}

var neverRecurs *scheduleRecurrence

func computeScheduleRecurrence(
	now time.Time, evalFn func() (string, error),
) (*scheduleRecurrence, error) {
	__antithesis_instrumentation__.Notify(9318)
	if evalFn == nil {
		__antithesis_instrumentation__.Notify(9322)
		return neverRecurs, nil
	} else {
		__antithesis_instrumentation__.Notify(9323)
	}
	__antithesis_instrumentation__.Notify(9319)
	cronStr, err := evalFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(9324)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9325)
	}
	__antithesis_instrumentation__.Notify(9320)
	expr, err := cron.ParseStandard(cronStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(9326)
		return nil, errors.Newf(
			`error parsing schedule expression: %q; it must be a valid cron expression`,
			cronStr)
	} else {
		__antithesis_instrumentation__.Notify(9327)
	}
	__antithesis_instrumentation__.Notify(9321)
	nextRun := expr.Next(now)
	frequency := expr.Next(nextRun).Sub(nextRun)
	return &scheduleRecurrence{cronStr, frequency}, nil
}

var forceFullBackup *scheduleRecurrence

func pickFullRecurrenceFromIncremental(inc *scheduleRecurrence) *scheduleRecurrence {
	__antithesis_instrumentation__.Notify(9328)
	if inc.frequency <= time.Hour {
		__antithesis_instrumentation__.Notify(9331)

		return &scheduleRecurrence{
			cron:      "@daily",
			frequency: 24 * time.Hour,
		}
	} else {
		__antithesis_instrumentation__.Notify(9332)
	}
	__antithesis_instrumentation__.Notify(9329)

	if inc.frequency <= 24*time.Hour {
		__antithesis_instrumentation__.Notify(9333)

		return &scheduleRecurrence{
			cron:      "@weekly",
			frequency: 7 * 24 * time.Hour,
		}
	} else {
		__antithesis_instrumentation__.Notify(9334)
	}
	__antithesis_instrumentation__.Notify(9330)

	return forceFullBackup
}

const scheduleBackupOp = "CREATE SCHEDULE FOR BACKUP"

func canChainProtectedTimestampRecords(p sql.PlanHookState, eval *scheduledBackupEval) bool {
	__antithesis_instrumentation__.Notify(9335)
	if !scheduledBackupGCProtectionEnabled.Get(&p.ExecCfg().Settings.SV) || func() bool {
		__antithesis_instrumentation__.Notify(9339)
		return !eval.BackupOptions.CaptureRevisionHistory == true
	}() == true {
		__antithesis_instrumentation__.Notify(9340)
		return false
	} else {
		__antithesis_instrumentation__.Notify(9341)
	}
	__antithesis_instrumentation__.Notify(9336)

	if eval.Coverage() == tree.AllDescriptors {
		__antithesis_instrumentation__.Notify(9342)
		return true
	} else {
		__antithesis_instrumentation__.Notify(9343)
	}
	__antithesis_instrumentation__.Notify(9337)

	for _, t := range eval.Targets.Tables {
		__antithesis_instrumentation__.Notify(9344)
		pattern, err := t.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(9346)
			return false
		} else {
			__antithesis_instrumentation__.Notify(9347)
		}
		__antithesis_instrumentation__.Notify(9345)
		if _, ok := pattern.(*tree.AllTablesSelector); ok {
			__antithesis_instrumentation__.Notify(9348)
			return false
		} else {
			__antithesis_instrumentation__.Notify(9349)
		}
	}
	__antithesis_instrumentation__.Notify(9338)

	return eval.Targets.Tables != nil || func() bool {
		__antithesis_instrumentation__.Notify(9350)
		return eval.Targets.TenantID.IsSet() == true
	}() == true
}

func doCreateBackupSchedules(
	ctx context.Context, p sql.PlanHookState, eval *scheduledBackupEval, resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(9351)
	if err := p.RequireAdminRole(ctx, scheduleBackupOp); err != nil {
		__antithesis_instrumentation__.Notify(9374)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9375)
	}
	__antithesis_instrumentation__.Notify(9352)

	if eval.ScheduleLabelSpec.IfNotExists {
		__antithesis_instrumentation__.Notify(9376)
		scheduleLabel, err := eval.scheduleLabel()
		if err != nil {
			__antithesis_instrumentation__.Notify(9379)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9380)
		}
		__antithesis_instrumentation__.Notify(9377)

		exists, err := checkScheduleAlreadyExists(ctx, p, scheduleLabel)
		if err != nil {
			__antithesis_instrumentation__.Notify(9381)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9382)
		}
		__antithesis_instrumentation__.Notify(9378)

		if exists {
			__antithesis_instrumentation__.Notify(9383)
			p.BufferClientNotice(ctx,
				pgnotice.Newf("schedule %q already exists, skipping", scheduleLabel),
			)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(9384)
		}
	} else {
		__antithesis_instrumentation__.Notify(9385)
	}
	__antithesis_instrumentation__.Notify(9353)

	env := sql.JobSchedulerEnv(p.ExecCfg())

	incRecurrence, err := computeScheduleRecurrence(env.Now(), eval.recurrence)
	if err != nil {
		__antithesis_instrumentation__.Notify(9386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9387)
	}
	__antithesis_instrumentation__.Notify(9354)
	fullRecurrence, err := computeScheduleRecurrence(env.Now(), eval.fullBackupRecurrence)
	if err != nil {
		__antithesis_instrumentation__.Notify(9388)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9389)
	}
	__antithesis_instrumentation__.Notify(9355)

	fullRecurrencePicked := false
	if incRecurrence != nil && func() bool {
		__antithesis_instrumentation__.Notify(9390)
		return fullRecurrence == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(9391)

		fullRecurrence = pickFullRecurrenceFromIncremental(incRecurrence)
		fullRecurrencePicked = true

		if fullRecurrence == forceFullBackup {
			__antithesis_instrumentation__.Notify(9392)
			fullRecurrence = incRecurrence
			incRecurrence = nil
		} else {
			__antithesis_instrumentation__.Notify(9393)
		}
	} else {
		__antithesis_instrumentation__.Notify(9394)
	}
	__antithesis_instrumentation__.Notify(9356)

	if fullRecurrence == nil {
		__antithesis_instrumentation__.Notify(9395)
		return errors.AssertionFailedf(" full backup recurrence should be set")
	} else {
		__antithesis_instrumentation__.Notify(9396)
	}
	__antithesis_instrumentation__.Notify(9357)

	backupNode := &tree.Backup{
		Options: tree.BackupOptions{
			CaptureRevisionHistory: eval.BackupOptions.CaptureRevisionHistory,
			Detached:               true,
		},
		Nested:         true,
		AppendToLatest: false,
	}

	if eval.encryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(9397)
		pw, err := eval.encryptionPassphrase()
		if err != nil {
			__antithesis_instrumentation__.Notify(9399)
			return errors.Wrapf(err, "failed to evaluate backup encryption_passphrase")
		} else {
			__antithesis_instrumentation__.Notify(9400)
		}
		__antithesis_instrumentation__.Notify(9398)
		backupNode.Options.EncryptionPassphrase = tree.NewStrVal(pw)
	} else {
		__antithesis_instrumentation__.Notify(9401)
	}
	__antithesis_instrumentation__.Notify(9358)

	var kmsURIs []string
	if eval.kmsURIs != nil {
		__antithesis_instrumentation__.Notify(9402)
		kmsURIs, err = eval.kmsURIs()
		if err != nil {
			__antithesis_instrumentation__.Notify(9404)
			return errors.Wrapf(err, "failed to evaluate backup kms_uri")
		} else {
			__antithesis_instrumentation__.Notify(9405)
		}
		__antithesis_instrumentation__.Notify(9403)
		for _, kmsURI := range kmsURIs {
			__antithesis_instrumentation__.Notify(9406)
			backupNode.Options.EncryptionKMSURI = append(backupNode.Options.EncryptionKMSURI,
				tree.NewStrVal(kmsURI))
		}
	} else {
		__antithesis_instrumentation__.Notify(9407)
	}
	__antithesis_instrumentation__.Notify(9359)

	destinations, err := eval.destination()
	if err != nil {
		__antithesis_instrumentation__.Notify(9408)
		return errors.Wrapf(err, "failed to evaluate backup destination paths")
	} else {
		__antithesis_instrumentation__.Notify(9409)
	}
	__antithesis_instrumentation__.Notify(9360)

	for _, dest := range destinations {
		__antithesis_instrumentation__.Notify(9410)
		backupNode.To = append(backupNode.To, tree.NewStrVal(dest))
	}
	__antithesis_instrumentation__.Notify(9361)

	backupNode.Targets = eval.Targets

	if err := dryRunBackup(ctx, p, backupNode); err != nil {
		__antithesis_instrumentation__.Notify(9411)
		return errors.Wrapf(err, "failed to dry run backup")
	} else {
		__antithesis_instrumentation__.Notify(9412)
	}
	__antithesis_instrumentation__.Notify(9362)

	var scheduleLabel string
	if eval.scheduleLabel != nil {
		__antithesis_instrumentation__.Notify(9413)
		label, err := eval.scheduleLabel()
		if err != nil {
			__antithesis_instrumentation__.Notify(9415)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9416)
		}
		__antithesis_instrumentation__.Notify(9414)
		scheduleLabel = label
	} else {
		__antithesis_instrumentation__.Notify(9417)
		scheduleLabel = fmt.Sprintf("BACKUP %d", env.Now().Unix())
	}
	__antithesis_instrumentation__.Notify(9363)

	scheduleOptions, err := eval.scheduleOpts()
	if err != nil {
		__antithesis_instrumentation__.Notify(9418)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9419)
	}
	__antithesis_instrumentation__.Notify(9364)

	if _, ignoreExisting := scheduleOptions[optIgnoreExistingBackups]; !ignoreExisting {
		__antithesis_instrumentation__.Notify(9420)
		if err := checkForExistingBackupsInCollection(ctx, p, destinations); err != nil {
			__antithesis_instrumentation__.Notify(9421)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9422)
		}
	} else {
		__antithesis_instrumentation__.Notify(9423)
	}
	__antithesis_instrumentation__.Notify(9365)

	_, updateMetricOnSuccess := scheduleOptions[optUpdatesLastBackupMetric]

	if updateMetricOnSuccess {
		__antithesis_instrumentation__.Notify(9424)

		if err := p.RequireAdminRole(ctx, optUpdatesLastBackupMetric); err != nil {
			__antithesis_instrumentation__.Notify(9425)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9426)
		}
	} else {
		__antithesis_instrumentation__.Notify(9427)
	}
	__antithesis_instrumentation__.Notify(9366)

	evalCtx := &p.ExtendedEvalContext().EvalContext
	firstRun, err := scheduleFirstRun(evalCtx, scheduleOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(9428)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9429)
	}
	__antithesis_instrumentation__.Notify(9367)

	details, err := makeScheduleDetails(scheduleOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(9430)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9431)
	}
	__antithesis_instrumentation__.Notify(9368)

	ex := p.ExecCfg().InternalExecutor

	unpauseOnSuccessID := jobs.InvalidScheduleID

	var chainProtectedTimestampRecords bool

	var inc *jobs.ScheduledJob
	var incScheduledBackupArgs *ScheduledBackupExecutionArgs
	if incRecurrence != nil {
		__antithesis_instrumentation__.Notify(9432)
		chainProtectedTimestampRecords = canChainProtectedTimestampRecords(p, eval)
		backupNode.AppendToLatest = true

		var incDests []string
		if eval.incrementalStorage != nil {
			__antithesis_instrumentation__.Notify(9437)
			incDests, err = eval.incrementalStorage()
			if err != nil {
				__antithesis_instrumentation__.Notify(9439)
				return err
			} else {
				__antithesis_instrumentation__.Notify(9440)
			}
			__antithesis_instrumentation__.Notify(9438)
			for _, incDest := range incDests {
				__antithesis_instrumentation__.Notify(9441)
				backupNode.Options.IncrementalStorage = append(backupNode.Options.IncrementalStorage, tree.NewStrVal(incDest))
			}
		} else {
			__antithesis_instrumentation__.Notify(9442)
		}
		__antithesis_instrumentation__.Notify(9433)
		inc, incScheduledBackupArgs, err = makeBackupSchedule(
			env, p.User(), scheduleLabel, incRecurrence, details, unpauseOnSuccessID,
			updateMetricOnSuccess, backupNode, chainProtectedTimestampRecords)
		if err != nil {
			__antithesis_instrumentation__.Notify(9443)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9444)
		}
		__antithesis_instrumentation__.Notify(9434)

		inc.Pause()
		inc.SetScheduleStatus("Waiting for initial backup to complete")

		if err := inc.Create(ctx, ex, p.ExtendedEvalContext().Txn); err != nil {
			__antithesis_instrumentation__.Notify(9445)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9446)
		}
		__antithesis_instrumentation__.Notify(9435)
		if err := emitSchedule(inc, backupNode, destinations, nil,
			kmsURIs, incDests, resultsCh); err != nil {
			__antithesis_instrumentation__.Notify(9447)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9448)
		}
		__antithesis_instrumentation__.Notify(9436)
		unpauseOnSuccessID = inc.ScheduleID()
	} else {
		__antithesis_instrumentation__.Notify(9449)
	}
	__antithesis_instrumentation__.Notify(9369)

	backupNode.AppendToLatest = false
	backupNode.Options.IncrementalStorage = nil
	var fullScheduledBackupArgs *ScheduledBackupExecutionArgs
	full, fullScheduledBackupArgs, err := makeBackupSchedule(
		env, p.User(), scheduleLabel, fullRecurrence, details, unpauseOnSuccessID,
		updateMetricOnSuccess, backupNode, chainProtectedTimestampRecords)
	if err != nil {
		__antithesis_instrumentation__.Notify(9450)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9451)
	}
	__antithesis_instrumentation__.Notify(9370)

	if firstRun != nil {
		__antithesis_instrumentation__.Notify(9452)
		full.SetNextRun(*firstRun)
	} else {
		__antithesis_instrumentation__.Notify(9453)
		if eval.isEnterpriseUser && func() bool {
			__antithesis_instrumentation__.Notify(9454)
			return fullRecurrencePicked == true
		}() == true {
			__antithesis_instrumentation__.Notify(9455)

			full.SetNextRun(env.Now())
		} else {
			__antithesis_instrumentation__.Notify(9456)
		}
	}
	__antithesis_instrumentation__.Notify(9371)

	if err := full.Create(ctx, ex, p.ExtendedEvalContext().Txn); err != nil {
		__antithesis_instrumentation__.Notify(9457)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9458)
	}
	__antithesis_instrumentation__.Notify(9372)

	if incRecurrence != nil {
		__antithesis_instrumentation__.Notify(9459)
		if err := setDependentSchedule(ctx, ex, fullScheduledBackupArgs, full, inc.ScheduleID(),
			p.ExtendedEvalContext().Txn); err != nil {
			__antithesis_instrumentation__.Notify(9461)
			return errors.Wrap(err,
				"failed to update full schedule with dependent incremental schedule id")
		} else {
			__antithesis_instrumentation__.Notify(9462)
		}
		__antithesis_instrumentation__.Notify(9460)
		if err := setDependentSchedule(ctx, ex, incScheduledBackupArgs, inc, full.ScheduleID(),
			p.ExtendedEvalContext().Txn); err != nil {
			__antithesis_instrumentation__.Notify(9463)
			return errors.Wrap(err,
				"failed to update incremental schedule with dependent full schedule id")
		} else {
			__antithesis_instrumentation__.Notify(9464)
		}
	} else {
		__antithesis_instrumentation__.Notify(9465)
	}
	__antithesis_instrumentation__.Notify(9373)

	collectScheduledBackupTelemetry(incRecurrence, firstRun, fullRecurrencePicked, details)
	return emitSchedule(full, backupNode, destinations, nil,
		kmsURIs, nil, resultsCh)
}

func setDependentSchedule(
	ctx context.Context,
	ex *sql.InternalExecutor,
	scheduleExecutionArgs *ScheduledBackupExecutionArgs,
	schedule *jobs.ScheduledJob,
	dependentID int64,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(9466)
	scheduleExecutionArgs.DependentScheduleID = dependentID
	any, err := pbtypes.MarshalAny(scheduleExecutionArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(9468)
		return errors.Wrap(err, "marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(9469)
	}
	__antithesis_instrumentation__.Notify(9467)
	schedule.SetExecutionDetails(
		schedule.ExecutorType(), jobspb.ExecutionArguments{Args: any},
	)
	return schedule.Update(ctx, ex, txn)
}

func checkForExistingBackupsInCollection(
	ctx context.Context, p sql.PlanHookState, destinations []string,
) error {
	__antithesis_instrumentation__.Notify(9470)
	makeCloudFactory := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI
	collectionURI, _, err := getURIsByLocalityKV(destinations, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(9474)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9475)
	}
	__antithesis_instrumentation__.Notify(9471)

	_, err = readLatestFile(ctx, collectionURI, makeCloudFactory, p.User())
	if err == nil {
		__antithesis_instrumentation__.Notify(9476)

		return errors.Newf("backups already created in %s; to ignore existing backups, "+
			"the schedule can be created with the 'ignore_existing_backups' option",
			collectionURI)
	} else {
		__antithesis_instrumentation__.Notify(9477)
	}
	__antithesis_instrumentation__.Notify(9472)
	if !errors.Is(err, cloud.ErrFileDoesNotExist) {
		__antithesis_instrumentation__.Notify(9478)
		return errors.Newf("unexpected error occurred when checking for existing backups in %s",
			collectionURI)
	} else {
		__antithesis_instrumentation__.Notify(9479)
	}
	__antithesis_instrumentation__.Notify(9473)

	return nil
}

func makeBackupSchedule(
	env scheduledjobs.JobSchedulerEnv,
	owner security.SQLUsername,
	label string,
	recurrence *scheduleRecurrence,
	details jobspb.ScheduleDetails,
	unpauseOnSuccess int64,
	updateLastMetricOnSuccess bool,
	backupNode *tree.Backup,
	chainProtectedTimestampRecords bool,
) (*jobs.ScheduledJob, *ScheduledBackupExecutionArgs, error) {
	__antithesis_instrumentation__.Notify(9480)
	sj := jobs.NewScheduledJob(env)
	sj.SetScheduleLabel(label)
	sj.SetOwner(owner)

	args := &ScheduledBackupExecutionArgs{
		UnpauseOnSuccess:               unpauseOnSuccess,
		UpdatesLastBackupMetric:        updateLastMetricOnSuccess,
		ChainProtectedTimestampRecords: chainProtectedTimestampRecords,
	}
	if backupNode.AppendToLatest {
		__antithesis_instrumentation__.Notify(9484)
		args.BackupType = ScheduledBackupExecutionArgs_INCREMENTAL
	} else {
		__antithesis_instrumentation__.Notify(9485)
		args.BackupType = ScheduledBackupExecutionArgs_FULL
	}
	__antithesis_instrumentation__.Notify(9481)

	if err := sj.SetSchedule(recurrence.cron); err != nil {
		__antithesis_instrumentation__.Notify(9486)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(9487)
	}
	__antithesis_instrumentation__.Notify(9482)

	sj.SetScheduleDetails(details)

	args.BackupStatement = tree.AsStringWithFlags(backupNode, tree.FmtParsable|tree.FmtShowPasswords)
	any, err := pbtypes.MarshalAny(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(9488)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(9489)
	}
	__antithesis_instrumentation__.Notify(9483)
	sj.SetExecutionDetails(
		tree.ScheduledBackupExecutor.InternalName(), jobspb.ExecutionArguments{Args: any},
	)

	return sj, args, nil
}

func emitSchedule(
	sj *jobs.ScheduledJob,
	backupNode *tree.Backup,
	to, incrementalFrom, kmsURIs []string,
	incrementalStorage []string,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(9490)
	var nextRun tree.Datum
	status := "ACTIVE"
	if sj.IsPaused() {
		__antithesis_instrumentation__.Notify(9493)
		nextRun = tree.DNull
		status = "PAUSED"
		if s := sj.ScheduleStatus(); s != "" {
			__antithesis_instrumentation__.Notify(9494)
			status += ": " + s
		} else {
			__antithesis_instrumentation__.Notify(9495)
		}
	} else {
		__antithesis_instrumentation__.Notify(9496)
		next, err := tree.MakeDTimestampTZ(sj.NextRun(), time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(9498)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9499)
		}
		__antithesis_instrumentation__.Notify(9497)
		nextRun = next
	}
	__antithesis_instrumentation__.Notify(9491)

	redactedBackupNode, err := GetRedactedBackupNode(backupNode, to, incrementalFrom, kmsURIs, "",
		incrementalStorage, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(9500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9501)
	}
	__antithesis_instrumentation__.Notify(9492)

	resultsCh <- tree.Datums{
		tree.NewDInt(tree.DInt(sj.ScheduleID())),
		tree.NewDString(sj.ScheduleLabel()),
		tree.NewDString(status),
		nextRun,
		tree.NewDString(sj.ScheduleExpr()),
		tree.NewDString(tree.AsString(redactedBackupNode)),
	}
	return nil
}

func checkScheduleAlreadyExists(
	ctx context.Context, p sql.PlanHookState, scheduleLabel string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(9502)

	row, err := p.ExecCfg().InternalExecutor.QueryRowEx(ctx, "check-sched",
		p.ExtendedEvalContext().Txn, sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("SELECT count(schedule_name) FROM %s WHERE schedule_name = '%s'",
			scheduledjobs.ProdJobSchedulerEnv.ScheduledJobsTableName(), scheduleLabel))

	if err != nil {
		__antithesis_instrumentation__.Notify(9504)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(9505)
	}
	__antithesis_instrumentation__.Notify(9503)
	return int64(tree.MustBeDInt(row[0])) != 0, nil
}

func dryRunBackup(ctx context.Context, p sql.PlanHookState, backupNode *tree.Backup) error {
	__antithesis_instrumentation__.Notify(9506)
	sp, err := p.ExtendedEvalContext().Txn.CreateSavepoint(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(9509)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9510)
	}
	__antithesis_instrumentation__.Notify(9507)
	err = dryRunInvokeBackup(ctx, p, backupNode)
	if rollbackErr := p.ExtendedEvalContext().Txn.RollbackToSavepoint(ctx, sp); rollbackErr != nil {
		__antithesis_instrumentation__.Notify(9511)
		return rollbackErr
	} else {
		__antithesis_instrumentation__.Notify(9512)
	}
	__antithesis_instrumentation__.Notify(9508)
	return err
}

func dryRunInvokeBackup(ctx context.Context, p sql.PlanHookState, backupNode *tree.Backup) error {
	__antithesis_instrumentation__.Notify(9513)
	backupFn, err := planBackup(ctx, p, backupNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(9515)
		return err
	} else {
		__antithesis_instrumentation__.Notify(9516)
	}
	__antithesis_instrumentation__.Notify(9514)
	return invokeBackup(ctx, backupFn)
}

func fullyQualifyScheduledBackupTargetTables(
	ctx context.Context, p sql.PlanHookState, tables tree.TablePatterns,
) ([]tree.TablePattern, error) {
	__antithesis_instrumentation__.Notify(9517)
	fqTablePatterns := make([]tree.TablePattern, len(tables))
	for i, target := range tables {
		__antithesis_instrumentation__.Notify(9519)
		tablePattern, err := target.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(9521)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9522)
		}
		__antithesis_instrumentation__.Notify(9520)
		switch tp := tablePattern.(type) {
		case *tree.TableName:
			__antithesis_instrumentation__.Notify(9523)
			if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn,
				col *descs.Collection) error {
				__antithesis_instrumentation__.Notify(9526)

				un := tp.ToUnresolvedObjectName()
				found, _, tableDesc, err := resolver.ResolveExisting(ctx, un, p, tree.ObjectLookupFlags{},
					p.CurrentDatabase(), p.CurrentSearchPath())
				if err != nil {
					__antithesis_instrumentation__.Notify(9532)
					return err
				} else {
					__antithesis_instrumentation__.Notify(9533)
				}
				__antithesis_instrumentation__.Notify(9527)
				if !found {
					__antithesis_instrumentation__.Notify(9534)
					return errors.Newf("target table %s could not be resolved", tp.String())
				} else {
					__antithesis_instrumentation__.Notify(9535)
				}
				__antithesis_instrumentation__.Notify(9528)

				found, dbDesc, err := col.GetImmutableDatabaseByID(ctx, txn, tableDesc.GetParentID(),
					tree.DatabaseLookupFlags{Required: true})
				if err != nil {
					__antithesis_instrumentation__.Notify(9536)
					return err
				} else {
					__antithesis_instrumentation__.Notify(9537)
				}
				__antithesis_instrumentation__.Notify(9529)
				if !found {
					__antithesis_instrumentation__.Notify(9538)
					return errors.Newf("database of target table %s could not be resolved", tp.String())
				} else {
					__antithesis_instrumentation__.Notify(9539)
				}
				__antithesis_instrumentation__.Notify(9530)

				schemaDesc, err := col.GetImmutableSchemaByID(ctx, txn, tableDesc.GetParentSchemaID(),
					tree.SchemaLookupFlags{Required: true})
				if err != nil {
					__antithesis_instrumentation__.Notify(9540)
					return err
				} else {
					__antithesis_instrumentation__.Notify(9541)
				}
				__antithesis_instrumentation__.Notify(9531)
				tn := tree.NewTableNameWithSchema(
					tree.Name(dbDesc.GetName()),
					tree.Name(schemaDesc.GetName()),
					tree.Name(tableDesc.GetName()),
				)
				fqTablePatterns[i] = tn
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(9542)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(9543)
			}
		case *tree.AllTablesSelector:
			__antithesis_instrumentation__.Notify(9524)
			if !tp.ExplicitSchema {
				__antithesis_instrumentation__.Notify(9544)
				tp.ExplicitSchema = true
				tp.SchemaName = tree.Name(p.CurrentDatabase())
			} else {
				__antithesis_instrumentation__.Notify(9545)
				if tp.ExplicitSchema && func() bool {
					__antithesis_instrumentation__.Notify(9546)
					return !tp.ExplicitCatalog == true
				}() == true {
					__antithesis_instrumentation__.Notify(9547)

					var schemaID descpb.ID
					if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
						__antithesis_instrumentation__.Notify(9549)
						flags := tree.DatabaseLookupFlags{Required: true}
						dbDesc, err := col.GetImmutableDatabaseByName(ctx, txn, p.CurrentDatabase(), flags)
						if err != nil {
							__antithesis_instrumentation__.Notify(9551)
							return err
						} else {
							__antithesis_instrumentation__.Notify(9552)
						}
						__antithesis_instrumentation__.Notify(9550)
						schemaID, err = col.Direct().ResolveSchemaID(ctx, txn, dbDesc.GetID(), tp.SchemaName.String())
						return err
					}); err != nil {
						__antithesis_instrumentation__.Notify(9553)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(9554)
					}
					__antithesis_instrumentation__.Notify(9548)

					if schemaID != descpb.InvalidID {
						__antithesis_instrumentation__.Notify(9555)
						tp.ExplicitCatalog = true
						tp.CatalogName = tree.Name(p.CurrentDatabase())
					} else {
						__antithesis_instrumentation__.Notify(9556)
					}
				} else {
					__antithesis_instrumentation__.Notify(9557)
				}
			}
			__antithesis_instrumentation__.Notify(9525)
			fqTablePatterns[i] = tp
		}
	}
	__antithesis_instrumentation__.Notify(9518)
	return fqTablePatterns, nil
}

func makeScheduledBackupEval(
	ctx context.Context, p sql.PlanHookState, schedule *tree.ScheduledBackup,
) (*scheduledBackupEval, error) {
	__antithesis_instrumentation__.Notify(9558)
	var err error
	if schedule.Targets != nil && func() bool {
		__antithesis_instrumentation__.Notify(9569)
		return schedule.Targets.Tables != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(9570)

		schedule.Targets.Tables, err = fullyQualifyScheduledBackupTargetTables(ctx, p,
			schedule.Targets.Tables)
		if err != nil {
			__antithesis_instrumentation__.Notify(9571)
			return nil, errors.Wrap(err, "qualifying backup target tables")
		} else {
			__antithesis_instrumentation__.Notify(9572)
		}
	} else {
		__antithesis_instrumentation__.Notify(9573)
	}
	__antithesis_instrumentation__.Notify(9559)

	eval := &scheduledBackupEval{ScheduledBackup: schedule}

	if schedule.ScheduleLabelSpec.Label != nil {
		__antithesis_instrumentation__.Notify(9574)
		eval.scheduleLabel, err = p.TypeAsString(ctx, schedule.ScheduleLabelSpec.Label, scheduleBackupOp)
		if err != nil {
			__antithesis_instrumentation__.Notify(9575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9576)
		}
	} else {
		__antithesis_instrumentation__.Notify(9577)
	}
	__antithesis_instrumentation__.Notify(9560)

	if schedule.Recurrence == nil {
		__antithesis_instrumentation__.Notify(9578)

		return nil, errors.New("RECURRING clause required")
	} else {
		__antithesis_instrumentation__.Notify(9579)
	}
	__antithesis_instrumentation__.Notify(9561)

	eval.recurrence, err = p.TypeAsString(ctx, schedule.Recurrence, scheduleBackupOp)
	if err != nil {
		__antithesis_instrumentation__.Notify(9580)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9581)
	}
	__antithesis_instrumentation__.Notify(9562)

	enterpriseCheckErr := utilccl.CheckEnterpriseEnabled(
		p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(),
		"BACKUP INTO LATEST")
	eval.isEnterpriseUser = enterpriseCheckErr == nil

	if eval.isEnterpriseUser && func() bool {
		__antithesis_instrumentation__.Notify(9582)
		return schedule.FullBackup != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(9583)
		if schedule.FullBackup.AlwaysFull {
			__antithesis_instrumentation__.Notify(9584)
			eval.fullBackupRecurrence = eval.recurrence
			eval.recurrence = nil
		} else {
			__antithesis_instrumentation__.Notify(9585)
			eval.fullBackupRecurrence, err = p.TypeAsString(
				ctx, schedule.FullBackup.Recurrence, scheduleBackupOp)
			if err != nil {
				__antithesis_instrumentation__.Notify(9586)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(9587)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(9588)
		if !eval.isEnterpriseUser {
			__antithesis_instrumentation__.Notify(9589)
			if schedule.FullBackup == nil || func() bool {
				__antithesis_instrumentation__.Notify(9590)
				return schedule.FullBackup.AlwaysFull == true
			}() == true {
				__antithesis_instrumentation__.Notify(9591)

				eval.fullBackupRecurrence = eval.recurrence
				eval.recurrence = nil
			} else {
				__antithesis_instrumentation__.Notify(9592)

				return nil, enterpriseCheckErr
			}
		} else {
			__antithesis_instrumentation__.Notify(9593)
		}
	}
	__antithesis_instrumentation__.Notify(9563)

	eval.scheduleOpts, err = p.TypeAsStringOpts(
		ctx, schedule.ScheduleOptions, scheduledBackupOptionExpectValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(9594)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9595)
	}
	__antithesis_instrumentation__.Notify(9564)

	eval.destination, err = p.TypeAsStringArray(ctx, tree.Exprs(schedule.To), scheduleBackupOp)
	if err != nil {
		__antithesis_instrumentation__.Notify(9596)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9597)
	}
	__antithesis_instrumentation__.Notify(9565)
	if schedule.BackupOptions.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(9598)
		eval.encryptionPassphrase, err =
			p.TypeAsString(ctx, schedule.BackupOptions.EncryptionPassphrase, scheduleBackupOp)
		if err != nil {
			__antithesis_instrumentation__.Notify(9599)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9600)
		}
	} else {
		__antithesis_instrumentation__.Notify(9601)
	}
	__antithesis_instrumentation__.Notify(9566)

	if schedule.BackupOptions.EncryptionKMSURI != nil {
		__antithesis_instrumentation__.Notify(9602)
		eval.kmsURIs, err = p.TypeAsStringArray(ctx, tree.Exprs(schedule.BackupOptions.EncryptionKMSURI),
			scheduleBackupOp)
		if err != nil {
			__antithesis_instrumentation__.Notify(9603)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9604)
		}
	} else {
		__antithesis_instrumentation__.Notify(9605)
	}
	__antithesis_instrumentation__.Notify(9567)
	if schedule.BackupOptions.IncrementalStorage != nil {
		__antithesis_instrumentation__.Notify(9606)
		eval.incrementalStorage, err = p.TypeAsStringArray(ctx,
			tree.Exprs(schedule.BackupOptions.IncrementalStorage),
			scheduleBackupOp)
		if err != nil {
			__antithesis_instrumentation__.Notify(9607)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9608)
		}
	} else {
		__antithesis_instrumentation__.Notify(9609)
	}
	__antithesis_instrumentation__.Notify(9568)
	return eval, nil
}

var scheduledBackupHeader = colinfo.ResultColumns{
	{Name: "schedule_id", Typ: types.Int},
	{Name: "label", Typ: types.String},
	{Name: "status", Typ: types.String},
	{Name: "first_run", Typ: types.TimestampTZ},
	{Name: "schedule", Typ: types.String},
	{Name: "backup_stmt", Typ: types.String},
}

func collectScheduledBackupTelemetry(
	incRecurrence *scheduleRecurrence,
	firstRun *time.Time,
	fullRecurrencePicked bool,
	details jobspb.ScheduleDetails,
) {
	__antithesis_instrumentation__.Notify(9610)
	telemetry.Count("scheduled-backup.create.success")
	if incRecurrence != nil {
		__antithesis_instrumentation__.Notify(9615)
		telemetry.Count("scheduled-backup.incremental")
	} else {
		__antithesis_instrumentation__.Notify(9616)
	}
	__antithesis_instrumentation__.Notify(9611)
	if firstRun != nil {
		__antithesis_instrumentation__.Notify(9617)
		telemetry.Count("scheduled-backup.first-run-picked")
	} else {
		__antithesis_instrumentation__.Notify(9618)
	}
	__antithesis_instrumentation__.Notify(9612)
	if fullRecurrencePicked {
		__antithesis_instrumentation__.Notify(9619)
		telemetry.Count("scheduled-backup.full-recurrence-picked")
	} else {
		__antithesis_instrumentation__.Notify(9620)
	}
	__antithesis_instrumentation__.Notify(9613)
	switch details.Wait {
	case jobspb.ScheduleDetails_WAIT:
		__antithesis_instrumentation__.Notify(9621)
		telemetry.Count("scheduled-backup.wait-policy.wait")
	case jobspb.ScheduleDetails_NO_WAIT:
		__antithesis_instrumentation__.Notify(9622)
		telemetry.Count("scheduled-backup.wait-policy.no-wait")
	case jobspb.ScheduleDetails_SKIP:
		__antithesis_instrumentation__.Notify(9623)
		telemetry.Count("scheduled-backup.wait-policy.skip")
	default:
		__antithesis_instrumentation__.Notify(9624)
	}
	__antithesis_instrumentation__.Notify(9614)
	switch details.OnError {
	case jobspb.ScheduleDetails_RETRY_SCHED:
		__antithesis_instrumentation__.Notify(9625)
		telemetry.Count("scheduled-backup.error-policy.retry-schedule")
	case jobspb.ScheduleDetails_RETRY_SOON:
		__antithesis_instrumentation__.Notify(9626)
		telemetry.Count("scheduled-backup.error-policy.retry-soon")
	case jobspb.ScheduleDetails_PAUSE_SCHED:
		__antithesis_instrumentation__.Notify(9627)
		telemetry.Count("scheduled-backup.error-policy.pause-schedule")
	default:
		__antithesis_instrumentation__.Notify(9628)
	}
}

func createBackupScheduleHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(9629)
	schedule, ok := stmt.(*tree.ScheduledBackup)
	if !ok {
		__antithesis_instrumentation__.Notify(9633)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(9634)
	}
	__antithesis_instrumentation__.Notify(9630)

	eval, err := makeScheduledBackupEval(ctx, p, schedule)
	if err != nil {
		__antithesis_instrumentation__.Notify(9635)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(9636)
	}
	__antithesis_instrumentation__.Notify(9631)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(9637)
		err := doCreateBackupSchedules(ctx, p, eval, resultsCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(9639)
			telemetry.Count("scheduled-backup.create.failed")
			return err
		} else {
			__antithesis_instrumentation__.Notify(9640)
		}
		__antithesis_instrumentation__.Notify(9638)

		return nil
	}
	__antithesis_instrumentation__.Notify(9632)
	return fn, scheduledBackupHeader, nil, false, nil
}

func (m ScheduledBackupExecutionArgs) MarshalJSONPB(marshaller *jsonpb.Marshaler) ([]byte, error) {
	__antithesis_instrumentation__.Notify(9641)
	if !protoreflect.ShouldRedact(marshaller) {
		__antithesis_instrumentation__.Notify(9650)
		return json.Marshal(m)
	} else {
		__antithesis_instrumentation__.Notify(9651)
	}
	__antithesis_instrumentation__.Notify(9642)

	stmt, err := parser.ParseOne(m.BackupStatement)
	if err != nil {
		__antithesis_instrumentation__.Notify(9652)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9653)
	}
	__antithesis_instrumentation__.Notify(9643)
	backup, ok := stmt.AST.(*tree.Backup)
	if !ok {
		__antithesis_instrumentation__.Notify(9654)
		return nil, errors.Errorf("unexpected %T statement in backup schedule: %v", backup, backup)
	} else {
		__antithesis_instrumentation__.Notify(9655)
	}
	__antithesis_instrumentation__.Notify(9644)

	for i := range backup.To {
		__antithesis_instrumentation__.Notify(9656)
		raw, ok := backup.To[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(9659)
			return nil, errors.Errorf("unexpected %T arg in backup schedule: %v", raw, raw)
		} else {
			__antithesis_instrumentation__.Notify(9660)
		}
		__antithesis_instrumentation__.Notify(9657)
		clean, err := cloud.SanitizeExternalStorageURI(raw.RawString(), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(9661)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9662)
		}
		__antithesis_instrumentation__.Notify(9658)
		backup.To[i] = tree.NewDString(clean)
	}
	__antithesis_instrumentation__.Notify(9645)

	for i := range backup.IncrementalFrom {
		__antithesis_instrumentation__.Notify(9663)
		raw, ok := backup.IncrementalFrom[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(9666)
			return nil, errors.Errorf("unexpected %T arg in backup schedule: %v", raw, raw)
		} else {
			__antithesis_instrumentation__.Notify(9667)
		}
		__antithesis_instrumentation__.Notify(9664)
		clean, err := cloud.SanitizeExternalStorageURI(raw.RawString(), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(9668)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9669)
		}
		__antithesis_instrumentation__.Notify(9665)
		backup.IncrementalFrom[i] = tree.NewDString(clean)
	}
	__antithesis_instrumentation__.Notify(9646)

	for i := range backup.Options.IncrementalStorage {
		__antithesis_instrumentation__.Notify(9670)
		raw, ok := backup.Options.IncrementalStorage[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(9673)
			return nil, errors.Errorf("unexpected %T arg in backup schedule: %v", raw, raw)
		} else {
			__antithesis_instrumentation__.Notify(9674)
		}
		__antithesis_instrumentation__.Notify(9671)
		clean, err := cloud.SanitizeExternalStorageURI(raw.RawString(), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(9675)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9676)
		}
		__antithesis_instrumentation__.Notify(9672)
		backup.Options.IncrementalStorage[i] = tree.NewDString(clean)
	}
	__antithesis_instrumentation__.Notify(9647)

	for i := range backup.Options.EncryptionKMSURI {
		__antithesis_instrumentation__.Notify(9677)
		raw, ok := backup.Options.EncryptionKMSURI[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(9680)
			return nil, errors.Errorf("unexpected %T arg in backup schedule: %v", raw, raw)
		} else {
			__antithesis_instrumentation__.Notify(9681)
		}
		__antithesis_instrumentation__.Notify(9678)
		clean, err := cloud.RedactKMSURI(raw.RawString())
		if err != nil {
			__antithesis_instrumentation__.Notify(9682)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(9683)
		}
		__antithesis_instrumentation__.Notify(9679)
		backup.Options.EncryptionKMSURI[i] = tree.NewDString(clean)
	}
	__antithesis_instrumentation__.Notify(9648)

	if backup.Options.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(9684)
		backup.Options.EncryptionPassphrase = tree.NewDString("redacted")
	} else {
		__antithesis_instrumentation__.Notify(9685)
	}
	__antithesis_instrumentation__.Notify(9649)

	m.BackupStatement = backup.String()
	return json.Marshal(m)
}

func init() {
	sql.AddPlanHook("schedule backup", createBackupScheduleHook)
}
