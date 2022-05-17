package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

var DefaultStorageEngine enginepb.EngineType

func init() {
	_ = DefaultStorageEngine.Set(envutil.EnvOrDefaultString("COCKROACH_STORAGE_ENGINE", "pebble"))
}

type SimpleMVCCIterator interface {
	Close()

	SeekGE(key MVCCKey)

	Valid() (bool, error)

	Next()

	NextKey()

	UnsafeKey() MVCCKey

	UnsafeValue() []byte
}

type IteratorStats struct {
	InternalDeleteSkippedCount int
	TimeBoundNumSSTs           int

	Stats pebble.IteratorStats
}

type MVCCIterator interface {
	SimpleMVCCIterator

	SeekLT(key MVCCKey)

	Prev()

	SeekIntentGE(key roachpb.Key, txnUUID uuid.UUID)

	Key() MVCCKey

	UnsafeRawKey() []byte

	UnsafeRawMVCCKey() []byte

	Value() []byte

	ValueProto(msg protoutil.Message) error

	ComputeStats(start, end roachpb.Key, nowNanos int64) (enginepb.MVCCStats, error)

	FindSplitKey(start, end, minSplitKey roachpb.Key, targetSize int64) (MVCCKey, error)

	SetUpperBound(roachpb.Key)

	Stats() IteratorStats

	SupportsPrev() bool
}

type EngineIterator interface {
	Close()

	SeekEngineKeyGE(key EngineKey) (valid bool, err error)

	SeekEngineKeyLT(key EngineKey) (valid bool, err error)

	NextEngineKey() (valid bool, err error)

	PrevEngineKey() (valid bool, err error)

	UnsafeEngineKey() (EngineKey, error)

	EngineKey() (EngineKey, error)

	UnsafeRawEngineKey() []byte

	UnsafeValue() []byte

	Value() []byte

	SetUpperBound(roachpb.Key)

	GetRawIter() *pebble.Iterator

	SeekEngineKeyGEWithLimit(key EngineKey, limit roachpb.Key) (state pebble.IterValidityState, err error)

	SeekEngineKeyLTWithLimit(key EngineKey, limit roachpb.Key) (state pebble.IterValidityState, err error)

	NextEngineKeyWithLimit(limit roachpb.Key) (state pebble.IterValidityState, err error)

	PrevEngineKeyWithLimit(limit roachpb.Key) (state pebble.IterValidityState, err error)

	Stats() IteratorStats
}

type IterOptions struct {
	Prefix bool

	LowerBound roachpb.Key

	UpperBound roachpb.Key

	WithStats bool

	MinTimestampHint, MaxTimestampHint hlc.Timestamp
}

type MVCCIterKind int

const (
	MVCCKeyAndIntentsIterKind MVCCIterKind = iota

	MVCCKeyIterKind
)

type ExportOptions struct {
	StartKey MVCCKey

	EndKey roachpb.Key

	StartTS, EndTS hlc.Timestamp

	ExportAllRevisions bool

	TargetSize uint64

	MaxSize uint64

	MaxIntents uint64

	StopMidKey bool

	ResourceLimiter ResourceLimiter

	UseTBI bool
}

type Reader interface {
	Close()

	Closed() bool

	ExportMVCCToSst(
		ctx context.Context, exportOptions ExportOptions, dest io.Writer,
	) (_ roachpb.BulkOpSummary, resumeKey roachpb.Key, resumeTS hlc.Timestamp, _ error)

	MVCCGet(key MVCCKey) ([]byte, error)

	MVCCGetProto(key MVCCKey, msg protoutil.Message) (ok bool, keyBytes, valBytes int64, err error)

	MVCCIterate(start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error) error

	NewMVCCIterator(iterKind MVCCIterKind, opts IterOptions) MVCCIterator

	NewEngineIterator(opts IterOptions) EngineIterator

	ConsistentIterators() bool

	PinEngineStateForIterators() error
}

type Writer interface {
	ApplyBatchRepr(repr []byte, sync bool) error

	ClearMVCC(key MVCCKey) error

	ClearUnversioned(key roachpb.Key) error

	ClearIntent(key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID) error

	ClearEngineKey(key EngineKey) error

	ClearRawRange(start, end roachpb.Key) error

	ClearMVCCRangeAndIntents(start, end roachpb.Key) error

	ClearMVCCRange(start, end MVCCKey) error

	ClearIterRange(iter MVCCIterator, start, end roachpb.Key) error

	Merge(key MVCCKey, value []byte) error

	PutMVCC(key MVCCKey, value []byte) error

	PutUnversioned(key roachpb.Key, value []byte) error

	PutIntent(ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID) error

	PutEngineKey(key EngineKey, value []byte) error

	LogData(data []byte) error

	LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails)

	SingleClearEngineKey(key EngineKey) error
}

