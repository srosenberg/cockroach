package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/tracker"
)

var looselyCoupledTruncationEnabled = func() *settings.BoolSetting {
	__antithesis_instrumentation__.Notify(112708)
	s := settings.RegisterBoolSetting(
		settings.SystemOnly,
		"kv.raft_log.loosely_coupled_truncation.enabled",
		"set to true to loosely couple the raft log truncation",
		false)
	s.SetVisibility(settings.Reserved)
	return s
}()

const (
	raftLogQueueTimerDuration = 0

	RaftLogQueueStaleThreshold = 100

	RaftLogQueueStaleSize = 64 << 10

	raftLogQueueConcurrency = 4

	RaftLogQueuePendingSnapshotGracePeriod = 3 * time.Second
)

type raftLogQueue struct {
	*baseQueue
	db *kv.DB

	logSnapshots util.EveryN
}

func newRaftLogQueue(store *Store, db *kv.DB) *raftLogQueue {
	__antithesis_instrumentation__.Notify(112709)
	rlq := &raftLogQueue{
		db:           db,
		logSnapshots: util.Every(10 * time.Second),
	}
	rlq.baseQueue = newBaseQueue(
		"raftlog", rlq, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			maxConcurrency:       raftLogQueueConcurrency,
			needsLease:           false,
			needsSystemConfig:    false,
			acceptsUnsplitRanges: true,
			successes:            store.metrics.RaftLogQueueSuccesses,
			failures:             store.metrics.RaftLogQueueFailures,
			pending:              store.metrics.RaftLogQueuePending,
			processingNanos:      store.metrics.RaftLogQueueProcessingNanos,
		},
	)
	return rlq
}

func newTruncateDecision(ctx context.Context, r *Replica) (truncateDecision, error) {
	__antithesis_instrumentation__.Notify(112710)
	rangeID := r.RangeID
	now := timeutil.Now()

	r.mu.Lock()
	raftLogSize := r.pendingLogTruncations.computePostTruncLogSize(r.mu.raftLogSize)

	targetSize := r.store.cfg.RaftLogTruncationThreshold
	if targetSize > r.mu.conf.RangeMaxBytes {
		__antithesis_instrumentation__.Notify(112717)
		targetSize = r.mu.conf.RangeMaxBytes
	} else {
		__antithesis_instrumentation__.Notify(112718)
	}
	__antithesis_instrumentation__.Notify(112711)
	raftStatus := r.raftStatusRLocked()

	const anyRecipientStore roachpb.StoreID = 0
	pendingSnapshotIndex := r.getAndGCSnapshotLogTruncationConstraintsLocked(now, anyRecipientStore)
	lastIndex := r.mu.lastIndex

	logSizeTrusted := r.mu.raftLogSizeTrusted
	firstIndex, err := r.raftFirstIndexLocked()
	r.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(112719)
		return truncateDecision{}, errors.Wrapf(err, "error retrieving first index for r%d", rangeID)
	} else {
		__antithesis_instrumentation__.Notify(112720)
	}
	__antithesis_instrumentation__.Notify(112712)
	firstIndex = r.pendingLogTruncations.computePostTruncFirstIndex(firstIndex)

	if raftStatus == nil {
		__antithesis_instrumentation__.Notify(112721)
		if log.V(6) {
			__antithesis_instrumentation__.Notify(112723)
			log.Infof(ctx, "the raft group doesn't exist for r%d", rangeID)
		} else {
			__antithesis_instrumentation__.Notify(112724)
		}
		__antithesis_instrumentation__.Notify(112722)
		return truncateDecision{}, nil
	} else {
		__antithesis_instrumentation__.Notify(112725)
	}
	__antithesis_instrumentation__.Notify(112713)

	if raftStatus.RaftState != raft.StateLeader {
		__antithesis_instrumentation__.Notify(112726)
		return truncateDecision{}, nil
	} else {
		__antithesis_instrumentation__.Notify(112727)
	}
	__antithesis_instrumentation__.Notify(112714)

	r.mu.RLock()
	log.Eventf(ctx, "raft status before lastUpdateTimes check: %+v", raftStatus.Progress)
	log.Eventf(ctx, "lastUpdateTimes: %+v", r.mu.lastUpdateTimes)
	updateRaftProgressFromActivity(
		ctx, raftStatus.Progress, r.descRLocked().Replicas().Descriptors(),
		func(replicaID roachpb.ReplicaID) bool {
			__antithesis_instrumentation__.Notify(112728)
			return r.mu.lastUpdateTimes.isFollowerActiveSince(
				ctx, replicaID, now, r.store.cfg.RangeLeaseActiveDuration())
		},
	)
	__antithesis_instrumentation__.Notify(112715)
	log.Eventf(ctx, "raft status after lastUpdateTimes check: %+v", raftStatus.Progress)
	r.mu.RUnlock()

	if pr, ok := raftStatus.Progress[raftStatus.Lead]; ok {
		__antithesis_instrumentation__.Notify(112729)

		pr.State = tracker.StateReplicate
		raftStatus.Progress[raftStatus.Lead] = pr
	} else {
		__antithesis_instrumentation__.Notify(112730)
	}
	__antithesis_instrumentation__.Notify(112716)

	input := truncateDecisionInput{
		RaftStatus:           *raftStatus,
		LogSize:              raftLogSize,
		MaxLogSize:           targetSize,
		LogSizeTrusted:       logSizeTrusted,
		FirstIndex:           firstIndex,
		LastIndex:            lastIndex,
		PendingSnapshotIndex: pendingSnapshotIndex,
	}

	decision := computeTruncateDecision(input)
	return decision, nil
}

