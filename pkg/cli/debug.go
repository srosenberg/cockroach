package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	gohex "encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/syncbench"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/gc"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/flagutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble/tool"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/ttycolor"
	humanize "github.com/dustin/go-humanize"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

var debugKeysCmd = &cobra.Command{
	Use:   "keys <directory>",
	Short: "dump all the keys in a store",
	Long: `
Pretty-prints all keys in a store.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugKeys),
}

var debugBallastCmd = &cobra.Command{
	Use:   "ballast <file>",
	Short: "create a ballast file",
	Long: `
Create a ballast file to fill the store directory up to a given amount
`,
	Args: cobra.ExactArgs(1),
	RunE: runDebugBallast,
}

var PopulateStorageConfigHook func(*base.StorageConfig) error

func parsePositiveInt(arg string) (int64, error) {
	__antithesis_instrumentation__.Notify(30422)
	i, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(30425)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(30426)
	}
	__antithesis_instrumentation__.Notify(30423)
	if i < 1 {
		__antithesis_instrumentation__.Notify(30427)
		return 0, fmt.Errorf("illegal val: %d < 1", i)
	} else {
		__antithesis_instrumentation__.Notify(30428)
	}
	__antithesis_instrumentation__.Notify(30424)
	return i, nil
}

func parseRangeID(arg string) (roachpb.RangeID, error) {
	__antithesis_instrumentation__.Notify(30429)
	rangeIDInt, err := parsePositiveInt(arg)
	if err != nil {
		__antithesis_instrumentation__.Notify(30431)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(30432)
	}
	__antithesis_instrumentation__.Notify(30430)
	return roachpb.RangeID(rangeIDInt), nil
}

func parsePositiveDuration(arg string) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(30433)
	duration, err := time.ParseDuration(arg)
	if err != nil {
		__antithesis_instrumentation__.Notify(30436)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(30437)
	}
	__antithesis_instrumentation__.Notify(30434)
	if duration <= 0 {
		__antithesis_instrumentation__.Notify(30438)
		return 0, fmt.Errorf("illegal val: %v <= 0", duration)
	} else {
		__antithesis_instrumentation__.Notify(30439)
	}
	__antithesis_instrumentation__.Notify(30435)
	return duration, nil
}

type OpenEngineOptions struct {
	ReadOnly                    bool
	MustExist                   bool
	DisableAutomaticCompactions bool
}

func (opts OpenEngineOptions) configOptions() []storage.ConfigOption {
	__antithesis_instrumentation__.Notify(30440)
	var cfgOpts []storage.ConfigOption
	if opts.ReadOnly {
		__antithesis_instrumentation__.Notify(30444)
		cfgOpts = append(cfgOpts, storage.ReadOnly)
	} else {
		__antithesis_instrumentation__.Notify(30445)
	}
	__antithesis_instrumentation__.Notify(30441)
	if opts.MustExist {
		__antithesis_instrumentation__.Notify(30446)
		cfgOpts = append(cfgOpts, storage.MustExist)
	} else {
		__antithesis_instrumentation__.Notify(30447)
	}
	__antithesis_instrumentation__.Notify(30442)
	if opts.DisableAutomaticCompactions {
		__antithesis_instrumentation__.Notify(30448)
		cfgOpts = append(cfgOpts, storage.DisableAutomaticCompactions)
	} else {
		__antithesis_instrumentation__.Notify(30449)
	}
	__antithesis_instrumentation__.Notify(30443)
	return cfgOpts
}

func OpenExistingStore(
	dir string, stopper *stop.Stopper, readOnly, disableAutomaticCompactions bool,
) (storage.Engine, error) {
	__antithesis_instrumentation__.Notify(30450)
	return OpenEngine(dir, stopper, OpenEngineOptions{
		ReadOnly: readOnly, MustExist: true, DisableAutomaticCompactions: disableAutomaticCompactions,
	})
}

func OpenEngine(dir string, stopper *stop.Stopper, opts OpenEngineOptions) (storage.Engine, error) {
	__antithesis_instrumentation__.Notify(30451)
	maxOpenFiles, err := server.SetOpenFileLimitForOneStore()
	if err != nil {
		__antithesis_instrumentation__.Notify(30454)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30455)
	}
	__antithesis_instrumentation__.Notify(30452)
	db, err := storage.Open(context.Background(),
		storage.Filesystem(dir),
		storage.MaxOpenFiles(int(maxOpenFiles)),
		storage.CacheSize(server.DefaultCacheSize),
		storage.Settings(serverCfg.Settings),
		storage.Hook(PopulateStorageConfigHook),
		storage.CombineOptions(opts.configOptions()...))
	if err != nil {
		__antithesis_instrumentation__.Notify(30456)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30457)
	}
	__antithesis_instrumentation__.Notify(30453)

	stopper.AddCloser(db)
	return db, nil
}

func printKey(kv storage.MVCCKeyValue) (bool, error) {
	__antithesis_instrumentation__.Notify(30458)
	fmt.Printf("%s %s: ", kv.Key.Timestamp, kv.Key.Key)
	if debugCtx.sizes {
		__antithesis_instrumentation__.Notify(30460)
		fmt.Printf(" %d %d", len(kv.Key.Key), len(kv.Value))
	} else {
		__antithesis_instrumentation__.Notify(30461)
	}
	__antithesis_instrumentation__.Notify(30459)
	fmt.Printf("\n")
	return false, nil
}

func transactionPredicate(kv storage.MVCCKeyValue) bool {
	__antithesis_instrumentation__.Notify(30462)
	if kv.Key.IsValue() {
		__antithesis_instrumentation__.Notify(30465)
		return false
	} else {
		__antithesis_instrumentation__.Notify(30466)
	}
	__antithesis_instrumentation__.Notify(30463)
	_, suffix, _, err := keys.DecodeRangeKey(kv.Key.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(30467)
		return false
	} else {
		__antithesis_instrumentation__.Notify(30468)
	}
	__antithesis_instrumentation__.Notify(30464)
	return keys.LocalTransactionSuffix.Equal(suffix)
}

func intentPredicate(kv storage.MVCCKeyValue) bool {
	__antithesis_instrumentation__.Notify(30469)
	if kv.Key.IsValue() {
		__antithesis_instrumentation__.Notify(30472)
		return false
	} else {
		__antithesis_instrumentation__.Notify(30473)
	}
	__antithesis_instrumentation__.Notify(30470)
	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(kv.Value, &meta); err != nil {
		__antithesis_instrumentation__.Notify(30474)
		return false
	} else {
		__antithesis_instrumentation__.Notify(30475)
	}
	__antithesis_instrumentation__.Notify(30471)
	return meta.Txn != nil
}

var keyTypeParams = map[keyTypeFilter]struct {
	predicate      func(kv storage.MVCCKeyValue) bool
	minKey, maxKey storage.MVCCKey
}{
	showAll: {
		predicate: func(kv storage.MVCCKeyValue) bool { __antithesis_instrumentation__.Notify(30476); return true },
		minKey:    storage.NilKey,
		maxKey:    storage.MVCCKeyMax,
	},
	showTxns: {
		predicate: transactionPredicate,
		minKey:    storage.NilKey,
		maxKey:    storage.MVCCKey{Key: keys.LocalMax},
	},
	showValues: {
		predicate: func(kv storage.MVCCKeyValue) bool {
			__antithesis_instrumentation__.Notify(30477)
			return kv.Key.IsValue()
		},
		minKey: storage.NilKey,
		maxKey: storage.MVCCKeyMax,
	},
	showIntents: {
		predicate: intentPredicate,
		minKey:    storage.NilKey,
		maxKey:    storage.MVCCKeyMax,
	},
}

func runDebugKeys(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30478)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30488)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30489)
	}
	__antithesis_instrumentation__.Notify(30479)

	if debugCtx.decodeAsTableDesc != "" {
		__antithesis_instrumentation__.Notify(30490)
		bytes, err := base64.StdEncoding.DecodeString(debugCtx.decodeAsTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(30495)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30496)
		}
		__antithesis_instrumentation__.Notify(30491)
		var desc descpb.Descriptor
		if err := protoutil.Unmarshal(bytes, &desc); err != nil {
			__antithesis_instrumentation__.Notify(30497)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30498)
		}
		__antithesis_instrumentation__.Notify(30492)
		b := descbuilder.NewBuilder(&desc)
		if b == nil || func() bool {
			__antithesis_instrumentation__.Notify(30499)
			return b.DescriptorType() != catalog.Table == true
		}() == true {
			__antithesis_instrumentation__.Notify(30500)
			return errors.Newf("expected a table descriptor")
		} else {
			__antithesis_instrumentation__.Notify(30501)
		}
		__antithesis_instrumentation__.Notify(30493)
		table := b.BuildImmutable().(catalog.TableDescriptor)

		fn := func(kv storage.MVCCKeyValue) (string, error) {
			__antithesis_instrumentation__.Notify(30502)
			var v roachpb.Value
			v.RawBytes = kv.Value
			_, names, values, err := row.DecodeRowInfo(context.Background(), table, kv.Key.Key, &v, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(30505)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(30506)
			}
			__antithesis_instrumentation__.Notify(30503)
			pairs := make([]string, len(names))
			for i := range pairs {
				__antithesis_instrumentation__.Notify(30507)
				pairs[i] = fmt.Sprintf("%s=%s", names[i], values[i])
			}
			__antithesis_instrumentation__.Notify(30504)
			return strings.Join(pairs, ", "), nil
		}
		__antithesis_instrumentation__.Notify(30494)
		kvserver.DebugSprintMVCCKeyValueDecoders = append(kvserver.DebugSprintMVCCKeyValueDecoders, fn)
	} else {
		__antithesis_instrumentation__.Notify(30508)
	}
	__antithesis_instrumentation__.Notify(30480)
	printer := printKey
	if debugCtx.values {
		__antithesis_instrumentation__.Notify(30509)
		printer = func(kv storage.MVCCKeyValue) (bool, error) {
			__antithesis_instrumentation__.Notify(30510)
			kvserver.PrintMVCCKeyValue(kv)
			return false, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(30511)
	}
	__antithesis_instrumentation__.Notify(30481)

	keyTypeOptions := keyTypeParams[debugCtx.keyTypes]
	if debugCtx.startKey.Equal(storage.NilKey) {
		__antithesis_instrumentation__.Notify(30512)
		debugCtx.startKey = keyTypeOptions.minKey
	} else {
		__antithesis_instrumentation__.Notify(30513)
	}
	__antithesis_instrumentation__.Notify(30482)
	if debugCtx.endKey.Equal(storage.NilKey) {
		__antithesis_instrumentation__.Notify(30514)
		debugCtx.endKey = keyTypeOptions.maxKey
	} else {
		__antithesis_instrumentation__.Notify(30515)
	}
	__antithesis_instrumentation__.Notify(30483)

	results := 0
	iterFunc := func(kv storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(30516)
		if !keyTypeOptions.predicate(kv) {
			__antithesis_instrumentation__.Notify(30521)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30522)
		}
		__antithesis_instrumentation__.Notify(30517)
		done, err := printer(kv)
		if err != nil {
			__antithesis_instrumentation__.Notify(30523)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30524)
		}
		__antithesis_instrumentation__.Notify(30518)
		if done {
			__antithesis_instrumentation__.Notify(30525)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(30526)
		}
		__antithesis_instrumentation__.Notify(30519)
		results++
		if results == debugCtx.maxResults {
			__antithesis_instrumentation__.Notify(30527)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(30528)
		}
		__antithesis_instrumentation__.Notify(30520)
		return nil
	}
	__antithesis_instrumentation__.Notify(30484)
	endKey := debugCtx.endKey.Key
	splitScan := false

	if (len(debugCtx.startKey.Key) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(30529)
		return keys.IsLocal(debugCtx.startKey.Key) == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(30530)
		return !(keys.IsLocal(endKey) || func() bool {
			__antithesis_instrumentation__.Notify(30531)
			return bytes.Equal(endKey, keys.LocalMax) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(30532)
		splitScan = true
		endKey = keys.LocalMax
	} else {
		__antithesis_instrumentation__.Notify(30533)
	}
	__antithesis_instrumentation__.Notify(30485)
	if err := db.MVCCIterate(
		debugCtx.startKey.Key, endKey, storage.MVCCKeyAndIntentsIterKind, iterFunc); err != nil {
		__antithesis_instrumentation__.Notify(30534)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30535)
	}
	__antithesis_instrumentation__.Notify(30486)
	if splitScan {
		__antithesis_instrumentation__.Notify(30536)
		if err := db.MVCCIterate(keys.LocalMax, debugCtx.endKey.Key, storage.MVCCKeyAndIntentsIterKind,
			iterFunc); err != nil {
			__antithesis_instrumentation__.Notify(30537)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30538)
		}
	} else {
		__antithesis_instrumentation__.Notify(30539)
	}
	__antithesis_instrumentation__.Notify(30487)
	return nil
}

func runDebugBallast(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30540)
	ballastFile := args[0]
	dataDirectory := filepath.Dir(ballastFile)

	du, err := vfs.Default.GetDiskUsage(dataDirectory)
	if err != nil {
		__antithesis_instrumentation__.Notify(30549)
		return errors.Wrapf(err, "failed to stat filesystem %s", dataDirectory)
	} else {
		__antithesis_instrumentation__.Notify(30550)
	}
	__antithesis_instrumentation__.Notify(30541)

	usedBytes := du.TotalBytes - du.AvailBytes

	var targetUsage uint64
	p := debugCtx.ballastSize.Percent
	if math.Abs(p) > 100 {
		__antithesis_instrumentation__.Notify(30551)
		return errors.Errorf("absolute percentage value %f greater than 100", p)
	} else {
		__antithesis_instrumentation__.Notify(30552)
	}
	__antithesis_instrumentation__.Notify(30542)
	b := debugCtx.ballastSize.InBytes
	if p != 0 && func() bool {
		__antithesis_instrumentation__.Notify(30553)
		return b != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(30554)
		return errors.New("expected exactly one of percentage or bytes non-zero, found both")
	} else {
		__antithesis_instrumentation__.Notify(30555)
	}
	__antithesis_instrumentation__.Notify(30543)
	switch {
	case p > 0:
		__antithesis_instrumentation__.Notify(30556)
		fillRatio := p / float64(100)
		targetUsage = usedBytes + uint64((fillRatio)*float64(du.TotalBytes))
	case p < 0:
		__antithesis_instrumentation__.Notify(30557)

		fillRatio := 1.0 + (p / float64(100))
		targetUsage = uint64((fillRatio) * float64(du.TotalBytes))
	case b > 0:
		__antithesis_instrumentation__.Notify(30558)
		targetUsage = usedBytes + uint64(b)
	case b < 0:
		__antithesis_instrumentation__.Notify(30559)

		targetUsage = du.TotalBytes - uint64(-b)
	default:
		__antithesis_instrumentation__.Notify(30560)
		return errors.New("expected exactly one of percentage or bytes non-zero, found none")
	}
	__antithesis_instrumentation__.Notify(30544)
	if usedBytes > targetUsage {
		__antithesis_instrumentation__.Notify(30561)
		return errors.Errorf(
			"Used space %s already more than needed to be filled %s\n",
			humanize.IBytes(usedBytes),
			humanize.IBytes(targetUsage),
		)
	} else {
		__antithesis_instrumentation__.Notify(30562)
	}
	__antithesis_instrumentation__.Notify(30545)
	if usedBytes == targetUsage {
		__antithesis_instrumentation__.Notify(30563)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(30564)
	}
	__antithesis_instrumentation__.Notify(30546)
	ballastSize := targetUsage - usedBytes

	if _, err := os.Stat(ballastFile); err == nil {
		__antithesis_instrumentation__.Notify(30565)
		return os.ErrExist
	} else {
		__antithesis_instrumentation__.Notify(30566)
		if !oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(30567)
			return errors.Wrap(err, "stating ballast file")
		} else {
			__antithesis_instrumentation__.Notify(30568)
		}
	}
	__antithesis_instrumentation__.Notify(30547)

	if err := sysutil.ResizeLargeFile(ballastFile, int64(ballastSize)); err != nil {
		__antithesis_instrumentation__.Notify(30569)
		return errors.Wrap(err, "error allocating ballast file")
	} else {
		__antithesis_instrumentation__.Notify(30570)
	}
	__antithesis_instrumentation__.Notify(30548)
	return nil
}

var debugRangeDataCmd = &cobra.Command{
	Use:   "range-data <directory> <range id>",
	Short: "dump all the data in a range",
	Long: `
Pretty-prints all keys and values in a range. By default, includes unreplicated
state like the raft HardState. With --replicated, only includes data covered by
 the consistency checker.
`,
	Args: cobra.ExactArgs(2),
	RunE: clierrorplus.MaybeDecorateError(runDebugRangeData),
}

func runDebugRangeData(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30571)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30577)
	}
	__antithesis_instrumentation__.Notify(30572)

	rangeID, err := parseRangeID(args[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(30578)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30579)
	}
	__antithesis_instrumentation__.Notify(30573)

	desc, err := loadRangeDescriptor(db, rangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(30580)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30581)
	}
	__antithesis_instrumentation__.Notify(30574)

	iter := rditer.NewReplicaEngineDataIterator(&desc, db, debugCtx.replicated)
	defer iter.Close()
	results := 0
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(30582)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(30584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30585)
			if !ok {
				__antithesis_instrumentation__.Notify(30586)
				break
			} else {
				__antithesis_instrumentation__.Notify(30587)
			}
		}
		__antithesis_instrumentation__.Notify(30583)
		kvserver.PrintEngineKeyValue(iter.UnsafeKey(), iter.UnsafeValue())
		results++
		if results == debugCtx.maxResults {
			__antithesis_instrumentation__.Notify(30588)
			break
		} else {
			__antithesis_instrumentation__.Notify(30589)
		}
	}
	__antithesis_instrumentation__.Notify(30575)
	return nil
}

var debugRangeDescriptorsCmd = &cobra.Command{
	Use:   "range-descriptors <directory>",
	Short: "print all range descriptors in a store",
	Long: `
Prints all range descriptors in a store with a history of changes.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugRangeDescriptors),
}

