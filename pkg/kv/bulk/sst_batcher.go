package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const maxScatterSize = 4 << 20

var (
	tooSmallSSTSize = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_io_write.small_write_size",
		"size below which a 'bulk' write will be performed as a normal write instead",
		400*1<<10,
	)

	ingestDelay = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"bulkio.ingest.flush_delay",
		"amount of time to wait before sending a file to the KV/Storage layer to ingest",
		0,
		settings.NonNegativeDuration,
	)
)

type SSTBatcher struct {
	name     string
	db       *kv.DB
	rc       *rangecache.RangeCache
	settings *cluster.Settings
	mem      mon.BoundAccount

	disallowShadowingBelow hlc.Timestamp

	skipDuplicates bool

	ingestAll bool

	batchTS hlc.Timestamp

	writeAtBatchTS bool

	initialSplitDone bool

	stats         ingestionPerformanceStats
	disableSplits bool

	sstWriter       storage.SSTWriter
	sstFile         *storage.MemFile
	batchStartKey   []byte
	batchEndKey     []byte
	batchEndValue   []byte
	flushKeyChecked bool
	flushKey        roachpb.Key

	lastRange struct {
		span            roachpb.Span
		remaining       sz
		nextExistingKey roachpb.Key
	}

	ms enginepb.MVCCStats

	rowCounter storage.RowCounter

	asyncAddSSTs ctxgroup.Group

	mu struct {
		syncutil.Mutex

		maxWriteTS hlc.Timestamp
		totalRows  roachpb.BulkOpSummary
	}
}

func MakeSSTBatcher(
	ctx context.Context,
	name string,
	db *kv.DB,
	settings *cluster.Settings,
	disallowShadowingBelow hlc.Timestamp,
	writeAtBatchTs bool,
	splitFilledRanges bool,
	mem mon.BoundAccount,
) (*SSTBatcher, error) {
	__antithesis_instrumentation__.Notify(86408)
	b := &SSTBatcher{
		name:                   name,
		db:                     db,
		settings:               settings,
		disallowShadowingBelow: disallowShadowingBelow,
		writeAtBatchTS:         writeAtBatchTs,
		disableSplits:          !splitFilledRanges,
		mem:                    mem,
	}
	err := b.Reset(ctx)
	return b, err
}

func MakeStreamSSTBatcher(
	ctx context.Context, db *kv.DB, settings *cluster.Settings, mem mon.BoundAccount,
) (*SSTBatcher, error) {
	__antithesis_instrumentation__.Notify(86409)
	b := &SSTBatcher{db: db, settings: settings, ingestAll: true, mem: mem}
	err := b.Reset(ctx)
	return b, err
}

func (b *SSTBatcher) updateMVCCStats(key storage.MVCCKey, value []byte) {
	__antithesis_instrumentation__.Notify(86410)
	metaKeySize := int64(len(key.Key)) + 1
	metaValSize := int64(0)
	b.ms.LiveBytes += metaKeySize
	b.ms.LiveCount++
	b.ms.KeyBytes += metaKeySize
	b.ms.ValBytes += metaValSize
	b.ms.KeyCount++

	totalBytes := int64(len(value)) + storage.MVCCVersionTimestampSize
	b.ms.LiveBytes += totalBytes
	b.ms.KeyBytes += storage.MVCCVersionTimestampSize
	b.ms.ValBytes += int64(len(value))
	b.ms.ValCount++
}

