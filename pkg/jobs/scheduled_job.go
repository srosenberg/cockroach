package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/robfig/cron/v3"
)

type scheduledJobRecord struct {
	ScheduleID      int64                     `col:"schedule_id"`
	ScheduleLabel   string                    `col:"schedule_name"`
	Owner           security.SQLUsername      `col:"owner"`
	NextRun         time.Time                 `col:"next_run"`
	ScheduleState   jobspb.ScheduleState      `col:"schedule_state"`
	ScheduleExpr    string                    `col:"schedule_expr"`
	ScheduleDetails jobspb.ScheduleDetails    `col:"schedule_details"`
	ExecutorType    string                    `col:"executor_type"`
	ExecutionArgs   jobspb.ExecutionArguments `col:"execution_args"`
}

const InvalidScheduleID int64 = 0

type ScheduledJob struct {
	env scheduledjobs.JobSchedulerEnv

	rec scheduledJobRecord

	scheduledTime time.Time

	dirty map[string]struct{}
}

func NewScheduledJob(env scheduledjobs.JobSchedulerEnv) *ScheduledJob {
	__antithesis_instrumentation__.Notify(84714)
	return &ScheduledJob{
		env:   env,
		dirty: make(map[string]struct{}),
	}
}

type scheduledJobNotFoundError struct {
	scheduleID int64
}

func (e *scheduledJobNotFoundError) Error() string {
	__antithesis_instrumentation__.Notify(84715)
	return fmt.Sprintf("scheduled job with ID %d does not exist", e.scheduleID)
}

func HasScheduledJobNotFoundError(err error) bool {
	__antithesis_instrumentation__.Notify(84716)
	return errors.HasType(err, (*scheduledJobNotFoundError)(nil))
}

func LoadScheduledJob(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	id int64,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) (*ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(84717)
	row, cols, err := ex.QueryRowExWithCols(ctx, "lookup-schedule", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("SELECT * FROM %s WHERE schedule_id = %d",
			env.ScheduledJobsTableName(), id))

	if err != nil {
		__antithesis_instrumentation__.Notify(84721)
		return nil, errors.CombineErrors(err, &scheduledJobNotFoundError{scheduleID: id})
	} else {
		__antithesis_instrumentation__.Notify(84722)
	}
	__antithesis_instrumentation__.Notify(84718)
	if row == nil {
		__antithesis_instrumentation__.Notify(84723)
		return nil, &scheduledJobNotFoundError{scheduleID: id}
	} else {
		__antithesis_instrumentation__.Notify(84724)
	}
	__antithesis_instrumentation__.Notify(84719)

	j := NewScheduledJob(env)
	if err := j.InitFromDatums(row, cols); err != nil {
		__antithesis_instrumentation__.Notify(84725)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84726)
	}
	__antithesis_instrumentation__.Notify(84720)
	return j, nil
}

func (j *ScheduledJob) ScheduleID() int64 {
	__antithesis_instrumentation__.Notify(84727)
	return j.rec.ScheduleID
}

func (j *ScheduledJob) ScheduleLabel() string {
	__antithesis_instrumentation__.Notify(84728)
	return j.rec.ScheduleLabel
}

func (j *ScheduledJob) SetScheduleLabel(label string) {
	__antithesis_instrumentation__.Notify(84729)
	j.rec.ScheduleLabel = label
	j.markDirty("schedule_name")
}

func (j *ScheduledJob) Owner() security.SQLUsername {
	__antithesis_instrumentation__.Notify(84730)
	return j.rec.Owner
}

func (j *ScheduledJob) SetOwner(owner security.SQLUsername) {
	__antithesis_instrumentation__.Notify(84731)
	j.rec.Owner = owner
	j.markDirty("owner")
}

func (j *ScheduledJob) NextRun() time.Time {
	__antithesis_instrumentation__.Notify(84732)
	return j.rec.NextRun
}

func (j *ScheduledJob) ScheduledRunTime() time.Time {
	__antithesis_instrumentation__.Notify(84733)
	return j.scheduledTime
}

