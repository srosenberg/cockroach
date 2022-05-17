package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/gogo/protobuf/jsonpb"
)

type jobDumpTraceMode int64

const (
	noDump jobDumpTraceMode = iota

	dumpOnFail

	dumpOnStop
)

var traceableJobDumpTraceMode = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"jobs.trace.force_dump_mode",
	"determines the state in which all traceable jobs will dump their cluster wide, inflight, "+
		"trace recordings. Traces may be dumped never, on fail, "+
		"or on any status change i.e paused, canceled, failed, succeeded.",
	"never",
	map[int64]string{
		int64(noDump):     "never",
		int64(dumpOnFail): "onFail",
		int64(dumpOnStop): "onStop",
	},
)

type Job struct {
	registry *Registry

	id        jobspb.JobID
	createdBy *CreatedByInfo
	session   sqlliveness.Session
	mu        struct {
		syncutil.Mutex
		payload  jobspb.Payload
		progress jobspb.Progress
		status   Status
		runStats *RunStats
	}
}

type CreatedByInfo struct {
	Name string
	ID   int64
}

type Record struct {
	JobID         jobspb.JobID
	Description   string
	Statements    []string
	Username      security.SQLUsername
	DescriptorIDs descpb.IDs
	Details       jobspb.Details
	Progress      jobspb.ProgressDetails
	RunningStatus RunningStatus

	NonCancelable bool

	CreatedBy *CreatedByInfo
}

func (r *Record) AppendDescription(description string) {
	__antithesis_instrumentation__.Notify(70470)
	if len(r.Description) == 0 {
		__antithesis_instrumentation__.Notify(70472)
		r.Description = description
		return
	} else {
		__antithesis_instrumentation__.Notify(70473)
	}
	__antithesis_instrumentation__.Notify(70471)
	r.Description = r.Description + "; " + description
}

func (r *Record) SetNonCancelable(ctx context.Context, updateFn NonCancelableUpdateFn) {
	__antithesis_instrumentation__.Notify(70474)
	r.NonCancelable = updateFn(ctx, r.NonCancelable)
}

type StartableJob struct {
	*Job
	txn        *kv.Txn
	resumer    Resumer
	resumerCtx context.Context
	cancel     context.CancelFunc
	execDone   chan struct{}
	execErr    error
	starts     int64
}

type TraceableJob interface {
	ForceRealSpan() bool
}

func init() {

	var jobPayload jobspb.Payload
	jobsDetailsInterfaceType := reflect.TypeOf(&jobPayload.Details).Elem()
	var jobProgress jobspb.Progress
	jobsProgressDetailsInterfaceType := reflect.TypeOf(&jobProgress.Details).Elem()
	protoutil.RegisterUnclonableType(jobsDetailsInterfaceType, reflect.Array)
	protoutil.RegisterUnclonableType(jobsProgressDetailsInterfaceType, reflect.Array)

}

type Status string

func (s Status) SafeFormat(sp redact.SafePrinter, verb rune) {
	__antithesis_instrumentation__.Notify(70475)
	sp.SafeString(redact.SafeString(s))
}

var _ redact.SafeFormatter = Status("")

type RunningStatus string

const (
	StatusPending Status = "pending"

	StatusRunning Status = "running"

	StatusPaused Status = "paused"

	StatusFailed Status = "failed"

	StatusReverting Status = "reverting"

	StatusSucceeded Status = "succeeded"

	StatusCanceled Status = "canceled"

	StatusCancelRequested Status = "cancel-requested"

	StatusPauseRequested Status = "pause-requested"

	StatusRevertFailed Status = "revert-failed"
)

var (
	errJobCanceled = errors.New("job canceled by user")
)

func HasErrJobCanceled(err error) bool {
	__antithesis_instrumentation__.Notify(70476)
	return errors.Is(err, errJobCanceled)
}

func deprecatedIsOldSchemaChangeJob(payload *jobspb.Payload) bool {
	__antithesis_instrumentation__.Notify(70477)
	schemaChangeDetails, ok := payload.UnwrapDetails().(jobspb.SchemaChangeDetails)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(70478)
		return schemaChangeDetails.FormatVersion < jobspb.JobResumerFormatVersion == true
	}() == true
}

func (s Status) Terminal() bool {
	__antithesis_instrumentation__.Notify(70479)
	return s == StatusFailed || func() bool {
		__antithesis_instrumentation__.Notify(70480)
		return s == StatusSucceeded == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(70481)
		return s == StatusCanceled == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(70482)
		return s == StatusRevertFailed == true
	}() == true
}

func (j *Job) ID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(70483)
	return j.id
}

func (j *Job) Session() sqlliveness.Session {
	__antithesis_instrumentation__.Notify(70484)
	return j.session
}

func (j *Job) CreatedBy() *CreatedByInfo {
	__antithesis_instrumentation__.Notify(70485)
	return j.createdBy
}

func (j *Job) taskName() string {
	__antithesis_instrumentation__.Notify(70486)
	return fmt.Sprintf(`job-%d`, j.ID())
}

