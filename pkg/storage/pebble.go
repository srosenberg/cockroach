package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"github.com/cockroachdb/pebble/sstable"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/redact"
	humanize "github.com/dustin/go-humanize"
)

const maxSyncDurationFatalOnExceededDefault = true

var maxSyncDurationDefault = envutil.EnvOrDefaultDuration("COCKROACH_ENGINE_MAX_SYNC_DURATION_DEFAULT", 60*time.Second)

var MaxSyncDuration = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"storage.max_sync_duration",
	"maximum duration for disk operations; any operations that take longer"+
		" than this setting trigger a warning log entry or process crash",
	maxSyncDurationDefault,
)

var MaxSyncDurationFatalOnExceeded = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"storage.max_sync_duration.fatal.enabled",
	"if true, fatal the process when a disk operation exceeds storage.max_sync_duration",
	maxSyncDurationFatalOnExceededDefault,
)

func EngineKeyCompare(a, b []byte) int {
	__antithesis_instrumentation__.Notify(641699)

	aEnd := len(a) - 1
	bEnd := len(b) - 1
	if aEnd < 0 || func() bool {
		__antithesis_instrumentation__.Notify(641705)
		return bEnd < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641706)

		return bytes.Compare(a, b)
	} else {
		__antithesis_instrumentation__.Notify(641707)
	}
	__antithesis_instrumentation__.Notify(641700)

	aSep := aEnd - int(a[aEnd])
	bSep := bEnd - int(b[bEnd])
	if aSep == -1 && func() bool {
		__antithesis_instrumentation__.Notify(641708)
		return bSep == -1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641709)
		aSep, bSep = 0, 0
	} else {
		__antithesis_instrumentation__.Notify(641710)
	}
	__antithesis_instrumentation__.Notify(641701)
	if aSep < 0 || func() bool {
		__antithesis_instrumentation__.Notify(641711)
		return bSep < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641712)

		return bytes.Compare(a, b)
	} else {
		__antithesis_instrumentation__.Notify(641713)
	}
	__antithesis_instrumentation__.Notify(641702)

	if c := bytes.Compare(a[:aSep], b[:bSep]); c != 0 {
		__antithesis_instrumentation__.Notify(641714)
		return c
	} else {
		__antithesis_instrumentation__.Notify(641715)
	}
	__antithesis_instrumentation__.Notify(641703)

	aTS := a[aSep:aEnd]
	bTS := b[bSep:bEnd]
	if len(aTS) == 0 {
		__antithesis_instrumentation__.Notify(641716)
		if len(bTS) == 0 {
			__antithesis_instrumentation__.Notify(641718)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(641719)
		}
		__antithesis_instrumentation__.Notify(641717)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(641720)
		if len(bTS) == 0 {
			__antithesis_instrumentation__.Notify(641721)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(641722)
		}
	}
	__antithesis_instrumentation__.Notify(641704)
	return bytes.Compare(bTS, aTS)
}

var EngineComparer = &pebble.Comparer{
	Compare: EngineKeyCompare,

	AbbreviatedKey: func(k []byte) uint64 {
		__antithesis_instrumentation__.Notify(641723)
		key, ok := GetKeyPartFromEngineKey(k)
		if !ok {
			__antithesis_instrumentation__.Notify(641725)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(641726)
		}
		__antithesis_instrumentation__.Notify(641724)
		return pebble.DefaultComparer.AbbreviatedKey(key)
	},

	FormatKey: func(k []byte) fmt.Formatter {
		__antithesis_instrumentation__.Notify(641727)
		decoded, ok := DecodeEngineKey(k)
		if !ok {
			__antithesis_instrumentation__.Notify(641730)
			return mvccKeyFormatter{err: errors.Errorf("invalid encoded engine key: %x", k)}
		} else {
			__antithesis_instrumentation__.Notify(641731)
		}
		__antithesis_instrumentation__.Notify(641728)
		if decoded.IsMVCCKey() {
			__antithesis_instrumentation__.Notify(641732)
			mvccKey, err := decoded.ToMVCCKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(641734)
				return mvccKeyFormatter{err: err}
			} else {
				__antithesis_instrumentation__.Notify(641735)
			}
			__antithesis_instrumentation__.Notify(641733)
			return mvccKeyFormatter{key: mvccKey}
		} else {
			__antithesis_instrumentation__.Notify(641736)
		}
		__antithesis_instrumentation__.Notify(641729)
		return EngineKeyFormatter{key: decoded}
	},

	Separator: func(dst, a, b []byte) []byte {
		__antithesis_instrumentation__.Notify(641737)
		aKey, ok := GetKeyPartFromEngineKey(a)
		if !ok {
			__antithesis_instrumentation__.Notify(641742)
			return append(dst, a...)
		} else {
			__antithesis_instrumentation__.Notify(641743)
		}
		__antithesis_instrumentation__.Notify(641738)
		bKey, ok := GetKeyPartFromEngineKey(b)
		if !ok {
			__antithesis_instrumentation__.Notify(641744)
			return append(dst, a...)
		} else {
			__antithesis_instrumentation__.Notify(641745)
		}
		__antithesis_instrumentation__.Notify(641739)

		if bytes.Equal(aKey, bKey) {
			__antithesis_instrumentation__.Notify(641746)
			return append(dst, a...)
		} else {
			__antithesis_instrumentation__.Notify(641747)
		}
		__antithesis_instrumentation__.Notify(641740)
		n := len(dst)

		dst = pebble.DefaultComparer.Separator(dst, aKey, bKey)

		buf := dst[n:]
		if bytes.Equal(aKey, buf) {
			__antithesis_instrumentation__.Notify(641748)
			return append(dst[:n], a...)
		} else {
			__antithesis_instrumentation__.Notify(641749)
		}
		__antithesis_instrumentation__.Notify(641741)

		return append(dst, 0)
	},

	Successor: func(dst, a []byte) []byte {
		__antithesis_instrumentation__.Notify(641750)
		aKey, ok := GetKeyPartFromEngineKey(a)
		if !ok {
			__antithesis_instrumentation__.Notify(641753)
			return append(dst, a...)
		} else {
			__antithesis_instrumentation__.Notify(641754)
		}
		__antithesis_instrumentation__.Notify(641751)
		n := len(dst)

		dst = pebble.DefaultComparer.Successor(dst, aKey)

		buf := dst[n:]
		if bytes.Equal(aKey, buf) {
			__antithesis_instrumentation__.Notify(641755)
			return append(dst[:n], a...)
		} else {
			__antithesis_instrumentation__.Notify(641756)
		}
		__antithesis_instrumentation__.Notify(641752)

		return append(dst, 0)
	},

	Split: func(k []byte) int {
		__antithesis_instrumentation__.Notify(641757)
		key, ok := GetKeyPartFromEngineKey(k)
		if !ok {
			__antithesis_instrumentation__.Notify(641759)
			return len(k)
		} else {
			__antithesis_instrumentation__.Notify(641760)
		}
		__antithesis_instrumentation__.Notify(641758)

		return len(key) + 1
	},

	Name: "cockroach_comparator",
}

var MVCCMerger = &pebble.Merger{
	Name: "cockroach_merge_operator",
	Merge: func(_, value []byte) (pebble.ValueMerger, error) {
		__antithesis_instrumentation__.Notify(641761)
		res := &MVCCValueMerger{}
		err := res.MergeNewer(value)
		if err != nil {
			__antithesis_instrumentation__.Notify(641763)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(641764)
		}
		__antithesis_instrumentation__.Notify(641762)
		return res, nil
	},
}

type pebbleTimeBoundPropCollector struct {
	min, max  []byte
	lastValue []byte
}

func (t *pebbleTimeBoundPropCollector) Add(key pebble.InternalKey, value []byte) error {
	__antithesis_instrumentation__.Notify(641765)
	engineKey, ok := DecodeEngineKey(key.UserKey)
	if !ok {
		__antithesis_instrumentation__.Notify(641768)
		return errors.Errorf("failed to split engine key")
	} else {
		__antithesis_instrumentation__.Notify(641769)
	}
	__antithesis_instrumentation__.Notify(641766)
	if engineKey.IsMVCCKey() && func() bool {
		__antithesis_instrumentation__.Notify(641770)
		return len(engineKey.Version) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641771)
		t.lastValue = t.lastValue[:0]
		t.updateBounds(engineKey.Version)
	} else {
		__antithesis_instrumentation__.Notify(641772)
		t.lastValue = append(t.lastValue[:0], value...)
	}
	__antithesis_instrumentation__.Notify(641767)
	return nil
}

func (t *pebbleTimeBoundPropCollector) Finish(userProps map[string]string) error {
	__antithesis_instrumentation__.Notify(641773)
	if len(t.lastValue) > 0 {
		__antithesis_instrumentation__.Notify(641775)

		meta := &enginepb.MVCCMetadata{}
		if err := protoutil.Unmarshal(t.lastValue, meta); err != nil {
			__antithesis_instrumentation__.Notify(641777)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(641778)
		}
		__antithesis_instrumentation__.Notify(641776)
		if meta.Txn != nil {
			__antithesis_instrumentation__.Notify(641779)
			ts := encodeMVCCTimestamp(meta.Timestamp.ToTimestamp())
			t.updateBounds(ts)
		} else {
			__antithesis_instrumentation__.Notify(641780)
		}
	} else {
		__antithesis_instrumentation__.Notify(641781)
	}
	__antithesis_instrumentation__.Notify(641774)

	userProps["crdb.ts.min"] = string(t.min)
	userProps["crdb.ts.max"] = string(t.max)
	return nil
}

func (t *pebbleTimeBoundPropCollector) updateBounds(ts []byte) {
	__antithesis_instrumentation__.Notify(641782)
	if len(t.min) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(641784)
		return bytes.Compare(ts, t.min) < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641785)
		t.min = append(t.min[:0], ts...)
	} else {
		__antithesis_instrumentation__.Notify(641786)
	}
	__antithesis_instrumentation__.Notify(641783)
	if len(t.max) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(641787)
		return bytes.Compare(ts, t.max) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641788)
		t.max = append(t.max[:0], ts...)
	} else {
		__antithesis_instrumentation__.Notify(641789)
	}
}

func (t *pebbleTimeBoundPropCollector) Name() string {
	__antithesis_instrumentation__.Notify(641790)

	return "TimeBoundTblPropCollectorFactory"
}

func (t *pebbleTimeBoundPropCollector) UpdateKeySuffixes(
	oldProps map[string]string, oldSuffix []byte, newSuffix []byte,
) error {
	__antithesis_instrumentation__.Notify(641791)
	t.updateBounds(newSuffix)
	return nil
}

var _ sstable.SuffixReplaceableTableCollector = (*pebbleTimeBoundPropCollector)(nil)

type pebbleDeleteRangeCollector struct{}

func (pebbleDeleteRangeCollector) Add(_ pebble.InternalKey, _ []byte) error {
	__antithesis_instrumentation__.Notify(641792)
	return nil
}