func loadRangeDescriptor(
	db storage.Engine, rangeID roachpb.RangeID,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(30590)
	var desc roachpb.RangeDescriptor
	handleKV := func(kv storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(30594)
		if kv.Key.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(30600)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(30601)
		}
		__antithesis_instrumentation__.Notify(30595)
		if err := kvserver.IsRangeDescriptorKey(kv.Key); err != nil {
			__antithesis_instrumentation__.Notify(30602)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(30603)
		}
		__antithesis_instrumentation__.Notify(30596)
		if len(kv.Value) == 0 {
			__antithesis_instrumentation__.Notify(30604)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(30605)
		}
		__antithesis_instrumentation__.Notify(30597)
		if err := (roachpb.Value{RawBytes: kv.Value}).GetProto(&desc); err != nil {
			__antithesis_instrumentation__.Notify(30606)
			log.Warningf(context.Background(), "ignoring range descriptor due to error %s: %+v", err, kv)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30607)
		}
		__antithesis_instrumentation__.Notify(30598)
		if desc.RangeID == rangeID {
			__antithesis_instrumentation__.Notify(30608)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(30609)
		}
		__antithesis_instrumentation__.Notify(30599)
		return nil
	}
	__antithesis_instrumentation__.Notify(30591)

	start := keys.LocalRangePrefix
	end := keys.LocalRangeMax

	if err := db.MVCCIterate(start, end, storage.MVCCKeyAndIntentsIterKind, handleKV); err != nil {
		__antithesis_instrumentation__.Notify(30610)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(30611)
	}
	__antithesis_instrumentation__.Notify(30592)
	if desc.RangeID == rangeID {
		__antithesis_instrumentation__.Notify(30612)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(30613)
	}
	__antithesis_instrumentation__.Notify(30593)
	return roachpb.RangeDescriptor{}, fmt.Errorf("range descriptor %d not found", rangeID)
}

