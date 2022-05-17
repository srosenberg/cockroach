package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/raft/v3/tracker"
)

const mergeApplicationTimeout = 5 * time.Second

var sendSnapshotTimeout = envutil.EnvOrDefaultDuration(
	"COCKROACH_RAFT_SEND_SNAPSHOT_TIMEOUT", 1*time.Hour)

func (r *Replica) AdminSplit(
	ctx context.Context, args roachpb.AdminSplitRequest, reason string,
) (reply roachpb.AdminSplitResponse, _ *roachpb.Error) {
	__antithesis_instrumentation__.Notify(115843)
	if len(args.SplitKey) == 0 {
		__antithesis_instrumentation__.Notify(115846)
		return roachpb.AdminSplitResponse{}, roachpb.NewErrorf("cannot split range with no key provided")
	} else {
		__antithesis_instrumentation__.Notify(115847)
	}
	__antithesis_instrumentation__.Notify(115844)

	err := r.executeAdminCommandWithDescriptor(ctx, func(desc *roachpb.RangeDescriptor) error {
		__antithesis_instrumentation__.Notify(115848)
		var err error
		reply, err = r.adminSplitWithDescriptor(ctx, args, desc, true, reason)
		return err
	})
	__antithesis_instrumentation__.Notify(115845)
	return reply, err
}

func maybeDescriptorChangedError(
	desc *roachpb.RangeDescriptor, err error,
) (ok bool, expectedDesc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(115849)
	if detail := (*roachpb.ConditionFailedError)(nil); errors.As(err, &detail) {
		__antithesis_instrumentation__.Notify(115851)

		var actualDesc roachpb.RangeDescriptor
		if !detail.ActualValue.IsPresent() {
			__antithesis_instrumentation__.Notify(115852)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(115853)
			if err := detail.ActualValue.GetProto(&actualDesc); err == nil && func() bool {
				__antithesis_instrumentation__.Notify(115854)
				return !desc.Equal(&actualDesc) == true
			}() == true {
				__antithesis_instrumentation__.Notify(115855)
				return true, &actualDesc
			} else {
				__antithesis_instrumentation__.Notify(115856)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(115857)
	}
	__antithesis_instrumentation__.Notify(115850)
	return false, nil
}

func splitSnapshotWarningStr(rangeID roachpb.RangeID, status *raft.Status) string {
	__antithesis_instrumentation__.Notify(115858)
	var s string
	if status != nil && func() bool {
		__antithesis_instrumentation__.Notify(115860)
		return status.RaftState == raft.StateLeader == true
	}() == true {
		__antithesis_instrumentation__.Notify(115861)
		for replicaID, pr := range status.Progress {
			__antithesis_instrumentation__.Notify(115862)
			if replicaID == status.Lead {
				__antithesis_instrumentation__.Notify(115865)

				continue
			} else {
				__antithesis_instrumentation__.Notify(115866)
			}
			__antithesis_instrumentation__.Notify(115863)
			if pr.State == tracker.StateReplicate {
				__antithesis_instrumentation__.Notify(115867)

				continue
			} else {
				__antithesis_instrumentation__.Notify(115868)
			}
			__antithesis_instrumentation__.Notify(115864)
			s += fmt.Sprintf("; r%d/%d is ", rangeID, replicaID)
			switch pr.State {
			case tracker.StateSnapshot:
				__antithesis_instrumentation__.Notify(115869)

				s += "waiting for a Raft snapshot"
			case tracker.StateProbe:
				__antithesis_instrumentation__.Notify(115870)

				s += "being probed (may or may not need a Raft snapshot)"
			default:
				__antithesis_instrumentation__.Notify(115871)

				s += "in unknown state " + pr.State.String()
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(115872)
	}
	__antithesis_instrumentation__.Notify(115859)
	return s
}

func prepareSplitDescs(
	ctx context.Context,
	st *cluster.Settings,
	rightRangeID roachpb.RangeID,
	splitKey roachpb.RKey,
	expiration hlc.Timestamp,
	leftDesc *roachpb.RangeDescriptor,
) (*roachpb.RangeDescriptor, *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(115873)

	rightDesc := roachpb.NewRangeDescriptor(rightRangeID, splitKey, leftDesc.EndKey, leftDesc.Replicas())

	{
		__antithesis_instrumentation__.Notify(115875)
		tmp := *leftDesc
		leftDesc = &tmp
	}
	__antithesis_instrumentation__.Notify(115874)

	leftDesc.IncrementGeneration()
	leftDesc.EndKey = splitKey

	rightDesc.Generation = leftDesc.Generation

	setStickyBit(rightDesc, expiration)
	return leftDesc, rightDesc
}

func setStickyBit(desc *roachpb.RangeDescriptor, expiration hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(115876)

	if !expiration.IsEmpty() {
		__antithesis_instrumentation__.Notify(115877)
		desc.StickyBit = &expiration
	} else {
		__antithesis_instrumentation__.Notify(115878)
	}
}

func splitTxnAttempt(
	ctx context.Context,
	store *Store,
	txn *kv.Txn,
	rightRangeID roachpb.RangeID,
	splitKey roachpb.RKey,
	expiration hlc.Timestamp,
	oldDesc *roachpb.RangeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(115879)
	txn.SetDebugName(splitTxnName)

	_, dbDescValue, err := conditionalGetDescValueFromDB(
		ctx, txn, oldDesc.StartKey, false, checkDescsEqual(oldDesc))
	if err != nil {
		__antithesis_instrumentation__.Notify(115885)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115886)
	}
	__antithesis_instrumentation__.Notify(115880)

	desc := oldDesc
	oldDesc = nil

	leftDesc, rightDesc := prepareSplitDescs(
		ctx, store.ClusterSettings(), rightRangeID, splitKey, expiration, desc)

	{
		__antithesis_instrumentation__.Notify(115887)
		b := txn.NewBatch()
		leftDescKey := keys.RangeDescriptorKey(leftDesc.StartKey)
		if err := updateRangeDescriptor(ctx, b, leftDescKey, dbDescValue, leftDesc); err != nil {
			__antithesis_instrumentation__.Notify(115889)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115890)
		}
		__antithesis_instrumentation__.Notify(115888)

		log.Event(ctx, "updating LHS descriptor")
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(115891)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115892)
		}
	}
	__antithesis_instrumentation__.Notify(115881)

	if err := store.logSplit(ctx, txn, *leftDesc, *rightDesc); err != nil {
		__antithesis_instrumentation__.Notify(115893)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115894)
	}
	__antithesis_instrumentation__.Notify(115882)

	b := txn.NewBatch()

	rightDescKey := keys.RangeDescriptorKey(rightDesc.StartKey)
	if err := updateRangeDescriptor(ctx, b, rightDescKey, nil, rightDesc); err != nil {
		__antithesis_instrumentation__.Notify(115895)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115896)
	}
	__antithesis_instrumentation__.Notify(115883)

	if err := splitRangeAddressing(b, rightDesc, leftDesc); err != nil {
		__antithesis_instrumentation__.Notify(115897)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115898)
	}
	__antithesis_instrumentation__.Notify(115884)

	b.AddRawRequest(&roachpb.EndTxnRequest{
		Commit: true,
		InternalCommitTrigger: &roachpb.InternalCommitTrigger{
			SplitTrigger: &roachpb.SplitTrigger{
				LeftDesc:  *leftDesc,
				RightDesc: *rightDesc,
			},
		},
	})

	log.Event(ctx, "commit txn with batch containing RHS descriptor and meta records")
	return txn.Run(ctx, b)
}

func splitTxnStickyUpdateAttempt(
	ctx context.Context, txn *kv.Txn, desc *roachpb.RangeDescriptor, expiration hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(115899)
	_, dbDescValue, err := conditionalGetDescValueFromDB(
		ctx, txn, desc.StartKey, false, checkDescsEqual(desc))
	if err != nil {
		__antithesis_instrumentation__.Notify(115903)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115904)
	}
	__antithesis_instrumentation__.Notify(115900)
	newDesc := *desc
	setStickyBit(&newDesc, expiration)

	b := txn.NewBatch()
	descKey := keys.RangeDescriptorKey(desc.StartKey)
	if err := updateRangeDescriptor(ctx, b, descKey, dbDescValue, &newDesc); err != nil {
		__antithesis_instrumentation__.Notify(115905)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115906)
	}
	__antithesis_instrumentation__.Notify(115901)
	if err := updateRangeAddressing(b, &newDesc); err != nil {
		__antithesis_instrumentation__.Notify(115907)
		return err
	} else {
		__antithesis_instrumentation__.Notify(115908)
	}
	__antithesis_instrumentation__.Notify(115902)

	b.AddRawRequest(&roachpb.EndTxnRequest{
		Commit: true,
		InternalCommitTrigger: &roachpb.InternalCommitTrigger{
			StickyBitTrigger: &roachpb.StickyBitTrigger{
				StickyBit: newDesc.GetStickyBit(),
			},
		},
	})
	return txn.Run(ctx, b)
}

func (r *Replica) adminSplitWithDescriptor(
	ctx context.Context,
	args roachpb.AdminSplitRequest,
	desc *roachpb.RangeDescriptor,
	delayable bool,
	reason string,
) (roachpb.AdminSplitResponse, error) {
	__antithesis_instrumentation__.Notify(115909)
	var err error
	var reply roachpb.AdminSplitResponse

	desc, err = r.maybeLeaveAtomicChangeReplicas(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(115916)
		return reply, err
	} else {
		__antithesis_instrumentation__.Notify(115917)
	}
	__antithesis_instrumentation__.Notify(115910)

	log.Event(ctx, "split begins")
	var splitKey roachpb.RKey
	{
		__antithesis_instrumentation__.Notify(115918)
		var foundSplitKey roachpb.Key
		if len(args.SplitKey) == 0 {
			__antithesis_instrumentation__.Notify(115924)

			var err error
			targetSize := r.GetMaxBytes() / 2
			foundSplitKey, err = storage.MVCCFindSplitKey(
				ctx, r.store.engine, desc.StartKey, desc.EndKey, targetSize)
			if err != nil {
				__antithesis_instrumentation__.Notify(115926)
				return reply, errors.Wrap(err, "unable to determine split key")
			} else {
				__antithesis_instrumentation__.Notify(115927)
			}
			__antithesis_instrumentation__.Notify(115925)
			if foundSplitKey == nil {
				__antithesis_instrumentation__.Notify(115928)

				return reply, unsplittableRangeError{}
			} else {
				__antithesis_instrumentation__.Notify(115929)
			}
		} else {
			__antithesis_instrumentation__.Notify(115930)

			if !kvserverbase.ContainsKey(desc, args.Key) {
				__antithesis_instrumentation__.Notify(115932)
				ri := r.GetRangeInfo(ctx)
				return reply, roachpb.NewRangeKeyMismatchErrorWithCTPolicy(ctx, args.Key, args.Key, desc, &ri.Lease, ri.ClosedTimestampPolicy)
			} else {
				__antithesis_instrumentation__.Notify(115933)
			}
			__antithesis_instrumentation__.Notify(115931)
			foundSplitKey = args.SplitKey
		}
		__antithesis_instrumentation__.Notify(115919)

		if !kvserverbase.ContainsKey(desc, foundSplitKey) {
			__antithesis_instrumentation__.Notify(115934)
			return reply, errors.Errorf("requested split key %s out of bounds of %s", args.SplitKey, r)
		} else {
			__antithesis_instrumentation__.Notify(115935)
		}
		__antithesis_instrumentation__.Notify(115920)

		for _, k := range args.PredicateKeys {
			__antithesis_instrumentation__.Notify(115936)
			if !kvserverbase.ContainsKey(desc, k) {
				__antithesis_instrumentation__.Notify(115937)
				return reply, errors.Errorf("requested predicate key %s out of bounds of %s", k, r)
			} else {
				__antithesis_instrumentation__.Notify(115938)
			}
		}
		__antithesis_instrumentation__.Notify(115921)

		var err error
		splitKey, err = keys.Addr(foundSplitKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(115939)
			return reply, err
		} else {
			__antithesis_instrumentation__.Notify(115940)
		}
		__antithesis_instrumentation__.Notify(115922)
		if !splitKey.Equal(foundSplitKey) {
			__antithesis_instrumentation__.Notify(115941)
			return reply, errors.Errorf("cannot split range at range-local key %s", splitKey)
		} else {
			__antithesis_instrumentation__.Notify(115942)
		}
		__antithesis_instrumentation__.Notify(115923)
		if !storage.IsValidSplitKey(foundSplitKey) {
			__antithesis_instrumentation__.Notify(115943)
			return reply, errors.Errorf("cannot split range at key %s", splitKey)
		} else {
			__antithesis_instrumentation__.Notify(115944)
		}
	}
	__antithesis_instrumentation__.Notify(115911)

	if desc.StartKey.Equal(splitKey) {
		__antithesis_instrumentation__.Notify(115945)
		if len(args.SplitKey) == 0 {
			__antithesis_instrumentation__.Notify(115948)
			log.Fatal(ctx, "MVCCFindSplitKey returned start key of range")
		} else {
			__antithesis_instrumentation__.Notify(115949)
		}
		__antithesis_instrumentation__.Notify(115946)
		log.Event(ctx, "range already split")

		if desc.GetStickyBit().Less(args.ExpirationTime) {
			__antithesis_instrumentation__.Notify(115950)
			err := r.store.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(115953)
				return splitTxnStickyUpdateAttempt(ctx, txn, desc, args.ExpirationTime)
			})
			__antithesis_instrumentation__.Notify(115951)

			if ok, actualDesc := maybeDescriptorChangedError(desc, err); ok {
				__antithesis_instrumentation__.Notify(115954)

				err = &benignError{wrapDescChangedError(err, desc, actualDesc)}
			} else {
				__antithesis_instrumentation__.Notify(115955)
			}
			__antithesis_instrumentation__.Notify(115952)
			return reply, err
		} else {
			__antithesis_instrumentation__.Notify(115956)
		}
		__antithesis_instrumentation__.Notify(115947)
		return reply, nil
	} else {
		__antithesis_instrumentation__.Notify(115957)
	}
	__antithesis_instrumentation__.Notify(115912)
	log.Event(ctx, "found split key")

	rightRangeID, err := r.store.AllocateRangeID(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(115958)
		return reply, errors.Wrap(err, "unable to allocate range id for right hand side")
	} else {
		__antithesis_instrumentation__.Notify(115959)
	}
	__antithesis_instrumentation__.Notify(115913)

	var extra string
	if delayable {
		__antithesis_instrumentation__.Notify(115960)
		extra += maybeDelaySplitToAvoidSnapshot(ctx, (*splitDelayHelper)(r))
	} else {
		__antithesis_instrumentation__.Notify(115961)
	}
	__antithesis_instrumentation__.Notify(115914)
	extra += splitSnapshotWarningStr(r.RangeID, r.RaftStatus())

	log.Infof(ctx, "initiating a split of this range at key %s [r%d] (%s)%s",
		splitKey.StringWithDirs(nil, 50), rightRangeID, reason, extra)

	if err := r.store.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(115962)
		return splitTxnAttempt(ctx, r.store, txn, rightRangeID, splitKey, args.ExpirationTime, desc)
	}); err != nil {
		__antithesis_instrumentation__.Notify(115963)

		if ok, actualDesc := maybeDescriptorChangedError(desc, err); ok {
			__antithesis_instrumentation__.Notify(115965)

			err = &benignError{wrapDescChangedError(err, desc, actualDesc)}
		} else {
			__antithesis_instrumentation__.Notify(115966)
		}
		__antithesis_instrumentation__.Notify(115964)
		return reply, errors.Wrapf(err, "split at key %s failed", splitKey)
	} else {
		__antithesis_instrumentation__.Notify(115967)
	}
	__antithesis_instrumentation__.Notify(115915)
	return reply, nil
}

func (r *Replica) AdminUnsplit(
	ctx context.Context, args roachpb.AdminUnsplitRequest, reason string,
) (roachpb.AdminUnsplitResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(115968)
	var reply roachpb.AdminUnsplitResponse
	err := r.executeAdminCommandWithDescriptor(ctx, func(desc *roachpb.RangeDescriptor) error {
		__antithesis_instrumentation__.Notify(115970)
		var err error
		reply, err = r.adminUnsplitWithDescriptor(ctx, args, desc, reason)
		return err
	})
	__antithesis_instrumentation__.Notify(115969)
	return reply, err
}

