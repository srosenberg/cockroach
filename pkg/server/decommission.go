package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func getPingCheckDecommissionFn(
	engines Engines,
) (*nodeTombstoneStorage, func(context.Context, roachpb.NodeID, codes.Code) error) {
	__antithesis_instrumentation__.Notify(190560)
	nodeTombStorage := &nodeTombstoneStorage{engs: engines}
	return nodeTombStorage, func(ctx context.Context, nodeID roachpb.NodeID, errorCode codes.Code) error {
		__antithesis_instrumentation__.Notify(190561)
		ts, err := nodeTombStorage.IsDecommissioned(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(190564)

			log.Fatalf(ctx, "unable to read decommissioned status for n%d: %v", nodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(190565)
		}
		__antithesis_instrumentation__.Notify(190562)
		if !ts.IsZero() {
			__antithesis_instrumentation__.Notify(190566)

			return grpcstatus.Errorf(errorCode,
				"n%d was permanently removed from the cluster at %s; it is not allowed to rejoin the cluster",
				nodeID, ts,
			)
		} else {
			__antithesis_instrumentation__.Notify(190567)
		}
		__antithesis_instrumentation__.Notify(190563)

		return nil
	}
}

func (s *Server) Decommission(
	ctx context.Context, targetStatus livenesspb.MembershipStatus, nodeIDs []roachpb.NodeID,
) error {
	__antithesis_instrumentation__.Notify(190568)

	if targetStatus == livenesspb.MembershipStatus_DECOMMISSIONED {
		__antithesis_instrumentation__.Notify(190572)
		orderedNodeIDs := make([]roachpb.NodeID, len(nodeIDs))
		copy(orderedNodeIDs, nodeIDs)
		sort.SliceStable(orderedNodeIDs, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(190574)
			return orderedNodeIDs[j] == s.NodeID()
		})
		__antithesis_instrumentation__.Notify(190573)
		nodeIDs = orderedNodeIDs
	} else {
		__antithesis_instrumentation__.Notify(190575)
	}
	__antithesis_instrumentation__.Notify(190569)

	var event eventpb.EventPayload
	var nodeDetails *eventpb.CommonNodeDecommissionDetails
	if targetStatus.Decommissioning() {
		__antithesis_instrumentation__.Notify(190576)
		ev := &eventpb.NodeDecommissioning{}
		nodeDetails = &ev.CommonNodeDecommissionDetails
		event = ev
	} else {
		__antithesis_instrumentation__.Notify(190577)
		if targetStatus.Decommissioned() {
			__antithesis_instrumentation__.Notify(190578)
			ev := &eventpb.NodeDecommissioned{}
			nodeDetails = &ev.CommonNodeDecommissionDetails
			event = ev
		} else {
			__antithesis_instrumentation__.Notify(190579)
			if targetStatus.Active() {
				__antithesis_instrumentation__.Notify(190580)
				ev := &eventpb.NodeRecommissioned{}
				nodeDetails = &ev.CommonNodeDecommissionDetails
				event = ev
			} else {
				__antithesis_instrumentation__.Notify(190581)
				panic("unexpected target membership status")
			}
		}
	}
	__antithesis_instrumentation__.Notify(190570)
	event.CommonDetails().Timestamp = timeutil.Now().UnixNano()
	nodeDetails.RequestingNodeID = int32(s.NodeID())

	for _, nodeID := range nodeIDs {
		__antithesis_instrumentation__.Notify(190582)
		statusChanged, err := s.nodeLiveness.SetMembershipStatus(ctx, nodeID, targetStatus)
		if err != nil {
			__antithesis_instrumentation__.Notify(190585)
			if errors.Is(err, liveness.ErrMissingRecord) {
				__antithesis_instrumentation__.Notify(190587)
				return grpcstatus.Error(codes.NotFound, liveness.ErrMissingRecord.Error())
			} else {
				__antithesis_instrumentation__.Notify(190588)
			}
			__antithesis_instrumentation__.Notify(190586)
			log.Errorf(ctx, "%+s", err)
			return grpcstatus.Errorf(codes.Internal, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(190589)
		}
		__antithesis_instrumentation__.Notify(190583)
		if statusChanged {
			__antithesis_instrumentation__.Notify(190590)
			nodeDetails.TargetNodeID = int32(nodeID)

			log.StructuredEvent(ctx, event)

			if err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(190591)
				return sql.InsertEventRecord(
					ctx,
					s.sqlServer.execCfg.InternalExecutor,
					txn,
					int32(s.NodeID()),
					sql.LogToSystemTable|sql.LogToDevChannelIfVerbose,
					int32(nodeID),
					event,
				)
			}); err != nil {
				__antithesis_instrumentation__.Notify(190592)
				log.Ops.Errorf(ctx, "unable to record event: %+v: %+v", event, err)
			} else {
				__antithesis_instrumentation__.Notify(190593)
			}
		} else {
			__antithesis_instrumentation__.Notify(190594)
		}
		__antithesis_instrumentation__.Notify(190584)

		if targetStatus.Decommissioned() {
			__antithesis_instrumentation__.Notify(190595)
			if err := s.db.PutInline(ctx, keys.NodeStatusKey(nodeID), nil); err != nil {
				__antithesis_instrumentation__.Notify(190596)
				log.Errorf(ctx, "unable to clean up node status data for node %d: %s", nodeID, err)
			} else {
				__antithesis_instrumentation__.Notify(190597)
			}
		} else {
			__antithesis_instrumentation__.Notify(190598)
		}
	}
	__antithesis_instrumentation__.Notify(190571)
	return nil
}