func runDebugRangeDescriptors(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30614)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30617)
	}
	__antithesis_instrumentation__.Notify(30615)

	start := keys.LocalRangePrefix
	end := keys.LocalRangeMax

	return db.MVCCIterate(start, end, storage.MVCCKeyAndIntentsIterKind, func(kv storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(30618)
		if kvserver.IsRangeDescriptorKey(kv.Key) != nil {
			__antithesis_instrumentation__.Notify(30620)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30621)
		}
		__antithesis_instrumentation__.Notify(30619)
		kvserver.PrintMVCCKeyValue(kv)
		return nil
	})
}

var debugDecodeKeyCmd = &cobra.Command{
	Use:   "decode-key",
	Short: "decode <key>",
	Long: `
Decode a hexadecimal-encoded key and pretty-print it. For example:

	$ decode-key BB89F902ADB43000151C2D1ED07DE6C009
	/Table/51/1/44938288/1521140384.514565824,0
`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(30622)
		for _, arg := range args {
			__antithesis_instrumentation__.Notify(30624)
			b, err := gohex.DecodeString(arg)
			if err != nil {
				__antithesis_instrumentation__.Notify(30627)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30628)
			}
			__antithesis_instrumentation__.Notify(30625)
			k, err := storage.DecodeMVCCKey(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(30629)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30630)
			}
			__antithesis_instrumentation__.Notify(30626)
			fmt.Println(k)
		}
		__antithesis_instrumentation__.Notify(30623)
		return nil
	},
}

var debugDecodeValueCmd = &cobra.Command{
	Use:   "decode-value",
	Short: "decode-value <key> <value>",
	Long: `
Decode and print a hexadecimal-encoded key-value pair.

	$ decode-value <TBD> <TBD>
	<TBD>
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(30631)
		var bs [][]byte
		for _, arg := range args {
			__antithesis_instrumentation__.Notify(30634)
			b, err := gohex.DecodeString(arg)
			if err != nil {
				__antithesis_instrumentation__.Notify(30636)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30637)
			}
			__antithesis_instrumentation__.Notify(30635)
			bs = append(bs, b)
		}
		__antithesis_instrumentation__.Notify(30632)

		isTS := bytes.HasPrefix(bs[0], keys.TimeseriesPrefix)
		k, err := storage.DecodeMVCCKey(bs[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(30638)

			if !isTS {
				__antithesis_instrumentation__.Notify(30640)
				if k, ok := storage.DecodeEngineKey(bs[0]); ok {
					__antithesis_instrumentation__.Notify(30642)
					kvserver.PrintEngineKeyValue(k, bs[1])
					return nil
				} else {
					__antithesis_instrumentation__.Notify(30643)
				}
				__antithesis_instrumentation__.Notify(30641)
				fmt.Printf("unable to decode key: %v, assuming it's a roachpb.Key with fake timestamp;\n"+
					"if the result below looks like garbage, then it likely is:\n\n", err)
			} else {
				__antithesis_instrumentation__.Notify(30644)
			}
			__antithesis_instrumentation__.Notify(30639)
			k = storage.MVCCKey{
				Key:       bs[0],
				Timestamp: hlc.Timestamp{WallTime: 987654321},
			}
		} else {
			__antithesis_instrumentation__.Notify(30645)
		}
		__antithesis_instrumentation__.Notify(30633)

		kvserver.PrintMVCCKeyValue(storage.MVCCKeyValue{
			Key:   k,
			Value: bs[1],
		})
		return nil
	},
}

var debugDecodeProtoName string
var debugDecodeProtoEmitDefaults bool
var debugDecodeProtoCmd = &cobra.Command{
	Use:   "decode-proto",
	Short: "decode-proto <proto> --name=<fully qualified proto name>",
	Long: `
Read from stdin and attempt to decode any hex or base64 encoded proto fields and
output them as JSON. All other fields will be outputted unchanged. Output fields
will be separated by tabs.

The default value for --schema is 'cockroach.sql.sqlbase.Descriptor'.
For example:

$ decode-proto < cat debug/system.decsriptor.txt
id	descriptor	hex_descriptor
1	\022!\012\006system\020\001\032\025\012\011\012\005admin\0200\012\010\012\004root\0200	{"database": {"id": 1, "modificationTime": {}, "name": "system", "privileges": {"users": [{"privileges": 48, "user": "admin"}, {"privileges": 48, "user": "root"}]}}}
...
`,
	Args: cobra.ArbitraryArgs,
	RunE: runDebugDecodeProto,
}

var debugRaftLogCmd = &cobra.Command{
	Use:   "raft-log <directory> <range id>",
	Short: "print the raft log for a range",
	Long: `
Prints all log entries in a store for the given range.
`,
	Args: cobra.ExactArgs(2),
	RunE: clierrorplus.MaybeDecorateError(runDebugRaftLog),
}

func runDebugRaftLog(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30646)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30649)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30650)
	}
	__antithesis_instrumentation__.Notify(30647)

	rangeID, err := parseRangeID(args[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(30651)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30652)
	}
	__antithesis_instrumentation__.Notify(30648)

	start := keys.RaftLogPrefix(rangeID)
	end := keys.RaftLogPrefix(rangeID).PrefixEnd()
	fmt.Printf("Printing keys %s -> %s (RocksDB keys: %#x - %#x )\n",
		start, end,
		string(storage.EncodeMVCCKey(storage.MakeMVCCMetadataKey(start))),
		string(storage.EncodeMVCCKey(storage.MakeMVCCMetadataKey(end))))

	return db.MVCCIterate(start, end, storage.MVCCKeyIterKind, func(kv storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(30653)
		kvserver.PrintMVCCKeyValue(kv)
		return nil
	})
}

var debugGCCmd = &cobra.Command{
	Use:   "estimate-gc <directory> [range id] [ttl-in-seconds] [intent-age-as-duration]",
	Short: "find out what a GC run would do",
	Long: `
Sets up (but does not run) a GC collection cycle, giving insight into how much
work would be done (assuming all intent resolution and pushes succeed).

Without a RangeID specified on the command line, runs the analysis for all
ranges individually.

Uses a configurable GC policy, with a default 24 hour TTL, for old versions and
2 hour intent resolution threshold.
`,
	Args: cobra.RangeArgs(1, 4),
	RunE: clierrorplus.MaybeDecorateError(runDebugGCCmd),
}

func runDebugGCCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30654)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	var rangeID roachpb.RangeID
	gcTTL := 24 * time.Hour
	intentAgeThreshold := gc.IntentAgeThreshold.Default()
	intentBatchSize := gc.MaxIntentsPerCleanupBatch.Default()

	if len(args) > 3 {
		__antithesis_instrumentation__.Notify(30662)
		var err error
		if intentAgeThreshold, err = parsePositiveDuration(args[3]); err != nil {
			__antithesis_instrumentation__.Notify(30663)
			return errors.Wrapf(err, "unable to parse %v as intent age threshold", args[3])
		} else {
			__antithesis_instrumentation__.Notify(30664)
		}
	} else {
		__antithesis_instrumentation__.Notify(30665)
	}
	__antithesis_instrumentation__.Notify(30655)
	if len(args) > 2 {
		__antithesis_instrumentation__.Notify(30666)
		gcTTLInSeconds, err := parsePositiveInt(args[2])
		if err != nil {
			__antithesis_instrumentation__.Notify(30668)
			return errors.Wrapf(err, "unable to parse %v as TTL", args[2])
		} else {
			__antithesis_instrumentation__.Notify(30669)
		}
		__antithesis_instrumentation__.Notify(30667)
		gcTTL = time.Duration(gcTTLInSeconds) * time.Second
	} else {
		__antithesis_instrumentation__.Notify(30670)
	}
	__antithesis_instrumentation__.Notify(30656)
	if len(args) > 1 {
		__antithesis_instrumentation__.Notify(30671)
		var err error
		if rangeID, err = parseRangeID(args[1]); err != nil {
			__antithesis_instrumentation__.Notify(30672)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30673)
		}
	} else {
		__antithesis_instrumentation__.Notify(30674)
	}
	__antithesis_instrumentation__.Notify(30657)

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30675)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30676)
	}
	__antithesis_instrumentation__.Notify(30658)

	start := keys.RangeDescriptorKey(roachpb.RKeyMin)
	end := keys.RangeDescriptorKey(roachpb.RKeyMax)

	var descs []roachpb.RangeDescriptor

	if _, err := storage.MVCCIterate(context.Background(), db, start, end, hlc.MaxTimestamp,
		storage.MVCCScanOptions{Inconsistent: true}, func(kv roachpb.KeyValue) error {
			__antithesis_instrumentation__.Notify(30677)
			var desc roachpb.RangeDescriptor
			_, suffix, _, err := keys.DecodeRangeKey(kv.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(30683)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30684)
			}
			__antithesis_instrumentation__.Notify(30678)
			if !bytes.Equal(suffix, keys.LocalRangeDescriptorSuffix) {
				__antithesis_instrumentation__.Notify(30685)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(30686)
			}
			__antithesis_instrumentation__.Notify(30679)
			if err := kv.Value.GetProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(30687)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30688)
			}
			__antithesis_instrumentation__.Notify(30680)
			if desc.RangeID == rangeID || func() bool {
				__antithesis_instrumentation__.Notify(30689)
				return rangeID == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(30690)
				descs = append(descs, desc)
			} else {
				__antithesis_instrumentation__.Notify(30691)
			}
			__antithesis_instrumentation__.Notify(30681)
			if desc.RangeID == rangeID {
				__antithesis_instrumentation__.Notify(30692)
				return iterutil.StopIteration()
			} else {
				__antithesis_instrumentation__.Notify(30693)
			}
			__antithesis_instrumentation__.Notify(30682)
			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(30694)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30695)
	}
	__antithesis_instrumentation__.Notify(30659)

	if len(descs) == 0 {
		__antithesis_instrumentation__.Notify(30696)
		return fmt.Errorf("no range matching the criteria found")
	} else {
		__antithesis_instrumentation__.Notify(30697)
	}
	__antithesis_instrumentation__.Notify(30660)

	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(30698)
		snap := db.NewSnapshot()
		defer snap.Close()
		now := hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}
		thresh := gc.CalculateThreshold(now, gcTTL)
		info, err := gc.Run(
			context.Background(),
			&desc, snap,
			now, thresh,
			gc.RunOptions{IntentAgeThreshold: intentAgeThreshold, MaxIntentsPerIntentCleanupBatch: intentBatchSize},
			gcTTL, gc.NoopGCer{},
			func(_ context.Context, _ []roachpb.Intent) error {
				__antithesis_instrumentation__.Notify(30701)
				return nil
			},
			func(_ context.Context, _ *roachpb.Transaction) error {
				__antithesis_instrumentation__.Notify(30702)
				return nil
			},
		)
		__antithesis_instrumentation__.Notify(30699)
		if err != nil {
			__antithesis_instrumentation__.Notify(30703)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30704)
		}
		__antithesis_instrumentation__.Notify(30700)
		fmt.Printf("RangeID: %d [%s, %s):\n", desc.RangeID, desc.StartKey, desc.EndKey)
		_, _ = pretty.Println(info)
	}
	__antithesis_instrumentation__.Notify(30661)
	return nil
}

var DebugPebbleCmd = &cobra.Command{
	Use:   "pebble [command]",
	Short: "run a Pebble introspection tool command",
	Long: `
Allows the use of pebble tools, such as to introspect manifests, SSTables, etc.
`,
}

var debugEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "output environment settings",
	Long: `
Output environment variables that influence configuration.
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		__antithesis_instrumentation__.Notify(30705)
		env := envutil.GetEnvReport()
		fmt.Print(env)
	},
}

