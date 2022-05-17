package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type relocateRange struct {
	optColumnsSlot

	rows            planNode
	subjectReplicas tree.RelocateSubject
	toStoreID       tree.TypedExpr
	fromStoreID     tree.TypedExpr
	run             relocateRunState
}

type relocateRunState struct {
	toStoreDesc   *roachpb.StoreDescriptor
	fromStoreDesc *roachpb.StoreDescriptor
	results       relocateResults
}

type relocateResults struct {
	rangeID   roachpb.RangeID
	rangeDesc *roachpb.RangeDescriptor
	err       error
}

type relocateRequest struct {
	rangeID         roachpb.RangeID
	subjectReplicas tree.RelocateSubject
	toStoreDesc     *roachpb.StoreDescriptor
	fromStoreDesc   *roachpb.StoreDescriptor
}

func (n *relocateRange) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565606)
	toStoreID, err := paramparse.DatumAsInt(params.EvalContext(), "TO", n.toStoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(565613)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565614)
	}
	__antithesis_instrumentation__.Notify(565607)
	var fromStoreID int64
	if n.subjectReplicas != tree.RelocateLease {
		__antithesis_instrumentation__.Notify(565615)

		fromStoreID, err = paramparse.DatumAsInt(params.EvalContext(), "FROM", n.fromStoreID)
		if err != nil {
			__antithesis_instrumentation__.Notify(565616)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565617)
		}
	} else {
		__antithesis_instrumentation__.Notify(565618)
	}
	__antithesis_instrumentation__.Notify(565608)

	if toStoreID <= 0 {
		__antithesis_instrumentation__.Notify(565619)
		return errors.Errorf("invalid target to store ID %d for RELOCATE", n.toStoreID)
	} else {
		__antithesis_instrumentation__.Notify(565620)
	}
	__antithesis_instrumentation__.Notify(565609)
	if n.subjectReplicas != tree.RelocateLease && func() bool {
		__antithesis_instrumentation__.Notify(565621)
		return fromStoreID <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(565622)
		return errors.Errorf("invalid target from store ID %d for RELOCATE", n.fromStoreID)
	} else {
		__antithesis_instrumentation__.Notify(565623)
	}
	__antithesis_instrumentation__.Notify(565610)

	n.run.toStoreDesc, err = lookupStoreDesc(roachpb.StoreID(toStoreID), params)
	if err != nil {
		__antithesis_instrumentation__.Notify(565624)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565625)
	}
	__antithesis_instrumentation__.Notify(565611)
	if n.subjectReplicas != tree.RelocateLease {
		__antithesis_instrumentation__.Notify(565626)
		n.run.fromStoreDesc, err = lookupStoreDesc(roachpb.StoreID(fromStoreID), params)
		if err != nil {
			__antithesis_instrumentation__.Notify(565627)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565628)
		}
	} else {
		__antithesis_instrumentation__.Notify(565629)
	}
	__antithesis_instrumentation__.Notify(565612)
	return nil
}

func (n *relocateRange) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565630)
	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(565633)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(565634)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(565635)
	}
	__antithesis_instrumentation__.Notify(565631)
	datum := n.rows.Values()[0]
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(565636)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(565637)
	}
	__antithesis_instrumentation__.Notify(565632)
	rangeID := roachpb.RangeID(tree.MustBeDInt(datum))

	rangeDesc, err := relocate(params, relocateRequest{
		rangeID:         rangeID,
		subjectReplicas: n.subjectReplicas,
		fromStoreDesc:   n.run.fromStoreDesc,
		toStoreDesc:     n.run.toStoreDesc,
	})

	n.run.results = relocateResults{
		rangeID:   rangeID,
		rangeDesc: rangeDesc,
		err:       err,
	}
	return true, nil
}

func (n *relocateRange) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565638)
	result := "ok"
	if n.run.results.err != nil {
		__antithesis_instrumentation__.Notify(565641)
		result = n.run.results.err.Error()
	} else {
		__antithesis_instrumentation__.Notify(565642)
	}
	__antithesis_instrumentation__.Notify(565639)
	pretty := ""
	if n.run.results.rangeDesc != nil {
		__antithesis_instrumentation__.Notify(565643)
		pretty = keys.PrettyPrint(nil, n.run.results.rangeDesc.StartKey.AsRawKey())
	} else {
		__antithesis_instrumentation__.Notify(565644)
	}
	__antithesis_instrumentation__.Notify(565640)
	return tree.Datums{
		tree.NewDInt(tree.DInt(n.run.results.rangeID)),
		tree.NewDString(pretty),
		tree.NewDString(result),
	}
}