func (pebbleDeleteRangeCollector) Finish(_ map[string]string) error {
	__antithesis_instrumentation__.Notify(641793)
	return nil
}

func (pebbleDeleteRangeCollector) Name() string {
	__antithesis_instrumentation__.Notify(641794)

	return "DeleteRangeTblPropCollectorFactory"
}

func (t *pebbleDeleteRangeCollector) UpdateKeySuffixes(
	_ map[string]string, _ []byte, _ []byte,
) error {
	__antithesis_instrumentation__.Notify(641795)
	return nil
}

var _ sstable.SuffixReplaceableTableCollector = (*pebbleTimeBoundPropCollector)(nil)

var PebbleTablePropertyCollectors = []func() pebble.TablePropertyCollector{
	func() pebble.TablePropertyCollector {
		__antithesis_instrumentation__.Notify(641796)
		return &pebbleTimeBoundPropCollector{}
	},
	func() pebble.TablePropertyCollector {
		__antithesis_instrumentation__.Notify(641797)
		return &pebbleDeleteRangeCollector{}
	},
}

type pebbleDataBlockMVCCTimeIntervalCollector struct {
	min, max []byte
}

var _ sstable.DataBlockIntervalCollector = &pebbleDataBlockMVCCTimeIntervalCollector{}
var _ sstable.SuffixReplaceableBlockCollector = (*pebbleDataBlockMVCCTimeIntervalCollector)(nil)

func (tc *pebbleDataBlockMVCCTimeIntervalCollector) Add(key pebble.InternalKey, _ []byte) error {
	__antithesis_instrumentation__.Notify(641798)
	return tc.add(key.UserKey)
}

func (tc *pebbleDataBlockMVCCTimeIntervalCollector) add(b []byte) error {
	__antithesis_instrumentation__.Notify(641799)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(641804)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641805)
	}
	__antithesis_instrumentation__.Notify(641800)

	versionLen := int(b[len(b)-1])
	if versionLen == 0 {
		__antithesis_instrumentation__.Notify(641806)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(641807)
	}
	__antithesis_instrumentation__.Notify(641801)

	prefixPartEnd := len(b) - 1 - versionLen

	if prefixPartEnd < -1 || func() bool {
		__antithesis_instrumentation__.Notify(641808)
		return (prefixPartEnd >= 0 && func() bool {
			__antithesis_instrumentation__.Notify(641809)
			return b[prefixPartEnd] != sentinel == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(641810)
		return errors.Errorf("invalid key %s", roachpb.Key(b).String())
	} else {
		__antithesis_instrumentation__.Notify(641811)
	}
	__antithesis_instrumentation__.Notify(641802)

	versionLen--

	if versionLen == engineKeyVersionWallTimeLen || func() bool {
		__antithesis_instrumentation__.Notify(641812)
		return versionLen == engineKeyVersionWallAndLogicalTimeLen == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(641813)
		return versionLen == engineKeyVersionWallLogicalAndSyntheticTimeLen == true
	}() == true {
		__antithesis_instrumentation__.Notify(641814)

		b = b[prefixPartEnd+1 : len(b)-1]

		if len(tc.min) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(641816)
			return bytes.Compare(b, tc.min) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(641817)
			tc.min = append(tc.min[:0], b...)
		} else {
			__antithesis_instrumentation__.Notify(641818)
		}
		__antithesis_instrumentation__.Notify(641815)
		if len(tc.max) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(641819)
			return bytes.Compare(b, tc.max) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(641820)
			tc.max = append(tc.max[:0], b...)
		} else {
			__antithesis_instrumentation__.Notify(641821)
		}
	} else {
		__antithesis_instrumentation__.Notify(641822)
	}
	__antithesis_instrumentation__.Notify(641803)
	return nil
}

func decodeWallTime(ts []byte) uint64 {
	__antithesis_instrumentation__.Notify(641823)
	return binary.BigEndian.Uint64(ts[0:engineKeyVersionWallTimeLen])
}

func (tc *pebbleDataBlockMVCCTimeIntervalCollector) FinishDataBlock() (
	lower uint64,
	upper uint64,
	err error,
) {
	__antithesis_instrumentation__.Notify(641824)
	if len(tc.min) == 0 {
		__antithesis_instrumentation__.Notify(641827)

		return 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(641828)
	}
	__antithesis_instrumentation__.Notify(641825)

	lower = decodeWallTime(tc.min)

	tc.min = tc.min[:0]

	upper = decodeWallTime(tc.max) + 1
	tc.max = tc.max[:0]
	if lower >= upper {
		__antithesis_instrumentation__.Notify(641829)
		return 0, 0,
			errors.Errorf("corrupt timestamps lower %d >= upper %d", lower, upper)
	} else {
		__antithesis_instrumentation__.Notify(641830)
	}
	__antithesis_instrumentation__.Notify(641826)
	return lower, upper, nil
}

func (tc *pebbleDataBlockMVCCTimeIntervalCollector) UpdateKeySuffixes(
	_ []byte, _, newSuffix []byte,
) error {
	__antithesis_instrumentation__.Notify(641831)
	return tc.add(newSuffix)
}

const mvccWallTimeIntervalCollector = "MVCCTimeInterval"

var PebbleBlockPropertyCollectors = []func() pebble.BlockPropertyCollector{
	func() pebble.BlockPropertyCollector {
		__antithesis_instrumentation__.Notify(641832)
		return sstable.NewBlockIntervalCollector(
			mvccWallTimeIntervalCollector,
			&pebbleDataBlockMVCCTimeIntervalCollector{},
			nil,
		)
	},
}

func DefaultPebbleOptions() *pebble.Options {
	__antithesis_instrumentation__.Notify(641833)

	maxConcurrentCompactions := rocksdbConcurrency - 1
	if maxConcurrentCompactions < 1 {
		__antithesis_instrumentation__.Notify(641840)
		maxConcurrentCompactions = 1
	} else {
		__antithesis_instrumentation__.Notify(641841)
	}
	__antithesis_instrumentation__.Notify(641834)

	opts := &pebble.Options{
		Comparer:                    EngineComparer,
		L0CompactionThreshold:       2,
		L0StopWritesThreshold:       1000,
		LBaseMaxBytes:               64 << 20,
		Levels:                      make([]pebble.LevelOptions, 7),
		MaxConcurrentCompactions:    maxConcurrentCompactions,
		MemTableSize:                64 << 20,
		MemTableStopWritesThreshold: 4,
		Merger:                      MVCCMerger,
		TablePropertyCollectors:     PebbleTablePropertyCollectors,
		BlockPropertyCollectors:     PebbleBlockPropertyCollectors,
	}

	opts.Experimental.DeleteRangeFlushDelay = 10 * time.Second

	opts.Experimental.MinDeletionRate = 128 << 20

	opts.Experimental.KeyValidationFunc = func(userKey []byte) error {
		__antithesis_instrumentation__.Notify(641842)
		engineKey, ok := DecodeEngineKey(userKey)
		if !ok {
			__antithesis_instrumentation__.Notify(641845)
			return errors.Newf("key %s could not be decoded as an EngineKey", string(userKey))
		} else {
			__antithesis_instrumentation__.Notify(641846)
		}
		__antithesis_instrumentation__.Notify(641843)
		if err := engineKey.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(641847)
			return err
		} else {
			__antithesis_instrumentation__.Notify(641848)
		}
		__antithesis_instrumentation__.Notify(641844)
		return nil
	}
	__antithesis_instrumentation__.Notify(641835)

	for i := 0; i < len(opts.Levels); i++ {
		__antithesis_instrumentation__.Notify(641849)
		l := &opts.Levels[i]
		l.BlockSize = 32 << 10
		l.IndexBlockSize = 256 << 10
		l.FilterPolicy = bloom.FilterPolicy(10)
		l.FilterType = pebble.TableFilter
		if i > 0 {
			__antithesis_instrumentation__.Notify(641851)
			l.TargetFileSize = opts.Levels[i-1].TargetFileSize * 2
		} else {
			__antithesis_instrumentation__.Notify(641852)
		}
		__antithesis_instrumentation__.Notify(641850)
		l.EnsureDefaults()
	}
	__antithesis_instrumentation__.Notify(641836)

	opts.Levels[6].FilterPolicy = nil

	diskHealthCheckInterval := 5 * time.Second
	if diskHealthCheckInterval.Seconds() > maxSyncDurationDefault.Seconds() {
		__antithesis_instrumentation__.Notify(641853)
		diskHealthCheckInterval = maxSyncDurationDefault
	} else {
		__antithesis_instrumentation__.Notify(641854)
	}
	__antithesis_instrumentation__.Notify(641837)

	opts.FS = vfs.WithDiskHealthChecks(vfs.Default, diskHealthCheckInterval,
		func(name string, duration time.Duration) {
			__antithesis_instrumentation__.Notify(641855)
			opts.EventListener.DiskSlow(pebble.DiskSlowInfo{
				Path:     name,
				Duration: duration,
			})
		})
	__antithesis_instrumentation__.Notify(641838)

	opts.FS = vfs.OnDiskFull(opts.FS, func() {
		__antithesis_instrumentation__.Notify(641856)
		exit.WithCode(exit.DiskFull())
	})
	__antithesis_instrumentation__.Notify(641839)
	return opts
}

type pebbleLogger struct {
	ctx   context.Context
	depth int
}

func (l pebbleLogger) Infof(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(641857)
	log.Storage.InfofDepth(l.ctx, l.depth, format, args...)
}

func (l pebbleLogger) Fatalf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(641858)
	log.Storage.FatalfDepth(l.ctx, l.depth, format, args...)
}

type PebbleConfig struct {
	base.StorageConfig

	Opts *pebble.Options
}

type EncryptionStatsHandler interface {
	GetEncryptionStatus() ([]byte, error)

	GetDataKeysRegistry() ([]byte, error)

	GetActiveDataKeyID() (string, error)

	GetActiveStoreKeyType() int32

	GetKeyIDFromSettings(settings []byte) (string, error)
}

type Pebble struct {
	db *pebble.DB

	closed      bool
	readOnly    bool
	path        string
	auxDir      string
	ballastPath string
	ballastSize int64
	maxSize     int64
	attrs       roachpb.Attributes
	properties  roachpb.StoreProperties

	settings     *cluster.Settings
	encryption   *EncryptionEnv
	fileRegistry *PebbleFileRegistry

	writeStallCount int64
	diskSlowCount   int64
	diskStallCount  int64

	fs            vfs.FS
	unencryptedFS vfs.FS
	logger        pebble.Logger
	eventListener *pebble.EventListener
	mu            struct {
		syncutil.Mutex
		flushCompletedCallback func()
	}

	wrappedIntentWriter intentDemuxWriter

	storeIDPebbleLog *base.StoreIDContainer
}

type EncryptionEnv struct {
	Closer io.Closer

	FS vfs.FS

	StatsHandler EncryptionStatsHandler
}

var _ Engine = &Pebble{}