var debugCompactCmd = &cobra.Command{
	Use:   "compact <directory>",
	Short: "compact the sstables in a store",
	Long: `
Compact the sstables in a store.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugCompact),
}

func runDebugCompact(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30706)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, false, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(30711)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30712)
	}

	{
		__antithesis_instrumentation__.Notify(30713)
		approxBytesBefore, err := db.ApproximateDiskBytes(roachpb.KeyMin, roachpb.KeyMax)
		if err != nil {
			__antithesis_instrumentation__.Notify(30715)
			return errors.Wrap(err, "while computing approximate size before compaction")
		} else {
			__antithesis_instrumentation__.Notify(30716)
		}
		__antithesis_instrumentation__.Notify(30714)
		fmt.Printf("approximate reported database size before compaction: %s\n", humanizeutil.IBytes(int64(approxBytesBefore)))
	}
	__antithesis_instrumentation__.Notify(30707)

	errCh := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(30717)
		errCh <- errors.Wrap(db.Compact(), "while compacting")
	}()
	__antithesis_instrumentation__.Notify(30708)

	ticker := time.NewTicker(time.Minute)
	for done := false; !done; {
		__antithesis_instrumentation__.Notify(30718)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(30719)
			fmt.Printf("%s\n", db.GetMetrics())
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(30720)
			ticker.Stop()
			if err != nil {
				__antithesis_instrumentation__.Notify(30722)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30723)
			}
			__antithesis_instrumentation__.Notify(30721)
			done = true
		}
	}
	__antithesis_instrumentation__.Notify(30709)
	fmt.Printf("%s\n", db.GetMetrics())

	{
		__antithesis_instrumentation__.Notify(30724)
		approxBytesAfter, err := db.ApproximateDiskBytes(roachpb.KeyMin, roachpb.KeyMax)
		if err != nil {
			__antithesis_instrumentation__.Notify(30726)
			return errors.Wrap(err, "while computing approximate size after compaction")
		} else {
			__antithesis_instrumentation__.Notify(30727)
		}
		__antithesis_instrumentation__.Notify(30725)
		fmt.Printf("approximate reported database size after compaction: %s\n", humanizeutil.IBytes(int64(approxBytesAfter)))
	}
	__antithesis_instrumentation__.Notify(30710)
	return nil
}

var debugGossipValuesCmd = &cobra.Command{
	Use:   "gossip-values",
	Short: "dump all the values in a node's gossip instance",
	Long: `
Pretty-prints the values in a node's gossip instance.

Can connect to a running server to get the values or can be provided with
a JSON file captured from a node's /_status/gossip/ debug endpoint.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runDebugGossipValues),
}

func runDebugGossipValues(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30728)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gossipInfo *gossip.InfoStatus
	if debugCtx.inputFile != "" {
		__antithesis_instrumentation__.Notify(30731)
		file, err := os.Open(debugCtx.inputFile)
		if err != nil {
			__antithesis_instrumentation__.Notify(30733)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30734)
		}
		__antithesis_instrumentation__.Notify(30732)
		defer file.Close()
		gossipInfo = new(gossip.InfoStatus)
		if err := jsonpb.Unmarshal(file, gossipInfo); err != nil {
			__antithesis_instrumentation__.Notify(30735)
			return errors.Wrap(err, "failed to parse provided file as gossip.InfoStatus")
		} else {
			__antithesis_instrumentation__.Notify(30736)
		}
	} else {
		__antithesis_instrumentation__.Notify(30737)
		conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(30739)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30740)
		}
		__antithesis_instrumentation__.Notify(30738)
		defer finish()

		status := serverpb.NewStatusClient(conn)
		gossipInfo, err = status.Gossip(ctx, &serverpb.GossipRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(30741)
			return errors.Wrap(err, "failed to retrieve gossip from server")
		} else {
			__antithesis_instrumentation__.Notify(30742)
		}
	}
	__antithesis_instrumentation__.Notify(30729)

	output, err := parseGossipValues(gossipInfo)
	if err != nil {
		__antithesis_instrumentation__.Notify(30743)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30744)
	}
	__antithesis_instrumentation__.Notify(30730)
	fmt.Println(output)
	return nil
}

