package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

const (
	MVCCVersionTimestampSize int64 = 12

	RecommendedMaxOpenFiles = 10000

	MinimumMaxOpenFiles = 1700

	maxIntentsPerWriteIntentErrorDefault = 5000
)

var minWALSyncInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"rocksdb.min_wal_sync_interval",
	"minimum duration between syncs of the RocksDB WAL",
	0*time.Millisecond,
)

var MaxIntentsPerWriteIntentError = settings.RegisterIntSetting(
	settings.TenantWritable,
	"storage.mvcc.max_intents_per_error",
	"maximum number of intents returned in error during export of scan requests",
	maxIntentsPerWriteIntentErrorDefault)

var rocksdbConcurrency = envutil.EnvOrDefaultInt(
	"COCKROACH_ROCKSDB_CONCURRENCY", func() int {
		__antithesis_instrumentation__.Notify(640287)

		const max = 4
		if n := runtime.GOMAXPROCS(0); n <= max {
			__antithesis_instrumentation__.Notify(640289)
			return n
		} else {
			__antithesis_instrumentation__.Notify(640290)
		}
		__antithesis_instrumentation__.Notify(640288)
		return max
	}())

func MakeValue(meta enginepb.MVCCMetadata) roachpb.Value {
	__antithesis_instrumentation__.Notify(640291)
	return roachpb.Value{RawBytes: meta.RawBytes}
}

func emptyKeyError() error {
	__antithesis_instrumentation__.Notify(640292)
	return errors.Errorf("attempted access to empty key")
}

type MVCCKeyValue struct {
	Key   MVCCKey
	Value []byte
}

type optionalValue struct {
	roachpb.Value
	exists bool
}

func makeOptionalValue(v roachpb.Value) optionalValue {
	__antithesis_instrumentation__.Notify(640293)
	return optionalValue{Value: v, exists: true}
}

func (v *optionalValue) IsPresent() bool {
	__antithesis_instrumentation__.Notify(640294)
	return v.exists && func() bool {
		__antithesis_instrumentation__.Notify(640295)
		return v.Value.IsPresent() == true
	}() == true
}

func (v *optionalValue) IsTombstone() bool {
	__antithesis_instrumentation__.Notify(640296)
	return v.exists && func() bool {
		__antithesis_instrumentation__.Notify(640297)
		return !v.Value.IsPresent() == true
	}() == true
}

func (v *optionalValue) ToPointer() *roachpb.Value {
	__antithesis_instrumentation__.Notify(640298)
	if !v.exists {
		__antithesis_instrumentation__.Notify(640300)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(640301)
	}
	__antithesis_instrumentation__.Notify(640299)

	cpy := v.Value
	return &cpy
}

func isSysLocal(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(640302)
	return key.Compare(keys.LocalMax) < 0
}

func isAbortSpanKey(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(640303)
	if !bytes.HasPrefix(key, keys.LocalRangeIDPrefix) {
		__antithesis_instrumentation__.Notify(640306)
		return false
	} else {
		__antithesis_instrumentation__.Notify(640307)
	}
	__antithesis_instrumentation__.Notify(640304)

	_, infix, suffix, _, err := keys.DecodeRangeIDKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(640308)
		return false
	} else {
		__antithesis_instrumentation__.Notify(640309)
	}
	__antithesis_instrumentation__.Notify(640305)
	hasAbortSpanSuffix := infix.Equal(keys.LocalRangeIDReplicatedInfix) && func() bool {
		__antithesis_instrumentation__.Notify(640310)
		return suffix.Equal(keys.LocalAbortSpanSuffix) == true
	}() == true
	return hasAbortSpanSuffix
}

func updateStatsForInline(
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	origMetaKeySize, origMetaValSize, metaKeySize, metaValSize int64,
) {
	__antithesis_instrumentation__.Notify(640311)
	sys := isSysLocal(key)

	if origMetaKeySize != 0 {
		__antithesis_instrumentation__.Notify(640313)
		if sys {
			__antithesis_instrumentation__.Notify(640314)
			ms.SysBytes -= (origMetaKeySize + origMetaValSize)
			ms.SysCount--

			if isAbortSpanKey(key) {
				__antithesis_instrumentation__.Notify(640315)
				ms.AbortSpanBytes -= (origMetaKeySize + origMetaValSize)
			} else {
				__antithesis_instrumentation__.Notify(640316)
			}
		} else {
			__antithesis_instrumentation__.Notify(640317)
			ms.LiveBytes -= (origMetaKeySize + origMetaValSize)
			ms.LiveCount--
			ms.KeyBytes -= origMetaKeySize
			ms.ValBytes -= origMetaValSize
			ms.KeyCount--
			ms.ValCount--
		}
	} else {
		__antithesis_instrumentation__.Notify(640318)
	}
	__antithesis_instrumentation__.Notify(640312)

	if metaKeySize != 0 {
		__antithesis_instrumentation__.Notify(640319)
		if sys {
			__antithesis_instrumentation__.Notify(640320)
			ms.SysBytes += metaKeySize + metaValSize
			ms.SysCount++
			if isAbortSpanKey(key) {
				__antithesis_instrumentation__.Notify(640321)
				ms.AbortSpanBytes += metaKeySize + metaValSize
			} else {
				__antithesis_instrumentation__.Notify(640322)
			}
		} else {
			__antithesis_instrumentation__.Notify(640323)
			ms.LiveBytes += metaKeySize + metaValSize
			ms.LiveCount++
			ms.KeyBytes += metaKeySize
			ms.ValBytes += metaValSize
			ms.KeyCount++
			ms.ValCount++
		}
	} else {
		__antithesis_instrumentation__.Notify(640324)
	}
}

func updateStatsOnMerge(key roachpb.Key, valSize, nowNanos int64) enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(640325)
	var ms enginepb.MVCCStats
	sys := isSysLocal(key)
	ms.AgeTo(nowNanos)

	ms.ContainsEstimates = 1

	if sys {
		__antithesis_instrumentation__.Notify(640327)
		ms.SysBytes += valSize
	} else {
		__antithesis_instrumentation__.Notify(640328)
		ms.LiveBytes += valSize
		ms.ValBytes += valSize
	}
	__antithesis_instrumentation__.Notify(640326)
	return ms
}

func updateStatsOnPut(
	key roachpb.Key,
	prevValSize int64,
	origMetaKeySize, origMetaValSize, metaKeySize, metaValSize int64,
	orig, meta *enginepb.MVCCMetadata,
) enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(640329)
	var ms enginepb.MVCCStats

	if isSysLocal(key) {
		__antithesis_instrumentation__.Notify(640334)

		if orig != nil {
			__antithesis_instrumentation__.Notify(640336)
			ms.SysBytes -= origMetaKeySize + origMetaValSize
			if orig.Txn != nil {
				__antithesis_instrumentation__.Notify(640338)

				ms.SysBytes -= orig.KeyBytes + orig.ValBytes
			} else {
				__antithesis_instrumentation__.Notify(640339)
			}
			__antithesis_instrumentation__.Notify(640337)
			ms.SysCount--
		} else {
			__antithesis_instrumentation__.Notify(640340)
		}
		__antithesis_instrumentation__.Notify(640335)
		ms.SysBytes += meta.KeyBytes + meta.ValBytes + metaKeySize + metaValSize
		ms.SysCount++
		return ms
	} else {
		__antithesis_instrumentation__.Notify(640341)
	}
	__antithesis_instrumentation__.Notify(640330)

	if orig != nil {
		__antithesis_instrumentation__.Notify(640342)
		ms.KeyCount--

		ms.AgeTo(orig.Timestamp.WallTime)

		if orig.Txn != nil {
			__antithesis_instrumentation__.Notify(640347)

			ms.ValCount--
			ms.IntentBytes -= (orig.KeyBytes + orig.ValBytes)
			ms.IntentCount--
			ms.SeparatedIntentCount--
		} else {
			__antithesis_instrumentation__.Notify(640348)
		}
		__antithesis_instrumentation__.Notify(640343)

		if orig.Deleted {
			__antithesis_instrumentation__.Notify(640349)
			ms.KeyBytes -= origMetaKeySize
			ms.ValBytes -= origMetaValSize

			if orig.Txn != nil {
				__antithesis_instrumentation__.Notify(640350)
				ms.KeyBytes -= orig.KeyBytes
				ms.ValBytes -= orig.ValBytes
			} else {
				__antithesis_instrumentation__.Notify(640351)
			}
		} else {
			__antithesis_instrumentation__.Notify(640352)
		}
		__antithesis_instrumentation__.Notify(640344)

		prevIsValue := prevValSize > 0

		if prevIsValue {
			__antithesis_instrumentation__.Notify(640353)

			ms.LiveBytes += MVCCVersionTimestampSize + prevValSize
		} else {
			__antithesis_instrumentation__.Notify(640354)
		}
		__antithesis_instrumentation__.Notify(640345)

		ms.AgeTo(meta.Timestamp.WallTime)

		if prevIsValue {
			__antithesis_instrumentation__.Notify(640355)

			ms.LiveBytes -= MVCCVersionTimestampSize + prevValSize
		} else {
			__antithesis_instrumentation__.Notify(640356)
		}
		__antithesis_instrumentation__.Notify(640346)

		if !orig.Deleted {
			__antithesis_instrumentation__.Notify(640357)
			ms.LiveBytes -= orig.KeyBytes + orig.ValBytes
			ms.LiveBytes -= origMetaKeySize + origMetaValSize
			ms.LiveCount--

			ms.KeyBytes -= origMetaKeySize
			ms.ValBytes -= origMetaValSize

			if orig.Txn != nil {
				__antithesis_instrumentation__.Notify(640358)
				ms.KeyBytes -= orig.KeyBytes
				ms.ValBytes -= orig.ValBytes
			} else {
				__antithesis_instrumentation__.Notify(640359)
			}
		} else {
			__antithesis_instrumentation__.Notify(640360)
		}
	} else {
		__antithesis_instrumentation__.Notify(640361)
		ms.AgeTo(meta.Timestamp.WallTime)
	}
	__antithesis_instrumentation__.Notify(640331)

	if !meta.Deleted {
		__antithesis_instrumentation__.Notify(640362)
		ms.LiveBytes += meta.KeyBytes + meta.ValBytes + metaKeySize + metaValSize
		ms.LiveCount++
	} else {
		__antithesis_instrumentation__.Notify(640363)
	}
	__antithesis_instrumentation__.Notify(640332)
	ms.KeyBytes += meta.KeyBytes + metaKeySize
	ms.ValBytes += meta.ValBytes + metaValSize
	ms.KeyCount++
	ms.ValCount++
	if meta.Txn != nil {
		__antithesis_instrumentation__.Notify(640364)
		ms.IntentBytes += meta.KeyBytes + meta.ValBytes
		ms.IntentCount++
		ms.SeparatedIntentCount++
	} else {
		__antithesis_instrumentation__.Notify(640365)
	}
	__antithesis_instrumentation__.Notify(640333)
	return ms
}

func updateStatsOnResolve(
	key roachpb.Key,
	prevValSize int64,
	origMetaKeySize, origMetaValSize, metaKeySize, metaValSize int64,
	orig, meta *enginepb.MVCCMetadata,
	commit bool,
) enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(640366)
	var ms enginepb.MVCCStats

	if isSysLocal(key) {
		__antithesis_instrumentation__.Notify(640374)

		ms.SysBytes += (metaKeySize + metaValSize) - (origMetaValSize + origMetaKeySize)
		return ms
	} else {
		__antithesis_instrumentation__.Notify(640375)
	}
	__antithesis_instrumentation__.Notify(640367)

	if orig.Deleted != meta.Deleted {
		__antithesis_instrumentation__.Notify(640376)
		log.Fatalf(context.TODO(), "on resolve, original meta was deleted=%t, but new one is deleted=%t",
			orig.Deleted, meta.Deleted)
	} else {
		__antithesis_instrumentation__.Notify(640377)
	}
	__antithesis_instrumentation__.Notify(640368)

	_ = updateStatsOnPut

	ms.AgeTo(orig.Timestamp.WallTime)

	ms.KeyBytes -= origMetaKeySize + orig.KeyBytes
	ms.ValBytes -= origMetaValSize + orig.ValBytes

	if !meta.Deleted {
		__antithesis_instrumentation__.Notify(640378)
		ms.LiveBytes -= origMetaKeySize + origMetaValSize
		ms.LiveBytes -= orig.KeyBytes + meta.ValBytes
	} else {
		__antithesis_instrumentation__.Notify(640379)
	}
	__antithesis_instrumentation__.Notify(640369)

	ms.IntentBytes -= orig.KeyBytes + orig.ValBytes
	ms.IntentCount--
	ms.SeparatedIntentCount--

	_ = updateStatsOnPut
	prevIsValue := prevValSize > 0

	if prevIsValue {
		__antithesis_instrumentation__.Notify(640380)
		ms.LiveBytes += MVCCVersionTimestampSize + prevValSize
	} else {
		__antithesis_instrumentation__.Notify(640381)
	}
	__antithesis_instrumentation__.Notify(640370)

	ms.AgeTo(meta.Timestamp.WallTime)

	if prevIsValue {
		__antithesis_instrumentation__.Notify(640382)

		ms.LiveBytes -= MVCCVersionTimestampSize + prevValSize
	} else {
		__antithesis_instrumentation__.Notify(640383)
	}
	__antithesis_instrumentation__.Notify(640371)

	ms.KeyBytes += metaKeySize + meta.KeyBytes
	ms.ValBytes += metaValSize + meta.ValBytes

	if !meta.Deleted {
		__antithesis_instrumentation__.Notify(640384)
		ms.LiveBytes += (metaKeySize + metaValSize) + (meta.KeyBytes + meta.ValBytes)
	} else {
		__antithesis_instrumentation__.Notify(640385)
	}
	__antithesis_instrumentation__.Notify(640372)

	if !commit {
		__antithesis_instrumentation__.Notify(640386)

		ms.IntentBytes += meta.KeyBytes + meta.ValBytes
		ms.IntentCount++
		ms.SeparatedIntentCount++
	} else {
		__antithesis_instrumentation__.Notify(640387)
	}
	__antithesis_instrumentation__.Notify(640373)
	return ms
}

func updateStatsOnClear(
	key roachpb.Key,
	origMetaKeySize, origMetaValSize, restoredMetaKeySize, restoredMetaValSize int64,
	orig, restored *enginepb.MVCCMetadata,
	restoredNanos int64,
) enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(640388)

	var ms enginepb.MVCCStats

	if isSysLocal(key) {
		__antithesis_instrumentation__.Notify(640393)
		if restored != nil {
			__antithesis_instrumentation__.Notify(640395)
			ms.SysBytes += restoredMetaKeySize + restoredMetaValSize
			ms.SysCount++
		} else {
			__antithesis_instrumentation__.Notify(640396)
		}
		__antithesis_instrumentation__.Notify(640394)

		ms.SysBytes -= (orig.KeyBytes + orig.ValBytes) + (origMetaKeySize + origMetaValSize)
		ms.SysCount--
		return ms
	} else {
		__antithesis_instrumentation__.Notify(640397)
	}
	__antithesis_instrumentation__.Notify(640389)

	if restored != nil {
		__antithesis_instrumentation__.Notify(640398)
		if restored.Txn != nil {
			__antithesis_instrumentation__.Notify(640401)
			panic("restored version should never be an intent")
		} else {
			__antithesis_instrumentation__.Notify(640402)
		}
		__antithesis_instrumentation__.Notify(640399)

		ms.AgeTo(restoredNanos)

		if restored.Deleted {
			__antithesis_instrumentation__.Notify(640403)

			ms.KeyBytes += restoredMetaKeySize
			ms.ValBytes += restoredMetaValSize
		} else {
			__antithesis_instrumentation__.Notify(640404)
		}
		__antithesis_instrumentation__.Notify(640400)

		ms.AgeTo(orig.Timestamp.WallTime)

		ms.KeyCount++

		if !restored.Deleted {
			__antithesis_instrumentation__.Notify(640405)

			ms.KeyBytes += restoredMetaKeySize
			ms.ValBytes += restoredMetaValSize

			ms.LiveBytes += restored.KeyBytes + restored.ValBytes
			ms.LiveCount++
			ms.LiveBytes += restoredMetaKeySize + restoredMetaValSize
		} else {
			__antithesis_instrumentation__.Notify(640406)
		}
	} else {
		__antithesis_instrumentation__.Notify(640407)
		ms.AgeTo(orig.Timestamp.WallTime)
	}
	__antithesis_instrumentation__.Notify(640390)

	if !orig.Deleted {
		__antithesis_instrumentation__.Notify(640408)
		ms.LiveBytes -= (orig.KeyBytes + orig.ValBytes) + (origMetaKeySize + origMetaValSize)
		ms.LiveCount--
	} else {
		__antithesis_instrumentation__.Notify(640409)
	}
	__antithesis_instrumentation__.Notify(640391)
	ms.KeyBytes -= (orig.KeyBytes + origMetaKeySize)
	ms.ValBytes -= (orig.ValBytes + origMetaValSize)
	ms.KeyCount--
	ms.ValCount--
	if orig.Txn != nil {
		__antithesis_instrumentation__.Notify(640410)
		ms.IntentBytes -= (orig.KeyBytes + orig.ValBytes)
		ms.IntentCount--
		ms.SeparatedIntentCount--
	} else {
		__antithesis_instrumentation__.Notify(640411)
	}
	__antithesis_instrumentation__.Notify(640392)
	return ms
}