type ReadWriter interface {
	Reader
	Writer
}

type DurabilityRequirement int8

const (
	StandardDurability DurabilityRequirement = iota

	GuaranteedDurability
)

type Engine interface {
	ReadWriter

	Attrs() roachpb.Attributes

	Capacity() (roachpb.StoreCapacity, error)

	Properties() roachpb.StoreProperties

	Compact() error

	Flush() error

	GetMetrics() Metrics

	GetEncryptionRegistries() (*EncryptionRegistries, error)

	GetEnvStats() (*EnvStats, error)

	GetAuxiliaryDir() string

	NewBatch() Batch

	NewReadOnly(durability DurabilityRequirement) ReadWriter

	NewUnindexedBatch(writeOnly bool) Batch

	NewSnapshot() Reader

	Type() enginepb.EngineType

	IngestExternalFiles(ctx context.Context, paths []string) error

	PreIngestDelay(ctx context.Context)

	ApproximateDiskBytes(from, to roachpb.Key) (uint64, error)

	CompactRange(start, end roachpb.Key) error

	InMem() bool

	RegisterFlushCompletedCallback(cb func())

	fs.FS

	ReadFile(filename string) ([]byte, error)

	WriteFile(filename string, data []byte) error

	CreateCheckpoint(dir string) error

	SetMinVersion(version roachpb.Version) error

	MinVersionIsAtLeastTargetVersion(target roachpb.Version) (bool, error)
}

type Batch interface {
	ReadWriter

	Commit(sync bool) error

	Empty() bool

	Count() uint32

	Len() int

	Repr() []byte
}

type Metrics struct {
	*pebble.Metrics

	WriteStallCount int64

	DiskSlowCount int64

	DiskStallCount int64
}

func (m *Metrics) NumSSTables() int64 {
	__antithesis_instrumentation__.Notify(633648)
	var num int64
	for _, lm := range m.Metrics.Levels {
		__antithesis_instrumentation__.Notify(633650)
		num += lm.NumFiles
	}
	__antithesis_instrumentation__.Notify(633649)
	return num
}

func (m *Metrics) IngestedBytes() uint64 {
	__antithesis_instrumentation__.Notify(633651)
	var ingestedBytes uint64
	for _, lm := range m.Metrics.Levels {
		__antithesis_instrumentation__.Notify(633653)
		ingestedBytes += lm.BytesIngested
	}
	__antithesis_instrumentation__.Notify(633652)
	return ingestedBytes
}

func (m *Metrics) CompactedBytes() (read, written uint64) {
	__antithesis_instrumentation__.Notify(633654)
	for _, lm := range m.Metrics.Levels {
		__antithesis_instrumentation__.Notify(633656)
		read += lm.BytesRead
		written += lm.BytesCompacted
	}
	__antithesis_instrumentation__.Notify(633655)
	return read, written
}

type EnvStats struct {
	TotalFiles uint64

	TotalBytes uint64

	ActiveKeyFiles uint64

	ActiveKeyBytes uint64

	EncryptionType int32

	EncryptionStatus []byte
}

type EncryptionRegistries struct {
	FileRegistry []byte

	KeyRegistry []byte
}

func Scan(reader Reader, start, end roachpb.Key, max int64) ([]MVCCKeyValue, error) {
	__antithesis_instrumentation__.Notify(633657)
	var kvs []MVCCKeyValue
	err := reader.MVCCIterate(start, end, MVCCKeyAndIntentsIterKind, func(kv MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(633659)
		if max != 0 && func() bool {
			__antithesis_instrumentation__.Notify(633661)
			return int64(len(kvs)) >= max == true
		}() == true {
			__antithesis_instrumentation__.Notify(633662)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(633663)
		}
		__antithesis_instrumentation__.Notify(633660)
		kvs = append(kvs, kv)
		return nil
	})
	__antithesis_instrumentation__.Notify(633658)
	return kvs, err
}