func updateRaftProgressFromActivity(
	ctx context.Context,
	prs map[uint64]tracker.Progress,
	replicas []roachpb.ReplicaDescriptor,
	replicaActive func(roachpb.ReplicaID) bool,
) {
	__antithesis_instrumentation__.Notify(112731)
	for _, replDesc := range replicas {
		__antithesis_instrumentation__.Notify(112732)
		replicaID := replDesc.ReplicaID
		pr, ok := prs[uint64(replicaID)]
		if !ok {
			__antithesis_instrumentation__.Notify(112734)
			continue
		} else {
			__antithesis_instrumentation__.Notify(112735)
		}
		__antithesis_instrumentation__.Notify(112733)
		pr.RecentActive = replicaActive(replicaID)

		pr.PendingSnapshot = 0
		prs[uint64(replicaID)] = pr
	}
}

const (
	truncatableIndexChosenViaCommitIndex     = "commit"
	truncatableIndexChosenViaFollowers       = "followers"
	truncatableIndexChosenViaProbingFollower = "probing follower"
	truncatableIndexChosenViaPendingSnap     = "pending snapshot"
	truncatableIndexChosenViaFirstIndex      = "first index"
	truncatableIndexChosenViaLastIndex       = "last index"
)

type truncateDecisionInput struct {
	RaftStatus            raft.Status
	LogSize, MaxLogSize   int64
	LogSizeTrusted        bool
	FirstIndex, LastIndex uint64
	PendingSnapshotIndex  uint64
}

func (input truncateDecisionInput) LogTooLarge() bool {
	__antithesis_instrumentation__.Notify(112736)
	return input.LogSize > input.MaxLogSize
}

type truncateDecision struct {
	Input       truncateDecisionInput
	CommitIndex uint64

	NewFirstIndex uint64
	ChosenVia     string
}

func (td *truncateDecision) raftSnapshotsForIndex(index uint64) int {
	__antithesis_instrumentation__.Notify(112737)
	var n int
	for _, p := range td.Input.RaftStatus.Progress {
		__antithesis_instrumentation__.Notify(112740)
		if p.State != tracker.StateReplicate {
			__antithesis_instrumentation__.Notify(112742)

			_ = truncatableIndexChosenViaProbingFollower
			continue
		} else {
			__antithesis_instrumentation__.Notify(112743)
		}
		__antithesis_instrumentation__.Notify(112741)

		if p.Match < index && func() bool {
			__antithesis_instrumentation__.Notify(112744)
			return p.Next <= index == true
		}() == true {
			__antithesis_instrumentation__.Notify(112745)
			n++
		} else {
			__antithesis_instrumentation__.Notify(112746)
		}
	}
	__antithesis_instrumentation__.Notify(112738)
	if td.Input.PendingSnapshotIndex != 0 && func() bool {
		__antithesis_instrumentation__.Notify(112747)
		return td.Input.PendingSnapshotIndex < index == true
	}() == true {
		__antithesis_instrumentation__.Notify(112748)
		n++
	} else {
		__antithesis_instrumentation__.Notify(112749)
	}
	__antithesis_instrumentation__.Notify(112739)

	return n
}