func (r *Replica) adminUnsplitWithDescriptor(
	ctx context.Context,
	args roachpb.AdminUnsplitRequest,
	desc *roachpb.RangeDescriptor,
	reason string,
) (roachpb.AdminUnsplitResponse, error) {
	__antithesis_instrumentation__.Notify(115971)
	var reply roachpb.AdminUnsplitResponse
	if !bytes.Equal(desc.StartKey.AsRawKey(), args.Header().Key) {
		__antithesis_instrumentation__.Notify(115975)
		return reply, errors.Errorf("key %s is not the start of a range", args.Header().Key)
	} else {
		__antithesis_instrumentation__.Notify(115976)
	}
	__antithesis_instrumentation__.Notify(115972)

	if desc.GetStickyBit().IsEmpty() {
		__antithesis_instrumentation__.Notify(115977)
		return reply, nil
	} else {
		__antithesis_instrumentation__.Notify(115978)
	}
	__antithesis_instrumentation__.Notify(115973)

	if err := r.store.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(115979)
		_, dbDescValue, err := conditionalGetDescValueFromDB(
			ctx, txn, desc.StartKey, false, checkDescsEqual(desc))
		if err != nil {
			__antithesis_instrumentation__.Notify(115983)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115984)
		}
		__antithesis_instrumentation__.Notify(115980)

		newDesc := *desc

		newDesc.StickyBit = nil
		descKey := keys.RangeDescriptorKey(newDesc.StartKey)

		b := txn.NewBatch()
		if err := updateRangeDescriptor(ctx, b, descKey, dbDescValue, &newDesc); err != nil {
			__antithesis_instrumentation__.Notify(115985)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115986)
		}
		__antithesis_instrumentation__.Notify(115981)
		if err := updateRangeAddressing(b, &newDesc); err != nil {
			__antithesis_instrumentation__.Notify(115987)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115988)
		}
		__antithesis_instrumentation__.Notify(115982)

		b.AddRawRequest(&roachpb.EndTxnRequest{
			Commit: true,
			InternalCommitTrigger: &roachpb.InternalCommitTrigger{
				StickyBitTrigger: &roachpb.StickyBitTrigger{

					StickyBit: hlc.Timestamp{},
				},
			},
		})
		return txn.Run(ctx, b)
	}); err != nil {
		__antithesis_instrumentation__.Notify(115989)

		if ok, actualDesc := maybeDescriptorChangedError(desc, err); ok {
			__antithesis_instrumentation__.Notify(115991)

			err = &benignError{wrapDescChangedError(err, desc, actualDesc)}
		} else {
			__antithesis_instrumentation__.Notify(115992)
		}
		__antithesis_instrumentation__.Notify(115990)
		return reply, err
	} else {
		__antithesis_instrumentation__.Notify(115993)
	}
	__antithesis_instrumentation__.Notify(115974)
	return reply, nil
}

func (r *Replica) executeAdminCommandWithDescriptor(
	ctx context.Context, updateDesc func(*roachpb.RangeDescriptor) error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(115994)

	retryOpts := base.DefaultRetryOptions()

	retryOpts.RandomizationFactor = 0.5
	var lastErr error
	for retryable := retry.StartWithCtx(ctx, retryOpts); retryable.Next(); {
		__antithesis_instrumentation__.Notify(115996)

		if _, err := r.IsDestroyed(); err != nil {
			__antithesis_instrumentation__.Notify(115999)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(116000)
		}
		__antithesis_instrumentation__.Notify(115997)

		if _, pErr := r.redirectOnOrAcquireLease(ctx); pErr != nil {
			__antithesis_instrumentation__.Notify(116001)
			return pErr
		} else {
			__antithesis_instrumentation__.Notify(116002)
		}
		__antithesis_instrumentation__.Notify(115998)

		lastErr = updateDesc(r.Desc())

		if !errors.HasType(lastErr, (*roachpb.ConditionFailedError)(nil)) && func() bool {
			__antithesis_instrumentation__.Notify(116003)
			return !errors.HasType(lastErr, (*roachpb.AmbiguousResultError)(nil)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(116004)
			break
		} else {
			__antithesis_instrumentation__.Notify(116005)
		}
	}
	__antithesis_instrumentation__.Notify(115995)
	return roachpb.NewError(lastErr)
}

func (r *Replica) AdminMerge(
	ctx context.Context, args roachpb.AdminMergeRequest, reason string,
) (roachpb.AdminMergeResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(116006)
	var reply roachpb.AdminMergeResponse

	runMergeTxn := func(txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(116008)
		log.Event(ctx, "merge txn begins")
		txn.SetDebugName(mergeTxnName)

		if err := txn.DisablePipelining(); err != nil {
			__antithesis_instrumentation__.Notify(116029)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116030)
		}
		__antithesis_instrumentation__.Notify(116009)

		origLeftDesc := r.Desc()
		if origLeftDesc.EndKey.Equal(roachpb.RKeyMax) {
			__antithesis_instrumentation__.Notify(116031)

			return errors.New("cannot merge final range")
		} else {
			__antithesis_instrumentation__.Notify(116032)
		}
		__antithesis_instrumentation__.Notify(116010)

		_, dbOrigLeftDescValue, err := conditionalGetDescValueFromDB(
			ctx, txn, origLeftDesc.StartKey, true, checkDescsEqual(origLeftDesc))
		if err != nil {
			__antithesis_instrumentation__.Notify(116033)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116034)
		}
		__antithesis_instrumentation__.Notify(116011)

		var rightDesc roachpb.RangeDescriptor
		rightDescKey := keys.RangeDescriptorKey(origLeftDesc.EndKey)
		dbRightDescKV, err := txn.GetForUpdate(ctx, rightDescKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(116035)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116036)
		}
		__antithesis_instrumentation__.Notify(116012)
		if err := dbRightDescKV.ValueProto(&rightDesc); err != nil {
			__antithesis_instrumentation__.Notify(116037)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116038)
		}
		__antithesis_instrumentation__.Notify(116013)

		if !bytes.Equal(origLeftDesc.EndKey, rightDesc.StartKey) {
			__antithesis_instrumentation__.Notify(116039)

			return errors.Errorf("ranges are not adjacent; %s != %s", origLeftDesc.EndKey, rightDesc.StartKey)
		} else {
			__antithesis_instrumentation__.Notify(116040)
		}
		__antithesis_instrumentation__.Notify(116014)

		lReplicas, rReplicas := origLeftDesc.Replicas(), rightDesc.Replicas()

		if len(lReplicas.VoterFullAndNonVoterDescriptors()) != len(lReplicas.Descriptors()) {
			__antithesis_instrumentation__.Notify(116041)
			return errors.Errorf("cannot merge ranges when lhs is in a joint state or has learners: %s",
				lReplicas)
		} else {
			__antithesis_instrumentation__.Notify(116042)
		}
		__antithesis_instrumentation__.Notify(116015)
		if len(rReplicas.VoterFullAndNonVoterDescriptors()) != len(rReplicas.Descriptors()) {
			__antithesis_instrumentation__.Notify(116043)
			return errors.Errorf("cannot merge ranges when rhs is in a joint state or has learners: %s",
				rReplicas)
		} else {
			__antithesis_instrumentation__.Notify(116044)
		}
		__antithesis_instrumentation__.Notify(116016)
		if !replicasCollocated(lReplicas.Descriptors(), rReplicas.Descriptors()) {
			__antithesis_instrumentation__.Notify(116045)
			return errors.Errorf("ranges not collocated; %s != %s", lReplicas, rReplicas)
		} else {
			__antithesis_instrumentation__.Notify(116046)
		}
		__antithesis_instrumentation__.Notify(116017)

		if err := waitForReplicasInit(
			ctx, r.store.cfg.NodeDialer, origLeftDesc.RangeID, origLeftDesc.Replicas().Descriptors(),
		); err != nil {
			__antithesis_instrumentation__.Notify(116047)
			return errors.Wrap(err, "waiting for all left-hand replicas to initialize")
		} else {
			__antithesis_instrumentation__.Notify(116048)
		}
		__antithesis_instrumentation__.Notify(116018)

		if err := waitForReplicasInit(
			ctx, r.store.cfg.NodeDialer, rightDesc.RangeID, rightDesc.Replicas().Descriptors(),
		); err != nil {
			__antithesis_instrumentation__.Notify(116049)
			return errors.Wrap(err, "waiting for all right-hand replicas to initialize")
		} else {
			__antithesis_instrumentation__.Notify(116050)
		}
		__antithesis_instrumentation__.Notify(116019)

		mergeReplicas := lReplicas.Descriptors()

		updatedLeftDesc := *origLeftDesc

		if updatedLeftDesc.Generation < rightDesc.Generation {
			__antithesis_instrumentation__.Notify(116051)
			updatedLeftDesc.Generation = rightDesc.Generation
		} else {
			__antithesis_instrumentation__.Notify(116052)
		}
		__antithesis_instrumentation__.Notify(116020)
		updatedLeftDesc.IncrementGeneration()
		updatedLeftDesc.EndKey = rightDesc.EndKey
		log.Infof(ctx, "initiating a merge of %s into this range (%s)", &rightDesc, reason)

		if err := r.store.logMerge(ctx, txn, updatedLeftDesc, rightDesc); err != nil {
			__antithesis_instrumentation__.Notify(116053)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116054)
		}
		__antithesis_instrumentation__.Notify(116021)

		b := txn.NewBatch()

		if err := mergeRangeAddressing(b, origLeftDesc, &updatedLeftDesc); err != nil {
			__antithesis_instrumentation__.Notify(116055)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116056)
		}
		__antithesis_instrumentation__.Notify(116022)

		leftDescKey := keys.RangeDescriptorKey(updatedLeftDesc.StartKey)
		if err := updateRangeDescriptor(ctx, b, leftDescKey,
			dbOrigLeftDescValue,
			&updatedLeftDesc,
		); err != nil {
			__antithesis_instrumentation__.Notify(116057)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116058)
		}
		__antithesis_instrumentation__.Notify(116023)

		if err := updateRangeDescriptor(ctx, b, rightDescKey,
			dbRightDescKV.Value.TagAndDataBytes(),
			nil,
		); err != nil {
			__antithesis_instrumentation__.Notify(116059)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116060)
		}
		__antithesis_instrumentation__.Notify(116024)

		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(116061)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116062)
		}
		__antithesis_instrumentation__.Notify(116025)

		br, pErr := kv.SendWrapped(ctx, r.store.DB().NonTransactionalSender(),
			&roachpb.SubsumeRequest{
				RequestHeader: roachpb.RequestHeader{Key: rightDesc.StartKey.AsRawKey()},
				LeftDesc:      *origLeftDesc,
				RightDesc:     rightDesc,
			})
		if pErr != nil {
			__antithesis_instrumentation__.Notify(116063)
			return pErr.GoError()
		} else {
			__antithesis_instrumentation__.Notify(116064)
		}
		__antithesis_instrumentation__.Notify(116026)
		rhsSnapshotRes := br.(*roachpb.SubsumeResponse)

		err = contextutil.RunWithTimeout(ctx, "waiting for merge application", mergeApplicationTimeout,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(116065)
				return waitForApplication(ctx, r.store.cfg.NodeDialer, rightDesc.RangeID, mergeReplicas,
					rhsSnapshotRes.LeaseAppliedIndex)
			})
		__antithesis_instrumentation__.Notify(116027)
		if err != nil {
			__antithesis_instrumentation__.Notify(116066)
			return errors.Wrap(err, "waiting for all right-hand replicas to catch up")
		} else {
			__antithesis_instrumentation__.Notify(116067)
		}
		__antithesis_instrumentation__.Notify(116028)

		b = txn.NewBatch()
		b.AddRawRequest(&roachpb.EndTxnRequest{
			Commit: true,
			InternalCommitTrigger: &roachpb.InternalCommitTrigger{
				MergeTrigger: &roachpb.MergeTrigger{
					LeftDesc:             updatedLeftDesc,
					RightDesc:            rightDesc,
					RightMVCCStats:       rhsSnapshotRes.MVCCStats,
					FreezeStart:          rhsSnapshotRes.FreezeStart,
					RightClosedTimestamp: rhsSnapshotRes.ClosedTimestamp,
					RightReadSummary:     rhsSnapshotRes.ReadSummary,
				},
			},
		})
		log.Event(ctx, "attempting commit")
		return txn.Run(ctx, b)
	}
	__antithesis_instrumentation__.Notify(116007)

	for {
		__antithesis_instrumentation__.Notify(116068)
		txn := kv.NewTxn(ctx, r.store.DB(), r.NodeID())
		err := runMergeTxn(txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(116070)
			log.VEventf(ctx, 2, "merge txn failed: %s", err)
			txn.CleanupOnError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(116071)
		}
		__antithesis_instrumentation__.Notify(116069)
		if !errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) {
			__antithesis_instrumentation__.Notify(116072)
			if err != nil {
				__antithesis_instrumentation__.Notify(116074)
				return reply, roachpb.NewErrorf("merge failed: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(116075)
			}
			__antithesis_instrumentation__.Notify(116073)
			return reply, nil
		} else {
			__antithesis_instrumentation__.Notify(116076)
		}
	}
}

func waitForApplication(
	ctx context.Context,
	dialer *nodedialer.Dialer,
	rangeID roachpb.RangeID,
	replicas []roachpb.ReplicaDescriptor,
	leaseIndex uint64,
) error {
	__antithesis_instrumentation__.Notify(116077)
	if dialer == nil && func() bool {
		__antithesis_instrumentation__.Notify(116080)
		return len(replicas) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(116081)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(116082)
	}
	__antithesis_instrumentation__.Notify(116078)
	g := ctxgroup.WithContext(ctx)
	for _, repl := range replicas {
		__antithesis_instrumentation__.Notify(116083)
		repl := repl
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(116084)
			conn, err := dialer.Dial(ctx, repl.NodeID, rpc.DefaultClass)
			if err != nil {
				__antithesis_instrumentation__.Notify(116086)
				return errors.Wrapf(err, "could not dial n%d", repl.NodeID)
			} else {
				__antithesis_instrumentation__.Notify(116087)
			}
			__antithesis_instrumentation__.Notify(116085)
			_, err = NewPerReplicaClient(conn).WaitForApplication(ctx, &WaitForApplicationRequest{
				StoreRequestHeader: StoreRequestHeader{NodeID: repl.NodeID, StoreID: repl.StoreID},
				RangeID:            rangeID,
				LeaseIndex:         leaseIndex,
			})
			return err
		})
	}
	__antithesis_instrumentation__.Notify(116079)
	return g.Wait()
}

func waitForReplicasInit(
	ctx context.Context,
	dialer *nodedialer.Dialer,
	rangeID roachpb.RangeID,
	replicas []roachpb.ReplicaDescriptor,
) error {
	__antithesis_instrumentation__.Notify(116088)
	if dialer == nil && func() bool {
		__antithesis_instrumentation__.Notify(116090)
		return len(replicas) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(116091)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(116092)
	}
	__antithesis_instrumentation__.Notify(116089)
	return contextutil.RunWithTimeout(ctx, "wait for replicas init", 5*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(116093)
		g := ctxgroup.WithContext(ctx)
		for _, repl := range replicas {
			__antithesis_instrumentation__.Notify(116095)
			repl := repl
			g.GoCtx(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(116096)
				conn, err := dialer.Dial(ctx, repl.NodeID, rpc.DefaultClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(116098)
					return errors.Wrapf(err, "could not dial n%d", repl.NodeID)
				} else {
					__antithesis_instrumentation__.Notify(116099)
				}
				__antithesis_instrumentation__.Notify(116097)
				_, err = NewPerReplicaClient(conn).WaitForReplicaInit(ctx, &WaitForReplicaInitRequest{
					StoreRequestHeader: StoreRequestHeader{NodeID: repl.NodeID, StoreID: repl.StoreID},
					RangeID:            rangeID,
				})
				return err
			})
		}
		__antithesis_instrumentation__.Notify(116094)
		return g.Wait()
	})
}