func ScanIntents(
	ctx context.Context, reader Reader, start, end roachpb.Key, maxIntents int64, targetBytes int64,
) ([]roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(633664)
	intents := []roachpb.Intent{}

	if bytes.Compare(start, end) >= 0 {
		__antithesis_instrumentation__.Notify(633668)
		return intents, nil
	} else {
		__antithesis_instrumentation__.Notify(633669)
	}
	__antithesis_instrumentation__.Notify(633665)

	ltStart, _ := keys.LockTableSingleKey(start, nil)
	ltEnd, _ := keys.LockTableSingleKey(end, nil)
	iter := reader.NewEngineIterator(IterOptions{LowerBound: ltStart, UpperBound: ltEnd})
	defer iter.Close()

	var meta enginepb.MVCCMetadata
	var intentBytes int64
	var ok bool
	var err error
	for ok, err = iter.SeekEngineKeyGE(EngineKey{Key: ltStart}); ok; ok, err = iter.NextEngineKey() {
		__antithesis_instrumentation__.Notify(633670)
		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(633677)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(633678)
		}
		__antithesis_instrumentation__.Notify(633671)
		if maxIntents != 0 && func() bool {
			__antithesis_instrumentation__.Notify(633679)
			return int64(len(intents)) >= maxIntents == true
		}() == true {
			__antithesis_instrumentation__.Notify(633680)
			break
		} else {
			__antithesis_instrumentation__.Notify(633681)
		}
		__antithesis_instrumentation__.Notify(633672)
		if targetBytes != 0 && func() bool {
			__antithesis_instrumentation__.Notify(633682)
			return intentBytes >= targetBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(633683)
			break
		} else {
			__antithesis_instrumentation__.Notify(633684)
		}
		__antithesis_instrumentation__.Notify(633673)
		key, err := iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(633685)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(633686)
		}
		__antithesis_instrumentation__.Notify(633674)
		lockedKey, err := keys.DecodeLockTableSingleKey(key.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(633687)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(633688)
		}
		__antithesis_instrumentation__.Notify(633675)
		if err = protoutil.Unmarshal(iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(633689)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(633690)
		}
		__antithesis_instrumentation__.Notify(633676)
		intents = append(intents, roachpb.MakeIntent(meta.Txn, lockedKey))
		intentBytes += int64(len(lockedKey)) + int64(len(iter.Value()))
	}
	__antithesis_instrumentation__.Notify(633666)
	if err != nil {
		__antithesis_instrumentation__.Notify(633691)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633692)
	}
	__antithesis_instrumentation__.Notify(633667)
	return intents, nil
}

func WriteSyncNoop(ctx context.Context, eng Engine) error {
	__antithesis_instrumentation__.Notify(633693)
	batch := eng.NewBatch()
	defer batch.Close()

	if err := batch.LogData(nil); err != nil {
		__antithesis_instrumentation__.Notify(633696)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633697)
	}
	__antithesis_instrumentation__.Notify(633694)

	if err := batch.Commit(true); err != nil {
		__antithesis_instrumentation__.Notify(633698)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633699)
	}
	__antithesis_instrumentation__.Notify(633695)
	return nil
}

func ClearRangeWithHeuristic(reader Reader, writer Writer, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(633700)
	iter := reader.NewEngineIterator(IterOptions{UpperBound: end})
	defer iter.Close()

	const clearRangeMinKeys = 64

	count := 0
	valid, err := iter.SeekEngineKeyGE(EngineKey{Key: start})
	for valid {
		__antithesis_instrumentation__.Notify(633705)
		count++
		if count > clearRangeMinKeys {
			__antithesis_instrumentation__.Notify(633707)
			break
		} else {
			__antithesis_instrumentation__.Notify(633708)
		}
		__antithesis_instrumentation__.Notify(633706)
		valid, err = iter.NextEngineKey()
	}
	__antithesis_instrumentation__.Notify(633701)
	if err != nil {
		__antithesis_instrumentation__.Notify(633709)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633710)
	}
	__antithesis_instrumentation__.Notify(633702)
	if count > clearRangeMinKeys {
		__antithesis_instrumentation__.Notify(633711)
		return writer.ClearRawRange(start, end)
	} else {
		__antithesis_instrumentation__.Notify(633712)
	}
	__antithesis_instrumentation__.Notify(633703)
	valid, err = iter.SeekEngineKeyGE(EngineKey{Key: start})
	for valid {
		__antithesis_instrumentation__.Notify(633713)
		var k EngineKey
		if k, err = iter.UnsafeEngineKey(); err != nil {
			__antithesis_instrumentation__.Notify(633716)
			break
		} else {
			__antithesis_instrumentation__.Notify(633717)
		}
		__antithesis_instrumentation__.Notify(633714)
		if err = writer.ClearEngineKey(k); err != nil {
			__antithesis_instrumentation__.Notify(633718)
			break
		} else {
			__antithesis_instrumentation__.Notify(633719)
		}
		__antithesis_instrumentation__.Notify(633715)
		valid, err = iter.NextEngineKey()
	}
	__antithesis_instrumentation__.Notify(633704)
	return err
}

