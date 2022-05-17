package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/redact"
)

type ingestionPerformanceStats struct {
	dataSizeAtomic int64

	bufferFlushes      int
	flushesDueToSize   int
	batches            int
	batchesDueToRange  int
	batchesDueToSize   int
	splitRetriesAtomic int64

	splits, scatters int
	scatterMoved     sz

	fillWait        time.Duration
	sortWait        time.Duration
	flushWait       time.Duration
	batchWaitAtomic int64
	sendWaitAtomic  int64
	splitWait       time.Duration
	scatterWait     time.Duration
	commitWait      time.Duration

	span roachpb.Span

	*sendWaitByStore
}

type sendWaitByStore struct {
	syncutil.Mutex
	timings map[roachpb.StoreID]time.Duration
}

func (s ingestionPerformanceStats) LogTimings(ctx context.Context, name, action string) {
	__antithesis_instrumentation__.Notify(86632)
	log.Infof(ctx,
		"%s adder %s; ingested %s: %s filling; %v sorting; %v / %v flushing; %v sending; %v splitting; %d; %v scattering, %d, %v; %v commit-wait",
		name,
		redact.Safe(action),
		sz(atomic.LoadInt64(&s.dataSizeAtomic)),
		timing(s.fillWait),
		timing(s.sortWait),
		timing(s.flushWait),
		timing(atomic.LoadInt64(&s.batchWaitAtomic)),
		timing(atomic.LoadInt64(&s.sendWaitAtomic)),
		timing(s.splitWait),
		s.splits,
		timing(s.scatterWait),
		s.scatters,
		s.scatterMoved,
		timing(s.commitWait),
	)
}

func (s ingestionPerformanceStats) LogFlushes(
	ctx context.Context, name, action string, bufSize sz,
) {
	__antithesis_instrumentation__.Notify(86633)
	log.Infof(ctx,
		"%s adder %s; flushed into %s %d times, %d due to buffer size (%s); flushing chunked into %d files (%d for ranges, %d for sst size) +%d split-retries",
		name,
		redact.Safe(action),
		s.span,
		s.bufferFlushes,
		s.flushesDueToSize,
		bufSize,
		s.batches,
		s.batchesDueToRange,
		s.batchesDueToSize,
		atomic.LoadInt64(&s.splitRetriesAtomic),
	)
}

func (s *sendWaitByStore) LogPerStoreTimings(ctx context.Context, name string) {
	__antithesis_instrumentation__.Notify(86634)
	s.Lock()
	defer s.Unlock()

	if len(s.timings) == 0 {
		__antithesis_instrumentation__.Notify(86638)
		return
	} else {
		__antithesis_instrumentation__.Notify(86639)
	}
	__antithesis_instrumentation__.Notify(86635)
	ids := make(roachpb.StoreIDSlice, 0, len(s.timings))
	for i := range s.timings {
		__antithesis_instrumentation__.Notify(86640)
		ids = append(ids, i)
	}
	__antithesis_instrumentation__.Notify(86636)
	sort.Sort(ids)

	var sb strings.Builder
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(86641)

		if i > 0 && func() bool {
			__antithesis_instrumentation__.Notify(86643)
			return ids[i-1] != id-1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(86644)
			s.timings[id-1] = 0
			fmt.Fprintf(&sb, "%d: %s;", id-1, timing(0))
		} else {
			__antithesis_instrumentation__.Notify(86645)
		}
		__antithesis_instrumentation__.Notify(86642)
		fmt.Fprintf(&sb, "%d: %s;", id, timing(s.timings[id]))

	}
	__antithesis_instrumentation__.Notify(86637)
	log.Infof(ctx, "%s waited on sending to: %s", name, redact.Safe(sb.String()))
}

type sz int64

func (b sz) String() string {
	__antithesis_instrumentation__.Notify(86646)
	return string(humanizeutil.IBytes(int64(b)))
}
func (b sz) SafeValue() { __antithesis_instrumentation__.Notify(86647) }

type timing time.Duration

func (t timing) String() string {
	__antithesis_instrumentation__.Notify(86648)
	return time.Duration(t).Round(time.Second).String()
}
func (t timing) SafeValue() { __antithesis_instrumentation__.Notify(86649) }