func (r *Replica) ChangeReplicas(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	priority kvserverpb.SnapshotRequest_Priority,
	reason kvserverpb.RangeLogEventReason,
	details string,
	chgs roachpb.ReplicationChanges,
) (updatedDesc *roachpb.RangeDescriptor, _ error) {
	__antithesis_instrumentation__.Notify(116100)
	if desc == nil {
		__antithesis_instrumentation__.Notify(116103)

		return nil, errors.Errorf("%s: the current RangeDescriptor must not be nil", r)
	} else {
		__antithesis_instrumentation__.Notify(116104)
	}
	__antithesis_instrumentation__.Notify(116101)

	if knobs := r.store.TestingKnobs(); util.RaceEnabled && func() bool {
		__antithesis_instrumentation__.Notify(116105)
		return !knobs.DisableReplicateQueue == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(116106)
		return !knobs.AllowUnsynchronizedReplicationChanges == true
	}() == true {
		__antithesis_instrumentation__.Notify(116107)
		bq := r.store.replicateQueue.baseQueue
		bq.mu.Lock()
		disabled := bq.mu.disabled
		bq.mu.Unlock()
		if !disabled {
			__antithesis_instrumentation__.Notify(116108)
			return nil, errors.New("must disable replicate queue to use ChangeReplicas manually")
		} else {
			__antithesis_instrumentation__.Notify(116109)
		}
	} else {
		__antithesis_instrumentation__.Notify(116110)
	}
	__antithesis_instrumentation__.Notify(116102)
	return r.changeReplicasImpl(ctx, desc, priority, reason, details, chgs)
}

func (r *Replica) changeReplicasImpl(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	priority kvserverpb.SnapshotRequest_Priority,
	reason kvserverpb.RangeLogEventReason,
	details string,
	chgs roachpb.ReplicationChanges,
) (updatedDesc *roachpb.RangeDescriptor, _ error) {
	__antithesis_instrumentation__.Notify(116111)
	var err error

	desc, err = r.maybeLeaveAtomicChangeReplicas(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(116120)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(116121)
	}
	__antithesis_instrumentation__.Notify(116112)

	if err := validateReplicationChanges(desc, chgs); err != nil {
		__antithesis_instrumentation__.Notify(116122)
		return nil, errors.Mark(err, errMarkInvalidReplicationChange)
	} else {
		__antithesis_instrumentation__.Notify(116123)
	}
	__antithesis_instrumentation__.Notify(116113)
	targets := synthesizeTargetsByChangeType(chgs)

	swaps := getInternalChangesForExplicitPromotionsAndDemotions(targets.voterDemotions, targets.nonVoterPromotions)
	if len(swaps) > 0 {
		__antithesis_instrumentation__.Notify(116124)
		desc, err = execChangeReplicasTxn(ctx, desc, reason, details, swaps, changeReplicasTxnArgs{
			db:                                   r.store.DB(),
			liveAndDeadReplicas:                  r.store.allocator.storePool.liveAndDeadReplicas,
			logChange:                            r.store.logChange,
			testForceJointConfig:                 r.store.TestingKnobs().ReplicationAlwaysUseJointConfig,
			testAllowDangerousReplicationChanges: r.store.TestingKnobs().AllowDangerousReplicationChanges,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(116125)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116126)
		}
	} else {
		__antithesis_instrumentation__.Notify(116127)
	}
	__antithesis_instrumentation__.Notify(116114)

	if adds := targets.voterAdditions; len(adds) > 0 {
		__antithesis_instrumentation__.Notify(116128)

		_ = roachpb.ReplicaSet.LearnerDescriptors
		var err error
		desc, err = r.initializeRaftLearners(
			ctx, desc, priority, reason, details, adds, roachpb.LEARNER,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116130)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116131)
		}
		__antithesis_instrumentation__.Notify(116129)

		if fn := r.store.cfg.TestingKnobs.VoterAddStopAfterLearnerSnapshot; fn != nil && func() bool {
			__antithesis_instrumentation__.Notify(116132)
			return fn(adds) == true
		}() == true {
			__antithesis_instrumentation__.Notify(116133)
			return desc, nil
		} else {
			__antithesis_instrumentation__.Notify(116134)
		}
	} else {
		__antithesis_instrumentation__.Notify(116135)
	}
	__antithesis_instrumentation__.Notify(116115)

	if len(targets.voterAdditions)+len(targets.voterRemovals) > 0 {
		__antithesis_instrumentation__.Notify(116136)
		desc, err = r.execReplicationChangesForVoters(
			ctx, desc, reason, details,
			targets.voterAdditions, targets.voterRemovals,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116137)

			if _, err := r.maybeLeaveAtomicChangeReplicas(ctx, r.Desc()); err != nil {
				__antithesis_instrumentation__.Notify(116141)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(116142)
			}
			__antithesis_instrumentation__.Notify(116138)
			if fn := r.store.cfg.TestingKnobs.ReplicaAddSkipLearnerRollback; fn != nil && func() bool {
				__antithesis_instrumentation__.Notify(116143)
				return fn() == true
			}() == true {
				__antithesis_instrumentation__.Notify(116144)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(116145)
			}
			__antithesis_instrumentation__.Notify(116139)

			if adds := targets.voterAdditions; len(adds) > 0 {
				__antithesis_instrumentation__.Notify(116146)
				log.Infof(ctx, "could not promote %v to voter, rolling back: %v", adds, err)
				for _, target := range adds {
					__antithesis_instrumentation__.Notify(116147)
					r.tryRollbackRaftLearner(ctx, r.Desc(), target, reason, details)
				}
			} else {
				__antithesis_instrumentation__.Notify(116148)
			}
			__antithesis_instrumentation__.Notify(116140)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116149)
		}
	} else {
		__antithesis_instrumentation__.Notify(116150)
	}
	__antithesis_instrumentation__.Notify(116116)

	if adds := targets.nonVoterAdditions; len(adds) > 0 {
		__antithesis_instrumentation__.Notify(116151)

		desc, err = r.initializeRaftLearners(
			ctx, desc, priority, reason, details, adds, roachpb.NON_VOTER,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116153)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116154)
		}
		__antithesis_instrumentation__.Notify(116152)
		if fn := r.store.TestingKnobs().NonVoterAfterInitialization; fn != nil {
			__antithesis_instrumentation__.Notify(116155)
			fn()
		} else {
			__antithesis_instrumentation__.Notify(116156)
		}
	} else {
		__antithesis_instrumentation__.Notify(116157)
	}
	__antithesis_instrumentation__.Notify(116117)

	if removals := targets.nonVoterRemovals; len(removals) > 0 {
		__antithesis_instrumentation__.Notify(116158)
		for _, rem := range removals {
			__antithesis_instrumentation__.Notify(116159)
			iChgs := []internalReplicationChange{{target: rem, typ: internalChangeTypeRemove}}
			var err error
			desc, err = execChangeReplicasTxn(ctx, desc, reason, details, iChgs,
				changeReplicasTxnArgs{
					db:                                   r.store.DB(),
					liveAndDeadReplicas:                  r.store.allocator.storePool.liveAndDeadReplicas,
					logChange:                            r.store.logChange,
					testForceJointConfig:                 r.store.TestingKnobs().ReplicationAlwaysUseJointConfig,
					testAllowDangerousReplicationChanges: r.store.TestingKnobs().AllowDangerousReplicationChanges,
				})
			if err != nil {
				__antithesis_instrumentation__.Notify(116160)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(116161)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(116162)
	}
	__antithesis_instrumentation__.Notify(116118)

	if len(targets.voterDemotions) > 0 {
		__antithesis_instrumentation__.Notify(116163)

		return r.maybeLeaveAtomicChangeReplicasAndRemoveLearners(ctx, desc)
	} else {
		__antithesis_instrumentation__.Notify(116164)
	}
	__antithesis_instrumentation__.Notify(116119)
	return desc, nil
}

type targetsForReplicationChanges struct {
	voterDemotions, nonVoterPromotions  []roachpb.ReplicationTarget
	voterAdditions, voterRemovals       []roachpb.ReplicationTarget
	nonVoterAdditions, nonVoterRemovals []roachpb.ReplicationTarget
}

func synthesizeTargetsByChangeType(
	chgs roachpb.ReplicationChanges,
) (result targetsForReplicationChanges) {
	__antithesis_instrumentation__.Notify(116165)

	result.voterDemotions = intersectTargets(chgs.VoterRemovals(), chgs.NonVoterAdditions())
	result.nonVoterPromotions = intersectTargets(chgs.NonVoterRemovals(), chgs.VoterAdditions())

	result.voterAdditions = subtractTargets(chgs.VoterAdditions(), chgs.NonVoterRemovals())
	result.voterRemovals = subtractTargets(chgs.VoterRemovals(), chgs.NonVoterAdditions())
	result.nonVoterAdditions = subtractTargets(chgs.NonVoterAdditions(), chgs.VoterRemovals())
	result.nonVoterRemovals = subtractTargets(chgs.NonVoterRemovals(), chgs.VoterAdditions())

	return result
}

func (r *Replica) maybeTransferLeaseDuringLeaveJoint(
	ctx context.Context, desc *roachpb.RangeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(116166)
	voters := desc.Replicas().VoterDescriptors()

	beingRemoved := true
	voterIncomingTarget := roachpb.ReplicaDescriptor{}
	for _, v := range voters {
		__antithesis_instrumentation__.Notify(116171)
		if beingRemoved && func() bool {
			__antithesis_instrumentation__.Notify(116173)
			return v.ReplicaID == r.ReplicaID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(116174)
			beingRemoved = false
		} else {
			__antithesis_instrumentation__.Notify(116175)
		}
		__antithesis_instrumentation__.Notify(116172)
		if voterIncomingTarget == (roachpb.ReplicaDescriptor{}) && func() bool {
			__antithesis_instrumentation__.Notify(116176)
			return v.GetType() == roachpb.VOTER_INCOMING == true
		}() == true {
			__antithesis_instrumentation__.Notify(116177)
			voterIncomingTarget = v
		} else {
			__antithesis_instrumentation__.Notify(116178)
		}
	}
	__antithesis_instrumentation__.Notify(116167)
	if !beingRemoved {
		__antithesis_instrumentation__.Notify(116179)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(116180)
	}
	__antithesis_instrumentation__.Notify(116168)

	if voterIncomingTarget == (roachpb.ReplicaDescriptor{}) {
		__antithesis_instrumentation__.Notify(116181)

		log.Infof(ctx, "no VOTER_INCOMING to transfer lease to. This replica probably lost the lease,"+
			" but still thinks its the leaseholder. In this case the new leaseholder is expected to "+
			"complete LEAVE_JOINT. Range descriptor: %v", desc)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(116182)
	}
	__antithesis_instrumentation__.Notify(116169)
	log.VEventf(ctx, 5, "current leaseholder %v is being removed through an"+
		" atomic replication change. Transferring lease to %v", r.String(), voterIncomingTarget)
	err := r.store.DB().AdminTransferLease(ctx, r.startKey, voterIncomingTarget.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(116183)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116184)
	}
	__antithesis_instrumentation__.Notify(116170)
	log.VEventf(ctx, 5, "leaseholder transfer to %v complete", voterIncomingTarget)
	return nil
}

func (r *Replica) maybeLeaveAtomicChangeReplicas(
	ctx context.Context, desc *roachpb.RangeDescriptor,
) (rangeDesc *roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(116185)

	if !desc.Replicas().InAtomicReplicationChange() {
		__antithesis_instrumentation__.Notify(116188)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(116189)
	}
	__antithesis_instrumentation__.Notify(116186)

	log.Eventf(ctx, "transitioning out of joint configuration %s", desc)

	if err := r.maybeTransferLeaseDuringLeaveJoint(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(116190)
		return desc, err
	} else {
		__antithesis_instrumentation__.Notify(116191)
	}
	__antithesis_instrumentation__.Notify(116187)

	s := r.store
	return execChangeReplicasTxn(
		ctx, desc, kvserverpb.ReasonUnknown, "", nil,
		changeReplicasTxnArgs{
			db:                                   s.DB(),
			liveAndDeadReplicas:                  s.allocator.storePool.liveAndDeadReplicas,
			logChange:                            s.logChange,
			testForceJointConfig:                 s.TestingKnobs().ReplicationAlwaysUseJointConfig,
			testAllowDangerousReplicationChanges: s.TestingKnobs().AllowDangerousReplicationChanges,
		})
}

func (r *Replica) maybeLeaveAtomicChangeReplicasAndRemoveLearners(
	ctx context.Context, desc *roachpb.RangeDescriptor,
) (rangeDesc *roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(116192)
	desc, err = r.maybeLeaveAtomicChangeReplicas(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(116197)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(116198)
	}
	__antithesis_instrumentation__.Notify(116193)

	learners := desc.Replicas().LearnerDescriptors()
	if len(learners) == 0 {
		__antithesis_instrumentation__.Notify(116199)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(116200)
	}
	__antithesis_instrumentation__.Notify(116194)
	targets := make([]roachpb.ReplicationTarget, len(learners))
	for i := range learners {
		__antithesis_instrumentation__.Notify(116201)
		targets[i].NodeID = learners[i].NodeID
		targets[i].StoreID = learners[i].StoreID
	}
	__antithesis_instrumentation__.Notify(116195)
	log.VEventf(ctx, 2, `removing learner replicas %v from %v`, targets, desc)

	origDesc := desc
	store := r.store
	for _, target := range targets {
		__antithesis_instrumentation__.Notify(116202)
		var err error
		desc, err = execChangeReplicasTxn(
			ctx, desc, kvserverpb.ReasonAbandonedLearner, "",
			[]internalReplicationChange{{target: target, typ: internalChangeTypeRemove}},
			changeReplicasTxnArgs{db: store.DB(),
				liveAndDeadReplicas:                  store.allocator.storePool.liveAndDeadReplicas,
				logChange:                            store.logChange,
				testForceJointConfig:                 store.TestingKnobs().ReplicationAlwaysUseJointConfig,
				testAllowDangerousReplicationChanges: store.TestingKnobs().AllowDangerousReplicationChanges,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116203)
			return nil, errors.Wrapf(err, `removing learners from %s`, origDesc)
		} else {
			__antithesis_instrumentation__.Notify(116204)
		}
	}
	__antithesis_instrumentation__.Notify(116196)
	return desc, nil
}

func validateAdditionsPerStore(
	desc *roachpb.RangeDescriptor, chgsByStoreID changesByStoreID,
) error {
	__antithesis_instrumentation__.Notify(116205)
	for storeID, chgs := range chgsByStoreID {
		__antithesis_instrumentation__.Notify(116207)
		for _, chg := range chgs {
			__antithesis_instrumentation__.Notify(116208)
			if chg.ChangeType.IsRemoval() {
				__antithesis_instrumentation__.Notify(116211)
				continue
			} else {
				__antithesis_instrumentation__.Notify(116212)
			}
			__antithesis_instrumentation__.Notify(116209)

			replDesc, found := desc.GetReplicaDescriptor(storeID)
			if !found {
				__antithesis_instrumentation__.Notify(116213)

				continue
			} else {
				__antithesis_instrumentation__.Notify(116214)
			}
			__antithesis_instrumentation__.Notify(116210)
			switch t := replDesc.GetType(); t {
			case roachpb.LEARNER:
				__antithesis_instrumentation__.Notify(116215)

				return errors.AssertionFailedf(
					"trying to add(%+v) to a store that already has a %s", chg, t)
			case roachpb.VOTER_FULL:
				__antithesis_instrumentation__.Notify(116216)
				if chg.ChangeType == roachpb.ADD_VOTER {
					__antithesis_instrumentation__.Notify(116219)
					return errors.AssertionFailedf(
						"trying to add a voter to a store that already has a %s", t)
				} else {
					__antithesis_instrumentation__.Notify(116220)
				}
			case roachpb.NON_VOTER:
				__antithesis_instrumentation__.Notify(116217)
				if chg.ChangeType == roachpb.ADD_NON_VOTER {
					__antithesis_instrumentation__.Notify(116221)
					return errors.AssertionFailedf(
						"trying to add a non-voter to a store that already has a %s", t)
				} else {
					__antithesis_instrumentation__.Notify(116222)
				}
			default:
				__antithesis_instrumentation__.Notify(116218)
				return errors.AssertionFailedf("store(%d) being added to already contains a"+
					" replica of an unexpected type: %s", storeID, t)
			}
		}
	}
	__antithesis_instrumentation__.Notify(116206)
	return nil
}