func (td *truncateDecision) NumNewRaftSnapshots() int {
	__antithesis_instrumentation__.Notify(112750)
	return td.raftSnapshotsForIndex(td.NewFirstIndex) - td.raftSnapshotsForIndex(td.Input.FirstIndex)
}

func (td *truncateDecision) String() string {
	__antithesis_instrumentation__.Notify(112751)
	var buf strings.Builder
	_, _ = fmt.Fprintf(&buf, "should truncate: %t [", td.ShouldTruncate())
	_, _ = fmt.Fprintf(
		&buf,
		"truncate %d entries to first index %d (chosen via: %s)",
		td.NumTruncatableIndexes(), td.NewFirstIndex, td.ChosenVia,
	)
	if td.Input.LogTooLarge() {
		__antithesis_instrumentation__.Notify(112755)
		_, _ = fmt.Fprintf(
			&buf,
			"; log too large (%s > %s)",
			humanizeutil.IBytes(td.Input.LogSize),
			humanizeutil.IBytes(td.Input.MaxLogSize),
		)
	} else {
		__antithesis_instrumentation__.Notify(112756)
	}
	__antithesis_instrumentation__.Notify(112752)
	if n := td.NumNewRaftSnapshots(); n > 0 {
		__antithesis_instrumentation__.Notify(112757)
		_, _ = fmt.Fprintf(&buf, "; implies %d Raft snapshot%s", n, util.Pluralize(int64(n)))
	} else {
		__antithesis_instrumentation__.Notify(112758)
	}
	__antithesis_instrumentation__.Notify(112753)
	if !td.Input.LogSizeTrusted {
		__antithesis_instrumentation__.Notify(112759)
		_, _ = fmt.Fprintf(&buf, "; log size untrusted")
	} else {
		__antithesis_instrumentation__.Notify(112760)
	}
	__antithesis_instrumentation__.Notify(112754)
	buf.WriteRune(']')

	return buf.String()
}

func (td *truncateDecision) NumTruncatableIndexes() int {
	__antithesis_instrumentation__.Notify(112761)
	if td.NewFirstIndex < td.Input.FirstIndex {
		__antithesis_instrumentation__.Notify(112763)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(112764)
	}
	__antithesis_instrumentation__.Notify(112762)
	return int(td.NewFirstIndex - td.Input.FirstIndex)
}

func (td *truncateDecision) ShouldTruncate() bool {
	__antithesis_instrumentation__.Notify(112765)
	n := td.NumTruncatableIndexes()
	return n >= RaftLogQueueStaleThreshold || func() bool {
		__antithesis_instrumentation__.Notify(112766)
		return (n > 0 && func() bool {
			__antithesis_instrumentation__.Notify(112767)
			return td.Input.LogSize >= RaftLogQueueStaleSize == true
		}() == true) == true
	}() == true
}

func (td *truncateDecision) ProtectIndex(index uint64, chosenVia string) {
	__antithesis_instrumentation__.Notify(112768)
	if td.NewFirstIndex > index {
		__antithesis_instrumentation__.Notify(112769)
		td.NewFirstIndex = index
		td.ChosenVia = chosenVia
	} else {
		__antithesis_instrumentation__.Notify(112770)
	}
}