func updateStatsOnGC(
	key roachpb.Key, keySize, valSize int64, meta *enginepb.MVCCMetadata, nonLiveMS int64,
) enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(640412)
	var ms enginepb.MVCCStats

	if isSysLocal(key) {
		__antithesis_instrumentation__.Notify(640415)
		ms.SysBytes -= (keySize + valSize)
		if meta != nil {
			__antithesis_instrumentation__.Notify(640417)
			ms.SysCount--
		} else {
			__antithesis_instrumentation__.Notify(640418)
		}
		__antithesis_instrumentation__.Notify(640416)
		return ms
	} else {
		__antithesis_instrumentation__.Notify(640419)
	}
	__antithesis_instrumentation__.Notify(640413)

	ms.AgeTo(nonLiveMS)
	ms.KeyBytes -= keySize
	ms.ValBytes -= valSize
	if meta != nil {
		__antithesis_instrumentation__.Notify(640420)
		ms.KeyCount--
	} else {
		__antithesis_instrumentation__.Notify(640421)
		ms.ValCount--
	}
	__antithesis_instrumentation__.Notify(640414)
	return ms
}

func MVCCGetProto(
	ctx context.Context,
	reader Reader,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	msg protoutil.Message,
	opts MVCCGetOptions,
) (bool, error) {
	__antithesis_instrumentation__.Notify(640422)

	value, _, mvccGetErr := MVCCGet(ctx, reader, key, timestamp, opts)
	found := value != nil

	if found && func() bool {
		__antithesis_instrumentation__.Notify(640424)
		return msg != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(640425)

		if err := value.GetProto(msg); err != nil {
			__antithesis_instrumentation__.Notify(640426)
			return found, err
		} else {
			__antithesis_instrumentation__.Notify(640427)
		}
	} else {
		__antithesis_instrumentation__.Notify(640428)
	}
	__antithesis_instrumentation__.Notify(640423)
	return found, mvccGetErr
}

func MVCCPutProto(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	txn *roachpb.Transaction,
	msg protoutil.Message,
) error {
	__antithesis_instrumentation__.Notify(640429)
	value := roachpb.Value{}
	if err := value.SetProto(msg); err != nil {
		__antithesis_instrumentation__.Notify(640431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640432)
	}
	__antithesis_instrumentation__.Notify(640430)
	value.InitChecksum(key)
	return MVCCPut(ctx, rw, ms, key, timestamp, value, txn)
}

func MVCCBlindPutProto(
	ctx context.Context,
	writer Writer,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	msg protoutil.Message,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640433)
	value := roachpb.Value{}
	if err := value.SetProto(msg); err != nil {
		__antithesis_instrumentation__.Notify(640435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640436)
	}
	__antithesis_instrumentation__.Notify(640434)
	value.InitChecksum(key)
	return MVCCBlindPut(ctx, writer, ms, key, timestamp, value, txn)
}

type MVCCGetOptions struct {
	Inconsistent     bool
	Tombstones       bool
	FailOnMoreRecent bool
	Txn              *roachpb.Transaction
	Uncertainty      uncertainty.Interval

	MemoryAccount *mon.BoundAccount
}

func (opts *MVCCGetOptions) validate() error {
	__antithesis_instrumentation__.Notify(640437)
	if opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640440)
		return opts.Txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(640441)
		return errors.Errorf("cannot allow inconsistent reads within a transaction")
	} else {
		__antithesis_instrumentation__.Notify(640442)
	}
	__antithesis_instrumentation__.Notify(640438)
	if opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640443)
		return opts.FailOnMoreRecent == true
	}() == true {
		__antithesis_instrumentation__.Notify(640444)
		return errors.Errorf("cannot allow inconsistent reads with fail on more recent option")
	} else {
		__antithesis_instrumentation__.Notify(640445)
	}
	__antithesis_instrumentation__.Notify(640439)
	return nil
}

func newMVCCIterator(reader Reader, inlineMeta bool, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(640446)
	iterKind := MVCCKeyAndIntentsIterKind
	if inlineMeta {
		__antithesis_instrumentation__.Notify(640448)
		iterKind = MVCCKeyIterKind
	} else {
		__antithesis_instrumentation__.Notify(640449)
	}
	__antithesis_instrumentation__.Notify(640447)
	return reader.NewMVCCIterator(iterKind, opts)
}

func MVCCGet(
	ctx context.Context, reader Reader, key roachpb.Key, timestamp hlc.Timestamp, opts MVCCGetOptions,
) (*roachpb.Value, *roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(640450)
	iter := newMVCCIterator(reader, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()
	value, intent, err := mvccGet(ctx, iter, key, timestamp, opts)
	return value.ToPointer(), intent, err
}

func mvccGet(
	ctx context.Context,
	iter MVCCIterator,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCGetOptions,
) (value optionalValue, intent *roachpb.Intent, err error) {
	__antithesis_instrumentation__.Notify(640451)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(640461)
		return optionalValue{}, nil, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(640462)
	}
	__antithesis_instrumentation__.Notify(640452)
	if timestamp.WallTime < 0 {
		__antithesis_instrumentation__.Notify(640463)
		return optionalValue{}, nil, errors.Errorf("cannot write to %q at timestamp %s", key, timestamp)
	} else {
		__antithesis_instrumentation__.Notify(640464)
	}
	__antithesis_instrumentation__.Notify(640453)
	if err := opts.validate(); err != nil {
		__antithesis_instrumentation__.Notify(640465)
		return optionalValue{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(640466)
	}
	__antithesis_instrumentation__.Notify(640454)

	mvccScanner := pebbleMVCCScannerPool.Get().(*pebbleMVCCScanner)
	defer mvccScanner.release()

	*mvccScanner = pebbleMVCCScanner{
		parent:           iter,
		memAccount:       opts.MemoryAccount,
		start:            key,
		ts:               timestamp,
		maxKeys:          1,
		inconsistent:     opts.Inconsistent,
		tombstones:       opts.Tombstones,
		failOnMoreRecent: opts.FailOnMoreRecent,
		keyBuf:           mvccScanner.keyBuf,
	}

	mvccScanner.init(opts.Txn, opts.Uncertainty, 0)
	mvccScanner.get(ctx)

	traceSpan := tracing.SpanFromContext(ctx)
	recordIteratorStats(traceSpan, mvccScanner.stats())

	if mvccScanner.err != nil {
		__antithesis_instrumentation__.Notify(640467)
		return optionalValue{}, nil, mvccScanner.err
	} else {
		__antithesis_instrumentation__.Notify(640468)
	}
	__antithesis_instrumentation__.Notify(640455)
	intents, err := buildScanIntents(mvccScanner.intentsRepr())
	if err != nil {
		__antithesis_instrumentation__.Notify(640469)
		return optionalValue{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(640470)
	}
	__antithesis_instrumentation__.Notify(640456)
	if !opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640471)
		return len(intents) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(640472)
		return optionalValue{}, nil, &roachpb.WriteIntentError{Intents: intents}
	} else {
		__antithesis_instrumentation__.Notify(640473)
	}
	__antithesis_instrumentation__.Notify(640457)

	if len(intents) > 1 {
		__antithesis_instrumentation__.Notify(640474)
		return optionalValue{}, nil, errors.Errorf("expected 0 or 1 intents, got %d", len(intents))
	} else {
		__antithesis_instrumentation__.Notify(640475)
		if len(intents) == 1 {
			__antithesis_instrumentation__.Notify(640476)
			intent = &intents[0]
		} else {
			__antithesis_instrumentation__.Notify(640477)
		}
	}
	__antithesis_instrumentation__.Notify(640458)

	if len(mvccScanner.results.repr) == 0 {
		__antithesis_instrumentation__.Notify(640478)
		return optionalValue{}, intent, nil
	} else {
		__antithesis_instrumentation__.Notify(640479)
	}
	__antithesis_instrumentation__.Notify(640459)

	mvccKey, rawValue, _, err := MVCCScanDecodeKeyValue(mvccScanner.results.repr)
	if err != nil {
		__antithesis_instrumentation__.Notify(640480)
		return optionalValue{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(640481)
	}
	__antithesis_instrumentation__.Notify(640460)

	value = makeOptionalValue(roachpb.Value{
		RawBytes:  rawValue,
		Timestamp: mvccKey.Timestamp,
	})
	return value, intent, nil
}

func MVCCGetAsTxn(
	ctx context.Context,
	reader Reader,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	txnMeta enginepb.TxnMeta,
) (*roachpb.Value, *roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(640482)
	return MVCCGet(ctx, reader, key, timestamp, MVCCGetOptions{
		Txn: &roachpb.Transaction{
			TxnMeta:                txnMeta,
			Status:                 roachpb.PENDING,
			ReadTimestamp:          txnMeta.WriteTimestamp,
			GlobalUncertaintyLimit: txnMeta.WriteTimestamp,
		}})
}

func mvccGetMetadata(
	iter MVCCIterator, metaKey MVCCKey, iterAlreadyPositioned bool, meta *enginepb.MVCCMetadata,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(640483)
	if iter == nil {
		__antithesis_instrumentation__.Notify(640489)
		return false, 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(640490)
	}
	__antithesis_instrumentation__.Notify(640484)
	if !iterAlreadyPositioned {
		__antithesis_instrumentation__.Notify(640491)
		iter.SeekGE(metaKey)
	} else {
		__antithesis_instrumentation__.Notify(640492)
	}
	__antithesis_instrumentation__.Notify(640485)
	if ok, err := iter.Valid(); !ok {
		__antithesis_instrumentation__.Notify(640493)
		return false, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640494)
	}
	__antithesis_instrumentation__.Notify(640486)

	unsafeKey := iter.UnsafeKey()
	if !unsafeKey.Key.Equal(metaKey.Key) {
		__antithesis_instrumentation__.Notify(640495)
		return false, 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(640496)
	}
	__antithesis_instrumentation__.Notify(640487)

	if !unsafeKey.IsValue() {
		__antithesis_instrumentation__.Notify(640497)
		if err := iter.ValueProto(meta); err != nil {
			__antithesis_instrumentation__.Notify(640499)
			return false, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(640500)
		}
		__antithesis_instrumentation__.Notify(640498)
		return true, int64(unsafeKey.EncodedSize()),
			int64(len(iter.UnsafeValue())), nil
	} else {
		__antithesis_instrumentation__.Notify(640501)
	}
	__antithesis_instrumentation__.Notify(640488)

	meta.Reset()

	meta.KeyBytes = MVCCVersionTimestampSize
	meta.ValBytes = int64(len(iter.UnsafeValue()))
	meta.Deleted = meta.ValBytes == 0
	meta.Timestamp = unsafeKey.Timestamp.ToLegacyTimestamp()
	return true, int64(unsafeKey.EncodedSize()) - meta.KeyBytes, 0, nil
}

type putBuffer struct {
	meta    enginepb.MVCCMetadata
	newMeta enginepb.MVCCMetadata
	ts      hlc.LegacyTimestamp
	tmpbuf  []byte
}

var putBufferPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(640502)
		return &putBuffer{}
	},
}

func newPutBuffer() *putBuffer {
	__antithesis_instrumentation__.Notify(640503)
	return putBufferPool.Get().(*putBuffer)
}

func (b *putBuffer) release() {
	__antithesis_instrumentation__.Notify(640504)
	*b = putBuffer{tmpbuf: b.tmpbuf[:0]}
	putBufferPool.Put(b)
}

func (b *putBuffer) marshalMeta(meta *enginepb.MVCCMetadata) (_ []byte, err error) {
	__antithesis_instrumentation__.Notify(640505)
	size := meta.Size()
	data := b.tmpbuf
	if cap(data) < size {
		__antithesis_instrumentation__.Notify(640508)
		data = make([]byte, size)
	} else {
		__antithesis_instrumentation__.Notify(640509)
		data = data[:size]
	}
	__antithesis_instrumentation__.Notify(640506)
	n, err := protoutil.MarshalTo(meta, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(640510)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(640511)
	}
	__antithesis_instrumentation__.Notify(640507)
	b.tmpbuf = data
	return data[:n], nil
}

func (b *putBuffer) putInlineMeta(
	writer Writer, key MVCCKey, meta *enginepb.MVCCMetadata,
) (keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(640512)
	bytes, err := b.marshalMeta(meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(640515)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640516)
	}
	__antithesis_instrumentation__.Notify(640513)
	if err := writer.PutUnversioned(key.Key, bytes); err != nil {
		__antithesis_instrumentation__.Notify(640517)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640518)
	}
	__antithesis_instrumentation__.Notify(640514)
	return int64(key.EncodedSize()), int64(len(bytes)), nil
}

var trueValue = true

func (b *putBuffer) putIntentMeta(
	ctx context.Context, writer Writer, key MVCCKey, meta *enginepb.MVCCMetadata, alreadyExists bool,
) (keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(640519)
	if meta.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(640524)
		return meta.Timestamp.ToTimestamp() != meta.Txn.WriteTimestamp == true
	}() == true {
		__antithesis_instrumentation__.Notify(640525)

		return 0, 0, errors.AssertionFailedf(
			"meta.Timestamp != meta.Txn.WriteTimestamp: %s != %s", meta.Timestamp, meta.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(640526)
	}
	__antithesis_instrumentation__.Notify(640520)
	if alreadyExists {
		__antithesis_instrumentation__.Notify(640527)

		meta.TxnDidNotUpdateMeta = nil
	} else {
		__antithesis_instrumentation__.Notify(640528)
		meta.TxnDidNotUpdateMeta = &trueValue
	}
	__antithesis_instrumentation__.Notify(640521)
	bytes, err := b.marshalMeta(meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(640529)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640530)
	}
	__antithesis_instrumentation__.Notify(640522)
	if err = writer.PutIntent(ctx, key.Key, bytes, meta.Txn.ID); err != nil {
		__antithesis_instrumentation__.Notify(640531)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640532)
	}
	__antithesis_instrumentation__.Notify(640523)
	return int64(key.EncodedSize()), int64(len(bytes)), nil
}

func MVCCPut(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640533)

	var iter MVCCIterator
	blind := ms == nil && func() bool {
		__antithesis_instrumentation__.Notify(640535)
		return timestamp.IsEmpty() == true
	}() == true
	if !blind {
		__antithesis_instrumentation__.Notify(640536)
		iter = rw.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{Prefix: true})
		defer iter.Close()
	} else {
		__antithesis_instrumentation__.Notify(640537)
	}
	__antithesis_instrumentation__.Notify(640534)
	return mvccPutUsingIter(ctx, rw, iter, ms, key, timestamp, value, txn, nil)
}

func MVCCBlindPut(
	ctx context.Context,
	writer Writer,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640538)
	return mvccPutUsingIter(ctx, writer, nil, ms, key, timestamp, value, txn, nil)
}

func MVCCDelete(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640539)
	iter := newMVCCIterator(rw, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()

	return mvccPutUsingIter(ctx, rw, iter, ms, key, timestamp, noValue, txn, nil)
}

var noValue = roachpb.Value{}

func mvccPutUsingIter(
	ctx context.Context,
	writer Writer,
	iter MVCCIterator,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	txn *roachpb.Transaction,
	valueFn func(optionalValue) ([]byte, error),
) error {
	__antithesis_instrumentation__.Notify(640540)
	var rawBytes []byte
	if valueFn == nil {
		__antithesis_instrumentation__.Notify(640542)
		if !value.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(640544)
			return errors.Errorf("cannot have timestamp set in value on Put")
		} else {
			__antithesis_instrumentation__.Notify(640545)
		}
		__antithesis_instrumentation__.Notify(640543)
		rawBytes = value.RawBytes
	} else {
		__antithesis_instrumentation__.Notify(640546)
	}
	__antithesis_instrumentation__.Notify(640541)

	buf := newPutBuffer()

	err := mvccPutInternal(ctx, writer, iter, ms, key, timestamp, rawBytes,
		txn, buf, valueFn)

	buf.release()
	return err
}