var NewEncryptedEnvFunc func(fs vfs.FS, fr *PebbleFileRegistry, dbDir string, readOnly bool, optionBytes []byte) (*EncryptionEnv, error)

type StoreIDSetter interface {
	SetStoreID(ctx context.Context, storeID int32)
}

func (p *Pebble) SetStoreID(ctx context.Context, storeID int32) {
	__antithesis_instrumentation__.Notify(641859)
	if p == nil {
		__antithesis_instrumentation__.Notify(641862)
		return
	} else {
		__antithesis_instrumentation__.Notify(641863)
	}
	__antithesis_instrumentation__.Notify(641860)
	if p.storeIDPebbleLog == nil {
		__antithesis_instrumentation__.Notify(641864)
		return
	} else {
		__antithesis_instrumentation__.Notify(641865)
	}
	__antithesis_instrumentation__.Notify(641861)
	p.storeIDPebbleLog.Set(ctx, storeID)
}

func ResolveEncryptedEnvOptions(cfg *PebbleConfig) (*PebbleFileRegistry, *EncryptionEnv, error) {
	__antithesis_instrumentation__.Notify(641866)
	fileRegistry := &PebbleFileRegistry{FS: cfg.Opts.FS, DBDir: cfg.Dir, ReadOnly: cfg.Opts.ReadOnly}
	if cfg.UseFileRegistry {
		__antithesis_instrumentation__.Notify(641869)
		if err := fileRegistry.Load(); err != nil {
			__antithesis_instrumentation__.Notify(641870)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(641871)
		}
	} else {
		__antithesis_instrumentation__.Notify(641872)
		if err := fileRegistry.CheckNoRegistryFile(); err != nil {
			__antithesis_instrumentation__.Notify(641874)
			return nil, nil, fmt.Errorf("encryption was used on this store before, but no encryption flags " +
				"specified. You need a CCL build and must fully specify the --enterprise-encryption flag")
		} else {
			__antithesis_instrumentation__.Notify(641875)
		}
		__antithesis_instrumentation__.Notify(641873)
		fileRegistry = nil
	}
	__antithesis_instrumentation__.Notify(641867)

	var env *EncryptionEnv
	if cfg.IsEncrypted() {
		__antithesis_instrumentation__.Notify(641876)

		if !cfg.UseFileRegistry {
			__antithesis_instrumentation__.Notify(641880)
			return nil, nil, fmt.Errorf("file registry is needed to support encryption")
		} else {
			__antithesis_instrumentation__.Notify(641881)
		}
		__antithesis_instrumentation__.Notify(641877)
		if NewEncryptedEnvFunc == nil {
			__antithesis_instrumentation__.Notify(641882)
			return nil, nil, fmt.Errorf("encryption is enabled but no function to create the encrypted env")
		} else {
			__antithesis_instrumentation__.Notify(641883)
		}
		__antithesis_instrumentation__.Notify(641878)
		var err error
		env, err = NewEncryptedEnvFunc(
			cfg.Opts.FS,
			fileRegistry,
			cfg.Dir,
			cfg.Opts.ReadOnly,
			cfg.EncryptionOptions,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(641884)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(641885)
		}
		__antithesis_instrumentation__.Notify(641879)

		cfg.Opts.FS = env.FS
	} else {
		__antithesis_instrumentation__.Notify(641886)
	}
	__antithesis_instrumentation__.Notify(641868)
	return fileRegistry, env, nil
}

func NewPebble(ctx context.Context, cfg PebbleConfig) (*Pebble, error) {
	__antithesis_instrumentation__.Notify(641887)

	if cfg.Opts == nil {
		__antithesis_instrumentation__.Notify(641897)
		cfg.Opts = DefaultPebbleOptions()
	} else {
		__antithesis_instrumentation__.Notify(641898)
	}
	__antithesis_instrumentation__.Notify(641888)
	cfg.Opts.EnsureDefaults()
	cfg.Opts.ErrorIfNotExists = cfg.MustExist
	if settings := cfg.Settings; settings != nil {
		__antithesis_instrumentation__.Notify(641899)
		cfg.Opts.WALMinSyncInterval = func() time.Duration {
			__antithesis_instrumentation__.Notify(641900)
			return minWALSyncInterval.Get(&settings.SV)
		}
	} else {
		__antithesis_instrumentation__.Notify(641901)
	}
	__antithesis_instrumentation__.Notify(641889)

	auxDir := cfg.Opts.FS.PathJoin(cfg.Dir, base.AuxiliaryDir)
	if err := cfg.Opts.FS.MkdirAll(auxDir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(641902)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641903)
	}
	__antithesis_instrumentation__.Notify(641890)
	ballastPath := base.EmergencyBallastFile(cfg.Opts.FS.PathJoin, cfg.Dir)

	unencryptedFS := cfg.Opts.FS

	fileRegistry, env, err := ResolveEncryptedEnvOptions(&cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(641904)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641905)
	}
	__antithesis_instrumentation__.Notify(641891)

	logCtx := logtags.WithTags(context.Background(), logtags.FromContext(ctx))
	logCtx = logtags.AddTag(logCtx, "pebble", nil)

	storeIDContainer := &base.StoreIDContainer{}
	logCtx = logtags.AddTag(logCtx, "s", storeIDContainer)

	if cfg.Opts.Logger == nil {
		__antithesis_instrumentation__.Notify(641906)
		cfg.Opts.Logger = pebbleLogger{
			ctx:   logCtx,
			depth: 1,
		}
	} else {
		__antithesis_instrumentation__.Notify(641907)
	}
	__antithesis_instrumentation__.Notify(641892)

	if !cfg.Opts.ReadOnly {
		__antithesis_instrumentation__.Notify(641908)
		du, err := unencryptedFS.GetDiskUsage(cfg.Dir)

		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(641909)
			return !errors.Is(err, vfs.ErrUnsupported) == true
		}() == true {
			__antithesis_instrumentation__.Notify(641910)
			return nil, errors.Wrap(err, "retrieving disk usage")
		} else {
			__antithesis_instrumentation__.Notify(641911)
			if err == nil {
				__antithesis_instrumentation__.Notify(641912)
				resized, err := maybeEstablishBallast(unencryptedFS, ballastPath, cfg.BallastSize, du)
				if err != nil {
					__antithesis_instrumentation__.Notify(641914)
					return nil, errors.Wrap(err, "resizing ballast")
				} else {
					__antithesis_instrumentation__.Notify(641915)
				}
				__antithesis_instrumentation__.Notify(641913)
				if resized {
					__antithesis_instrumentation__.Notify(641916)
					cfg.Opts.Logger.Infof("resized ballast %s to size %s",
						ballastPath, humanizeutil.IBytes(cfg.BallastSize))
				} else {
					__antithesis_instrumentation__.Notify(641917)
				}
			} else {
				__antithesis_instrumentation__.Notify(641918)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(641919)
	}
	__antithesis_instrumentation__.Notify(641893)

	storeProps := computeStoreProperties(ctx, cfg.Dir, cfg.Opts.ReadOnly, env != nil)

	p := &Pebble{
		readOnly:         cfg.Opts.ReadOnly,
		path:             cfg.Dir,
		auxDir:           auxDir,
		ballastPath:      ballastPath,
		ballastSize:      cfg.BallastSize,
		maxSize:          cfg.MaxSize,
		attrs:            cfg.Attrs,
		properties:       storeProps,
		settings:         cfg.Settings,
		encryption:       env,
		fileRegistry:     fileRegistry,
		fs:               cfg.Opts.FS,
		unencryptedFS:    unencryptedFS,
		logger:           cfg.Opts.Logger,
		storeIDPebbleLog: storeIDContainer,
	}
	cfg.Opts.EventListener = pebble.TeeEventListener(
		pebble.MakeLoggingEventListener(pebbleLogger{
			ctx:   logCtx,
			depth: 2,
		}),
		p.makeMetricEtcEventListener(ctx),
	)
	p.eventListener = &cfg.Opts.EventListener
	p.wrappedIntentWriter = wrapIntentWriter(ctx, p)

	storeClusterVersion, err := getMinVersion(unencryptedFS, cfg.Dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(641920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641921)
	}
	__antithesis_instrumentation__.Notify(641894)

	db, err := pebble.Open(cfg.StorageConfig.Dir, cfg.Opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(641922)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641923)
	}
	__antithesis_instrumentation__.Notify(641895)
	p.db = db

	if storeClusterVersion != (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(641924)

		if err := p.SetMinVersion(storeClusterVersion); err != nil {
			__antithesis_instrumentation__.Notify(641925)
			p.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(641926)
		}
	} else {
		__antithesis_instrumentation__.Notify(641927)
	}
	__antithesis_instrumentation__.Notify(641896)

	return p, nil
}

func (p *Pebble) makeMetricEtcEventListener(ctx context.Context) pebble.EventListener {
	__antithesis_instrumentation__.Notify(641928)
	return pebble.EventListener{
		WriteStallBegin: func(info pebble.WriteStallBeginInfo) {
			__antithesis_instrumentation__.Notify(641929)
			atomic.AddInt64(&p.writeStallCount, 1)
		},
		DiskSlow: func(info pebble.DiskSlowInfo) {
			__antithesis_instrumentation__.Notify(641930)
			maxSyncDuration := maxSyncDurationDefault
			fatalOnExceeded := maxSyncDurationFatalOnExceededDefault
			if p.settings != nil {
				__antithesis_instrumentation__.Notify(641933)
				maxSyncDuration = MaxSyncDuration.Get(&p.settings.SV)
				fatalOnExceeded = MaxSyncDurationFatalOnExceeded.Get(&p.settings.SV)
			} else {
				__antithesis_instrumentation__.Notify(641934)
			}
			__antithesis_instrumentation__.Notify(641931)
			if info.Duration.Seconds() >= maxSyncDuration.Seconds() {
				__antithesis_instrumentation__.Notify(641935)
				atomic.AddInt64(&p.diskStallCount, 1)

				if fatalOnExceeded {
					__antithesis_instrumentation__.Notify(641937)
					log.Fatalf(ctx, "disk stall detected: pebble unable to write to %s in %.2f seconds",
						info.Path, redact.Safe(info.Duration.Seconds()))
				} else {
					__antithesis_instrumentation__.Notify(641938)
					log.Errorf(ctx, "disk stall detected: pebble unable to write to %s in %.2f seconds",
						info.Path, redact.Safe(info.Duration.Seconds()))
				}
				__antithesis_instrumentation__.Notify(641936)
				return
			} else {
				__antithesis_instrumentation__.Notify(641939)
			}
			__antithesis_instrumentation__.Notify(641932)
			atomic.AddInt64(&p.diskSlowCount, 1)
		},
		FlushEnd: func(info pebble.FlushInfo) {
			__antithesis_instrumentation__.Notify(641940)
			if info.Err != nil {
				__antithesis_instrumentation__.Notify(641942)
				return
			} else {
				__antithesis_instrumentation__.Notify(641943)
			}
			__antithesis_instrumentation__.Notify(641941)
			p.mu.Lock()
			cb := p.mu.flushCompletedCallback
			p.mu.Unlock()
			if cb != nil {
				__antithesis_instrumentation__.Notify(641944)
				cb()
			} else {
				__antithesis_instrumentation__.Notify(641945)
			}
		},
	}
}