func validateRemovals(desc *roachpb.RangeDescriptor, chgsByStoreID changesByStoreID) error {
	__antithesis_instrumentation__.Notify(116223)
	for storeID, chgs := range chgsByStoreID {
		__antithesis_instrumentation__.Notify(116225)
		for _, chg := range chgs {
			__antithesis_instrumentation__.Notify(116226)
			if chg.ChangeType.IsAddition() {
				__antithesis_instrumentation__.Notify(116229)
				continue
			} else {
				__antithesis_instrumentation__.Notify(116230)
			}
			__antithesis_instrumentation__.Notify(116227)
			replDesc, found := desc.GetReplicaDescriptor(storeID)
			if !found {
				__antithesis_instrumentation__.Notify(116231)
				return errors.AssertionFailedf("trying to remove a replica that doesn't exist: %+v", chg)
			} else {
				__antithesis_instrumentation__.Notify(116232)
			}
			__antithesis_instrumentation__.Notify(116228)

			switch t := replDesc.GetType(); t {
			case roachpb.VOTER_FULL, roachpb.LEARNER:
				__antithesis_instrumentation__.Notify(116233)
				if chg.ChangeType != roachpb.REMOVE_VOTER {
					__antithesis_instrumentation__.Notify(116236)
					return errors.AssertionFailedf("type of replica being removed (%s) does not match"+
						" expectation for change: %+v", t, chg)
				} else {
					__antithesis_instrumentation__.Notify(116237)
				}
			case roachpb.NON_VOTER:
				__antithesis_instrumentation__.Notify(116234)
				if chg.ChangeType != roachpb.REMOVE_NON_VOTER {
					__antithesis_instrumentation__.Notify(116238)
					return errors.AssertionFailedf("type of replica being removed (%s) does not match"+
						" expectation for change: %+v", t, chg)
				} else {
					__antithesis_instrumentation__.Notify(116239)
				}
			default:
				__antithesis_instrumentation__.Notify(116235)
				return errors.AssertionFailedf("unexpected replica type for removal %+v: %s", chg, t)
			}
		}
	}
	__antithesis_instrumentation__.Notify(116224)
	return nil
}

func validatePromotionsAndDemotions(
	desc *roachpb.RangeDescriptor, chgsByStoreID changesByStoreID,
) error {
	__antithesis_instrumentation__.Notify(116240)
	for storeID, chgs := range chgsByStoreID {
		__antithesis_instrumentation__.Notify(116242)
		replDesc, found := desc.GetReplicaDescriptor(storeID)
		switch len(chgs) {
		case 0:
			__antithesis_instrumentation__.Notify(116243)
			continue
		case 1:
			__antithesis_instrumentation__.Notify(116244)
			if chgs[0].ChangeType.IsAddition() {
				__antithesis_instrumentation__.Notify(116248)

				if found {
					__antithesis_instrumentation__.Notify(116249)
					return errors.AssertionFailedf("trying to add(%+v) to a store(%s) that already"+
						" has a replica(%s)", chgs[0], storeID, replDesc.GetType())
				} else {
					__antithesis_instrumentation__.Notify(116250)
				}
			} else {
				__antithesis_instrumentation__.Notify(116251)
			}
		case 2:
			__antithesis_instrumentation__.Notify(116245)
			c1, c2 := chgs[0], chgs[1]
			if !found {
				__antithesis_instrumentation__.Notify(116252)
				return errors.AssertionFailedf("found 2 changes(%+v) for a store(%d)"+
					" that has no replicas", chgs, storeID)
			} else {
				__antithesis_instrumentation__.Notify(116253)
			}
			__antithesis_instrumentation__.Notify(116246)
			if c1.ChangeType.IsAddition() && func() bool {
				__antithesis_instrumentation__.Notify(116254)
				return c2.ChangeType.IsRemoval() == true
			}() == true {
				__antithesis_instrumentation__.Notify(116255)

				isPromotion := c1.ChangeType == roachpb.ADD_VOTER && func() bool {
					__antithesis_instrumentation__.Notify(116256)
					return c2.ChangeType == roachpb.REMOVE_NON_VOTER == true
				}() == true
				isDemotion := c1.ChangeType == roachpb.ADD_NON_VOTER && func() bool {
					__antithesis_instrumentation__.Notify(116257)
					return c2.ChangeType == roachpb.REMOVE_VOTER == true
				}() == true
				if !(isPromotion || func() bool {
					__antithesis_instrumentation__.Notify(116258)
					return isDemotion == true
				}() == true) {
					__antithesis_instrumentation__.Notify(116259)
					return errors.AssertionFailedf("trying to add-remove the same replica(%s):"+
						" %+v", replDesc.GetType(), chgs)
				} else {
					__antithesis_instrumentation__.Notify(116260)
				}
			} else {
				__antithesis_instrumentation__.Notify(116261)

				return errors.AssertionFailedf("the only permissible order of operations within a"+
					" store is add-remove; got %+v", chgs)
			}
		default:
			__antithesis_instrumentation__.Notify(116247)
			return errors.AssertionFailedf("more than 2 changes referring to the same store: %+v", chgs)
		}
	}
	__antithesis_instrumentation__.Notify(116241)
	return nil
}

func validateOneReplicaPerNode(desc *roachpb.RangeDescriptor, chgsByNodeID changesByNodeID) error {
	__antithesis_instrumentation__.Notify(116262)
	replsByNodeID := make(map[roachpb.NodeID]int)
	for _, repl := range desc.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(116265)
		replsByNodeID[repl.NodeID]++
	}
	__antithesis_instrumentation__.Notify(116263)

	for nodeID, chgs := range chgsByNodeID {
		__antithesis_instrumentation__.Notify(116266)
		if len(chgs) > 2 {
			__antithesis_instrumentation__.Notify(116268)
			return errors.AssertionFailedf("more than 2 changes for the same node(%d): %+v",
				nodeID, chgs)
		} else {
			__antithesis_instrumentation__.Notify(116269)
		}
		__antithesis_instrumentation__.Notify(116267)
		switch replsByNodeID[nodeID] {
		case 0:
			__antithesis_instrumentation__.Notify(116270)

			if len(chgs) > 1 {
				__antithesis_instrumentation__.Notify(116274)
				return errors.AssertionFailedf("unexpected set of changes(%+v) for node %d, which has"+
					" no existing replicas for the range", chgs, nodeID)
			} else {
				__antithesis_instrumentation__.Notify(116275)
			}
		case 1:
			__antithesis_instrumentation__.Notify(116271)

			switch n := len(chgs); n {
			case 1:
				__antithesis_instrumentation__.Notify(116276)

				if !chgs[0].ChangeType.IsRemoval() && func() bool {
					__antithesis_instrumentation__.Notify(116279)
					return len(desc.Replicas().Descriptors()) > 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(116280)
					return errors.AssertionFailedf("node %d already has a replica; only valid actions"+
						" are a removal or a rebalance(add/remove); got %+v", nodeID, chgs)
				} else {
					__antithesis_instrumentation__.Notify(116281)
				}
			case 2:
				__antithesis_instrumentation__.Notify(116277)

				c1, c2 := chgs[0], chgs[1]
				if !(c1.ChangeType.IsAddition() && func() bool {
					__antithesis_instrumentation__.Notify(116282)
					return c2.ChangeType.IsRemoval() == true
				}() == true) {
					__antithesis_instrumentation__.Notify(116283)
					return errors.AssertionFailedf("node %d already has a replica; only valid actions"+
						" are a removal or a rebalance(add/remove); got %+v", nodeID, chgs)
				} else {
					__antithesis_instrumentation__.Notify(116284)
				}
			default:
				__antithesis_instrumentation__.Notify(116278)
				panic(fmt.Sprintf("unexpected number of changes for node %d: %+v", nodeID, chgs))
			}
		case 2:
			__antithesis_instrumentation__.Notify(116272)

			if !(len(chgs) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(116285)
				return chgs[0].ChangeType.IsRemoval() == true
			}() == true) {
				__antithesis_instrumentation__.Notify(116286)
				return errors.AssertionFailedf("node %d has 2 replicas, expected exactly one of them"+
					" to be removed; got %+v", nodeID, chgs)
			} else {
				__antithesis_instrumentation__.Notify(116287)
			}
		default:
			__antithesis_instrumentation__.Notify(116273)
			return errors.AssertionFailedf("node %d unexpectedly has more than 2 replicas: %s",
				nodeID, desc.Replicas().Descriptors())
		}
	}
	__antithesis_instrumentation__.Notify(116264)
	return nil
}

func validateReplicationChanges(
	desc *roachpb.RangeDescriptor, chgs roachpb.ReplicationChanges,
) error {
	__antithesis_instrumentation__.Notify(116288)
	chgsByStoreID := getChangesByStoreID(chgs)
	chgsByNodeID := getChangesByNodeID(chgs)

	if err := validateAdditionsPerStore(desc, chgsByStoreID); err != nil {
		__antithesis_instrumentation__.Notify(116293)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116294)
	}
	__antithesis_instrumentation__.Notify(116289)
	if err := validateRemovals(desc, chgsByStoreID); err != nil {
		__antithesis_instrumentation__.Notify(116295)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116296)
	}
	__antithesis_instrumentation__.Notify(116290)
	if err := validatePromotionsAndDemotions(desc, chgsByStoreID); err != nil {
		__antithesis_instrumentation__.Notify(116297)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116298)
	}
	__antithesis_instrumentation__.Notify(116291)
	if err := validateOneReplicaPerNode(desc, chgsByNodeID); err != nil {
		__antithesis_instrumentation__.Notify(116299)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116300)
	}
	__antithesis_instrumentation__.Notify(116292)

	return nil
}

type changesByStoreID map[roachpb.StoreID][]roachpb.ReplicationChange

type changesByNodeID map[roachpb.NodeID][]roachpb.ReplicationChange

func getChangesByStoreID(chgs roachpb.ReplicationChanges) changesByStoreID {
	__antithesis_instrumentation__.Notify(116301)
	chgsByStoreID := make(map[roachpb.StoreID][]roachpb.ReplicationChange, len(chgs))
	for _, chg := range chgs {
		__antithesis_instrumentation__.Notify(116303)
		if _, ok := chgsByStoreID[chg.Target.StoreID]; !ok {
			__antithesis_instrumentation__.Notify(116305)
			chgsByStoreID[chg.Target.StoreID] = make([]roachpb.ReplicationChange, 0, 2)
		} else {
			__antithesis_instrumentation__.Notify(116306)
		}
		__antithesis_instrumentation__.Notify(116304)
		chgsByStoreID[chg.Target.StoreID] = append(chgsByStoreID[chg.Target.StoreID], chg)
	}
	__antithesis_instrumentation__.Notify(116302)
	return chgsByStoreID
}

func getChangesByNodeID(chgs roachpb.ReplicationChanges) changesByNodeID {
	__antithesis_instrumentation__.Notify(116307)
	chgsByNodeID := make(map[roachpb.NodeID][]roachpb.ReplicationChange, len(chgs))
	for _, chg := range chgs {
		__antithesis_instrumentation__.Notify(116309)
		if _, ok := chgsByNodeID[chg.Target.NodeID]; !ok {
			__antithesis_instrumentation__.Notify(116311)
			chgsByNodeID[chg.Target.NodeID] = make([]roachpb.ReplicationChange, 0, 2)
		} else {
			__antithesis_instrumentation__.Notify(116312)
		}
		__antithesis_instrumentation__.Notify(116310)
		chgsByNodeID[chg.Target.NodeID] = append(chgsByNodeID[chg.Target.NodeID], chg)
	}
	__antithesis_instrumentation__.Notify(116308)
	return chgsByNodeID
}

func (r *Replica) initializeRaftLearners(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	priority kvserverpb.SnapshotRequest_Priority,
	reason kvserverpb.RangeLogEventReason,
	details string,
	targets []roachpb.ReplicationTarget,
	replicaType roachpb.ReplicaType,
) (afterDesc *roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(116313)
	var iChangeType internalChangeType
	switch replicaType {
	case roachpb.LEARNER:
		__antithesis_instrumentation__.Notify(116320)
		iChangeType = internalChangeTypeAddLearner
	case roachpb.NON_VOTER:
		__antithesis_instrumentation__.Notify(116321)
		iChangeType = internalChangeTypeAddNonVoter
	default:
		__antithesis_instrumentation__.Notify(116322)
		log.Fatalf(ctx, "unexpected replicaType %s", replicaType)
	}
	__antithesis_instrumentation__.Notify(116314)

	_ = (*raftSnapshotQueue)(nil).processRaftSnapshot

	releaseSnapshotLockFn := r.lockLearnerSnapshot(ctx, targets)
	defer releaseSnapshotLockFn()

	defer func() {
		__antithesis_instrumentation__.Notify(116323)
		if err != nil {
			__antithesis_instrumentation__.Notify(116324)
			log.Infof(
				ctx,
				"could not successfully add and upreplicate %s replica(s) on %s, rolling back: %v",
				replicaType,
				targets,
				err,
			)
			if fn := r.store.cfg.TestingKnobs.ReplicaAddSkipLearnerRollback; fn != nil && func() bool {
				__antithesis_instrumentation__.Notify(116326)
				return fn() == true
			}() == true {
				__antithesis_instrumentation__.Notify(116327)
				return
			} else {
				__antithesis_instrumentation__.Notify(116328)
			}
			__antithesis_instrumentation__.Notify(116325)

			for _, target := range targets {
				__antithesis_instrumentation__.Notify(116329)

				r.tryRollbackRaftLearner(ctx, r.Desc(), target, reason, details)
			}
		} else {
			__antithesis_instrumentation__.Notify(116330)
		}
	}()
	__antithesis_instrumentation__.Notify(116315)

	for _, target := range targets {
		__antithesis_instrumentation__.Notify(116331)
		iChgs := []internalReplicationChange{{target: target, typ: iChangeType}}
		var err error
		desc, err = execChangeReplicasTxn(
			ctx, desc, reason, details, iChgs, changeReplicasTxnArgs{
				db:                                   r.store.DB(),
				liveAndDeadReplicas:                  r.store.allocator.storePool.liveAndDeadReplicas,
				logChange:                            r.store.logChange,
				testForceJointConfig:                 r.store.TestingKnobs().ReplicationAlwaysUseJointConfig,
				testAllowDangerousReplicationChanges: r.store.TestingKnobs().AllowDangerousReplicationChanges,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116332)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116333)
		}
	}
	__antithesis_instrumentation__.Notify(116316)

	descriptorOK := false
	start := timeutil.Now()
	retOpts := retry.Options{InitialBackoff: time.Second, MaxBackoff: time.Second, MaxRetries: 10}
	for re := retry.StartWithCtx(ctx, retOpts); re.Next(); {
		__antithesis_instrumentation__.Notify(116334)
		rDesc := r.Desc()
		if rDesc.Generation >= desc.Generation {
			__antithesis_instrumentation__.Notify(116336)
			descriptorOK = true
			break
		} else {
			__antithesis_instrumentation__.Notify(116337)
		}
		__antithesis_instrumentation__.Notify(116335)
		log.VEventf(ctx, 1, "stale descriptor detected; waiting to catch up to replication. want: %s, have: %s",
			desc, rDesc)
		if _, err := r.IsDestroyed(); err != nil {
			__antithesis_instrumentation__.Notify(116338)
			return nil, errors.Wrapf(
				err,
				"replica destroyed while waiting desc replication",
			)
		} else {
			__antithesis_instrumentation__.Notify(116339)
		}
	}
	__antithesis_instrumentation__.Notify(116317)
	if !descriptorOK {
		__antithesis_instrumentation__.Notify(116340)
		return nil, errors.Newf(
			"waited for %s and replication hasn't caught up with descriptor update",
			timeutil.Since(start),
		)
	} else {
		__antithesis_instrumentation__.Notify(116341)
	}
	__antithesis_instrumentation__.Notify(116318)

	for _, target := range targets {
		__antithesis_instrumentation__.Notify(116342)
		rDesc, ok := desc.GetReplicaDescriptor(target.StoreID)
		if !ok {
			__antithesis_instrumentation__.Notify(116346)
			return nil, errors.Errorf("programming error: replica %v not found in %v", target, desc)
		} else {
			__antithesis_instrumentation__.Notify(116347)
		}
		__antithesis_instrumentation__.Notify(116343)

		if rDesc.GetType() != replicaType {
			__antithesis_instrumentation__.Notify(116348)
			return nil, errors.Errorf("programming error: cannot promote replica of type %s", rDesc.Type)
		} else {
			__antithesis_instrumentation__.Notify(116349)
		}
		__antithesis_instrumentation__.Notify(116344)

		if fn := r.store.cfg.TestingKnobs.ReplicaSkipInitialSnapshot; fn != nil && func() bool {
			__antithesis_instrumentation__.Notify(116350)
			return fn() == true
		}() == true {
			__antithesis_instrumentation__.Notify(116351)
			continue
		} else {
			__antithesis_instrumentation__.Notify(116352)
		}
		__antithesis_instrumentation__.Notify(116345)

		if err := r.sendSnapshot(ctx, rDesc, kvserverpb.SnapshotRequest_INITIAL, priority); err != nil {
			__antithesis_instrumentation__.Notify(116353)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(116354)
		}
	}
	__antithesis_instrumentation__.Notify(116319)
	return desc, nil
}

