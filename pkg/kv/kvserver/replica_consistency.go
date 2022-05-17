package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const replicaChecksumGCInterval = time.Hour

var fatalOnStatsMismatch = envutil.EnvOrDefaultBool("COCKROACH_ENFORCE_CONSISTENT_STATS", false)

type replicaChecksum struct {
	CollectChecksumResponse

	started bool

	gcTimestamp time.Time

	notify chan struct{}
}

func (r *Replica) CheckConsistency(
	ctx context.Context, args roachpb.CheckConsistencyRequest,
) (roachpb.CheckConsistencyResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(116824)
	startKey := r.Desc().StartKey.AsRawKey()

	checkArgs := roachpb.ComputeChecksumRequest{
		RequestHeader: roachpb.RequestHeader{Key: startKey},
		Version:       batcheval.ReplicaChecksumVersion,
		Snapshot:      args.WithDiff,
		Mode:          args.Mode,
		Checkpoint:    args.Checkpoint,
		Terminate:     args.Terminate,
	}

	isQueue := args.Mode == roachpb.ChecksumMode_CHECK_VIA_QUEUE

	results, err := r.RunConsistencyCheck(ctx, checkArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(116837)
		return roachpb.CheckConsistencyResponse{}, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(116838)
	}
	__antithesis_instrumentation__.Notify(116825)

	res := roachpb.CheckConsistencyResponse_Result{}
	res.RangeID = r.RangeID

	shaToIdxs := map[string][]int{}
	var missing []ConsistencyCheckResult
	for i, result := range results {
		__antithesis_instrumentation__.Notify(116839)
		if result.Err != nil {
			__antithesis_instrumentation__.Notify(116841)
			missing = append(missing, result)
			continue
		} else {
			__antithesis_instrumentation__.Notify(116842)
		}
		__antithesis_instrumentation__.Notify(116840)
		s := string(result.Response.Checksum)
		shaToIdxs[s] = append(shaToIdxs[s], i)
	}
	__antithesis_instrumentation__.Notify(116826)

	var minoritySHA string
	if len(shaToIdxs) > 1 {
		__antithesis_instrumentation__.Notify(116843)
		for sha, idxs := range shaToIdxs {
			__antithesis_instrumentation__.Notify(116844)
			if minoritySHA == "" || func() bool {
				__antithesis_instrumentation__.Notify(116845)
				return len(shaToIdxs[minoritySHA]) > len(idxs) == true
			}() == true {
				__antithesis_instrumentation__.Notify(116846)
				minoritySHA = sha
			} else {
				__antithesis_instrumentation__.Notify(116847)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(116848)
	}
	__antithesis_instrumentation__.Notify(116827)

	if minoritySHA != "" {
		__antithesis_instrumentation__.Notify(116849)
		var buf redact.StringBuilder
		buf.Printf("\n")
		for sha, idxs := range shaToIdxs {
			__antithesis_instrumentation__.Notify(116852)
			minority := redact.Safe("")
			if sha == minoritySHA {
				__antithesis_instrumentation__.Notify(116855)
				minority = redact.Safe(" [minority]")
			} else {
				__antithesis_instrumentation__.Notify(116856)
			}
			__antithesis_instrumentation__.Notify(116853)
			for _, idx := range idxs {
				__antithesis_instrumentation__.Notify(116857)
				buf.Printf("%s: checksum %x%s\n"+
					"- stats: %+v\n"+
					"- stats.Sub(recomputation): %+v\n",
					&results[idx].Replica,
					redact.Safe(sha),
					minority,
					&results[idx].Response.Persisted,
					&results[idx].Response.Delta,
				)
			}
			__antithesis_instrumentation__.Notify(116854)
			minoritySnap := results[shaToIdxs[minoritySHA][0]].Response.Snapshot
			curSnap := results[shaToIdxs[sha][0]].Response.Snapshot
			if sha != minoritySHA && func() bool {
				__antithesis_instrumentation__.Notify(116858)
				return minoritySnap != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(116859)
				return curSnap != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(116860)
				diff := diffRange(curSnap, minoritySnap)
				if report := r.store.cfg.TestingKnobs.ConsistencyTestingKnobs.BadChecksumReportDiff; report != nil {
					__antithesis_instrumentation__.Notify(116862)
					report(*r.store.Ident, diff)
				} else {
					__antithesis_instrumentation__.Notify(116863)
				}
				__antithesis_instrumentation__.Notify(116861)
				buf.Printf("====== diff(%x, [minority]) ======\n%v", redact.Safe(sha), diff)
			} else {
				__antithesis_instrumentation__.Notify(116864)
			}
		}
		__antithesis_instrumentation__.Notify(116850)

		if isQueue {
			__antithesis_instrumentation__.Notify(116865)
			log.Errorf(ctx, "%v", &buf)
		} else {
			__antithesis_instrumentation__.Notify(116866)
		}
		__antithesis_instrumentation__.Notify(116851)
		res.Detail += buf.String()
	} else {
		__antithesis_instrumentation__.Notify(116867)
		res.Detail += fmt.Sprintf("stats: %+v\n", results[0].Response.Persisted)
	}
	__antithesis_instrumentation__.Notify(116828)
	for _, result := range missing {
		__antithesis_instrumentation__.Notify(116868)
		res.Detail += fmt.Sprintf("%s: error: %v\n", result.Replica, result.Err)
	}
	__antithesis_instrumentation__.Notify(116829)

	delta := enginepb.MVCCStats(results[0].Response.Delta)
	var haveDelta bool
	{
		__antithesis_instrumentation__.Notify(116869)
		d2 := delta
		d2.AgeTo(0)
		haveDelta = d2 != enginepb.MVCCStats{}
	}
	__antithesis_instrumentation__.Notify(116830)

	res.StartKey = []byte(startKey)
	res.Status = roachpb.CheckConsistencyResponse_RANGE_CONSISTENT
	if minoritySHA != "" {
		__antithesis_instrumentation__.Notify(116870)
		res.Status = roachpb.CheckConsistencyResponse_RANGE_INCONSISTENT
	} else {
		__antithesis_instrumentation__.Notify(116871)
		if args.Mode != roachpb.ChecksumMode_CHECK_STATS && func() bool {
			__antithesis_instrumentation__.Notify(116872)
			return haveDelta == true
		}() == true {
			__antithesis_instrumentation__.Notify(116873)
			if delta.ContainsEstimates > 0 {
				__antithesis_instrumentation__.Notify(116875)

				res.Status = roachpb.CheckConsistencyResponse_RANGE_CONSISTENT_STATS_ESTIMATED
			} else {
				__antithesis_instrumentation__.Notify(116876)

				res.Status = roachpb.CheckConsistencyResponse_RANGE_CONSISTENT_STATS_INCORRECT
			}
			__antithesis_instrumentation__.Notify(116874)
			res.Detail += fmt.Sprintf("stats - recomputation: %+v\n", enginepb.MVCCStats(results[0].Response.Delta))
		} else {
			__antithesis_instrumentation__.Notify(116877)
			if len(missing) > 0 {
				__antithesis_instrumentation__.Notify(116878)

				res.Status = roachpb.CheckConsistencyResponse_RANGE_INDETERMINATE
			} else {
				__antithesis_instrumentation__.Notify(116879)
			}
		}
	}
	__antithesis_instrumentation__.Notify(116831)
	var resp roachpb.CheckConsistencyResponse
	resp.Result = append(resp.Result, res)

	if !isQueue {
		__antithesis_instrumentation__.Notify(116880)
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(116881)
	}
	__antithesis_instrumentation__.Notify(116832)

	if minoritySHA == "" {
		__antithesis_instrumentation__.Notify(116882)

		if !haveDelta {
			__antithesis_instrumentation__.Notify(116885)
			return resp, nil
		} else {
			__antithesis_instrumentation__.Notify(116886)
		}
		__antithesis_instrumentation__.Notify(116883)

		if delta.ContainsEstimates <= 0 && func() bool {
			__antithesis_instrumentation__.Notify(116887)
			return fatalOnStatsMismatch == true
		}() == true {
			__antithesis_instrumentation__.Notify(116888)

			var v roachpb.Version
			if err := r.store.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(116890)
				return txn.GetProto(ctx, keys.BootstrapVersionKey, &v)
			}); err != nil {
				__antithesis_instrumentation__.Notify(116891)
				log.Infof(ctx, "while retrieving cluster bootstrap version: %s", err)

				v = r.store.cfg.Settings.Version.ActiveVersion(ctx).Version
			} else {
				__antithesis_instrumentation__.Notify(116892)
			}
			__antithesis_instrumentation__.Notify(116889)

			if !v.Less(roachpb.Version{Major: 19, Minor: 1}) {
				__antithesis_instrumentation__.Notify(116893)

				if v.Less(roachpb.Version{Major: 20, Minor: 1, Internal: 14}) {
					__antithesis_instrumentation__.Notify(116896)
					delta.AbortSpanBytes = 0
					haveDelta = delta != enginepb.MVCCStats{}
				} else {
					__antithesis_instrumentation__.Notify(116897)
				}
				__antithesis_instrumentation__.Notify(116894)
				if !haveDelta {
					__antithesis_instrumentation__.Notify(116898)
					return resp, nil
				} else {
					__antithesis_instrumentation__.Notify(116899)
				}
				__antithesis_instrumentation__.Notify(116895)
				log.Fatalf(ctx, "found a delta of %+v", redact.Safe(delta))
			} else {
				__antithesis_instrumentation__.Notify(116900)
			}
		} else {
			__antithesis_instrumentation__.Notify(116901)
		}
		__antithesis_instrumentation__.Notify(116884)

		log.Infof(ctx, "triggering stats recomputation to resolve delta of %+v", results[0].Response.Delta)

		req := roachpb.RecomputeStatsRequest{
			RequestHeader: roachpb.RequestHeader{Key: startKey},
		}

		var b kv.Batch
		b.AddRawRequest(&req)

		err := r.store.db.Run(ctx, &b)
		return resp, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(116902)
	}
	__antithesis_instrumentation__.Notify(116833)

	if args.WithDiff {
		__antithesis_instrumentation__.Notify(116903)

		log.Errorf(ctx, "consistency check failed")
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(116904)
	}
	__antithesis_instrumentation__.Notify(116834)

	args.WithDiff = true
	args.Checkpoint = true
	for _, idxs := range shaToIdxs[minoritySHA] {
		__antithesis_instrumentation__.Notify(116905)
		args.Terminate = append(args.Terminate, results[idxs].Replica)
	}

	{
		__antithesis_instrumentation__.Notify(116906)
		var tmp redact.SafeFormatter = roachpb.MakeReplicaSet(args.Terminate)
		log.Errorf(ctx, "consistency check failed; fetching details and shutting down minority %v", tmp)
	}
	__antithesis_instrumentation__.Notify(116835)

	defer log.TemporarilyDisableFileGCForMainLogger()()

	if _, pErr := r.CheckConsistency(ctx, args); pErr != nil {
		__antithesis_instrumentation__.Notify(116907)
		log.Errorf(ctx, "replica inconsistency detected; could not obtain actual diff: %s", pErr)
	} else {
		__antithesis_instrumentation__.Notify(116908)
	}
	__antithesis_instrumentation__.Notify(116836)

	return resp, nil
}