func (p *Pebble) String() string {
	__antithesis_instrumentation__.Notify(641946)
	dir := p.path
	if dir == "" {
		__antithesis_instrumentation__.Notify(641949)
		dir = "<in-mem>"
	} else {
		__antithesis_instrumentation__.Notify(641950)
	}
	__antithesis_instrumentation__.Notify(641947)
	attrs := p.attrs.String()
	if attrs == "" {
		__antithesis_instrumentation__.Notify(641951)
		attrs = "<no-attributes>"
	} else {
		__antithesis_instrumentation__.Notify(641952)
	}
	__antithesis_instrumentation__.Notify(641948)
	return fmt.Sprintf("%s=%s", attrs, dir)
}

func (p *Pebble) Close() {
	__antithesis_instrumentation__.Notify(641953)
	if p.closed {
		__antithesis_instrumentation__.Notify(641956)
		p.logger.Infof("closing unopened pebble instance")
		return
	} else {
		__antithesis_instrumentation__.Notify(641957)
	}
	__antithesis_instrumentation__.Notify(641954)
	p.closed = true
	_ = p.db.Close()
	if p.fileRegistry != nil {
		__antithesis_instrumentation__.Notify(641958)
		_ = p.fileRegistry.Close()
	} else {
		__antithesis_instrumentation__.Notify(641959)
	}
	__antithesis_instrumentation__.Notify(641955)
	if p.encryption != nil {
		__antithesis_instrumentation__.Notify(641960)
		_ = p.encryption.Closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(641961)
	}
}

func (p *Pebble) Closed() bool {
	__antithesis_instrumentation__.Notify(641962)
	return p.closed
}

func (p *Pebble) ExportMVCCToSst(
	ctx context.Context, exportOptions ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, roachpb.Key, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(641963)
	r := wrapReader(p)

	summary, k, err := pebbleExportToSst(ctx, p.settings, r, exportOptions, dest)
	r.Free()
	return summary, k.Key, k.Timestamp, err
}

func (p *Pebble) MVCCGet(key MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(641964)
	return mvccGetHelper(key, p)
}

func mvccGetHelper(key MVCCKey, reader wrappableReader) ([]byte, error) {
	__antithesis_instrumentation__.Notify(641965)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(641967)
		return nil, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(641968)
	}
	__antithesis_instrumentation__.Notify(641966)
	r := wrapReader(reader)

	v, err := r.MVCCGet(key)
	r.Free()
	return v, err
}

func (p *Pebble) rawMVCCGet(key []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(641969)
	ret, closer, err := p.db.Get(key)
	if closer != nil {
		__antithesis_instrumentation__.Notify(641972)
		retCopy := make([]byte, len(ret))
		copy(retCopy, ret)
		ret = retCopy
		closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(641973)
	}
	__antithesis_instrumentation__.Notify(641970)
	if errors.Is(err, pebble.ErrNotFound) || func() bool {
		__antithesis_instrumentation__.Notify(641974)
		return len(ret) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(641975)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(641976)
	}
	__antithesis_instrumentation__.Notify(641971)
	return ret, err
}

func (p *Pebble) MVCCGetProto(
	key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(641977)
	return pebbleGetProto(p, key, msg)
}

func (p *Pebble) MVCCIterate(
	start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(641978)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(641980)
		r := wrapReader(p)

		err := iterateOnReader(r, start, end, iterKind, f)
		r.Free()
		return err
	} else {
		__antithesis_instrumentation__.Notify(641981)
	}
	__antithesis_instrumentation__.Notify(641979)
	return iterateOnReader(p, start, end, iterKind, f)
}

func (p *Pebble) NewMVCCIterator(iterKind MVCCIterKind, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(641982)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(641986)
		r := wrapReader(p)

		iter := r.NewMVCCIterator(iterKind, opts)
		r.Free()
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(641988)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(641989)
		}
		__antithesis_instrumentation__.Notify(641987)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(641990)
	}
	__antithesis_instrumentation__.Notify(641983)

	iter := newPebbleIterator(p.db, nil, opts, StandardDurability)
	if iter == nil {
		__antithesis_instrumentation__.Notify(641991)
		panic("couldn't create a new iterator")
	} else {
		__antithesis_instrumentation__.Notify(641992)
	}
	__antithesis_instrumentation__.Notify(641984)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(641993)
		return wrapInUnsafeIter(iter)
	} else {
		__antithesis_instrumentation__.Notify(641994)
	}
	__antithesis_instrumentation__.Notify(641985)
	return iter
}

func (p *Pebble) NewEngineIterator(opts IterOptions) EngineIterator {
	__antithesis_instrumentation__.Notify(641995)
	iter := newPebbleIterator(p.db, nil, opts, StandardDurability)
	if iter == nil {
		__antithesis_instrumentation__.Notify(641997)
		panic("couldn't create a new iterator")
	} else {
		__antithesis_instrumentation__.Notify(641998)
	}
	__antithesis_instrumentation__.Notify(641996)
	return iter
}

func (p *Pebble) ConsistentIterators() bool {
	__antithesis_instrumentation__.Notify(641999)
	return false
}

func (p *Pebble) PinEngineStateForIterators() error {
	__antithesis_instrumentation__.Notify(642000)
	return errors.AssertionFailedf(
		"PinEngineStateForIterators must not be called when ConsistentIterators returns false")
}

func (p *Pebble) ApplyBatchRepr(repr []byte, sync bool) error {
	__antithesis_instrumentation__.Notify(642001)

	reprCopy := make([]byte, len(repr))
	copy(reprCopy, repr)

	batch := p.db.NewBatch()
	if err := batch.SetRepr(reprCopy); err != nil {
		__antithesis_instrumentation__.Notify(642004)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642005)
	}
	__antithesis_instrumentation__.Notify(642002)

	opts := pebble.NoSync
	if sync {
		__antithesis_instrumentation__.Notify(642006)
		opts = pebble.Sync
	} else {
		__antithesis_instrumentation__.Notify(642007)
	}
	__antithesis_instrumentation__.Notify(642003)
	return batch.Commit(opts)
}

func (p *Pebble) ClearMVCC(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(642008)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(642010)
		panic("ClearMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(642011)
	}
	__antithesis_instrumentation__.Notify(642009)
	return p.clear(key)
}

func (p *Pebble) ClearUnversioned(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642012)
	return p.clear(MVCCKey{Key: key})
}

func (p *Pebble) ClearIntent(key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID) error {
	__antithesis_instrumentation__.Notify(642013)
	_, err := p.wrappedIntentWriter.ClearIntent(key, txnDidNotUpdateMeta, txnUUID, nil)
	return err
}

func (p *Pebble) ClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642014)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642016)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642017)
	}
	__antithesis_instrumentation__.Notify(642015)
	return p.db.Delete(key.Encode(), pebble.Sync)
}

func (p *Pebble) clear(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(642018)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642020)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642021)
	}
	__antithesis_instrumentation__.Notify(642019)
	return p.db.Delete(EncodeMVCCKey(key), pebble.Sync)
}

func (p *Pebble) SingleClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642022)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642024)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642025)
	}
	__antithesis_instrumentation__.Notify(642023)
	return p.db.SingleDelete(key.Encode(), pebble.Sync)
}

func (p *Pebble) ClearRawRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642026)
	return p.clearRange(MVCCKey{Key: start}, MVCCKey{Key: end})
}

func (p *Pebble) ClearMVCCRangeAndIntents(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642027)
	_, err := p.wrappedIntentWriter.ClearMVCCRangeAndIntents(start, end, nil)
	return err

}

func (p *Pebble) ClearMVCCRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(642028)
	return p.clearRange(start, end)
}

func (p *Pebble) clearRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(642029)
	bufStart := EncodeMVCCKey(start)
	bufEnd := EncodeMVCCKey(end)
	return p.db.DeleteRange(bufStart, bufEnd, pebble.Sync)
}

func (p *Pebble) ClearIterRange(iter MVCCIterator, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642030)

	batch := p.NewUnindexedBatch(true)
	defer batch.Close()

	if err := batch.ClearIterRange(iter, start, end); err != nil {
		__antithesis_instrumentation__.Notify(642032)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642033)
	}
	__antithesis_instrumentation__.Notify(642031)
	return batch.Commit(true)
}

func (p *Pebble) Merge(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642034)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642036)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642037)
	}
	__antithesis_instrumentation__.Notify(642035)
	return p.db.Merge(EncodeMVCCKey(key), value, pebble.Sync)
}

func (p *Pebble) PutMVCC(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642038)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(642040)
		panic("PutMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(642041)
	}
	__antithesis_instrumentation__.Notify(642039)
	return p.put(key, value)
}

func (p *Pebble) PutUnversioned(key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(642042)
	return p.put(MVCCKey{Key: key}, value)
}

func (p *Pebble) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(642043)
	_, err := p.wrappedIntentWriter.PutIntent(ctx, key, value, txnUUID, nil)
	return err
}

func (p *Pebble) PutEngineKey(key EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642044)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642046)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642047)
	}
	__antithesis_instrumentation__.Notify(642045)
	return p.db.Set(key.Encode(), value, pebble.Sync)
}

func (p *Pebble) put(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642048)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642050)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642051)
	}
	__antithesis_instrumentation__.Notify(642049)
	return p.db.Set(EncodeMVCCKey(key), value, pebble.Sync)
}

func (p *Pebble) LogData(data []byte) error {
	__antithesis_instrumentation__.Notify(642052)
	return p.db.LogData(data, pebble.Sync)
}

func (p *Pebble) LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(642053)

}

func (p *Pebble) Attrs() roachpb.Attributes {
	__antithesis_instrumentation__.Notify(642054)
	return p.attrs
}

func (p *Pebble) Properties() roachpb.StoreProperties {
	__antithesis_instrumentation__.Notify(642055)
	return p.properties
}

