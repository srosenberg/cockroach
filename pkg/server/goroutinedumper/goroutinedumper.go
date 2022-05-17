package goroutinedumper

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	goroutineDumpPrefix = "goroutine_dump"
	timeFormat          = "2006-01-02T15_04_05.000"
)

var (
	numGoroutinesThreshold = settings.RegisterIntSetting(
		settings.TenantWritable,
		"server.goroutine_dump.num_goroutines_threshold",
		"a threshold beyond which if number of goroutines increases, "+
			"then goroutine dump can be triggered",
		1000,
	)
	totalDumpSizeLimit = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"server.goroutine_dump.total_dump_size_limit",
		"total size of goroutine dumps to be kept. "+
			"Dumps are GC'ed in the order of creation time. The latest dump is "+
			"always kept even if its size exceeds the limit.",
		500<<20,
	)
)

type heuristic struct {
	name   string
	isTrue func(s *GoroutineDumper) bool
}

var doubleSinceLastDumpHeuristic = heuristic{
	name: "double_since_last_dump",
	isTrue: func(gd *GoroutineDumper) bool {
		__antithesis_instrumentation__.Notify(193410)
		return gd.goroutines > gd.goroutinesThreshold && func() bool {
			__antithesis_instrumentation__.Notify(193411)
			return gd.goroutines >= 2*gd.maxGoroutinesDumped == true
		}() == true
	},
}

type GoroutineDumper struct {
	goroutines          int64
	goroutinesThreshold int64
	maxGoroutinesDumped int64
	heuristics          []heuristic
	currentTime         func() time.Time
	takeGoroutineDump   func(path string) error
	store               *dumpstore.DumpStore
	st                  *cluster.Settings
}

func (gd *GoroutineDumper) MaybeDump(ctx context.Context, st *cluster.Settings, goroutines int64) {
	__antithesis_instrumentation__.Notify(193412)
	gd.goroutines = goroutines
	if gd.goroutinesThreshold != numGoroutinesThreshold.Get(&st.SV) {
		__antithesis_instrumentation__.Notify(193414)
		gd.goroutinesThreshold = numGoroutinesThreshold.Get(&st.SV)
		gd.maxGoroutinesDumped = 0
	} else {
		__antithesis_instrumentation__.Notify(193415)
	}
	__antithesis_instrumentation__.Notify(193413)
	for _, h := range gd.heuristics {
		__antithesis_instrumentation__.Notify(193416)
		if h.isTrue(gd) {
			__antithesis_instrumentation__.Notify(193417)
			now := gd.currentTime()
			filename := fmt.Sprintf(
				"%s.%s.%s.%09d",
				goroutineDumpPrefix,
				now.Format(timeFormat),
				h.name,
				goroutines,
			)
			path := gd.store.GetFullPath(filename)
			if err := gd.takeGoroutineDump(path); err != nil {
				__antithesis_instrumentation__.Notify(193419)
				log.Warningf(ctx, "error dumping goroutines: %s", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(193420)
			}
			__antithesis_instrumentation__.Notify(193418)
			gd.maxGoroutinesDumped = goroutines
			gd.gcDumps(ctx, now)
			break
		} else {
			__antithesis_instrumentation__.Notify(193421)
		}
	}
}

func NewGoroutineDumper(
	ctx context.Context, dir string, st *cluster.Settings,
) (*GoroutineDumper, error) {
	__antithesis_instrumentation__.Notify(193422)
	if dir == "" {
		__antithesis_instrumentation__.Notify(193424)
		return nil, errors.New("directory to store dumps could not be determined")
	} else {
		__antithesis_instrumentation__.Notify(193425)
	}
	__antithesis_instrumentation__.Notify(193423)

	log.Infof(ctx, "writing goroutine dumps to %s", dir)

	gd := &GoroutineDumper{
		heuristics: []heuristic{
			doubleSinceLastDumpHeuristic,
		},
		goroutinesThreshold: 0,
		maxGoroutinesDumped: 0,
		currentTime:         timeutil.Now,
		takeGoroutineDump:   takeGoroutineDump,
		store:               dumpstore.NewStore(dir, totalDumpSizeLimit, st),
		st:                  st,
	}
	return gd, nil
}

func (gd *GoroutineDumper) gcDumps(ctx context.Context, now time.Time) {
	__antithesis_instrumentation__.Notify(193426)
	gd.store.GC(ctx, now, gd)
}

func (gd *GoroutineDumper) PreFilter(
	ctx context.Context, files []os.FileInfo, cleanupFn func(fileName string) error,
) (preserved map[int]bool, _ error) {
	__antithesis_instrumentation__.Notify(193427)
	preserved = make(map[int]bool)
	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(193429)

		if gd.CheckOwnsFile(ctx, files[i]) {
			__antithesis_instrumentation__.Notify(193430)
			preserved[i] = true
			break
		} else {
			__antithesis_instrumentation__.Notify(193431)
		}
	}
	__antithesis_instrumentation__.Notify(193428)
	return
}

func (gd *GoroutineDumper) CheckOwnsFile(_ context.Context, fi os.FileInfo) bool {
	__antithesis_instrumentation__.Notify(193432)
	return strings.HasPrefix(fi.Name(), goroutineDumpPrefix)
}

func takeGoroutineDump(path string) error {
	__antithesis_instrumentation__.Notify(193433)
	path += ".txt.gz"
	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(193437)
		return errors.Wrapf(err, "error creating file %s for goroutine dump", path)
	} else {
		__antithesis_instrumentation__.Notify(193438)
	}
	__antithesis_instrumentation__.Notify(193434)
	defer f.Close()
	w := gzip.NewWriter(f)
	if err = pprof.Lookup("goroutine").WriteTo(w, 2); err != nil {
		__antithesis_instrumentation__.Notify(193439)
		return errors.Wrapf(err, "error writing goroutine dump to %s", path)
	} else {
		__antithesis_instrumentation__.Notify(193440)
	}
	__antithesis_instrumentation__.Notify(193435)

	if err := w.Close(); err != nil {
		__antithesis_instrumentation__.Notify(193441)
		return errors.Wrapf(err, "error closing gzip writer for %s", path)
	} else {
		__antithesis_instrumentation__.Notify(193442)
	}
	__antithesis_instrumentation__.Notify(193436)

	return f.Close()
}