func (b *SSTBatcher) AddMVCCKey(ctx context.Context, key storage.MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(86411)
	if len(b.batchEndKey) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(86418)
		return bytes.Equal(b.batchEndKey, key.Key) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(86419)
		return !b.ingestAll == true
	}() == true {
		__antithesis_instrumentation__.Notify(86420)
		if b.skipDuplicates && func() bool {
			__antithesis_instrumentation__.Notify(86422)
			return bytes.Equal(b.batchEndValue, value) == true
		}() == true {
			__antithesis_instrumentation__.Notify(86423)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(86424)
		}
		__antithesis_instrumentation__.Notify(86421)

		err := &kvserverbase.DuplicateKeyError{}
		err.Key = append(err.Key, key.Key...)
		err.Value = append(err.Value, value...)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86425)
	}
	__antithesis_instrumentation__.Notify(86412)

	if err := b.flushIfNeeded(ctx, key.Key); err != nil {
		__antithesis_instrumentation__.Notify(86426)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86427)
	}
	__antithesis_instrumentation__.Notify(86413)

	if b.writeAtBatchTS {
		__antithesis_instrumentation__.Notify(86428)
		if b.batchTS.IsEmpty() {
			__antithesis_instrumentation__.Notify(86430)
			b.batchTS = b.db.Clock().Now()
		} else {
			__antithesis_instrumentation__.Notify(86431)
		}
		__antithesis_instrumentation__.Notify(86429)
		key.Timestamp = b.batchTS
	} else {
		__antithesis_instrumentation__.Notify(86432)
	}
	__antithesis_instrumentation__.Notify(86414)

	if len(b.batchStartKey) == 0 {
		__antithesis_instrumentation__.Notify(86433)
		b.batchStartKey = append(b.batchStartKey[:0], key.Key...)
	} else {
		__antithesis_instrumentation__.Notify(86434)
	}
	__antithesis_instrumentation__.Notify(86415)
	b.batchEndKey = append(b.batchEndKey[:0], key.Key...)
	b.batchEndValue = append(b.batchEndValue[:0], value...)

	if err := b.rowCounter.Count(key.Key); err != nil {
		__antithesis_instrumentation__.Notify(86435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86436)
	}
	__antithesis_instrumentation__.Notify(86416)

	if !b.disallowShadowingBelow.IsEmpty() {
		__antithesis_instrumentation__.Notify(86437)
		b.updateMVCCStats(key, value)
	} else {
		__antithesis_instrumentation__.Notify(86438)
	}
	__antithesis_instrumentation__.Notify(86417)

	return b.sstWriter.Put(key, value)
}

func (b *SSTBatcher) Reset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(86439)
	b.sstWriter.Close()
	b.sstFile = &storage.MemFile{}

	b.sstWriter = storage.MakeIngestionSSTWriter(ctx, b.settings, b.sstFile)
	b.batchStartKey = b.batchStartKey[:0]
	b.batchEndKey = b.batchEndKey[:0]
	b.batchEndValue = b.batchEndValue[:0]
	b.flushKey = nil
	b.flushKeyChecked = false
	b.ms.Reset()

	if b.writeAtBatchTS {
		__antithesis_instrumentation__.Notify(86442)
		b.batchTS = hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(86443)
	}
	__antithesis_instrumentation__.Notify(86440)

	b.rowCounter.BulkOpSummary.Reset()

	if b.stats.sendWaitByStore == nil {
		__antithesis_instrumentation__.Notify(86444)
		b.stats.sendWaitByStore = &sendWaitByStore{timings: make(map[roachpb.StoreID]time.Duration)}
	} else {
		__antithesis_instrumentation__.Notify(86445)
	}
	__antithesis_instrumentation__.Notify(86441)

	return nil
}

const (
	manualFlush = iota
	sizeFlush
	rangeFlush
)