func computeTruncateDecision(input truncateDecisionInput) truncateDecision {
	__antithesis_instrumentation__.Notify(112771)
	decision := truncateDecision{Input: input}
	decision.CommitIndex = input.RaftStatus.Commit

	decision.NewFirstIndex = input.LastIndex
	decision.ChosenVia = truncatableIndexChosenViaLastIndex

	decision.ProtectIndex(decision.CommitIndex, truncatableIndexChosenViaCommitIndex)

	for _, progress := range input.RaftStatus.Progress {
		__antithesis_instrumentation__.Notify(112776)

		if progress.RecentActive {
			__antithesis_instrumentation__.Notify(112778)
			if progress.State == tracker.StateProbe {
				__antithesis_instrumentation__.Notify(112780)
				decision.ProtectIndex(input.FirstIndex, truncatableIndexChosenViaProbingFollower)
			} else {
				__antithesis_instrumentation__.Notify(112781)
				decision.ProtectIndex(progress.Match, truncatableIndexChosenViaFollowers)
			}
			__antithesis_instrumentation__.Notify(112779)
			continue
		} else {
			__antithesis_instrumentation__.Notify(112782)
		}
		__antithesis_instrumentation__.Notify(112777)

		if !input.LogTooLarge() {
			__antithesis_instrumentation__.Notify(112783)
			decision.ProtectIndex(progress.Match, truncatableIndexChosenViaFollowers)
		} else {
			__antithesis_instrumentation__.Notify(112784)
		}

	}
	__antithesis_instrumentation__.Notify(112772)

	if input.PendingSnapshotIndex > 0 {
		__antithesis_instrumentation__.Notify(112785)
		decision.ProtectIndex(input.PendingSnapshotIndex, truncatableIndexChosenViaPendingSnap)
	} else {
		__antithesis_instrumentation__.Notify(112786)
	}
	__antithesis_instrumentation__.Notify(112773)

	if decision.NewFirstIndex < input.FirstIndex {
		__antithesis_instrumentation__.Notify(112787)
		decision.NewFirstIndex = input.FirstIndex
		decision.ChosenVia = truncatableIndexChosenViaFirstIndex
	} else {
		__antithesis_instrumentation__.Notify(112788)
	}
	__antithesis_instrumentation__.Notify(112774)

	logEmpty := input.FirstIndex > input.LastIndex
	noCommittedEntries := input.FirstIndex > input.RaftStatus.Commit

	logIndexValid := logEmpty || func() bool {
		__antithesis_instrumentation__.Notify(112789)
		return ((decision.NewFirstIndex >= input.FirstIndex) && func() bool {
			__antithesis_instrumentation__.Notify(112790)
			return (decision.NewFirstIndex <= input.LastIndex) == true
		}() == true) == true
	}() == true
	commitIndexValid := noCommittedEntries || func() bool {
		__antithesis_instrumentation__.Notify(112791)
		return (decision.NewFirstIndex <= decision.CommitIndex) == true
	}() == true
	valid := logIndexValid && func() bool {
		__antithesis_instrumentation__.Notify(112792)
		return commitIndexValid == true
	}() == true
	if !valid {
		__antithesis_instrumentation__.Notify(112793)
		err := fmt.Sprintf("invalid truncation decision: output = %d, input: [%d, %d], commit idx = %d",
			decision.NewFirstIndex, input.FirstIndex, input.LastIndex, decision.CommitIndex)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(112794)
	}
	__antithesis_instrumentation__.Notify(112775)

	return decision
}

