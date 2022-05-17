package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type relocateNode struct {
	optColumnsSlot

	subjectReplicas tree.RelocateSubject
	tableDesc       catalog.TableDescriptor
	index           catalog.Index
	rows            planNode

	run relocateRun
}

type relocateRun struct {
	lastRangeStartKey []byte

	storeMap map[roachpb.StoreID]roachpb.NodeID
}

func (n *relocateNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(565542)
	n.run.storeMap = make(map[roachpb.StoreID]roachpb.NodeID)
	return nil
}

func (n *relocateNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565543)

	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(565549)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(565550)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(565551)
	}
	__antithesis_instrumentation__.Notify(565544)

	data := n.rows.Values()

	var relocationTargets []roachpb.ReplicationTarget
	var leaseStoreID roachpb.StoreID
	if n.subjectReplicas == tree.RelocateLease {
		__antithesis_instrumentation__.Notify(565552)
		leaseStoreID = roachpb.StoreID(tree.MustBeDInt(data[0]))
		if leaseStoreID <= 0 {
			__antithesis_instrumentation__.Notify(565553)
			return false, errors.Errorf("invalid target leaseholder store ID %d for EXPERIMENTAL_RELOCATE LEASE", leaseStoreID)
		} else {
			__antithesis_instrumentation__.Notify(565554)
		}
	} else {
		__antithesis_instrumentation__.Notify(565555)
		if !data[0].ResolvedType().Equivalent(types.IntArray) {
			__antithesis_instrumentation__.Notify(565558)
			return false, errors.Errorf(
				"expected int array in the first EXPERIMENTAL_RELOCATE data column; got %s",
				data[0].ResolvedType(),
			)
		} else {
			__antithesis_instrumentation__.Notify(565559)
		}
		__antithesis_instrumentation__.Notify(565556)
		relocation := data[0].(*tree.DArray)
		if n.subjectReplicas != tree.RelocateNonVoters && func() bool {
			__antithesis_instrumentation__.Notify(565560)
			return len(relocation.Array) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(565561)

			return false, errors.Errorf("empty relocation array for EXPERIMENTAL_RELOCATE")
		} else {
			__antithesis_instrumentation__.Notify(565562)
		}
		__antithesis_instrumentation__.Notify(565557)

		relocationTargets = make([]roachpb.ReplicationTarget, len(relocation.Array))
		for i, d := range relocation.Array {
			__antithesis_instrumentation__.Notify(565563)
			storeID := roachpb.StoreID(*d.(*tree.DInt))
			nodeID, ok := n.run.storeMap[storeID]
			if !ok {
				__antithesis_instrumentation__.Notify(565565)

				storeDesc, err := lookupStoreDesc(storeID, params)
				if err != nil {
					__antithesis_instrumentation__.Notify(565567)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(565568)
				}
				__antithesis_instrumentation__.Notify(565566)
				nodeID = storeDesc.Node.NodeID
				n.run.storeMap[storeID] = nodeID
			} else {
				__antithesis_instrumentation__.Notify(565569)
			}
			__antithesis_instrumentation__.Notify(565564)
			relocationTargets[i] = roachpb.ReplicationTarget{NodeID: nodeID, StoreID: storeID}
		}
	}
	__antithesis_instrumentation__.Notify(565545)

	rowKey, err := getRowKey(params.ExecCfg().Codec, n.tableDesc, n.index, data[1:])
	if err != nil {
		__antithesis_instrumentation__.Notify(565570)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(565571)
	}
	__antithesis_instrumentation__.Notify(565546)
	rowKey = keys.MakeFamilyKey(rowKey, 0)

	rangeDesc, err := lookupRangeDescriptor(params.ctx, params.extendedEvalCtx.ExecCfg.DB, rowKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(565572)
		return false, errors.Wrapf(err, "error looking up range descriptor")
	} else {
		__antithesis_instrumentation__.Notify(565573)
	}
	__antithesis_instrumentation__.Notify(565547)
	n.run.lastRangeStartKey = rangeDesc.StartKey.AsRawKey()

	existingVoters := rangeDesc.Replicas().Voters().ReplicationTargets()
	existingNonVoters := rangeDesc.Replicas().NonVoters().ReplicationTargets()
	switch n.subjectReplicas {
	case tree.RelocateLease:
		__antithesis_instrumentation__.Notify(565574)
		if err := params.p.ExecCfg().DB.AdminTransferLease(params.ctx, rowKey, leaseStoreID); err != nil {
			__antithesis_instrumentation__.Notify(565578)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(565579)
		}
	case tree.RelocateNonVoters:
		__antithesis_instrumentation__.Notify(565575)
		if err := params.p.ExecCfg().DB.AdminRelocateRange(
			params.ctx,
			rowKey,
			existingVoters,
			relocationTargets,
			true,
		); err != nil {
			__antithesis_instrumentation__.Notify(565580)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(565581)
		}
	case tree.RelocateVoters:
		__antithesis_instrumentation__.Notify(565576)
		if err := params.p.ExecCfg().DB.AdminRelocateRange(
			params.ctx,
			rowKey,
			relocationTargets,
			existingNonVoters,
			true,
		); err != nil {
			__antithesis_instrumentation__.Notify(565582)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(565583)
		}
	default:
		__antithesis_instrumentation__.Notify(565577)
		return false, errors.AssertionFailedf("unknown relocate mode: %v", n.subjectReplicas)
	}
	__antithesis_instrumentation__.Notify(565548)

	return true, nil
}