func (b *SSTBatcher) flushIfNeeded(ctx context.Context, nextKey roachpb.Key) error {
	__antithesis_instrumentation__.Notify(86446)

	if !b.flushKeyChecked && func() bool {
		__antithesis_instrumentation__.Notify(86450)
		return b.rc != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(86451)
		b.flushKeyChecked = true
		if k, err := keys.Addr(nextKey); err != nil {
			__antithesis_instrumentation__.Notify(86452)
			log.Warningf(ctx, "failed to get RKey for flush key lookup")
		} else {
			__antithesis_instrumentation__.Notify(86453)
			r := b.rc.GetCached(ctx, k, false)
			if r != nil {
				__antithesis_instrumentation__.Notify(86454)
				k := r.Desc().EndKey.AsRawKey()
				b.flushKey = k
				log.VEventf(ctx, 3, "%s building sstable that will flush before %v", b.name, k)
			} else {
				__antithesis_instrumentation__.Notify(86455)
				log.VEventf(ctx, 2, "%s no cached range desc available to determine sst flush key", b.name)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(86456)
	}
	__antithesis_instrumentation__.Notify(86447)

	shouldFlushDueToRange := b.flushKey != nil && func() bool {
		__antithesis_instrumentation__.Notify(86457)
		return b.flushKey.Compare(nextKey) <= 0 == true
	}() == true

	if shouldFlushDueToRange {
		__antithesis_instrumentation__.Notify(86458)
		if err := b.doFlush(ctx, rangeFlush); err != nil {
			__antithesis_instrumentation__.Notify(86460)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86461)
		}
		__antithesis_instrumentation__.Notify(86459)
		return b.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(86462)
	}
	__antithesis_instrumentation__.Notify(86448)

	if b.sstWriter.DataSize >= ingestFileSize(b.settings) {
		__antithesis_instrumentation__.Notify(86463)
		if err := b.doFlush(ctx, sizeFlush); err != nil {
			__antithesis_instrumentation__.Notify(86465)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86466)
		}
		__antithesis_instrumentation__.Notify(86464)
		return b.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(86467)
	}
	__antithesis_instrumentation__.Notify(86449)
	return nil
}

func (b *SSTBatcher) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(86468)
	if err := b.asyncAddSSTs.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(86472)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86473)
	}
	__antithesis_instrumentation__.Notify(86469)

	b.asyncAddSSTs = ctxgroup.Group{}

	if err := b.doFlush(ctx, manualFlush); err != nil {
		__antithesis_instrumentation__.Notify(86474)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86475)
	}
	__antithesis_instrumentation__.Notify(86470)

	if !b.mu.maxWriteTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(86476)
		if now := b.db.Clock().Now(); now.Less(b.mu.maxWriteTS) {
			__antithesis_instrumentation__.Notify(86478)
			guess := timing(b.mu.maxWriteTS.WallTime - now.WallTime)
			log.VEventf(ctx, 1, "%s batcher waiting %s until max write time %s", b.name, guess, b.mu.maxWriteTS)
			if err := b.db.Clock().SleepUntil(ctx, b.mu.maxWriteTS); err != nil {
				__antithesis_instrumentation__.Notify(86480)
				return err
			} else {
				__antithesis_instrumentation__.Notify(86481)
			}
			__antithesis_instrumentation__.Notify(86479)
			b.stats.commitWait += timeutil.Since(now.GoTime())
		} else {
			__antithesis_instrumentation__.Notify(86482)
		}
		__antithesis_instrumentation__.Notify(86477)
		b.mu.maxWriteTS.Reset()
	} else {
		__antithesis_instrumentation__.Notify(86483)
	}
	__antithesis_instrumentation__.Notify(86471)
	return nil
}

