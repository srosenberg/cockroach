package tracedumper

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/dumpstore"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/zipper"
	"github.com/cockroachdb/errors"
)

const (
	jobTraceDumpPrefix = "job_trace_dump"
	timeFormat         = "2006-01-02T15_04_05.000"
)

var (
	totalDumpSizeLimit = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"server.job_trace.total_dump_size_limit",
		"total size of job trace dumps to be kept. "+
			"Dumps are GC'ed in the order of creation time. The latest dump is "+
			"always kept even if its size exceeds the limit.",
		500<<20,
	)
)

type TraceDumper struct {
	currentTime func() time.Time
	store       *dumpstore.DumpStore
}

func (t *TraceDumper) PreFilter(
	ctx context.Context, files []os.FileInfo, _ func(fileName string) error,
) (preserved map[int]bool, err error) {
	__antithesis_instrumentation__.Notify(239504)
	preserved = make(map[int]bool)
	for i := len(files) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(239506)

		if t.CheckOwnsFile(ctx, files[i]) {
			__antithesis_instrumentation__.Notify(239507)
			preserved[i] = true
			break
		} else {
			__antithesis_instrumentation__.Notify(239508)
		}
	}
	__antithesis_instrumentation__.Notify(239505)
	return
}

func (t *TraceDumper) CheckOwnsFile(ctx context.Context, fi os.FileInfo) bool {
	__antithesis_instrumentation__.Notify(239509)
	return strings.HasPrefix(fi.Name(), jobTraceDumpPrefix)
}

var _ dumpstore.Dumper = &TraceDumper{}

func (t *TraceDumper) Dump(
	ctx context.Context, name string, traceID int64, ie sqlutil.InternalExecutor,
) {
	__antithesis_instrumentation__.Notify(239510)
	err := func() error {
		__antithesis_instrumentation__.Notify(239512)
		now := t.currentTime()
		traceZipFile := fmt.Sprintf(
			"%s.%s.%s.zip",
			jobTraceDumpPrefix,
			now.Format(timeFormat),
			name,
		)
		z := zipper.MakeInternalExecutorInflightTraceZipper(ie)
		zipBytes, err := z.Zip(ctx, traceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(239516)
			return errors.Wrap(err, "failed to collect inflight trace zip")
		} else {
			__antithesis_instrumentation__.Notify(239517)
		}
		__antithesis_instrumentation__.Notify(239513)
		path := t.store.GetFullPath(traceZipFile)
		f, err := os.Create(path)
		if err != nil {
			__antithesis_instrumentation__.Notify(239518)
			return errors.Wrapf(err, "error creating file %q for trace dump", path)
		} else {
			__antithesis_instrumentation__.Notify(239519)
		}
		__antithesis_instrumentation__.Notify(239514)
		defer f.Close()
		_, err = f.Write(zipBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(239520)
			return errors.Newf("error writing zip file %q for trace dump", path)
		} else {
			__antithesis_instrumentation__.Notify(239521)
		}
		__antithesis_instrumentation__.Notify(239515)
		return nil
	}()
	__antithesis_instrumentation__.Notify(239511)
	if err != nil {
		__antithesis_instrumentation__.Notify(239522)
		log.Errorf(ctx, "failed to dump trace %v", err)
	} else {
		__antithesis_instrumentation__.Notify(239523)
	}
}

func NewTraceDumper(ctx context.Context, dir string, st *cluster.Settings) *TraceDumper {
	__antithesis_instrumentation__.Notify(239524)
	if dir == "" {
		__antithesis_instrumentation__.Notify(239526)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(239527)
	}
	__antithesis_instrumentation__.Notify(239525)

	log.Infof(ctx, "writing job trace dumps to %s", dir)

	td := &TraceDumper{
		currentTime: timeutil.Now,
		store:       dumpstore.NewStore(dir, totalDumpSizeLimit, st),
	}
	return td
}
