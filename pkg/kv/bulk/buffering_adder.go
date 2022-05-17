package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type BufferingAdder struct {
	sink SSTBatcher

	timestamp hlc.Timestamp

	maxBufferLimit func() int64

	curBuf kvBuf

	sorted bool

	initialSplits int

	lastFlush time.Time

	name string

	bulkMon *mon.BytesMonitor
	memAcc  mon.BoundAccount

	onFlush func(summary roachpb.BulkOpSummary)

	underfill sz
}

var _ kvserverbase.BulkAdder = &BufferingAdder{}

func MakeBulkAdder(
	ctx context.Context,
	db *kv.DB,
	rangeCache *rangecache.RangeCache,
	settings *cluster.Settings,
	timestamp hlc.Timestamp,
	opts kvserverbase.BulkAdderOptions,
	bulkMon *mon.BytesMonitor,
) (*BufferingAdder, error) {
	__antithesis_instrumentation__.Notify(86191)
	if opts.MinBufferSize == 0 {
		__antithesis_instrumentation__.Notify(86195)
		opts.MinBufferSize = 32 << 20
	} else {
		__antithesis_instrumentation__.Notify(86196)
	}
	__antithesis_instrumentation__.Notify(86192)
	if opts.MaxBufferSize == nil {
		__antithesis_instrumentation__.Notify(86197)
		opts.MaxBufferSize = func() int64 { __antithesis_instrumentation__.Notify(86198); return 128 << 20 }
	} else {
		__antithesis_instrumentation__.Notify(86199)
	}
	__antithesis_instrumentation__.Notify(86193)

	b := &BufferingAdder{
		name: opts.Name,
		sink: SSTBatcher{
			name:                   opts.Name,
			db:                     db,
			rc:                     rangeCache,
			settings:               settings,
			skipDuplicates:         opts.SkipDuplicates,
			disallowShadowingBelow: opts.DisallowShadowingBelow,
			batchTS:                opts.BatchTimestamp,
			writeAtBatchTS:         opts.WriteAtBatchTimestamp,
			mem:                    bulkMon.MakeBoundAccount(),
		},
		timestamp:      timestamp,
		maxBufferLimit: opts.MaxBufferSize,
		bulkMon:        bulkMon,
		sorted:         true,
		initialSplits:  opts.InitialSplitsIfUnordered,
		lastFlush:      timeutil.Now(),
	}

	b.sink.mem.Mu = &syncutil.Mutex{}

	b.memAcc = bulkMon.MakeBoundAccount()
	if opts.MinBufferSize > 0 {
		__antithesis_instrumentation__.Notify(86200)
		if err := b.memAcc.Reserve(ctx, opts.MinBufferSize); err != nil {
			__antithesis_instrumentation__.Notify(86201)
			return nil, errors.WithHint(
				errors.Wrap(err, "not enough memory available to create a BulkAdder"),
				"Try setting a higher --max-sql-memory.")
		} else {
			__antithesis_instrumentation__.Notify(86202)
		}
	} else {
		__antithesis_instrumentation__.Notify(86203)
	}
	__antithesis_instrumentation__.Notify(86194)
	return b, nil
}

func (b *BufferingAdder) SetOnFlush(fn func(summary roachpb.BulkOpSummary)) {
	__antithesis_instrumentation__.Notify(86204)
	b.onFlush = fn
}

func (b *BufferingAdder) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(86205)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(86207)
		if b.sink.stats.bufferFlushes > 0 {
			__antithesis_instrumentation__.Notify(86208)
			b.sink.stats.LogTimings(ctx, b.name, "closing")
			if log.V(3) {
				__antithesis_instrumentation__.Notify(86210)
				b.sink.stats.LogPerStoreTimings(ctx, b.name)
			} else {
				__antithesis_instrumentation__.Notify(86211)
			}
			__antithesis_instrumentation__.Notify(86209)
			b.sink.stats.LogFlushes(ctx, b.name, "closing", sz(b.memAcc.Used()))
		} else {
			__antithesis_instrumentation__.Notify(86212)
			log.Infof(ctx, "%s adder closing; ingested nothing", b.name)
		}
	} else {
		__antithesis_instrumentation__.Notify(86213)
	}
	__antithesis_instrumentation__.Notify(86206)
	b.sink.Close(ctx)

	if b.bulkMon != nil {
		__antithesis_instrumentation__.Notify(86214)
		b.memAcc.Close(ctx)
		b.bulkMon.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(86215)
	}
}