type ConsistencyCheckResult struct {
	Replica  roachpb.ReplicaDescriptor
	Response CollectChecksumResponse
	Err      error
}

func (r *Replica) collectChecksumFromReplica(
	ctx context.Context, replica roachpb.ReplicaDescriptor, id uuid.UUID, checksum []byte,
) (CollectChecksumResponse, error) {
	__antithesis_instrumentation__.Notify(116909)
	conn, err := r.store.cfg.NodeDialer.Dial(ctx, replica.NodeID, rpc.DefaultClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(116912)
		return CollectChecksumResponse{},
			errors.Wrapf(err, "could not dial node ID %d", replica.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(116913)
	}
	__antithesis_instrumentation__.Notify(116910)
	client := NewPerReplicaClient(conn)
	req := &CollectChecksumRequest{
		StoreRequestHeader: StoreRequestHeader{NodeID: replica.NodeID, StoreID: replica.StoreID},
		RangeID:            r.RangeID,
		ChecksumID:         id,
		Checksum:           checksum,
	}
	resp, err := client.CollectChecksum(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(116914)
		return CollectChecksumResponse{}, err
	} else {
		__antithesis_instrumentation__.Notify(116915)
	}
	__antithesis_instrumentation__.Notify(116911)
	return *resp, nil
}

func (r *Replica) RunConsistencyCheck(
	ctx context.Context, req roachpb.ComputeChecksumRequest,
) ([]ConsistencyCheckResult, error) {
	__antithesis_instrumentation__.Notify(116916)

	res, pErr := kv.SendWrapped(ctx, r.store.db.NonTransactionalSender(), &req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(116921)
		return nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(116922)
	}
	__antithesis_instrumentation__.Notify(116917)
	ccRes := res.(*roachpb.ComputeChecksumResponse)

	var orderedReplicas []roachpb.ReplicaDescriptor
	{
		__antithesis_instrumentation__.Notify(116923)
		desc := r.Desc()
		localReplica, err := r.GetReplicaDescriptor()
		if err != nil {
			__antithesis_instrumentation__.Notify(116925)
			return nil, errors.Wrap(err, "could not get replica descriptor")
		} else {
			__antithesis_instrumentation__.Notify(116926)
		}
		__antithesis_instrumentation__.Notify(116924)

		orderedReplicas = append(orderedReplicas, desc.Replicas().Descriptors()...)

		sort.Slice(orderedReplicas, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(116927)
			return orderedReplicas[i] == localReplica
		})
	}
	__antithesis_instrumentation__.Notify(116918)

	resultCh := make(chan ConsistencyCheckResult, len(orderedReplicas))
	var results []ConsistencyCheckResult
	var wg sync.WaitGroup

	for _, replica := range orderedReplicas {
		__antithesis_instrumentation__.Notify(116928)
		wg.Add(1)
		replica := replica
		if err := r.store.Stopper().RunAsyncTask(ctx, "storage.Replica: checking consistency",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(116930)
				defer wg.Done()

				var masterChecksum []byte
				if len(results) > 0 {
					__antithesis_instrumentation__.Notify(116932)
					masterChecksum = results[0].Response.Checksum
				} else {
					__antithesis_instrumentation__.Notify(116933)
				}
				__antithesis_instrumentation__.Notify(116931)
				resp, err := r.collectChecksumFromReplica(ctx, replica, ccRes.ChecksumID, masterChecksum)
				resultCh <- ConsistencyCheckResult{
					Replica:  replica,
					Response: resp,
					Err:      err,
				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(116934)
			wg.Done()

			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116935)
		}
		__antithesis_instrumentation__.Notify(116929)

		if len(results) == 0 {
			__antithesis_instrumentation__.Notify(116936)
			wg.Wait()
			result := <-resultCh
			if err := result.Err; err != nil {
				__antithesis_instrumentation__.Notify(116938)

				return nil, errors.Wrap(err, "computing own checksum")
			} else {
				__antithesis_instrumentation__.Notify(116939)
			}
			__antithesis_instrumentation__.Notify(116937)
			results = append(results, result)
		} else {
			__antithesis_instrumentation__.Notify(116940)
		}
	}
	__antithesis_instrumentation__.Notify(116919)

	wg.Wait()
	close(resultCh)

	for result := range resultCh {
		__antithesis_instrumentation__.Notify(116941)
		results = append(results, result)
	}
	__antithesis_instrumentation__.Notify(116920)

	return results, nil
}