func (p *Pebble) Capacity() (roachpb.StoreCapacity, error) {
	__antithesis_instrumentation__.Notify(642056)
	dir := p.path
	if dir != "" {
		__antithesis_instrumentation__.Notify(642066)
		var err error

		if dir, err = filepath.EvalSymlinks(dir); err != nil {
			__antithesis_instrumentation__.Notify(642067)
			return roachpb.StoreCapacity{}, err
		} else {
			__antithesis_instrumentation__.Notify(642068)
		}
	} else {
		__antithesis_instrumentation__.Notify(642069)
	}
	__antithesis_instrumentation__.Notify(642057)
	du, err := p.unencryptedFS.GetDiskUsage(dir)
	if errors.Is(err, vfs.ErrUnsupported) {
		__antithesis_instrumentation__.Notify(642070)

		return roachpb.StoreCapacity{
			Capacity:  p.maxSize,
			Available: p.maxSize,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(642071)
		if err != nil {
			__antithesis_instrumentation__.Notify(642072)
			return roachpb.StoreCapacity{}, err
		} else {
			__antithesis_instrumentation__.Notify(642073)
		}
	}
	__antithesis_instrumentation__.Notify(642058)

	if du.TotalBytes > math.MaxInt64 {
		__antithesis_instrumentation__.Notify(642074)
		return roachpb.StoreCapacity{}, fmt.Errorf("unsupported disk size %s, max supported size is %s",
			humanize.IBytes(du.TotalBytes), humanizeutil.IBytes(math.MaxInt64))
	} else {
		__antithesis_instrumentation__.Notify(642075)
	}
	__antithesis_instrumentation__.Notify(642059)
	if du.AvailBytes > math.MaxInt64 {
		__antithesis_instrumentation__.Notify(642076)
		return roachpb.StoreCapacity{}, fmt.Errorf("unsupported disk size %s, max supported size is %s",
			humanize.IBytes(du.AvailBytes), humanizeutil.IBytes(math.MaxInt64))
	} else {
		__antithesis_instrumentation__.Notify(642077)
	}
	__antithesis_instrumentation__.Notify(642060)
	fsuTotal := int64(du.TotalBytes)
	fsuAvail := int64(du.AvailBytes)

	if !p.readOnly {
		__antithesis_instrumentation__.Notify(642078)
		resized, err := maybeEstablishBallast(p.unencryptedFS, p.ballastPath, p.ballastSize, du)
		if err != nil {
			__antithesis_instrumentation__.Notify(642080)
			return roachpb.StoreCapacity{}, errors.Wrap(err, "resizing ballast")
		} else {
			__antithesis_instrumentation__.Notify(642081)
		}
		__antithesis_instrumentation__.Notify(642079)
		if resized {
			__antithesis_instrumentation__.Notify(642082)
			p.logger.Infof("resized ballast %s to size %s",
				p.ballastPath, humanizeutil.IBytes(p.ballastSize))
			du, err = p.unencryptedFS.GetDiskUsage(dir)
			if err != nil {
				__antithesis_instrumentation__.Notify(642083)
				return roachpb.StoreCapacity{}, err
			} else {
				__antithesis_instrumentation__.Notify(642084)
			}
		} else {
			__antithesis_instrumentation__.Notify(642085)
		}
	} else {
		__antithesis_instrumentation__.Notify(642086)
	}
	__antithesis_instrumentation__.Notify(642061)

	m := p.db.Metrics()
	totalUsedBytes := int64(m.DiskSpaceUsage())

	if errOuter := filepath.Walk(p.auxDir, func(path string, info os.FileInfo, err error) error {
		__antithesis_instrumentation__.Notify(642087)
		if err != nil {
			__antithesis_instrumentation__.Notify(642091)

			if oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(642094)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(642095)
			}
			__antithesis_instrumentation__.Notify(642092)

			if oserror.IsPermission(err) && func() bool {
				__antithesis_instrumentation__.Notify(642096)
				return filepath.Base(path) == "lost+found" == true
			}() == true {
				__antithesis_instrumentation__.Notify(642097)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(642098)
			}
			__antithesis_instrumentation__.Notify(642093)
			return err
		} else {
			__antithesis_instrumentation__.Notify(642099)
		}
		__antithesis_instrumentation__.Notify(642088)
		if path == p.ballastPath {
			__antithesis_instrumentation__.Notify(642100)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(642101)
		}
		__antithesis_instrumentation__.Notify(642089)
		if info.Mode().IsRegular() {
			__antithesis_instrumentation__.Notify(642102)
			totalUsedBytes += info.Size()
		} else {
			__antithesis_instrumentation__.Notify(642103)
		}
		__antithesis_instrumentation__.Notify(642090)
		return nil
	}); errOuter != nil {
		__antithesis_instrumentation__.Notify(642104)
		return roachpb.StoreCapacity{}, errOuter
	} else {
		__antithesis_instrumentation__.Notify(642105)
	}
	__antithesis_instrumentation__.Notify(642062)

	if p.maxSize == 0 || func() bool {
		__antithesis_instrumentation__.Notify(642106)
		return p.maxSize >= fsuTotal == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(642107)
		return p.path == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(642108)
		return roachpb.StoreCapacity{
			Capacity:  fsuTotal,
			Available: fsuAvail,
			Used:      totalUsedBytes,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(642109)
	}
	__antithesis_instrumentation__.Notify(642063)

	available := p.maxSize - totalUsedBytes
	if available > fsuAvail {
		__antithesis_instrumentation__.Notify(642110)
		available = fsuAvail
	} else {
		__antithesis_instrumentation__.Notify(642111)
	}
	__antithesis_instrumentation__.Notify(642064)
	if available < 0 {
		__antithesis_instrumentation__.Notify(642112)
		available = 0
	} else {
		__antithesis_instrumentation__.Notify(642113)
	}
	__antithesis_instrumentation__.Notify(642065)

	return roachpb.StoreCapacity{
		Capacity:  p.maxSize,
		Available: available,
		Used:      totalUsedBytes,
	}, nil
}

func (p *Pebble) Flush() error {
	__antithesis_instrumentation__.Notify(642114)
	return p.db.Flush()
}

func (p *Pebble) GetMetrics() Metrics {
	__antithesis_instrumentation__.Notify(642115)
	m := p.db.Metrics()
	return Metrics{
		Metrics:         m,
		WriteStallCount: atomic.LoadInt64(&p.writeStallCount),
		DiskSlowCount:   atomic.LoadInt64(&p.diskSlowCount),
		DiskStallCount:  atomic.LoadInt64(&p.diskStallCount),
	}
}

func (p *Pebble) GetEncryptionRegistries() (*EncryptionRegistries, error) {
	__antithesis_instrumentation__.Notify(642116)
	rv := &EncryptionRegistries{}
	var err error
	if p.encryption != nil {
		__antithesis_instrumentation__.Notify(642119)
		rv.KeyRegistry, err = p.encryption.StatsHandler.GetDataKeysRegistry()
		if err != nil {
			__antithesis_instrumentation__.Notify(642120)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(642121)
		}
	} else {
		__antithesis_instrumentation__.Notify(642122)
	}
	__antithesis_instrumentation__.Notify(642117)
	if p.fileRegistry != nil {
		__antithesis_instrumentation__.Notify(642123)
		rv.FileRegistry, err = protoutil.Marshal(p.fileRegistry.getRegistryCopy())
		if err != nil {
			__antithesis_instrumentation__.Notify(642124)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(642125)
		}
	} else {
		__antithesis_instrumentation__.Notify(642126)
	}
	__antithesis_instrumentation__.Notify(642118)
	return rv, nil
}

func (p *Pebble) GetEnvStats() (*EnvStats, error) {
	__antithesis_instrumentation__.Notify(642127)

	stats := &EnvStats{}
	if p.encryption == nil {
		__antithesis_instrumentation__.Notify(642135)
		return stats, nil
	} else {
		__antithesis_instrumentation__.Notify(642136)
	}
	__antithesis_instrumentation__.Notify(642128)
	stats.EncryptionType = p.encryption.StatsHandler.GetActiveStoreKeyType()
	var err error
	stats.EncryptionStatus, err = p.encryption.StatsHandler.GetEncryptionStatus()
	if err != nil {
		__antithesis_instrumentation__.Notify(642137)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(642138)
	}
	__antithesis_instrumentation__.Notify(642129)
	fr := p.fileRegistry.getRegistryCopy()
	activeKeyID, err := p.encryption.StatsHandler.GetActiveDataKeyID()
	if err != nil {
		__antithesis_instrumentation__.Notify(642139)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(642140)
	}
	__antithesis_instrumentation__.Notify(642130)

	m := p.db.Metrics()
	stats.TotalFiles = 3
	stats.TotalFiles += uint64(m.WAL.Files + m.Table.ZombieCount + m.WAL.ObsoleteFiles + m.Table.ObsoleteCount)
	stats.TotalBytes = m.WAL.Size + m.Table.ZombieSize + m.Table.ObsoleteSize
	for _, l := range m.Levels {
		__antithesis_instrumentation__.Notify(642141)
		stats.TotalFiles += uint64(l.NumFiles)
		stats.TotalBytes += uint64(l.Size)
	}
	__antithesis_instrumentation__.Notify(642131)

	sstSizes := make(map[pebble.FileNum]uint64)
	sstInfos, err := p.db.SSTables()
	if err != nil {
		__antithesis_instrumentation__.Notify(642142)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(642143)
	}
	__antithesis_instrumentation__.Notify(642132)
	for _, ssts := range sstInfos {
		__antithesis_instrumentation__.Notify(642144)
		for _, sst := range ssts {
			__antithesis_instrumentation__.Notify(642145)
			sstSizes[sst.FileNum] = sst.Size
		}
	}
	__antithesis_instrumentation__.Notify(642133)

	for filePath, entry := range fr.Files {
		__antithesis_instrumentation__.Notify(642146)
		keyID, err := p.encryption.StatsHandler.GetKeyIDFromSettings(entry.EncryptionSettings)
		if err != nil {
			__antithesis_instrumentation__.Notify(642152)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(642153)
		}
		__antithesis_instrumentation__.Notify(642147)
		if len(keyID) == 0 {
			__antithesis_instrumentation__.Notify(642154)
			keyID = "plain"
		} else {
			__antithesis_instrumentation__.Notify(642155)
		}
		__antithesis_instrumentation__.Notify(642148)
		if keyID != activeKeyID {
			__antithesis_instrumentation__.Notify(642156)
			continue
		} else {
			__antithesis_instrumentation__.Notify(642157)
		}
		__antithesis_instrumentation__.Notify(642149)
		stats.ActiveKeyFiles++

		filename := p.fs.PathBase(filePath)
		numStr := strings.TrimSuffix(filename, ".sst")
		if len(numStr) == len(filename) {
			__antithesis_instrumentation__.Notify(642158)
			continue
		} else {
			__antithesis_instrumentation__.Notify(642159)
		}
		__antithesis_instrumentation__.Notify(642150)
		u, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(642160)
			return nil, errors.Wrapf(err, "parsing filename %q", errors.Safe(filename))
		} else {
			__antithesis_instrumentation__.Notify(642161)
		}
		__antithesis_instrumentation__.Notify(642151)
		stats.ActiveKeyBytes += sstSizes[pebble.FileNum(u)]
	}
	__antithesis_instrumentation__.Notify(642134)
	return stats, nil
}

func (p *Pebble) GetAuxiliaryDir() string {
	__antithesis_instrumentation__.Notify(642162)
	return p.auxDir
}

func (p *Pebble) NewBatch() Batch {
	__antithesis_instrumentation__.Notify(642163)
	return newPebbleBatch(p.db, p.db.NewIndexedBatch(), false)
}