var ingestDelayL0Threshold = settings.RegisterIntSetting(
	settings.TenantWritable,
	"rocksdb.ingest_backpressure.l0_file_count_threshold",
	"number of L0 files after which to backpressure SST ingestions",
	20,
)

var ingestDelayTime = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"rocksdb.ingest_backpressure.max_delay",
	"maximum amount of time to backpressure a single SST ingestion",
	time.Second*5,
)

func preIngestDelay(ctx context.Context, eng Engine, settings *cluster.Settings) {
	__antithesis_instrumentation__.Notify(633720)
	if settings == nil {
		__antithesis_instrumentation__.Notify(633723)
		return
	} else {
		__antithesis_instrumentation__.Notify(633724)
	}
	__antithesis_instrumentation__.Notify(633721)
	metrics := eng.GetMetrics()
	targetDelay := calculatePreIngestDelay(settings, metrics.Metrics)

	if targetDelay == 0 {
		__antithesis_instrumentation__.Notify(633725)
		return
	} else {
		__antithesis_instrumentation__.Notify(633726)
	}
	__antithesis_instrumentation__.Notify(633722)
	log.VEventf(ctx, 2, "delaying SST ingestion %s. %d L0 files, %d L0 Sublevels",
		targetDelay, metrics.Levels[0].NumFiles, metrics.Levels[0].Sublevels)

	select {
	case <-time.After(targetDelay):
		__antithesis_instrumentation__.Notify(633727)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(633728)
	}
}

func calculatePreIngestDelay(settings *cluster.Settings, metrics *pebble.Metrics) time.Duration {
	__antithesis_instrumentation__.Notify(633729)
	maxDelay := ingestDelayTime.Get(&settings.SV)
	l0ReadAmpLimit := ingestDelayL0Threshold.Get(&settings.SV)

	const ramp = 10
	l0ReadAmp := metrics.Levels[0].NumFiles
	if metrics.Levels[0].Sublevels >= 0 {
		__antithesis_instrumentation__.Notify(633732)
		l0ReadAmp = int64(metrics.Levels[0].Sublevels)
	} else {
		__antithesis_instrumentation__.Notify(633733)
	}
	__antithesis_instrumentation__.Notify(633730)

	if l0ReadAmp > l0ReadAmpLimit {
		__antithesis_instrumentation__.Notify(633734)
		delayPerFile := maxDelay / time.Duration(ramp)
		targetDelay := time.Duration(l0ReadAmp-l0ReadAmpLimit) * delayPerFile
		if targetDelay > maxDelay {
			__antithesis_instrumentation__.Notify(633736)
			return maxDelay
		} else {
			__antithesis_instrumentation__.Notify(633737)
		}
		__antithesis_instrumentation__.Notify(633735)
		return targetDelay
	} else {
		__antithesis_instrumentation__.Notify(633738)
	}
	__antithesis_instrumentation__.Notify(633731)
	return 0
}

func iterateOnReader(
	reader Reader, start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(633739)
	if reader.Closed() {
		__antithesis_instrumentation__.Notify(633743)
		return errors.New("cannot call MVCCIterate on a closed batch")
	} else {
		__antithesis_instrumentation__.Notify(633744)
	}
	__antithesis_instrumentation__.Notify(633740)
	if start.Compare(end) >= 0 {
		__antithesis_instrumentation__.Notify(633745)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(633746)
	}
	__antithesis_instrumentation__.Notify(633741)

	it := reader.NewMVCCIterator(iterKind, IterOptions{UpperBound: end})
	defer it.Close()

	it.SeekGE(MakeMVCCMetadataKey(start))
	for ; ; it.Next() {
		__antithesis_instrumentation__.Notify(633747)
		ok, err := it.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(633749)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633750)
			if !ok {
				__antithesis_instrumentation__.Notify(633751)
				break
			} else {
				__antithesis_instrumentation__.Notify(633752)
			}
		}
		__antithesis_instrumentation__.Notify(633748)
		if err := f(MVCCKeyValue{Key: it.Key(), Value: it.Value()}); err != nil {
			__antithesis_instrumentation__.Notify(633753)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(633755)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(633756)
			}
			__antithesis_instrumentation__.Notify(633754)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633757)
		}
	}
	__antithesis_instrumentation__.Notify(633742)
	return nil
}