func (n *relocateNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565584)
	return tree.Datums{
		tree.NewDBytes(tree.DBytes(n.run.lastRangeStartKey)),
		tree.NewDString(keys.PrettyPrint(catalogkeys.IndexKeyValDirs(n.index), n.run.lastRangeStartKey)),
	}
}

func (n *relocateNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(565585)
	n.rows.Close(ctx)
}

func lookupStoreDesc(storeID roachpb.StoreID, params runParams) (*roachpb.StoreDescriptor, error) {
	__antithesis_instrumentation__.Notify(565586)
	var storeDesc roachpb.StoreDescriptor
	gossipStoreKey := gossip.MakeStoreKey(storeID)
	g, err := params.extendedEvalCtx.ExecCfg.Gossip.OptionalErr(54250)
	if err != nil {
		__antithesis_instrumentation__.Notify(565589)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565590)
	}
	__antithesis_instrumentation__.Notify(565587)
	if err := g.GetInfoProto(
		gossipStoreKey, &storeDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(565591)
		return nil, errors.Wrapf(err, "error looking up store %d", storeID)
	} else {
		__antithesis_instrumentation__.Notify(565592)
	}
	__antithesis_instrumentation__.Notify(565588)
	return &storeDesc, nil
}

func lookupRangeDescriptor(
	ctx context.Context, db *kv.DB, rowKey []byte,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(565593)
	startKey := keys.RangeMetaKey(keys.MustAddr(rowKey))
	endKey := keys.Meta2Prefix.PrefixEnd()
	kvs, err := db.Scan(ctx, startKey, endKey, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(565598)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(565599)
	}
	__antithesis_instrumentation__.Notify(565594)
	if len(kvs) != 1 {
		__antithesis_instrumentation__.Notify(565600)
		log.Fatalf(ctx, "expected 1 KV, got %v", kvs)
	} else {
		__antithesis_instrumentation__.Notify(565601)
	}
	__antithesis_instrumentation__.Notify(565595)
	var desc roachpb.RangeDescriptor
	if err := kvs[0].ValueProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(565602)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(565603)
	}
	__antithesis_instrumentation__.Notify(565596)
	if desc.EndKey.Equal(rowKey) {
		__antithesis_instrumentation__.Notify(565604)
		log.Fatalf(ctx, "row key should not be valid range split point: %s", keys.PrettyPrint(nil, rowKey))
	} else {
		__antithesis_instrumentation__.Notify(565605)
	}
	__antithesis_instrumentation__.Notify(565597)
	return desc, nil
}