func (j *ScheduledJob) IsPaused() bool {
	__antithesis_instrumentation__.Notify(84734)
	return j.rec.NextRun == time.Time{}
}

func (j *ScheduledJob) ExecutorType() string {
	__antithesis_instrumentation__.Notify(84735)
	return j.rec.ExecutorType
}

func (j *ScheduledJob) ExecutionArgs() *jobspb.ExecutionArguments {
	__antithesis_instrumentation__.Notify(84736)
	return &j.rec.ExecutionArgs
}

func (j *ScheduledJob) SetSchedule(scheduleExpr string) error {
	__antithesis_instrumentation__.Notify(84737)
	j.rec.ScheduleExpr = scheduleExpr
	j.markDirty("schedule_expr")
	return j.ScheduleNextRun()
}

func (j *ScheduledJob) HasRecurringSchedule() bool {
	__antithesis_instrumentation__.Notify(84738)
	return len(j.rec.ScheduleExpr) > 0
}

func (j *ScheduledJob) Frequency() (time.Duration, error) {
	__antithesis_instrumentation__.Notify(84739)
	if !j.HasRecurringSchedule() {
		__antithesis_instrumentation__.Notify(84742)
		return 0, errors.Newf(
			"schedule %d is not periodic", j.rec.ScheduleID)
	} else {
		__antithesis_instrumentation__.Notify(84743)
	}
	__antithesis_instrumentation__.Notify(84740)
	expr, err := cron.ParseStandard(j.rec.ScheduleExpr)
	if err != nil {
		__antithesis_instrumentation__.Notify(84744)
		return 0, errors.Wrapf(err,
			"parsing schedule expression: %q; it must be a valid cron expression",
			j.rec.ScheduleExpr)
	} else {
		__antithesis_instrumentation__.Notify(84745)
	}
	__antithesis_instrumentation__.Notify(84741)

	next := expr.Next(j.env.Now())
	nextNext := expr.Next(next)
	return nextNext.Sub(next), nil
}

func (j *ScheduledJob) ScheduleNextRun() error {
	__antithesis_instrumentation__.Notify(84746)
	if !j.HasRecurringSchedule() {
		__antithesis_instrumentation__.Notify(84749)
		return errors.Newf(
			"cannot set next run for schedule %d (empty schedule)", j.rec.ScheduleID)
	} else {
		__antithesis_instrumentation__.Notify(84750)
	}
	__antithesis_instrumentation__.Notify(84747)
	expr, err := cron.ParseStandard(j.rec.ScheduleExpr)
	if err != nil {
		__antithesis_instrumentation__.Notify(84751)
		return errors.Wrapf(err, "parsing schedule expression: %q", j.rec.ScheduleExpr)
	} else {
		__antithesis_instrumentation__.Notify(84752)
	}
	__antithesis_instrumentation__.Notify(84748)
	j.SetNextRun(expr.Next(j.env.Now()))
	return nil
}

func (j *ScheduledJob) SetNextRun(t time.Time) {
	__antithesis_instrumentation__.Notify(84753)
	j.rec.NextRun = t
	j.markDirty("next_run")
}

func (j *ScheduledJob) ScheduleDetails() *jobspb.ScheduleDetails {
	__antithesis_instrumentation__.Notify(84754)
	return &j.rec.ScheduleDetails
}

func (j *ScheduledJob) SetScheduleDetails(details jobspb.ScheduleDetails) {
	__antithesis_instrumentation__.Notify(84755)
	j.rec.ScheduleDetails = details
	j.markDirty("schedule_details")
}

func (j *ScheduledJob) SetScheduleStatus(fmtOrMsg string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(84756)
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(84758)
		j.rec.ScheduleState.Status = fmtOrMsg
	} else {
		__antithesis_instrumentation__.Notify(84759)
		j.rec.ScheduleState.Status = fmt.Sprintf(fmtOrMsg, args...)
	}
	__antithesis_instrumentation__.Notify(84757)
	j.markDirty("schedule_state")
}

func (j *ScheduledJob) ScheduleStatus() string {
	__antithesis_instrumentation__.Notify(84760)
	return j.rec.ScheduleState.Status
}