func (j *Job) started(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(70487)
	return j.Update(ctx, txn, func(_ *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70488)
		if md.Status != StatusPending && func() bool {
			__antithesis_instrumentation__.Notify(70492)
			return md.Status != StatusRunning == true
		}() == true {
			__antithesis_instrumentation__.Notify(70493)
			return errors.Errorf("job with status %s cannot be marked started", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70494)
		}
		__antithesis_instrumentation__.Notify(70489)
		if md.Payload.StartedMicros == 0 {
			__antithesis_instrumentation__.Notify(70495)
			ju.UpdateStatus(StatusRunning)
			md.Payload.StartedMicros = timeutil.ToUnixMicros(j.registry.clock.Now().GoTime())
			ju.UpdatePayload(md.Payload)
		} else {
			__antithesis_instrumentation__.Notify(70496)
		}
		__antithesis_instrumentation__.Notify(70490)

		if md.RunStats != nil {
			__antithesis_instrumentation__.Notify(70497)
			ju.UpdateRunStats(md.RunStats.NumRuns+1, j.registry.clock.Now().GoTime())
		} else {
			__antithesis_instrumentation__.Notify(70498)
		}
		__antithesis_instrumentation__.Notify(70491)
		return nil
	})
}

func (j *Job) CheckStatus(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(70499)
	return j.Update(ctx, txn, func(_ *kv.Txn, md JobMetadata, _ *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70500)
		return md.CheckRunningOrReverting()
	})
}

func (j *Job) CheckTerminalStatus(ctx context.Context, txn *kv.Txn) bool {
	__antithesis_instrumentation__.Notify(70501)
	err := j.Update(ctx, txn, func(_ *kv.Txn, md JobMetadata, _ *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70503)
		if !md.Status.Terminal() {
			__antithesis_instrumentation__.Notify(70505)
			return &InvalidStatusError{md.ID, md.Status, "checking that job status is success", md.Payload.Error}
		} else {
			__antithesis_instrumentation__.Notify(70506)
		}
		__antithesis_instrumentation__.Notify(70504)
		return nil
	})
	__antithesis_instrumentation__.Notify(70502)

	return err == nil
}

func (j *Job) RunningStatus(
	ctx context.Context, txn *kv.Txn, runningStatusFn RunningStatusFn,
) error {
	__antithesis_instrumentation__.Notify(70507)
	return j.Update(ctx, txn, func(_ *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70508)
		if err := md.CheckRunningOrReverting(); err != nil {
			__antithesis_instrumentation__.Notify(70511)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70512)
		}
		__antithesis_instrumentation__.Notify(70509)
		runningStatus, err := runningStatusFn(ctx, md.Progress.Details)
		if err != nil {
			__antithesis_instrumentation__.Notify(70513)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70514)
		}
		__antithesis_instrumentation__.Notify(70510)
		md.Progress.RunningStatus = string(runningStatus)
		ju.UpdateProgress(md.Progress)
		return nil
	})
}

type RunningStatusFn func(ctx context.Context, details jobspb.Details) (RunningStatus, error)

type NonCancelableUpdateFn func(ctx context.Context, nonCancelable bool) bool

type FractionProgressedFn func(ctx context.Context, details jobspb.ProgressDetails) float32

func FractionUpdater(f float32) FractionProgressedFn {
	__antithesis_instrumentation__.Notify(70515)
	return func(ctx context.Context, details jobspb.ProgressDetails) float32 {
		__antithesis_instrumentation__.Notify(70516)
		return f
	}
}

func (j *Job) FractionProgressed(
	ctx context.Context, txn *kv.Txn, progressedFn FractionProgressedFn,
) error {
	__antithesis_instrumentation__.Notify(70517)
	return j.Update(ctx, txn, func(_ *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70518)
		if err := md.CheckRunningOrReverting(); err != nil {
			__antithesis_instrumentation__.Notify(70522)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70523)
		}
		__antithesis_instrumentation__.Notify(70519)
		fractionCompleted := progressedFn(ctx, md.Progress.Details)

		if fractionCompleted > 1.0 && func() bool {
			__antithesis_instrumentation__.Notify(70524)
			return fractionCompleted < 1.01 == true
		}() == true {
			__antithesis_instrumentation__.Notify(70525)
			fractionCompleted = 1.0
		} else {
			__antithesis_instrumentation__.Notify(70526)
		}
		__antithesis_instrumentation__.Notify(70520)
		if fractionCompleted < 0.0 || func() bool {
			__antithesis_instrumentation__.Notify(70527)
			return fractionCompleted > 1.0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(70528)
			return errors.Errorf(
				"job %d: fractionCompleted %f is outside allowable range [0.0, 1.0]",
				j.ID(), fractionCompleted,
			)
		} else {
			__antithesis_instrumentation__.Notify(70529)
		}
		__antithesis_instrumentation__.Notify(70521)
		md.Progress.Progress = &jobspb.Progress_FractionCompleted{
			FractionCompleted: fractionCompleted,
		}
		ju.UpdateProgress(md.Progress)
		return nil
	})
}

