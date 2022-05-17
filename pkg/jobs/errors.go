package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var errRetryJobSentinel = errors.New("retriable job error")

func MarkAsRetryJobError(err error) error {
	__antithesis_instrumentation__.Notify(70214)
	return errors.Mark(err, errRetryJobSentinel)
}

var errJobPermanentSentinel = errors.New("permanent job error")

func MarkAsPermanentJobError(err error) error {
	__antithesis_instrumentation__.Notify(70215)
	return errors.Mark(err, errJobPermanentSentinel)
}

func IsPermanentJobError(err error) bool {
	__antithesis_instrumentation__.Notify(70216)
	return errors.Is(err, errJobPermanentSentinel)
}

func IsPauseSelfError(err error) bool {
	__antithesis_instrumentation__.Notify(70217)
	return errors.Is(err, errPauseSelfSentinel)
}

var errPauseSelfSentinel = errors.New("job requested it be paused")

func MarkPauseRequestError(reason error) error {
	__antithesis_instrumentation__.Notify(70218)
	return errors.Mark(reason, errPauseSelfSentinel)
}

const PauseRequestExplained = "pausing due to error; use RESUME JOB to try to proceed once the issue is resolved, or CANCEL JOB to rollback"

type InvalidStatusError struct {
	id     jobspb.JobID
	status Status
	op     string
	err    string
}

func (e *InvalidStatusError) Error() string {
	__antithesis_instrumentation__.Notify(70219)
	if e.err != "" {
		__antithesis_instrumentation__.Notify(70221)
		return fmt.Sprintf("cannot %s %s job (id %d, err: %q)", e.op, e.status, e.id, e.err)
	} else {
		__antithesis_instrumentation__.Notify(70222)
	}
	__antithesis_instrumentation__.Notify(70220)
	return fmt.Sprintf("cannot %s %s job (id %d)", e.op, e.status, e.id)
}

func SimplifyInvalidStatusError(err error) error {
	__antithesis_instrumentation__.Notify(70223)
	if ierr := (*InvalidStatusError)(nil); errors.As(err, &ierr) {
		__antithesis_instrumentation__.Notify(70225)
		return errors.Errorf("job %s", ierr.status)
	} else {
		__antithesis_instrumentation__.Notify(70226)
	}
	__antithesis_instrumentation__.Notify(70224)
	return err
}

type retriableExecutionError struct {
	instanceID base.SQLInstanceID
	start, end time.Time
	status     Status
	cause      error
}

func newRetriableExecutionError(
	instanceID base.SQLInstanceID, status Status, start, end time.Time, cause error,
) *retriableExecutionError {
	__antithesis_instrumentation__.Notify(70227)
	return &retriableExecutionError{
		instanceID: instanceID,
		status:     status,
		start:      start,
		end:        end,
		cause:      cause,
	}
}

func (e *retriableExecutionError) Error() string {
	__antithesis_instrumentation__.Notify(70228)
	return formatRetriableExecutionFailure(
		e.instanceID, e.status, e.start, e.end, e.cause,
	)

}

func formatRetriableExecutionFailure(
	instanceID base.SQLInstanceID, status Status, start, end time.Time, cause error,
) string {
	__antithesis_instrumentation__.Notify(70229)
	mustTimestamp := func(ts time.Time) *tree.DTimestamp {
		__antithesis_instrumentation__.Notify(70231)
		ret, _ := tree.MakeDTimestamp(ts, time.Microsecond)
		return ret
	}
	__antithesis_instrumentation__.Notify(70230)
	return fmt.Sprintf(
		"%s execution from %v to %v on %d failed: %v",
		status,
		mustTimestamp(start),
		mustTimestamp(end),
		instanceID,
		cause,
	)
}

func (e *retriableExecutionError) Cause() error {
	__antithesis_instrumentation__.Notify(70232)
	return e.cause
}

func (e *retriableExecutionError) Unwrap() error {
	__antithesis_instrumentation__.Notify(70233)
	return e.cause
}

func (e *retriableExecutionError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(70234)
	errors.FormatError(e, s, verb)
}

func (e *retriableExecutionError) SafeFormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(70235)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(70237)
		p.Printf("retriable execution error")
	} else {
		__antithesis_instrumentation__.Notify(70238)
	}
	__antithesis_instrumentation__.Notify(70236)
	return e.cause
}

func (e *retriableExecutionError) toRetriableExecutionFailure(
	ctx context.Context, maxErrorSize int,
) *jobspb.RetriableExecutionFailure {
	__antithesis_instrumentation__.Notify(70239)

	ef := &jobspb.RetriableExecutionFailure{
		Status:               string(e.status),
		ExecutionStartMicros: timeutil.ToUnixMicros(e.start),
		ExecutionEndMicros:   timeutil.ToUnixMicros(e.end),
		InstanceID:           e.instanceID,
	}
	if encodedCause := errors.EncodeError(ctx, e.cause); encodedCause.Size() < maxErrorSize {
		__antithesis_instrumentation__.Notify(70241)
		ef.Error = &encodedCause
	} else {
		__antithesis_instrumentation__.Notify(70242)
		formatted := e.cause.Error()
		if len(formatted) > maxErrorSize {
			__antithesis_instrumentation__.Notify(70244)
			formatted = formatted[:maxErrorSize]
		} else {
			__antithesis_instrumentation__.Notify(70245)
		}
		__antithesis_instrumentation__.Notify(70243)
		ef.TruncatedError = formatted
	}
	__antithesis_instrumentation__.Notify(70240)
	return ef
}