func (r *Replica) gcOldChecksumEntriesLocked(now time.Time) {
	__antithesis_instrumentation__.Notify(116942)
	for id, val := range r.mu.checksums {
		__antithesis_instrumentation__.Notify(116943)

		if !val.gcTimestamp.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(116944)
			return now.After(val.gcTimestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(116945)
			delete(r.mu.checksums, id)
		} else {
			__antithesis_instrumentation__.Notify(116946)
		}
	}
}

func (r *Replica) getChecksum(ctx context.Context, id uuid.UUID) (replicaChecksum, error) {
	__antithesis_instrumentation__.Notify(116947)
	now := timeutil.Now()
	r.mu.Lock()
	r.gcOldChecksumEntriesLocked(now)
	c, ok := r.mu.checksums[id]
	if !ok {
		__antithesis_instrumentation__.Notify(116953)

		if d, dOk := ctx.Deadline(); dOk {
			__antithesis_instrumentation__.Notify(116955)
			c.gcTimestamp = d
		} else {
			__antithesis_instrumentation__.Notify(116956)
		}
		__antithesis_instrumentation__.Notify(116954)
		c.notify = make(chan struct{})
		r.mu.checksums[id] = c
	} else {
		__antithesis_instrumentation__.Notify(116957)
	}
	__antithesis_instrumentation__.Notify(116948)
	r.mu.Unlock()

	computed, err := r.checksumInitialWait(ctx, id, c.notify)
	if err != nil {
		__antithesis_instrumentation__.Notify(116958)
		return replicaChecksum{}, err
	} else {
		__antithesis_instrumentation__.Notify(116959)
	}
	__antithesis_instrumentation__.Notify(116949)

	if !computed {
		__antithesis_instrumentation__.Notify(116960)
		_, err = r.checksumWait(ctx, id, c.notify, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(116961)
			return replicaChecksum{}, err
		} else {
			__antithesis_instrumentation__.Notify(116962)
		}
	} else {
		__antithesis_instrumentation__.Notify(116963)
	}
	__antithesis_instrumentation__.Notify(116950)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(116964)
		log.Infof(ctx, "waited for compute checksum for %s", timeutil.Since(now))
	} else {
		__antithesis_instrumentation__.Notify(116965)
	}
	__antithesis_instrumentation__.Notify(116951)
	r.mu.RLock()
	c, ok = r.mu.checksums[id]
	r.mu.RUnlock()

	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(116966)
		return c.Checksum == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(116967)
		return replicaChecksum{}, errors.Errorf("no checksum found (ID = %s)", id)
	} else {
		__antithesis_instrumentation__.Notify(116968)
	}
	__antithesis_instrumentation__.Notify(116952)
	return c, nil
}