func (j *Job) paused(
	ctx context.Context, txn *kv.Txn, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70530)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70531)
		if md.Status == StatusPaused {
			__antithesis_instrumentation__.Notify(70535)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(70536)
		}
		__antithesis_instrumentation__.Notify(70532)
		if md.Status != StatusPauseRequested {
			__antithesis_instrumentation__.Notify(70537)
			return fmt.Errorf("job with status %s cannot be set to paused", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70538)
		}
		__antithesis_instrumentation__.Notify(70533)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70539)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70540)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70541)
			}
		} else {
			__antithesis_instrumentation__.Notify(70542)
		}
		__antithesis_instrumentation__.Notify(70534)
		ju.UpdateStatus(StatusPaused)
		return nil
	})
}

func (j *Job) unpaused(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(70543)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70544)
		if md.Status == StatusRunning || func() bool {
			__antithesis_instrumentation__.Notify(70548)
			return md.Status == StatusReverting == true
		}() == true {
			__antithesis_instrumentation__.Notify(70549)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(70550)
		}
		__antithesis_instrumentation__.Notify(70545)
		if md.Status != StatusPaused {
			__antithesis_instrumentation__.Notify(70551)
			return fmt.Errorf("job with status %s cannot be resumed", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70552)
		}
		__antithesis_instrumentation__.Notify(70546)

		if md.Payload.FinalResumeError == nil {
			__antithesis_instrumentation__.Notify(70553)
			ju.UpdateStatus(StatusRunning)
		} else {
			__antithesis_instrumentation__.Notify(70554)
			ju.UpdateStatus(StatusReverting)
		}
		__antithesis_instrumentation__.Notify(70547)
		ju.UpdatePayload(md.Payload)
		return nil
	})
}

func (j *Job) cancelRequested(
	ctx context.Context, txn *kv.Txn, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70555)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70556)

		if deprecatedIsOldSchemaChangeJob(md.Payload) {
			__antithesis_instrumentation__.Notify(70563)
			return errors.Newf(
				"schema change job was created in earlier version, and cannot be " +
					"canceled in this version until the upgrade is finalized and an internal migration is complete")
		} else {
			__antithesis_instrumentation__.Notify(70564)
		}
		__antithesis_instrumentation__.Notify(70557)

		if md.Payload.Noncancelable {
			__antithesis_instrumentation__.Notify(70565)
			return errors.Newf("job %d: not cancelable", j.ID())
		} else {
			__antithesis_instrumentation__.Notify(70566)
		}
		__antithesis_instrumentation__.Notify(70558)
		if md.Status == StatusCancelRequested || func() bool {
			__antithesis_instrumentation__.Notify(70567)
			return md.Status == StatusCanceled == true
		}() == true {
			__antithesis_instrumentation__.Notify(70568)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(70569)
		}
		__antithesis_instrumentation__.Notify(70559)
		if md.Status != StatusPending && func() bool {
			__antithesis_instrumentation__.Notify(70570)
			return md.Status != StatusRunning == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(70571)
			return md.Status != StatusPaused == true
		}() == true {
			__antithesis_instrumentation__.Notify(70572)
			return fmt.Errorf("job with status %s cannot be requested to be canceled", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70573)
		}
		__antithesis_instrumentation__.Notify(70560)
		if md.Status == StatusPaused && func() bool {
			__antithesis_instrumentation__.Notify(70574)
			return md.Payload.FinalResumeError != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(70575)
			decodedErr := errors.DecodeError(ctx, *md.Payload.FinalResumeError)
			return errors.Wrapf(decodedErr, "job %d is paused and has non-nil FinalResumeError "+
				"hence cannot be canceled and should be reverted", j.ID())
		} else {
			__antithesis_instrumentation__.Notify(70576)
		}
		__antithesis_instrumentation__.Notify(70561)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70577)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70578)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70579)
			}
		} else {
			__antithesis_instrumentation__.Notify(70580)
		}
		__antithesis_instrumentation__.Notify(70562)
		ju.UpdateStatus(StatusCancelRequested)
		return nil
	})
}

type onPauseRequestFunc func(
	ctx context.Context, planHookState interface{}, txn *kv.Txn, progress *jobspb.Progress,
) error

func (j *Job) PauseRequested(
	ctx context.Context, txn *kv.Txn, fn onPauseRequestFunc, reason string,
) error {
	__antithesis_instrumentation__.Notify(70581)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70582)

		if deprecatedIsOldSchemaChangeJob(md.Payload) {
			__antithesis_instrumentation__.Notify(70587)
			return errors.Newf(
				"schema change job was created in earlier version, and cannot be " +
					"paused in this version until the upgrade is finalized and an internal migration is complete")
		} else {
			__antithesis_instrumentation__.Notify(70588)
		}
		__antithesis_instrumentation__.Notify(70583)

		if md.Status == StatusPauseRequested || func() bool {
			__antithesis_instrumentation__.Notify(70589)
			return md.Status == StatusPaused == true
		}() == true {
			__antithesis_instrumentation__.Notify(70590)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(70591)
		}
		__antithesis_instrumentation__.Notify(70584)
		if md.Status != StatusPending && func() bool {
			__antithesis_instrumentation__.Notify(70592)
			return md.Status != StatusRunning == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(70593)
			return md.Status != StatusReverting == true
		}() == true {
			__antithesis_instrumentation__.Notify(70594)
			return fmt.Errorf("job with status %s cannot be requested to be paused", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70595)
		}
		__antithesis_instrumentation__.Notify(70585)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70596)
			execCtx, cleanup := j.registry.execCtx("pause request", j.Payload().UsernameProto.Decode())
			defer cleanup()
			if err := fn(ctx, execCtx, txn, md.Progress); err != nil {
				__antithesis_instrumentation__.Notify(70598)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70599)
			}
			__antithesis_instrumentation__.Notify(70597)
			ju.UpdateProgress(md.Progress)
		} else {
			__antithesis_instrumentation__.Notify(70600)
		}
		__antithesis_instrumentation__.Notify(70586)
		ju.UpdateStatus(StatusPauseRequested)
		md.Payload.PauseReason = reason
		ju.UpdatePayload(md.Payload)
		log.Infof(ctx, "job %d: pause requested recorded with reason %s", j.ID(), reason)
		return nil
	})
}