func (rlq *raftLogQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, r *Replica, _ spanconfig.StoreReader,
) (shouldQueue bool, priority float64) {
	__antithesis_instrumentation__.Notify(112795)
	decision, err := newTruncateDecision(ctx, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(112797)
		log.Warningf(ctx, "%v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(112798)
	}
	__antithesis_instrumentation__.Notify(112796)

	shouldQ, _, prio := rlq.shouldQueueImpl(ctx, decision)
	return shouldQ, prio
}

func (rlq *raftLogQueue) shouldQueueImpl(
	ctx context.Context, decision truncateDecision,
) (shouldQ bool, recomputeRaftLogSize bool, priority float64) {
	__antithesis_instrumentation__.Notify(112799)
	if decision.ShouldTruncate() {
		__antithesis_instrumentation__.Notify(112802)
		return true, !decision.Input.LogSizeTrusted, float64(decision.Input.LogSize)
	} else {
		__antithesis_instrumentation__.Notify(112803)
	}
	__antithesis_instrumentation__.Notify(112800)
	if decision.Input.LogSizeTrusted || func() bool {
		__antithesis_instrumentation__.Notify(112804)
		return decision.Input.LastIndex == decision.Input.FirstIndex == true
	}() == true {
		__antithesis_instrumentation__.Notify(112805)

		return false, false, 0
	} else {
		__antithesis_instrumentation__.Notify(112806)
	}
	__antithesis_instrumentation__.Notify(112801)

	return true, true, 1.0 + float64(decision.Input.MaxLogSize)/2.0
}

func (rlq *raftLogQueue) process(
	ctx context.Context, r *Replica, _ spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(112807)
	decision, err := newTruncateDecision(ctx, r)
	if err != nil {
		__antithesis_instrumentation__.Notify(112814)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(112815)
	}
	__antithesis_instrumentation__.Notify(112808)

	if _, recompute, _ := rlq.shouldQueueImpl(ctx, decision); recompute {
		__antithesis_instrumentation__.Notify(112816)
		log.VEventf(ctx, 2, "recomputing raft log based on decision %+v", decision)

		r.raftMu.Lock()
		n, err := ComputeRaftLogSize(ctx, r.RangeID, r.Engine(), r.raftMu.sideloaded)
		if err == nil {
			__antithesis_instrumentation__.Notify(112819)
			r.mu.Lock()
			r.mu.raftLogSize = n
			r.mu.raftLogLastCheckSize = n
			r.mu.raftLogSizeTrusted = true
			r.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(112820)
		}
		__antithesis_instrumentation__.Notify(112817)
		r.raftMu.Unlock()

		if err != nil {
			__antithesis_instrumentation__.Notify(112821)
			return false, errors.Wrap(err, "recomputing raft log size")
		} else {
			__antithesis_instrumentation__.Notify(112822)
		}
		__antithesis_instrumentation__.Notify(112818)

		log.VEventf(ctx, 2, "recomputed raft log size to %s", humanizeutil.IBytes(n))

		decision, err = newTruncateDecision(ctx, r)
		if err != nil {
			__antithesis_instrumentation__.Notify(112823)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(112824)
		}
	} else {
		__antithesis_instrumentation__.Notify(112825)
	}
	__antithesis_instrumentation__.Notify(112809)

	if !decision.ShouldTruncate() {
		__antithesis_instrumentation__.Notify(112826)
		log.VEventf(ctx, 3, "%s", redact.Safe(decision.String()))
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(112827)
	}
	__antithesis_instrumentation__.Notify(112810)

	if n := decision.NumNewRaftSnapshots(); log.V(1) || func() bool {
		__antithesis_instrumentation__.Notify(112828)
		return (n > 0 && func() bool {
			__antithesis_instrumentation__.Notify(112829)
			return rlq.logSnapshots.ShouldProcess(timeutil.Now()) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(112830)
		log.Infof(ctx, "%v", redact.Safe(decision.String()))
	} else {
		__antithesis_instrumentation__.Notify(112831)
		log.VEventf(ctx, 1, "%v", redact.Safe(decision.String()))
	}
	__antithesis_instrumentation__.Notify(112811)
	b := &kv.Batch{}
	truncRequest := &roachpb.TruncateLogRequest{
		RequestHeader: roachpb.RequestHeader{Key: r.Desc().StartKey.AsRawKey()},
		Index:         decision.NewFirstIndex,
		RangeID:       r.RangeID,
	}
	if rlq.store.ClusterSettings().Version.IsActive(
		ctx, clusterversion.LooselyCoupledRaftLogTruncation) {
		__antithesis_instrumentation__.Notify(112832)
		truncRequest.ExpectedFirstIndex = decision.Input.FirstIndex
	} else {
		__antithesis_instrumentation__.Notify(112833)
	}
	__antithesis_instrumentation__.Notify(112812)
	b.AddRawRequest(truncRequest)
	if err := rlq.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(112834)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(112835)
	}
	__antithesis_instrumentation__.Notify(112813)
	r.store.metrics.RaftLogTruncated.Inc(int64(decision.NumTruncatableIndexes()))
	return true, nil
}

func (*raftLogQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(112836)
	return raftLogQueueTimerDuration
}

func (*raftLogQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(112837)
	return nil
}

func isLooselyCoupledRaftLogTruncationEnabled(
	ctx context.Context, settings *cluster.Settings,
) bool {
	__antithesis_instrumentation__.Notify(112838)
	return settings.Version.IsActive(
		ctx, clusterversion.LooselyCoupledRaftLogTruncation) && func() bool {
		__antithesis_instrumentation__.Notify(112839)
		return looselyCoupledTruncationEnabled.Get(&settings.SV) == true
	}() == true
}