func maybeGetValue(
	ctx context.Context,
	iter MVCCIterator,
	key roachpb.Key,
	value []byte,
	exists bool,
	readTimestamp hlc.Timestamp,
	valueFn func(optionalValue) ([]byte, error),
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(640547)

	if valueFn == nil {
		__antithesis_instrumentation__.Notify(640550)
		return value, nil
	} else {
		__antithesis_instrumentation__.Notify(640551)
	}
	__antithesis_instrumentation__.Notify(640548)
	var exVal optionalValue
	if exists {
		__antithesis_instrumentation__.Notify(640552)
		var err error
		exVal, _, err = mvccGet(ctx, iter, key, readTimestamp, MVCCGetOptions{Tombstones: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(640553)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(640554)
		}
	} else {
		__antithesis_instrumentation__.Notify(640555)
	}
	__antithesis_instrumentation__.Notify(640549)
	return valueFn(exVal)
}

func MVCCScanDecodeKeyValue(repr []byte) (key MVCCKey, value []byte, orepr []byte, err error) {
	__antithesis_instrumentation__.Notify(640556)
	k, ts, value, orepr, err := enginepb.ScanDecodeKeyValue(repr)
	return MVCCKey{k, ts}, value, orepr, err
}

func MVCCScanDecodeKeyValues(repr [][]byte, fn func(key MVCCKey, rawBytes []byte) error) error {
	__antithesis_instrumentation__.Notify(640557)
	var k MVCCKey
	var rawBytes []byte
	var err error
	for _, data := range repr {
		__antithesis_instrumentation__.Notify(640559)
		for len(data) > 0 {
			__antithesis_instrumentation__.Notify(640560)
			k, rawBytes, data, err = MVCCScanDecodeKeyValue(data)
			if err != nil {
				__antithesis_instrumentation__.Notify(640562)
				return err
			} else {
				__antithesis_instrumentation__.Notify(640563)
			}
			__antithesis_instrumentation__.Notify(640561)
			if err = fn(k, rawBytes); err != nil {
				__antithesis_instrumentation__.Notify(640564)
				return err
			} else {
				__antithesis_instrumentation__.Notify(640565)
			}
		}
	}
	__antithesis_instrumentation__.Notify(640558)
	return nil
}

func replayTransactionalWrite(
	ctx context.Context,
	iter MVCCIterator,
	meta *enginepb.MVCCMetadata,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value []byte,
	txn *roachpb.Transaction,
	valueFn func(optionalValue) ([]byte, error),
) error {
	__antithesis_instrumentation__.Notify(640566)
	var found bool
	var writtenValue []byte
	var err error
	if txn.Sequence == meta.Txn.Sequence {
		__antithesis_instrumentation__.Notify(640571)

		exVal, _, err := mvccGet(ctx, iter, key, timestamp, MVCCGetOptions{Txn: txn, Tombstones: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(640573)
			return err
		} else {
			__antithesis_instrumentation__.Notify(640574)
		}
		__antithesis_instrumentation__.Notify(640572)
		writtenValue = exVal.RawBytes
		found = true
	} else {
		__antithesis_instrumentation__.Notify(640575)

		writtenValue, found = meta.GetIntentValue(txn.Sequence)
	}
	__antithesis_instrumentation__.Notify(640567)
	if !found {
		__antithesis_instrumentation__.Notify(640576)

		err := errors.AssertionFailedf("transaction %s with sequence %d missing an intent with lower sequence %d",
			txn.ID, meta.Txn.Sequence, txn.Sequence)
		errWithIssue := errors.WithIssueLink(err,
			errors.IssueLink{
				IssueURL: "https://github.com/cockroachdb/cockroach/issues/71236",
				Detail: "This error may be caused by `DelRange` operation in a batch that also contains a " +
					"write on an intersecting key, as in the case the other write hits a `WriteTooOld` " +
					"error, it is possible for the `DelRange` operation to pick up a new key to delete on " +
					"replay, which will cause sanity checks of intent history to fail as the first iteration " +
					"of the operation would not have placed an intent on this new key.",
			})
		return errWithIssue
	} else {
		__antithesis_instrumentation__.Notify(640577)
	}
	__antithesis_instrumentation__.Notify(640568)

	if valueFn != nil {
		__antithesis_instrumentation__.Notify(640578)
		var exVal optionalValue

		prevIntent, prevValueWritten := meta.GetPrevIntentSeq(txn.Sequence, txn.IgnoredSeqNums)
		if prevValueWritten {
			__antithesis_instrumentation__.Notify(640580)

			prevVal := prevIntent.Value

			exVal = makeOptionalValue(roachpb.Value{RawBytes: prevVal})
		} else {
			__antithesis_instrumentation__.Notify(640581)

			exVal, _, err = mvccGet(ctx, iter, key, timestamp, MVCCGetOptions{Inconsistent: true, Tombstones: true})
			if err != nil {
				__antithesis_instrumentation__.Notify(640582)
				return err
			} else {
				__antithesis_instrumentation__.Notify(640583)
			}
		}
		__antithesis_instrumentation__.Notify(640579)

		value, err = valueFn(exVal)
		if err != nil {
			__antithesis_instrumentation__.Notify(640584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(640585)
		}
	} else {
		__antithesis_instrumentation__.Notify(640586)
	}
	__antithesis_instrumentation__.Notify(640569)

	if !bytes.Equal(value, writtenValue) {
		__antithesis_instrumentation__.Notify(640587)
		return errors.AssertionFailedf("transaction %s with sequence %d has a different value %+v after recomputing from what was written: %+v",
			txn.ID, txn.Sequence, value, writtenValue)
	} else {
		__antithesis_instrumentation__.Notify(640588)
	}
	__antithesis_instrumentation__.Notify(640570)
	return nil
}

func mvccPutInternal(
	ctx context.Context,
	writer Writer,
	iter MVCCIterator,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value []byte,
	txn *roachpb.Transaction,
	buf *putBuffer,
	valueFn func(optionalValue) ([]byte, error),
) error {
	__antithesis_instrumentation__.Notify(640589)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(640602)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(640603)
	}
	__antithesis_instrumentation__.Notify(640590)

	if timestamp.WallTime < 0 {
		__antithesis_instrumentation__.Notify(640604)
		return errors.Errorf("cannot write to %q at timestamp %s", key, timestamp)
	} else {
		__antithesis_instrumentation__.Notify(640605)
	}
	__antithesis_instrumentation__.Notify(640591)

	metaKey := MakeMVCCMetadataKey(key)
	ok, origMetaKeySize, origMetaValSize, err :=
		mvccGetMetadata(iter, metaKey, false, &buf.meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(640606)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640607)
	}
	__antithesis_instrumentation__.Notify(640592)

	putIsInline := timestamp.IsEmpty()
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(640608)
		return putIsInline != buf.meta.IsInline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(640609)
		return errors.Errorf("%q: put is inline=%t, but existing value is inline=%t",
			metaKey, putIsInline, buf.meta.IsInline())
	} else {
		__antithesis_instrumentation__.Notify(640610)
	}
	__antithesis_instrumentation__.Notify(640593)

	if putIsInline {
		__antithesis_instrumentation__.Notify(640611)
		if txn != nil {
			__antithesis_instrumentation__.Notify(640617)
			return errors.Errorf("%q: inline writes not allowed within transactions", metaKey)
		} else {
			__antithesis_instrumentation__.Notify(640618)
		}
		__antithesis_instrumentation__.Notify(640612)
		var metaKeySize, metaValSize int64
		if value, err = maybeGetValue(ctx, iter, key, value, ok, timestamp, valueFn); err != nil {
			__antithesis_instrumentation__.Notify(640619)
			return err
		} else {
			__antithesis_instrumentation__.Notify(640620)
		}
		__antithesis_instrumentation__.Notify(640613)
		if value == nil {
			__antithesis_instrumentation__.Notify(640621)
			metaKeySize, metaValSize, err = 0, 0, writer.ClearUnversioned(metaKey.Key)
		} else {
			__antithesis_instrumentation__.Notify(640622)
			buf.meta = enginepb.MVCCMetadata{RawBytes: value}
			metaKeySize, metaValSize, err = buf.putInlineMeta(writer, metaKey, &buf.meta)
		}
		__antithesis_instrumentation__.Notify(640614)
		if ms != nil {
			__antithesis_instrumentation__.Notify(640623)
			updateStatsForInline(ms, key, origMetaKeySize, origMetaValSize, metaKeySize, metaValSize)
		} else {
			__antithesis_instrumentation__.Notify(640624)
		}
		__antithesis_instrumentation__.Notify(640615)
		if err == nil {
			__antithesis_instrumentation__.Notify(640625)
			writer.LogLogicalOp(MVCCWriteValueOpType, MVCCLogicalOpDetails{
				Key:  key,
				Safe: true,
			})
		} else {
			__antithesis_instrumentation__.Notify(640626)
		}
		__antithesis_instrumentation__.Notify(640616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640627)
	}
	__antithesis_instrumentation__.Notify(640594)

	readTimestamp := timestamp
	writeTimestamp := timestamp
	if txn != nil {
		__antithesis_instrumentation__.Notify(640628)
		readTimestamp = txn.ReadTimestamp
		if readTimestamp != timestamp {
			__antithesis_instrumentation__.Notify(640630)
			return errors.AssertionFailedf(
				"mvccPutInternal: txn's read timestamp %s does not match timestamp %s",
				readTimestamp, timestamp)
		} else {
			__antithesis_instrumentation__.Notify(640631)
		}
		__antithesis_instrumentation__.Notify(640629)
		writeTimestamp = txn.WriteTimestamp
	} else {
		__antithesis_instrumentation__.Notify(640632)
	}
	__antithesis_instrumentation__.Notify(640595)

	timestamp = hlc.Timestamp{}

	logicalOp := MVCCWriteValueOpType
	if txn != nil {
		__antithesis_instrumentation__.Notify(640633)
		logicalOp = MVCCWriteIntentOpType
	} else {
		__antithesis_instrumentation__.Notify(640634)
	}
	__antithesis_instrumentation__.Notify(640596)

	var meta *enginepb.MVCCMetadata
	buf.newMeta = enginepb.MVCCMetadata{
		IntentHistory: buf.meta.IntentHistory,
	}

	var maybeTooOldErr error
	var prevValSize int64
	if ok {
		__antithesis_instrumentation__.Notify(640635)

		meta = &buf.meta
		metaTimestamp := meta.Timestamp.ToTimestamp()

		if meta.Txn != nil {
			__antithesis_instrumentation__.Notify(640636)

			if txn == nil || func() bool {
				__antithesis_instrumentation__.Notify(640642)
				return meta.Txn.ID != txn.ID == true
			}() == true {
				__antithesis_instrumentation__.Notify(640643)

				return &roachpb.WriteIntentError{Intents: []roachpb.Intent{
					roachpb.MakeIntent(meta.Txn, key),
				}}
			} else {
				__antithesis_instrumentation__.Notify(640644)
				if txn.Epoch < meta.Txn.Epoch {
					__antithesis_instrumentation__.Notify(640645)
					return errors.Errorf("put with epoch %d came after put with epoch %d in txn %s",
						txn.Epoch, meta.Txn.Epoch, txn.ID)
				} else {
					__antithesis_instrumentation__.Notify(640646)
					if txn.Epoch == meta.Txn.Epoch && func() bool {
						__antithesis_instrumentation__.Notify(640647)
						return txn.Sequence <= meta.Txn.Sequence == true
					}() == true {
						__antithesis_instrumentation__.Notify(640648)

						return replayTransactionalWrite(ctx, iter, meta, key, readTimestamp, value, txn, valueFn)
					} else {
						__antithesis_instrumentation__.Notify(640649)
					}
				}
			}
			__antithesis_instrumentation__.Notify(640637)

			var exVal optionalValue

			var curProvNotIgnored bool
			if txn.Epoch == meta.Txn.Epoch {
				__antithesis_instrumentation__.Notify(640650)
				if !enginepb.TxnSeqIsIgnored(meta.Txn.Sequence, txn.IgnoredSeqNums) {
					__antithesis_instrumentation__.Notify(640651)

					exVal, _, err = mvccGet(ctx, iter, key, readTimestamp, MVCCGetOptions{Txn: txn, Tombstones: true})
					if err != nil {
						__antithesis_instrumentation__.Notify(640653)
						return err
					} else {
						__antithesis_instrumentation__.Notify(640654)
					}
					__antithesis_instrumentation__.Notify(640652)
					curProvNotIgnored = true
				} else {
					__antithesis_instrumentation__.Notify(640655)

					prevIntent, prevValueWritten := meta.GetPrevIntentSeq(txn.Sequence, txn.IgnoredSeqNums)
					if prevValueWritten {
						__antithesis_instrumentation__.Notify(640656)
						exVal = makeOptionalValue(roachpb.Value{RawBytes: prevIntent.Value})
					} else {
						__antithesis_instrumentation__.Notify(640657)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(640658)
			}
			__antithesis_instrumentation__.Notify(640638)
			if !exVal.exists {
				__antithesis_instrumentation__.Notify(640659)

				exVal, _, err = mvccGet(ctx, iter, key, readTimestamp, MVCCGetOptions{Inconsistent: true, Tombstones: true})
				if err != nil {
					__antithesis_instrumentation__.Notify(640660)
					return err
				} else {
					__antithesis_instrumentation__.Notify(640661)
				}
			} else {
				__antithesis_instrumentation__.Notify(640662)
			}
			__antithesis_instrumentation__.Notify(640639)

			if valueFn != nil {
				__antithesis_instrumentation__.Notify(640663)
				value, err = valueFn(exVal)
				if err != nil {
					__antithesis_instrumentation__.Notify(640664)
					return err
				} else {
					__antithesis_instrumentation__.Notify(640665)
				}
			} else {
				__antithesis_instrumentation__.Notify(640666)
			}
			__antithesis_instrumentation__.Notify(640640)

			logicalOp = MVCCUpdateIntentOpType
			if metaTimestamp.Less(writeTimestamp) {
				__antithesis_instrumentation__.Notify(640667)
				{
					__antithesis_instrumentation__.Notify(640669)

					latestKey := MVCCKey{Key: key, Timestamp: metaTimestamp}
					_, prevUnsafeVal, haveNextVersion, err := unsafeNextVersion(iter, latestKey)
					if err != nil {
						__antithesis_instrumentation__.Notify(640672)
						return err
					} else {
						__antithesis_instrumentation__.Notify(640673)
					}
					__antithesis_instrumentation__.Notify(640670)
					if haveNextVersion {
						__antithesis_instrumentation__.Notify(640674)
						prevValSize = int64(len(prevUnsafeVal))
					} else {
						__antithesis_instrumentation__.Notify(640675)
					}
					__antithesis_instrumentation__.Notify(640671)
					iter = nil
				}
				__antithesis_instrumentation__.Notify(640668)

				versionKey := metaKey
				versionKey.Timestamp = metaTimestamp
				if err := writer.ClearMVCC(versionKey); err != nil {
					__antithesis_instrumentation__.Notify(640676)
					return err
				} else {
					__antithesis_instrumentation__.Notify(640677)
				}
			} else {
				__antithesis_instrumentation__.Notify(640678)
				if writeTimestamp.Less(metaTimestamp) {
					__antithesis_instrumentation__.Notify(640679)

					writeTimestamp = metaTimestamp
				} else {
					__antithesis_instrumentation__.Notify(640680)
				}
			}
			__antithesis_instrumentation__.Notify(640641)

			if txn.Epoch == meta.Txn.Epoch && func() bool {
				__antithesis_instrumentation__.Notify(640681)
				return exVal.exists == true
			}() == true {
				__antithesis_instrumentation__.Notify(640682)

				if curProvNotIgnored {
					__antithesis_instrumentation__.Notify(640683)
					prevIntentValBytes := exVal.RawBytes
					prevIntentSequence := meta.Txn.Sequence
					buf.newMeta.AddToIntentHistory(prevIntentSequence, prevIntentValBytes)
				} else {
					__antithesis_instrumentation__.Notify(640684)
				}
			} else {
				__antithesis_instrumentation__.Notify(640685)
				buf.newMeta.IntentHistory = nil
			}
		} else {
			__antithesis_instrumentation__.Notify(640686)
			if readTimestamp.LessEq(metaTimestamp) {
				__antithesis_instrumentation__.Notify(640687)

				writeTimestamp.Forward(metaTimestamp.Next())
				maybeTooOldErr = roachpb.NewWriteTooOldError(readTimestamp, writeTimestamp, key)

				if txn == nil {
					__antithesis_instrumentation__.Notify(640689)
					readTimestamp = writeTimestamp
				} else {
					__antithesis_instrumentation__.Notify(640690)
				}
				__antithesis_instrumentation__.Notify(640688)
				if value, err = maybeGetValue(ctx, iter, key, value, ok, readTimestamp, valueFn); err != nil {
					__antithesis_instrumentation__.Notify(640691)
					return err
				} else {
					__antithesis_instrumentation__.Notify(640692)
				}
			} else {
				__antithesis_instrumentation__.Notify(640693)
				if value, err = maybeGetValue(ctx, iter, key, value, ok, readTimestamp, valueFn); err != nil {
					__antithesis_instrumentation__.Notify(640694)
					return err
				} else {
					__antithesis_instrumentation__.Notify(640695)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(640696)

		if valueFn != nil {
			__antithesis_instrumentation__.Notify(640697)
			value, err = valueFn(optionalValue{exists: false})
			if err != nil {
				__antithesis_instrumentation__.Notify(640698)
				return err
			} else {
				__antithesis_instrumentation__.Notify(640699)
			}
		} else {
			__antithesis_instrumentation__.Notify(640700)
		}
	}

	{
		__antithesis_instrumentation__.Notify(640701)
		var txnMeta *enginepb.TxnMeta
		if txn != nil {
			__antithesis_instrumentation__.Notify(640703)
			txnMeta = &txn.TxnMeta

			if txnMeta.WriteTimestamp != writeTimestamp {
				__antithesis_instrumentation__.Notify(640704)
				txnMetaCpy := *txnMeta
				txnMetaCpy.WriteTimestamp.Forward(writeTimestamp)
				txnMeta = &txnMetaCpy
			} else {
				__antithesis_instrumentation__.Notify(640705)
			}
		} else {
			__antithesis_instrumentation__.Notify(640706)
		}
		__antithesis_instrumentation__.Notify(640702)
		buf.newMeta.Txn = txnMeta
		buf.newMeta.Timestamp = writeTimestamp.ToLegacyTimestamp()
	}
	__antithesis_instrumentation__.Notify(640597)
	newMeta := &buf.newMeta

	newMeta.KeyBytes = MVCCVersionTimestampSize
	newMeta.ValBytes = int64(len(value))
	newMeta.Deleted = value == nil

	var metaKeySize, metaValSize int64
	if newMeta.Txn != nil {
		__antithesis_instrumentation__.Notify(640707)

		alreadyExists := ok && func() bool {
			__antithesis_instrumentation__.Notify(640708)
			return buf.meta.Txn != nil == true
		}() == true

		metaKeySize, metaValSize, err = buf.putIntentMeta(
			ctx, writer, metaKey, newMeta, alreadyExists)
		if err != nil {
			__antithesis_instrumentation__.Notify(640709)
			return err
		} else {
			__antithesis_instrumentation__.Notify(640710)
		}
	} else {
		__antithesis_instrumentation__.Notify(640711)

		metaKeySize = int64(metaKey.EncodedSize())
	}
	__antithesis_instrumentation__.Notify(640598)

	versionKey := metaKey
	versionKey.Timestamp = writeTimestamp
	if err := writer.PutMVCC(versionKey, value); err != nil {
		__antithesis_instrumentation__.Notify(640712)
		return err
	} else {
		__antithesis_instrumentation__.Notify(640713)
	}
	__antithesis_instrumentation__.Notify(640599)

	if ms != nil {
		__antithesis_instrumentation__.Notify(640714)
		ms.Add(updateStatsOnPut(key, prevValSize, origMetaKeySize, origMetaValSize,
			metaKeySize, metaValSize, meta, newMeta))
	} else {
		__antithesis_instrumentation__.Notify(640715)
	}
	__antithesis_instrumentation__.Notify(640600)

	logicalOpDetails := MVCCLogicalOpDetails{
		Key:       key,
		Timestamp: writeTimestamp,
		Safe:      true,
	}
	if txn := buf.newMeta.Txn; txn != nil {
		__antithesis_instrumentation__.Notify(640716)
		logicalOpDetails.Txn = *txn
	} else {
		__antithesis_instrumentation__.Notify(640717)
	}
	__antithesis_instrumentation__.Notify(640601)
	writer.LogLogicalOp(logicalOp, logicalOpDetails)

	return maybeTooOldErr
}

func MVCCIncrement(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	txn *roachpb.Transaction,
	inc int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(640718)
	iter := newMVCCIterator(rw, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()

	var int64Val int64
	var newInt64Val int64
	err := mvccPutUsingIter(ctx, rw, iter, ms, key, timestamp, noValue, txn, func(value optionalValue) ([]byte, error) {
		__antithesis_instrumentation__.Notify(640720)
		if value.IsPresent() {
			__antithesis_instrumentation__.Notify(640723)
			var err error
			if int64Val, err = value.GetInt(); err != nil {
				__antithesis_instrumentation__.Notify(640724)
				return nil, errors.Errorf("key %q does not contain an integer value", key)
			} else {
				__antithesis_instrumentation__.Notify(640725)
			}
		} else {
			__antithesis_instrumentation__.Notify(640726)
		}
		__antithesis_instrumentation__.Notify(640721)

		if willOverflow(int64Val, inc) {
			__antithesis_instrumentation__.Notify(640727)

			newInt64Val = int64Val
			return nil, &roachpb.IntegerOverflowError{
				Key:            key,
				CurrentValue:   int64Val,
				IncrementValue: inc,
			}
		} else {
			__antithesis_instrumentation__.Notify(640728)
		}
		__antithesis_instrumentation__.Notify(640722)
		newInt64Val = int64Val + inc

		newValue := roachpb.Value{}
		newValue.SetInt(newInt64Val)
		newValue.InitChecksum(key)
		return newValue.RawBytes, nil
	})
	__antithesis_instrumentation__.Notify(640719)

	return newInt64Val, err
}

type CPutMissingBehavior bool

const (
	CPutAllowIfMissing CPutMissingBehavior = true

	CPutFailIfMissing CPutMissingBehavior = false
)

func MVCCConditionalPut(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	expVal []byte,
	allowIfDoesNotExist CPutMissingBehavior,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640729)
	iter := newMVCCIterator(rw, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()

	return mvccConditionalPutUsingIter(ctx, rw, iter, ms, key, timestamp, value, expVal, allowIfDoesNotExist, txn)
}

func MVCCBlindConditionalPut(
	ctx context.Context,
	writer Writer,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	expVal []byte,
	allowIfDoesNotExist CPutMissingBehavior,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640730)
	return mvccConditionalPutUsingIter(ctx, writer, nil, ms, key, timestamp, value, expVal, allowIfDoesNotExist, txn)
}

func mvccConditionalPutUsingIter(
	ctx context.Context,
	writer Writer,
	iter MVCCIterator,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	expBytes []byte,
	allowNoExisting CPutMissingBehavior,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640731)
	return mvccPutUsingIter(
		ctx, writer, iter, ms, key, timestamp, noValue, txn,
		func(existVal optionalValue) ([]byte, error) {
			__antithesis_instrumentation__.Notify(640732)
			if expValPresent, existValPresent := len(expBytes) != 0, existVal.IsPresent(); expValPresent && func() bool {
				__antithesis_instrumentation__.Notify(640734)
				return existValPresent == true
			}() == true {
				__antithesis_instrumentation__.Notify(640735)
				if !bytes.Equal(expBytes, existVal.TagAndDataBytes()) {
					__antithesis_instrumentation__.Notify(640736)
					return nil, &roachpb.ConditionFailedError{
						ActualValue: existVal.ToPointer(),
					}
				} else {
					__antithesis_instrumentation__.Notify(640737)
				}
			} else {
				__antithesis_instrumentation__.Notify(640738)
				if expValPresent != existValPresent && func() bool {
					__antithesis_instrumentation__.Notify(640739)
					return (existValPresent || func() bool {
						__antithesis_instrumentation__.Notify(640740)
						return !bool(allowNoExisting) == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(640741)
					return nil, &roachpb.ConditionFailedError{
						ActualValue: existVal.ToPointer(),
					}
				} else {
					__antithesis_instrumentation__.Notify(640742)
				}
			}
			__antithesis_instrumentation__.Notify(640733)
			return value.RawBytes, nil
		})
}

func MVCCInitPut(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	failOnTombstones bool,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640743)
	iter := newMVCCIterator(rw, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()
	return mvccInitPutUsingIter(ctx, rw, iter, ms, key, timestamp, value, failOnTombstones, txn)
}

func MVCCBlindInitPut(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	failOnTombstones bool,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640744)
	return mvccInitPutUsingIter(ctx, rw, nil, ms, key, timestamp, value, failOnTombstones, txn)
}

func mvccInitPutUsingIter(
	ctx context.Context,
	rw ReadWriter,
	iter MVCCIterator,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
	failOnTombstones bool,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(640745)
	return mvccPutUsingIter(
		ctx, rw, iter, ms, key, timestamp, noValue, txn,
		func(existVal optionalValue) ([]byte, error) {
			__antithesis_instrumentation__.Notify(640746)
			if failOnTombstones && func() bool {
				__antithesis_instrumentation__.Notify(640749)
				return existVal.IsTombstone() == true
			}() == true {
				__antithesis_instrumentation__.Notify(640750)

				return nil, &roachpb.ConditionFailedError{
					ActualValue: existVal.ToPointer(),
				}
			} else {
				__antithesis_instrumentation__.Notify(640751)
			}
			__antithesis_instrumentation__.Notify(640747)
			if existVal.IsPresent() && func() bool {
				__antithesis_instrumentation__.Notify(640752)
				return !existVal.EqualTagAndData(value) == true
			}() == true {
				__antithesis_instrumentation__.Notify(640753)

				return nil, &roachpb.ConditionFailedError{
					ActualValue: existVal.ToPointer(),
				}
			} else {
				__antithesis_instrumentation__.Notify(640754)
			}
			__antithesis_instrumentation__.Notify(640748)
			return value.RawBytes, nil
		})
}

type mvccKeyFormatter struct {
	key MVCCKey
	err error
}

var _ fmt.Formatter = mvccKeyFormatter{}

func (m mvccKeyFormatter) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(640755)
	if m.err != nil {
		__antithesis_instrumentation__.Notify(640757)
		errors.FormatError(m.err, f, c)
		return
	} else {
		__antithesis_instrumentation__.Notify(640758)
	}
	__antithesis_instrumentation__.Notify(640756)
	m.key.Format(f, c)
}

func MVCCMerge(
	_ context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key roachpb.Key,
	timestamp hlc.Timestamp,
	value roachpb.Value,
) error {
	__antithesis_instrumentation__.Notify(640759)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(640763)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(640764)
	}
	__antithesis_instrumentation__.Notify(640760)
	metaKey := MakeMVCCMetadataKey(key)

	buf := newPutBuffer()

	rawBytes := value.RawBytes

	meta := &buf.meta
	*meta = enginepb.MVCCMetadata{RawBytes: rawBytes}

	if !timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(640765)
		buf.ts = timestamp.ToLegacyTimestamp()
		meta.MergeTimestamp = &buf.ts
	} else {
		__antithesis_instrumentation__.Notify(640766)
	}
	__antithesis_instrumentation__.Notify(640761)
	data, err := buf.marshalMeta(meta)
	if err == nil {
		__antithesis_instrumentation__.Notify(640767)
		if err = rw.Merge(metaKey, data); err == nil && func() bool {
			__antithesis_instrumentation__.Notify(640768)
			return ms != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(640769)
			ms.Add(updateStatsOnMerge(
				key, int64(len(rawBytes)), timestamp.WallTime))
		} else {
			__antithesis_instrumentation__.Notify(640770)
		}
	} else {
		__antithesis_instrumentation__.Notify(640771)
	}
	__antithesis_instrumentation__.Notify(640762)
	buf.release()
	return err
}

func MVCCClearTimeRange(
	_ context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key, endKey roachpb.Key,
	startTime, endTime hlc.Timestamp,
	maxBatchSize, maxBatchByteSize int64,
	useTBI bool,
) (*roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(640772)
	var batchSize int64
	var batchByteSize int64
	var resume *roachpb.Span

	const useClearRangeThreshold = 64
	var buf [useClearRangeThreshold]MVCCKey
	var bufSize int
	var clearRangeStart MVCCKey

	if ms == nil {
		__antithesis_instrumentation__.Notify(640778)
		return nil, errors.AssertionFailedf(
			"MVCCStats passed in to MVCCClearTimeRange must be non-nil to ensure proper stats" +
				" computation during Clear operations")
	} else {
		__antithesis_instrumentation__.Notify(640779)
	}
	__antithesis_instrumentation__.Notify(640773)
	clearMatchingKey := func(k MVCCKey) {
		__antithesis_instrumentation__.Notify(640780)
		if len(clearRangeStart.Key) == 0 {
			__antithesis_instrumentation__.Notify(640781)

			if bufSize < useClearRangeThreshold {
				__antithesis_instrumentation__.Notify(640782)
				buf[bufSize].Key = append(buf[bufSize].Key[:0], k.Key...)
				buf[bufSize].Timestamp = k.Timestamp
				bufSize++
			} else {
				__antithesis_instrumentation__.Notify(640783)

				clearRangeStart = buf[0]
				bufSize = 0
			}
		} else {
			__antithesis_instrumentation__.Notify(640784)
		}
	}
	__antithesis_instrumentation__.Notify(640774)

	flushClearedKeys := func(nonMatch MVCCKey) error {
		__antithesis_instrumentation__.Notify(640785)
		if len(clearRangeStart.Key) != 0 {
			__antithesis_instrumentation__.Notify(640787)
			if err := rw.ClearMVCCRange(clearRangeStart, nonMatch); err != nil {
				__antithesis_instrumentation__.Notify(640789)
				return err
			} else {
				__antithesis_instrumentation__.Notify(640790)
			}
			__antithesis_instrumentation__.Notify(640788)
			batchByteSize += int64(clearRangeStart.EncodedSize() + nonMatch.EncodedSize())
			batchSize++
			clearRangeStart = MVCCKey{}
		} else {
			__antithesis_instrumentation__.Notify(640791)
			if bufSize > 0 {
				__antithesis_instrumentation__.Notify(640792)
				var encodedBufSize int64
				for i := 0; i < bufSize; i++ {
					__antithesis_instrumentation__.Notify(640795)
					encodedBufSize += int64(buf[i].EncodedSize())
				}
				__antithesis_instrumentation__.Notify(640793)

				if batchByteSize+encodedBufSize >= maxBatchByteSize {
					__antithesis_instrumentation__.Notify(640796)
					if err := rw.ClearMVCCRange(buf[0], nonMatch); err != nil {
						__antithesis_instrumentation__.Notify(640798)
						return err
					} else {
						__antithesis_instrumentation__.Notify(640799)
					}
					__antithesis_instrumentation__.Notify(640797)
					batchByteSize += int64(buf[0].EncodedSize() + nonMatch.EncodedSize())
					batchSize++
				} else {
					__antithesis_instrumentation__.Notify(640800)
					for i := 0; i < bufSize; i++ {
						__antithesis_instrumentation__.Notify(640802)
						if buf[i].Timestamp.IsEmpty() {
							__antithesis_instrumentation__.Notify(640803)

							if err := rw.ClearUnversioned(buf[i].Key); err != nil {
								__antithesis_instrumentation__.Notify(640804)
								return err
							} else {
								__antithesis_instrumentation__.Notify(640805)
							}
						} else {
							__antithesis_instrumentation__.Notify(640806)
							if err := rw.ClearMVCC(buf[i]); err != nil {
								__antithesis_instrumentation__.Notify(640807)
								return err
							} else {
								__antithesis_instrumentation__.Notify(640808)
							}
						}
					}
					__antithesis_instrumentation__.Notify(640801)
					batchByteSize += encodedBufSize
					batchSize += int64(bufSize)
				}
				__antithesis_instrumentation__.Notify(640794)
				bufSize = 0
			} else {
				__antithesis_instrumentation__.Notify(640809)
			}
		}
		__antithesis_instrumentation__.Notify(640786)
		return nil
	}
	__antithesis_instrumentation__.Notify(640775)

	iter := NewMVCCIncrementalIterator(rw, MVCCIncrementalIterOptions{
		EnableTimeBoundIteratorOptimization: useTBI,
		EndKey:                              endKey,
		StartTime:                           startTime,
		EndTime:                             endTime,
	})
	defer iter.Close()

	var clearedMetaKey MVCCKey
	var clearedMeta enginepb.MVCCMetadata
	var restoredMeta enginepb.MVCCMetadata
	iter.SeekGE(MVCCKey{Key: key})
	for {
		__antithesis_instrumentation__.Notify(640810)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(640813)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(640814)
			if !ok {
				__antithesis_instrumentation__.Notify(640815)
				break
			} else {
				__antithesis_instrumentation__.Notify(640816)
			}
		}
		__antithesis_instrumentation__.Notify(640811)

		k := iter.UnsafeKey()

		if len(clearedMetaKey.Key) > 0 {
			__antithesis_instrumentation__.Notify(640817)
			metaKeySize := int64(clearedMetaKey.EncodedSize())
			if bytes.Equal(clearedMetaKey.Key, k.Key) {
				__antithesis_instrumentation__.Notify(640819)

				valueSize := int64(len(iter.Value()))
				restoredMeta.KeyBytes = MVCCVersionTimestampSize
				restoredMeta.Deleted = valueSize == 0
				restoredMeta.ValBytes = valueSize
				restoredMeta.Timestamp = k.Timestamp.ToLegacyTimestamp()

				ms.Add(updateStatsOnClear(
					clearedMetaKey.Key, metaKeySize, 0, metaKeySize, 0, &clearedMeta, &restoredMeta, k.Timestamp.WallTime,
				))
			} else {
				__antithesis_instrumentation__.Notify(640820)

				ms.Add(updateStatsOnClear(clearedMetaKey.Key, metaKeySize, 0, 0, 0, &clearedMeta, nil, 0))
			}
			__antithesis_instrumentation__.Notify(640818)
			clearedMetaKey.Key = clearedMetaKey.Key[:0]
		} else {
			__antithesis_instrumentation__.Notify(640821)
		}
		__antithesis_instrumentation__.Notify(640812)

		if startTime.Less(k.Timestamp) && func() bool {
			__antithesis_instrumentation__.Notify(640822)
			return k.Timestamp.LessEq(endTime) == true
		}() == true {
			__antithesis_instrumentation__.Notify(640823)
			if batchSize >= maxBatchSize {
				__antithesis_instrumentation__.Notify(640826)
				resume = &roachpb.Span{Key: append([]byte{}, k.Key...), EndKey: endKey}
				break
			} else {
				__antithesis_instrumentation__.Notify(640827)
			}
			__antithesis_instrumentation__.Notify(640824)
			if batchByteSize > maxBatchByteSize {
				__antithesis_instrumentation__.Notify(640828)
				resume = &roachpb.Span{Key: append([]byte{}, k.Key...), EndKey: endKey}
				break
			} else {
				__antithesis_instrumentation__.Notify(640829)
			}
			__antithesis_instrumentation__.Notify(640825)
			clearMatchingKey(k)
			clearedMetaKey.Key = append(clearedMetaKey.Key[:0], k.Key...)
			clearedMeta.KeyBytes = MVCCVersionTimestampSize
			clearedMeta.ValBytes = int64(len(iter.UnsafeValue()))
			clearedMeta.Deleted = clearedMeta.ValBytes == 0
			clearedMeta.Timestamp = k.Timestamp.ToLegacyTimestamp()

			iter.NextIgnoringTime()
		} else {
			__antithesis_instrumentation__.Notify(640830)

			if err := flushClearedKeys(k); err != nil {
				__antithesis_instrumentation__.Notify(640832)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(640833)
			}
			__antithesis_instrumentation__.Notify(640831)

			iter.Next()
		}
	}
	__antithesis_instrumentation__.Notify(640776)

	if len(clearedMetaKey.Key) > 0 {
		__antithesis_instrumentation__.Notify(640834)

		origMetaKeySize := int64(clearedMetaKey.EncodedSize())
		ms.Add(updateStatsOnClear(clearedMetaKey.Key, origMetaKeySize, 0, 0, 0, &clearedMeta, nil, 0))
	} else {
		__antithesis_instrumentation__.Notify(640835)
	}
	__antithesis_instrumentation__.Notify(640777)

	return resume, flushClearedKeys(MVCCKey{Key: endKey})
}

func MVCCDeleteRange(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	key, endKey roachpb.Key,
	max int64,
	timestamp hlc.Timestamp,
	txn *roachpb.Transaction,
	returnKeys bool,
) ([]roachpb.Key, *roachpb.Span, int64, error) {
	__antithesis_instrumentation__.Notify(640836)

	var scanTxn *roachpb.Transaction
	if txn != nil {
		__antithesis_instrumentation__.Notify(640840)
		prevSeqTxn := txn.Clone()
		prevSeqTxn.Sequence--
		scanTxn = prevSeqTxn
	} else {
		__antithesis_instrumentation__.Notify(640841)
	}
	__antithesis_instrumentation__.Notify(640837)
	res, err := MVCCScan(ctx, rw, key, endKey, timestamp, MVCCScanOptions{
		FailOnMoreRecent: true, Txn: scanTxn, MaxKeys: max,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(640842)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(640843)
	}
	__antithesis_instrumentation__.Notify(640838)

	buf := newPutBuffer()
	defer buf.release()
	iter := newMVCCIterator(rw, timestamp.IsEmpty(), IterOptions{Prefix: true})
	defer iter.Close()

	var keys []roachpb.Key
	for i, kv := range res.KVs {
		__antithesis_instrumentation__.Notify(640844)
		if err := mvccPutInternal(ctx, rw, iter, ms, kv.Key, timestamp, nil, txn, buf, nil); err != nil {
			__antithesis_instrumentation__.Notify(640846)
			return nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(640847)
		}
		__antithesis_instrumentation__.Notify(640845)
		if returnKeys {
			__antithesis_instrumentation__.Notify(640848)
			if i == 0 {
				__antithesis_instrumentation__.Notify(640850)
				keys = make([]roachpb.Key, len(res.KVs))
			} else {
				__antithesis_instrumentation__.Notify(640851)
			}
			__antithesis_instrumentation__.Notify(640849)
			keys[i] = kv.Key
		} else {
			__antithesis_instrumentation__.Notify(640852)
		}
	}
	__antithesis_instrumentation__.Notify(640839)
	return keys, res.ResumeSpan, res.NumKeys, nil
}

func recordIteratorStats(traceSpan *tracing.Span, iteratorStats IteratorStats) {
	__antithesis_instrumentation__.Notify(640853)
	stats := iteratorStats.Stats
	if traceSpan != nil {
		__antithesis_instrumentation__.Notify(640854)
		steps := stats.ReverseStepCount[pebble.InterfaceCall] + stats.ForwardStepCount[pebble.InterfaceCall]
		seeks := stats.ReverseSeekCount[pebble.InterfaceCall] + stats.ForwardSeekCount[pebble.InterfaceCall]
		internalSteps := stats.ReverseStepCount[pebble.InternalIterCall] + stats.ForwardStepCount[pebble.InternalIterCall]
		internalSeeks := stats.ReverseSeekCount[pebble.InternalIterCall] + stats.ForwardSeekCount[pebble.InternalIterCall]
		traceSpan.RecordStructured(&roachpb.ScanStats{
			NumInterfaceSeeks:              uint64(seeks),
			NumInternalSeeks:               uint64(internalSeeks),
			NumInterfaceSteps:              uint64(steps),
			NumInternalSteps:               uint64(internalSteps),
			BlockBytes:                     stats.InternalStats.BlockBytes,
			BlockBytesInCache:              stats.InternalStats.BlockBytesInCache,
			KeyBytes:                       stats.InternalStats.KeyBytes,
			ValueBytes:                     stats.InternalStats.ValueBytes,
			PointCount:                     stats.InternalStats.PointCount,
			PointsCoveredByRangeTombstones: stats.InternalStats.PointsCoveredByRangeTombstones,
		})
	} else {
		__antithesis_instrumentation__.Notify(640855)
	}
}

func mvccScanToBytes(
	ctx context.Context,
	iter MVCCIterator,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCScanOptions,
) (MVCCScanResult, error) {
	__antithesis_instrumentation__.Notify(640856)
	if len(endKey) == 0 {
		__antithesis_instrumentation__.Notify(640865)
		return MVCCScanResult{}, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(640866)
	}
	__antithesis_instrumentation__.Notify(640857)
	if err := opts.validate(); err != nil {
		__antithesis_instrumentation__.Notify(640867)
		return MVCCScanResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(640868)
	}
	__antithesis_instrumentation__.Notify(640858)
	if opts.MaxKeys < 0 {
		__antithesis_instrumentation__.Notify(640869)
		return MVCCScanResult{
			ResumeSpan:   &roachpb.Span{Key: key, EndKey: endKey},
			ResumeReason: roachpb.RESUME_KEY_LIMIT,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(640870)
	}
	__antithesis_instrumentation__.Notify(640859)
	if opts.TargetBytes < 0 {
		__antithesis_instrumentation__.Notify(640871)
		return MVCCScanResult{
			ResumeSpan:   &roachpb.Span{Key: key, EndKey: endKey},
			ResumeReason: roachpb.RESUME_BYTE_LIMIT,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(640872)
	}
	__antithesis_instrumentation__.Notify(640860)

	mvccScanner := pebbleMVCCScannerPool.Get().(*pebbleMVCCScanner)
	defer mvccScanner.release()

	*mvccScanner = pebbleMVCCScanner{
		parent:                 iter,
		memAccount:             opts.MemoryAccount,
		reverse:                opts.Reverse,
		start:                  key,
		end:                    endKey,
		ts:                     timestamp,
		maxKeys:                opts.MaxKeys,
		targetBytes:            opts.TargetBytes,
		targetBytesAvoidExcess: opts.TargetBytesAvoidExcess,
		allowEmpty:             opts.AllowEmpty,
		wholeRows:              opts.WholeRowsOfSize > 1,
		maxIntents:             opts.MaxIntents,
		inconsistent:           opts.Inconsistent,
		tombstones:             opts.Tombstones,
		failOnMoreRecent:       opts.FailOnMoreRecent,
		keyBuf:                 mvccScanner.keyBuf,
	}

	var trackLastOffsets int
	if opts.WholeRowsOfSize > 1 {
		__antithesis_instrumentation__.Notify(640873)
		trackLastOffsets = int(opts.WholeRowsOfSize)
	} else {
		__antithesis_instrumentation__.Notify(640874)
	}
	__antithesis_instrumentation__.Notify(640861)
	mvccScanner.init(opts.Txn, opts.Uncertainty, trackLastOffsets)

	var res MVCCScanResult
	var err error
	res.ResumeSpan, res.ResumeReason, res.ResumeNextBytes, err = mvccScanner.scan(ctx)

	if err != nil {
		__antithesis_instrumentation__.Notify(640875)
		return MVCCScanResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(640876)
	}
	__antithesis_instrumentation__.Notify(640862)

	res.KVData = mvccScanner.results.finish()
	res.NumKeys = mvccScanner.results.count
	res.NumBytes = mvccScanner.results.bytes

	traceSpan := tracing.SpanFromContext(ctx)

	recordIteratorStats(traceSpan, mvccScanner.stats())

	res.Intents, err = buildScanIntents(mvccScanner.intentsRepr())
	if err != nil {
		__antithesis_instrumentation__.Notify(640877)
		return MVCCScanResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(640878)
	}
	__antithesis_instrumentation__.Notify(640863)

	if !opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640879)
		return len(res.Intents) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(640880)
		return MVCCScanResult{}, &roachpb.WriteIntentError{Intents: res.Intents}
	} else {
		__antithesis_instrumentation__.Notify(640881)
	}
	__antithesis_instrumentation__.Notify(640864)
	return res, nil
}

func mvccScanToKvs(
	ctx context.Context,
	iter MVCCIterator,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCScanOptions,
) (MVCCScanResult, error) {
	__antithesis_instrumentation__.Notify(640882)
	res, err := mvccScanToBytes(ctx, iter, key, endKey, timestamp, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(640885)
		return MVCCScanResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(640886)
	}
	__antithesis_instrumentation__.Notify(640883)
	res.KVs = make([]roachpb.KeyValue, res.NumKeys)
	kvData := res.KVData
	res.KVData = nil

	var i int
	if err := MVCCScanDecodeKeyValues(kvData, func(key MVCCKey, rawBytes []byte) error {
		__antithesis_instrumentation__.Notify(640887)
		res.KVs[i].Key = key.Key
		res.KVs[i].Value.RawBytes = rawBytes
		res.KVs[i].Value.Timestamp = key.Timestamp
		i++
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(640888)
		return MVCCScanResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(640889)
	}
	__antithesis_instrumentation__.Notify(640884)
	return res, err
}

func buildScanIntents(data []byte) ([]roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(640890)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(640895)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(640896)
	}
	__antithesis_instrumentation__.Notify(640891)

	reader, err := NewRocksDBBatchReader(data)
	if err != nil {
		__antithesis_instrumentation__.Notify(640897)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(640898)
	}
	__antithesis_instrumentation__.Notify(640892)

	intents := make([]roachpb.Intent, 0, reader.Count())
	var meta enginepb.MVCCMetadata
	for reader.Next() {
		__antithesis_instrumentation__.Notify(640899)
		key, err := reader.MVCCKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(640902)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(640903)
		}
		__antithesis_instrumentation__.Notify(640900)
		if err := protoutil.Unmarshal(reader.Value(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(640904)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(640905)
		}
		__antithesis_instrumentation__.Notify(640901)
		intents = append(intents, roachpb.MakeIntent(meta.Txn, key.Key))
	}
	__antithesis_instrumentation__.Notify(640893)

	if err := reader.Error(); err != nil {
		__antithesis_instrumentation__.Notify(640906)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(640907)
	}
	__antithesis_instrumentation__.Notify(640894)
	return intents, nil
}

type MVCCScanOptions struct {
	Inconsistent     bool
	Tombstones       bool
	Reverse          bool
	FailOnMoreRecent bool
	Txn              *roachpb.Transaction
	Uncertainty      uncertainty.Interval

	MaxKeys int64

	TargetBytes int64

	TargetBytesAvoidExcess bool

	AllowEmpty bool

	WholeRowsOfSize int32

	MaxIntents int64

	MemoryAccount *mon.BoundAccount
}

func (opts *MVCCScanOptions) validate() error {
	__antithesis_instrumentation__.Notify(640908)
	if opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640911)
		return opts.Txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(640912)
		return errors.Errorf("cannot allow inconsistent reads within a transaction")
	} else {
		__antithesis_instrumentation__.Notify(640913)
	}
	__antithesis_instrumentation__.Notify(640909)
	if opts.Inconsistent && func() bool {
		__antithesis_instrumentation__.Notify(640914)
		return opts.FailOnMoreRecent == true
	}() == true {
		__antithesis_instrumentation__.Notify(640915)
		return errors.Errorf("cannot allow inconsistent reads with fail on more recent option")
	} else {
		__antithesis_instrumentation__.Notify(640916)
	}
	__antithesis_instrumentation__.Notify(640910)
	return nil
}