func (j *Job) reverted(
	ctx context.Context, txn *kv.Txn, err error, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70601)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70602)
		if md.Status != StatusReverting && func() bool {
			__antithesis_instrumentation__.Notify(70606)
			return md.Status != StatusCancelRequested == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(70607)
			return md.Status != StatusRunning == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(70608)
			return md.Status != StatusPending == true
		}() == true {
			__antithesis_instrumentation__.Notify(70609)
			return fmt.Errorf("job with status %s cannot be reverted", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70610)
		}
		__antithesis_instrumentation__.Notify(70603)
		if md.Status != StatusReverting {
			__antithesis_instrumentation__.Notify(70611)
			if fn != nil {
				__antithesis_instrumentation__.Notify(70614)
				if err := fn(ctx, txn); err != nil {
					__antithesis_instrumentation__.Notify(70615)
					return err
				} else {
					__antithesis_instrumentation__.Notify(70616)
				}
			} else {
				__antithesis_instrumentation__.Notify(70617)
			}
			__antithesis_instrumentation__.Notify(70612)
			if err != nil {
				__antithesis_instrumentation__.Notify(70618)
				md.Payload.Error = err.Error()
				encodedErr := errors.EncodeError(ctx, err)
				md.Payload.FinalResumeError = &encodedErr
				ju.UpdatePayload(md.Payload)
			} else {
				__antithesis_instrumentation__.Notify(70619)
				if md.Payload.FinalResumeError == nil {
					__antithesis_instrumentation__.Notify(70620)
					return errors.AssertionFailedf(
						"tried to mark job as reverting, but no error was provided or recorded")
				} else {
					__antithesis_instrumentation__.Notify(70621)
				}
			}
			__antithesis_instrumentation__.Notify(70613)
			ju.UpdateStatus(StatusReverting)
		} else {
			__antithesis_instrumentation__.Notify(70622)
		}
		__antithesis_instrumentation__.Notify(70604)

		if md.RunStats != nil {
			__antithesis_instrumentation__.Notify(70623)

			numRuns := md.RunStats.NumRuns + 1
			if md.Status != StatusReverting {
				__antithesis_instrumentation__.Notify(70625)

				numRuns = 1
			} else {
				__antithesis_instrumentation__.Notify(70626)
			}
			__antithesis_instrumentation__.Notify(70624)
			ju.UpdateRunStats(numRuns, j.registry.clock.Now().GoTime())
		} else {
			__antithesis_instrumentation__.Notify(70627)
		}
		__antithesis_instrumentation__.Notify(70605)
		return nil
	})
}

func (j *Job) canceled(
	ctx context.Context, txn *kv.Txn, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70628)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70629)
		if md.Status == StatusCanceled {
			__antithesis_instrumentation__.Notify(70633)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(70634)
		}
		__antithesis_instrumentation__.Notify(70630)
		if md.Status != StatusReverting {
			__antithesis_instrumentation__.Notify(70635)
			return fmt.Errorf("job with status %s cannot be requested to be canceled", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70636)
		}
		__antithesis_instrumentation__.Notify(70631)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70637)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70638)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70639)
			}
		} else {
			__antithesis_instrumentation__.Notify(70640)
		}
		__antithesis_instrumentation__.Notify(70632)
		ju.UpdateStatus(StatusCanceled)
		md.Payload.FinishedMicros = timeutil.ToUnixMicros(j.registry.clock.Now().GoTime())
		ju.UpdatePayload(md.Payload)
		return nil
	})
}

func (j *Job) failed(
	ctx context.Context, txn *kv.Txn, err error, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70641)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70642)

		if md.Status.Terminal() {
			__antithesis_instrumentation__.Notify(70646)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(70647)
		}
		__antithesis_instrumentation__.Notify(70643)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70648)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70649)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70650)
			}
		} else {
			__antithesis_instrumentation__.Notify(70651)
		}
		__antithesis_instrumentation__.Notify(70644)

		ju.UpdateStatus(StatusFailed)

		const (
			jobErrMaxRuneCount    = 1024
			jobErrTruncatedMarker = " -- TRUNCATED"
		)
		errStr := err.Error()
		if len(errStr) > jobErrMaxRuneCount {
			__antithesis_instrumentation__.Notify(70652)
			errStr = util.TruncateString(errStr, jobErrMaxRuneCount) + jobErrTruncatedMarker
		} else {
			__antithesis_instrumentation__.Notify(70653)
		}
		__antithesis_instrumentation__.Notify(70645)
		md.Payload.Error = errStr

		md.Payload.FinishedMicros = timeutil.ToUnixMicros(j.registry.clock.Now().GoTime())
		ju.UpdatePayload(md.Payload)
		return nil
	})
}

