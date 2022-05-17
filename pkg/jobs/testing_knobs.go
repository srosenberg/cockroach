package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type TestingKnobs struct {
	SchedulerDaemonInitialScanDelay func() time.Duration

	SchedulerDaemonScanDelay func() time.Duration

	JobSchedulerEnv scheduledjobs.JobSchedulerEnv

	TakeOverJobsScheduling func(func(ctx context.Context, maxSchedules int64) error)

	CaptureJobScheduler func(scheduler interface{})

	CaptureJobExecutionConfig func(config *scheduledjobs.JobExecutionConfig)

	OverrideAsOfClause func(clause *tree.AsOfClause)

	BeforeUpdate func(orig, updated JobMetadata) error

	IntervalOverrides TestingIntervalOverrides

	AfterJobStateMachine func()

	TimeSource *hlc.Clock

	DisableAdoptions bool
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(84946) }

type TestingIntervalOverrides struct {
	Adopt *time.Duration

	Cancel *time.Duration

	Gc *time.Duration

	RetentionTime *time.Duration

	RetryInitialDelay *time.Duration

	RetryMaxDelay *time.Duration
}

func NewTestingKnobsWithShortIntervals() *TestingKnobs {
	__antithesis_instrumentation__.Notify(84947)
	defaultShortInterval := 10 * time.Millisecond
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(84949)
		defaultShortInterval *= 5
	} else {
		__antithesis_instrumentation__.Notify(84950)
	}
	__antithesis_instrumentation__.Notify(84948)
	return NewTestingKnobsWithIntervals(
		defaultShortInterval, defaultShortInterval, defaultShortInterval, defaultShortInterval,
	)
}

func NewTestingKnobsWithIntervals(
	adopt, cancel, initialDelay, maxDelay time.Duration,
) *TestingKnobs {
	__antithesis_instrumentation__.Notify(84951)
	return &TestingKnobs{
		IntervalOverrides: TestingIntervalOverrides{
			Adopt:             &adopt,
			Cancel:            &cancel,
			RetryInitialDelay: &initialDelay,
			RetryMaxDelay:     &maxDelay,
		},
	}
}