func (p *Pebble) NewReadOnly(durability DurabilityRequirement) ReadWriter {
	__antithesis_instrumentation__.Notify(642164)
	return newPebbleReadOnly(p, durability)
}

func (p *Pebble) NewUnindexedBatch(writeOnly bool) Batch {
	__antithesis_instrumentation__.Notify(642165)
	return newPebbleBatch(p.db, p.db.NewBatch(), writeOnly)
}

func (p *Pebble) NewSnapshot() Reader {
	__antithesis_instrumentation__.Notify(642166)
	return &pebbleSnapshot{
		snapshot: p.db.NewSnapshot(),
		settings: p.settings,
	}
}

func (p *Pebble) Type() enginepb.EngineType {
	__antithesis_instrumentation__.Notify(642167)
	return enginepb.EngineTypePebble
}

func (p *Pebble) IngestExternalFiles(ctx context.Context, paths []string) error {
	__antithesis_instrumentation__.Notify(642168)
	return p.db.Ingest(paths)
}

func (p *Pebble) PreIngestDelay(ctx context.Context) {
	__antithesis_instrumentation__.Notify(642169)
	preIngestDelay(ctx, p, p.settings)
}

func (p *Pebble) ApproximateDiskBytes(from, to roachpb.Key) (uint64, error) {
	__antithesis_instrumentation__.Notify(642170)
	count, err := p.db.EstimateDiskUsage(from, to)
	if err != nil {
		__antithesis_instrumentation__.Notify(642172)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(642173)
	}
	__antithesis_instrumentation__.Notify(642171)
	return count, nil
}

func (p *Pebble) Compact() error {
	__antithesis_instrumentation__.Notify(642174)
	return p.db.Compact(nil, EncodeMVCCKey(MVCCKeyMax), true)
}

func (p *Pebble) CompactRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642175)
	bufStart := EncodeMVCCKey(MVCCKey{start, hlc.Timestamp{}})
	bufEnd := EncodeMVCCKey(MVCCKey{end, hlc.Timestamp{}})
	return p.db.Compact(bufStart, bufEnd, true)
}

func (p *Pebble) InMem() bool {
	__antithesis_instrumentation__.Notify(642176)
	return p.path == ""
}

func (p *Pebble) RegisterFlushCompletedCallback(cb func()) {
	__antithesis_instrumentation__.Notify(642177)
	p.mu.Lock()
	p.mu.flushCompletedCallback = cb
	p.mu.Unlock()
}

func (p *Pebble) ReadFile(filename string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642178)
	file, err := p.fs.Open(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(642180)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(642181)
	}
	__antithesis_instrumentation__.Notify(642179)
	defer file.Close()

	return ioutil.ReadAll(file)
}

func (p *Pebble) WriteFile(filename string, data []byte) error {
	__antithesis_instrumentation__.Notify(642182)
	file, err := p.fs.Create(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(642184)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642185)
	}
	__antithesis_instrumentation__.Notify(642183)
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(data))
	return err
}

func (p *Pebble) Remove(filename string) error {
	__antithesis_instrumentation__.Notify(642186)
	return p.fs.Remove(filename)
}

func (p *Pebble) RemoveAll(dir string) error {
	__antithesis_instrumentation__.Notify(642187)
	return p.fs.RemoveAll(dir)
}

func (p *Pebble) Link(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(642188)
	return p.fs.Link(oldname, newname)
}

var _ fs.FS = &Pebble{}

func (p *Pebble) Create(name string) (fs.File, error) {
	__antithesis_instrumentation__.Notify(642189)
	return p.fs.Create(name)
}

func (p *Pebble) CreateWithSync(name string, bytesPerSync int) (fs.File, error) {
	__antithesis_instrumentation__.Notify(642190)
	f, err := p.fs.Create(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(642192)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(642193)
	}
	__antithesis_instrumentation__.Notify(642191)
	return vfs.NewSyncingFile(f, vfs.SyncingFileOptions{BytesPerSync: bytesPerSync}), nil
}

func (p *Pebble) Open(name string) (fs.File, error) {
	__antithesis_instrumentation__.Notify(642194)
	return p.fs.Open(name)
}

func (p *Pebble) OpenDir(name string) (fs.File, error) {
	__antithesis_instrumentation__.Notify(642195)
	return p.fs.OpenDir(name)
}

func (p *Pebble) Rename(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(642196)
	return p.fs.Rename(oldname, newname)
}

func (p *Pebble) MkdirAll(name string) error {
	__antithesis_instrumentation__.Notify(642197)
	return p.fs.MkdirAll(name, 0755)
}

func (p *Pebble) List(name string) ([]string, error) {
	__antithesis_instrumentation__.Notify(642198)
	dirents, err := p.fs.List(name)
	sort.Strings(dirents)
	return dirents, err
}

func (p *Pebble) Stat(name string) (os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(642199)
	return p.fs.Stat(name)
}

func (p *Pebble) CreateCheckpoint(dir string) error {
	__antithesis_instrumentation__.Notify(642200)
	return p.db.Checkpoint(dir)
}

func (p *Pebble) SetMinVersion(version roachpb.Version) error {
	__antithesis_instrumentation__.Notify(642201)

	if err := writeMinVersionFile(p.unencryptedFS, p.path, version); err != nil {
		__antithesis_instrumentation__.Notify(642205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642206)
	}
	__antithesis_instrumentation__.Notify(642202)

	formatVers := pebble.FormatMostCompatible

	switch {
	case !version.Less(clusterversion.ByKey(clusterversion.PebbleFormatSplitUserKeysMarked)):
		__antithesis_instrumentation__.Notify(642207)
		if formatVers < pebble.FormatSplitUserKeysMarked {
			__antithesis_instrumentation__.Notify(642211)
			formatVers = pebble.FormatSplitUserKeysMarked
		} else {
			__antithesis_instrumentation__.Notify(642212)
		}
	case !version.Less(clusterversion.ByKey(clusterversion.PebbleFormatBlockPropertyCollector)):
		__antithesis_instrumentation__.Notify(642208)
		if formatVers < pebble.FormatBlockPropertyCollector {
			__antithesis_instrumentation__.Notify(642213)
			formatVers = pebble.FormatBlockPropertyCollector
		} else {
			__antithesis_instrumentation__.Notify(642214)
		}
	case !version.Less(clusterversion.ByKey(clusterversion.TODOPreV21_2)):
		__antithesis_instrumentation__.Notify(642209)
		if formatVers < pebble.FormatSetWithDelete {
			__antithesis_instrumentation__.Notify(642215)
			formatVers = pebble.FormatSetWithDelete
		} else {
			__antithesis_instrumentation__.Notify(642216)
		}
	default:
		__antithesis_instrumentation__.Notify(642210)
	}
	__antithesis_instrumentation__.Notify(642203)
	if p.db.FormatMajorVersion() < formatVers {
		__antithesis_instrumentation__.Notify(642217)
		if err := p.db.RatchetFormatMajorVersion(formatVers); err != nil {
			__antithesis_instrumentation__.Notify(642218)
			return errors.Wrap(err, "ratcheting format major version")
		} else {
			__antithesis_instrumentation__.Notify(642219)
		}
	} else {
		__antithesis_instrumentation__.Notify(642220)
	}
	__antithesis_instrumentation__.Notify(642204)
	return nil
}

func (p *Pebble) MinVersionIsAtLeastTargetVersion(target roachpb.Version) (bool, error) {
	__antithesis_instrumentation__.Notify(642221)
	return MinVersionIsAtLeastTargetVersion(p.unencryptedFS, p.path, target)
}

type pebbleReadOnly struct {
	parent *Pebble

	prefixIter       pebbleIterator
	normalIter       pebbleIterator
	prefixEngineIter pebbleIterator
	normalEngineIter pebbleIterator
	iter             cloneableIter
	durability       DurabilityRequirement
	closed           bool
}

var _ ReadWriter = &pebbleReadOnly{}

var pebbleReadOnlyPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(642222)
		return &pebbleReadOnly{

			prefixIter:       pebbleIterator{reusable: true},
			normalIter:       pebbleIterator{reusable: true},
			prefixEngineIter: pebbleIterator{reusable: true},
			normalEngineIter: pebbleIterator{reusable: true},
		}
	},
}

func newPebbleReadOnly(parent *Pebble, durability DurabilityRequirement) *pebbleReadOnly {
	__antithesis_instrumentation__.Notify(642223)
	p := pebbleReadOnlyPool.Get().(*pebbleReadOnly)

	*p = pebbleReadOnly{
		parent:           parent,
		prefixIter:       p.prefixIter,
		normalIter:       p.normalIter,
		prefixEngineIter: p.prefixEngineIter,
		normalEngineIter: p.normalEngineIter,
		durability:       durability,
	}
	return p
}

func (p *pebbleReadOnly) Close() {
	__antithesis_instrumentation__.Notify(642224)
	if p.closed {
		__antithesis_instrumentation__.Notify(642226)
		panic("closing an already-closed pebbleReadOnly")
	} else {
		__antithesis_instrumentation__.Notify(642227)
	}
	__antithesis_instrumentation__.Notify(642225)
	p.closed = true

	p.iter = nil
	p.prefixIter.destroy()
	p.normalIter.destroy()
	p.prefixEngineIter.destroy()
	p.normalEngineIter.destroy()
	p.durability = StandardDurability

	pebbleReadOnlyPool.Put(p)
}

func (p *pebbleReadOnly) Closed() bool {
	__antithesis_instrumentation__.Notify(642228)
	return p.closed
}

func (p *pebbleReadOnly) ExportMVCCToSst(
	ctx context.Context, exportOptions ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, roachpb.Key, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(642229)
	r := wrapReader(p)

	summary, k, err := pebbleExportToSst(ctx, p.parent.settings, r, exportOptions, dest)
	r.Free()
	return summary, k.Key, k.Timestamp, err
}

func (p *pebbleReadOnly) MVCCGet(key MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642230)
	return mvccGetHelper(key, p)
}

func (p *pebbleReadOnly) rawMVCCGet(key []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642231)
	if p.closed {
		__antithesis_instrumentation__.Notify(642236)
		panic("using a closed pebbleReadOnly")
	} else {
		__antithesis_instrumentation__.Notify(642237)
	}
	__antithesis_instrumentation__.Notify(642232)

	onlyReadGuaranteedDurable := false
	if p.durability == GuaranteedDurability {
		__antithesis_instrumentation__.Notify(642238)
		onlyReadGuaranteedDurable = true
	} else {
		__antithesis_instrumentation__.Notify(642239)
	}
	__antithesis_instrumentation__.Notify(642233)
	options := pebble.IterOptions{
		LowerBound:                key,
		UpperBound:                roachpb.BytesNext(key),
		OnlyReadGuaranteedDurable: onlyReadGuaranteedDurable,
	}
	iter := p.parent.db.NewIter(&options)
	defer func() {
		__antithesis_instrumentation__.Notify(642240)

		_ = iter.Close()
	}()
	__antithesis_instrumentation__.Notify(642234)
	valid := iter.SeekGE(key)
	if !valid {
		__antithesis_instrumentation__.Notify(642241)
		return nil, iter.Error()
	} else {
		__antithesis_instrumentation__.Notify(642242)
	}
	__antithesis_instrumentation__.Notify(642235)
	return iter.Value(), nil
}