func (b *SSTBatcher) doFlush(ctx context.Context, reason int) error {
	__antithesis_instrumentation__.Notify(86484)
	if b.sstWriter.DataSize == 0 {
		__antithesis_instrumentation__.Notify(86493)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(86494)
	}
	__antithesis_instrumentation__.Notify(86485)
	beforeFlush := timeutil.Now()

	b.stats.batches++

	if delay := ingestDelay.Get(&b.settings.SV); delay != 0 {
		__antithesis_instrumentation__.Notify(86495)
		if delay > time.Second || func() bool {
			__antithesis_instrumentation__.Notify(86497)
			return log.V(1) == true
		}() == true {
			__antithesis_instrumentation__.Notify(86498)
			log.Infof(ctx, "%s delaying %s before flushing ingestion buffer...", b.name, delay)
		} else {
			__antithesis_instrumentation__.Notify(86499)
		}
		__antithesis_instrumentation__.Notify(86496)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(86500)
			return ctx.Err()
		case <-time.After(delay):
			__antithesis_instrumentation__.Notify(86501)
		}
	} else {
		__antithesis_instrumentation__.Notify(86502)
	}
	__antithesis_instrumentation__.Notify(86486)

	if err := b.sstWriter.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(86503)
		return errors.Wrapf(err, "finishing constructed sstable")
	} else {
		__antithesis_instrumentation__.Notify(86504)
	}
	__antithesis_instrumentation__.Notify(86487)

	start := roachpb.Key(append([]byte(nil), b.batchStartKey...))

	end := roachpb.Key(append([]byte(nil), b.batchEndKey...)).Next()

	size := sz(b.sstWriter.DataSize)

	if reason == sizeFlush {
		__antithesis_instrumentation__.Notify(86505)
		log.VEventf(ctx, 3, "%s flushing %s SST due to size > %s", b.name, size, sz(ingestFileSize(b.settings)))
		b.stats.batchesDueToSize++
	} else {
		__antithesis_instrumentation__.Notify(86506)
		if reason == rangeFlush {
			__antithesis_instrumentation__.Notify(86507)
			log.VEventf(ctx, 3, "%s flushing %s SST due to range boundary", b.name, size)
			b.stats.batchesDueToRange++
		} else {
			__antithesis_instrumentation__.Notify(86508)
		}
	}
	__antithesis_instrumentation__.Notify(86488)

	if !b.disableSplits && func() bool {
		__antithesis_instrumentation__.Notify(86509)
		return b.lastRange.span.ContainsKey(start) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(86510)
		return size >= b.lastRange.remaining == true
	}() == true {
		__antithesis_instrumentation__.Notify(86511)
		log.VEventf(ctx, 2, "%s batcher splitting full range before adding file starting at %s",
			b.name, start)

		expire := hlc.Timestamp{WallTime: timeutil.Now().Add(time.Minute * 10).UnixNano()}

		nextKey := b.lastRange.nextExistingKey

		if len(nextKey) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(86513)
			return end.Compare(nextKey) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(86514)
			log.VEventf(ctx, 2, "%s splitting above file span %s at existing key %s",
				b.name, roachpb.Span{Key: start, EndKey: end}, nextKey)

			splitAbove, err := keys.EnsureSafeSplitKey(nextKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(86515)
				log.Warningf(ctx, "%s failed to generate split-above key: %v", b.name, err)
			} else {
				__antithesis_instrumentation__.Notify(86516)
				beforeSplit := timeutil.Now()
				err := b.db.AdminSplit(ctx, splitAbove, expire)
				b.stats.splitWait += timeutil.Since(beforeSplit)
				if err != nil {
					__antithesis_instrumentation__.Notify(86517)
					log.Warningf(ctx, "%s failed to split-above: %v", b.name, err)
				} else {
					__antithesis_instrumentation__.Notify(86518)
					b.stats.splits++
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(86519)
		}
		__antithesis_instrumentation__.Notify(86512)

		splitAt, err := keys.EnsureSafeSplitKey(start)
		if err != nil {
			__antithesis_instrumentation__.Notify(86520)
			log.Warningf(ctx, "%s failed to generate split key: %v", b.name, err)
		} else {
			__antithesis_instrumentation__.Notify(86521)
			beforeSplit := timeutil.Now()
			err := b.db.AdminSplit(ctx, splitAt, expire)
			b.stats.splitWait += timeutil.Since(beforeSplit)
			if err != nil {
				__antithesis_instrumentation__.Notify(86522)
				log.Warningf(ctx, "%s failed to split: %v", b.name, err)
			} else {
				__antithesis_instrumentation__.Notify(86523)
				b.stats.splits++

				beforeScatter := timeutil.Now()
				resp, err := b.db.AdminScatter(ctx, splitAt, maxScatterSize)
				b.stats.scatterWait += timeutil.Since(beforeScatter)
				if err != nil {
					__antithesis_instrumentation__.Notify(86524)

					log.Warningf(ctx, "%s failed to scatter	: %v", b.name, err)
				} else {
					__antithesis_instrumentation__.Notify(86525)
					b.stats.scatters++
					if resp.MVCCStats != nil {
						__antithesis_instrumentation__.Notify(86526)
						moved := sz(resp.MVCCStats.Total())
						b.stats.scatterMoved += moved
						if moved > 0 {
							__antithesis_instrumentation__.Notify(86527)
							log.VEventf(ctx, 1, "%s split scattered %s in non-empty range %s", b.name, moved, resp.RangeInfos[0].Desc.KeySpan().AsRawSpanWithNoLocals())
						} else {
							__antithesis_instrumentation__.Notify(86528)
						}
					} else {
						__antithesis_instrumentation__.Notify(86529)
					}
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(86530)
	}
	__antithesis_instrumentation__.Notify(86489)

	if (b.ms != enginepb.MVCCStats{}) {
		__antithesis_instrumentation__.Notify(86531)
		b.ms.LastUpdateNanos = timeutil.Now().UnixNano()
	} else {
		__antithesis_instrumentation__.Notify(86532)
	}
	__antithesis_instrumentation__.Notify(86490)

	stats := b.ms
	summary := b.rowCounter.BulkOpSummary
	data := b.sstFile.Data()
	batchTS := b.batchTS
	var reserved int64

	updatesLastRange := reason != rangeFlush
	fn := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(86533)
		if err := b.addSSTable(ctx, batchTS, start, end, data, stats, updatesLastRange); err != nil {
			__antithesis_instrumentation__.Notify(86536)
			b.mem.Shrink(ctx, reserved)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86537)
		}
		__antithesis_instrumentation__.Notify(86534)
		b.mu.Lock()
		summary.DataSize += int64(size)
		b.mu.totalRows.Add(summary)
		atomic.AddInt64(&b.stats.batchWaitAtomic, int64(timeutil.Since(beforeFlush)))
		atomic.AddInt64(&b.stats.dataSizeAtomic, int64(size))
		b.mu.Unlock()
		if reserved != 0 {
			__antithesis_instrumentation__.Notify(86538)
			b.mem.Shrink(ctx, reserved)
		} else {
			__antithesis_instrumentation__.Notify(86539)
		}
		__antithesis_instrumentation__.Notify(86535)
		return nil
	}
	__antithesis_instrumentation__.Notify(86491)

	if reason == rangeFlush {
		__antithesis_instrumentation__.Notify(86540)

		if b.asyncAddSSTs == (ctxgroup.Group{}) {
			__antithesis_instrumentation__.Notify(86542)
			b.asyncAddSSTs = ctxgroup.WithContext(ctx)
		} else {
			__antithesis_instrumentation__.Notify(86543)
		}
		__antithesis_instrumentation__.Notify(86541)
		if err := b.mem.Grow(ctx, int64(cap(data))); err != nil {
			__antithesis_instrumentation__.Notify(86544)
			log.VEventf(ctx, 3, "%s unable to reserve enough memory to flush async: %v", b.name, err)
		} else {
			__antithesis_instrumentation__.Notify(86545)
			reserved = int64(cap(data))
			b.asyncAddSSTs.GoCtx(fn)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(86546)
	}
	__antithesis_instrumentation__.Notify(86492)

	return fn(ctx)
}

func (b *SSTBatcher) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(86547)
	b.sstWriter.Close()
	if err := b.asyncAddSSTs.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(86549)
		log.Warningf(ctx, "closing with flushes in-progress encountered an error: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(86550)
	}
	__antithesis_instrumentation__.Notify(86548)
	b.mem.Close(ctx)
}

func (b *SSTBatcher) GetBatchSummary() roachpb.BulkOpSummary {
	__antithesis_instrumentation__.Notify(86551)
	return b.rowCounter.BulkOpSummary
}

func (b *SSTBatcher) GetSummary() roachpb.BulkOpSummary {
	__antithesis_instrumentation__.Notify(86552)
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.mu.totalRows
}

type sstSpan struct {
	start, end roachpb.Key
	sstBytes   []byte
	stats      enginepb.MVCCStats
}

func (b *SSTBatcher) addSSTable(
	ctx context.Context,
	batchTS hlc.Timestamp,
	start, end roachpb.Key,
	sstBytes []byte,
	stats enginepb.MVCCStats,
	updatesLastRange bool,
) error {
	__antithesis_instrumentation__.Notify(86553)
	sendStart := timeutil.Now()
	iter, err := storage.NewMemSSTIterator(sstBytes, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(86557)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86558)
	}
	__antithesis_instrumentation__.Notify(86554)
	defer iter.Close()

	if (stats == enginepb.MVCCStats{}) {
		__antithesis_instrumentation__.Notify(86559)
		stats, err = storage.ComputeStatsForRange(iter, start, end, sendStart.UnixNano())
		if err != nil {
			__antithesis_instrumentation__.Notify(86560)
			return errors.Wrapf(err, "computing stats for SST [%s, %s)", start, end)
		} else {
			__antithesis_instrumentation__.Notify(86561)
		}
	} else {
		__antithesis_instrumentation__.Notify(86562)
	}
	__antithesis_instrumentation__.Notify(86555)

	work := []*sstSpan{{start: start, end: end, sstBytes: sstBytes, stats: stats}}
	var files int
	const maxAddSSTableRetries = 10
	for len(work) > 0 {
		__antithesis_instrumentation__.Notify(86563)
		item := work[0]
		work = work[1:]
		if err := func() error {
			__antithesis_instrumentation__.Notify(86565)
			var err error
			for i := 0; i < maxAddSSTableRetries; i++ {
				__antithesis_instrumentation__.Notify(86567)
				log.VEventf(ctx, 4, "sending %s AddSSTable [%s,%s)", sz(len(item.sstBytes)), item.start, item.end)

				ingestAsWriteBatch := false
				if b.settings != nil && func() bool {
					__antithesis_instrumentation__.Notify(86573)
					return int64(len(item.sstBytes)) < tooSmallSSTSize.Get(&b.settings.SV) == true
				}() == true {
					__antithesis_instrumentation__.Notify(86574)
					log.VEventf(ctx, 3, "ingest data is too small (%d keys/%d bytes) for SSTable, adding via regular batch", item.stats.KeyCount, len(item.sstBytes))
					ingestAsWriteBatch = true
				} else {
					__antithesis_instrumentation__.Notify(86575)
				}
				__antithesis_instrumentation__.Notify(86568)

				req := &roachpb.AddSSTableRequest{
					RequestHeader:                          roachpb.RequestHeader{Key: item.start, EndKey: item.end},
					Data:                                   item.sstBytes,
					DisallowShadowing:                      !b.disallowShadowingBelow.IsEmpty(),
					DisallowShadowingBelow:                 b.disallowShadowingBelow,
					MVCCStats:                              &item.stats,
					IngestAsWrites:                         ingestAsWriteBatch,
					ReturnFollowingLikelyNonEmptySpanStart: true,
				}
				if b.writeAtBatchTS {
					__antithesis_instrumentation__.Notify(86576)
					req.SSTTimestampToRequestTimestamp = batchTS
				} else {
					__antithesis_instrumentation__.Notify(86577)
				}
				__antithesis_instrumentation__.Notify(86569)

				ba := roachpb.BatchRequest{
					Header: roachpb.Header{Timestamp: batchTS, ClientRangeInfo: roachpb.ClientRangeInfo{ExplicitlyRequested: true}},
					AdmissionHeader: roachpb.AdmissionHeader{
						Priority:                 int32(admission.BulkNormalPri),
						CreateTime:               timeutil.Now().UnixNano(),
						Source:                   roachpb.AdmissionHeader_FROM_SQL,
						NoMemoryReservedAtSource: true,
					},
				}
				ba.Add(req)
				beforeSend := timeutil.Now()
				br, pErr := b.db.NonTransactionalSender().Send(ctx, ba)
				sendTime := timeutil.Since(beforeSend)

				atomic.AddInt64(&b.stats.sendWaitAtomic, int64(sendTime))
				if br != nil && func() bool {
					__antithesis_instrumentation__.Notify(86578)
					return len(br.BatchResponse_Header.RangeInfos) > 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(86579)

					b.stats.sendWaitByStore.Lock()
					for i := range br.BatchResponse_Header.RangeInfos {
						__antithesis_instrumentation__.Notify(86581)
						b.stats.sendWaitByStore.timings[br.BatchResponse_Header.RangeInfos[i].Lease.Replica.StoreID] += sendTime
					}
					__antithesis_instrumentation__.Notify(86580)
					b.stats.sendWaitByStore.Unlock()
				} else {
					__antithesis_instrumentation__.Notify(86582)
				}
				__antithesis_instrumentation__.Notify(86570)

				if pErr == nil {
					__antithesis_instrumentation__.Notify(86583)
					resp := br.Responses[0].GetInner().(*roachpb.AddSSTableResponse)
					b.mu.Lock()
					if b.writeAtBatchTS {
						__antithesis_instrumentation__.Notify(86586)
						b.mu.maxWriteTS.Forward(br.Timestamp)
					} else {
						__antithesis_instrumentation__.Notify(86587)
					}
					__antithesis_instrumentation__.Notify(86584)
					b.mu.Unlock()

					if updatesLastRange {
						__antithesis_instrumentation__.Notify(86588)
						b.lastRange.span = resp.RangeSpan
						if resp.RangeSpan.Valid() {
							__antithesis_instrumentation__.Notify(86589)
							b.lastRange.remaining = sz(resp.AvailableBytes)
							b.lastRange.nextExistingKey = resp.FollowingLikelyNonEmptySpanStart
						} else {
							__antithesis_instrumentation__.Notify(86590)
						}
					} else {
						__antithesis_instrumentation__.Notify(86591)
					}
					__antithesis_instrumentation__.Notify(86585)
					files++
					log.VEventf(ctx, 3, "adding %s AddSSTable [%s,%s) took %v", sz(len(item.sstBytes)), item.start, item.end, sendTime)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(86592)
				}
				__antithesis_instrumentation__.Notify(86571)

				err = pErr.GoError()

				if errors.HasType(err, (*roachpb.AmbiguousResultError)(nil)) {
					__antithesis_instrumentation__.Notify(86593)
					log.Warningf(ctx, "addsstable [%s,%s) attempt %d failed: %+v", start, end, i, err)
					continue
				} else {
					__antithesis_instrumentation__.Notify(86594)
				}
				__antithesis_instrumentation__.Notify(86572)

				if m := (*roachpb.RangeKeyMismatchError)(nil); errors.As(err, &m) {
					__antithesis_instrumentation__.Notify(86595)

					mr, err := m.MismatchedRange()
					if err != nil {
						__antithesis_instrumentation__.Notify(86599)
						return err
					} else {
						__antithesis_instrumentation__.Notify(86600)
					}
					__antithesis_instrumentation__.Notify(86596)
					split := mr.Desc.EndKey.AsRawKey()
					log.Infof(ctx, "SSTable cannot be added spanning range bounds %v, retrying...", split)
					left, right, err := createSplitSSTable(ctx, item.start, split, iter, b.settings)
					if err != nil {
						__antithesis_instrumentation__.Notify(86601)
						return err
					} else {
						__antithesis_instrumentation__.Notify(86602)
					}
					__antithesis_instrumentation__.Notify(86597)

					right.stats, err = storage.ComputeStatsForRange(
						iter, right.start, right.end, sendStart.Unix(),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(86603)
						return err
					} else {
						__antithesis_instrumentation__.Notify(86604)
					}
					__antithesis_instrumentation__.Notify(86598)
					left.stats = item.stats
					left.stats.Subtract(right.stats)

					work = append([]*sstSpan{left, right}, work...)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(86605)
				}
			}
			__antithesis_instrumentation__.Notify(86566)
			return err
		}(); err != nil {
			__antithesis_instrumentation__.Notify(86606)
			return errors.Wrapf(err, "addsstable [%s,%s)", item.start, item.end)
		} else {
			__antithesis_instrumentation__.Notify(86607)
		}
		__antithesis_instrumentation__.Notify(86564)

		item.sstBytes = nil
	}
	__antithesis_instrumentation__.Notify(86556)
	atomic.AddInt64(&b.stats.splitRetriesAtomic, int64(files-1))

	log.VEventf(ctx, 3, "AddSSTable [%v, %v) added %d files and took %v", start, end, files, timeutil.Since(sendStart))
	return nil
}

