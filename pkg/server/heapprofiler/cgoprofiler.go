package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type NonGoAllocProfiler struct {
	profiler
}

const JemallocFileNamePrefix = "jeprof"

const JemallocFileNameSuffix = ".jeprof"

func NewNonGoAllocProfiler(
	ctx context.Context, dir string, st *cluster.Settings,
) (*NonGoAllocProfiler, error) {
	__antithesis_instrumentation__.Notify(193545)
	if dir == "" {
		__antithesis_instrumentation__.Notify(193548)
		return nil, errors.AssertionFailedf("need to specify dir for NewHeapProfiler")
	} else {
		__antithesis_instrumentation__.Notify(193549)
	}
	__antithesis_instrumentation__.Notify(193546)

	if jemallocHeapDump != nil {
		__antithesis_instrumentation__.Notify(193550)
		log.Infof(ctx, "writing jemalloc profiles to %s at last every %s", dir, resetHighWaterMarkInterval)
	} else {
		__antithesis_instrumentation__.Notify(193551)
		log.Infof(ctx, `to enable jmalloc profiling: "export MALLOC_CONF=prof:true" or "ln -s prof:true /etc/malloc.conf"`)
	}
	__antithesis_instrumentation__.Notify(193547)

	dumpStore := dumpstore.NewStore(dir, maxCombinedFileSize, st)

	hp := &NonGoAllocProfiler{
		profiler{
			store: newProfileStore(dumpStore, JemallocFileNamePrefix, JemallocFileNameSuffix, st),
		},
	}
	return hp, nil
}

func (o *NonGoAllocProfiler) MaybeTakeProfile(ctx context.Context, curNonGoAlloc int64) {
	__antithesis_instrumentation__.Notify(193552)
	o.maybeTakeProfile(ctx, curNonGoAlloc, takeJemallocProfile)
}

func takeJemallocProfile(ctx context.Context, path string) (success bool) {
	__antithesis_instrumentation__.Notify(193553)
	if jemallocHeapDump == nil {
		__antithesis_instrumentation__.Notify(193556)
		return true
	} else {
		__antithesis_instrumentation__.Notify(193557)
	}
	__antithesis_instrumentation__.Notify(193554)
	if err := jemallocHeapDump(path); err != nil {
		__antithesis_instrumentation__.Notify(193558)
		log.Warningf(ctx, "error writing jemalloc heap %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193559)
	}
	__antithesis_instrumentation__.Notify(193555)
	return true
}

var jemallocHeapDump func(string) error

func SetJemallocHeapDumpFn(fn func(filename string) error) {
	__antithesis_instrumentation__.Notify(193560)
	if jemallocHeapDump != nil {
		__antithesis_instrumentation__.Notify(193562)
		panic("jemallocHeapDump is already set")
	} else {
		__antithesis_instrumentation__.Notify(193563)
	}
	__antithesis_instrumentation__.Notify(193561)
	jemallocHeapDump = fn
}