func (r *Replica) lockLearnerSnapshot(
	ctx context.Context, additions []roachpb.ReplicationTarget,
) (unlock func()) {
	__antithesis_instrumentation__.Notify(116355)

	var lockUUIDs []uuid.UUID
	for _, addition := range additions {
		__antithesis_instrumentation__.Notify(116357)
		lockUUID := uuid.MakeV4()
		lockUUIDs = append(lockUUIDs, lockUUID)
		r.addSnapshotLogTruncationConstraint(ctx, lockUUID, 1, addition.StoreID)
	}
	__antithesis_instrumentation__.Notify(116356)
	return func() {
		__antithesis_instrumentation__.Notify(116358)
		now := timeutil.Now()
		for _, lockUUID := range lockUUIDs {
			__antithesis_instrumentation__.Notify(116359)
			r.completeSnapshotLogTruncationConstraint(ctx, lockUUID, now)
		}
	}
}

func (r *Replica) execReplicationChangesForVoters(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	reason kvserverpb.RangeLogEventReason,
	details string,
	voterAdditions, voterRemovals []roachpb.ReplicationTarget,
) (rangeDesc *roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(116360)

	iChgs := make([]internalReplicationChange, 0, len(voterAdditions)+len(voterRemovals))
	for _, target := range voterAdditions {
		__antithesis_instrumentation__.Notify(116365)
		iChgs = append(iChgs, internalReplicationChange{target: target, typ: internalChangeTypePromoteLearner})
	}
	__antithesis_instrumentation__.Notify(116361)

	for _, target := range voterRemovals {
		__antithesis_instrumentation__.Notify(116366)
		typ := internalChangeTypeRemove
		if rDesc, ok := desc.GetReplicaDescriptor(target.StoreID); ok && func() bool {
			__antithesis_instrumentation__.Notify(116368)
			return rDesc.GetType() == roachpb.VOTER_FULL == true
		}() == true {
			__antithesis_instrumentation__.Notify(116369)
			typ = internalChangeTypeDemoteVoterToLearner
		} else {
			__antithesis_instrumentation__.Notify(116370)
		}
		__antithesis_instrumentation__.Notify(116367)
		iChgs = append(iChgs, internalReplicationChange{target: target, typ: typ})
	}
	__antithesis_instrumentation__.Notify(116362)

	desc, err = execChangeReplicasTxn(ctx, desc, reason, details, iChgs, changeReplicasTxnArgs{
		db:                                   r.store.DB(),
		liveAndDeadReplicas:                  r.store.allocator.storePool.liveAndDeadReplicas,
		logChange:                            r.store.logChange,
		testForceJointConfig:                 r.store.TestingKnobs().ReplicationAlwaysUseJointConfig,
		testAllowDangerousReplicationChanges: r.store.TestingKnobs().AllowDangerousReplicationChanges,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(116371)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(116372)
	}
	__antithesis_instrumentation__.Notify(116363)

	if fn := r.store.cfg.TestingKnobs.VoterAddStopAfterJointConfig; fn != nil && func() bool {
		__antithesis_instrumentation__.Notify(116373)
		return fn() == true
	}() == true {
		__antithesis_instrumentation__.Notify(116374)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(116375)
	}
	__antithesis_instrumentation__.Notify(116364)

	return r.maybeLeaveAtomicChangeReplicasAndRemoveLearners(ctx, desc)
}

func (r *Replica) tryRollbackRaftLearner(
	ctx context.Context,
	rangeDesc *roachpb.RangeDescriptor,
	target roachpb.ReplicationTarget,
	reason kvserverpb.RangeLogEventReason,
	details string,
) {
	__antithesis_instrumentation__.Notify(116376)
	repDesc, ok := rangeDesc.GetReplicaDescriptor(target.StoreID)
	isLearnerOrNonVoter := repDesc.GetType() == roachpb.LEARNER || func() bool {
		__antithesis_instrumentation__.Notify(116379)
		return repDesc.GetType() == roachpb.NON_VOTER == true
	}() == true
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(116380)
		return !isLearnerOrNonVoter == true
	}() == true {
		__antithesis_instrumentation__.Notify(116381)

		log.Event(ctx, "learner to roll back not found; skipping")
		return
	} else {
		__antithesis_instrumentation__.Notify(116382)
	}
	__antithesis_instrumentation__.Notify(116377)

	const rollbackTimeout = 10 * time.Second

	rollbackFn := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(116383)
		_, err := execChangeReplicasTxn(
			ctx, rangeDesc, reason, details,
			[]internalReplicationChange{{target: target, typ: internalChangeTypeRemove}},
			changeReplicasTxnArgs{
				db:                                   r.store.DB(),
				liveAndDeadReplicas:                  r.store.allocator.storePool.liveAndDeadReplicas,
				logChange:                            r.store.logChange,
				testForceJointConfig:                 r.store.TestingKnobs().ReplicationAlwaysUseJointConfig,
				testAllowDangerousReplicationChanges: r.store.TestingKnobs().AllowDangerousReplicationChanges,
			})
		return err
	}
	__antithesis_instrumentation__.Notify(116378)
	rollbackCtx := r.AnnotateCtx(context.Background())

	rollbackCtx = logtags.AddTags(rollbackCtx, logtags.FromContext(ctx))
	if err := contextutil.RunWithTimeout(
		rollbackCtx, "learner rollback", rollbackTimeout, rollbackFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(116384)
		log.Infof(
			ctx,
			"failed to rollback %s %s, abandoning it for the replicate queue: %v",
			repDesc.GetType(),
			target,
			err,
		)
		r.store.replicateQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(116385)
		log.Infof(ctx, "rolled back %s %s in %s", repDesc.GetType(), target, rangeDesc)
	}
}

type internalChangeType byte

const (
	_ internalChangeType = iota + 1
	internalChangeTypeAddLearner
	internalChangeTypeAddNonVoter

	internalChangeTypePromoteLearner
	internalChangeTypePromoteNonVoter

	internalChangeTypeDemoteVoterToLearner

	internalChangeTypeDemoteVoterToNonVoter

	internalChangeTypeRemove
)

type internalReplicationChange struct {
	target roachpb.ReplicationTarget
	typ    internalChangeType
}

type internalReplicationChanges []internalReplicationChange

func (c internalReplicationChanges) leaveJoint() bool {
	__antithesis_instrumentation__.Notify(116386)
	return len(c) == 0
}
func (c internalReplicationChanges) useJoint() bool {
	__antithesis_instrumentation__.Notify(116387)

	isDemotion := c[0].typ == internalChangeTypeDemoteVoterToNonVoter || func() bool {
		__antithesis_instrumentation__.Notify(116388)
		return c[0].typ == internalChangeTypeDemoteVoterToLearner == true
	}() == true
	return len(c) > 1 || func() bool {
		__antithesis_instrumentation__.Notify(116389)
		return isDemotion == true
	}() == true
}

func prepareChangeReplicasTrigger(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	chgs internalReplicationChanges,
	testingForceJointConfig func() bool,
) (*roachpb.ChangeReplicasTrigger, error) {
	__antithesis_instrumentation__.Notify(116390)
	updatedDesc := *desc
	updatedDesc.SetReplicas(desc.Replicas().DeepCopy())
	updatedDesc.IncrementGeneration()

	var added, removed []roachpb.ReplicaDescriptor
	if !chgs.leaveJoint() {
		__antithesis_instrumentation__.Notify(116394)
		if desc.Replicas().InAtomicReplicationChange() {
			__antithesis_instrumentation__.Notify(116397)
			return nil, errors.Errorf("must transition out of joint config first: %s", desc)
		} else {
			__antithesis_instrumentation__.Notify(116398)
		}
		__antithesis_instrumentation__.Notify(116395)

		useJoint := chgs.useJoint()
		if fn := testingForceJointConfig; fn != nil && func() bool {
			__antithesis_instrumentation__.Notify(116399)
			return fn() == true
		}() == true {
			__antithesis_instrumentation__.Notify(116400)
			useJoint = true
		} else {
			__antithesis_instrumentation__.Notify(116401)
		}
		__antithesis_instrumentation__.Notify(116396)
		for _, chg := range chgs {
			__antithesis_instrumentation__.Notify(116402)
			switch chg.typ {
			case internalChangeTypeAddLearner:
				__antithesis_instrumentation__.Notify(116403)
				added = append(added,
					updatedDesc.AddReplica(chg.target.NodeID, chg.target.StoreID, roachpb.LEARNER))
			case internalChangeTypeAddNonVoter:
				__antithesis_instrumentation__.Notify(116404)
				added = append(added,
					updatedDesc.AddReplica(chg.target.NodeID, chg.target.StoreID, roachpb.NON_VOTER))
			case internalChangeTypePromoteLearner:
				__antithesis_instrumentation__.Notify(116405)
				typ := roachpb.VOTER_FULL
				if useJoint {
					__antithesis_instrumentation__.Notify(116423)
					typ = roachpb.VOTER_INCOMING
				} else {
					__antithesis_instrumentation__.Notify(116424)
				}
				__antithesis_instrumentation__.Notify(116406)
				rDesc, prevTyp, ok := updatedDesc.SetReplicaType(chg.target.NodeID, chg.target.StoreID, typ)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(116425)
					return prevTyp != roachpb.LEARNER == true
				}() == true {
					__antithesis_instrumentation__.Notify(116426)
					return nil, errors.Errorf("cannot promote target %v which is missing as LEARNER",
						chg.target)
				} else {
					__antithesis_instrumentation__.Notify(116427)
				}
				__antithesis_instrumentation__.Notify(116407)
				added = append(added, rDesc)
			case internalChangeTypePromoteNonVoter:
				__antithesis_instrumentation__.Notify(116408)
				typ := roachpb.VOTER_FULL
				if useJoint {
					__antithesis_instrumentation__.Notify(116428)
					typ = roachpb.VOTER_INCOMING
				} else {
					__antithesis_instrumentation__.Notify(116429)
				}
				__antithesis_instrumentation__.Notify(116409)
				rDesc, prevTyp, ok := updatedDesc.SetReplicaType(chg.target.NodeID, chg.target.StoreID, typ)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(116430)
					return prevTyp != roachpb.NON_VOTER == true
				}() == true {
					__antithesis_instrumentation__.Notify(116431)
					return nil, errors.Errorf("cannot promote target %v which is missing as NON_VOTER",
						chg.target)
				} else {
					__antithesis_instrumentation__.Notify(116432)
				}
				__antithesis_instrumentation__.Notify(116410)
				added = append(added, rDesc)
			case internalChangeTypeRemove:
				__antithesis_instrumentation__.Notify(116411)
				rDesc, ok := updatedDesc.GetReplicaDescriptor(chg.target.StoreID)
				if !ok {
					__antithesis_instrumentation__.Notify(116433)
					return nil, errors.Errorf("target %s not found", chg.target)
				} else {
					__antithesis_instrumentation__.Notify(116434)
				}
				__antithesis_instrumentation__.Notify(116412)
				prevTyp := rDesc.GetType()
				isRaftLearner := prevTyp == roachpb.LEARNER || func() bool {
					__antithesis_instrumentation__.Notify(116435)
					return prevTyp == roachpb.NON_VOTER == true
				}() == true
				if !useJoint || func() bool {
					__antithesis_instrumentation__.Notify(116436)
					return isRaftLearner == true
				}() == true {
					__antithesis_instrumentation__.Notify(116437)
					rDesc, _ = updatedDesc.RemoveReplica(chg.target.NodeID, chg.target.StoreID)
				} else {
					__antithesis_instrumentation__.Notify(116438)
					if prevTyp != roachpb.VOTER_FULL {
						__antithesis_instrumentation__.Notify(116439)

						return nil, errors.AssertionFailedf("cannot transition from %s to VOTER_OUTGOING", prevTyp)
					} else {
						__antithesis_instrumentation__.Notify(116440)
						rDesc, _, _ = updatedDesc.SetReplicaType(chg.target.NodeID, chg.target.StoreID, roachpb.VOTER_OUTGOING)
					}
				}
				__antithesis_instrumentation__.Notify(116413)
				removed = append(removed, rDesc)
			case internalChangeTypeDemoteVoterToLearner:
				__antithesis_instrumentation__.Notify(116414)

				rDesc, ok := updatedDesc.GetReplicaDescriptor(chg.target.StoreID)
				if !ok {
					__antithesis_instrumentation__.Notify(116441)
					return nil, errors.Errorf("target %s not found", chg.target)
				} else {
					__antithesis_instrumentation__.Notify(116442)
				}
				__antithesis_instrumentation__.Notify(116415)
				if !useJoint {
					__antithesis_instrumentation__.Notify(116443)

					return nil, errors.AssertionFailedf("demotions require joint consensus")
				} else {
					__antithesis_instrumentation__.Notify(116444)
				}
				__antithesis_instrumentation__.Notify(116416)
				if prevTyp := rDesc.GetType(); prevTyp != roachpb.VOTER_FULL {
					__antithesis_instrumentation__.Notify(116445)
					return nil, errors.Errorf("cannot transition from %s to VOTER_DEMOTING_LEARNER", prevTyp)
				} else {
					__antithesis_instrumentation__.Notify(116446)
				}
				__antithesis_instrumentation__.Notify(116417)
				rDesc, _, _ = updatedDesc.SetReplicaType(chg.target.NodeID, chg.target.StoreID, roachpb.VOTER_DEMOTING_LEARNER)
				removed = append(removed, rDesc)
			case internalChangeTypeDemoteVoterToNonVoter:
				__antithesis_instrumentation__.Notify(116418)
				rDesc, ok := updatedDesc.GetReplicaDescriptor(chg.target.StoreID)
				if !ok {
					__antithesis_instrumentation__.Notify(116447)
					return nil, errors.Errorf("target %s not found", chg.target)
				} else {
					__antithesis_instrumentation__.Notify(116448)
				}
				__antithesis_instrumentation__.Notify(116419)
				if !useJoint {
					__antithesis_instrumentation__.Notify(116449)

					return nil, errors.Errorf("demotions require joint consensus")
				} else {
					__antithesis_instrumentation__.Notify(116450)
				}
				__antithesis_instrumentation__.Notify(116420)
				if prevTyp := rDesc.GetType(); prevTyp != roachpb.VOTER_FULL {
					__antithesis_instrumentation__.Notify(116451)
					return nil, errors.Errorf("cannot transition from %s to VOTER_DEMOTING_NON_VOTER", prevTyp)
				} else {
					__antithesis_instrumentation__.Notify(116452)
				}
				__antithesis_instrumentation__.Notify(116421)
				rDesc, _, _ = updatedDesc.SetReplicaType(chg.target.NodeID, chg.target.StoreID, roachpb.VOTER_DEMOTING_NON_VOTER)
				removed = append(removed, rDesc)
			default:
				__antithesis_instrumentation__.Notify(116422)
				return nil, errors.Errorf("unsupported internal change type %d", chg.typ)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(116453)

		var isJoint bool

		for _, rDesc := range updatedDesc.Replicas().DeepCopy().Descriptors() {
			__antithesis_instrumentation__.Notify(116455)
			switch rDesc.GetType() {
			case roachpb.VOTER_INCOMING:
				__antithesis_instrumentation__.Notify(116456)
				updatedDesc.SetReplicaType(rDesc.NodeID, rDesc.StoreID, roachpb.VOTER_FULL)
				isJoint = true
			case roachpb.VOTER_OUTGOING:
				__antithesis_instrumentation__.Notify(116457)
				updatedDesc.RemoveReplica(rDesc.NodeID, rDesc.StoreID)
				isJoint = true
			case roachpb.VOTER_DEMOTING_LEARNER:
				__antithesis_instrumentation__.Notify(116458)
				updatedDesc.SetReplicaType(rDesc.NodeID, rDesc.StoreID, roachpb.LEARNER)
				isJoint = true
			case roachpb.VOTER_DEMOTING_NON_VOTER:
				__antithesis_instrumentation__.Notify(116459)
				updatedDesc.SetReplicaType(rDesc.NodeID, rDesc.StoreID, roachpb.NON_VOTER)
				isJoint = true
			default:
				__antithesis_instrumentation__.Notify(116460)
			}
		}
		__antithesis_instrumentation__.Notify(116454)
		if !isJoint {
			__antithesis_instrumentation__.Notify(116461)
			return nil, errors.Errorf("cannot leave a joint config; desc not joint: %s", &updatedDesc)
		} else {
			__antithesis_instrumentation__.Notify(116462)
		}
	}
	__antithesis_instrumentation__.Notify(116391)

	if err := updatedDesc.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(116463)
		return nil, errors.Wrapf(err, "validating updated descriptor %s", &updatedDesc)
	} else {
		__antithesis_instrumentation__.Notify(116464)
	}
	__antithesis_instrumentation__.Notify(116392)

	crt := &roachpb.ChangeReplicasTrigger{
		Desc:                    &updatedDesc,
		InternalAddedReplicas:   added,
		InternalRemovedReplicas: removed,
	}

	if _, err := crt.ConfChange(nil); err != nil {
		__antithesis_instrumentation__.Notify(116465)
		return nil, errors.Wrapf(err, "programming error: malformed trigger created from desc %s to %s", desc, &updatedDesc)
	} else {
		__antithesis_instrumentation__.Notify(116466)
	}
	__antithesis_instrumentation__.Notify(116393)
	return crt, nil
}

