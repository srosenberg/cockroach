package heapprofiler

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type StatsProfiler struct {
	profiler
}

const StatsFileNamePrefix = "memstats"

const StatsFileNameSuffix = ".txt"

func NewStatsProfiler(
	ctx context.Context, dir string, st *cluster.Settings,
) (*StatsProfiler, error) {
	__antithesis_instrumentation__.Notify(193634)
	if dir == "" {
		__antithesis_instrumentation__.Notify(193636)
		return nil, errors.AssertionFailedf("need to specify dir for NewStatsProfiler")
	} else {
		__antithesis_instrumentation__.Notify(193637)
	}
	__antithesis_instrumentation__.Notify(193635)

	log.Infof(ctx, "writing memory stats to %s at last every %s", dir, resetHighWaterMarkInterval)

	dumpStore := dumpstore.NewStore(dir, maxCombinedFileSize, st)

	hp := &StatsProfiler{
		profiler{
			store: newProfileStore(dumpStore, StatsFileNamePrefix, StatsFileNameSuffix, st),
		},
	}
	return hp, nil
}

func (o *StatsProfiler) MaybeTakeProfile(
	ctx context.Context, curRSS int64, ms *status.GoMemStats, cs *status.CGoMemStats,
) {
	__antithesis_instrumentation__.Notify(193638)
	o.maybeTakeProfile(ctx, curRSS, func(ctx context.Context, path string) bool {
		__antithesis_instrumentation__.Notify(193639)
		return saveStats(ctx, path, ms, cs)
	})
}

func saveStats(
	ctx context.Context, path string, ms *status.GoMemStats, cs *status.CGoMemStats,
) bool {
	__antithesis_instrumentation__.Notify(193640)
	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(193645)
		log.Warningf(ctx, "error creating stats profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193646)
	}
	__antithesis_instrumentation__.Notify(193641)
	defer f.Close()
	msJ, err := json.MarshalIndent(&ms.MemStats, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(193647)
		log.Warningf(ctx, "error marshaling stats profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193648)
	}
	__antithesis_instrumentation__.Notify(193642)
	csJ, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(193649)
		log.Warningf(ctx, "error marshaling stats profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193650)
	}
	__antithesis_instrumentation__.Notify(193643)
	_, err = fmt.Fprintf(f, "Go memory stats:\n%s\n----\nNon-Go stats:\n%s\n", msJ, csJ)
	if err != nil {
		__antithesis_instrumentation__.Notify(193651)
		log.Warningf(ctx, "error writing stats profile %s: %v", path, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193652)
	}
	__antithesis_instrumentation__.Notify(193644)
	return true
}