func (r *Replica) checksumInitialWait(
	ctx context.Context, id uuid.UUID, notify chan struct{},
) (bool, error) {
	__antithesis_instrumentation__.Notify(116969)
	d, dOk := ctx.Deadline()

	maxInitialWait := 5 * time.Second
	var initialWait <-chan time.Time
	if dOk {
		__antithesis_instrumentation__.Notify(116971)
		duration := time.Duration(timeutil.Until(d).Nanoseconds() / 10)
		if duration > maxInitialWait {
			__antithesis_instrumentation__.Notify(116973)
			duration = maxInitialWait
		} else {
			__antithesis_instrumentation__.Notify(116974)
		}
		__antithesis_instrumentation__.Notify(116972)
		initialWait = time.After(duration)
	} else {
		__antithesis_instrumentation__.Notify(116975)
		initialWait = time.After(maxInitialWait)
	}
	__antithesis_instrumentation__.Notify(116970)
	return r.checksumWait(ctx, id, notify, initialWait)
}

func (r *Replica) checksumWait(
	ctx context.Context, id uuid.UUID, notify chan struct{}, initialWait <-chan time.Time,
) (bool, error) {
	__antithesis_instrumentation__.Notify(116976)

	select {
	case <-r.store.Stopper().ShouldQuiesce():
		__antithesis_instrumentation__.Notify(116977)
		return false,
			errors.Errorf("store quiescing while waiting for compute checksum (ID = %s)", id)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(116978)
		return false,
			errors.Wrapf(ctx.Err(), "while waiting for compute checksum (ID = %s)", id)
	case <-initialWait:
		__antithesis_instrumentation__.Notify(116979)
		{
			__antithesis_instrumentation__.Notify(116981)
			r.mu.Lock()
			started := r.mu.checksums[id].started
			r.mu.Unlock()
			if !started {
				__antithesis_instrumentation__.Notify(116983)
				return false,
					errors.Errorf("checksum computation did not start in time for (ID = %s)", id)
			} else {
				__antithesis_instrumentation__.Notify(116984)
			}
			__antithesis_instrumentation__.Notify(116982)
			return false, nil
		}
	case <-notify:
		__antithesis_instrumentation__.Notify(116980)
		return true, nil
	}
}