func (j *Job) revertFailed(
	ctx context.Context, txn *kv.Txn, err error, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70654)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70655)
		if md.Status != StatusReverting {
			__antithesis_instrumentation__.Notify(70658)
			return fmt.Errorf("job with status %s cannot fail during a revert", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70659)
		}
		__antithesis_instrumentation__.Notify(70656)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70660)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70661)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70662)
			}
		} else {
			__antithesis_instrumentation__.Notify(70663)
		}
		__antithesis_instrumentation__.Notify(70657)
		ju.UpdateStatus(StatusRevertFailed)
		md.Payload.FinishedMicros = timeutil.ToUnixMicros(j.registry.clock.Now().GoTime())
		md.Payload.Error = err.Error()
		ju.UpdatePayload(md.Payload)
		return nil
	})
}

func (j *Job) succeeded(
	ctx context.Context, txn *kv.Txn, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70664)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70665)
		if md.Status == StatusSucceeded {
			__antithesis_instrumentation__.Notify(70669)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(70670)
		}
		__antithesis_instrumentation__.Notify(70666)
		if md.Status != StatusRunning && func() bool {
			__antithesis_instrumentation__.Notify(70671)
			return md.Status != StatusPending == true
		}() == true {
			__antithesis_instrumentation__.Notify(70672)
			return errors.Errorf("job with status %s cannot be marked as succeeded", md.Status)
		} else {
			__antithesis_instrumentation__.Notify(70673)
		}
		__antithesis_instrumentation__.Notify(70667)
		if fn != nil {
			__antithesis_instrumentation__.Notify(70674)
			if err := fn(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(70675)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70676)
			}
		} else {
			__antithesis_instrumentation__.Notify(70677)
		}
		__antithesis_instrumentation__.Notify(70668)
		ju.UpdateStatus(StatusSucceeded)
		md.Payload.FinishedMicros = timeutil.ToUnixMicros(j.registry.clock.Now().GoTime())
		ju.UpdatePayload(md.Payload)
		md.Progress.Progress = &jobspb.Progress_FractionCompleted{
			FractionCompleted: 1.0,
		}
		ju.UpdateProgress(md.Progress)
		return nil
	})
}

func (j *Job) SetDetails(ctx context.Context, txn *kv.Txn, details interface{}) error {
	__antithesis_instrumentation__.Notify(70678)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70679)
		if err := md.CheckRunningOrReverting(); err != nil {
			__antithesis_instrumentation__.Notify(70681)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70682)
		}
		__antithesis_instrumentation__.Notify(70680)
		md.Payload.Details = jobspb.WrapPayloadDetails(details)
		ju.UpdatePayload(md.Payload)
		return nil
	})
}

func (j *Job) SetProgress(ctx context.Context, txn *kv.Txn, details interface{}) error {
	__antithesis_instrumentation__.Notify(70683)
	return j.Update(ctx, txn, func(txn *kv.Txn, md JobMetadata, ju *JobUpdater) error {
		__antithesis_instrumentation__.Notify(70684)
		if err := md.CheckRunningOrReverting(); err != nil {
			__antithesis_instrumentation__.Notify(70686)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70687)
		}
		__antithesis_instrumentation__.Notify(70685)
		md.Progress.Details = jobspb.WrapProgressDetails(details)
		ju.UpdateProgress(md.Progress)
		return nil
	})
}

func (j *Job) Payload() jobspb.Payload {
	__antithesis_instrumentation__.Notify(70688)
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.mu.payload
}

func (j *Job) Progress() jobspb.Progress {
	__antithesis_instrumentation__.Notify(70689)
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.mu.progress
}

func (j *Job) Details() jobspb.Details {
	__antithesis_instrumentation__.Notify(70690)
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.mu.payload.UnwrapDetails()
}

func (j *Job) Status() Status {
	__antithesis_instrumentation__.Notify(70691)
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.mu.status
}

func (j *Job) FractionCompleted() float32 {
	__antithesis_instrumentation__.Notify(70692)
	progress := j.Progress()
	return progress.GetFractionCompleted()
}

func (j *Job) MakeSessionBoundInternalExecutor(
	ctx context.Context, sd *sessiondata.SessionData,
) sqlutil.InternalExecutor {
	__antithesis_instrumentation__.Notify(70693)
	return j.registry.sessionBoundInternalExecutorFactory(ctx, sd)
}

func (j *Job) MarkIdle(isIdle bool) {
	__antithesis_instrumentation__.Notify(70694)
	j.registry.MarkIdle(j, isIdle)
}