func parseGossipValues(gossipInfo *gossip.InfoStatus) (string, error) {
	__antithesis_instrumentation__.Notify(30745)
	var output []string
	for key, info := range gossipInfo.Infos {
		__antithesis_instrumentation__.Notify(30747)
		bytes, err := info.Value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(30749)
			return "", errors.Wrapf(err, "failed to extract bytes for key %q", key)
		} else {
			__antithesis_instrumentation__.Notify(30750)
		}
		__antithesis_instrumentation__.Notify(30748)
		if key == gossip.KeyClusterID || func() bool {
			__antithesis_instrumentation__.Notify(30751)
			return key == gossip.KeySentinel == true
		}() == true {
			__antithesis_instrumentation__.Notify(30752)
			clusterID, err := uuid.FromBytes(bytes)
			if err != nil {
				__antithesis_instrumentation__.Notify(30754)
				return "", errors.Wrapf(err, "failed to parse value for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(30755)
			}
			__antithesis_instrumentation__.Notify(30753)
			output = append(output, fmt.Sprintf("%q: %v", key, clusterID))
		} else {
			__antithesis_instrumentation__.Notify(30756)
			if key == gossip.KeyDeprecatedSystemConfig {
				__antithesis_instrumentation__.Notify(30757)
				if debugCtx.printSystemConfig {
					__antithesis_instrumentation__.Notify(30758)
					var config config.SystemConfigEntries
					if err := protoutil.Unmarshal(bytes, &config); err != nil {
						__antithesis_instrumentation__.Notify(30760)
						return "", errors.Wrapf(err, "failed to parse value for key %q", key)
					} else {
						__antithesis_instrumentation__.Notify(30761)
					}
					__antithesis_instrumentation__.Notify(30759)
					output = append(output, fmt.Sprintf("%q: %+v", key, config))
				} else {
					__antithesis_instrumentation__.Notify(30762)
					output = append(output, fmt.Sprintf("%q: omitted", key))
				}
			} else {
				__antithesis_instrumentation__.Notify(30763)
				if key == gossip.KeyFirstRangeDescriptor {
					__antithesis_instrumentation__.Notify(30764)
					var desc roachpb.RangeDescriptor
					if err := protoutil.Unmarshal(bytes, &desc); err != nil {
						__antithesis_instrumentation__.Notify(30766)
						return "", errors.Wrapf(err, "failed to parse value for key %q", key)
					} else {
						__antithesis_instrumentation__.Notify(30767)
					}
					__antithesis_instrumentation__.Notify(30765)
					output = append(output, fmt.Sprintf("%q: %v", key, desc))
				} else {
					__antithesis_instrumentation__.Notify(30768)
					if gossip.IsNodeIDKey(key) {
						__antithesis_instrumentation__.Notify(30769)
						var desc roachpb.NodeDescriptor
						if err := protoutil.Unmarshal(bytes, &desc); err != nil {
							__antithesis_instrumentation__.Notify(30771)
							return "", errors.Wrapf(err, "failed to parse value for key %q", key)
						} else {
							__antithesis_instrumentation__.Notify(30772)
						}
						__antithesis_instrumentation__.Notify(30770)
						output = append(output, fmt.Sprintf("%q: %+v", key, desc))
					} else {
						__antithesis_instrumentation__.Notify(30773)
						if strings.HasPrefix(key, gossip.KeyStorePrefix) {
							__antithesis_instrumentation__.Notify(30774)
							var desc roachpb.StoreDescriptor
							if err := protoutil.Unmarshal(bytes, &desc); err != nil {
								__antithesis_instrumentation__.Notify(30776)
								return "", errors.Wrapf(err, "failed to parse value for key %q", key)
							} else {
								__antithesis_instrumentation__.Notify(30777)
							}
							__antithesis_instrumentation__.Notify(30775)
							output = append(output, fmt.Sprintf("%q: %+v", key, desc))
						} else {
							__antithesis_instrumentation__.Notify(30778)
							if strings.HasPrefix(key, gossip.KeyNodeLivenessPrefix) {
								__antithesis_instrumentation__.Notify(30779)
								var liveness livenesspb.Liveness
								if err := protoutil.Unmarshal(bytes, &liveness); err != nil {
									__antithesis_instrumentation__.Notify(30781)
									return "", errors.Wrapf(err, "failed to parse value for key %q", key)
								} else {
									__antithesis_instrumentation__.Notify(30782)
								}
								__antithesis_instrumentation__.Notify(30780)
								output = append(output, fmt.Sprintf("%q: %+v", key, liveness))
							} else {
								__antithesis_instrumentation__.Notify(30783)
								if strings.HasPrefix(key, gossip.KeyNodeHealthAlertPrefix) {
									__antithesis_instrumentation__.Notify(30784)
									var healthAlert statuspb.HealthCheckResult
									if err := protoutil.Unmarshal(bytes, &healthAlert); err != nil {
										__antithesis_instrumentation__.Notify(30786)
										return "", errors.Wrapf(err, "failed to parse value for key %q", key)
									} else {
										__antithesis_instrumentation__.Notify(30787)
									}
									__antithesis_instrumentation__.Notify(30785)
									output = append(output, fmt.Sprintf("%q: %+v", key, healthAlert))
								} else {
									__antithesis_instrumentation__.Notify(30788)
									if strings.HasPrefix(key, gossip.KeyDistSQLNodeVersionKeyPrefix) {
										__antithesis_instrumentation__.Notify(30789)
										var version execinfrapb.DistSQLVersionGossipInfo
										if err := protoutil.Unmarshal(bytes, &version); err != nil {
											__antithesis_instrumentation__.Notify(30791)
											return "", errors.Wrapf(err, "failed to parse value for key %q", key)
										} else {
											__antithesis_instrumentation__.Notify(30792)
										}
										__antithesis_instrumentation__.Notify(30790)
										output = append(output, fmt.Sprintf("%q: %+v", key, version))
									} else {
										__antithesis_instrumentation__.Notify(30793)
										if strings.HasPrefix(key, gossip.KeyDistSQLDrainingPrefix) {
											__antithesis_instrumentation__.Notify(30794)
											var drainingInfo execinfrapb.DistSQLDrainingInfo
											if err := protoutil.Unmarshal(bytes, &drainingInfo); err != nil {
												__antithesis_instrumentation__.Notify(30796)
												return "", errors.Wrapf(err, "failed to parse value for key %q", key)
											} else {
												__antithesis_instrumentation__.Notify(30797)
											}
											__antithesis_instrumentation__.Notify(30795)
											output = append(output, fmt.Sprintf("%q: %+v", key, drainingInfo))
										} else {
											__antithesis_instrumentation__.Notify(30798)
											if strings.HasPrefix(key, gossip.KeyGossipClientsPrefix) {
												__antithesis_instrumentation__.Notify(30799)
												output = append(output, fmt.Sprintf("%q: %v", key, string(bytes)))
											} else {
												__antithesis_instrumentation__.Notify(30800)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(30746)

	sort.Strings(output)
	return strings.Join(output, "\n"), nil
}

var debugSyncBenchCmd = &cobra.Command{
	Use:   "syncbench [directory]",
	Short: "Run a performance test for WAL sync speed",
	Long: `
`,
	Args:   cobra.MaximumNArgs(1),
	Hidden: true,
	RunE:   clierrorplus.MaybeDecorateError(runDebugSyncBench),
}

var syncBenchOpts = syncbench.Options{
	Concurrency: 1,
	Duration:    10 * time.Second,
	LogOnly:     true,
}

func runDebugSyncBench(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30801)
	syncBenchOpts.Dir = "./testdb"
	if len(args) == 1 {
		__antithesis_instrumentation__.Notify(30803)
		syncBenchOpts.Dir = args[0]
	} else {
		__antithesis_instrumentation__.Notify(30804)
	}
	__antithesis_instrumentation__.Notify(30802)
	return syncbench.Run(syncBenchOpts)
}

var debugUnsafeRemoveDeadReplicasCmd = &cobra.Command{
	Use:   "unsafe-remove-dead-replicas --dead-store-ids=[store ID,...] [path]",
	Short: "Unsafely attempt to recover a range that has lost quorum",
	Long: `

This command is UNSAFE and should only be used with the supervision of
a Cockroach Labs engineer. It is a last-resort option to recover data
after multiple node failures. The recovered data is not guaranteed to
be consistent. If a suitable backup exists, restore it instead of
using this tool.

The --dead-store-ids flag takes a comma-separated list of dead store IDs and
scans this store for any ranges unable to make progress (as indicated by the
remaining replicas not marked as dead). For each such replica in which the local
store is the voter with the highest StoreID, the range descriptors will be (only
when run on that store) rewritten in-place to reflect a replication
configuration in which the local store is the sole voter (and thus able to make
progress).

The intention is for the tool to be run against all stores in the cluster, which
will attempt to recover all ranges in the system for which such an operation is
necessary. This is the safest and most straightforward option, but incurs global
downtime. When availability problems are isolated to a small number of ranges,
it is also possible to restrict the set of nodes to be restarted (all nodes that
have a store that has the highest voting StoreID in one of the ranges that need
to be recovered), and to perform the restarts one-by-one (assuming system ranges
are not affected). With this latter strategy, restarts of additional replicas
may still be necessary, owing to the fact that the former leaseholder's epoch
may still be live, even though that leaseholder may now be on the "unrecovered"
side of the range and thus still unavailable. This case can be detected via the
range status of the affected range(s) and by restarting the listed leaseholder.

This command will prompt for confirmation before committing its changes.

WARNINGS

This tool may cause previously committed data to be lost. It does not preserve
atomicity of transactions, so further inconsistencies and undefined behavior may
result. In the worst case, a corruption of the cluster-internal metadata may
occur, which would complicate the recovery further. It is recommended to take a
filesystem-level backup or snapshot of the nodes to be affected before running
this command (it is not safe to take a filesystem-level backup of a running
node, but it is possible while the node is stopped). A cluster that had this
tool used against it is no longer fit for production use. It must be
re-initialized from a backup.

Before proceeding at the yes/no prompt, review the ranges that are affected to
consider the possible impact of inconsistencies. Further remediation may be
necessary after running this tool, including dropping and recreating affected
indexes, or in the worst case creating a new backup or export of this cluster's
data for restoration into a brand new cluster. Because of the latter
possibilities, this tool is a slower means of disaster recovery than restoring
from a backup.

Must only be used when the dead stores are lost and unrecoverable. If the dead
stores were to rejoin the cluster after this command was used, data may be
corrupted.

After this command is used, the node should not be restarted until at least 10
seconds have passed since it was stopped. Restarting it too early may lead to
things getting stuck (if it happens, it can be fixed by restarting a second
time).
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugUnsafeRemoveDeadReplicas),
}

var removeDeadReplicasOpts struct {
	deadStoreIDs []int
}

func runDebugUnsafeRemoveDeadReplicas(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30805)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(args[0], stopper, false, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30811)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30812)
	}
	__antithesis_instrumentation__.Notify(30806)

	deadStoreIDs := map[roachpb.StoreID]struct{}{}
	for _, id := range removeDeadReplicasOpts.deadStoreIDs {
		__antithesis_instrumentation__.Notify(30813)
		deadStoreIDs[roachpb.StoreID(id)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(30807)
	batch, err := removeDeadReplicas(db, deadStoreIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(30814)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30815)
		if batch == nil {
			__antithesis_instrumentation__.Notify(30816)
			fmt.Printf("Nothing to do\n")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30817)
		}
	}
	__antithesis_instrumentation__.Notify(30808)
	defer batch.Close()

	fmt.Printf("Proceed with the above rewrites? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		__antithesis_instrumentation__.Notify(30818)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30819)
	}
	__antithesis_instrumentation__.Notify(30809)
	fmt.Printf("\n")
	if line[0] == 'y' || func() bool {
		__antithesis_instrumentation__.Notify(30820)
		return line[0] == 'Y' == true
	}() == true {
		__antithesis_instrumentation__.Notify(30821)
		fmt.Printf("Committing\n")
		if err := batch.Commit(true); err != nil {
			__antithesis_instrumentation__.Notify(30822)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30823)
		}
	} else {
		__antithesis_instrumentation__.Notify(30824)
		fmt.Printf("Aborting\n")
	}
	__antithesis_instrumentation__.Notify(30810)
	return nil
}

func removeDeadReplicas(
	db storage.Engine, deadStoreIDs map[roachpb.StoreID]struct{},
) (storage.Batch, error) {
	__antithesis_instrumentation__.Notify(30825)
	clock := hlc.NewClock(hlc.UnixNano, 0)

	ctx := context.Background()

	storeIdent, err := kvserver.ReadStoreIdent(ctx, db)
	if err != nil {
		__antithesis_instrumentation__.Notify(30832)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30833)
	}
	__antithesis_instrumentation__.Notify(30826)
	fmt.Printf("Scanning replicas on store %s for dead peers %v\n", storeIdent.String(),
		removeDeadReplicasOpts.deadStoreIDs)

	if _, ok := deadStoreIDs[storeIdent.StoreID]; ok {
		__antithesis_instrumentation__.Notify(30834)
		return nil, errors.Errorf("this store's ID (%s) marked as dead, aborting", storeIdent.StoreID)
	} else {
		__antithesis_instrumentation__.Notify(30835)
	}
	__antithesis_instrumentation__.Notify(30827)

	var newDescs []roachpb.RangeDescriptor

	err = kvserver.IterateRangeDescriptorsFromDisk(ctx, db, func(desc roachpb.RangeDescriptor) error {
		__antithesis_instrumentation__.Notify(30836)
		numDeadPeers := 0
		allReplicas := desc.Replicas().Descriptors()
		maxLiveVoter := roachpb.StoreID(-1)
		for _, rep := range allReplicas {
			__antithesis_instrumentation__.Notify(30841)
			if _, ok := deadStoreIDs[rep.StoreID]; ok {
				__antithesis_instrumentation__.Notify(30843)
				numDeadPeers++
				continue
			} else {
				__antithesis_instrumentation__.Notify(30844)
			}
			__antithesis_instrumentation__.Notify(30842)

			if rep.IsVoterNewConfig() && func() bool {
				__antithesis_instrumentation__.Notify(30845)
				return rep.StoreID > maxLiveVoter == true
			}() == true {
				__antithesis_instrumentation__.Notify(30846)
				maxLiveVoter = rep.StoreID
			} else {
				__antithesis_instrumentation__.Notify(30847)
			}
		}
		__antithesis_instrumentation__.Notify(30837)

		if numDeadPeers == 0 {
			__antithesis_instrumentation__.Notify(30848)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30849)
		}
		__antithesis_instrumentation__.Notify(30838)
		if storeIdent.StoreID != maxLiveVoter {
			__antithesis_instrumentation__.Notify(30850)
			fmt.Printf("not designated survivor, skipping: %s\n", &desc)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30851)
		}
		__antithesis_instrumentation__.Notify(30839)

		if desc.Replicas().CanMakeProgress(func(rep roachpb.ReplicaDescriptor) bool {
			__antithesis_instrumentation__.Notify(30852)
			_, ok := deadStoreIDs[rep.StoreID]
			return !ok
		}) {
			__antithesis_instrumentation__.Notify(30853)
			fmt.Printf("replica has not lost quorum, skipping: %s\n", &desc)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(30854)
		}
		__antithesis_instrumentation__.Notify(30840)

		newDesc := desc

		replicas := []roachpb.ReplicaDescriptor{{
			NodeID:    storeIdent.NodeID,
			StoreID:   storeIdent.StoreID,
			ReplicaID: desc.NextReplicaID,
		}}
		newDesc.SetReplicas(roachpb.MakeReplicaSet(replicas))
		newDesc.NextReplicaID++
		fmt.Printf("replica has lost quorum, recovering: %s -> %s\n", &desc, &newDesc)
		newDescs = append(newDescs, newDesc)
		return nil
	})
	__antithesis_instrumentation__.Notify(30828)
	if err != nil {
		__antithesis_instrumentation__.Notify(30855)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(30856)
	}
	__antithesis_instrumentation__.Notify(30829)

	if len(newDescs) == 0 {
		__antithesis_instrumentation__.Notify(30857)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(30858)
	}
	__antithesis_instrumentation__.Notify(30830)

	batch := db.NewBatch()
	for _, desc := range newDescs {
		__antithesis_instrumentation__.Notify(30859)

		key := keys.RangeDescriptorKey(desc.StartKey)
		sl := stateloader.Make(desc.RangeID)
		ms, err := sl.LoadMVCCStats(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(30864)
			return nil, errors.Wrap(err, "loading MVCCStats")
		} else {
			__antithesis_instrumentation__.Notify(30865)
		}
		__antithesis_instrumentation__.Notify(30860)
		err = storage.MVCCPutProto(ctx, batch, &ms, key, clock.Now(), nil, &desc)
		if wiErr := (*roachpb.WriteIntentError)(nil); errors.As(err, &wiErr) {
			__antithesis_instrumentation__.Notify(30866)
			if len(wiErr.Intents) != 1 {
				__antithesis_instrumentation__.Notify(30870)
				return nil, errors.Errorf("expected 1 intent, found %d: %s", len(wiErr.Intents), wiErr)
			} else {
				__antithesis_instrumentation__.Notify(30871)
			}
			__antithesis_instrumentation__.Notify(30867)
			intent := wiErr.Intents[0]

			fmt.Printf("aborting intent: %s (txn %s)\n", key, intent.Txn.ID)

			txnKey := keys.TransactionKey(intent.Txn.Key, intent.Txn.ID)
			if err := storage.MVCCDelete(ctx, batch, &ms, txnKey, hlc.Timestamp{}, nil); err != nil {
				__antithesis_instrumentation__.Notify(30872)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(30873)
			}
			__antithesis_instrumentation__.Notify(30868)
			update := roachpb.LockUpdate{
				Span:   roachpb.Span{Key: intent.Key},
				Txn:    intent.Txn,
				Status: roachpb.ABORTED,
			}
			if _, err := storage.MVCCResolveWriteIntent(ctx, batch, &ms, update); err != nil {
				__antithesis_instrumentation__.Notify(30874)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(30875)
			}
			__antithesis_instrumentation__.Notify(30869)

			if err := storage.MVCCPutProto(ctx, batch, &ms, key, clock.Now(),
				nil, &desc); err != nil {
				__antithesis_instrumentation__.Notify(30876)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(30877)
			}
		} else {
			__antithesis_instrumentation__.Notify(30878)
			if err != nil {
				__antithesis_instrumentation__.Notify(30879)
				batch.Close()
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(30880)
			}
		}
		__antithesis_instrumentation__.Notify(30861)

		replicas := desc.Replicas().Descriptors()
		if len(replicas) != 1 {
			__antithesis_instrumentation__.Notify(30881)
			return nil, errors.Errorf("expected 1 replica, got %v", replicas)
		} else {
			__antithesis_instrumentation__.Notify(30882)
		}
		__antithesis_instrumentation__.Notify(30862)
		if err := sl.SetRaftReplicaID(ctx, batch, replicas[0].ReplicaID); err != nil {
			__antithesis_instrumentation__.Notify(30883)
			return nil, errors.Wrapf(err, "failed to write new replica ID for range %d", desc.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(30884)
		}
		__antithesis_instrumentation__.Notify(30863)

		if err := sl.SetMVCCStats(ctx, batch, &ms); err != nil {
			__antithesis_instrumentation__.Notify(30885)
			return nil, errors.Wrap(err, "updating MVCCStats")
		} else {
			__antithesis_instrumentation__.Notify(30886)
		}
	}
	__antithesis_instrumentation__.Notify(30831)

	return batch, nil
}

var debugMergeLogsCmd = &cobra.Command{
	Use:   "merge-logs <log file globs>",
	Short: "merge multiple log files from different machines into a single stream",
	Long: `
Takes a list of glob patterns (not left exclusively to the shell because of
MAX_ARG_STRLEN, usually 128kB) which will be walked and whose contained log
files and merged them into a single stream printed to stdout. Files not matching
the log file name pattern are ignored. If log lines appear out of order within
a file (which happens), the timestamp is ratcheted to the highest value seen so far.
The command supports efficient time filtering as well as multiline regexp pattern
matching via flags. If the filter regexp contains captures, such as
'^abc(hello)def(world)', only the captured parts will be printed.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDebugMergeLogs,
}

const logFilePattern = "^(?:(?P<fpath>.*)/)?" + log.FileNamePattern + "$"

var debugMergeLogsOpts = struct {
	from           time.Time
	to             time.Time
	filter         *regexp.Regexp
	program        *regexp.Regexp
	file           *regexp.Regexp
	keepRedactable bool
	prefix         string
	redactInput    bool
	format         string
	useColor       forceColor
}{
	program:        nil,
	file:           regexp.MustCompile(logFilePattern),
	keepRedactable: true,
	redactInput:    false,
}

func runDebugMergeLogs(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30887)
	o := debugMergeLogsOpts
	p := newFilePrefixer(withTemplate(o.prefix))

	inputEditMode := log.SelectEditMode(o.redactInput, o.keepRedactable)

	s, err := newMergedStreamFromPatterns(context.Background(),
		args, o.file, o.program, o.from, o.to, inputEditMode, o.format, p)
	if err != nil {
		__antithesis_instrumentation__.Notify(30891)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30892)
	}
	__antithesis_instrumentation__.Notify(30888)

	autoDetect := func(outStream io.Writer) (ttycolor.Profile, error) {
		__antithesis_instrumentation__.Notify(30893)
		if f, ok := outStream.(*os.File); ok {
			__antithesis_instrumentation__.Notify(30895)

			return ttycolor.DetectProfile(f)
		} else {
			__antithesis_instrumentation__.Notify(30896)
		}
		__antithesis_instrumentation__.Notify(30894)
		return nil, nil
	}
	__antithesis_instrumentation__.Notify(30889)
	outStream := cmd.OutOrStdout()
	var cp ttycolor.Profile

	switch o.useColor {
	case forceColorOff:
		__antithesis_instrumentation__.Notify(30897)

	case forceColorOn:
		__antithesis_instrumentation__.Notify(30898)

		var err error
		cp, err = autoDetect(outStream)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(30901)
			return cp == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(30902)

			cp = ttycolor.Profile8
		} else {
			__antithesis_instrumentation__.Notify(30903)
		}
	case forceColorAuto:
		__antithesis_instrumentation__.Notify(30899)
		var err error
		cp, err = autoDetect(outStream)
		if err != nil {
			__antithesis_instrumentation__.Notify(30904)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30905)
		}
	default:
		__antithesis_instrumentation__.Notify(30900)
	}
	__antithesis_instrumentation__.Notify(30890)

	return writeLogStream(s, outStream, o.filter, o.keepRedactable, cp)
}

var debugIntentCount = &cobra.Command{
	Use:   "intent-count <store directory>",
	Short: "return a count of intents in directory",
	Long: `
Returns a count of interleaved and separated intents in the store directory.
Used to investigate stores with lots of unresolved intents, or to confirm
if the migration away from interleaved intents was successful.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDebugIntentCount,
}

func runDebugIntentCount(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30906)
	stopper := stop.NewStopper()
	ctx := context.Background()
	defer stopper.Stop(ctx)

	db, err := OpenExistingStore(args[0], stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(30911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30912)
	}
	__antithesis_instrumentation__.Notify(30907)
	defer db.Close()

	var interleavedIntentCount, separatedIntentCount int
	var keysCount uint64
	var wg sync.WaitGroup
	closer := make(chan bool)

	wg.Add(1)
	_ = stopper.RunAsyncTask(ctx, "intent-count-progress-indicator", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(30913)
		defer wg.Done()
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(30914)
			fmt.Printf("scanned %d keys\n", atomic.LoadUint64(&keysCount))
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(30915)
			return
		case <-closer:
			__antithesis_instrumentation__.Notify(30916)
			return
		}
	})
	__antithesis_instrumentation__.Notify(30908)

	iter := db.NewEngineIterator(storage.IterOptions{
		LowerBound: roachpb.KeyMin,
		UpperBound: roachpb.KeyMax,
	})
	defer iter.Close()
	valid, err := iter.SeekEngineKeyGE(storage.EngineKey{Key: roachpb.KeyMin})
	var meta enginepb.MVCCMetadata
	for ; valid && func() bool {
		__antithesis_instrumentation__.Notify(30917)
		return err == nil == true
	}() == true; valid, err = iter.NextEngineKey() {
		__antithesis_instrumentation__.Notify(30918)
		key, err := iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(30926)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30927)
		}
		__antithesis_instrumentation__.Notify(30919)
		atomic.AddUint64(&keysCount, 1)
		if key.IsLockTableKey() {
			__antithesis_instrumentation__.Notify(30928)
			separatedIntentCount++
			continue
		} else {
			__antithesis_instrumentation__.Notify(30929)
		}
		__antithesis_instrumentation__.Notify(30920)
		if !key.IsMVCCKey() {
			__antithesis_instrumentation__.Notify(30930)
			continue
		} else {
			__antithesis_instrumentation__.Notify(30931)
		}
		__antithesis_instrumentation__.Notify(30921)
		mvccKey, err := key.ToMVCCKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(30932)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30933)
		}
		__antithesis_instrumentation__.Notify(30922)
		if !mvccKey.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(30934)
			continue
		} else {
			__antithesis_instrumentation__.Notify(30935)
		}
		__antithesis_instrumentation__.Notify(30923)
		val := iter.UnsafeValue()
		if err := protoutil.Unmarshal(val, &meta); err != nil {
			__antithesis_instrumentation__.Notify(30936)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30937)
		}
		__antithesis_instrumentation__.Notify(30924)
		if meta.IsInline() {
			__antithesis_instrumentation__.Notify(30938)
			continue
		} else {
			__antithesis_instrumentation__.Notify(30939)
		}
		__antithesis_instrumentation__.Notify(30925)
		interleavedIntentCount++
	}
	__antithesis_instrumentation__.Notify(30909)
	if err != nil {
		__antithesis_instrumentation__.Notify(30940)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30941)
	}
	__antithesis_instrumentation__.Notify(30910)
	close(closer)
	wg.Wait()
	fmt.Printf("interleaved intents: %d\nseparated intents: %d\n",
		interleavedIntentCount, separatedIntentCount)
	return nil
}

var DebugCommandsRequiringEncryption = []*cobra.Command{
	debugCheckStoreCmd,
	debugCompactCmd,
	debugGCCmd,
	debugIntentCount,
	debugKeysCmd,
	debugRaftLogCmd,
	debugRangeDataCmd,
	debugRangeDescriptorsCmd,
	debugUnsafeRemoveDeadReplicasCmd,
	debugRecoverCollectInfoCmd,
	debugRecoverExecuteCmd,
}

var debugCmds = []*cobra.Command{
	debugCheckStoreCmd,
	debugCompactCmd,
	debugGCCmd,
	debugIntentCount,
	debugKeysCmd,
	debugRaftLogCmd,
	debugRangeDataCmd,
	debugRangeDescriptorsCmd,
	debugUnsafeRemoveDeadReplicasCmd,
	debugBallastCmd,
	debugCheckLogConfigCmd,
	debugDecodeKeyCmd,
	debugDecodeValueCmd,
	debugDecodeProtoCmd,
	debugGossipValuesCmd,
	debugTimeSeriesDumpCmd,
	debugSyncBenchCmd,
	debugSyncTestCmd,
	debugEnvCmd,
	debugZipCmd,
	debugMergeLogsCmd,
	debugListFilesCmd,
	debugResetQuorumCmd,
	debugSendKVBatchCmd,
	debugRecoverCmd,
}

var DebugCmd = &cobra.Command{
	Use:   "debug [command]",
	Short: "debugging commands",
	Long: `Various commands for debugging.

These commands are useful for extracting data from the data files of a
process that has failed and cannot restart.
`,
	RunE: UsageAndErr,
}

type mvccValueFormatter struct {
	kv  storage.MVCCKeyValue
	err error
}

func (m mvccValueFormatter) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(30942)
	if m.err != nil {
		__antithesis_instrumentation__.Notify(30944)
		errors.FormatError(m.err, f, c)
		return
	} else {
		__antithesis_instrumentation__.Notify(30945)
	}
	__antithesis_instrumentation__.Notify(30943)
	fmt.Fprint(f, kvserver.SprintMVCCKeyValue(m.kv, false))
}

type lockValueFormatter struct {
	value []byte
}

func (m lockValueFormatter) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(30946)
	fmt.Fprint(f, kvserver.SprintIntent(m.value))
}

var pebbleToolFS = &swappableFS{vfs.Default}

func init() {
	DebugCmd.AddCommand(debugCmds...)

	storage.EngineComparer.FormatValue = func(key, value []byte) fmt.Formatter {
		decoded, ok := storage.DecodeEngineKey(key)
		if !ok {
			return mvccValueFormatter{err: errors.Errorf("invalid encoded engine key: %x", key)}
		}
		if decoded.IsMVCCKey() {
			mvccKey, err := decoded.ToMVCCKey()
			if err != nil {
				return mvccValueFormatter{err: err}
			}
			return mvccValueFormatter{kv: storage.MVCCKeyValue{Key: mvccKey, Value: value}}
		}
		return lockValueFormatter{value: value}
	}

	pebbleTool := tool.New(tool.Mergers(storage.MVCCMerger),
		tool.DefaultComparer(storage.EngineComparer),
		tool.FS(&absoluteFS{pebbleToolFS}),
	)
	DebugPebbleCmd.AddCommand(pebbleTool.Commands...)
	initPebbleCmds(DebugPebbleCmd)
	DebugCmd.AddCommand(DebugPebbleCmd)

	doctorExamineCmd.AddCommand(doctorExamineClusterCmd, doctorExamineZipDirCmd)
	doctorRecreateCmd.AddCommand(doctorRecreateClusterCmd, doctorRecreateZipDirCmd)
	debugDoctorCmd.AddCommand(doctorExamineCmd, doctorRecreateCmd, doctorExamineFallbackClusterCmd, doctorExamineFallbackZipDirCmd)
	DebugCmd.AddCommand(debugDoctorCmd)

	debugStatementBundleCmd.AddCommand(statementBundleRecreateCmd)
	DebugCmd.AddCommand(debugStatementBundleCmd)

	DebugCmd.AddCommand(debugJobTraceFromClusterCmd)

	f := debugSyncBenchCmd.Flags()
	f.IntVarP(&syncBenchOpts.Concurrency, "concurrency", "c", syncBenchOpts.Concurrency,
		"number of concurrent writers")
	f.DurationVarP(&syncBenchOpts.Duration, "duration", "d", syncBenchOpts.Duration,
		"duration to run the test for")
	f.BoolVarP(&syncBenchOpts.LogOnly, "log-only", "l", syncBenchOpts.LogOnly,
		"only write to the WAL, not to sstables")

	f = debugUnsafeRemoveDeadReplicasCmd.Flags()
	f.IntSliceVar(&removeDeadReplicasOpts.deadStoreIDs, "dead-store-ids", nil,
		"list of dead store IDs")

	f = debugRecoverCollectInfoCmd.Flags()
	f.VarP(&debugRecoverCollectInfoOpts.Stores, cliflags.RecoverStore.Name, cliflags.RecoverStore.Shorthand, cliflags.RecoverStore.Usage())

	f = debugRecoverPlanCmd.Flags()
	f.StringVarP(&debugRecoverPlanOpts.outputFileName, "plan", "o", "",
		"filename to write plan to")
	f.IntSliceVar(&debugRecoverPlanOpts.deadStoreIDs, "dead-store-ids", nil,
		"list of dead store IDs")
	f.VarP(&debugRecoverPlanOpts.confirmAction, cliflags.ConfirmActions.Name, cliflags.ConfirmActions.Shorthand,
		cliflags.ConfirmActions.Usage())
	f.BoolVar(&debugRecoverPlanOpts.force, "force", false,
		"force creation of plan even when problems were encountered; applying this plan may "+
			"result in additional problems and should be done only with care and as a last resort")

	f = debugRecoverExecuteCmd.Flags()
	f.VarP(&debugRecoverExecuteOpts.Stores, cliflags.RecoverStore.Name, cliflags.RecoverStore.Shorthand, cliflags.RecoverStore.Usage())
	f.VarP(&debugRecoverExecuteOpts.confirmAction, cliflags.ConfirmActions.Name, cliflags.ConfirmActions.Shorthand,
		cliflags.ConfirmActions.Usage())

	f = debugMergeLogsCmd.Flags()
	f.Var(flagutil.Time(&debugMergeLogsOpts.from), "from",
		"time before which messages should be filtered")

	f.Var(flagutil.Time(&debugMergeLogsOpts.to), "to",
		"time after which messages should be filtered")
	f.Var(flagutil.Regexp(&debugMergeLogsOpts.filter), "filter",
		"re which filters log messages")
	f.Var(flagutil.Regexp(&debugMergeLogsOpts.file), "file-pattern",
		"re which filters log files based on path, also used with prefix and program-filter")
	f.Var(flagutil.Regexp(&debugMergeLogsOpts.program), "program-filter",
		"re which filter log files that operates on the capture group named \"program\" in file-pattern, "+
			"if no such group exists, program-filter is ignored")
	f.StringVar(&debugMergeLogsOpts.prefix, "prefix", "${host}> ",
		"expansion template (see regexp.Expand) used as prefix to merged log messages evaluated on file-pattern")
	f.BoolVar(&debugMergeLogsOpts.keepRedactable, "redactable-output", debugMergeLogsOpts.keepRedactable,
		"keep the output log file redactable")
	f.BoolVar(&debugMergeLogsOpts.redactInput, "redact", debugMergeLogsOpts.redactInput,
		"redact the input files to remove sensitive information")
	f.StringVar(&debugMergeLogsOpts.format, "format", "",
		"log format of the input files")
	f.Var(&debugMergeLogsOpts.useColor, "color",
		"force use of TTY escape codes to colorize the output")

	f = debugDecodeProtoCmd.Flags()
	f.StringVar(&debugDecodeProtoName, "schema", "cockroach.sql.sqlbase.Descriptor",
		"fully qualified name of the proto to decode")
	f.BoolVar(&debugDecodeProtoEmitDefaults, "emit-defaults", false,
		"encode default values for every field")

	f = debugCheckLogConfigCmd.Flags()
	f.Var(&debugLogChanSel, "only-channels", "selection of channels to include in the output diagram.")

	f = debugTimeSeriesDumpCmd.Flags()
	f.Var(&debugTimeSeriesDumpOpts.format, "format", "output format (text, csv, tsv, raw)")
	f.Var(&debugTimeSeriesDumpOpts.from, "from", "oldest timestamp to include (inclusive)")
	f.Var(&debugTimeSeriesDumpOpts.to, "to", "newest timestamp to include (inclusive)")

	f = debugSendKVBatchCmd.Flags()
	f.StringVar(&debugSendKVBatchContext.traceFormat, "trace", debugSendKVBatchContext.traceFormat,
		"which format to use for the trace output (off, text, jaeger)")
	f.BoolVar(&debugSendKVBatchContext.keepCollectedSpans, "keep-collected-spans", debugSendKVBatchContext.keepCollectedSpans,
		"whether to keep the CollectedSpans field on the response, to learn about how traces work")
	f.StringVar(&debugSendKVBatchContext.traceFile, "trace-output", debugSendKVBatchContext.traceFile,
		"the output file to use for the trace. If left empty, output to stderr.")
}

func initPebbleCmds(cmd *cobra.Command) {
	__antithesis_instrumentation__.Notify(30947)
	for _, c := range cmd.Commands() {
		__antithesis_instrumentation__.Notify(30948)
		wrapped := c.PreRunE
		c.PreRunE = func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(30950)
			if wrapped != nil {
				__antithesis_instrumentation__.Notify(30952)
				if err := wrapped(cmd, args); err != nil {
					__antithesis_instrumentation__.Notify(30953)
					return err
				} else {
					__antithesis_instrumentation__.Notify(30954)
				}
			} else {
				__antithesis_instrumentation__.Notify(30955)
			}
			__antithesis_instrumentation__.Notify(30951)
			return pebbleCryptoInitializer()
		}
		__antithesis_instrumentation__.Notify(30949)
		initPebbleCmds(c)
	}
}

func pebbleCryptoInitializer() error {
	__antithesis_instrumentation__.Notify(30956)
	storageConfig := base.StorageConfig{
		Settings: serverCfg.Settings,
		Dir:      serverCfg.Stores.Specs[0].Path,
	}

	if PopulateStorageConfigHook != nil {
		__antithesis_instrumentation__.Notify(30959)
		if err := PopulateStorageConfigHook(&storageConfig); err != nil {
			__antithesis_instrumentation__.Notify(30960)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30961)
		}
	} else {
		__antithesis_instrumentation__.Notify(30962)
	}
	__antithesis_instrumentation__.Notify(30957)

	cfg := storage.PebbleConfig{
		StorageConfig: storageConfig,
		Opts:          storage.DefaultPebbleOptions(),
	}

	_, _, err := storage.ResolveEncryptedEnvOptions(&cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(30963)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30964)
	}
	__antithesis_instrumentation__.Notify(30958)

	pebbleToolFS.set(cfg.Opts.FS)
	return nil
}