type MVCCScanResult struct {
	KVData  [][]byte
	KVs     []roachpb.KeyValue
	NumKeys int64

	NumBytes int64

	ResumeSpan      *roachpb.Span
	ResumeReason    roachpb.ResumeReason
	ResumeNextBytes int64
	Intents         []roachpb.Intent
}

func MVCCScan(
	ctx context.Context,
	reader Reader,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCScanOptions,
) (MVCCScanResult, error) {
	__antithesis_instrumentation__.Notify(640917)
	iter := newMVCCIterator(reader, timestamp.IsEmpty(), IterOptions{LowerBound: key, UpperBound: endKey})
	defer iter.Close()
	return mvccScanToKvs(ctx, iter, key, endKey, timestamp, opts)
}

func MVCCScanToBytes(
	ctx context.Context,
	reader Reader,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCScanOptions,
) (MVCCScanResult, error) {
	__antithesis_instrumentation__.Notify(640918)
	iter := newMVCCIterator(reader, timestamp.IsEmpty(), IterOptions{LowerBound: key, UpperBound: endKey})
	defer iter.Close()
	return mvccScanToBytes(ctx, iter, key, endKey, timestamp, opts)
}

func MVCCScanAsTxn(
	ctx context.Context,
	reader Reader,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	txnMeta enginepb.TxnMeta,
) (MVCCScanResult, error) {
	__antithesis_instrumentation__.Notify(640919)
	return MVCCScan(ctx, reader, key, endKey, timestamp, MVCCScanOptions{
		Txn: &roachpb.Transaction{
			TxnMeta:                txnMeta,
			Status:                 roachpb.PENDING,
			ReadTimestamp:          txnMeta.WriteTimestamp,
			GlobalUncertaintyLimit: txnMeta.WriteTimestamp,
		}})
}

