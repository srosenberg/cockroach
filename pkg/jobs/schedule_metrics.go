package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type ExecutorMetrics struct {
	NumStarted   *metric.Counter
	NumSucceeded *metric.Counter
	NumFailed    *metric.Counter
}

var _ metric.Struct = &ExecutorMetrics{}

func (m *ExecutorMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(84710) }

type SchedulerMetrics struct {
	NumStarted *metric.Gauge

	RescheduleSkip *metric.Gauge

	RescheduleWait *metric.Gauge

	NumErrSchedules *metric.Gauge

	NumMalformedSchedules *metric.Gauge
}

func MakeSchedulerMetrics() SchedulerMetrics {
	__antithesis_instrumentation__.Notify(84711)
	return SchedulerMetrics{
		NumStarted: metric.NewGauge(metric.Metadata{
			Name:        "schedules.round.jobs-started",
			Help:        "The number of jobs started",
			Measurement: "Jobs",
			Unit:        metric.Unit_COUNT,
		}),

		RescheduleSkip: metric.NewGauge(metric.Metadata{
			Name:        "schedules.round.reschedule-skip",
			Help:        "The number of schedules rescheduled due to SKIP policy",
			Measurement: "Schedules",
			Unit:        metric.Unit_COUNT,
		}),

		RescheduleWait: metric.NewGauge(metric.Metadata{
			Name:        "schedules.round.reschedule-wait",
			Help:        "The number of schedules rescheduled due to WAIT policy",
			Measurement: "Schedules",
			Unit:        metric.Unit_COUNT,
		}),

		NumErrSchedules: metric.NewGauge(metric.Metadata{
			Name:        "schedules.error",
			Help:        "Number of schedules which did not execute successfully",
			Measurement: "Schedules",
			Unit:        metric.Unit_COUNT,
		}),

		NumMalformedSchedules: metric.NewGauge(metric.Metadata{
			Name:        "schedules.malformed",
			Help:        "Number of malformed schedules",
			Measurement: "Schedules",
			Unit:        metric.Unit_COUNT,
		}),
	}
}

func (m *SchedulerMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(84712) }

var _ metric.Struct = &SchedulerMetrics{}

func MakeExecutorMetrics(name string) ExecutorMetrics {
	__antithesis_instrumentation__.Notify(84713)
	return ExecutorMetrics{
		NumStarted: metric.NewCounter(metric.Metadata{
			Name:        fmt.Sprintf("schedules.%s.started", name),
			Help:        fmt.Sprintf("Number of %s jobs started", name),
			Measurement: "Jobs",
			Unit:        metric.Unit_COUNT,
		}),

		NumSucceeded: metric.NewCounter(metric.Metadata{
			Name:        fmt.Sprintf("schedules.%s.succeeded", name),
			Help:        fmt.Sprintf("Number of %s jobs succeeded", name),
			Measurement: "Jobs",
			Unit:        metric.Unit_COUNT,
		}),

		NumFailed: metric.NewCounter(metric.Metadata{
			Name:        fmt.Sprintf("schedules.%s.failed", name),
			Help:        fmt.Sprintf("Number of %s jobs failed", name),
			Measurement: "Jobs",
			Unit:        metric.Unit_COUNT,
		}),
	}
}