func (n *relocateRange) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(565645)
	n.rows.Close(ctx)
}

func relocate(params runParams, req relocateRequest) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(565646)
	rangeDesc, err := lookupRangeDescriptorByRangeID(params.ctx, params.extendedEvalCtx.ExecCfg.DB, req.rangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(565650)
		return nil, errors.Wrapf(err, "error looking up range descriptor")
	} else {
		__antithesis_instrumentation__.Notify(565651)
	}
	__antithesis_instrumentation__.Notify(565647)

	if req.subjectReplicas == tree.RelocateLease {
		__antithesis_instrumentation__.Notify(565652)
		err := params.p.ExecCfg().DB.AdminTransferLease(params.ctx, rangeDesc.StartKey, req.toStoreDesc.StoreID)
		return rangeDesc, err
	} else {
		__antithesis_instrumentation__.Notify(565653)
	}
	__antithesis_instrumentation__.Notify(565648)

	toTarget := roachpb.ReplicationTarget{NodeID: req.toStoreDesc.Node.NodeID, StoreID: req.toStoreDesc.StoreID}
	fromTarget := roachpb.ReplicationTarget{NodeID: req.fromStoreDesc.Node.NodeID, StoreID: req.fromStoreDesc.StoreID}
	if req.subjectReplicas == tree.RelocateNonVoters {
		__antithesis_instrumentation__.Notify(565654)
		_, err := params.p.ExecCfg().DB.AdminChangeReplicas(
			params.ctx, rangeDesc.StartKey, *rangeDesc, []roachpb.ReplicationChange{
				{ChangeType: roachpb.ADD_NON_VOTER, Target: toTarget},
				{ChangeType: roachpb.REMOVE_NON_VOTER, Target: fromTarget},
			},
		)
		return rangeDesc, err
	} else {
		__antithesis_instrumentation__.Notify(565655)
	}
	__antithesis_instrumentation__.Notify(565649)
	_, err = params.p.ExecCfg().DB.AdminChangeReplicas(
		params.ctx, rangeDesc.StartKey, *rangeDesc, []roachpb.ReplicationChange{
			{ChangeType: roachpb.ADD_VOTER, Target: toTarget},
			{ChangeType: roachpb.REMOVE_VOTER, Target: fromTarget},
		},
	)
	return rangeDesc, err
}

func lookupRangeDescriptorByRangeID(
	ctx context.Context, db *kv.DB, rangeID roachpb.RangeID,
) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(565656)
	var descriptor roachpb.RangeDescriptor
	sentinelErr := errors.Errorf("sentinel")
	err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(565660)
		return txn.Iterate(ctx, keys.MetaMin, keys.MetaMax, 100,
			func(rows []kv.KeyValue) error {
				__antithesis_instrumentation__.Notify(565661)
				var desc roachpb.RangeDescriptor
				for _, row := range rows {
					__antithesis_instrumentation__.Notify(565663)
					err := row.ValueProto(&desc)
					if err != nil {
						__antithesis_instrumentation__.Notify(565665)
						return errors.Wrapf(err, "unable to unmarshal range descriptor from %s", row.Key)
					} else {
						__antithesis_instrumentation__.Notify(565666)
					}
					__antithesis_instrumentation__.Notify(565664)

					if desc.RangeID == rangeID {
						__antithesis_instrumentation__.Notify(565667)
						descriptor = desc
						return sentinelErr
					} else {
						__antithesis_instrumentation__.Notify(565668)
					}
				}
				__antithesis_instrumentation__.Notify(565662)
				return nil
			})
	})
	__antithesis_instrumentation__.Notify(565657)
	if errors.Is(err, sentinelErr) {
		__antithesis_instrumentation__.Notify(565669)
		return &descriptor, nil
	} else {
		__antithesis_instrumentation__.Notify(565670)
	}
	__antithesis_instrumentation__.Notify(565658)
	if err != nil {
		__antithesis_instrumentation__.Notify(565671)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565672)
	}
	__antithesis_instrumentation__.Notify(565659)
	return nil, errors.Errorf("Descriptor for range %d is not found", rangeID)
}