func MVCCIterate(
	ctx context.Context,
	reader Reader,
	key, endKey roachpb.Key,
	timestamp hlc.Timestamp,
	opts MVCCScanOptions,
	f func(roachpb.KeyValue) error,
) ([]roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(640920)
	iter := newMVCCIterator(
		reader, timestamp.IsEmpty(), IterOptions{LowerBound: key, UpperBound: endKey})
	defer iter.Close()

	var intents []roachpb.Intent
	for {
		__antithesis_instrumentation__.Notify(640922)
		const maxKeysPerScan = 1000
		opts := opts
		opts.MaxKeys = maxKeysPerScan
		res, err := mvccScanToKvs(
			ctx, iter, key, endKey, timestamp, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(640927)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(640928)
		}
		__antithesis_instrumentation__.Notify(640923)

		if len(res.Intents) > 0 {
			__antithesis_instrumentation__.Notify(640929)
			if intents == nil {
				__antithesis_instrumentation__.Notify(640930)
				intents = res.Intents
			} else {
				__antithesis_instrumentation__.Notify(640931)
				intents = append(intents, res.Intents...)
			}
		} else {
			__antithesis_instrumentation__.Notify(640932)
		}
		__antithesis_instrumentation__.Notify(640924)

		for i := range res.KVs {
			__antithesis_instrumentation__.Notify(640933)
			if err := f(res.KVs[i]); err != nil {
				__antithesis_instrumentation__.Notify(640934)
				if iterutil.Done(err) {
					__antithesis_instrumentation__.Notify(640936)
					return intents, nil
				} else {
					__antithesis_instrumentation__.Notify(640937)
				}
				__antithesis_instrumentation__.Notify(640935)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(640938)
			}
		}
		__antithesis_instrumentation__.Notify(640925)

		if res.ResumeSpan == nil {
			__antithesis_instrumentation__.Notify(640939)
			break
		} else {
			__antithesis_instrumentation__.Notify(640940)
		}
		__antithesis_instrumentation__.Notify(640926)
		if opts.Reverse {
			__antithesis_instrumentation__.Notify(640941)
			endKey = res.ResumeSpan.EndKey
		} else {
			__antithesis_instrumentation__.Notify(640942)
			key = res.ResumeSpan.Key
		}
	}
	__antithesis_instrumentation__.Notify(640921)

	return intents, nil
}