func (j *ScheduledJob) ClearScheduleStatus() {
	__antithesis_instrumentation__.Notify(84761)
	j.rec.ScheduleState.Status = ""
	j.markDirty("schedule_state")
}

func (j *ScheduledJob) ScheduleExpr() string {
	__antithesis_instrumentation__.Notify(84762)
	return j.rec.ScheduleExpr
}

func (j *ScheduledJob) Pause() {
	__antithesis_instrumentation__.Notify(84763)
	j.rec.NextRun = time.Time{}
	j.markDirty("next_run")
}

func (j *ScheduledJob) SetExecutionDetails(executor string, args jobspb.ExecutionArguments) {
	__antithesis_instrumentation__.Notify(84764)
	j.rec.ExecutorType = executor
	j.rec.ExecutionArgs = args
	j.markDirty("executor_type", "execution_args")
}

func (j *ScheduledJob) ClearDirty() {
	__antithesis_instrumentation__.Notify(84765)
	j.dirty = make(map[string]struct{})
}

func (j *ScheduledJob) InitFromDatums(datums []tree.Datum, cols []colinfo.ResultColumn) error {
	__antithesis_instrumentation__.Notify(84766)
	if len(datums) != len(cols) {
		__antithesis_instrumentation__.Notify(84770)
		return errors.Errorf(
			"datums length != columns length: %d != %d", len(datums), len(cols))
	} else {
		__antithesis_instrumentation__.Notify(84771)
	}
	__antithesis_instrumentation__.Notify(84767)

	record := reflect.ValueOf(&j.rec).Elem()

	numInitialized := 0
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(84772)
		native, err := datumToNative(datums[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(84777)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84778)
		}
		__antithesis_instrumentation__.Notify(84773)

		if native == nil {
			__antithesis_instrumentation__.Notify(84779)
			continue
		} else {
			__antithesis_instrumentation__.Notify(84780)
		}
		__antithesis_instrumentation__.Notify(84774)

		fieldNum, ok := columnNameToField[col.Name]
		if !ok {
			__antithesis_instrumentation__.Notify(84781)

			continue
		} else {
			__antithesis_instrumentation__.Notify(84782)
		}
		__antithesis_instrumentation__.Notify(84775)

		field := record.Field(fieldNum)

		if data, ok := native.([]byte); ok {
			__antithesis_instrumentation__.Notify(84783)

			if pb, ok := field.Addr().Interface().(protoutil.Message); ok {
				__antithesis_instrumentation__.Notify(84784)
				if err := protoutil.Unmarshal(data, pb); err != nil {
					__antithesis_instrumentation__.Notify(84785)
					return err
				} else {
					__antithesis_instrumentation__.Notify(84786)
				}
			} else {
				__antithesis_instrumentation__.Notify(84787)
				return errors.Newf(
					"field %s with value of type %T is does not appear to be a protocol message",
					field.String(), field.Addr().Interface())
			}
		} else {
			__antithesis_instrumentation__.Notify(84788)

			rv := reflect.ValueOf(native)
			if !rv.Type().AssignableTo(field.Type()) {
				__antithesis_instrumentation__.Notify(84790)

				ok := false
				if col.Name == "owner" {
					__antithesis_instrumentation__.Notify(84792)

					var s string
					s, ok = native.(string)
					if ok {
						__antithesis_instrumentation__.Notify(84793)

						rv = reflect.ValueOf(security.MakeSQLUsernameFromPreNormalizedString(s))
					} else {
						__antithesis_instrumentation__.Notify(84794)
					}
				} else {
					__antithesis_instrumentation__.Notify(84795)
				}
				__antithesis_instrumentation__.Notify(84791)
				if !ok {
					__antithesis_instrumentation__.Notify(84796)
					return errors.Newf("value of type %T cannot be assigned to %s",
						native, field.Type().String())
				} else {
					__antithesis_instrumentation__.Notify(84797)
				}
			} else {
				__antithesis_instrumentation__.Notify(84798)
			}
			__antithesis_instrumentation__.Notify(84789)
			field.Set(rv)
		}
		__antithesis_instrumentation__.Notify(84776)
		numInitialized++
	}
	__antithesis_instrumentation__.Notify(84768)

	if numInitialized == 0 {
		__antithesis_instrumentation__.Notify(84799)
		return errors.New("did not initialize any schedule field")
	} else {
		__antithesis_instrumentation__.Notify(84800)
	}
	__antithesis_instrumentation__.Notify(84769)

	j.scheduledTime = j.rec.NextRun
	return nil
}

