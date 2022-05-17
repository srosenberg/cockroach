package loqrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type PrepareStoreReport struct {
	MissingStores []roachpb.StoreID

	UpdatedReplicas []PrepareReplicaReport

	SkippedReplicas []PrepareReplicaReport
}

type PrepareReplicaReport struct {
	Replica    roachpb.ReplicaDescriptor
	Descriptor roachpb.RangeDescriptor
	OldReplica roachpb.ReplicaDescriptor

	AlreadyUpdated bool

	RemovedReplicas roachpb.ReplicaSet

	AbortedTransaction   bool
	AbortedTransactionID uuid.UUID
}

func (r PrepareReplicaReport) RangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(108733)
	return r.Descriptor.RangeID
}

func (r PrepareReplicaReport) StartKey() roachpb.RKey {
	__antithesis_instrumentation__.Notify(108734)
	return r.Descriptor.StartKey
}

func PrepareUpdateReplicas(
	ctx context.Context,
	plan loqrecoverypb.ReplicaUpdatePlan,
	uuidGen uuid.Generator,
	updateTime time.Time,
	nodeID roachpb.NodeID,
	batches map[roachpb.StoreID]storage.Batch,
) (PrepareStoreReport, error) {
	__antithesis_instrumentation__.Notify(108735)
	var report PrepareStoreReport

	missing := make(map[roachpb.StoreID]struct{})
	for _, update := range plan.Updates {
		__antithesis_instrumentation__.Notify(108738)
		if nodeID != update.NodeID() {
			__antithesis_instrumentation__.Notify(108740)
			continue
		} else {
			__antithesis_instrumentation__.Notify(108741)
		}
		__antithesis_instrumentation__.Notify(108739)
		if readWriter, ok := batches[update.StoreID()]; !ok {
			__antithesis_instrumentation__.Notify(108742)
			missing[update.StoreID()] = struct{}{}
			continue
		} else {
			__antithesis_instrumentation__.Notify(108743)
			replicaReport, err := applyReplicaUpdate(ctx, readWriter, update)
			if err != nil {
				__antithesis_instrumentation__.Notify(108745)
				return PrepareStoreReport{}, errors.Wrapf(
					err,
					"failed to prepare update replica for range r%v on store s%d", update.RangeID,
					update.StoreID())
			} else {
				__antithesis_instrumentation__.Notify(108746)
			}
			__antithesis_instrumentation__.Notify(108744)
			if !replicaReport.AlreadyUpdated {
				__antithesis_instrumentation__.Notify(108747)
				report.UpdatedReplicas = append(report.UpdatedReplicas, replicaReport)
				uuid, err := uuidGen.NewV1()
				if err != nil {
					__antithesis_instrumentation__.Notify(108749)
					return PrepareStoreReport{}, errors.Wrap(err,
						"failed to generate uuid to write replica recovery evidence record")
				} else {
					__antithesis_instrumentation__.Notify(108750)
				}
				__antithesis_instrumentation__.Notify(108748)
				if err := writeReplicaRecoveryStoreRecord(
					uuid, updateTime.UnixNano(), update, replicaReport, readWriter); err != nil {
					__antithesis_instrumentation__.Notify(108751)
					return PrepareStoreReport{}, errors.Wrap(err,
						"failed writing replica recovery evidence record")
				} else {
					__antithesis_instrumentation__.Notify(108752)
				}
			} else {
				__antithesis_instrumentation__.Notify(108753)
				report.SkippedReplicas = append(report.SkippedReplicas, replicaReport)
			}
		}
	}
	__antithesis_instrumentation__.Notify(108736)

	if len(missing) > 0 {
		__antithesis_instrumentation__.Notify(108754)
		report.MissingStores = storeSliceFromSet(missing)
	} else {
		__antithesis_instrumentation__.Notify(108755)
	}
	__antithesis_instrumentation__.Notify(108737)
	return report, nil
}