func MVCCResolveWriteIntent(
	ctx context.Context, rw ReadWriter, ms *enginepb.MVCCStats, intent roachpb.LockUpdate,
) (bool, error) {
	__antithesis_instrumentation__.Notify(640943)
	if len(intent.Key) == 0 {
		__antithesis_instrumentation__.Notify(640946)
		return false, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(640947)
	}
	__antithesis_instrumentation__.Notify(640944)
	if len(intent.EndKey) > 0 {
		__antithesis_instrumentation__.Notify(640948)
		return false, errors.Errorf("can't resolve range intent as point intent")
	} else {
		__antithesis_instrumentation__.Notify(640949)
	}
	__antithesis_instrumentation__.Notify(640945)

	iterAndBuf := GetBufUsingIter(rw.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{Prefix: true}))
	iterAndBuf.iter.SeekIntentGE(intent.Key, intent.Txn.ID)
	ok, err := mvccResolveWriteIntent(ctx, rw, iterAndBuf.iter, ms, intent, iterAndBuf.buf)

	iterAndBuf.Cleanup()
	return ok, err
}

func unsafeNextVersion(iter MVCCIterator, latestKey MVCCKey) (MVCCKey, []byte, bool, error) {
	__antithesis_instrumentation__.Notify(640950)

	nextKey := latestKey.Next()
	iter.SeekGE(nextKey)

	if ok, err := iter.Valid(); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(640953)
		return !ok == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(640954)
		return !iter.UnsafeKey().Key.Equal(latestKey.Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(640955)
		return MVCCKey{}, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(640956)
	}
	__antithesis_instrumentation__.Notify(640951)
	unsafeKey := iter.UnsafeKey()
	if !unsafeKey.IsValue() {
		__antithesis_instrumentation__.Notify(640957)
		return MVCCKey{}, nil, false, errors.Errorf("expected an MVCC value key: %s", unsafeKey)
	} else {
		__antithesis_instrumentation__.Notify(640958)
	}
	__antithesis_instrumentation__.Notify(640952)
	return unsafeKey, iter.UnsafeValue(), true, nil
}

type iterForKeyVersions interface {
	Valid() (bool, error)
	SeekGE(key MVCCKey)
	Next()
	UnsafeKey() MVCCKey
	UnsafeValue() []byte
	ValueProto(msg protoutil.Message) error
}

type separatedIntentAndVersionIter struct {
	engineIter EngineIterator
	mvccIter   MVCCIterator

	meta            *enginepb.MVCCMetadata
	atMVCCIter      bool
	engineIterValid bool
	engineIterErr   error
	intentKey       roachpb.Key
}

var _ iterForKeyVersions = &separatedIntentAndVersionIter{}

func (s *separatedIntentAndVersionIter) seekEngineKeyGE(key EngineKey) {
	__antithesis_instrumentation__.Notify(640959)
	s.atMVCCIter = false
	s.meta = nil
	s.engineIterValid, s.engineIterErr = s.engineIter.SeekEngineKeyGE(key)
	s.initIntentKey()
}

func (s *separatedIntentAndVersionIter) nextEngineKey() {
	__antithesis_instrumentation__.Notify(640960)
	s.atMVCCIter = false
	s.meta = nil
	s.engineIterValid, s.engineIterErr = s.engineIter.NextEngineKey()
	s.initIntentKey()
}

func (s *separatedIntentAndVersionIter) initIntentKey() {
	__antithesis_instrumentation__.Notify(640961)
	if s.engineIterValid {
		__antithesis_instrumentation__.Notify(640962)
		engineKey, err := s.engineIter.UnsafeEngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(640964)
			s.engineIterErr = err
			s.engineIterValid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(640965)
		}
		__antithesis_instrumentation__.Notify(640963)
		if s.intentKey, err = keys.DecodeLockTableSingleKey(engineKey.Key); err != nil {
			__antithesis_instrumentation__.Notify(640966)
			s.engineIterErr = err
			s.engineIterValid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(640967)
		}
	} else {
		__antithesis_instrumentation__.Notify(640968)
	}
}

func (s *separatedIntentAndVersionIter) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(640969)
	if s.atMVCCIter {
		__antithesis_instrumentation__.Notify(640971)
		return s.mvccIter.Valid()
	} else {
		__antithesis_instrumentation__.Notify(640972)
	}
	__antithesis_instrumentation__.Notify(640970)
	return s.engineIterValid, s.engineIterErr
}

func (s *separatedIntentAndVersionIter) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(640973)
	if !key.IsValue() {
		__antithesis_instrumentation__.Notify(640975)
		panic(errors.AssertionFailedf("SeekGE only permitted for values"))
	} else {
		__antithesis_instrumentation__.Notify(640976)
	}
	__antithesis_instrumentation__.Notify(640974)
	s.mvccIter.SeekGE(key)
	s.atMVCCIter = true
}

func (s *separatedIntentAndVersionIter) Next() {
	__antithesis_instrumentation__.Notify(640977)
	if !s.atMVCCIter {
		__antithesis_instrumentation__.Notify(640979)
		panic(errors.AssertionFailedf("Next not preceded by SeekGE"))
	} else {
		__antithesis_instrumentation__.Notify(640980)
	}
	__antithesis_instrumentation__.Notify(640978)
	s.mvccIter.Next()
}

func (s *separatedIntentAndVersionIter) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(640981)
	if s.atMVCCIter {
		__antithesis_instrumentation__.Notify(640983)
		return s.mvccIter.UnsafeKey()
	} else {
		__antithesis_instrumentation__.Notify(640984)
	}
	__antithesis_instrumentation__.Notify(640982)
	return MVCCKey{Key: s.intentKey}
}

func (s *separatedIntentAndVersionIter) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(640985)
	if s.atMVCCIter {
		__antithesis_instrumentation__.Notify(640987)
		return s.mvccIter.UnsafeValue()
	} else {
		__antithesis_instrumentation__.Notify(640988)
	}
	__antithesis_instrumentation__.Notify(640986)
	return s.engineIter.UnsafeValue()
}

func (s *separatedIntentAndVersionIter) ValueProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(640989)
	if s.atMVCCIter {
		__antithesis_instrumentation__.Notify(640992)
		return s.mvccIter.ValueProto(msg)
	} else {
		__antithesis_instrumentation__.Notify(640993)
	}
	__antithesis_instrumentation__.Notify(640990)
	meta, ok := msg.(*enginepb.MVCCMetadata)
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(640994)
		return meta == s.meta == true
	}() == true {
		__antithesis_instrumentation__.Notify(640995)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(640996)
	}
	__antithesis_instrumentation__.Notify(640991)
	v := s.engineIter.UnsafeValue()
	return protoutil.Unmarshal(v, msg)
}

func mvccGetIntent(
	iter iterForKeyVersions, metaKey MVCCKey, meta *enginepb.MVCCMetadata,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(640997)
	if ok, err := iter.Valid(); !ok {
		__antithesis_instrumentation__.Notify(641002)
		return false, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(641003)
	}
	__antithesis_instrumentation__.Notify(640998)
	unsafeKey := iter.UnsafeKey()
	if !unsafeKey.Key.Equal(metaKey.Key) {
		__antithesis_instrumentation__.Notify(641004)
		return false, 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(641005)
	}
	__antithesis_instrumentation__.Notify(640999)
	if unsafeKey.IsValue() {
		__antithesis_instrumentation__.Notify(641006)
		return false, 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(641007)
	}
	__antithesis_instrumentation__.Notify(641000)
	if err := iter.ValueProto(meta); err != nil {
		__antithesis_instrumentation__.Notify(641008)
		return false, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(641009)
	}
	__antithesis_instrumentation__.Notify(641001)
	return true, int64(unsafeKey.EncodedSize()),
		int64(len(iter.UnsafeValue())), nil
}

type singleDelOptimizationHelper struct {
	_didNotUpdateMeta *bool
	_hasIgnoredSeqs   bool
	_epoch            enginepb.TxnEpoch
}

func (h singleDelOptimizationHelper) v() bool {
	__antithesis_instrumentation__.Notify(641010)
	if h._didNotUpdateMeta == nil {
		__antithesis_instrumentation__.Notify(641012)
		return false
	} else {
		__antithesis_instrumentation__.Notify(641013)
	}
	__antithesis_instrumentation__.Notify(641011)
	return *h._didNotUpdateMeta
}

func (h singleDelOptimizationHelper) onCommitIntent() bool {
	__antithesis_instrumentation__.Notify(641014)

	return h.v() && func() bool {
		__antithesis_instrumentation__.Notify(641015)
		return !h._hasIgnoredSeqs == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(641016)
		return h._epoch == 0 == true
	}() == true
}

func (h singleDelOptimizationHelper) onAbortIntent() bool {
	__antithesis_instrumentation__.Notify(641017)
	return false
}

func mvccResolveWriteIntent(
	ctx context.Context,
	rw ReadWriter,
	iter iterForKeyVersions,
	ms *enginepb.MVCCStats,
	intent roachpb.LockUpdate,
	buf *putBuffer,
) (bool, error) {
	__antithesis_instrumentation__.Notify(641018)
	metaKey := MakeMVCCMetadataKey(intent.Key)
	meta := &buf.meta
	ok, origMetaKeySize, origMetaValSize, err :=
		mvccGetIntent(iter, metaKey, meta)
	if err != nil {
		__antithesis_instrumentation__.Notify(641029)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(641030)
	}
	__antithesis_instrumentation__.Notify(641019)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(641031)
		return meta.Txn == nil == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(641032)
		return intent.Txn.ID != meta.Txn.ID == true
	}() == true {
		__antithesis_instrumentation__.Notify(641033)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(641034)
	}
	__antithesis_instrumentation__.Notify(641020)
	metaTimestamp := meta.Timestamp.ToTimestamp()
	canSingleDelHelper := singleDelOptimizationHelper{
		_didNotUpdateMeta: meta.TxnDidNotUpdateMeta,
		_hasIgnoredSeqs:   len(intent.IgnoredSeqNums) > 0,

		_epoch: intent.Txn.Epoch,
	}

	epochsMatch := meta.Txn.Epoch == intent.Txn.Epoch
	timestampsValid := metaTimestamp.LessEq(intent.Txn.WriteTimestamp)
	timestampChanged := metaTimestamp.Less(intent.Txn.WriteTimestamp)
	commit := intent.Status == roachpb.COMMITTED && func() bool {
		__antithesis_instrumentation__.Notify(641035)
		return epochsMatch == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(641036)
		return timestampsValid == true
	}() == true

	inProgress := !intent.Status.IsFinalized() && func() bool {
		__antithesis_instrumentation__.Notify(641037)
		return meta.Txn.Epoch >= intent.Txn.Epoch == true
	}() == true
	pushed := inProgress && func() bool {
		__antithesis_instrumentation__.Notify(641038)
		return timestampChanged == true
	}() == true
	latestKey := MVCCKey{Key: intent.Key, Timestamp: metaTimestamp}

	var rolledBackVal []byte
	if len(intent.IgnoredSeqNums) > 0 {
		__antithesis_instrumentation__.Notify(641039)

		var removeIntent bool
		removeIntent, rolledBackVal, err = mvccMaybeRewriteIntentHistory(ctx, rw, intent.IgnoredSeqNums, meta, latestKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(641042)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(641043)
		}
		__antithesis_instrumentation__.Notify(641040)

		if removeIntent {
			__antithesis_instrumentation__.Notify(641044)

			commit = false
			pushed = false
			inProgress = false
			rolledBackVal = nil
		} else {
			__antithesis_instrumentation__.Notify(641045)
		}
		__antithesis_instrumentation__.Notify(641041)

		if rolledBackVal != nil {
			__antithesis_instrumentation__.Notify(641046)

			intent.Txn.WriteTimestamp.Forward(metaTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(641047)
		}
	} else {
		__antithesis_instrumentation__.Notify(641048)
	}
	__antithesis_instrumentation__.Notify(641021)

	if inProgress && func() bool {
		__antithesis_instrumentation__.Notify(641049)
		return !pushed == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(641050)
		return rolledBackVal == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(641051)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(641052)
	}
	__antithesis_instrumentation__.Notify(641022)

	if commit || func() bool {
		__antithesis_instrumentation__.Notify(641053)
		return pushed == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(641054)
		return rolledBackVal != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(641055)

		newTimestamp := intent.Txn.WriteTimestamp

		if newTimestamp.Less(metaTimestamp) {
			__antithesis_instrumentation__.Notify(641062)
			return false, errors.AssertionFailedf("timestamp regression (%s -> %s) "+
				"during intent resolution, commit=%t pushed=%t rolledBackVal=%t",
				metaTimestamp, newTimestamp, commit, pushed, rolledBackVal != nil)
		} else {
			__antithesis_instrumentation__.Notify(641063)
		}
		__antithesis_instrumentation__.Notify(641056)

		buf.newMeta = *meta

		buf.newMeta.Timestamp = newTimestamp.ToLegacyTimestamp()
		buf.newMeta.Txn.WriteTimestamp = newTimestamp

		var metaKeySize, metaValSize int64
		if !commit {
			__antithesis_instrumentation__.Notify(641064)

			metaKeySize, metaValSize, err = buf.putIntentMeta(
				ctx, rw, metaKey, &buf.newMeta, true)
		} else {
			__antithesis_instrumentation__.Notify(641065)
			metaKeySize = int64(metaKey.EncodedSize())
			err = rw.ClearIntent(metaKey.Key, canSingleDelHelper.onCommitIntent(), meta.Txn.ID)
		}
		__antithesis_instrumentation__.Notify(641057)
		if err != nil {
			__antithesis_instrumentation__.Notify(641066)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(641067)
		}
		__antithesis_instrumentation__.Notify(641058)

		var prevValSize int64
		if timestampChanged {
			__antithesis_instrumentation__.Notify(641068)
			oldKey := MVCCKey{Key: intent.Key, Timestamp: metaTimestamp}
			newKey := MVCCKey{Key: intent.Key, Timestamp: newTimestamp}

			iter.SeekGE(oldKey)
			if valid, err := iter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(641073)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(641074)
				if !valid || func() bool {
					__antithesis_instrumentation__.Notify(641075)
					return !iter.UnsafeKey().Equal(oldKey) == true
				}() == true {
					__antithesis_instrumentation__.Notify(641076)
					return false, errors.Errorf("existing intent value missing: %s", oldKey)
				} else {
					__antithesis_instrumentation__.Notify(641077)
				}
			}
			__antithesis_instrumentation__.Notify(641069)
			value := iter.UnsafeValue()

			if rolledBackVal != nil {
				__antithesis_instrumentation__.Notify(641078)
				value = rolledBackVal
			} else {
				__antithesis_instrumentation__.Notify(641079)
			}
			__antithesis_instrumentation__.Notify(641070)
			if err = rw.PutMVCC(newKey, value); err != nil {
				__antithesis_instrumentation__.Notify(641080)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(641081)
			}
			__antithesis_instrumentation__.Notify(641071)
			if err = rw.ClearMVCC(oldKey); err != nil {
				__antithesis_instrumentation__.Notify(641082)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(641083)
			}
			__antithesis_instrumentation__.Notify(641072)

			iter.Next()
			if valid, err := iter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(641084)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(641085)
				if valid && func() bool {
					__antithesis_instrumentation__.Notify(641086)
					return iter.UnsafeKey().Key.Equal(oldKey.Key) == true
				}() == true {
					__antithesis_instrumentation__.Notify(641087)
					prevValSize = int64(len(iter.UnsafeValue()))
				} else {
					__antithesis_instrumentation__.Notify(641088)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(641089)
		}
		__antithesis_instrumentation__.Notify(641059)

		if ms != nil {
			__antithesis_instrumentation__.Notify(641090)
			ms.Add(updateStatsOnResolve(intent.Key, prevValSize, origMetaKeySize, origMetaValSize,
				metaKeySize, metaValSize, meta, &buf.newMeta, commit))
		} else {
			__antithesis_instrumentation__.Notify(641091)
		}
		__antithesis_instrumentation__.Notify(641060)

		logicalOp := MVCCCommitIntentOpType
		if pushed {
			__antithesis_instrumentation__.Notify(641092)
			logicalOp = MVCCUpdateIntentOpType
		} else {
			__antithesis_instrumentation__.Notify(641093)
		}
		__antithesis_instrumentation__.Notify(641061)
		rw.LogLogicalOp(logicalOp, MVCCLogicalOpDetails{
			Txn:       intent.Txn,
			Key:       intent.Key,
			Timestamp: intent.Txn.WriteTimestamp,
		})

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(641094)
	}
	__antithesis_instrumentation__.Notify(641023)

	if err := rw.ClearMVCC(latestKey); err != nil {
		__antithesis_instrumentation__.Notify(641095)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(641096)
	}
	__antithesis_instrumentation__.Notify(641024)

	rw.LogLogicalOp(MVCCAbortIntentOpType, MVCCLogicalOpDetails{
		Txn: intent.Txn,
		Key: intent.Key,
	})

	nextKey := latestKey.Next()
	ok = false
	var unsafeNextKey MVCCKey
	var unsafeNextValue []byte
	if nextKey.IsValue() {
		__antithesis_instrumentation__.Notify(641097)

		iter.SeekGE(nextKey)
		ok, err = iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(641100)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(641101)
		}
		__antithesis_instrumentation__.Notify(641098)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(641102)
			return iter.UnsafeKey().Key.Equal(latestKey.Key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(641103)
			unsafeNextKey = iter.UnsafeKey()
			if !unsafeNextKey.IsValue() {
				__antithesis_instrumentation__.Notify(641105)

				return false, errors.Errorf("expected an MVCC value key: %s", unsafeNextKey)
			} else {
				__antithesis_instrumentation__.Notify(641106)
			}
			__antithesis_instrumentation__.Notify(641104)
			unsafeNextValue = iter.UnsafeValue()
		} else {
			__antithesis_instrumentation__.Notify(641107)
			ok = false
		}
		__antithesis_instrumentation__.Notify(641099)
		iter = nil
	} else {
		__antithesis_instrumentation__.Notify(641108)
	}
	__antithesis_instrumentation__.Notify(641025)

	if !ok {
		__antithesis_instrumentation__.Notify(641109)

		if err = rw.ClearIntent(metaKey.Key, canSingleDelHelper.onAbortIntent(), meta.Txn.ID); err != nil {
			__antithesis_instrumentation__.Notify(641112)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(641113)
		}
		__antithesis_instrumentation__.Notify(641110)

		if ms != nil {
			__antithesis_instrumentation__.Notify(641114)
			ms.Add(updateStatsOnClear(
				intent.Key, origMetaKeySize, origMetaValSize, 0, 0, meta, nil, 0))
		} else {
			__antithesis_instrumentation__.Notify(641115)
		}
		__antithesis_instrumentation__.Notify(641111)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(641116)
	}
	__antithesis_instrumentation__.Notify(641026)

	valueSize := int64(len(unsafeNextValue))

	buf.newMeta = enginepb.MVCCMetadata{
		Deleted:  valueSize == 0,
		KeyBytes: MVCCVersionTimestampSize,
		ValBytes: valueSize,
	}
	if err = rw.ClearIntent(metaKey.Key, canSingleDelHelper.onAbortIntent(), meta.Txn.ID); err != nil {
		__antithesis_instrumentation__.Notify(641117)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(641118)
	}
	__antithesis_instrumentation__.Notify(641027)
	metaKeySize := int64(metaKey.EncodedSize())
	metaValSize := int64(0)

	if ms != nil {
		__antithesis_instrumentation__.Notify(641119)
		ms.Add(updateStatsOnClear(intent.Key, origMetaKeySize, origMetaValSize, metaKeySize,
			metaValSize, meta, &buf.newMeta, unsafeNextKey.Timestamp.WallTime))
	} else {
		__antithesis_instrumentation__.Notify(641120)
	}
	__antithesis_instrumentation__.Notify(641028)

	return true, nil
}