func within10s(ctx context.Context, fn func() error) error {
	__antithesis_instrumentation__.Notify(116467)
	retOpts := retry.Options{InitialBackoff: time.Second, MaxBackoff: time.Second, MaxRetries: 10}
	var err error
	for re := retry.StartWithCtx(ctx, retOpts); re.Next(); {
		__antithesis_instrumentation__.Notify(116469)
		err = fn()
		if err == nil {
			__antithesis_instrumentation__.Notify(116470)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(116471)
		}
	}
	__antithesis_instrumentation__.Notify(116468)
	return ctx.Err()
}

type changeReplicasTxnArgs struct {
	db *kv.DB

	liveAndDeadReplicas func(
		repls []roachpb.ReplicaDescriptor, includeSuspectAndDrainingStores bool,
	) (liveReplicas, deadReplicas []roachpb.ReplicaDescriptor)

	logChange                            logChangeFn
	testForceJointConfig                 func() bool
	testAllowDangerousReplicationChanges bool
}

func execChangeReplicasTxn(
	ctx context.Context,
	referenceDesc *roachpb.RangeDescriptor,
	reason kvserverpb.RangeLogEventReason,
	details string,
	chgs internalReplicationChanges,
	args changeReplicasTxnArgs,
) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(116472)
	var returnDesc *roachpb.RangeDescriptor

	descKey := keys.RangeDescriptorKey(referenceDesc.StartKey)

	check := func(kvDesc *roachpb.RangeDescriptor) bool {
		__antithesis_instrumentation__.Notify(116475)

		if kvDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(116477)
			return kvDesc.RangeID == referenceDesc.RangeID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(116478)
			return chgs.leaveJoint() == true
		}() == true {
			__antithesis_instrumentation__.Notify(116479)

			return true
		} else {
			__antithesis_instrumentation__.Notify(116480)
		}
		__antithesis_instrumentation__.Notify(116476)

		return checkDescsEqual(referenceDesc)(kvDesc)
	}
	__antithesis_instrumentation__.Notify(116473)

	if err := args.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(116481)
		log.Event(ctx, "attempting txn")
		txn.SetDebugName(replicaChangeTxnName)
		desc, dbDescValue, err := conditionalGetDescValueFromDB(
			ctx, txn, referenceDesc.StartKey, false, check)
		if err != nil {
			__antithesis_instrumentation__.Notify(116490)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116491)
		}
		__antithesis_instrumentation__.Notify(116482)
		if chgs.leaveJoint() && func() bool {
			__antithesis_instrumentation__.Notify(116492)
			return !desc.Replicas().InAtomicReplicationChange() == true
		}() == true {
			__antithesis_instrumentation__.Notify(116493)

			returnDesc = desc
			return nil
		} else {
			__antithesis_instrumentation__.Notify(116494)
		}
		__antithesis_instrumentation__.Notify(116483)

		crt, err := prepareChangeReplicasTrigger(ctx, desc, chgs, args.testForceJointConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(116495)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116496)
		}
		__antithesis_instrumentation__.Notify(116484)
		log.Infof(ctx, "change replicas (add %v remove %v): existing descriptor %s", crt.Added(), crt.Removed(), desc)

		if err := within10s(ctx, func() error {
			__antithesis_instrumentation__.Notify(116497)
			if args.testAllowDangerousReplicationChanges {
				__antithesis_instrumentation__.Notify(116500)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(116501)
			}
			__antithesis_instrumentation__.Notify(116498)

			replicas := crt.Desc.Replicas()

			liveReplicas, _ := args.liveAndDeadReplicas(replicas.Descriptors(), true)
			if !replicas.CanMakeProgress(
				func(rDesc roachpb.ReplicaDescriptor) bool {
					__antithesis_instrumentation__.Notify(116502)
					for _, inner := range liveReplicas {
						__antithesis_instrumentation__.Notify(116504)
						if inner.ReplicaID == rDesc.ReplicaID {
							__antithesis_instrumentation__.Notify(116505)
							return true
						} else {
							__antithesis_instrumentation__.Notify(116506)
						}
					}
					__antithesis_instrumentation__.Notify(116503)
					return false
				}) {
				__antithesis_instrumentation__.Notify(116507)

				return newQuorumError("range %s cannot make progress with proposed changes add=%v del=%v "+
					"based on live replicas %v", crt.Desc, crt.Added(), crt.Removed(), liveReplicas)
			} else {
				__antithesis_instrumentation__.Notify(116508)
			}
			__antithesis_instrumentation__.Notify(116499)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(116509)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116510)
		}

		{
			__antithesis_instrumentation__.Notify(116511)
			b := txn.NewBatch()

			if err := updateRangeDescriptor(ctx, b, descKey, dbDescValue, crt.Desc); err != nil {
				__antithesis_instrumentation__.Notify(116513)
				return err
			} else {
				__antithesis_instrumentation__.Notify(116514)
			}
			__antithesis_instrumentation__.Notify(116512)

			if err := txn.Run(ctx, b); err != nil {
				__antithesis_instrumentation__.Notify(116515)
				return err
			} else {
				__antithesis_instrumentation__.Notify(116516)
			}
		}
		__antithesis_instrumentation__.Notify(116485)

		err = recordRangeEventsInLog(
			ctx, txn, true, crt.Added(), crt.Desc, reason, details, args.logChange,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116517)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116518)
		}
		__antithesis_instrumentation__.Notify(116486)
		err = recordRangeEventsInLog(
			ctx, txn, false, crt.Removed(), crt.Desc, reason, details, args.logChange,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116519)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116520)
		}
		__antithesis_instrumentation__.Notify(116487)

		b := txn.NewBatch()

		if err := updateRangeAddressing(b, crt.Desc); err != nil {
			__antithesis_instrumentation__.Notify(116521)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116522)
		}
		__antithesis_instrumentation__.Notify(116488)

		b.AddRawRequest(&roachpb.EndTxnRequest{
			Commit: true,
			InternalCommitTrigger: &roachpb.InternalCommitTrigger{
				ChangeReplicasTrigger: crt,
			},
		})
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(116523)
			log.Eventf(ctx, "%v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116524)
		}
		__antithesis_instrumentation__.Notify(116489)

		returnDesc = crt.Desc
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(116525)
		log.Eventf(ctx, "%v", err)

		if ok, actualDesc := maybeDescriptorChangedError(referenceDesc, err); ok {
			__antithesis_instrumentation__.Notify(116527)

			err = errors.WithSecondaryError(newDescChangedError(referenceDesc, actualDesc), err)
			err = &benignError{err}
		} else {
			__antithesis_instrumentation__.Notify(116528)
		}
		__antithesis_instrumentation__.Notify(116526)
		return nil, errors.Wrapf(err, "change replicas of r%d failed", referenceDesc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(116529)
	}
	__antithesis_instrumentation__.Notify(116474)
	log.Event(ctx, "txn complete")
	return returnDesc, nil
}

func getInternalChangesForExplicitPromotionsAndDemotions(
	voterDemotions, nonVoterPromotions []roachpb.ReplicationTarget,
) []internalReplicationChange {
	__antithesis_instrumentation__.Notify(116530)
	iChgs := make([]internalReplicationChange, len(voterDemotions)+len(nonVoterPromotions))
	for i := range voterDemotions {
		__antithesis_instrumentation__.Notify(116533)
		iChgs[i] = internalReplicationChange{
			target: voterDemotions[i],
			typ:    internalChangeTypeDemoteVoterToNonVoter,
		}
	}
	__antithesis_instrumentation__.Notify(116531)

	for j := range nonVoterPromotions {
		__antithesis_instrumentation__.Notify(116534)
		iChgs[j+len(voterDemotions)] = internalReplicationChange{
			target: nonVoterPromotions[j],
			typ:    internalChangeTypePromoteNonVoter,
		}
	}
	__antithesis_instrumentation__.Notify(116532)

	return iChgs
}

type logChangeFn func(
	ctx context.Context,
	txn *kv.Txn,
	changeType roachpb.ReplicaChangeType,
	replica roachpb.ReplicaDescriptor,
	desc roachpb.RangeDescriptor,
	reason kvserverpb.RangeLogEventReason,
	details string,
) error

func recordRangeEventsInLog(
	ctx context.Context,
	txn *kv.Txn,
	added bool,
	repDescs []roachpb.ReplicaDescriptor,
	rangeDesc *roachpb.RangeDescriptor,
	reason kvserverpb.RangeLogEventReason,
	details string,
	logChange logChangeFn,
) error {
	__antithesis_instrumentation__.Notify(116535)
	for _, repDesc := range repDescs {
		__antithesis_instrumentation__.Notify(116537)
		isNonVoter := repDesc.GetType() == roachpb.NON_VOTER
		var typ roachpb.ReplicaChangeType
		if added {
			__antithesis_instrumentation__.Notify(116539)
			typ = roachpb.ADD_VOTER
			if isNonVoter {
				__antithesis_instrumentation__.Notify(116540)
				typ = roachpb.ADD_NON_VOTER
			} else {
				__antithesis_instrumentation__.Notify(116541)
			}
		} else {
			__antithesis_instrumentation__.Notify(116542)
			typ = roachpb.REMOVE_VOTER
			if isNonVoter {
				__antithesis_instrumentation__.Notify(116543)
				typ = roachpb.REMOVE_NON_VOTER
			} else {
				__antithesis_instrumentation__.Notify(116544)
			}
		}
		__antithesis_instrumentation__.Notify(116538)
		if err := logChange(
			ctx, txn, typ, repDesc, *rangeDesc, reason, details,
		); err != nil {
			__antithesis_instrumentation__.Notify(116545)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116546)
		}
	}
	__antithesis_instrumentation__.Notify(116536)
	return nil
}

func (r *Replica) sendSnapshot(
	ctx context.Context,
	recipient roachpb.ReplicaDescriptor,
	snapType kvserverpb.SnapshotRequest_Type,
	priority kvserverpb.SnapshotRequest_Priority,
) (retErr error) {
	__antithesis_instrumentation__.Notify(116547)
	defer func() {
		__antithesis_instrumentation__.Notify(116557)

		r.reportSnapshotStatus(ctx, recipient.ReplicaID, retErr)
	}()
	__antithesis_instrumentation__.Notify(116548)

	snap, err := r.GetSnapshot(ctx, snapType, recipient.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(116558)
		err = errors.Wrapf(err, "%s: failed to generate %s snapshot", r, snapType)
		return errors.Mark(err, errMarkSnapshotError)
	} else {
		__antithesis_instrumentation__.Notify(116559)
	}
	__antithesis_instrumentation__.Notify(116549)
	defer snap.Close()
	log.Event(ctx, "generated snapshot")

	if _, ok := snap.State.Desc.GetReplicaDescriptor(recipient.StoreID); !ok {
		__antithesis_instrumentation__.Notify(116560)
		return errors.Wrapf(errMarkSnapshotError,
			"attempting to send snapshot that does not contain the recipient as a replica; "+
				"snapshot type: %s, recipient: s%d, desc: %s", snapType, recipient, snap.State.Desc)
	} else {
		__antithesis_instrumentation__.Notify(116561)
	}
	__antithesis_instrumentation__.Notify(116550)

	sender, err := r.GetReplicaDescriptor()
	if err != nil {
		__antithesis_instrumentation__.Notify(116562)
		return errors.Wrapf(err, "%s: change replicas failed", r)
	} else {
		__antithesis_instrumentation__.Notify(116563)
	}
	__antithesis_instrumentation__.Notify(116551)

	status := r.RaftStatus()
	if status == nil {
		__antithesis_instrumentation__.Notify(116564)

		return &benignError{errors.Wrap(errMarkSnapshotError, "raft status not initialized")}
	} else {
		__antithesis_instrumentation__.Notify(116565)
	}
	__antithesis_instrumentation__.Notify(116552)

	_ = (*kvBatchSnapshotStrategy)(nil).Send

	snap.State.TruncatedState = &roachpb.RaftTruncatedState{
		Index: snap.RaftSnap.Metadata.Index,
		Term:  snap.RaftSnap.Metadata.Term,
	}

	snap.State.DeprecatedUsingAppliedStateKey = true

	req := kvserverpb.SnapshotRequest_Header{
		State:                                snap.State,
		DeprecatedUnreplicatedTruncatedState: true,
		RaftMessageRequest: kvserverpb.RaftMessageRequest{
			RangeID:     r.RangeID,
			FromReplica: sender,
			ToReplica:   recipient,
			Message: raftpb.Message{
				Type:     raftpb.MsgSnap,
				To:       uint64(recipient.ReplicaID),
				From:     uint64(sender.ReplicaID),
				Term:     status.Term,
				Snapshot: snap.RaftSnap,
			},
		},
		RangeSize: r.GetMVCCStats().Total(),
		Priority:  priority,
		Strategy:  kvserverpb.SnapshotRequest_KV_BATCH,
		Type:      snapType,
	}
	newBatchFn := func() storage.Batch {
		__antithesis_instrumentation__.Notify(116566)
		return r.store.Engine().NewUnindexedBatch(true)
	}
	__antithesis_instrumentation__.Notify(116553)
	sent := func() {
		__antithesis_instrumentation__.Notify(116567)
		r.store.metrics.RangeSnapshotsGenerated.Inc(1)
	}
	__antithesis_instrumentation__.Notify(116554)
	err = contextutil.RunWithTimeout(
		ctx, "send-snapshot", sendSnapshotTimeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(116568)
			return r.store.cfg.Transport.SendSnapshot(
				ctx,
				r.store.allocator.storePool,
				req,
				snap,
				newBatchFn,
				sent,
				r.store.metrics.RangeSnapshotSentBytes,
			)
		})
	__antithesis_instrumentation__.Notify(116555)
	if err != nil {
		__antithesis_instrumentation__.Notify(116569)
		if errors.Is(err, errMalformedSnapshot) {
			__antithesis_instrumentation__.Notify(116571)
			tag := fmt.Sprintf("r%d_%s", r.RangeID, snap.SnapUUID.Short())
			if dir, err := r.store.checkpoint(ctx, tag); err != nil {
				__antithesis_instrumentation__.Notify(116573)
				log.Warningf(ctx, "unable to create checkpoint %s: %+v", dir, err)
			} else {
				__antithesis_instrumentation__.Notify(116574)
				log.Warningf(ctx, "created checkpoint %s", dir)
			}
			__antithesis_instrumentation__.Notify(116572)

			log.Fatal(ctx, "malformed snapshot generated")
		} else {
			__antithesis_instrumentation__.Notify(116575)
		}
		__antithesis_instrumentation__.Notify(116570)
		return errors.Mark(err, errMarkSnapshotError)
	} else {
		__antithesis_instrumentation__.Notify(116576)
	}
	__antithesis_instrumentation__.Notify(116556)
	return nil
}