func (j *Job) runInTxn(
	ctx context.Context, txn *kv.Txn, fn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(70695)
	if txn != nil {
		__antithesis_instrumentation__.Notify(70697)

		return fn(ctx, txn)
	} else {
		__antithesis_instrumentation__.Notify(70698)
	}
	__antithesis_instrumentation__.Notify(70696)
	return j.registry.db.Txn(ctx, fn)
}

type JobNotFoundError struct {
	jobID     jobspb.JobID
	sessionID sqlliveness.SessionID
}

func (e *JobNotFoundError) Error() string {
	__antithesis_instrumentation__.Notify(70699)
	if e.sessionID != "" {
		__antithesis_instrumentation__.Notify(70701)
		return fmt.Sprintf("job with ID %d does not exist with claim session id %q", e.jobID, e.sessionID.String())
	} else {
		__antithesis_instrumentation__.Notify(70702)
	}
	__antithesis_instrumentation__.Notify(70700)
	return fmt.Sprintf("job with ID %d does not exist", e.jobID)
}

func HasJobNotFoundError(err error) bool {
	__antithesis_instrumentation__.Notify(70703)
	return errors.HasType(err, (*JobNotFoundError)(nil))
}

func (j *Job) load(ctx context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(70704)
	var payload *jobspb.Payload
	var progress *jobspb.Progress
	var createdBy *CreatedByInfo
	var status Status

	if err := j.runInTxn(ctx, txn, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(70706)
		const (
			queryNoSessionID   = "SELECT payload, progress, created_by_type, created_by_id, status FROM system.jobs WHERE id = $1"
			queryWithSessionID = queryNoSessionID + " AND claim_session_id = $2"
		)
		sess := sessiondata.InternalExecutorOverride{User: security.RootUserName()}

		var err error
		var row tree.Datums
		if j.session == nil {
			__antithesis_instrumentation__.Notify(70713)
			row, err = j.registry.ex.QueryRowEx(ctx, "load-job-query", txn, sess,
				queryNoSessionID, j.ID())
		} else {
			__antithesis_instrumentation__.Notify(70714)
			row, err = j.registry.ex.QueryRowEx(ctx, "load-job-query", txn, sess,
				queryWithSessionID, j.ID(), j.session.ID().UnsafeBytes())
		}
		__antithesis_instrumentation__.Notify(70707)
		if err != nil {
			__antithesis_instrumentation__.Notify(70715)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70716)
		}
		__antithesis_instrumentation__.Notify(70708)
		if row == nil {
			__antithesis_instrumentation__.Notify(70717)
			return &JobNotFoundError{jobID: j.ID()}
		} else {
			__antithesis_instrumentation__.Notify(70718)
		}
		__antithesis_instrumentation__.Notify(70709)
		payload, err = UnmarshalPayload(row[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(70719)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70720)
		}
		__antithesis_instrumentation__.Notify(70710)
		progress, err = UnmarshalProgress(row[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(70721)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70722)
		}
		__antithesis_instrumentation__.Notify(70711)
		createdBy, err = unmarshalCreatedBy(row[2], row[3])
		if err != nil {
			__antithesis_instrumentation__.Notify(70723)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70724)
		}
		__antithesis_instrumentation__.Notify(70712)
		status, err = unmarshalStatus(row[4])
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(70725)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70726)
	}
	__antithesis_instrumentation__.Notify(70705)
	j.mu.payload = *payload
	j.mu.progress = *progress
	j.mu.status = status
	j.createdBy = createdBy
	return nil
}

func UnmarshalPayload(datum tree.Datum) (*jobspb.Payload, error) {
	__antithesis_instrumentation__.Notify(70727)
	payload := &jobspb.Payload{}
	bytes, ok := datum.(*tree.DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(70730)
		return nil, errors.Errorf(
			"job: failed to unmarshal payload as DBytes (was %T)", datum)
	} else {
		__antithesis_instrumentation__.Notify(70731)
	}
	__antithesis_instrumentation__.Notify(70728)
	if err := protoutil.Unmarshal([]byte(*bytes), payload); err != nil {
		__antithesis_instrumentation__.Notify(70732)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70733)
	}
	__antithesis_instrumentation__.Notify(70729)
	return payload, nil
}

func UnmarshalProgress(datum tree.Datum) (*jobspb.Progress, error) {
	__antithesis_instrumentation__.Notify(70734)
	progress := &jobspb.Progress{}
	bytes, ok := datum.(*tree.DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(70737)
		return nil, errors.Errorf(
			"job: failed to unmarshal Progress as DBytes (was %T)", datum)
	} else {
		__antithesis_instrumentation__.Notify(70738)
	}
	__antithesis_instrumentation__.Notify(70735)
	if err := protoutil.Unmarshal([]byte(*bytes), progress); err != nil {
		__antithesis_instrumentation__.Notify(70739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70740)
	}
	__antithesis_instrumentation__.Notify(70736)
	return progress, nil
}