func (b *BufferingAdder) Add(ctx context.Context, key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(86216)
	if b.sorted {
		__antithesis_instrumentation__.Notify(86221)
		if l := len(b.curBuf.entries); l > 0 && func() bool {
			__antithesis_instrumentation__.Notify(86222)
			return key.Compare(b.curBuf.Key(l-1)) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(86223)
			b.sorted = false
		} else {
			__antithesis_instrumentation__.Notify(86224)
		}
	} else {
		__antithesis_instrumentation__.Notify(86225)
	}
	__antithesis_instrumentation__.Notify(86217)

	need := sz(len(key) + len(value))

	if b.curBuf.fits(ctx, need, sz(b.maxBufferLimit()), &b.memAcc) {
		__antithesis_instrumentation__.Notify(86226)
		return b.curBuf.append(key, value)
	} else {
		__antithesis_instrumentation__.Notify(86227)
	}
	__antithesis_instrumentation__.Notify(86218)

	b.sink.stats.flushesDueToSize++
	log.VEventf(ctx, 3, "%s adder triggering flush of %s of KVs in %s buffer",
		b.name, b.curBuf.KVSize(), b.bufferedMemSize())

	unusedEntries, unusedSlab := b.curBuf.unusedCap()
	b.underfill += unusedEntries + unusedSlab

	if err := b.doFlush(ctx, true); err != nil {
		__antithesis_instrumentation__.Notify(86228)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86229)
	}
	__antithesis_instrumentation__.Notify(86219)

	if b.underfill > 1<<30 {
		__antithesis_instrumentation__.Notify(86230)
		b.memAcc.Shrink(ctx, int64(b.curBuf.MemSize()))
		b.curBuf.entries = nil
		b.curBuf.slab = nil
		b.underfill = 0
	} else {
		__antithesis_instrumentation__.Notify(86231)
	}
	__antithesis_instrumentation__.Notify(86220)

	return b.curBuf.append(key, value)
}

func (b *BufferingAdder) bufferedKeys() int {
	__antithesis_instrumentation__.Notify(86232)
	return len(b.curBuf.entries)
}

func (b *BufferingAdder) bufferedMemSize() sz {
	__antithesis_instrumentation__.Notify(86233)
	return b.curBuf.MemSize()
}

func (b *BufferingAdder) CurrentBufferFill() float32 {
	__antithesis_instrumentation__.Notify(86234)
	return float32(b.curBuf.KVSize()) / float32(b.curBuf.MemSize())
}

func (b *BufferingAdder) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(86235)
	return b.curBuf.Len() == 0
}

func (b *BufferingAdder) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(86236)
	return b.doFlush(ctx, false)
}

