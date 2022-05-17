package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const FirstNodeID = roachpb.NodeID(1)

const FirstStoreID = roachpb.StoreID(1)

func InitEngine(ctx context.Context, eng storage.Engine, ident roachpb.StoreIdent) error {
	__antithesis_instrumentation__.Notify(124903)
	exIdent, err := ReadStoreIdent(ctx, eng)
	if err == nil {
		__antithesis_instrumentation__.Notify(124909)
		return errors.Errorf("engine %s is already initialized with ident %s", eng, exIdent.String())
	} else {
		__antithesis_instrumentation__.Notify(124910)
	}
	__antithesis_instrumentation__.Notify(124904)
	if !errors.HasType(err, (*NotBootstrappedError)(nil)) {
		__antithesis_instrumentation__.Notify(124911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124912)
	}
	__antithesis_instrumentation__.Notify(124905)

	if err := checkCanInitializeEngine(ctx, eng); err != nil {
		__antithesis_instrumentation__.Notify(124913)
		return errors.Wrap(err, "while trying to initialize engine")
	} else {
		__antithesis_instrumentation__.Notify(124914)
	}
	__antithesis_instrumentation__.Notify(124906)

	batch := eng.NewBatch()
	if err := storage.MVCCPutProto(
		ctx,
		batch,
		nil,
		keys.StoreIdentKey(),
		hlc.Timestamp{},
		nil,
		&ident,
	); err != nil {
		__antithesis_instrumentation__.Notify(124915)
		batch.Close()
		return err
	} else {
		__antithesis_instrumentation__.Notify(124916)
	}
	__antithesis_instrumentation__.Notify(124907)
	if err := batch.Commit(true); err != nil {
		__antithesis_instrumentation__.Notify(124917)
		return errors.Wrap(err, "persisting engine initialization data")
	} else {
		__antithesis_instrumentation__.Notify(124918)
	}
	__antithesis_instrumentation__.Notify(124908)

	return nil
}