func unmarshalCreatedBy(createdByType, createdByID tree.Datum) (*CreatedByInfo, error) {
	__antithesis_instrumentation__.Notify(70741)
	if createdByType == tree.DNull || func() bool {
		__antithesis_instrumentation__.Notify(70744)
		return createdByID == tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(70745)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(70746)
	}
	__antithesis_instrumentation__.Notify(70742)
	if ds, ok := createdByType.(*tree.DString); ok {
		__antithesis_instrumentation__.Notify(70747)
		if id, ok := createdByID.(*tree.DInt); ok {
			__antithesis_instrumentation__.Notify(70749)
			return &CreatedByInfo{Name: string(*ds), ID: int64(*id)}, nil
		} else {
			__antithesis_instrumentation__.Notify(70750)
		}
		__antithesis_instrumentation__.Notify(70748)
		return nil, errors.Errorf(
			"job: failed to unmarshal created_by_type as DInt (was %T)", createdByID)
	} else {
		__antithesis_instrumentation__.Notify(70751)
	}
	__antithesis_instrumentation__.Notify(70743)
	return nil, errors.Errorf(
		"job: failed to unmarshal created_by_type as DString (was %T)", createdByType)
}

func unmarshalStatus(datum tree.Datum) (Status, error) {
	__antithesis_instrumentation__.Notify(70752)
	statusString, ok := datum.(*tree.DString)
	if !ok {
		__antithesis_instrumentation__.Notify(70754)
		return "", errors.AssertionFailedf("expected string status, but got %T", datum)
	} else {
		__antithesis_instrumentation__.Notify(70755)
	}
	__antithesis_instrumentation__.Notify(70753)
	return Status(*statusString), nil
}

func (j *Job) getRunStats() (rs RunStats) {
	__antithesis_instrumentation__.Notify(70756)
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.mu.runStats != nil {
		__antithesis_instrumentation__.Notify(70758)
		rs = *j.mu.runStats
	} else {
		__antithesis_instrumentation__.Notify(70759)
	}
	__antithesis_instrumentation__.Notify(70757)
	return rs
}

func (sj *StartableJob) Start(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(70760)
	if alreadyStarted := sj.recordStart(); alreadyStarted {
		__antithesis_instrumentation__.Notify(70766)
		return errors.AssertionFailedf(
			"StartableJob %d cannot be started more than once", sj.ID())
	} else {
		__antithesis_instrumentation__.Notify(70767)
	}
	__antithesis_instrumentation__.Notify(70761)

	if sj.session == nil {
		__antithesis_instrumentation__.Notify(70768)
		return errors.AssertionFailedf(
			"StartableJob %d cannot be started without sqlliveness session", sj.ID())
	} else {
		__antithesis_instrumentation__.Notify(70769)
	}
	__antithesis_instrumentation__.Notify(70762)

	defer func() {
		__antithesis_instrumentation__.Notify(70770)
		if err != nil {
			__antithesis_instrumentation__.Notify(70771)
			sj.registry.unregister(sj.ID())
		} else {
			__antithesis_instrumentation__.Notify(70772)
		}
	}()
	__antithesis_instrumentation__.Notify(70763)
	if !sj.txn.IsCommitted() {
		__antithesis_instrumentation__.Notify(70773)
		return fmt.Errorf("cannot resume %T job which is not committed", sj.resumer)
	} else {
		__antithesis_instrumentation__.Notify(70774)
	}
	__antithesis_instrumentation__.Notify(70764)

	if err := sj.registry.stopper.RunAsyncTask(ctx, sj.taskName(), func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(70775)
		sj.execErr = sj.registry.runJob(sj.resumerCtx, sj.resumer, sj.Job, StatusRunning, sj.taskName())
		close(sj.execDone)
	}); err != nil {
		__antithesis_instrumentation__.Notify(70776)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70777)
	}
	__antithesis_instrumentation__.Notify(70765)

	return nil
}

func (sj *StartableJob) AwaitCompletion(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(70778)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(70779)
		return ctx.Err()
	case <-sj.execDone:
		__antithesis_instrumentation__.Notify(70780)
		return sj.execErr
	}
}

func (sj *StartableJob) ReportExecutionResults(
	ctx context.Context, resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(70781)
	if reporter, ok := sj.resumer.(JobResultsReporter); ok {
		__antithesis_instrumentation__.Notify(70783)
		return reporter.ReportResults(ctx, resultsCh)
	} else {
		__antithesis_instrumentation__.Notify(70784)
	}
	__antithesis_instrumentation__.Notify(70782)
	return errors.AssertionFailedf("job does not produce results")
}

func (sj *StartableJob) CleanupOnRollback(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(70785)

	if sj.txn.IsCommitted() && func() bool {
		__antithesis_instrumentation__.Notify(70789)
		return ctx.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70790)
		return errors.AssertionFailedf(
			"cannot call CleanupOnRollback for a StartableJob created by a committed transaction")
	} else {
		__antithesis_instrumentation__.Notify(70791)
	}
	__antithesis_instrumentation__.Notify(70786)

	sj.registry.unregister(sj.ID())
	if sj.cancel != nil {
		__antithesis_instrumentation__.Notify(70792)
		sj.cancel()
	} else {
		__antithesis_instrumentation__.Notify(70793)
	}
	__antithesis_instrumentation__.Notify(70787)

	if !sj.txn.Sender().TxnStatus().IsFinalized() && func() bool {
		__antithesis_instrumentation__.Notify(70794)
		return ctx.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70795)
		return errors.New(
			"cannot call CleanupOnRollback for a StartableJob with a non-finalized transaction")
	} else {
		__antithesis_instrumentation__.Notify(70796)
	}
	__antithesis_instrumentation__.Notify(70788)
	return nil
}