func applyReplicaUpdate(
	ctx context.Context, readWriter storage.ReadWriter, update loqrecoverypb.ReplicaUpdate,
) (PrepareReplicaReport, error) {
	__antithesis_instrumentation__.Notify(108756)
	clock := hlc.NewClock(hlc.UnixNano, 0)
	report := PrepareReplicaReport{
		Replica: update.NewReplica,
	}

	key := keys.RangeDescriptorKey(update.StartKey.AsRKey())
	value, intent, err := storage.MVCCGet(
		ctx, readWriter, key, clock.Now(), storage.MVCCGetOptions{Inconsistent: true})
	if value == nil {
		__antithesis_instrumentation__.Notify(108767)
		return PrepareReplicaReport{}, errors.Errorf(
			"failed to find a range descriptor for range %v", key)
	} else {
		__antithesis_instrumentation__.Notify(108768)
	}
	__antithesis_instrumentation__.Notify(108757)
	if err != nil {
		__antithesis_instrumentation__.Notify(108769)
		return PrepareReplicaReport{}, err
	} else {
		__antithesis_instrumentation__.Notify(108770)
	}
	__antithesis_instrumentation__.Notify(108758)
	var localDesc roachpb.RangeDescriptor
	if err := value.GetProto(&localDesc); err != nil {
		__antithesis_instrumentation__.Notify(108771)
		return PrepareReplicaReport{}, err
	} else {
		__antithesis_instrumentation__.Notify(108772)
	}
	__antithesis_instrumentation__.Notify(108759)

	if localDesc.RangeID != update.RangeID {
		__antithesis_instrumentation__.Notify(108773)
		return PrepareReplicaReport{}, errors.Errorf(
			"unexpected range ID at key: expected r%d but found r%d", update.RangeID, localDesc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(108774)
	}
	__antithesis_instrumentation__.Notify(108760)

	if len(localDesc.InternalReplicas) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(108775)
		return localDesc.InternalReplicas[0].ReplicaID == update.NewReplica.ReplicaID == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108776)
		return localDesc.NextReplicaID == update.NextReplicaID == true
	}() == true {
		__antithesis_instrumentation__.Notify(108777)
		report.AlreadyUpdated = true
		return report, nil
	} else {
		__antithesis_instrumentation__.Notify(108778)
	}
	__antithesis_instrumentation__.Notify(108761)

	sl := stateloader.Make(localDesc.RangeID)
	ms, err := sl.LoadMVCCStats(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(108779)
		return PrepareReplicaReport{}, errors.Wrap(err, "loading MVCCStats")
	} else {
		__antithesis_instrumentation__.Notify(108780)
	}
	__antithesis_instrumentation__.Notify(108762)

	if intent != nil {
		__antithesis_instrumentation__.Notify(108781)

		report.AbortedTransaction = true
		report.AbortedTransactionID = intent.Txn.ID

		txnKey := keys.TransactionKey(intent.Txn.Key, intent.Txn.ID)
		if err := storage.MVCCDelete(ctx, readWriter, &ms, txnKey, hlc.Timestamp{}, nil); err != nil {
			__antithesis_instrumentation__.Notify(108784)
			return PrepareReplicaReport{}, err
		} else {
			__antithesis_instrumentation__.Notify(108785)
		}
		__antithesis_instrumentation__.Notify(108782)
		update := roachpb.LockUpdate{
			Span:   roachpb.Span{Key: intent.Key},
			Txn:    intent.Txn,
			Status: roachpb.ABORTED,
		}
		if _, err := storage.MVCCResolveWriteIntent(ctx, readWriter, &ms, update); err != nil {
			__antithesis_instrumentation__.Notify(108786)
			return PrepareReplicaReport{}, err
		} else {
			__antithesis_instrumentation__.Notify(108787)
		}
		__antithesis_instrumentation__.Notify(108783)
		report.AbortedTransaction = true
		report.AbortedTransactionID = intent.Txn.ID
	} else {
		__antithesis_instrumentation__.Notify(108788)
	}
	__antithesis_instrumentation__.Notify(108763)
	newDesc := localDesc
	replicas := []roachpb.ReplicaDescriptor{
		{
			NodeID:    update.NewReplica.NodeID,
			StoreID:   update.NewReplica.StoreID,
			ReplicaID: update.NewReplica.ReplicaID,
			Type:      update.NewReplica.Type,
		},
	}
	newDesc.SetReplicas(roachpb.MakeReplicaSet(replicas))
	newDesc.NextReplicaID = update.NextReplicaID

	if err := storage.MVCCPutProto(
		ctx, readWriter, &ms, key, clock.Now(),
		nil, &newDesc); err != nil {
		__antithesis_instrumentation__.Notify(108789)
		return PrepareReplicaReport{}, err
	} else {
		__antithesis_instrumentation__.Notify(108790)
	}
	__antithesis_instrumentation__.Notify(108764)
	report.Descriptor = newDesc
	report.RemovedReplicas = localDesc.Replicas()
	report.OldReplica, _ = report.RemovedReplicas.RemoveReplica(
		update.NewReplica.NodeID, update.NewReplica.StoreID)

	if err := sl.SetRaftReplicaID(ctx, readWriter, update.NewReplica.ReplicaID); err != nil {
		__antithesis_instrumentation__.Notify(108791)
		return PrepareReplicaReport{}, errors.Wrap(err, "setting new replica ID")
	} else {
		__antithesis_instrumentation__.Notify(108792)
	}
	__antithesis_instrumentation__.Notify(108765)

	if err := sl.SetMVCCStats(ctx, readWriter, &ms); err != nil {
		__antithesis_instrumentation__.Notify(108793)
		return PrepareReplicaReport{}, errors.Wrap(err, "updating MVCCStats")
	} else {
		__antithesis_instrumentation__.Notify(108794)
	}
	__antithesis_instrumentation__.Notify(108766)

	return report, nil
}

type ApplyUpdateReport struct {
	UpdatedStores []roachpb.StoreID
}

func CommitReplicaChanges(batches map[roachpb.StoreID]storage.Batch) (ApplyUpdateReport, error) {
	__antithesis_instrumentation__.Notify(108795)
	var report ApplyUpdateReport
	failed := false
	var updateErrors []string

	for id, batch := range batches {
		__antithesis_instrumentation__.Notify(108798)
		if batch.Empty() {
			__antithesis_instrumentation__.Notify(108800)
			continue
		} else {
			__antithesis_instrumentation__.Notify(108801)
		}
		__antithesis_instrumentation__.Notify(108799)
		if err := batch.Commit(true); err != nil {
			__antithesis_instrumentation__.Notify(108802)

			updateErrors = append(updateErrors, fmt.Sprintf("failed to update store s%d: %v", id, err))
			failed = true
		} else {
			__antithesis_instrumentation__.Notify(108803)
			report.UpdatedStores = append(report.UpdatedStores, id)
		}
	}
	__antithesis_instrumentation__.Notify(108796)
	if failed {
		__antithesis_instrumentation__.Notify(108804)
		return report, errors.Errorf(
			"failed to commit update to one or more stores: %s", strings.Join(updateErrors, "; "))
	} else {
		__antithesis_instrumentation__.Notify(108805)
	}
	__antithesis_instrumentation__.Notify(108797)
	return report, nil
}