func (p *pebbleReadOnly) MVCCGetProto(
	key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(642243)
	return pebbleGetProto(p, key, msg)
}

func (p *pebbleReadOnly) MVCCIterate(
	start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(642244)
	if p.closed {
		__antithesis_instrumentation__.Notify(642247)
		panic("using a closed pebbleReadOnly")
	} else {
		__antithesis_instrumentation__.Notify(642248)
	}
	__antithesis_instrumentation__.Notify(642245)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642249)
		r := wrapReader(p)

		err := iterateOnReader(r, start, end, iterKind, f)
		r.Free()
		return err
	} else {
		__antithesis_instrumentation__.Notify(642250)
	}
	__antithesis_instrumentation__.Notify(642246)
	return iterateOnReader(p, start, end, iterKind, f)
}

func (p *pebbleReadOnly) NewMVCCIterator(iterKind MVCCIterKind, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(642251)
	if p.closed {
		__antithesis_instrumentation__.Notify(642259)
		panic("using a closed pebbleReadOnly")
	} else {
		__antithesis_instrumentation__.Notify(642260)
	}
	__antithesis_instrumentation__.Notify(642252)

	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642261)
		r := wrapReader(p)

		iter := r.NewMVCCIterator(iterKind, opts)
		r.Free()
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(642263)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(642264)
		}
		__antithesis_instrumentation__.Notify(642262)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(642265)
	}
	__antithesis_instrumentation__.Notify(642253)

	if !opts.MinTimestampHint.IsEmpty() {
		__antithesis_instrumentation__.Notify(642266)

		iter := MVCCIterator(newPebbleIterator(p.parent.db, nil, opts, p.durability))
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(642268)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(642269)
		}
		__antithesis_instrumentation__.Notify(642267)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(642270)
	}
	__antithesis_instrumentation__.Notify(642254)

	iter := &p.normalIter
	if opts.Prefix {
		__antithesis_instrumentation__.Notify(642271)
		iter = &p.prefixIter
	} else {
		__antithesis_instrumentation__.Notify(642272)
	}
	__antithesis_instrumentation__.Notify(642255)
	if iter.inuse {
		__antithesis_instrumentation__.Notify(642273)
		panic("iterator already in use")
	} else {
		__antithesis_instrumentation__.Notify(642274)
	}
	__antithesis_instrumentation__.Notify(642256)

	checkOptionsForIterReuse(opts)

	if iter.iter != nil {
		__antithesis_instrumentation__.Notify(642275)
		iter.setBounds(opts.LowerBound, opts.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(642276)
		iter.init(p.parent.db, p.iter, opts, p.durability)
		if p.iter == nil {
			__antithesis_instrumentation__.Notify(642278)

			p.iter = iter.iter
		} else {
			__antithesis_instrumentation__.Notify(642279)
		}
		__antithesis_instrumentation__.Notify(642277)
		iter.reusable = true
	}
	__antithesis_instrumentation__.Notify(642257)

	iter.inuse = true
	var rv MVCCIterator = iter
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(642280)
		rv = wrapInUnsafeIter(rv)
	} else {
		__antithesis_instrumentation__.Notify(642281)
	}
	__antithesis_instrumentation__.Notify(642258)
	return rv
}

func (p *pebbleReadOnly) NewEngineIterator(opts IterOptions) EngineIterator {
	__antithesis_instrumentation__.Notify(642282)
	if p.closed {
		__antithesis_instrumentation__.Notify(642287)
		panic("using a closed pebbleReadOnly")
	} else {
		__antithesis_instrumentation__.Notify(642288)
	}
	__antithesis_instrumentation__.Notify(642283)

	iter := &p.normalEngineIter
	if opts.Prefix {
		__antithesis_instrumentation__.Notify(642289)
		iter = &p.prefixEngineIter
	} else {
		__antithesis_instrumentation__.Notify(642290)
	}
	__antithesis_instrumentation__.Notify(642284)
	if iter.inuse {
		__antithesis_instrumentation__.Notify(642291)
		panic("iterator already in use")
	} else {
		__antithesis_instrumentation__.Notify(642292)
	}
	__antithesis_instrumentation__.Notify(642285)

	checkOptionsForIterReuse(opts)

	if iter.iter != nil {
		__antithesis_instrumentation__.Notify(642293)
		iter.setBounds(opts.LowerBound, opts.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(642294)
		iter.init(p.parent.db, p.iter, opts, p.durability)
		if p.iter == nil {
			__antithesis_instrumentation__.Notify(642296)

			p.iter = iter.iter
		} else {
			__antithesis_instrumentation__.Notify(642297)
		}
		__antithesis_instrumentation__.Notify(642295)
		iter.reusable = true
	}
	__antithesis_instrumentation__.Notify(642286)

	iter.inuse = true
	return iter
}

func checkOptionsForIterReuse(opts IterOptions) {
	__antithesis_instrumentation__.Notify(642298)
	if !opts.MinTimestampHint.IsEmpty() || func() bool {
		__antithesis_instrumentation__.Notify(642300)
		return !opts.MaxTimestampHint.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(642301)
		panic("iterator with timestamp hints cannot be reused")
	} else {
		__antithesis_instrumentation__.Notify(642302)
	}
	__antithesis_instrumentation__.Notify(642299)
	if !opts.Prefix && func() bool {
		__antithesis_instrumentation__.Notify(642303)
		return len(opts.UpperBound) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(642304)
		return len(opts.LowerBound) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(642305)
		panic("iterator must set prefix or upper bound or lower bound")
	} else {
		__antithesis_instrumentation__.Notify(642306)
	}
}

func (p *pebbleReadOnly) ConsistentIterators() bool {
	__antithesis_instrumentation__.Notify(642307)
	return true
}

func (p *pebbleReadOnly) PinEngineStateForIterators() error {
	__antithesis_instrumentation__.Notify(642308)
	if p.iter == nil {
		__antithesis_instrumentation__.Notify(642310)
		o := (*pebble.IterOptions)(nil)
		if p.durability == GuaranteedDurability {
			__antithesis_instrumentation__.Notify(642312)
			o = &pebble.IterOptions{OnlyReadGuaranteedDurable: true}
		} else {
			__antithesis_instrumentation__.Notify(642313)
		}
		__antithesis_instrumentation__.Notify(642311)
		p.iter = p.parent.db.NewIter(o)
	} else {
		__antithesis_instrumentation__.Notify(642314)
	}
	__antithesis_instrumentation__.Notify(642309)
	return nil
}

func (p *pebbleReadOnly) ApplyBatchRepr(repr []byte, sync bool) error {
	__antithesis_instrumentation__.Notify(642315)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearMVCC(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(642316)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearUnversioned(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642317)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearIntent(
	key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(642318)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642319)
	panic("not implemented")
}

func (p *pebbleReadOnly) SingleClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642320)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearRawRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642321)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearMVCCRangeAndIntents(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642322)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearMVCCRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(642323)
	panic("not implemented")
}

func (p *pebbleReadOnly) ClearIterRange(iter MVCCIterator, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642324)
	panic("not implemented")
}

func (p *pebbleReadOnly) Merge(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642325)
	panic("not implemented")
}

func (p *pebbleReadOnly) PutMVCC(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642326)
	panic("not implemented")
}

func (p *pebbleReadOnly) PutUnversioned(key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(642327)
	panic("not implemented")
}

func (p *pebbleReadOnly) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(642328)
	panic("not implemented")
}

func (p *pebbleReadOnly) PutEngineKey(key EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642329)
	panic("not implemented")
}

func (p *pebbleReadOnly) LogData(data []byte) error {
	__antithesis_instrumentation__.Notify(642330)
	panic("not implemented")
}

func (p *pebbleReadOnly) LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(642331)
	panic("not implemented")
}

type pebbleSnapshot struct {
	snapshot *pebble.Snapshot
	settings *cluster.Settings
	closed   bool
}

var _ Reader = &pebbleSnapshot{}

func (p *pebbleSnapshot) Close() {
	__antithesis_instrumentation__.Notify(642332)
	_ = p.snapshot.Close()
	p.closed = true
}

func (p *pebbleSnapshot) Closed() bool {
	__antithesis_instrumentation__.Notify(642333)
	return p.closed
}

func (p *pebbleSnapshot) ExportMVCCToSst(
	ctx context.Context, exportOptions ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, roachpb.Key, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(642334)
	r := wrapReader(p)

	summary, k, err := pebbleExportToSst(ctx, p.settings, r, exportOptions, dest)
	r.Free()
	return summary, k.Key, k.Timestamp, err
}

func (p *pebbleSnapshot) MVCCGet(key MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642335)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642337)
		return nil, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642338)
	}
	__antithesis_instrumentation__.Notify(642336)
	r := wrapReader(p)

	v, err := r.MVCCGet(key)
	r.Free()
	return v, err
}

func (p *pebbleSnapshot) rawMVCCGet(key []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642339)
	ret, closer, err := p.snapshot.Get(key)
	if closer != nil {
		__antithesis_instrumentation__.Notify(642342)
		retCopy := make([]byte, len(ret))
		copy(retCopy, ret)
		ret = retCopy
		closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(642343)
	}
	__antithesis_instrumentation__.Notify(642340)
	if errors.Is(err, pebble.ErrNotFound) || func() bool {
		__antithesis_instrumentation__.Notify(642344)
		return len(ret) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(642345)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(642346)
	}
	__antithesis_instrumentation__.Notify(642341)
	return ret, err
}

func (p *pebbleSnapshot) MVCCGetProto(
	key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(642347)
	return pebbleGetProto(p, key, msg)
}

func (p *pebbleSnapshot) MVCCIterate(
	start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(642348)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642350)
		r := wrapReader(p)

		err := iterateOnReader(r, start, end, iterKind, f)
		r.Free()
		return err
	} else {
		__antithesis_instrumentation__.Notify(642351)
	}
	__antithesis_instrumentation__.Notify(642349)
	return iterateOnReader(p, start, end, iterKind, f)
}

func (p *pebbleSnapshot) NewMVCCIterator(iterKind MVCCIterKind, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(642352)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642355)
		r := wrapReader(p)

		iter := r.NewMVCCIterator(iterKind, opts)
		r.Free()
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(642357)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(642358)
		}
		__antithesis_instrumentation__.Notify(642356)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(642359)
	}
	__antithesis_instrumentation__.Notify(642353)
	iter := MVCCIterator(newPebbleIterator(p.snapshot, nil, opts, StandardDurability))
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(642360)
		iter = wrapInUnsafeIter(iter)
	} else {
		__antithesis_instrumentation__.Notify(642361)
	}
	__antithesis_instrumentation__.Notify(642354)
	return iter
}