func replicasCollocated(a, b []roachpb.ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(116577)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(116582)
		return false
	} else {
		__antithesis_instrumentation__.Notify(116583)
	}
	__antithesis_instrumentation__.Notify(116578)

	set := make(map[roachpb.StoreID]int)
	for _, replica := range a {
		__antithesis_instrumentation__.Notify(116584)
		set[replica.StoreID]++
	}
	__antithesis_instrumentation__.Notify(116579)

	for _, replica := range b {
		__antithesis_instrumentation__.Notify(116585)
		set[replica.StoreID]--
	}
	__antithesis_instrumentation__.Notify(116580)

	for _, value := range set {
		__antithesis_instrumentation__.Notify(116586)
		if value != 0 {
			__antithesis_instrumentation__.Notify(116587)
			return false
		} else {
			__antithesis_instrumentation__.Notify(116588)
		}
	}
	__antithesis_instrumentation__.Notify(116581)

	return true
}

func checkDescsEqual(desc *roachpb.RangeDescriptor) func(*roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(116589)
	return func(desc2 *roachpb.RangeDescriptor) bool {
		__antithesis_instrumentation__.Notify(116590)
		return desc.Equal(desc2)
	}
}

func conditionalGetDescValueFromDB(
	ctx context.Context,
	txn *kv.Txn,
	startKey roachpb.RKey,
	forUpdate bool,
	check func(*roachpb.RangeDescriptor) bool,
) (*roachpb.RangeDescriptor, []byte, error) {
	__antithesis_instrumentation__.Notify(116591)
	get := txn.Get
	if forUpdate {
		__antithesis_instrumentation__.Notify(116596)
		get = txn.GetForUpdate
	} else {
		__antithesis_instrumentation__.Notify(116597)
	}
	__antithesis_instrumentation__.Notify(116592)
	descKey := keys.RangeDescriptorKey(startKey)
	existingDescKV, err := get(ctx, descKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(116598)
		return nil, nil, errors.Wrap(err, "fetching current range descriptor value")
	} else {
		__antithesis_instrumentation__.Notify(116599)
	}
	__antithesis_instrumentation__.Notify(116593)
	var existingDesc *roachpb.RangeDescriptor
	if existingDescKV.Value != nil {
		__antithesis_instrumentation__.Notify(116600)
		existingDesc = &roachpb.RangeDescriptor{}
		if err := existingDescKV.Value.GetProto(existingDesc); err != nil {
			__antithesis_instrumentation__.Notify(116601)
			return nil, nil, errors.Wrap(err, "decoding current range descriptor value")
		} else {
			__antithesis_instrumentation__.Notify(116602)
		}
	} else {
		__antithesis_instrumentation__.Notify(116603)
	}
	__antithesis_instrumentation__.Notify(116594)

	if !check(existingDesc) {
		__antithesis_instrumentation__.Notify(116604)
		return nil, nil, &roachpb.ConditionFailedError{ActualValue: existingDescKV.Value}
	} else {
		__antithesis_instrumentation__.Notify(116605)
	}
	__antithesis_instrumentation__.Notify(116595)
	return existingDesc, existingDescKV.Value.TagAndDataBytes(), nil
}

func updateRangeDescriptor(
	ctx context.Context,
	b *kv.Batch,
	descKey roachpb.Key,
	oldValue []byte,
	newDesc *roachpb.RangeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(116606)

	var newValue interface{}
	if newDesc != nil {
		__antithesis_instrumentation__.Notify(116608)
		if err := newDesc.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(116611)
			return errors.Wrapf(err, "validating new descriptor %+v (old descriptor is %+v)",
				newDesc, oldValue)
		} else {
			__antithesis_instrumentation__.Notify(116612)
		}
		__antithesis_instrumentation__.Notify(116609)
		newBytes, err := protoutil.Marshal(newDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(116613)
			return err
		} else {
			__antithesis_instrumentation__.Notify(116614)
		}
		__antithesis_instrumentation__.Notify(116610)
		newValue = newBytes
	} else {
		__antithesis_instrumentation__.Notify(116615)
	}
	__antithesis_instrumentation__.Notify(116607)
	b.CPut(descKey, newValue, oldValue)
	return nil
}

func (r *Replica) AdminRelocateRange(
	ctx context.Context,
	rangeDesc roachpb.RangeDescriptor,
	voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
	transferLeaseToFirstVoter bool,
) error {
	__antithesis_instrumentation__.Notify(116616)
	if containsDuplicates(voterTargets) {
		__antithesis_instrumentation__.Notify(116622)
		return errors.AssertionFailedf(
			"list of desired voter targets contains duplicates: %+v",
			voterTargets,
		)
	} else {
		__antithesis_instrumentation__.Notify(116623)
	}
	__antithesis_instrumentation__.Notify(116617)
	if containsDuplicates(nonVoterTargets) {
		__antithesis_instrumentation__.Notify(116624)
		return errors.AssertionFailedf(
			"list of desired non-voter targets contains duplicates: %+v",
			nonVoterTargets,
		)
	} else {
		__antithesis_instrumentation__.Notify(116625)
	}
	__antithesis_instrumentation__.Notify(116618)
	if containsDuplicates(append(voterTargets, nonVoterTargets...)) {
		__antithesis_instrumentation__.Notify(116626)
		return errors.AssertionFailedf(
			"list of voter targets overlaps with the list of non-voter targets: voters: %+v, non-voters: %+v",
			voterTargets, nonVoterTargets,
		)
	} else {
		__antithesis_instrumentation__.Notify(116627)
	}
	__antithesis_instrumentation__.Notify(116619)

	newDesc, err :=
		r.maybeLeaveAtomicChangeReplicasAndRemoveLearners(ctx, &rangeDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(116628)
		log.Warningf(ctx, "%v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116629)
	}
	__antithesis_instrumentation__.Notify(116620)
	rangeDesc = *newDesc

	rangeDesc, err = r.relocateReplicas(
		ctx, rangeDesc, voterTargets, nonVoterTargets, transferLeaseToFirstVoter,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(116630)
		return err
	} else {
		__antithesis_instrumentation__.Notify(116631)
	}
	__antithesis_instrumentation__.Notify(116621)
	return nil
}

func (r *Replica) relocateReplicas(
	ctx context.Context,
	rangeDesc roachpb.RangeDescriptor,
	voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
	transferLeaseToFirstVoter bool,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(116632)
	startKey := rangeDesc.StartKey.AsRawKey()
	transferLease := func(target roachpb.ReplicationTarget) error {
		__antithesis_instrumentation__.Notify(116634)

		if err := r.store.DB().AdminTransferLease(
			ctx, startKey, target.StoreID,
		); err != nil {
			__antithesis_instrumentation__.Notify(116636)
			log.Warningf(ctx, "while transferring lease: %+v", err)
			if r.store.TestingKnobs().DontIgnoreFailureToTransferLease {
				__antithesis_instrumentation__.Notify(116637)
				return err
			} else {
				__antithesis_instrumentation__.Notify(116638)
			}
		} else {
			__antithesis_instrumentation__.Notify(116639)
		}
		__antithesis_instrumentation__.Notify(116635)
		return nil
	}
	__antithesis_instrumentation__.Notify(116633)

	every := log.Every(time.Minute)
	for {
		__antithesis_instrumentation__.Notify(116640)
		for re := retry.StartWithCtx(ctx, retry.Options{MaxBackoff: 5 * time.Second}); re.Next(); {
			__antithesis_instrumentation__.Notify(116641)
			if err := ctx.Err(); err != nil {
				__antithesis_instrumentation__.Notify(116647)
				return rangeDesc, err
			} else {
				__antithesis_instrumentation__.Notify(116648)
			}
			__antithesis_instrumentation__.Notify(116642)

			ops, leaseTarget, err := r.relocateOne(
				ctx, &rangeDesc, voterTargets, nonVoterTargets, transferLeaseToFirstVoter,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(116649)
				return rangeDesc, err
			} else {
				__antithesis_instrumentation__.Notify(116650)
			}
			__antithesis_instrumentation__.Notify(116643)
			if leaseTarget != nil {
				__antithesis_instrumentation__.Notify(116651)
				if err := transferLease(*leaseTarget); err != nil {
					__antithesis_instrumentation__.Notify(116652)
					return rangeDesc, err
				} else {
					__antithesis_instrumentation__.Notify(116653)
				}
			} else {
				__antithesis_instrumentation__.Notify(116654)
			}
			__antithesis_instrumentation__.Notify(116644)
			if len(ops) == 0 {
				__antithesis_instrumentation__.Notify(116655)

				return rangeDesc, ctx.Err()
			} else {
				__antithesis_instrumentation__.Notify(116656)
			}
			__antithesis_instrumentation__.Notify(116645)

			opss := [][]roachpb.ReplicationChange{ops}
			success := true
			for _, ops := range opss {
				__antithesis_instrumentation__.Notify(116657)
				newDesc, err := r.store.DB().AdminChangeReplicas(ctx, startKey, rangeDesc, ops)
				if err != nil {
					__antithesis_instrumentation__.Notify(116659)
					returnErr := errors.Wrapf(err, "while carrying out changes %v", ops)
					if !isSnapshotError(err) {
						__antithesis_instrumentation__.Notify(116662)
						return rangeDesc, returnErr
					} else {
						__antithesis_instrumentation__.Notify(116663)
					}
					__antithesis_instrumentation__.Notify(116660)
					if every.ShouldLog() {
						__antithesis_instrumentation__.Notify(116664)
						log.Infof(ctx, "%v", returnErr)
					} else {
						__antithesis_instrumentation__.Notify(116665)
					}
					__antithesis_instrumentation__.Notify(116661)
					success = false
					break
				} else {
					__antithesis_instrumentation__.Notify(116666)
				}
				__antithesis_instrumentation__.Notify(116658)
				rangeDesc = *newDesc
			}
			__antithesis_instrumentation__.Notify(116646)
			if success {
				__antithesis_instrumentation__.Notify(116667)
				if fn := r.store.cfg.TestingKnobs.OnRelocatedOne; fn != nil {
					__antithesis_instrumentation__.Notify(116669)
					fn(ops, &voterTargets[0])
				} else {
					__antithesis_instrumentation__.Notify(116670)
				}
				__antithesis_instrumentation__.Notify(116668)

				break
			} else {
				__antithesis_instrumentation__.Notify(116671)
			}
		}
	}
}

type relocationArgs struct {
	votersToAdd, votersToRemove             []roachpb.ReplicationTarget
	nonVotersToAdd, nonVotersToRemove       []roachpb.ReplicationTarget
	finalVoterTargets, finalNonVoterTargets []roachpb.ReplicationTarget
	targetType                              targetReplicaType
}

func (r *relocationArgs) targetsToAdd() []roachpb.ReplicationTarget {
	__antithesis_instrumentation__.Notify(116672)
	switch r.targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(116673)
		return r.votersToAdd
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(116674)
		return r.nonVotersToAdd
	default:
		__antithesis_instrumentation__.Notify(116675)
		panic(fmt.Sprintf("unknown targetReplicaType: %s", r.targetType))
	}
}

func (r *relocationArgs) targetsToRemove() []roachpb.ReplicationTarget {
	__antithesis_instrumentation__.Notify(116676)
	switch r.targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(116677)
		return r.votersToRemove
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(116678)
		return r.nonVotersToRemove
	default:
		__antithesis_instrumentation__.Notify(116679)
		panic(fmt.Sprintf("unknown targetReplicaType: %s", r.targetType))
	}
}

func (r *relocationArgs) finalRelocationTargets() []roachpb.ReplicationTarget {
	__antithesis_instrumentation__.Notify(116680)
	switch r.targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(116681)
		return r.finalVoterTargets
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(116682)
		return r.finalNonVoterTargets
	default:
		__antithesis_instrumentation__.Notify(116683)
		panic(fmt.Sprintf("unknown targetReplicaType: %s", r.targetType))
	}
}

