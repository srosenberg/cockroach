package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"golang.org/x/sync/errgroup"
)

var debugCheckStoreCmd = &cobra.Command{
	Use:   "check-store <directory>",
	Short: "consistency check for a single store",
	Long: `
Perform local consistency checks of a single store.

Capable of detecting the following errors:
* Raft logs that are inconsistent with their metadata
* MVCC stats that are inconsistent with the data within the range
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugCheckStoreCmd),
}

var errCheckFoundProblem = errors.New("check-store found problems")

func runDebugCheckStoreCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(30965)
	ctx := context.Background()
	dir := args[0]
	foundProblem := false

	err := checkStoreRangeStats(ctx, dir, func(args ...interface{}) {
		__antithesis_instrumentation__.Notify(30971)
		fmt.Println(args...)
	})
	__antithesis_instrumentation__.Notify(30966)
	foundProblem = foundProblem || func() bool {
		__antithesis_instrumentation__.Notify(30972)
		return err != nil == true
	}() == true
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(30973)
		return !errors.Is(err, errCheckFoundProblem) == true
	}() == true {
		__antithesis_instrumentation__.Notify(30974)
		_, _ = fmt.Println(err)
	} else {
		__antithesis_instrumentation__.Notify(30975)
	}
	__antithesis_instrumentation__.Notify(30967)

	err = checkStoreRaftState(ctx, dir, func(format string, args ...interface{}) {
		__antithesis_instrumentation__.Notify(30976)
		_, _ = fmt.Printf(format, args...)
	})
	__antithesis_instrumentation__.Notify(30968)
	foundProblem = foundProblem || func() bool {
		__antithesis_instrumentation__.Notify(30977)
		return err != nil == true
	}() == true
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(30978)
		return !errors.Is(err, errCheckFoundProblem) == true
	}() == true {
		__antithesis_instrumentation__.Notify(30979)
		fmt.Println(err)
	} else {
		__antithesis_instrumentation__.Notify(30980)
	}
	__antithesis_instrumentation__.Notify(30969)
	if foundProblem {
		__antithesis_instrumentation__.Notify(30981)
		return errCheckFoundProblem
	} else {
		__antithesis_instrumentation__.Notify(30982)
	}
	__antithesis_instrumentation__.Notify(30970)
	return nil
}

type replicaCheckInfo struct {
	truncatedIndex uint64
	appliedIndex   uint64
	firstIndex     uint64
	lastIndex      uint64
	committedIndex uint64
}

type checkInput struct {
	eng  storage.Engine
	desc *roachpb.RangeDescriptor
	sl   stateloader.StateLoader
}

type checkResult struct {
	desc           *roachpb.RangeDescriptor
	err            error
	claimMS, actMS enginepb.MVCCStats
}

func (cr *checkResult) Error() error {
	__antithesis_instrumentation__.Notify(30983)
	var err error
	if cr.err != nil {
		__antithesis_instrumentation__.Notify(30987)
		err = cr.err
	} else {
		__antithesis_instrumentation__.Notify(30988)
	}
	__antithesis_instrumentation__.Notify(30984)
	if !cr.actMS.Equal(enginepb.MVCCStats{}) && func() bool {
		__antithesis_instrumentation__.Notify(30989)
		return !cr.actMS.Equal(cr.claimMS) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(30990)
		return cr.claimMS.ContainsEstimates <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(30991)
		thisErr := errors.Newf(
			"stats inconsistency:\n- stored:\n%+v\n- recomputed:\n%+v\n- diff:\n%s",
			cr.claimMS, cr.actMS, strings.Join(pretty.Diff(cr.claimMS, cr.actMS), ","),
		)
		err = errors.CombineErrors(err, thisErr)
	} else {
		__antithesis_instrumentation__.Notify(30992)
	}
	__antithesis_instrumentation__.Notify(30985)
	if err != nil {
		__antithesis_instrumentation__.Notify(30993)
		if cr.desc != nil {
			__antithesis_instrumentation__.Notify(30994)
			err = errors.Wrapf(err, "%s", cr.desc)
		} else {
			__antithesis_instrumentation__.Notify(30995)
		}
	} else {
		__antithesis_instrumentation__.Notify(30996)
	}
	__antithesis_instrumentation__.Notify(30986)
	return err
}

func worker(ctx context.Context, in checkInput) checkResult {
	__antithesis_instrumentation__.Notify(30997)
	desc, eng := in.desc, in.eng

	res := checkResult{desc: desc}
	claimedMS, err := in.sl.LoadMVCCStats(ctx, eng)
	if err != nil {
		__antithesis_instrumentation__.Notify(31000)
		res.err = err
		return res
	} else {
		__antithesis_instrumentation__.Notify(31001)
	}
	__antithesis_instrumentation__.Notify(30998)
	ms, err := rditer.ComputeStatsForRange(desc, eng, claimedMS.LastUpdateNanos)
	if err != nil {
		__antithesis_instrumentation__.Notify(31002)
		res.err = err
		return res
	} else {
		__antithesis_instrumentation__.Notify(31003)
	}
	__antithesis_instrumentation__.Notify(30999)
	res.claimMS = claimedMS
	res.actMS = ms
	return res
}

func checkStoreRangeStats(
	ctx context.Context,
	dir string,
	println func(...interface{}),
) error {
	__antithesis_instrumentation__.Notify(31004)
	stopper := stop.NewStopper()
	defer stopper.Stop(ctx)

	eng, err := OpenExistingStore(dir, stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(31010)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31011)
	}
	__antithesis_instrumentation__.Notify(31005)

	inCh := make(chan checkInput)
	outCh := make(chan checkResult, 1000)

	n := runtime.GOMAXPROCS(0)
	var g errgroup.Group
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(31012)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(31013)
			for in := range inCh {
				__antithesis_instrumentation__.Notify(31015)
				outCh <- worker(ctx, in)
			}
			__antithesis_instrumentation__.Notify(31014)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(31006)

	go func() {
		__antithesis_instrumentation__.Notify(31016)
		if err := kvserver.IterateRangeDescriptorsFromDisk(ctx, eng,
			func(desc roachpb.RangeDescriptor) error {
				__antithesis_instrumentation__.Notify(31019)
				inCh <- checkInput{eng: eng, desc: &desc, sl: stateloader.Make(desc.RangeID)}
				return nil
			}); err != nil {
			__antithesis_instrumentation__.Notify(31020)
			outCh <- checkResult{err: err}
		} else {
			__antithesis_instrumentation__.Notify(31021)
		}
		__antithesis_instrumentation__.Notify(31017)
		close(inCh)
		if err := g.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(31022)
			outCh <- checkResult{err: err}
		} else {
			__antithesis_instrumentation__.Notify(31023)
		}
		__antithesis_instrumentation__.Notify(31018)
		close(outCh)
	}()
	__antithesis_instrumentation__.Notify(31007)

	foundProblem := false
	var total enginepb.MVCCStats
	var cR, cE int
	for res := range outCh {
		__antithesis_instrumentation__.Notify(31024)
		cR++
		if err := res.Error(); err != nil {
			__antithesis_instrumentation__.Notify(31025)
			foundProblem = true
			errS := err.Error()
			println(errS)
		} else {
			__antithesis_instrumentation__.Notify(31026)
			if res.claimMS.ContainsEstimates > 0 {
				__antithesis_instrumentation__.Notify(31028)
				cE++
			} else {
				__antithesis_instrumentation__.Notify(31029)
			}
			__antithesis_instrumentation__.Notify(31027)
			total.Add(res.actMS)
		}
	}
	__antithesis_instrumentation__.Notify(31008)

	println(fmt.Sprintf("scanned %d ranges (%d with estimates), total stats %s", cR, cE, &total))

	if foundProblem {
		__antithesis_instrumentation__.Notify(31030)

		return errCheckFoundProblem
	} else {
		__antithesis_instrumentation__.Notify(31031)
	}
	__antithesis_instrumentation__.Notify(31009)
	return nil
}

func checkStoreRaftState(
	ctx context.Context,
	dir string,
	printf func(string, ...interface{}),
) error {
	__antithesis_instrumentation__.Notify(31032)
	foundProblem := false
	goldenPrintf := printf
	printf = func(format string, args ...interface{}) {
		__antithesis_instrumentation__.Notify(31039)
		foundProblem = true
		goldenPrintf(format, args...)
	}
	__antithesis_instrumentation__.Notify(31033)
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())

	db, err := OpenExistingStore(dir, stopper, true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(31040)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31041)
	}
	__antithesis_instrumentation__.Notify(31034)

	start := roachpb.Key(keys.LocalRangeIDPrefix)
	end := start.PrefixEnd()

	replicaInfo := map[roachpb.RangeID]*replicaCheckInfo{}
	getReplicaInfo := func(rangeID roachpb.RangeID) *replicaCheckInfo {
		__antithesis_instrumentation__.Notify(31042)
		if info, ok := replicaInfo[rangeID]; ok {
			__antithesis_instrumentation__.Notify(31044)
			return info
		} else {
			__antithesis_instrumentation__.Notify(31045)
		}
		__antithesis_instrumentation__.Notify(31043)
		replicaInfo[rangeID] = &replicaCheckInfo{}
		return replicaInfo[rangeID]
	}
	__antithesis_instrumentation__.Notify(31035)

	if _, err := storage.MVCCIterate(ctx, db, start, end, hlc.MaxTimestamp,
		storage.MVCCScanOptions{Inconsistent: true}, func(kv roachpb.KeyValue) error {
			__antithesis_instrumentation__.Notify(31046)
			rangeID, _, suffix, detail, err := keys.DecodeRangeIDKey(kv.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(31049)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31050)
			}
			__antithesis_instrumentation__.Notify(31047)

			switch {
			case bytes.Equal(suffix, keys.LocalRaftHardStateSuffix):
				__antithesis_instrumentation__.Notify(31051)
				var hs raftpb.HardState
				if err := kv.Value.GetProto(&hs); err != nil {
					__antithesis_instrumentation__.Notify(31060)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31061)
				}
				__antithesis_instrumentation__.Notify(31052)
				getReplicaInfo(rangeID).committedIndex = hs.Commit
			case bytes.Equal(suffix, keys.LocalRaftTruncatedStateSuffix):
				__antithesis_instrumentation__.Notify(31053)
				var trunc roachpb.RaftTruncatedState
				if err := kv.Value.GetProto(&trunc); err != nil {
					__antithesis_instrumentation__.Notify(31062)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31063)
				}
				__antithesis_instrumentation__.Notify(31054)
				getReplicaInfo(rangeID).truncatedIndex = trunc.Index
			case bytes.Equal(suffix, keys.LocalRangeAppliedStateSuffix):
				__antithesis_instrumentation__.Notify(31055)
				var state enginepb.RangeAppliedState
				if err := kv.Value.GetProto(&state); err != nil {
					__antithesis_instrumentation__.Notify(31064)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31065)
				}
				__antithesis_instrumentation__.Notify(31056)
				getReplicaInfo(rangeID).appliedIndex = state.RaftAppliedIndex
			case bytes.Equal(suffix, keys.LocalRaftLogSuffix):
				__antithesis_instrumentation__.Notify(31057)
				_, index, err := encoding.DecodeUint64Ascending(detail)
				if err != nil {
					__antithesis_instrumentation__.Notify(31066)
					return err
				} else {
					__antithesis_instrumentation__.Notify(31067)
				}
				__antithesis_instrumentation__.Notify(31058)
				ri := getReplicaInfo(rangeID)
				if ri.firstIndex == 0 {
					__antithesis_instrumentation__.Notify(31068)
					ri.firstIndex = index
					ri.lastIndex = index
				} else {
					__antithesis_instrumentation__.Notify(31069)
					if index != ri.lastIndex+1 {
						__antithesis_instrumentation__.Notify(31071)
						printf("range %s: log index anomaly: %v followed by %v\n",
							rangeID, ri.lastIndex, index)
					} else {
						__antithesis_instrumentation__.Notify(31072)
					}
					__antithesis_instrumentation__.Notify(31070)
					ri.lastIndex = index
				}
			default:
				__antithesis_instrumentation__.Notify(31059)
			}
			__antithesis_instrumentation__.Notify(31048)

			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(31073)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31074)
	}
	__antithesis_instrumentation__.Notify(31036)

	for rangeID, info := range replicaInfo {
		__antithesis_instrumentation__.Notify(31075)
		if info.truncatedIndex != 0 && func() bool {
			__antithesis_instrumentation__.Notify(31080)
			return info.truncatedIndex != info.firstIndex-1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(31081)
			printf("range %s: truncated index %v should equal first index %v - 1\n",
				rangeID, info.truncatedIndex, info.firstIndex)
		} else {
			__antithesis_instrumentation__.Notify(31082)
		}
		__antithesis_instrumentation__.Notify(31076)
		if info.firstIndex > info.lastIndex {
			__antithesis_instrumentation__.Notify(31083)
			printf("range %s: [first index, last index] is [%d, %d]\n",
				rangeID, info.firstIndex, info.lastIndex)
		} else {
			__antithesis_instrumentation__.Notify(31084)
		}
		__antithesis_instrumentation__.Notify(31077)
		if info.appliedIndex < info.firstIndex || func() bool {
			__antithesis_instrumentation__.Notify(31085)
			return info.appliedIndex > info.lastIndex == true
		}() == true {
			__antithesis_instrumentation__.Notify(31086)
			printf("range %s: applied index %v should be between first index %v and last index %v\n",
				rangeID, info.appliedIndex, info.firstIndex, info.lastIndex)
		} else {
			__antithesis_instrumentation__.Notify(31087)
		}
		__antithesis_instrumentation__.Notify(31078)
		if info.appliedIndex > info.committedIndex {
			__antithesis_instrumentation__.Notify(31088)
			printf("range %s: committed index %d must not trail applied index %d\n",
				rangeID, info.committedIndex, info.appliedIndex)
		} else {
			__antithesis_instrumentation__.Notify(31089)
		}
		__antithesis_instrumentation__.Notify(31079)
		if info.committedIndex > info.lastIndex {
			__antithesis_instrumentation__.Notify(31090)
			printf("range %s: committed index %d ahead of last index  %d\n",
				rangeID, info.committedIndex, info.lastIndex)
		} else {
			__antithesis_instrumentation__.Notify(31091)
		}
	}
	__antithesis_instrumentation__.Notify(31037)
	if foundProblem {
		__antithesis_instrumentation__.Notify(31092)
		return errCheckFoundProblem
	} else {
		__antithesis_instrumentation__.Notify(31093)
	}
	__antithesis_instrumentation__.Notify(31038)

	return nil
}
