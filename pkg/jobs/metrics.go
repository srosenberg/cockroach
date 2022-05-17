package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type Metrics struct {
	JobMetrics [jobspb.NumJobTypes]*JobTypeMetrics

	RunningNonIdleJobs *metric.Gauge

	RowLevelTTL  metric.Struct
	Changefeed   metric.Struct
	StreamIngest metric.Struct

	AdoptIterations *metric.Counter

	ClaimedJobs *metric.Counter

	ResumedJobs *metric.Counter
}

type JobTypeMetrics struct {
	CurrentlyRunning       *metric.Gauge
	CurrentlyIdle          *metric.Gauge
	ResumeCompleted        *metric.Counter
	ResumeRetryError       *metric.Counter
	ResumeFailed           *metric.Counter
	FailOrCancelCompleted  *metric.Counter
	FailOrCancelRetryError *metric.Counter

	FailOrCancelFailed *metric.Counter
}

func (JobTypeMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(84202) }

func makeMetaCurrentlyRunning(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84203)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.currently_running", typeStr),
		Help: fmt.Sprintf("Number of %s jobs currently running in Resume or OnFailOrCancel state",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaCurrentlyIdle(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84204)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.currently_idle", typeStr),
		Help: fmt.Sprintf("Number of %s jobs currently considered Idle and can be freely shut down",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaResumeCompeted(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84205)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.resume_completed", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which successfully resumed to completion",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaResumeRetryError(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84206)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.resume_retry_error", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which failed with a retriable error",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaResumeFailed(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84207)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.resume_failed", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which failed with a non-retriable error",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaFailOrCancelCompeted(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84208)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.fail_or_cancel_completed", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which successfully completed "+
			"their failure or cancelation process",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaFailOrCancelRetryError(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84209)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.fail_or_cancel_retry_error", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which failed with a retriable "+
			"error on their failure or cancelation process",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

func makeMetaFailOrCancelFailed(typeStr string) metric.Metadata {
	__antithesis_instrumentation__.Notify(84210)
	return metric.Metadata{
		Name: fmt.Sprintf("jobs.%s.fail_or_cancel_failed", typeStr),
		Help: fmt.Sprintf("Number of %s jobs which failed with a "+
			"non-retriable error on their failure or cancelation process",
			typeStr),
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
}

var (
	metaAdoptIterations = metric.Metadata{
		Name:        "jobs.adopt_iterations",
		Help:        "number of job-adopt iterations performed by the registry",
		Measurement: "iterations",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}

	metaClaimedJobs = metric.Metadata{
		Name:        "jobs.claimed_jobs",
		Help:        "number of jobs claimed in job-adopt iterations",
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}

	metaResumedClaimedJobs = metric.Metadata{
		Name:        "jobs.resumed_claimed_jobs",
		Help:        "number of claimed-jobs resumed in job-adopt iterations",
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}

	MetaRunningNonIdleJobs = metric.Metadata{
		Name:        "jobs.running_non_idle",
		Help:        "number of running jobs that are not idle",
		Measurement: "jobs",
		Unit:        metric.Unit_COUNT,
		MetricType:  io_prometheus_client.MetricType_GAUGE,
	}
)

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(84211) }

func (m *Metrics) init(histogramWindowInterval time.Duration) {
	if MakeRowLevelTTLMetricsHook != nil {
		m.RowLevelTTL = MakeRowLevelTTLMetricsHook(histogramWindowInterval)
	}
	if MakeChangefeedMetricsHook != nil {
		m.Changefeed = MakeChangefeedMetricsHook(histogramWindowInterval)
	}
	if MakeStreamIngestMetricsHook != nil {
		m.StreamIngest = MakeStreamIngestMetricsHook(histogramWindowInterval)
	}
	m.AdoptIterations = metric.NewCounter(metaAdoptIterations)
	m.ClaimedJobs = metric.NewCounter(metaClaimedJobs)
	m.ResumedJobs = metric.NewCounter(metaResumedClaimedJobs)
	m.RunningNonIdleJobs = metric.NewGauge(MetaRunningNonIdleJobs)
	for i := 0; i < jobspb.NumJobTypes; i++ {
		jt := jobspb.Type(i)
		if jt == jobspb.TypeUnspecified {
			continue
		}
		typeStr := strings.ToLower(strings.Replace(jt.String(), " ", "_", -1))
		m.JobMetrics[jt] = &JobTypeMetrics{
			CurrentlyRunning:       metric.NewGauge(makeMetaCurrentlyRunning(typeStr)),
			CurrentlyIdle:          metric.NewGauge(makeMetaCurrentlyIdle(typeStr)),
			ResumeCompleted:        metric.NewCounter(makeMetaResumeCompeted(typeStr)),
			ResumeRetryError:       metric.NewCounter(makeMetaResumeRetryError(typeStr)),
			ResumeFailed:           metric.NewCounter(makeMetaResumeFailed(typeStr)),
			FailOrCancelCompleted:  metric.NewCounter(makeMetaFailOrCancelCompeted(typeStr)),
			FailOrCancelRetryError: metric.NewCounter(makeMetaFailOrCancelRetryError(typeStr)),
			FailOrCancelFailed:     metric.NewCounter(makeMetaFailOrCancelFailed(typeStr)),
		}
	}
}

var MakeChangefeedMetricsHook func(time.Duration) metric.Struct

var MakeStreamIngestMetricsHook func(duration time.Duration) metric.Struct

var MakeRowLevelTTLMetricsHook func(time.Duration) metric.Struct

type JobTelemetryMetrics struct {
	Successful telemetry.Counter
	Failed     telemetry.Counter
	Canceled   telemetry.Counter
}

func newJobTelemetryMetrics(jobName string) *JobTelemetryMetrics {
	__antithesis_instrumentation__.Notify(84212)
	return &JobTelemetryMetrics{
		Successful: telemetry.GetCounterOnce(fmt.Sprintf("job.%s.successful", jobName)),
		Failed:     telemetry.GetCounterOnce(fmt.Sprintf("job.%s.failed", jobName)),
		Canceled:   telemetry.GetCounterOnce(fmt.Sprintf("job.%s.canceled", jobName)),
	}
}

func getJobTelemetryMetricsArray() [jobspb.NumJobTypes]*JobTelemetryMetrics {
	__antithesis_instrumentation__.Notify(84213)
	var metrics [jobspb.NumJobTypes]*JobTelemetryMetrics
	for i := 0; i < jobspb.NumJobTypes; i++ {
		__antithesis_instrumentation__.Notify(84215)
		jt := jobspb.Type(i)
		if jt == jobspb.TypeUnspecified {
			__antithesis_instrumentation__.Notify(84217)
			continue
		} else {
			__antithesis_instrumentation__.Notify(84218)
		}
		__antithesis_instrumentation__.Notify(84216)
		typeStr := strings.ToLower(strings.Replace(jt.String(), " ", "_", -1))
		metrics[i] = newJobTelemetryMetrics(typeStr)
	}
	__antithesis_instrumentation__.Notify(84214)
	return metrics
}

var TelemetryMetrics = getJobTelemetryMetricsArray()