func (p pebbleSnapshot) NewEngineIterator(opts IterOptions) EngineIterator {
	__antithesis_instrumentation__.Notify(642362)
	return newPebbleIterator(p.snapshot, nil, opts, StandardDurability)
}

func (p pebbleSnapshot) ConsistentIterators() bool {
	__antithesis_instrumentation__.Notify(642363)
	return true
}

func (p *pebbleSnapshot) PinEngineStateForIterators() error {
	__antithesis_instrumentation__.Notify(642364)

	return nil
}

func pebbleGetProto(
	reader Reader, key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(642365)
	val, err := reader.MVCCGet(key)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(642368)
		return val == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(642369)
		return false, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(642370)
	}
	__antithesis_instrumentation__.Notify(642366)
	keyBytes = int64(key.Len())
	valBytes = int64(len(val))
	if msg != nil {
		__antithesis_instrumentation__.Notify(642371)
		err = protoutil.Unmarshal(val, msg)
	} else {
		__antithesis_instrumentation__.Notify(642372)
	}
	__antithesis_instrumentation__.Notify(642367)
	return true, keyBytes, valBytes, err
}

type ExceedMaxSizeError struct {
	reached int64
	maxSize uint64
}

var _ error = &ExceedMaxSizeError{}

func (e *ExceedMaxSizeError) Error() string {
	__antithesis_instrumentation__.Notify(642373)
	return fmt.Sprintf("export size (%d bytes) exceeds max size (%d bytes)", e.reached, e.maxSize)
}

func pebbleExportToSst(
	ctx context.Context, cs *cluster.Settings, reader Reader, options ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, MVCCKey, error) {
	__antithesis_instrumentation__.Notify(642374)
	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, "pebbleExportToSst")
	defer span.Finish()
	sstWriter := MakeBackupSSTWriter(ctx, cs, dest)
	defer sstWriter.Close()

	var rows RowCounter
	iter := NewMVCCIncrementalIterator(
		reader,
		MVCCIncrementalIterOptions{
			EndKey:                              options.EndKey,
			EnableTimeBoundIteratorOptimization: options.UseTBI,
			StartTime:                           options.StartTS,
			EndTime:                             options.EndTS,
			IntentPolicy:                        MVCCIncrementalIterIntentPolicyAggregate,
		})
	defer iter.Close()
	var curKey roachpb.Key
	var resumeKey roachpb.Key
	var resumeTS hlc.Timestamp
	paginated := options.TargetSize > 0
	trackKeyBoundary := paginated || func() bool {
		__antithesis_instrumentation__.Notify(642379)
		return options.ResourceLimiter != nil == true
	}() == true
	firstIteration := true
	for iter.SeekGE(options.StartKey); ; {
		__antithesis_instrumentation__.Notify(642380)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(642388)
			return roachpb.BulkOpSummary{}, MVCCKey{}, err
		} else {
			__antithesis_instrumentation__.Notify(642389)
		}
		__antithesis_instrumentation__.Notify(642381)
		if !ok {
			__antithesis_instrumentation__.Notify(642390)
			break
		} else {
			__antithesis_instrumentation__.Notify(642391)
		}
		__antithesis_instrumentation__.Notify(642382)
		unsafeKey := iter.UnsafeKey()
		if unsafeKey.Key.Compare(options.EndKey) >= 0 {
			__antithesis_instrumentation__.Notify(642392)
			break
		} else {
			__antithesis_instrumentation__.Notify(642393)
		}
		__antithesis_instrumentation__.Notify(642383)

		if iter.NumCollectedIntents() > 0 {
			__antithesis_instrumentation__.Notify(642394)
			break
		} else {
			__antithesis_instrumentation__.Notify(642395)
		}
		__antithesis_instrumentation__.Notify(642384)

		unsafeValue := iter.UnsafeValue()
		isNewKey := !options.ExportAllRevisions || func() bool {
			__antithesis_instrumentation__.Notify(642396)
			return !unsafeKey.Key.Equal(curKey) == true
		}() == true
		if trackKeyBoundary && func() bool {
			__antithesis_instrumentation__.Notify(642397)
			return options.ExportAllRevisions == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(642398)
			return isNewKey == true
		}() == true {
			__antithesis_instrumentation__.Notify(642399)
			curKey = append(curKey[:0], unsafeKey.Key...)
		} else {
			__antithesis_instrumentation__.Notify(642400)
		}
		__antithesis_instrumentation__.Notify(642385)

		if options.ResourceLimiter != nil {
			__antithesis_instrumentation__.Notify(642401)

			if firstIteration {
				__antithesis_instrumentation__.Notify(642402)
				firstIteration = false
			} else {
				__antithesis_instrumentation__.Notify(642403)

				limit := options.ResourceLimiter.IsExhausted()

				if limit >= ResourceLimitReachedSoft && func() bool {
					__antithesis_instrumentation__.Notify(642404)
					return isNewKey == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(642405)
					return (limit == ResourceLimitReachedHard && func() bool {
						__antithesis_instrumentation__.Notify(642406)
						return options.StopMidKey == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(642407)

					resumeKey = append(make(roachpb.Key, 0, len(unsafeKey.Key)), unsafeKey.Key...)
					if !isNewKey {
						__antithesis_instrumentation__.Notify(642409)
						resumeTS = unsafeKey.Timestamp
					} else {
						__antithesis_instrumentation__.Notify(642410)
					}
					__antithesis_instrumentation__.Notify(642408)
					break
				} else {
					__antithesis_instrumentation__.Notify(642411)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(642412)
		}
		__antithesis_instrumentation__.Notify(642386)

		skipTombstones := !options.ExportAllRevisions && func() bool {
			__antithesis_instrumentation__.Notify(642413)
			return options.StartTS.IsEmpty() == true
		}() == true
		if len(unsafeValue) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(642414)
			return !skipTombstones == true
		}() == true {
			__antithesis_instrumentation__.Notify(642415)
			if err := rows.Count(unsafeKey.Key); err != nil {
				__antithesis_instrumentation__.Notify(642420)
				return roachpb.BulkOpSummary{}, MVCCKey{}, errors.Wrapf(err, "decoding %s", unsafeKey)
			} else {
				__antithesis_instrumentation__.Notify(642421)
			}
			__antithesis_instrumentation__.Notify(642416)
			curSize := rows.BulkOpSummary.DataSize
			reachedTargetSize := curSize > 0 && func() bool {
				__antithesis_instrumentation__.Notify(642422)
				return uint64(curSize) >= options.TargetSize == true
			}() == true
			newSize := curSize + int64(len(unsafeKey.Key)+len(unsafeValue))
			reachedMaxSize := options.MaxSize > 0 && func() bool {
				__antithesis_instrumentation__.Notify(642423)
				return newSize > int64(options.MaxSize) == true
			}() == true

			if paginated && func() bool {
				__antithesis_instrumentation__.Notify(642424)
				return (isNewKey && func() bool {
					__antithesis_instrumentation__.Notify(642425)
					return reachedTargetSize == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(642426)
					return (options.StopMidKey && func() bool {
						__antithesis_instrumentation__.Notify(642427)
						return reachedMaxSize == true
					}() == true) == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(642428)

				resumeKey = append(make(roachpb.Key, 0, len(unsafeKey.Key)), unsafeKey.Key...)
				if options.StopMidKey && func() bool {
					__antithesis_instrumentation__.Notify(642430)
					return !isNewKey == true
				}() == true {
					__antithesis_instrumentation__.Notify(642431)
					resumeTS = unsafeKey.Timestamp
				} else {
					__antithesis_instrumentation__.Notify(642432)
				}
				__antithesis_instrumentation__.Notify(642429)
				break
			} else {
				__antithesis_instrumentation__.Notify(642433)
			}
			__antithesis_instrumentation__.Notify(642417)
			if reachedMaxSize {
				__antithesis_instrumentation__.Notify(642434)
				return roachpb.BulkOpSummary{}, MVCCKey{}, &ExceedMaxSizeError{reached: newSize, maxSize: options.MaxSize}
			} else {
				__antithesis_instrumentation__.Notify(642435)
			}
			__antithesis_instrumentation__.Notify(642418)
			if unsafeKey.Timestamp.IsEmpty() {
				__antithesis_instrumentation__.Notify(642436)

				if err := sstWriter.PutUnversioned(unsafeKey.Key, unsafeValue); err != nil {
					__antithesis_instrumentation__.Notify(642437)
					return roachpb.BulkOpSummary{}, MVCCKey{}, errors.Wrapf(err, "adding key %s", unsafeKey)
				} else {
					__antithesis_instrumentation__.Notify(642438)
				}
			} else {
				__antithesis_instrumentation__.Notify(642439)
				if err := sstWriter.PutMVCC(unsafeKey, unsafeValue); err != nil {
					__antithesis_instrumentation__.Notify(642440)
					return roachpb.BulkOpSummary{}, MVCCKey{}, errors.Wrapf(err, "adding key %s", unsafeKey)
				} else {
					__antithesis_instrumentation__.Notify(642441)
				}
			}
			__antithesis_instrumentation__.Notify(642419)
			rows.BulkOpSummary.DataSize = newSize
		} else {
			__antithesis_instrumentation__.Notify(642442)
		}
		__antithesis_instrumentation__.Notify(642387)

		if options.ExportAllRevisions {
			__antithesis_instrumentation__.Notify(642443)
			iter.Next()
		} else {
			__antithesis_instrumentation__.Notify(642444)
			iter.NextKey()
		}
	}
	__antithesis_instrumentation__.Notify(642375)

	if iter.NumCollectedIntents() > 0 {
		__antithesis_instrumentation__.Notify(642445)
		for uint64(iter.NumCollectedIntents()) < options.MaxIntents {
			__antithesis_instrumentation__.Notify(642447)
			iter.NextKey()

			ok, _ := iter.Valid()
			if !ok {
				__antithesis_instrumentation__.Notify(642448)
				break
			} else {
				__antithesis_instrumentation__.Notify(642449)
			}
		}
		__antithesis_instrumentation__.Notify(642446)
		err := iter.TryGetIntentError()
		return roachpb.BulkOpSummary{}, MVCCKey{}, err
	} else {
		__antithesis_instrumentation__.Notify(642450)
	}
	__antithesis_instrumentation__.Notify(642376)

	if rows.BulkOpSummary.DataSize == 0 {
		__antithesis_instrumentation__.Notify(642451)

		return roachpb.BulkOpSummary{}, MVCCKey{}, nil
	} else {
		__antithesis_instrumentation__.Notify(642452)
	}
	__antithesis_instrumentation__.Notify(642377)

	if err := sstWriter.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(642453)
		return roachpb.BulkOpSummary{}, MVCCKey{}, err
	} else {
		__antithesis_instrumentation__.Notify(642454)
	}
	__antithesis_instrumentation__.Notify(642378)

	return rows.BulkOpSummary, MVCCKey{Key: resumeKey, Timestamp: resumeTS}, nil
}