func (b *BufferingAdder) doFlush(ctx context.Context, forSize bool) error {
	__antithesis_instrumentation__.Notify(86237)
	b.sink.stats.fillWait += timeutil.Since(b.lastFlush)

	if b.bufferedKeys() == 0 {
		__antithesis_instrumentation__.Notify(86250)
		if b.onFlush != nil {
			__antithesis_instrumentation__.Notify(86252)
			b.onFlush(b.sink.GetBatchSummary())
		} else {
			__antithesis_instrumentation__.Notify(86253)
		}
		__antithesis_instrumentation__.Notify(86251)
		b.lastFlush = timeutil.Now()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(86254)
	}
	__antithesis_instrumentation__.Notify(86238)
	if err := b.sink.Reset(ctx); err != nil {
		__antithesis_instrumentation__.Notify(86255)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86256)
	}
	__antithesis_instrumentation__.Notify(86239)
	b.sink.stats.bufferFlushes++

	before := b.sink.stats
	beforeSize := b.sink.mu.totalRows.DataSize

	beforeSort := timeutil.Now()

	if !b.sorted {
		__antithesis_instrumentation__.Notify(86257)
		sort.Sort(&b.curBuf)
	} else {
		__antithesis_instrumentation__.Notify(86258)
	}
	__antithesis_instrumentation__.Notify(86240)
	mvccKey := storage.MVCCKey{Timestamp: b.timestamp}

	beforeFlush := timeutil.Now()
	b.sink.stats.sortWait += beforeFlush.Sub(beforeSort)

	if b.initialSplits > 0 {
		__antithesis_instrumentation__.Notify(86259)
		if forSize && func() bool {
			__antithesis_instrumentation__.Notify(86261)
			return !b.sorted == true
		}() == true {
			__antithesis_instrumentation__.Notify(86262)
			if err := b.createInitialSplits(ctx); err != nil {
				__antithesis_instrumentation__.Notify(86263)
				return err
			} else {
				__antithesis_instrumentation__.Notify(86264)
			}
		} else {
			__antithesis_instrumentation__.Notify(86265)
		}
		__antithesis_instrumentation__.Notify(86260)

		b.initialSplits = 0
	} else {
		__antithesis_instrumentation__.Notify(86266)
	}
	__antithesis_instrumentation__.Notify(86241)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(86267)
		if len(b.sink.stats.span.Key) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(86268)
			return b.curBuf.Key(0).Compare(b.sink.stats.span.Key) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(86269)
			b.sink.stats.span.Key = b.curBuf.Key(0).Clone()
		} else {
			__antithesis_instrumentation__.Notify(86270)
		}
	} else {
		__antithesis_instrumentation__.Notify(86271)
	}
	__antithesis_instrumentation__.Notify(86242)

	for i := range b.curBuf.entries {
		__antithesis_instrumentation__.Notify(86272)
		mvccKey.Key = b.curBuf.Key(i)
		if err := b.sink.AddMVCCKey(ctx, mvccKey, b.curBuf.Value(i)); err != nil {
			__antithesis_instrumentation__.Notify(86273)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86274)
		}
	}
	__antithesis_instrumentation__.Notify(86243)
	if err := b.sink.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(86275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86276)
	}
	__antithesis_instrumentation__.Notify(86244)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(86277)
		if b.sink.stats.span.EndKey.Compare(mvccKey.Key) < 0 {
			__antithesis_instrumentation__.Notify(86278)
			b.sink.stats.span.EndKey = mvccKey.Key.Clone()
		} else {
			__antithesis_instrumentation__.Notify(86279)
		}
	} else {
		__antithesis_instrumentation__.Notify(86280)
	}
	__antithesis_instrumentation__.Notify(86245)

	b.sink.stats.flushWait += timeutil.Since(beforeFlush)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(86281)
		written := b.sink.mu.totalRows.DataSize - beforeSize
		files := b.sink.stats.batches - before.batches
		dueToSplits := b.sink.stats.batchesDueToRange - before.batchesDueToRange
		dueToSize := b.sink.stats.batchesDueToRange - before.batchesDueToRange

		log.Infof(ctx,
			"%s adder flushing %s (%s buffered/%0.2gx) wrote %d SSTs (avg: %s) with %d for splits, %d for size, took %v",
			b.name,
			b.curBuf.KVSize(),
			b.curBuf.MemSize(),
			float64(b.curBuf.KVSize())/float64(b.curBuf.MemSize()),
			files,
			sz(written/int64(files)),
			dueToSplits,
			dueToSize,
			timing(timeutil.Since(beforeSort)),
		)
	} else {
		__antithesis_instrumentation__.Notify(86282)
	}
	__antithesis_instrumentation__.Notify(86246)

	if log.V(2) {
		__antithesis_instrumentation__.Notify(86283)
		b.sink.stats.LogTimings(ctx, b.name, "flushed")
		if log.V(3) {
			__antithesis_instrumentation__.Notify(86284)
			b.sink.stats.LogPerStoreTimings(ctx, b.name)
		} else {
			__antithesis_instrumentation__.Notify(86285)
		}
	} else {
		__antithesis_instrumentation__.Notify(86286)
	}
	__antithesis_instrumentation__.Notify(86247)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(86287)
		b.sink.stats.LogFlushes(ctx, b.name, "flushed", sz(b.memAcc.Used()))
	} else {
		__antithesis_instrumentation__.Notify(86288)
	}
	__antithesis_instrumentation__.Notify(86248)

	if b.onFlush != nil {
		__antithesis_instrumentation__.Notify(86289)
		b.onFlush(b.sink.GetBatchSummary())
	} else {
		__antithesis_instrumentation__.Notify(86290)
	}
	__antithesis_instrumentation__.Notify(86249)
	b.curBuf.Reset()
	b.lastFlush = timeutil.Now()
	return nil
}