func (j *ScheduledJob) Create(ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(84801)
	if j.rec.ScheduleID != 0 {
		__antithesis_instrumentation__.Notify(84807)
		return errors.New("cannot specify schedule id when creating new cron job")
	} else {
		__antithesis_instrumentation__.Notify(84808)
	}
	__antithesis_instrumentation__.Notify(84802)

	if !j.isDirty() {
		__antithesis_instrumentation__.Notify(84809)
		return errors.New("no settings specified for scheduled job")
	} else {
		__antithesis_instrumentation__.Notify(84810)
	}
	__antithesis_instrumentation__.Notify(84803)

	cols, qargs, err := j.marshalChanges()
	if err != nil {
		__antithesis_instrumentation__.Notify(84811)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84812)
	}
	__antithesis_instrumentation__.Notify(84804)

	row, retCols, err := ex.QueryRowExWithCols(ctx, "sched-create", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s) RETURNING schedule_id",
			j.env.ScheduledJobsTableName(), strings.Join(cols, ","), generatePlaceholders(len(qargs))),
		qargs...,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(84813)
		return errors.Wrapf(err, "failed to create new schedule")
	} else {
		__antithesis_instrumentation__.Notify(84814)
	}
	__antithesis_instrumentation__.Notify(84805)
	if row == nil {
		__antithesis_instrumentation__.Notify(84815)
		return errors.New("failed to create new schedule")
	} else {
		__antithesis_instrumentation__.Notify(84816)
	}
	__antithesis_instrumentation__.Notify(84806)

	return j.InitFromDatums(row, retCols)
}

func (j *ScheduledJob) Update(ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(84817)
	if !j.isDirty() {
		__antithesis_instrumentation__.Notify(84824)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(84825)
	}
	__antithesis_instrumentation__.Notify(84818)

	if j.rec.ScheduleID == 0 {
		__antithesis_instrumentation__.Notify(84826)
		return errors.New("cannot update schedule: missing schedule id")
	} else {
		__antithesis_instrumentation__.Notify(84827)
	}
	__antithesis_instrumentation__.Notify(84819)

	cols, qargs, err := j.marshalChanges()
	if err != nil {
		__antithesis_instrumentation__.Notify(84828)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84829)
	}
	__antithesis_instrumentation__.Notify(84820)

	if len(qargs) == 0 {
		__antithesis_instrumentation__.Notify(84830)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(84831)
	}
	__antithesis_instrumentation__.Notify(84821)

	n, err := ex.ExecEx(ctx, "sched-update", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("UPDATE %s SET (%s) = (%s) WHERE schedule_id = %d",
			j.env.ScheduledJobsTableName(), strings.Join(cols, ","),
			generatePlaceholders(len(qargs)), j.ScheduleID()),
		qargs...,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(84832)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84833)
	}
	__antithesis_instrumentation__.Notify(84822)

	if n != 1 {
		__antithesis_instrumentation__.Notify(84834)
		return fmt.Errorf("expected to update 1 schedule, updated %d instead", n)
	} else {
		__antithesis_instrumentation__.Notify(84835)
	}
	__antithesis_instrumentation__.Notify(84823)

	return nil
}

