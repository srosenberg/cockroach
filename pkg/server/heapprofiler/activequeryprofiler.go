package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/cgroups"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

var (
	varianceBytes          = int64(64 * 1024 * 1024)
	memLimitFn             = cgroups.GetMemoryLimit
	memUsageFn             = cgroups.GetMemoryUsage
	memInactiveFileUsageFn = cgroups.GetMemoryInactiveFileUsage
)

type ActiveQueryProfiler struct {
	profiler
	cgroupMemLimit int64
	mu             struct {
		syncutil.Mutex
		prevMemUsage int64
	}
}

const (
	QueryFileNamePrefix = "activequeryprof"

	QueryFileNameSuffix = ".csv"
)

func NewActiveQueryProfiler(
	ctx context.Context, dir string, st *cluster.Settings,
) (*ActiveQueryProfiler, error) {
	__antithesis_instrumentation__.Notify(193503)
	if dir == "" {
		__antithesis_instrumentation__.Notify(193507)
		return nil, errors.AssertionFailedf("need to specify dir for NewQueryProfiler")
	} else {
		__antithesis_instrumentation__.Notify(193508)
	}
	__antithesis_instrumentation__.Notify(193504)

	dumpStore := dumpstore.NewStore(dir, maxCombinedFileSize, st)

	maxMem, warn, err := memLimitFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(193509)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193510)
	}
	__antithesis_instrumentation__.Notify(193505)
	if warn != "" {
		__antithesis_instrumentation__.Notify(193511)
		log.Warningf(ctx, "warning when reading cgroup memory limit: %s", warn)
	} else {
		__antithesis_instrumentation__.Notify(193512)
	}
	__antithesis_instrumentation__.Notify(193506)

	log.Infof(ctx, "writing go query profiles to %s", dir)
	qp := &ActiveQueryProfiler{
		profiler: profiler{
			store: newProfileStore(dumpStore, QueryFileNamePrefix, QueryFileNameSuffix, st),
		},
		cgroupMemLimit: maxMem,
	}
	return qp, nil
}

func (o *ActiveQueryProfiler) MaybeDumpQueries(
	ctx context.Context, registry *sql.SessionRegistry, st *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(193513)
	defer func() {
		__antithesis_instrumentation__.Notify(193515)
		if p := recover(); p != nil {
			__antithesis_instrumentation__.Notify(193516)
			logcrash.ReportPanic(ctx, &st.SV, p, 1)
		} else {
			__antithesis_instrumentation__.Notify(193517)
		}
	}()
	__antithesis_instrumentation__.Notify(193514)
	shouldDump, memUsage := o.shouldDump(ctx, st)
	now := o.now()
	if shouldDump && func() bool {
		__antithesis_instrumentation__.Notify(193518)
		return o.takeQueryProfile(ctx, registry, now, memUsage) == true
	}() == true {
		__antithesis_instrumentation__.Notify(193519)

		o.store.gcProfiles(ctx, now)
	} else {
		__antithesis_instrumentation__.Notify(193520)
	}
}

func (o *ActiveQueryProfiler) shouldDump(ctx context.Context, st *cluster.Settings) (bool, int64) {
	__antithesis_instrumentation__.Notify(193521)
	if !ActiveQueryDumpsEnabled.Get(&st.SV) {
		__antithesis_instrumentation__.Notify(193527)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(193528)
	}
	__antithesis_instrumentation__.Notify(193522)
	cgMemUsage, _, err := memUsageFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(193529)
		log.Errorf(ctx, "failed to fetch cgroup memory usage: %v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(193530)
	}
	__antithesis_instrumentation__.Notify(193523)
	cgInactiveFileUsage, _, err := memInactiveFileUsageFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(193531)
		log.Errorf(ctx, "failed to fetch cgroup memory inactive file usage: %v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(193532)
	}
	__antithesis_instrumentation__.Notify(193524)
	curMemUsage := cgMemUsage - cgInactiveFileUsage

	defer func() {
		__antithesis_instrumentation__.Notify(193533)
		o.mu.Lock()
		defer o.mu.Unlock()
		o.mu.prevMemUsage = curMemUsage
	}()
	__antithesis_instrumentation__.Notify(193525)

	o.mu.Lock()
	defer o.mu.Unlock()
	if o.mu.prevMemUsage == 0 || func() bool {
		__antithesis_instrumentation__.Notify(193534)
		return curMemUsage <= o.mu.prevMemUsage == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(193535)
		return o.knobs.dontWriteProfiles == true
	}() == true {
		__antithesis_instrumentation__.Notify(193536)
		return false, curMemUsage
	} else {
		__antithesis_instrumentation__.Notify(193537)
	}
	__antithesis_instrumentation__.Notify(193526)
	diff := curMemUsage - o.mu.prevMemUsage

	return curMemUsage+diff >= o.cgroupMemLimit-varianceBytes, curMemUsage
}

func (o *ActiveQueryProfiler) takeQueryProfile(
	ctx context.Context, registry *sql.SessionRegistry, now time.Time, curMemUsage int64,
) bool {
	__antithesis_instrumentation__.Notify(193538)
	path := o.store.makeNewFileName(now, curMemUsage)

	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(193541)
		log.Errorf(ctx, "error creating query profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193542)
	}
	__antithesis_instrumentation__.Notify(193539)
	defer f.Close()

	writer := debug.NewActiveQueriesWriter(registry.SerializeAll(), f)
	err = writer.Write()
	if err != nil {
		__antithesis_instrumentation__.Notify(193543)
		log.Errorf(ctx, "error writing active queries to profile: %v", err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193544)
	}
	__antithesis_instrumentation__.Notify(193540)
	return true
}