func (r *Replica) computeChecksumDone(
	ctx context.Context, id uuid.UUID, result *replicaHash, snapshot *roachpb.RaftSnapshotData,
) {
	__antithesis_instrumentation__.Notify(116985)
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.mu.checksums[id]; ok {
		__antithesis_instrumentation__.Notify(116986)
		if result != nil {
			__antithesis_instrumentation__.Notify(116988)
			c.Checksum = result.SHA512[:]

			delta := result.PersistedMS
			delta.Subtract(result.RecomputedMS)
			c.Delta = enginepb.MVCCStatsDelta(delta)
			c.Persisted = result.PersistedMS
		} else {
			__antithesis_instrumentation__.Notify(116989)
		}
		__antithesis_instrumentation__.Notify(116987)
		c.gcTimestamp = timeutil.Now().Add(replicaChecksumGCInterval)
		c.Snapshot = snapshot
		r.mu.checksums[id] = c

		close(c.notify)
	} else {
		__antithesis_instrumentation__.Notify(116990)

		log.Errorf(ctx, "no map entry for checksum (ID = %s)", id)
	}
}

type replicaHash struct {
	SHA512                    [sha512.Size]byte
	PersistedMS, RecomputedMS enginepb.MVCCStats
}

func (*Replica) sha512(
	ctx context.Context,
	desc roachpb.RangeDescriptor,
	snap storage.Reader,
	snapshot *roachpb.RaftSnapshotData,
	mode roachpb.ChecksumMode,
	limiter *quotapool.RateLimiter,
) (*replicaHash, error) {
	__antithesis_instrumentation__.Notify(116991)
	statsOnly := mode == roachpb.ChecksumMode_CHECK_STATS

	var alloc bufalloc.ByteAllocator
	var intBuf [8]byte
	var legacyTimestamp hlc.LegacyTimestamp
	var timestampBuf []byte
	hasher := sha512.New()

	visitor := func(unsafeKey storage.MVCCKey, unsafeValue []byte) error {
		__antithesis_instrumentation__.Notify(116996)

		if err := limiter.WaitN(ctx, int64(len(unsafeKey.Key)+len(unsafeValue))); err != nil {
			__antithesis_instrumentation__.Notify(117005)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117006)
		}
		__antithesis_instrumentation__.Notify(116997)

		if snapshot != nil {
			__antithesis_instrumentation__.Notify(117007)

			kv := roachpb.RaftSnapshotData_KeyValue{
				Timestamp: unsafeKey.Timestamp,
			}
			alloc, kv.Key = alloc.Copy(unsafeKey.Key, 0)
			alloc, kv.Value = alloc.Copy(unsafeValue, 0)
			snapshot.KV = append(snapshot.KV, kv)
		} else {
			__antithesis_instrumentation__.Notify(117008)
		}
		__antithesis_instrumentation__.Notify(116998)

		binary.LittleEndian.PutUint64(intBuf[:], uint64(len(unsafeKey.Key)))
		if _, err := hasher.Write(intBuf[:]); err != nil {
			__antithesis_instrumentation__.Notify(117009)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117010)
		}
		__antithesis_instrumentation__.Notify(116999)
		binary.LittleEndian.PutUint64(intBuf[:], uint64(len(unsafeValue)))
		if _, err := hasher.Write(intBuf[:]); err != nil {
			__antithesis_instrumentation__.Notify(117011)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117012)
		}
		__antithesis_instrumentation__.Notify(117000)
		if _, err := hasher.Write(unsafeKey.Key); err != nil {
			__antithesis_instrumentation__.Notify(117013)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117014)
		}
		__antithesis_instrumentation__.Notify(117001)
		legacyTimestamp = unsafeKey.Timestamp.ToLegacyTimestamp()
		if size := legacyTimestamp.Size(); size > cap(timestampBuf) {
			__antithesis_instrumentation__.Notify(117015)
			timestampBuf = make([]byte, size)
		} else {
			__antithesis_instrumentation__.Notify(117016)
			timestampBuf = timestampBuf[:size]
		}
		__antithesis_instrumentation__.Notify(117002)
		if _, err := protoutil.MarshalTo(&legacyTimestamp, timestampBuf); err != nil {
			__antithesis_instrumentation__.Notify(117017)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117018)
		}
		__antithesis_instrumentation__.Notify(117003)
		if _, err := hasher.Write(timestampBuf); err != nil {
			__antithesis_instrumentation__.Notify(117019)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117020)
		}
		__antithesis_instrumentation__.Notify(117004)
		_, err := hasher.Write(unsafeValue)
		return err
	}
	__antithesis_instrumentation__.Notify(116992)

	var ms enginepb.MVCCStats

	if !statsOnly {
		__antithesis_instrumentation__.Notify(117021)

		for _, span := range rditer.MakeReplicatedKeyRangesExceptLockTable(&desc) {
			__antithesis_instrumentation__.Notify(117022)
			iter := snap.NewMVCCIterator(storage.MVCCKeyAndIntentsIterKind,
				storage.IterOptions{UpperBound: span.End})
			spanMS, err := storage.ComputeStatsForRange(
				iter, span.Start, span.End, 0, visitor,
			)
			iter.Close()
			if err != nil {
				__antithesis_instrumentation__.Notify(117024)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(117025)
			}
			__antithesis_instrumentation__.Notify(117023)
			ms.Add(spanMS)
		}
	} else {
		__antithesis_instrumentation__.Notify(117026)
	}
	__antithesis_instrumentation__.Notify(116993)

	var result replicaHash
	result.RecomputedMS = ms

	rangeAppliedState, err := stateloader.Make(desc.RangeID).LoadRangeAppliedState(ctx, snap)
	if err != nil {
		__antithesis_instrumentation__.Notify(117027)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(117028)
	}
	__antithesis_instrumentation__.Notify(116994)
	result.PersistedMS = rangeAppliedState.RangeStats.ToStats()

	if statsOnly {
		__antithesis_instrumentation__.Notify(117029)
		b, err := protoutil.Marshal(&rangeAppliedState)
		if err != nil {
			__antithesis_instrumentation__.Notify(117032)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(117033)
		}
		__antithesis_instrumentation__.Notify(117030)
		if snapshot != nil {
			__antithesis_instrumentation__.Notify(117034)

			kv := roachpb.RaftSnapshotData_KeyValue{
				Timestamp: hlc.Timestamp{},
			}
			kv.Key = keys.RangeAppliedStateKey(desc.RangeID)
			var v roachpb.Value
			if err := v.SetProto(&rangeAppliedState); err != nil {
				__antithesis_instrumentation__.Notify(117036)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(117037)
			}
			__antithesis_instrumentation__.Notify(117035)
			kv.Value = v.RawBytes
			snapshot.KV = append(snapshot.KV, kv)
		} else {
			__antithesis_instrumentation__.Notify(117038)
		}
		__antithesis_instrumentation__.Notify(117031)
		if _, err := hasher.Write(b); err != nil {
			__antithesis_instrumentation__.Notify(117039)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(117040)
		}
	} else {
		__antithesis_instrumentation__.Notify(117041)
	}
	__antithesis_instrumentation__.Notify(116995)

	hasher.Sum(result.SHA512[:0])

	result.RecomputedMS.AgeTo(result.PersistedMS.LastUpdateNanos)

	return &result, nil
}