func (j *ScheduledJob) marshalChanges() ([]string, []interface{}, error) {
	__antithesis_instrumentation__.Notify(84836)
	var cols []string
	var qargs []interface{}

	for col := range j.dirty {
		__antithesis_instrumentation__.Notify(84838)
		var arg tree.Datum
		var err error

		switch col {
		case `schedule_name`:
			__antithesis_instrumentation__.Notify(84841)
			arg = tree.NewDString(j.rec.ScheduleLabel)
		case `owner`:
			__antithesis_instrumentation__.Notify(84842)
			arg = tree.NewDString(j.rec.Owner.Normalized())
		case `next_run`:
			__antithesis_instrumentation__.Notify(84843)
			if (j.rec.NextRun == time.Time{}) {
				__antithesis_instrumentation__.Notify(84850)
				arg = tree.DNull
			} else {
				__antithesis_instrumentation__.Notify(84851)
				arg, err = tree.MakeDTimestampTZ(j.rec.NextRun, time.Microsecond)
			}
		case `schedule_state`:
			__antithesis_instrumentation__.Notify(84844)
			arg, err = marshalProto(&j.rec.ScheduleState)
		case `schedule_expr`:
			__antithesis_instrumentation__.Notify(84845)
			arg = tree.NewDString(j.rec.ScheduleExpr)
		case `schedule_details`:
			__antithesis_instrumentation__.Notify(84846)
			arg, err = marshalProto(&j.rec.ScheduleDetails)
		case `executor_type`:
			__antithesis_instrumentation__.Notify(84847)
			arg = tree.NewDString(j.rec.ExecutorType)
		case `execution_args`:
			__antithesis_instrumentation__.Notify(84848)
			arg, err = marshalProto(&j.rec.ExecutionArgs)
		default:
			__antithesis_instrumentation__.Notify(84849)
			return nil, nil, errors.Newf("cannot marshal column %q", col)
		}
		__antithesis_instrumentation__.Notify(84839)

		if err != nil {
			__antithesis_instrumentation__.Notify(84852)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(84853)
		}
		__antithesis_instrumentation__.Notify(84840)
		cols = append(cols, col)
		qargs = append(qargs, arg)
	}
	__antithesis_instrumentation__.Notify(84837)

	j.dirty = make(map[string]struct{})
	return cols, qargs, nil
}

func (j *ScheduledJob) markDirty(cols ...string) {
	__antithesis_instrumentation__.Notify(84854)
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(84855)
		j.dirty[col] = struct{}{}
	}
}

func (j *ScheduledJob) isDirty() bool {
	__antithesis_instrumentation__.Notify(84856)
	return len(j.dirty) > 0
}

func generatePlaceholders(n int) string {
	__antithesis_instrumentation__.Notify(84857)
	placeholders := strings.Builder{}
	for i := 1; i <= n; i++ {
		__antithesis_instrumentation__.Notify(84859)
		if i > 1 {
			__antithesis_instrumentation__.Notify(84861)
			placeholders.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(84862)
		}
		__antithesis_instrumentation__.Notify(84860)
		placeholders.WriteString(fmt.Sprintf("$%d", i))
	}
	__antithesis_instrumentation__.Notify(84858)
	return placeholders.String()
}

func marshalProto(message protoutil.Message) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(84863)
	data := make([]byte, message.Size())
	if _, err := message.MarshalTo(data); err != nil {
		__antithesis_instrumentation__.Notify(84865)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84866)
	}
	__antithesis_instrumentation__.Notify(84864)
	return tree.NewDBytes(tree.DBytes(data)), nil
}

func datumToNative(datum tree.Datum) (interface{}, error) {
	__antithesis_instrumentation__.Notify(84867)
	datum = tree.UnwrapDatum(nil, datum)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(84870)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(84871)
	}
	__antithesis_instrumentation__.Notify(84868)
	switch d := datum.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(84872)
		return string(*d), nil
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(84873)
		return int64(*d), nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(84874)
		return d.Time, nil
	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(84875)
		return []byte(*d), nil
	}
	__antithesis_instrumentation__.Notify(84869)
	return nil, errors.Newf("cannot handle type %T", datum)
}

var columnNameToField = make(map[string]int)

func init() {

	j := reflect.TypeOf(scheduledJobRecord{})

	for f := 0; f < j.NumField(); f++ {
		field := j.Field(f)
		col := field.Tag.Get("col")
		if col != "" {
			columnNameToField[col] = f
		}
	}
}