func (r *Replica) relocateOne(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
	transferLeaseToFirstVoter bool,
) ([]roachpb.ReplicationChange, *roachpb.ReplicationTarget, error) {
	__antithesis_instrumentation__.Notify(116684)
	if repls := desc.Replicas(); len(repls.VoterFullAndNonVoterDescriptors()) != len(repls.Descriptors()) {
		__antithesis_instrumentation__.Notify(116692)

		return nil, nil, errors.AssertionFailedf(
			`range %s was either in a joint configuration or had learner replicas: %v`, desc, desc.Replicas())
	} else {
		__antithesis_instrumentation__.Notify(116693)
	}
	__antithesis_instrumentation__.Notify(116685)

	confReader, err := r.store.GetConfReader(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(116694)
		return nil, nil, errors.Wrap(err, "can't relocate range")
	} else {
		__antithesis_instrumentation__.Notify(116695)
	}
	__antithesis_instrumentation__.Notify(116686)
	conf, err := confReader.GetSpanConfigForKey(ctx, desc.StartKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(116696)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(116697)
	}
	__antithesis_instrumentation__.Notify(116687)

	storeList, _, _ := r.store.allocator.storePool.getStoreList(storeFilterNone)
	storeMap := storeListToMap(storeList)

	args := getRelocationArgs(desc, voterTargets, nonVoterTargets)
	existingVoters := desc.Replicas().VoterDescriptors()
	existingNonVoters := desc.Replicas().NonVoterDescriptors()
	existingReplicas := desc.Replicas().Descriptors()

	var additionTarget, removalTarget roachpb.ReplicationTarget
	var shouldAdd, shouldRemove, canPromoteNonVoter, canDemoteVoter bool
	if len(args.targetsToAdd()) > 0 {
		__antithesis_instrumentation__.Notify(116698)

		candidateTargets := args.targetsToAdd()
		if args.targetType == voterTarget && func() bool {
			__antithesis_instrumentation__.Notify(116703)
			return storeHasReplica(args.finalRelocationTargets()[0].StoreID, candidateTargets) == true
		}() == true {
			__antithesis_instrumentation__.Notify(116704)
			candidateTargets = []roachpb.ReplicationTarget{args.finalRelocationTargets()[0]}
		} else {
			__antithesis_instrumentation__.Notify(116705)
		}
		__antithesis_instrumentation__.Notify(116699)

		candidateDescs := make([]roachpb.StoreDescriptor, 0, len(candidateTargets))
		for _, candidate := range candidateTargets {
			__antithesis_instrumentation__.Notify(116706)
			store, ok := storeMap[candidate.StoreID]
			if !ok {
				__antithesis_instrumentation__.Notify(116708)
				return nil, nil, fmt.Errorf(
					"cannot up-replicate to s%d; missing gossiped StoreDescriptor"+
						" (the store is likely dead, draining or decommissioning)", candidate.StoreID,
				)
			} else {
				__antithesis_instrumentation__.Notify(116709)
			}
			__antithesis_instrumentation__.Notify(116707)
			candidateDescs = append(candidateDescs, *store)
		}
		__antithesis_instrumentation__.Notify(116700)
		candidateStoreList := makeStoreList(candidateDescs)

		additionTarget, _ = r.store.allocator.allocateTargetFromList(
			ctx,
			candidateStoreList,
			conf,
			existingVoters,
			existingNonVoters,
			r.store.allocator.scorerOptions(),

			true,
			args.targetType,
		)
		if roachpb.Empty(additionTarget) {
			__antithesis_instrumentation__.Notify(116710)
			return nil, nil, fmt.Errorf(
				"none of the remaining %ss %v are legal additions to %v",
				args.targetType, args.targetsToAdd(), desc.Replicas(),
			)
		} else {
			__antithesis_instrumentation__.Notify(116711)
		}
		__antithesis_instrumentation__.Notify(116701)

		if args.targetType == voterTarget {
			__antithesis_instrumentation__.Notify(116712)
			existingVoters = append(
				existingVoters, roachpb.ReplicaDescriptor{
					NodeID:    additionTarget.NodeID,
					StoreID:   additionTarget.StoreID,
					ReplicaID: desc.NextReplicaID,
					Type:      roachpb.ReplicaTypeVoterFull(),
				},
			)

			for i, nonVoter := range existingNonVoters {
				__antithesis_instrumentation__.Notify(116713)
				if nonVoter.StoreID == additionTarget.StoreID {
					__antithesis_instrumentation__.Notify(116714)
					canPromoteNonVoter = true

					existingNonVoters[i] = existingNonVoters[len(existingNonVoters)-1]
					existingNonVoters = existingNonVoters[:len(existingNonVoters)-1]
					break
				} else {
					__antithesis_instrumentation__.Notify(116715)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(116716)
			existingNonVoters = append(
				existingNonVoters, roachpb.ReplicaDescriptor{
					NodeID:    additionTarget.NodeID,
					StoreID:   additionTarget.StoreID,
					ReplicaID: desc.NextReplicaID,
					Type:      roachpb.ReplicaTypeNonVoter(),
				},
			)
		}
		__antithesis_instrumentation__.Notify(116702)
		shouldAdd = true
	} else {
		__antithesis_instrumentation__.Notify(116717)
	}
	__antithesis_instrumentation__.Notify(116688)

	lhRemovalAllowed := false
	var transferTarget *roachpb.ReplicationTarget
	if len(args.targetsToRemove()) > 0 {
		__antithesis_instrumentation__.Notify(116718)

		targetStore, _, err := r.store.allocator.removeTarget(
			ctx,
			conf,
			r.store.allocator.storeListForTargets(args.targetsToRemove()),
			existingVoters,
			existingNonVoters,
			args.targetType,
			r.store.allocator.scorerOptions(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116721)
			return nil, nil, errors.Wrapf(
				err, "unable to select removal target from %v; current replicas %v",
				args.targetsToRemove(), existingReplicas,
			)
		} else {
			__antithesis_instrumentation__.Notify(116722)
		}
		__antithesis_instrumentation__.Notify(116719)
		removalTarget = roachpb.ReplicationTarget{
			NodeID:  targetStore.NodeID,
			StoreID: targetStore.StoreID,
		}

		var b kv.Batch
		liReq := &roachpb.LeaseInfoRequest{}
		liReq.Key = desc.StartKey.AsRawKey()
		b.AddRawRequest(liReq)
		if err := r.store.DB().Run(ctx, &b); err != nil {
			__antithesis_instrumentation__.Notify(116723)
			return nil, nil, errors.Wrap(err, "looking up lease")
		} else {
			__antithesis_instrumentation__.Notify(116724)
		}
		__antithesis_instrumentation__.Notify(116720)

		lhRemovalAllowed = len(args.votersToAdd) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(116725)
			return r.store.cfg.Settings.Version.IsActive(ctx, clusterversion.EnableLeaseHolderRemoval) == true
		}() == true
		curLeaseholder := b.RawResponse().Responses[0].GetLeaseInfo().Lease.Replica
		shouldRemove = (curLeaseholder.StoreID != removalTarget.StoreID) || func() bool {
			__antithesis_instrumentation__.Notify(116726)
			return lhRemovalAllowed == true
		}() == true
		if args.targetType == voterTarget {
			__antithesis_instrumentation__.Notify(116727)

			for _, target := range args.nonVotersToAdd {
				__antithesis_instrumentation__.Notify(116729)
				if target.StoreID == removalTarget.StoreID {
					__antithesis_instrumentation__.Notify(116730)
					canDemoteVoter = true
				} else {
					__antithesis_instrumentation__.Notify(116731)
				}
			}
			__antithesis_instrumentation__.Notify(116728)
			if !shouldRemove {
				__antithesis_instrumentation__.Notify(116732)

				added := 0
				if shouldAdd {
					__antithesis_instrumentation__.Notify(116735)
					added++
				} else {
					__antithesis_instrumentation__.Notify(116736)
				}
				__antithesis_instrumentation__.Notify(116733)
				sortedTargetReplicas := append(
					[]roachpb.ReplicaDescriptor(nil),
					existingVoters[:len(existingVoters)-added]...,
				)
				sort.Slice(
					sortedTargetReplicas, func(i, j int) bool {
						__antithesis_instrumentation__.Notify(116737)
						sl := sortedTargetReplicas

						return sl[i].StoreID == args.finalRelocationTargets()[0].StoreID
					},
				)
				__antithesis_instrumentation__.Notify(116734)
				for _, rDesc := range sortedTargetReplicas {
					__antithesis_instrumentation__.Notify(116738)
					if rDesc.StoreID != curLeaseholder.StoreID {
						__antithesis_instrumentation__.Notify(116739)
						transferTarget = &roachpb.ReplicationTarget{
							NodeID:  rDesc.NodeID,
							StoreID: rDesc.StoreID,
						}
						shouldRemove = true
						break
					} else {
						__antithesis_instrumentation__.Notify(116740)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(116741)
			}
		} else {
			__antithesis_instrumentation__.Notify(116742)
		}
	} else {
		__antithesis_instrumentation__.Notify(116743)
	}
	__antithesis_instrumentation__.Notify(116689)

	var ops []roachpb.ReplicationChange
	if shouldAdd && func() bool {
		__antithesis_instrumentation__.Notify(116744)
		return shouldRemove == true
	}() == true {
		__antithesis_instrumentation__.Notify(116745)
		ops, _, err = replicationChangesForRebalance(
			ctx, desc, len(existingVoters), additionTarget, removalTarget, args.targetType,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116746)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(116747)
		}
	} else {
		__antithesis_instrumentation__.Notify(116748)
		if shouldAdd {
			__antithesis_instrumentation__.Notify(116749)
			if canPromoteNonVoter {
				__antithesis_instrumentation__.Notify(116750)
				ops = roachpb.ReplicationChangesForPromotion(additionTarget)
			} else {
				__antithesis_instrumentation__.Notify(116751)
				ops = roachpb.MakeReplicationChanges(args.targetType.AddChangeType(), additionTarget)
			}
		} else {
			__antithesis_instrumentation__.Notify(116752)
			if shouldRemove {
				__antithesis_instrumentation__.Notify(116753)

				if canDemoteVoter {
					__antithesis_instrumentation__.Notify(116754)
					ops = roachpb.ReplicationChangesForDemotion(removalTarget)
				} else {
					__antithesis_instrumentation__.Notify(116755)
					ops = roachpb.MakeReplicationChanges(args.targetType.RemoveChangeType(), removalTarget)
				}
			} else {
				__antithesis_instrumentation__.Notify(116756)
			}
		}
	}
	__antithesis_instrumentation__.Notify(116690)

	if len(ops) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(116757)
		return transferLeaseToFirstVoter == true
	}() == true {
		__antithesis_instrumentation__.Notify(116758)

		transferTarget = &voterTargets[0]
	} else {
		__antithesis_instrumentation__.Notify(116759)
	}
	__antithesis_instrumentation__.Notify(116691)
	return ops, transferTarget, nil
}

func getRelocationArgs(
	desc *roachpb.RangeDescriptor, voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
) relocationArgs {
	__antithesis_instrumentation__.Notify(116760)
	args := relocationArgs{
		votersToAdd: subtractTargets(
			voterTargets,
			desc.Replicas().Voters().ReplicationTargets(),
		),
		votersToRemove: subtractTargets(
			desc.Replicas().Voters().ReplicationTargets(),
			voterTargets,
		),
		nonVotersToAdd: subtractTargets(
			nonVoterTargets,
			desc.Replicas().NonVoters().ReplicationTargets(),
		),
		nonVotersToRemove: subtractTargets(
			desc.Replicas().NonVoters().ReplicationTargets(),
			nonVoterTargets,
		),
		finalVoterTargets:    voterTargets,
		finalNonVoterTargets: nonVoterTargets,
		targetType:           voterTarget,
	}

	if len(args.votersToAdd) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(116762)
		return len(args.votersToRemove) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(116763)
		args.targetType = nonVoterTarget
	} else {
		__antithesis_instrumentation__.Notify(116764)
	}
	__antithesis_instrumentation__.Notify(116761)
	return args
}

func containsDuplicates(targets []roachpb.ReplicationTarget) bool {
	__antithesis_instrumentation__.Notify(116765)
	for i := range targets {
		__antithesis_instrumentation__.Notify(116767)
		for j := i + 1; j < len(targets); j++ {
			__antithesis_instrumentation__.Notify(116768)
			if targets[i] == targets[j] {
				__antithesis_instrumentation__.Notify(116769)
				return true
			} else {
				__antithesis_instrumentation__.Notify(116770)
			}
		}
	}
	__antithesis_instrumentation__.Notify(116766)
	return false
}

func subtractTargets(left, right []roachpb.ReplicationTarget) (diff []roachpb.ReplicationTarget) {
	__antithesis_instrumentation__.Notify(116771)
	for _, t := range left {
		__antithesis_instrumentation__.Notify(116773)
		found := false
		for _, replicaDesc := range right {
			__antithesis_instrumentation__.Notify(116775)
			if replicaDesc.StoreID == t.StoreID && func() bool {
				__antithesis_instrumentation__.Notify(116776)
				return replicaDesc.NodeID == t.NodeID == true
			}() == true {
				__antithesis_instrumentation__.Notify(116777)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(116778)
			}
		}
		__antithesis_instrumentation__.Notify(116774)
		if !found {
			__antithesis_instrumentation__.Notify(116779)
			diff = append(diff, t)
		} else {
			__antithesis_instrumentation__.Notify(116780)
		}
	}
	__antithesis_instrumentation__.Notify(116772)
	return diff
}

func intersectTargets(
	left, right []roachpb.ReplicationTarget,
) (intersection []roachpb.ReplicationTarget) {
	__antithesis_instrumentation__.Notify(116781)
	isInLeft := func(id roachpb.StoreID) bool {
		__antithesis_instrumentation__.Notify(116784)
		for _, r := range left {
			__antithesis_instrumentation__.Notify(116786)
			if r.StoreID == id {
				__antithesis_instrumentation__.Notify(116787)
				return true
			} else {
				__antithesis_instrumentation__.Notify(116788)
			}
		}
		__antithesis_instrumentation__.Notify(116785)
		return false
	}
	__antithesis_instrumentation__.Notify(116782)
	for i := range right {
		__antithesis_instrumentation__.Notify(116789)
		if isInLeft(right[i].StoreID) {
			__antithesis_instrumentation__.Notify(116790)
			intersection = append(intersection, right[i])
		} else {
			__antithesis_instrumentation__.Notify(116791)
		}
	}
	__antithesis_instrumentation__.Notify(116783)
	return intersection
}

func (r *Replica) adminScatter(
	ctx context.Context, args roachpb.AdminScatterRequest,
) (roachpb.AdminScatterResponse, error) {
	__antithesis_instrumentation__.Notify(116792)
	rq := r.store.replicateQueue
	retryOpts := retry.Options{
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2,
		MaxRetries:     5,
	}

	maxAttempts := len(r.Desc().Replicas().Descriptors())
	currentAttempt := 0

	if args.MaxSize > 0 {
		__antithesis_instrumentation__.Notify(116797)
		if existing, limit := r.GetMVCCStats().Total(), args.MaxSize; existing > limit {
			__antithesis_instrumentation__.Notify(116798)
			return roachpb.AdminScatterResponse{}, errors.Errorf("existing range size %d exceeds specified limit %d", existing, limit)
		} else {
			__antithesis_instrumentation__.Notify(116799)
		}
	} else {
		__antithesis_instrumentation__.Notify(116800)
	}
	__antithesis_instrumentation__.Notify(116793)

	var allowLeaseTransfer bool
	var err error
	requeue := true
	canTransferLease := func(ctx context.Context, repl *Replica) bool {
		__antithesis_instrumentation__.Notify(116801)
		return allowLeaseTransfer
	}
	__antithesis_instrumentation__.Notify(116794)
	for re := retry.StartWithCtx(ctx, retryOpts); re.Next(); {
		__antithesis_instrumentation__.Notify(116802)
		if currentAttempt == maxAttempts {
			__antithesis_instrumentation__.Notify(116806)
			break
		} else {
			__antithesis_instrumentation__.Notify(116807)
		}
		__antithesis_instrumentation__.Notify(116803)
		if currentAttempt == maxAttempts-1 || func() bool {
			__antithesis_instrumentation__.Notify(116808)
			return !requeue == true
		}() == true {
			__antithesis_instrumentation__.Notify(116809)
			allowLeaseTransfer = true
		} else {
			__antithesis_instrumentation__.Notify(116810)
		}
		__antithesis_instrumentation__.Notify(116804)
		requeue, err = rq.processOneChange(
			ctx, r, canTransferLease, true, false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(116811)

			if isSnapshotError(err) {
				__antithesis_instrumentation__.Notify(116813)
				continue
			} else {
				__antithesis_instrumentation__.Notify(116814)
			}
			__antithesis_instrumentation__.Notify(116812)
			break
		} else {
			__antithesis_instrumentation__.Notify(116815)
		}
		__antithesis_instrumentation__.Notify(116805)
		currentAttempt++
		re.Reset()
	}
	__antithesis_instrumentation__.Notify(116795)

	if args.RandomizeLeases && func() bool {
		__antithesis_instrumentation__.Notify(116816)
		return r.OwnsValidLease(ctx, r.store.Clock().NowAsClockTimestamp()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(116817)
		desc := r.Desc()

		voterReplicas := desc.Replicas().VoterDescriptors()
		newLeaseholderIdx := rand.Intn(len(voterReplicas))
		targetStoreID := voterReplicas[newLeaseholderIdx].StoreID
		if targetStoreID != r.store.StoreID() {
			__antithesis_instrumentation__.Notify(116818)
			if err := r.AdminTransferLease(ctx, targetStoreID); err != nil {
				__antithesis_instrumentation__.Notify(116819)
				log.Warningf(ctx, "failed to scatter lease to s%d: %+v", targetStoreID, err)
			} else {
				__antithesis_instrumentation__.Notify(116820)
			}
		} else {
			__antithesis_instrumentation__.Notify(116821)
		}
	} else {
		__antithesis_instrumentation__.Notify(116822)
	}
	__antithesis_instrumentation__.Notify(116796)

	ri := r.GetRangeInfo(ctx)
	stats := r.GetMVCCStats()
	return roachpb.AdminScatterResponse{
		RangeInfos: []roachpb.RangeInfo{ri},
		MVCCStats:  &stats,
	}, nil
}

func (r *Replica) adminVerifyProtectedTimestamp(
	ctx context.Context, _ roachpb.AdminVerifyProtectedTimestampRequest,
) (resp roachpb.AdminVerifyProtectedTimestampResponse, err error) {
	__antithesis_instrumentation__.Notify(116823)

	resp.Verified = true
	return resp, nil
}