func (r *Replica) computeChecksumPostApply(ctx context.Context, cc kvserverpb.ComputeChecksum) {
	__antithesis_instrumentation__.Notify(117042)
	stopper := r.store.Stopper()
	now := timeutil.Now()
	r.mu.Lock()
	var notify chan struct{}
	if c, ok := r.mu.checksums[cc.ChecksumID]; !ok {
		__antithesis_instrumentation__.Notify(117047)

		notify = make(chan struct{})
	} else {
		__antithesis_instrumentation__.Notify(117048)
		if !c.started {
			__antithesis_instrumentation__.Notify(117049)

			notify = c.notify
		} else {
			__antithesis_instrumentation__.Notify(117050)
			log.Fatalf(ctx, "attempted to apply ComputeChecksum command with duplicated checksum ID %s",
				cc.ChecksumID)
		}
	}
	__antithesis_instrumentation__.Notify(117043)

	r.gcOldChecksumEntriesLocked(now)

	r.mu.checksums[cc.ChecksumID] = replicaChecksum{started: true, notify: notify}
	desc := *r.mu.state.Desc
	r.mu.Unlock()

	if cc.Version != batcheval.ReplicaChecksumVersion {
		__antithesis_instrumentation__.Notify(117051)
		r.computeChecksumDone(ctx, cc.ChecksumID, nil, nil)
		log.Infof(ctx, "incompatible ComputeChecksum versions (requested: %d, have: %d)",
			cc.Version, batcheval.ReplicaChecksumVersion)
		return
	} else {
		__antithesis_instrumentation__.Notify(117052)
	}
	__antithesis_instrumentation__.Notify(117044)

	snap := r.store.engine.NewSnapshot()
	if cc.Checkpoint {
		__antithesis_instrumentation__.Notify(117053)
		sl := stateloader.Make(r.RangeID)
		as, err := sl.LoadRangeAppliedState(ctx, snap)
		if err != nil {
			__antithesis_instrumentation__.Notify(117055)
			log.Warningf(ctx, "unable to load applied index, continuing anyway")
		} else {
			__antithesis_instrumentation__.Notify(117056)
		}
		__antithesis_instrumentation__.Notify(117054)

		tag := fmt.Sprintf("r%d_at_%d", r.RangeID, as.RaftAppliedIndex)
		if dir, err := r.store.checkpoint(ctx, tag); err != nil {
			__antithesis_instrumentation__.Notify(117057)
			log.Warningf(ctx, "unable to create checkpoint %s: %+v", dir, err)
		} else {
			__antithesis_instrumentation__.Notify(117058)
			log.Warningf(ctx, "created checkpoint %s", dir)
		}
	} else {
		__antithesis_instrumentation__.Notify(117059)
	}
	__antithesis_instrumentation__.Notify(117045)

	const taskName = "storage.Replica: computing checksum"
	sem := r.store.consistencySem
	if cc.Mode == roachpb.ChecksumMode_CHECK_STATS {
		__antithesis_instrumentation__.Notify(117060)

		sem = nil
	} else {
		__antithesis_instrumentation__.Notify(117061)
	}
	__antithesis_instrumentation__.Notify(117046)
	if err := stopper.RunAsyncTaskEx(r.AnnotateCtx(context.Background()), stop.TaskOpts{
		TaskName:   taskName,
		Sem:        sem,
		WaitForSem: false,
	}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(117062)
		if err := contextutil.RunWithTimeout(ctx, taskName, consistencyCheckAsyncTimeout,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(117065)
				defer snap.Close()
				var snapshot *roachpb.RaftSnapshotData
				if cc.SaveSnapshot {
					__antithesis_instrumentation__.Notify(117068)
					snapshot = &roachpb.RaftSnapshotData{}
				} else {
					__antithesis_instrumentation__.Notify(117069)
				}
				__antithesis_instrumentation__.Notify(117066)

				result, err := r.sha512(ctx, desc, snap, snapshot, cc.Mode, r.store.consistencyLimiter)
				if err != nil {
					__antithesis_instrumentation__.Notify(117070)
					result = nil
				} else {
					__antithesis_instrumentation__.Notify(117071)
				}
				__antithesis_instrumentation__.Notify(117067)
				r.computeChecksumDone(ctx, cc.ChecksumID, result, snapshot)
				return err
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(117072)
			log.Errorf(ctx, "checksum computation failed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(117073)
		}
		__antithesis_instrumentation__.Notify(117063)

		var shouldFatal bool
		for _, rDesc := range cc.Terminate {
			__antithesis_instrumentation__.Notify(117074)
			if rDesc.StoreID == r.store.StoreID() && func() bool {
				__antithesis_instrumentation__.Notify(117075)
				return rDesc.ReplicaID == r.replicaID == true
			}() == true {
				__antithesis_instrumentation__.Notify(117076)
				shouldFatal = true
			} else {
				__antithesis_instrumentation__.Notify(117077)
			}
		}
		__antithesis_instrumentation__.Notify(117064)

		if shouldFatal {
			__antithesis_instrumentation__.Notify(117078)

			auxDir := r.store.engine.GetAuxiliaryDir()
			_ = r.store.engine.MkdirAll(auxDir)
			path := base.PreventedStartupFile(auxDir)

			const attentionFmt = `ATTENTION:

this node is terminating because a replica inconsistency was detected between %s
and its other replicas. Please check your cluster-wide log files for more
information and contact the CockroachDB support team. It is not necessarily safe
to replace this node; cluster data may still be at risk of corruption.

A checkpoints directory to aid (expert) debugging should be present in:
%s

A file preventing this node from restarting was placed at:
%s
`
			preventStartupMsg := fmt.Sprintf(attentionFmt, r, auxDir, path)
			if err := fs.WriteFile(r.store.engine, path, []byte(preventStartupMsg)); err != nil {
				__antithesis_instrumentation__.Notify(117080)
				log.Warningf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(117081)
			}
			__antithesis_instrumentation__.Notify(117079)

			if p := r.store.cfg.TestingKnobs.ConsistencyTestingKnobs.OnBadChecksumFatal; p != nil {
				__antithesis_instrumentation__.Notify(117082)
				p(*r.store.Ident)
			} else {
				__antithesis_instrumentation__.Notify(117083)
				time.Sleep(10 * time.Second)
				log.Fatalf(r.AnnotateCtx(context.Background()), attentionFmt, r, auxDir, path)
			}
		} else {
			__antithesis_instrumentation__.Notify(117084)
		}

	}); err != nil {
		__antithesis_instrumentation__.Notify(117085)
		defer snap.Close()
		log.Errorf(ctx, "could not run async checksum computation (ID = %s): %v", cc.ChecksumID, err)

		r.computeChecksumDone(ctx, cc.ChecksumID, nil, nil)
	} else {
		__antithesis_instrumentation__.Notify(117086)
	}
}