func mvccMaybeRewriteIntentHistory(
	ctx context.Context,
	engine ReadWriter,
	ignoredSeqNums []enginepb.IgnoredSeqNumRange,
	meta *enginepb.MVCCMetadata,
	latestKey MVCCKey,
) (remove bool, updatedVal []byte, err error) {
	__antithesis_instrumentation__.Notify(641121)
	if !enginepb.TxnSeqIsIgnored(meta.Txn.Sequence, ignoredSeqNums) {
		__antithesis_instrumentation__.Notify(641125)

		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(641126)
	}
	__antithesis_instrumentation__.Notify(641122)

	var i int
	for i = len(meta.IntentHistory) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(641127)
		e := &meta.IntentHistory[i]
		if !enginepb.TxnSeqIsIgnored(e.Sequence, ignoredSeqNums) {
			__antithesis_instrumentation__.Notify(641128)
			break
		} else {
			__antithesis_instrumentation__.Notify(641129)
		}
	}
	__antithesis_instrumentation__.Notify(641123)

	if i < 0 {
		__antithesis_instrumentation__.Notify(641130)
		return true, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(641131)
	}
	__antithesis_instrumentation__.Notify(641124)

	restoredVal := meta.IntentHistory[i].Value
	meta.Txn.Sequence = meta.IntentHistory[i].Sequence
	meta.IntentHistory = meta.IntentHistory[:i]
	meta.Deleted = len(restoredVal) == 0
	meta.ValBytes = int64(len(restoredVal))

	err = engine.PutMVCC(latestKey, restoredVal)

	return false, restoredVal, err
}

type IterAndBuf struct {
	buf  *putBuffer
	iter MVCCIterator
}

func GetIterAndBuf(reader Reader, opts IterOptions) IterAndBuf {
	__antithesis_instrumentation__.Notify(641132)
	return GetBufUsingIter(reader.NewMVCCIterator(MVCCKeyAndIntentsIterKind, opts))
}

func GetBufUsingIter(iter MVCCIterator) IterAndBuf {
	__antithesis_instrumentation__.Notify(641133)
	return IterAndBuf{
		buf:  newPutBuffer(),
		iter: iter,
	}
}

func (b IterAndBuf) Cleanup() {
	__antithesis_instrumentation__.Notify(641134)
	b.buf.release()
	if b.iter != nil {
		__antithesis_instrumentation__.Notify(641135)
		b.iter.Close()
	} else {
		__antithesis_instrumentation__.Notify(641136)
	}
}

func MVCCResolveWriteIntentRange(
	ctx context.Context, rw ReadWriter, ms *enginepb.MVCCStats, intent roachpb.LockUpdate, max int64,
) (int64, *roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(641137)
	if max < 0 {
		__antithesis_instrumentation__.Notify(641141)
		resumeSpan := intent.Span
		return 0, &resumeSpan, nil
	} else {
		__antithesis_instrumentation__.Notify(641142)
	}
	__antithesis_instrumentation__.Notify(641138)

	var putBuf *putBuffer

	var sepIter *separatedIntentAndVersionIter
	var mvccIter MVCCIterator

	var iter iterForKeyVersions

	if rw.ConsistentIterators() {
		__antithesis_instrumentation__.Notify(641143)
		ltStart, _ := keys.LockTableSingleKey(intent.Key, nil)
		ltEnd, _ := keys.LockTableSingleKey(intent.EndKey, nil)
		engineIter := rw.NewEngineIterator(IterOptions{LowerBound: ltStart, UpperBound: ltEnd})
		iterAndBuf :=
			GetBufUsingIter(rw.NewMVCCIterator(MVCCKeyIterKind, IterOptions{UpperBound: intent.EndKey}))
		defer func() {
			__antithesis_instrumentation__.Notify(641145)
			engineIter.Close()
			iterAndBuf.Cleanup()
		}()
		__antithesis_instrumentation__.Notify(641144)
		putBuf = iterAndBuf.buf
		sepIter = &separatedIntentAndVersionIter{
			engineIter: engineIter,
			mvccIter:   iterAndBuf.iter,
		}
		iter = sepIter

		sepIter.seekEngineKeyGE(EngineKey{Key: ltStart})
	} else {
		__antithesis_instrumentation__.Notify(641146)
		iterAndBuf := GetIterAndBuf(rw, IterOptions{UpperBound: intent.EndKey})
		defer iterAndBuf.Cleanup()
		putBuf = iterAndBuf.buf
		mvccIter = iterAndBuf.iter
		iter = mvccIter
	}
	__antithesis_instrumentation__.Notify(641139)
	nextKey := MakeMVCCMetadataKey(intent.Key)
	intentEndKey := intent.EndKey
	intent.EndKey = nil

	var keyBuf []byte
	num := int64(0)
	for {
		__antithesis_instrumentation__.Notify(641147)
		if max > 0 && func() bool {
			__antithesis_instrumentation__.Notify(641153)
			return num == max == true
		}() == true {
			__antithesis_instrumentation__.Notify(641154)
			return num, &roachpb.Span{Key: nextKey.Key, EndKey: intentEndKey}, nil
		} else {
			__antithesis_instrumentation__.Notify(641155)
		}
		__antithesis_instrumentation__.Notify(641148)
		var key MVCCKey
		if sepIter != nil {
			__antithesis_instrumentation__.Notify(641156)

			if valid, err := sepIter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(641160)
				return 0, nil, err
			} else {
				__antithesis_instrumentation__.Notify(641161)
				if !valid {
					__antithesis_instrumentation__.Notify(641162)

					break
				} else {
					__antithesis_instrumentation__.Notify(641163)
				}
			}
			__antithesis_instrumentation__.Notify(641157)

			meta := &putBuf.meta
			if err := sepIter.ValueProto(meta); err != nil {
				__antithesis_instrumentation__.Notify(641164)
				return 0, nil, err
			} else {
				__antithesis_instrumentation__.Notify(641165)
			}
			__antithesis_instrumentation__.Notify(641158)
			if meta.Txn == nil {
				__antithesis_instrumentation__.Notify(641166)
				return 0, nil, errors.Errorf("intent with no txn")
			} else {
				__antithesis_instrumentation__.Notify(641167)
			}
			__antithesis_instrumentation__.Notify(641159)
			if intent.Txn.ID == meta.Txn.ID {
				__antithesis_instrumentation__.Notify(641168)

				sepIter.meta = meta

				key = sepIter.UnsafeKey()
				keyBuf = append(keyBuf[:0], key.Key...)
				key.Key = keyBuf
			} else {
				__antithesis_instrumentation__.Notify(641169)
				sepIter.nextEngineKey()
				continue
			}
		} else {
			__antithesis_instrumentation__.Notify(641170)

			mvccIter.SeekGE(nextKey)
			if valid, err := mvccIter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(641172)
				return 0, nil, err
			} else {
				__antithesis_instrumentation__.Notify(641173)
				if !valid {
					__antithesis_instrumentation__.Notify(641174)

					break
				} else {
					__antithesis_instrumentation__.Notify(641175)
				}
			}
			__antithesis_instrumentation__.Notify(641171)
			key = mvccIter.UnsafeKey()

			keyBuf = append(keyBuf[:0], key.Key...)
			key.Key = keyBuf
		}
		__antithesis_instrumentation__.Notify(641149)

		var err error
		var ok bool
		if !key.IsValue() {
			__antithesis_instrumentation__.Notify(641176)

			intent.Key = key.Key
			ok, err = mvccResolveWriteIntent(ctx, rw, iter, ms, intent, putBuf)
		} else {
			__antithesis_instrumentation__.Notify(641177)
		}
		__antithesis_instrumentation__.Notify(641150)
		if err != nil {
			__antithesis_instrumentation__.Notify(641178)
			log.Warningf(ctx, "failed to resolve intent for key %q: %+v", key.Key, err)
		} else {
			__antithesis_instrumentation__.Notify(641179)
			if ok {
				__antithesis_instrumentation__.Notify(641180)
				num++
			} else {
				__antithesis_instrumentation__.Notify(641181)
			}
		}
		__antithesis_instrumentation__.Notify(641151)

		if sepIter != nil {
			__antithesis_instrumentation__.Notify(641182)
			sepIter.nextEngineKey()

		} else {
			__antithesis_instrumentation__.Notify(641183)
		}
		__antithesis_instrumentation__.Notify(641152)

		nextKey.Key = key.Key.Next()
		if nextKey.Key.Compare(intentEndKey) >= 0 {
			__antithesis_instrumentation__.Notify(641184)

			break
		} else {
			__antithesis_instrumentation__.Notify(641185)
		}
	}
	__antithesis_instrumentation__.Notify(641140)
	return num, nil, nil
}