func createSplitSSTable(
	ctx context.Context,
	start, splitKey roachpb.Key,
	iter storage.SimpleMVCCIterator,
	settings *cluster.Settings,
) (*sstSpan, *sstSpan, error) {
	__antithesis_instrumentation__.Notify(86608)
	sstFile := &storage.MemFile{}
	w := storage.MakeIngestionSSTWriter(ctx, settings, sstFile)
	defer w.Close()

	split := false
	var first, last roachpb.Key
	var left, right *sstSpan

	iter.SeekGE(storage.MVCCKey{Key: start})
	for {
		__antithesis_instrumentation__.Notify(86611)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(86616)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(86617)
			if !ok {
				__antithesis_instrumentation__.Notify(86618)
				break
			} else {
				__antithesis_instrumentation__.Notify(86619)
			}
		}
		__antithesis_instrumentation__.Notify(86612)

		key := iter.UnsafeKey()

		if !split && func() bool {
			__antithesis_instrumentation__.Notify(86620)
			return key.Key.Compare(splitKey) >= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(86621)
			err := w.Finish()
			if err != nil {
				__antithesis_instrumentation__.Notify(86623)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(86624)
			}
			__antithesis_instrumentation__.Notify(86622)

			left = &sstSpan{start: first, end: last.PrefixEnd(), sstBytes: sstFile.Data()}
			*sstFile = storage.MemFile{}
			w = storage.MakeIngestionSSTWriter(ctx, settings, sstFile)
			split = true
			first = nil
			last = nil
		} else {
			__antithesis_instrumentation__.Notify(86625)
		}
		__antithesis_instrumentation__.Notify(86613)

		if len(first) == 0 {
			__antithesis_instrumentation__.Notify(86626)
			first = append(first[:0], key.Key...)
		} else {
			__antithesis_instrumentation__.Notify(86627)
		}
		__antithesis_instrumentation__.Notify(86614)
		last = append(last[:0], key.Key...)

		if err := w.Put(key, iter.UnsafeValue()); err != nil {
			__antithesis_instrumentation__.Notify(86628)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(86629)
		}
		__antithesis_instrumentation__.Notify(86615)

		iter.Next()
	}
	__antithesis_instrumentation__.Notify(86609)

	err := w.Finish()
	if err != nil {
		__antithesis_instrumentation__.Notify(86630)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(86631)
	}
	__antithesis_instrumentation__.Notify(86610)
	right = &sstSpan{start: first, end: last.PrefixEnd(), sstBytes: sstFile.Data()}
	return left, right, nil
}
