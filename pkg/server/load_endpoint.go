package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func makeStatusLoadHandler(
	ctx context.Context, rsr *status.RuntimeStatSampler, metricSource metricMarshaler,
) func(http.ResponseWriter, *http.Request) {
	__antithesis_instrumentation__.Notify(194236)
	cpuUserNanos := metric.NewGauge(rsr.CPUUserNS.GetMetadata())
	cpuSysNanos := metric.NewGauge(rsr.CPUSysNS.GetMetadata())
	cpuNowNanos := metric.NewGauge(rsr.CPUNowNS.GetMetadata())
	registry := metric.NewRegistry()
	registry.AddMetric(cpuUserNanos)
	registry.AddMetric(cpuSysNanos)
	registry.AddMetric(cpuNowNanos)

	exporter := metric.MakePrometheusExporter()
	regScrape := func(pm *metric.PrometheusExporter) {
		__antithesis_instrumentation__.Notify(194238)
		pm.ScrapeRegistry(registry, true)
	}
	__antithesis_instrumentation__.Notify(194237)

	exporter2 := metric.MakePrometheusExporterForSelectedMetrics(map[string]struct{}{
		sql.MetaQueryExecuted.Name:       {},
		pgwire.MetaConns.Name:            {},
		jobs.MetaRunningNonIdleJobs.Name: {},
	})

	return func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(194239)
		userTimeMillis, sysTimeMillis, err := status.GetCPUTime(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(194242)

			log.Ops.Errorf(ctx, "unable to get cpu usage: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(194243)
		}
		__antithesis_instrumentation__.Notify(194240)

		utime := userTimeMillis * 1e6
		stime := sysTimeMillis * 1e6
		cpuUserNanos.Update(utime)
		cpuSysNanos.Update(stime)
		cpuNowNanos.Update(timeutil.Now().UnixNano())

		if err := exporter.ScrapeAndPrintAsText(w, regScrape); err != nil {
			__antithesis_instrumentation__.Notify(194244)
			log.Errorf(r.Context(), "%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(194245)
		}
		__antithesis_instrumentation__.Notify(194241)

		if err := exporter2.ScrapeAndPrintAsText(w, metricSource.ScrapeIntoPrometheus); err != nil {
			__antithesis_instrumentation__.Notify(194246)
			log.Errorf(r.Context(), "%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(194247)
		}
	}
}