func WriteInitialClusterData(
	ctx context.Context,
	eng storage.Engine,
	initialValues []roachpb.KeyValue,
	bootstrapVersion roachpb.Version,
	numStores int,
	splits []roachpb.RKey,
	nowNanos int64,
	knobs StoreTestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(124919)

	bootstrapVal := roachpb.Value{}
	if err := bootstrapVal.SetProto(&bootstrapVersion); err != nil {
		__antithesis_instrumentation__.Notify(124925)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124926)
	}
	__antithesis_instrumentation__.Notify(124920)
	initialValues = append(initialValues,
		roachpb.KeyValue{Key: keys.BootstrapVersionKey, Value: bootstrapVal})

	var nodeIDVal, storeIDVal, rangeIDVal, livenessVal roachpb.Value

	nodeIDVal.SetInt(int64(FirstNodeID))

	storeIDVal.SetInt(int64(FirstStoreID) + int64(numStores) - 1)

	rangeIDVal.SetInt(int64(len(splits) + 1))

	livenessRecord := livenesspb.Liveness{NodeID: FirstNodeID, Epoch: 0}
	if err := livenessVal.SetProto(&livenessRecord); err != nil {
		__antithesis_instrumentation__.Notify(124927)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124928)
	}
	__antithesis_instrumentation__.Notify(124921)
	initialValues = append(initialValues,
		roachpb.KeyValue{Key: keys.NodeIDGenerator, Value: nodeIDVal},
		roachpb.KeyValue{Key: keys.StoreIDGenerator, Value: storeIDVal},
		roachpb.KeyValue{Key: keys.RangeIDGenerator, Value: rangeIDVal},
		roachpb.KeyValue{Key: keys.NodeLivenessKey(FirstNodeID), Value: livenessVal})

	firstRangeMS := &enginepb.MVCCStats{}

	filterInitialValues := func(desc *roachpb.RangeDescriptor) []roachpb.KeyValue {
		__antithesis_instrumentation__.Notify(124929)
		var r []roachpb.KeyValue
		for _, kv := range initialValues {
			__antithesis_instrumentation__.Notify(124931)
			if desc.ContainsKey(roachpb.RKey(kv.Key)) {
				__antithesis_instrumentation__.Notify(124932)
				r = append(r, kv)
			} else {
				__antithesis_instrumentation__.Notify(124933)
			}
		}
		__antithesis_instrumentation__.Notify(124930)
		return r
	}
	__antithesis_instrumentation__.Notify(124922)

	initialReplicaVersion := bootstrapVersion
	if knobs.InitialReplicaVersionOverride != nil {
		__antithesis_instrumentation__.Notify(124934)
		initialReplicaVersion = *knobs.InitialReplicaVersionOverride
	} else {
		__antithesis_instrumentation__.Notify(124935)
	}
	__antithesis_instrumentation__.Notify(124923)

	startKey := roachpb.RKeyMax
	for i := len(splits) - 1; i >= -1; i-- {
		__antithesis_instrumentation__.Notify(124936)
		endKey := startKey
		rangeID := roachpb.RangeID(i + 2)
		if i >= 0 {
			__antithesis_instrumentation__.Notify(124947)
			startKey = splits[i]
		} else {
			__antithesis_instrumentation__.Notify(124948)
			startKey = roachpb.RKeyMin
		}
		__antithesis_instrumentation__.Notify(124937)

		desc := &roachpb.RangeDescriptor{
			RangeID:       rangeID,
			StartKey:      startKey,
			EndKey:        endKey,
			NextReplicaID: 2,
		}
		const firstReplicaID = 1
		replicas := []roachpb.ReplicaDescriptor{
			{
				NodeID:    FirstNodeID,
				StoreID:   FirstStoreID,
				ReplicaID: firstReplicaID,
			},
		}
		desc.SetReplicas(roachpb.MakeReplicaSet(replicas))
		if err := desc.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(124949)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124950)
		}
		__antithesis_instrumentation__.Notify(124938)
		rangeInitialValues := filterInitialValues(desc)
		log.VEventf(
			ctx, 2, "creating range %d [%s, %s). Initial values: %d",
			desc.RangeID, desc.StartKey, desc.EndKey, len(rangeInitialValues))
		batch := eng.NewBatch()
		defer batch.Close()

		now := hlc.Timestamp{
			WallTime: nowNanos,
			Logical:  0,
		}

		if err := storage.MVCCPutProto(
			ctx, batch, nil, keys.RangeDescriptorKey(desc.StartKey),
			now, nil, desc,
		); err != nil {
			__antithesis_instrumentation__.Notify(124951)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124952)
		}
		__antithesis_instrumentation__.Notify(124939)

		if err := storage.MVCCPutProto(
			ctx, batch, nil, keys.RangeLastReplicaGCTimestampKey(desc.RangeID),
			hlc.Timestamp{}, nil, &now,
		); err != nil {
			__antithesis_instrumentation__.Notify(124953)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124954)
		}
		__antithesis_instrumentation__.Notify(124940)

		meta2Key := keys.RangeMetaKey(endKey)
		if err := storage.MVCCPutProto(ctx, batch, firstRangeMS, meta2Key.AsRawKey(),
			now, nil, desc,
		); err != nil {
			__antithesis_instrumentation__.Notify(124955)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124956)
		}
		__antithesis_instrumentation__.Notify(124941)

		if startKey.Equal(roachpb.RKeyMin) {
			__antithesis_instrumentation__.Notify(124957)

			meta1Key := keys.RangeMetaKey(keys.RangeMetaKey(roachpb.RKeyMax))
			if err := storage.MVCCPutProto(
				ctx, batch, nil, meta1Key.AsRawKey(), now, nil, desc,
			); err != nil {
				__antithesis_instrumentation__.Notify(124958)
				return err
			} else {
				__antithesis_instrumentation__.Notify(124959)
			}
		} else {
			__antithesis_instrumentation__.Notify(124960)
		}
		__antithesis_instrumentation__.Notify(124942)

		for _, kv := range rangeInitialValues {
			__antithesis_instrumentation__.Notify(124961)

			kv.Value.InitChecksum(kv.Key)
			if err := storage.MVCCPut(
				ctx, batch, nil, kv.Key, now, kv.Value, nil,
			); err != nil {
				__antithesis_instrumentation__.Notify(124962)
				return err
			} else {
				__antithesis_instrumentation__.Notify(124963)
			}
		}
		__antithesis_instrumentation__.Notify(124943)

		if err := stateloader.WriteInitialRangeState(
			ctx, batch, *desc, firstReplicaID, initialReplicaVersion); err != nil {
			__antithesis_instrumentation__.Notify(124964)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124965)
		}
		__antithesis_instrumentation__.Notify(124944)
		computedStats, err := rditer.ComputeStatsForRange(desc, batch, now.WallTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(124966)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124967)
		}
		__antithesis_instrumentation__.Notify(124945)

		sl := stateloader.Make(rangeID)
		if err := sl.SetMVCCStats(ctx, batch, &computedStats); err != nil {
			__antithesis_instrumentation__.Notify(124968)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124969)
		}
		__antithesis_instrumentation__.Notify(124946)

		if err := batch.Commit(true); err != nil {
			__antithesis_instrumentation__.Notify(124970)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124971)
		}
	}
	__antithesis_instrumentation__.Notify(124924)

	return nil
}
