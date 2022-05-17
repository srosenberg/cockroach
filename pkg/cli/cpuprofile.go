package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var maxCombinedCPUProfFileSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"server.cpu_profile.total_dump_size_limit",
	"maximum combined disk size of preserved CPU profiles",
	128<<20,
)

const cpuProfTimeFormat = "2006-01-02T15_04_05.000"
const cpuProfFileNamePrefix = "cpuprof."

type cpuProfiler struct{}

func (s cpuProfiler) PreFilter(
	ctx context.Context, files []os.FileInfo, cleanupFn func(fileName string) error,
) (preserved map[int]bool, _ error) {
	__antithesis_instrumentation__.Notify(30388)
	preserved = make(map[int]bool)

	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(30390)
		if s.CheckOwnsFile(ctx, files[i]) {
			__antithesis_instrumentation__.Notify(30391)
			preserved[i] = true
			break
		} else {
			__antithesis_instrumentation__.Notify(30392)
		}
	}
	__antithesis_instrumentation__.Notify(30389)
	return
}

func (s cpuProfiler) CheckOwnsFile(_ context.Context, fi os.FileInfo) bool {
	__antithesis_instrumentation__.Notify(30393)
	return strings.HasPrefix(fi.Name(), cpuProfFileNamePrefix)
}

func initCPUProfile(ctx context.Context, dir string, st *cluster.Settings) {
	__antithesis_instrumentation__.Notify(30394)
	cpuProfileInterval := envutil.EnvOrDefaultDuration("COCKROACH_CPUPROF_INTERVAL", -1)
	if cpuProfileInterval <= 0 {
		__antithesis_instrumentation__.Notify(30399)
		return
	} else {
		__antithesis_instrumentation__.Notify(30400)
	}
	__antithesis_instrumentation__.Notify(30395)

	if dir == "" {
		__antithesis_instrumentation__.Notify(30401)
		return
	} else {
		__antithesis_instrumentation__.Notify(30402)
	}
	__antithesis_instrumentation__.Notify(30396)
	if err := os.MkdirAll(dir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(30403)

		log.Warningf(ctx, "cannot create CPU profile dump dir -- CPU profiles will be disabled: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(30404)
	}
	__antithesis_instrumentation__.Notify(30397)

	if min := time.Second; cpuProfileInterval < min {
		__antithesis_instrumentation__.Notify(30405)
		log.Infof(ctx, "fixing excessively short cpu profiling interval: %s -> %s",
			cpuProfileInterval, min)
		cpuProfileInterval = min
	} else {
		__antithesis_instrumentation__.Notify(30406)
	}
	__antithesis_instrumentation__.Notify(30398)

	profilestore := dumpstore.NewStore(dir, maxCombinedCPUProfFileSize, st)
	profiler := dumpstore.Dumper(cpuProfiler{})

	go func() {
		__antithesis_instrumentation__.Notify(30407)
		defer logcrash.RecoverAndReportPanic(ctx, &serverCfg.Settings.SV)

		ctx := context.Background()

		t := time.NewTicker(cpuProfileInterval)
		defer t.Stop()

		var currentProfile *os.File
		defer func() {
			__antithesis_instrumentation__.Notify(30409)
			if currentProfile != nil {
				__antithesis_instrumentation__.Notify(30410)
				pprof.StopCPUProfile()
				currentProfile.Close()
			} else {
				__antithesis_instrumentation__.Notify(30411)
			}
		}()
		__antithesis_instrumentation__.Notify(30408)

		for {
			__antithesis_instrumentation__.Notify(30412)

			if err := debug.CPUProfileDo(st, cluster.CPUProfileDefault, func() error {
				__antithesis_instrumentation__.Notify(30413)

				var buf bytes.Buffer

				if err := pprof.StartCPUProfile(&buf); err != nil {
					__antithesis_instrumentation__.Notify(30416)
					return err
				} else {
					__antithesis_instrumentation__.Notify(30417)
				}
				__antithesis_instrumentation__.Notify(30414)

				<-t.C

				pprof.StopCPUProfile()

				now := timeutil.Now()
				name := cpuProfFileNamePrefix + now.Format(cpuProfTimeFormat)
				path := profilestore.GetFullPath(name)
				if err := ioutil.WriteFile(path, buf.Bytes(), 0644); err != nil {
					__antithesis_instrumentation__.Notify(30418)
					return err
				} else {
					__antithesis_instrumentation__.Notify(30419)
				}
				__antithesis_instrumentation__.Notify(30415)
				profilestore.GC(ctx, now, profiler)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(30420)

				log.Infof(ctx, "error during CPU profile: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(30421)
			}
		}
	}()
}