func (sj *StartableJob) Cancel(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(70797)
	alreadyStarted := sj.recordStart()
	defer func() {
		__antithesis_instrumentation__.Notify(70799)
		if alreadyStarted {
			__antithesis_instrumentation__.Notify(70800)
			sj.registry.cancelRegisteredJobContext(sj.ID())
		} else {
			__antithesis_instrumentation__.Notify(70801)
			sj.registry.unregister(sj.ID())
		}
	}()
	__antithesis_instrumentation__.Notify(70798)
	return sj.registry.CancelRequested(ctx, nil, sj.ID())
}

func (sj *StartableJob) recordStart() (alreadyStarted bool) {
	__antithesis_instrumentation__.Notify(70802)
	return atomic.AddInt64(&sj.starts, 1) != 1
}

func ParseRetriableExecutionErrorLogFromJSON(
	log []byte,
) ([]*jobspb.RetriableExecutionFailure, error) {
	__antithesis_instrumentation__.Notify(70803)
	var jsonArr []gojson.RawMessage
	if err := gojson.Unmarshal(log, &jsonArr); err != nil {
		__antithesis_instrumentation__.Notify(70806)
		return nil, errors.Wrap(err, "failed to decode json array for execution log")
	} else {
		__antithesis_instrumentation__.Notify(70807)
	}
	__antithesis_instrumentation__.Notify(70804)
	ret := make([]*jobspb.RetriableExecutionFailure, len(jsonArr))

	json := jsonpb.Unmarshaler{AllowUnknownFields: true}
	var reader bytes.Reader
	for i, data := range jsonArr {
		__antithesis_instrumentation__.Notify(70808)
		msgI, err := protoreflect.NewMessage("cockroach.sql.jobs.jobspb.RetriableExecutionFailure")
		if err != nil {
			__antithesis_instrumentation__.Notify(70811)
			return nil, errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(70812)
		}
		__antithesis_instrumentation__.Notify(70809)
		msg := msgI.(*jobspb.RetriableExecutionFailure)
		reader.Reset(data)
		if err := json.Unmarshal(&reader, msg); err != nil {
			__antithesis_instrumentation__.Notify(70813)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(70814)
		}
		__antithesis_instrumentation__.Notify(70810)
		ret[i] = msg
	}
	__antithesis_instrumentation__.Notify(70805)
	return ret, nil
}

func FormatRetriableExecutionErrorLogToJSON(
	ctx context.Context, log []*jobspb.RetriableExecutionFailure,
) (*tree.DJSON, error) {
	__antithesis_instrumentation__.Notify(70815)
	ab := json.NewArrayBuilder(len(log))
	for i := range log {
		__antithesis_instrumentation__.Notify(70817)
		ev := *log[i]
		if ev.Error != nil {
			__antithesis_instrumentation__.Notify(70820)
			ev.TruncatedError = errors.DecodeError(ctx, *ev.Error).Error()
			ev.Error = nil
		} else {
			__antithesis_instrumentation__.Notify(70821)
		}
		__antithesis_instrumentation__.Notify(70818)
		msg, err := protoreflect.MessageToJSON(&ev, protoreflect.FmtFlags{
			EmitDefaults: false,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(70822)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(70823)
		}
		__antithesis_instrumentation__.Notify(70819)
		ab.Add(msg)
	}
	__antithesis_instrumentation__.Notify(70816)
	return tree.NewDJSON(ab.Build()), nil
}

func FormatRetriableExecutionErrorLogToStringArray(
	ctx context.Context, log []*jobspb.RetriableExecutionFailure,
) *tree.DArray {
	__antithesis_instrumentation__.Notify(70824)
	arr := tree.NewDArray(types.String)
	for _, ev := range log {
		__antithesis_instrumentation__.Notify(70826)
		if ev == nil {
			__antithesis_instrumentation__.Notify(70829)
			continue
		} else {
			__antithesis_instrumentation__.Notify(70830)
		}
		__antithesis_instrumentation__.Notify(70827)
		var cause error
		if ev.Error != nil {
			__antithesis_instrumentation__.Notify(70831)
			cause = errors.DecodeError(ctx, *ev.Error)
		} else {
			__antithesis_instrumentation__.Notify(70832)
			cause = fmt.Errorf("(truncated) %s", ev.TruncatedError)
		}
		__antithesis_instrumentation__.Notify(70828)
		msg := formatRetriableExecutionFailure(
			ev.InstanceID,
			Status(ev.Status),
			timeutil.FromUnixMicros(ev.ExecutionStartMicros),
			timeutil.FromUnixMicros(ev.ExecutionEndMicros),
			cause,
		)

		_ = arr.Append(tree.NewDString(msg))
	}
	__antithesis_instrumentation__.Notify(70825)
	return arr
}