func (b *BufferingAdder) createInitialSplits(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(86291)
	log.Infof(ctx, "%s adder creating up to %d initial splits from %d KVs in %s buffer",
		b.name, b.initialSplits, b.curBuf.Len(), b.curBuf.KVSize())

	beforeSplits := timeutil.Now()
	hour := hlc.Timestamp{WallTime: beforeSplits.Add(time.Hour).UnixNano()}
	width := len(b.curBuf.entries) / b.initialSplits
	var toScatter []roachpb.Key
	for i := 0; i < b.initialSplits; i++ {
		__antithesis_instrumentation__.Notify(86294)
		expire := hour
		if i == 0 {
			__antithesis_instrumentation__.Notify(86300)

			expire = hour.Add(time.Hour.Nanoseconds(), 0)
		} else {
			__antithesis_instrumentation__.Notify(86301)
		}
		__antithesis_instrumentation__.Notify(86295)

		splitAt := i * width
		if splitAt >= len(b.curBuf.entries) {
			__antithesis_instrumentation__.Notify(86302)
			break
		} else {
			__antithesis_instrumentation__.Notify(86303)
		}
		__antithesis_instrumentation__.Notify(86296)

		predicateAt := splitAt - width
		if predicateAt < 0 {
			__antithesis_instrumentation__.Notify(86304)
			next := splitAt + width
			if next > len(b.curBuf.entries)-1 {
				__antithesis_instrumentation__.Notify(86306)
				next = len(b.curBuf.entries) - 1
			} else {
				__antithesis_instrumentation__.Notify(86307)
			}
			__antithesis_instrumentation__.Notify(86305)
			predicateAt = next
		} else {
			__antithesis_instrumentation__.Notify(86308)
		}
		__antithesis_instrumentation__.Notify(86297)
		splitKey, err := keys.EnsureSafeSplitKey(b.curBuf.Key(splitAt))
		if err != nil {
			__antithesis_instrumentation__.Notify(86309)
			log.Warningf(ctx, "failed to generate pre-split key for key %s", b.curBuf.Key(splitAt))
			continue
		} else {
			__antithesis_instrumentation__.Notify(86310)
		}
		__antithesis_instrumentation__.Notify(86298)
		predicateKey := b.curBuf.Key(predicateAt)
		log.VEventf(ctx, 1, "pre-splitting span %d of %d at %s", i, b.initialSplits, splitKey)
		if err := b.sink.db.AdminSplit(ctx, splitKey, expire, predicateKey); err != nil {
			__antithesis_instrumentation__.Notify(86311)

			if strings.Contains(err.Error(), "predicate") {
				__antithesis_instrumentation__.Notify(86313)
				log.VEventf(ctx, 1, "%s adder split at %s rejected, had previously split and no longer included %s",
					b.name, splitKey, predicateKey)
				continue
			} else {
				__antithesis_instrumentation__.Notify(86314)
			}
			__antithesis_instrumentation__.Notify(86312)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86315)
		}
		__antithesis_instrumentation__.Notify(86299)
		toScatter = append(toScatter, splitKey)
	}
	__antithesis_instrumentation__.Notify(86292)

	beforeScatters := timeutil.Now()
	splitsWait := beforeScatters.Sub(beforeSplits)
	log.Infof(ctx, "%s adder created %d initial splits in %v from %d keys in %s buffer",
		b.name, len(toScatter), timing(splitsWait), b.curBuf.Len(), b.curBuf.MemSize())
	b.sink.stats.splits += len(toScatter)
	b.sink.stats.splitWait += splitsWait

	for _, splitKey := range toScatter {
		__antithesis_instrumentation__.Notify(86316)
		resp, err := b.sink.db.AdminScatter(ctx, splitKey, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(86318)
			log.Warningf(ctx, "failed to scatter: %v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(86319)
		}
		__antithesis_instrumentation__.Notify(86317)
		b.sink.stats.scatters++
		if resp.MVCCStats != nil {
			__antithesis_instrumentation__.Notify(86320)
			moved := sz(resp.MVCCStats.Total())
			b.sink.stats.scatterMoved += moved
			if moved > 0 {
				__antithesis_instrumentation__.Notify(86321)
				log.VEventf(ctx, 1, "pre-split scattered %s in non-empty range %s",
					moved, resp.RangeInfos[0].Desc.KeySpan().AsRawSpanWithNoLocals())
			} else {
				__antithesis_instrumentation__.Notify(86322)
			}
		} else {
			__antithesis_instrumentation__.Notify(86323)
		}
	}
	__antithesis_instrumentation__.Notify(86293)
	scattersWait := timeutil.Since(beforeScatters)
	b.sink.stats.scatterWait += scattersWait
	log.Infof(ctx, "%s adder scattered %d initial split spans in %v",
		b.name, len(toScatter), timing(scattersWait))

	b.sink.initialSplitDone = true
	return nil
}

func (b *BufferingAdder) GetSummary() roachpb.BulkOpSummary {
	__antithesis_instrumentation__.Notify(86324)
	return b.sink.GetSummary()
}