func MVCCGarbageCollect(
	ctx context.Context,
	rw ReadWriter,
	ms *enginepb.MVCCStats,
	keys []roachpb.GCRequest_GCKey,
	timestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(641186)

	var count int64
	defer func(begin time.Time) {
		__antithesis_instrumentation__.Notify(641191)
		log.Eventf(ctx, "done with GC evaluation for %d keys at %.2f keys/sec. Deleted %d entries",
			len(keys), float64(len(keys))*1e9/float64(timeutil.Since(begin)), count)
	}(timeutil.Now())
	__antithesis_instrumentation__.Notify(641187)

	if len(keys) == 0 {
		__antithesis_instrumentation__.Notify(641192)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641193)
	}
	__antithesis_instrumentation__.Notify(641188)

	sort.Slice(keys, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(641194)
		iKey := MVCCKey{Key: keys[i].Key, Timestamp: keys[i].Timestamp}
		jKey := MVCCKey{Key: keys[j].Key, Timestamp: keys[j].Timestamp}
		return iKey.Less(jKey)
	})
	__antithesis_instrumentation__.Notify(641189)

	iter := rw.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{
		LowerBound: keys[0].Key,
		UpperBound: keys[len(keys)-1].Key.Next(),
	})
	defer iter.Close()
	supportsPrev := iter.SupportsPrev()

	meta := &enginepb.MVCCMetadata{}
	for _, gcKey := range keys {
		__antithesis_instrumentation__.Notify(641195)
		encKey := MakeMVCCMetadataKey(gcKey.Key)
		ok, metaKeySize, metaValSize, err :=
			mvccGetMetadata(iter, encKey, false, meta)
		if err != nil {
			__antithesis_instrumentation__.Notify(641201)
			return err
		} else {
			__antithesis_instrumentation__.Notify(641202)
		}
		__antithesis_instrumentation__.Notify(641196)
		if !ok {
			__antithesis_instrumentation__.Notify(641203)
			continue
		} else {
			__antithesis_instrumentation__.Notify(641204)
		}
		__antithesis_instrumentation__.Notify(641197)
		inlinedValue := meta.IsInline()
		implicitMeta := iter.UnsafeKey().IsValue()

		if meta.Timestamp.ToTimestamp().LessEq(gcKey.Timestamp) {
			__antithesis_instrumentation__.Notify(641205)

			if !meta.Deleted && func() bool {
				__antithesis_instrumentation__.Notify(641209)
				return !inlinedValue == true
			}() == true {
				__antithesis_instrumentation__.Notify(641210)
				return errors.Errorf("request to GC non-deleted, latest value of %q", gcKey.Key)
			} else {
				__antithesis_instrumentation__.Notify(641211)
			}
			__antithesis_instrumentation__.Notify(641206)
			if meta.Txn != nil {
				__antithesis_instrumentation__.Notify(641212)
				return errors.Errorf("request to GC intent at %q", gcKey.Key)
			} else {
				__antithesis_instrumentation__.Notify(641213)
			}
			__antithesis_instrumentation__.Notify(641207)
			if ms != nil {
				__antithesis_instrumentation__.Notify(641214)
				if inlinedValue {
					__antithesis_instrumentation__.Notify(641215)
					updateStatsForInline(ms, gcKey.Key, metaKeySize, metaValSize, 0, 0)
					ms.AgeTo(timestamp.WallTime)
				} else {
					__antithesis_instrumentation__.Notify(641216)
					ms.Add(updateStatsOnGC(gcKey.Key, metaKeySize, metaValSize, meta, meta.Timestamp.WallTime))
				}
			} else {
				__antithesis_instrumentation__.Notify(641217)
			}
			__antithesis_instrumentation__.Notify(641208)
			if !implicitMeta {
				__antithesis_instrumentation__.Notify(641218)

				if err := rw.ClearUnversioned(iter.UnsafeKey().Key); err != nil {
					__antithesis_instrumentation__.Notify(641220)
					return err
				} else {
					__antithesis_instrumentation__.Notify(641221)
				}
				__antithesis_instrumentation__.Notify(641219)
				count++
			} else {
				__antithesis_instrumentation__.Notify(641222)
			}
		} else {
			__antithesis_instrumentation__.Notify(641223)
		}
		__antithesis_instrumentation__.Notify(641198)

		if !implicitMeta {
			__antithesis_instrumentation__.Notify(641224)

			iter.Next()
		} else {
			__antithesis_instrumentation__.Notify(641225)
		}
		__antithesis_instrumentation__.Notify(641199)

		prevNanos := timestamp.WallTime
		{
			__antithesis_instrumentation__.Notify(641226)

			var foundPrevNanos bool
			{
				__antithesis_instrumentation__.Notify(641228)

				var foundNextKey bool

				const nextsBeforeSeekLT = 4
				for i := 0; !supportsPrev || func() bool {
					__antithesis_instrumentation__.Notify(641230)
					return i < nextsBeforeSeekLT == true
				}() == true; i++ {
					__antithesis_instrumentation__.Notify(641231)
					if i > 0 {
						__antithesis_instrumentation__.Notify(641236)
						iter.Next()
					} else {
						__antithesis_instrumentation__.Notify(641237)
					}
					__antithesis_instrumentation__.Notify(641232)
					if ok, err := iter.Valid(); err != nil {
						__antithesis_instrumentation__.Notify(641238)
						return err
					} else {
						__antithesis_instrumentation__.Notify(641239)
						if !ok {
							__antithesis_instrumentation__.Notify(641240)
							foundNextKey = true
							break
						} else {
							__antithesis_instrumentation__.Notify(641241)
						}
					}
					__antithesis_instrumentation__.Notify(641233)
					unsafeIterKey := iter.UnsafeKey()
					if !unsafeIterKey.Key.Equal(encKey.Key) {
						__antithesis_instrumentation__.Notify(641242)
						foundNextKey = true
						break
					} else {
						__antithesis_instrumentation__.Notify(641243)
					}
					__antithesis_instrumentation__.Notify(641234)
					if unsafeIterKey.Timestamp.LessEq(gcKey.Timestamp) {
						__antithesis_instrumentation__.Notify(641244)
						foundPrevNanos = true
						break
					} else {
						__antithesis_instrumentation__.Notify(641245)
					}
					__antithesis_instrumentation__.Notify(641235)
					prevNanos = unsafeIterKey.Timestamp.WallTime
				}
				__antithesis_instrumentation__.Notify(641229)

				if foundNextKey {
					__antithesis_instrumentation__.Notify(641246)
					continue
				} else {
					__antithesis_instrumentation__.Notify(641247)
				}
			}
			__antithesis_instrumentation__.Notify(641227)

			if !foundPrevNanos {
				__antithesis_instrumentation__.Notify(641248)
				if !supportsPrev {
					__antithesis_instrumentation__.Notify(641250)
					log.Fatalf(ctx, "failed to find first garbage key without"+
						"support for reverse iteration")
				} else {
					__antithesis_instrumentation__.Notify(641251)
				}
				__antithesis_instrumentation__.Notify(641249)
				gcKeyMVCC := MVCCKey{Key: gcKey.Key, Timestamp: gcKey.Timestamp}
				iter.SeekLT(gcKeyMVCC)
				if ok, err := iter.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(641252)
					return err
				} else {
					__antithesis_instrumentation__.Notify(641253)
					if ok {
						__antithesis_instrumentation__.Notify(641254)

						if iter.UnsafeKey().Key.Equal(gcKey.Key) {
							__antithesis_instrumentation__.Notify(641256)
							prevNanos = iter.UnsafeKey().Timestamp.WallTime
						} else {
							__antithesis_instrumentation__.Notify(641257)
						}
						__antithesis_instrumentation__.Notify(641255)

						iter.Next()
					} else {
						__antithesis_instrumentation__.Notify(641258)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(641259)
			}
		}
		__antithesis_instrumentation__.Notify(641200)

		for ; ; iter.Next() {
			__antithesis_instrumentation__.Notify(641260)
			if ok, err := iter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(641266)
				return err
			} else {
				__antithesis_instrumentation__.Notify(641267)
				if !ok {
					__antithesis_instrumentation__.Notify(641268)
					break
				} else {
					__antithesis_instrumentation__.Notify(641269)
				}
			}
			__antithesis_instrumentation__.Notify(641261)
			unsafeIterKey := iter.UnsafeKey()
			if !unsafeIterKey.Key.Equal(encKey.Key) {
				__antithesis_instrumentation__.Notify(641270)
				break
			} else {
				__antithesis_instrumentation__.Notify(641271)
			}
			__antithesis_instrumentation__.Notify(641262)
			if !unsafeIterKey.IsValue() {
				__antithesis_instrumentation__.Notify(641272)
				break
			} else {
				__antithesis_instrumentation__.Notify(641273)
			}
			__antithesis_instrumentation__.Notify(641263)
			if ms != nil {
				__antithesis_instrumentation__.Notify(641274)

				valSize := int64(len(iter.UnsafeValue()))

				fromNS := prevNanos
				if valSize == 0 {
					__antithesis_instrumentation__.Notify(641276)
					fromNS = unsafeIterKey.Timestamp.WallTime
				} else {
					__antithesis_instrumentation__.Notify(641277)
				}
				__antithesis_instrumentation__.Notify(641275)

				ms.Add(updateStatsOnGC(gcKey.Key, MVCCVersionTimestampSize,
					valSize, nil, fromNS))
			} else {
				__antithesis_instrumentation__.Notify(641278)
			}
			__antithesis_instrumentation__.Notify(641264)
			count++
			if err := rw.ClearMVCC(unsafeIterKey); err != nil {
				__antithesis_instrumentation__.Notify(641279)
				return err
			} else {
				__antithesis_instrumentation__.Notify(641280)
			}
			__antithesis_instrumentation__.Notify(641265)
			prevNanos = unsafeIterKey.Timestamp.WallTime
		}
	}
	__antithesis_instrumentation__.Notify(641190)

	return nil
}

func MVCCFindSplitKey(
	_ context.Context, reader Reader, key, endKey roachpb.RKey, targetSize int64,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(641281)
	if key.Less(roachpb.RKey(keys.LocalMax)) {
		__antithesis_instrumentation__.Notify(641287)
		key = roachpb.RKey(keys.LocalMax)
	} else {
		__antithesis_instrumentation__.Notify(641288)
	}
	__antithesis_instrumentation__.Notify(641282)

	it := reader.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{UpperBound: endKey.AsRawKey()})
	defer it.Close()

	it.SeekGE(MakeMVCCMetadataKey(key.AsRawKey()))
	if ok, err := it.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(641289)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641290)
		if !ok {
			__antithesis_instrumentation__.Notify(641291)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(641292)
		}
	}
	__antithesis_instrumentation__.Notify(641283)
	var minSplitKey roachpb.Key
	if _, tenID, err := keys.DecodeTenantPrefix(it.UnsafeKey().Key); err == nil {
		__antithesis_instrumentation__.Notify(641293)
		if _, _, err := keys.MakeSQLCodec(tenID).DecodeTablePrefix(it.UnsafeKey().Key); err == nil {
			__antithesis_instrumentation__.Notify(641294)

			firstRowKey, err := keys.EnsureSafeSplitKey(it.Key().Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(641296)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(641297)
			}
			__antithesis_instrumentation__.Notify(641295)

			minSplitKey = firstRowKey.PrefixEnd()
		} else {
			__antithesis_instrumentation__.Notify(641298)
		}
	} else {
		__antithesis_instrumentation__.Notify(641299)
	}
	__antithesis_instrumentation__.Notify(641284)
	if minSplitKey == nil {
		__antithesis_instrumentation__.Notify(641300)

		minSplitKey = it.Key().Key.Next()
	} else {
		__antithesis_instrumentation__.Notify(641301)
	}
	__antithesis_instrumentation__.Notify(641285)

	splitKey, err := it.FindSplitKey(key.AsRawKey(), endKey.AsRawKey(), minSplitKey, targetSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(641302)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641303)
	}
	__antithesis_instrumentation__.Notify(641286)

	return keys.EnsureSafeSplitKey(splitKey.Key)
}

func willOverflow(a, b int64) bool {
	__antithesis_instrumentation__.Notify(641304)

	if a > b {
		__antithesis_instrumentation__.Notify(641307)
		a, b = b, a
	} else {
		__antithesis_instrumentation__.Notify(641308)
	}
	__antithesis_instrumentation__.Notify(641305)

	if b > 0 {
		__antithesis_instrumentation__.Notify(641309)
		return a > math.MaxInt64-b
	} else {
		__antithesis_instrumentation__.Notify(641310)
	}
	__antithesis_instrumentation__.Notify(641306)
	return math.MinInt64-b > a
}

func ComputeStatsForRange(
	iter SimpleMVCCIterator,
	start, end roachpb.Key,
	nowNanos int64,
	callbacks ...func(MVCCKey, []byte) error,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(641311)
	var ms enginepb.MVCCStats

	var meta enginepb.MVCCMetadata
	var prevKey []byte
	first := false

	var accrueGCAgeNanos int64
	mvccEndKey := MakeMVCCMetadataKey(end)

	iter.SeekGE(MakeMVCCMetadataKey(start))
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(641313)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(641320)
			return ms, err
		} else {
			__antithesis_instrumentation__.Notify(641321)
		}
		__antithesis_instrumentation__.Notify(641314)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(641322)
			return !iter.UnsafeKey().Less(mvccEndKey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(641323)
			break
		} else {
			__antithesis_instrumentation__.Notify(641324)
		}
		__antithesis_instrumentation__.Notify(641315)

		unsafeKey := iter.UnsafeKey()
		unsafeValue := iter.UnsafeValue()

		for _, f := range callbacks {
			__antithesis_instrumentation__.Notify(641325)
			if err := f(unsafeKey, unsafeValue); err != nil {
				__antithesis_instrumentation__.Notify(641326)
				return enginepb.MVCCStats{}, err
			} else {
				__antithesis_instrumentation__.Notify(641327)
			}
		}
		__antithesis_instrumentation__.Notify(641316)

		if bytes.HasPrefix(unsafeKey.Key, keys.LocalRangeIDPrefix) {
			__antithesis_instrumentation__.Notify(641328)

			_, infix, suffix, _, err := keys.DecodeRangeIDKey(unsafeKey.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(641330)
				return enginepb.MVCCStats{}, errors.Wrap(err, "unable to decode rangeID key")
			} else {
				__antithesis_instrumentation__.Notify(641331)
			}
			__antithesis_instrumentation__.Notify(641329)

			if infix.Equal(keys.LocalRangeIDReplicatedInfix) {
				__antithesis_instrumentation__.Notify(641332)

				if suffix.Equal(keys.LocalRangeAppliedStateSuffix) {
					__antithesis_instrumentation__.Notify(641333)

					continue
				} else {
					__antithesis_instrumentation__.Notify(641334)
				}
			} else {
				__antithesis_instrumentation__.Notify(641335)
			}
		} else {
			__antithesis_instrumentation__.Notify(641336)
		}
		__antithesis_instrumentation__.Notify(641317)

		isSys := isSysLocal(unsafeKey.Key)
		isValue := unsafeKey.IsValue()
		implicitMeta := isValue && func() bool {
			__antithesis_instrumentation__.Notify(641337)
			return !bytes.Equal(unsafeKey.Key, prevKey) == true
		}() == true
		prevKey = append(prevKey[:0], unsafeKey.Key...)

		if implicitMeta {
			__antithesis_instrumentation__.Notify(641338)

			meta.Reset()
			meta.KeyBytes = MVCCVersionTimestampSize
			meta.ValBytes = int64(len(unsafeValue))
			meta.Deleted = len(unsafeValue) == 0
			meta.Timestamp.WallTime = unsafeKey.Timestamp.WallTime
		} else {
			__antithesis_instrumentation__.Notify(641339)
		}
		__antithesis_instrumentation__.Notify(641318)

		if !isValue || func() bool {
			__antithesis_instrumentation__.Notify(641340)
			return implicitMeta == true
		}() == true {
			__antithesis_instrumentation__.Notify(641341)
			metaKeySize := int64(len(unsafeKey.Key)) + 1
			var metaValSize int64
			if !implicitMeta {
				__antithesis_instrumentation__.Notify(641345)
				metaValSize = int64(len(unsafeValue))
			} else {
				__antithesis_instrumentation__.Notify(641346)
			}
			__antithesis_instrumentation__.Notify(641342)
			totalBytes := metaKeySize + metaValSize
			first = true

			if !implicitMeta {
				__antithesis_instrumentation__.Notify(641347)
				if err := protoutil.Unmarshal(unsafeValue, &meta); err != nil {
					__antithesis_instrumentation__.Notify(641348)
					return ms, errors.Wrap(err, "unable to decode MVCCMetadata")
				} else {
					__antithesis_instrumentation__.Notify(641349)
				}
			} else {
				__antithesis_instrumentation__.Notify(641350)
			}
			__antithesis_instrumentation__.Notify(641343)

			if isSys {
				__antithesis_instrumentation__.Notify(641351)
				ms.SysBytes += totalBytes
				ms.SysCount++
				if isAbortSpanKey(unsafeKey.Key) {
					__antithesis_instrumentation__.Notify(641352)
					ms.AbortSpanBytes += totalBytes
				} else {
					__antithesis_instrumentation__.Notify(641353)
				}
			} else {
				__antithesis_instrumentation__.Notify(641354)
				if !meta.Deleted {
					__antithesis_instrumentation__.Notify(641356)
					ms.LiveBytes += totalBytes
					ms.LiveCount++
				} else {
					__antithesis_instrumentation__.Notify(641357)

					ms.GCBytesAge += totalBytes * (nowNanos/1e9 - meta.Timestamp.WallTime/1e9)
				}
				__antithesis_instrumentation__.Notify(641355)
				ms.KeyBytes += metaKeySize
				ms.ValBytes += metaValSize
				ms.KeyCount++
				if meta.IsInline() {
					__antithesis_instrumentation__.Notify(641358)
					ms.ValCount++
				} else {
					__antithesis_instrumentation__.Notify(641359)
				}
			}
			__antithesis_instrumentation__.Notify(641344)
			if !implicitMeta {
				__antithesis_instrumentation__.Notify(641360)
				continue
			} else {
				__antithesis_instrumentation__.Notify(641361)
			}
		} else {
			__antithesis_instrumentation__.Notify(641362)
		}
		__antithesis_instrumentation__.Notify(641319)

		totalBytes := int64(len(unsafeValue)) + MVCCVersionTimestampSize
		if isSys {
			__antithesis_instrumentation__.Notify(641363)
			ms.SysBytes += totalBytes
		} else {
			__antithesis_instrumentation__.Notify(641364)
			if first {
				__antithesis_instrumentation__.Notify(641366)
				first = false
				if !meta.Deleted {
					__antithesis_instrumentation__.Notify(641371)
					ms.LiveBytes += totalBytes
				} else {
					__antithesis_instrumentation__.Notify(641372)

					ms.GCBytesAge += totalBytes * (nowNanos/1e9 - meta.Timestamp.WallTime/1e9)
				}
				__antithesis_instrumentation__.Notify(641367)
				if meta.Txn != nil {
					__antithesis_instrumentation__.Notify(641373)
					ms.IntentBytes += totalBytes
					ms.IntentCount++
					ms.SeparatedIntentCount++
					ms.IntentAge += nowNanos/1e9 - meta.Timestamp.WallTime/1e9
				} else {
					__antithesis_instrumentation__.Notify(641374)
				}
				__antithesis_instrumentation__.Notify(641368)
				if meta.KeyBytes != MVCCVersionTimestampSize {
					__antithesis_instrumentation__.Notify(641375)
					return ms, errors.Errorf("expected mvcc metadata key bytes to equal %d; got %d "+
						"(meta: %s)", MVCCVersionTimestampSize, meta.KeyBytes, &meta)
				} else {
					__antithesis_instrumentation__.Notify(641376)
				}
				__antithesis_instrumentation__.Notify(641369)
				if meta.ValBytes != int64(len(unsafeValue)) {
					__antithesis_instrumentation__.Notify(641377)
					return ms, errors.Errorf("expected mvcc metadata val bytes to equal %d; got %d "+
						"(meta: %s)", len(unsafeValue), meta.ValBytes, &meta)
				} else {
					__antithesis_instrumentation__.Notify(641378)
				}
				__antithesis_instrumentation__.Notify(641370)
				accrueGCAgeNanos = meta.Timestamp.WallTime
			} else {
				__antithesis_instrumentation__.Notify(641379)

				isTombstone := len(unsafeValue) == 0
				if isTombstone {
					__antithesis_instrumentation__.Notify(641381)

					ms.GCBytesAge += totalBytes * (nowNanos/1e9 - unsafeKey.Timestamp.WallTime/1e9)
				} else {
					__antithesis_instrumentation__.Notify(641382)

					ms.GCBytesAge += totalBytes * (nowNanos/1e9 - accrueGCAgeNanos/1e9)
				}
				__antithesis_instrumentation__.Notify(641380)

				accrueGCAgeNanos = unsafeKey.Timestamp.WallTime
			}
			__antithesis_instrumentation__.Notify(641365)
			ms.KeyBytes += MVCCVersionTimestampSize
			ms.ValBytes += int64(len(unsafeValue))
			ms.ValCount++
		}
	}
	__antithesis_instrumentation__.Notify(641312)

	ms.LastUpdateNanos = nowNanos
	return ms, nil
}
