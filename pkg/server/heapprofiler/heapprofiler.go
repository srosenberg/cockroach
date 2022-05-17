package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"runtime/pprof"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type HeapProfiler struct {
	profiler
}

const HeapFileNamePrefix = "memprof"

const HeapFileNameSuffix = ".pprof"

func NewHeapProfiler(ctx context.Context, dir string, st *cluster.Settings) (*HeapProfiler, error) {
	__antithesis_instrumentation__.Notify(193564)
	if dir == "" {
		__antithesis_instrumentation__.Notify(193566)
		return nil, errors.AssertionFailedf("need to specify dir for NewHeapProfiler")
	} else {
		__antithesis_instrumentation__.Notify(193567)
	}
	__antithesis_instrumentation__.Notify(193565)

	log.Infof(ctx, "writing go heap profiles to %s at least every %s", dir, resetHighWaterMarkInterval)

	dumpStore := dumpstore.NewStore(dir, maxCombinedFileSize, st)

	hp := &HeapProfiler{
		profiler{
			store: newProfileStore(dumpStore, HeapFileNamePrefix, HeapFileNameSuffix, st),
		},
	}
	return hp, nil
}

func (o *HeapProfiler) MaybeTakeProfile(ctx context.Context, curHeap int64) {
	__antithesis_instrumentation__.Notify(193568)
	o.maybeTakeProfile(ctx, curHeap, takeHeapProfile)
}

func takeHeapProfile(ctx context.Context, path string) (success bool) {
	__antithesis_instrumentation__.Notify(193569)

	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(193572)
		log.Warningf(ctx, "error creating go heap profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193573)
	}
	__antithesis_instrumentation__.Notify(193570)
	defer f.Close()
	if err = pprof.WriteHeapProfile(f); err != nil {
		__antithesis_instrumentation__.Notify(193574)
		log.Warningf(ctx, "error writing go heap profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193575)
	}
	__antithesis_instrumentation__.Notify(193571)
	return true
}
