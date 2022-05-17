package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/server/goroutinedumper"
	"github.com/cockroachdb/cockroach/pkg/server/heapprofiler"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type sampleEnvironmentCfg struct {
	st                   *cluster.Settings
	stopper              *stop.Stopper
	minSampleInterval    time.Duration
	goroutineDumpDirName string
	heapProfileDirName   string
	runtime              *status.RuntimeStatSampler
	sessionRegistry      *sql.SessionRegistry
}

func startSampleEnvironment(
	ctx context.Context,
	settings *cluster.Settings,
	stopper *stop.Stopper,
	goroutineDumpDirName string,
	heapProfileDirName string,
	runtimeSampler *status.RuntimeStatSampler,
	sessionRegistry *sql.SessionRegistry,
) error {
	__antithesis_instrumentation__.Notify(193351)
	cfg := sampleEnvironmentCfg{
		st:                   settings,
		stopper:              stopper,
		minSampleInterval:    base.DefaultMetricsSampleInterval,
		goroutineDumpDirName: goroutineDumpDirName,
		heapProfileDirName:   heapProfileDirName,
		runtime:              runtimeSampler,
		sessionRegistry:      sessionRegistry,
	}

	var goroutineDumper *goroutinedumper.GoroutineDumper
	if cfg.goroutineDumpDirName != "" {
		__antithesis_instrumentation__.Notify(193354)
		hasValidDumpDir := true
		if err := os.MkdirAll(cfg.goroutineDumpDirName, 0755); err != nil {
			__antithesis_instrumentation__.Notify(193356)

			log.Warningf(ctx, "cannot create goroutine dump dir -- goroutine dumps will be disabled: %v", err)
			hasValidDumpDir = false
		} else {
			__antithesis_instrumentation__.Notify(193357)
		}
		__antithesis_instrumentation__.Notify(193355)
		if hasValidDumpDir {
			__antithesis_instrumentation__.Notify(193358)
			var err error
			goroutineDumper, err = goroutinedumper.NewGoroutineDumper(ctx, cfg.goroutineDumpDirName, cfg.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(193359)
				return errors.Wrap(err, "starting goroutine dumper worker")
			} else {
				__antithesis_instrumentation__.Notify(193360)
			}
		} else {
			__antithesis_instrumentation__.Notify(193361)
		}
	} else {
		__antithesis_instrumentation__.Notify(193362)
	}
	__antithesis_instrumentation__.Notify(193352)

	var heapProfiler *heapprofiler.HeapProfiler
	var nonGoAllocProfiler *heapprofiler.NonGoAllocProfiler
	var statsProfiler *heapprofiler.StatsProfiler
	var queryProfiler *heapprofiler.ActiveQueryProfiler
	if cfg.heapProfileDirName != "" {
		__antithesis_instrumentation__.Notify(193363)
		hasValidDumpDir := true
		if err := os.MkdirAll(cfg.heapProfileDirName, 0755); err != nil {
			__antithesis_instrumentation__.Notify(193365)

			log.Warningf(ctx, "cannot create memory dump dir -- memory profile dumps will be disabled: %v", err)
			hasValidDumpDir = false
		} else {
			__antithesis_instrumentation__.Notify(193366)
		}
		__antithesis_instrumentation__.Notify(193364)

		if hasValidDumpDir {
			__antithesis_instrumentation__.Notify(193367)
			var err error
			heapProfiler, err = heapprofiler.NewHeapProfiler(ctx, cfg.heapProfileDirName, cfg.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(193371)
				return errors.Wrap(err, "starting heap profiler worker")
			} else {
				__antithesis_instrumentation__.Notify(193372)
			}
			__antithesis_instrumentation__.Notify(193368)
			nonGoAllocProfiler, err = heapprofiler.NewNonGoAllocProfiler(ctx, cfg.heapProfileDirName, cfg.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(193373)
				return errors.Wrap(err, "starting non-go alloc profiler worker")
			} else {
				__antithesis_instrumentation__.Notify(193374)
			}
			__antithesis_instrumentation__.Notify(193369)
			statsProfiler, err = heapprofiler.NewStatsProfiler(ctx, cfg.heapProfileDirName, cfg.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(193375)
				return errors.Wrap(err, "starting memory stats collector worker")
			} else {
				__antithesis_instrumentation__.Notify(193376)
			}
			__antithesis_instrumentation__.Notify(193370)
			queryProfiler, err = heapprofiler.NewActiveQueryProfiler(ctx, cfg.heapProfileDirName, cfg.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(193377)
				log.Warningf(ctx, "failed to start query profiler worker: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(193378)
			}
		} else {
			__antithesis_instrumentation__.Notify(193379)
		}
	} else {
		__antithesis_instrumentation__.Notify(193380)
	}
	__antithesis_instrumentation__.Notify(193353)

	return cfg.stopper.RunAsyncTaskEx(ctx,
		stop.TaskOpts{TaskName: "mem-logger", SpanOpt: stop.SterileRootSpan},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(193381)
			var goMemStats atomic.Value
			goMemStats.Store(&status.GoMemStats{})
			var collectingMemStats int32

			timer := timeutil.NewTimer()
			defer timer.Stop()
			timer.Reset(cfg.minSampleInterval)

			for {
				__antithesis_instrumentation__.Notify(193382)
				select {
				case <-cfg.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(193383)
					return
				case <-timer.C:
					__antithesis_instrumentation__.Notify(193384)
					timer.Read = true
					timer.Reset(cfg.minSampleInterval)

					statsCollected := make(chan struct{})
					if atomic.CompareAndSwapInt32(&collectingMemStats, 0, 1) {
						__antithesis_instrumentation__.Notify(193389)
						if err := cfg.stopper.RunAsyncTaskEx(ctx,
							stop.TaskOpts{TaskName: "get-mem-stats"},
							func(ctx context.Context) {
								__antithesis_instrumentation__.Notify(193390)
								var ms status.GoMemStats
								runtime.ReadMemStats(&ms.MemStats)
								ms.Collected = timeutil.Now()
								log.VEventf(ctx, 2, "memstats: %+v", ms)

								goMemStats.Store(&ms)
								atomic.StoreInt32(&collectingMemStats, 0)
								close(statsCollected)
							}); err != nil {
							__antithesis_instrumentation__.Notify(193391)
							close(statsCollected)
						} else {
							__antithesis_instrumentation__.Notify(193392)
						}
					} else {
						__antithesis_instrumentation__.Notify(193393)
					}
					__antithesis_instrumentation__.Notify(193385)

					select {
					case <-statsCollected:
						__antithesis_instrumentation__.Notify(193394)

					case <-time.After(time.Second):
						__antithesis_instrumentation__.Notify(193395)
					}
					__antithesis_instrumentation__.Notify(193386)

					curStats := goMemStats.Load().(*status.GoMemStats)
					cgoStats := status.GetCGoMemStats(ctx)
					cfg.runtime.SampleEnvironment(ctx, curStats, cgoStats)

					if goroutineDumper != nil {
						__antithesis_instrumentation__.Notify(193396)
						goroutineDumper.MaybeDump(ctx, cfg.st, cfg.runtime.Goroutines.Value())
					} else {
						__antithesis_instrumentation__.Notify(193397)
					}
					__antithesis_instrumentation__.Notify(193387)
					if heapProfiler != nil {
						__antithesis_instrumentation__.Notify(193398)
						heapProfiler.MaybeTakeProfile(ctx, cfg.runtime.GoAllocBytes.Value())
						nonGoAllocProfiler.MaybeTakeProfile(ctx, cfg.runtime.CgoTotalBytes.Value())
						statsProfiler.MaybeTakeProfile(ctx, cfg.runtime.RSSBytes.Value(), curStats, cgoStats)
					} else {
						__antithesis_instrumentation__.Notify(193399)
					}
					__antithesis_instrumentation__.Notify(193388)
					if queryProfiler != nil {
						__antithesis_instrumentation__.Notify(193400)
						queryProfiler.MaybeDumpQueries(ctx, cfg.sessionRegistry, cfg.st)
					} else {
						__antithesis_instrumentation__.Notify(193401)
					}
				}
			}
		})
}
