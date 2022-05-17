package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *statusServer) ProblemRanges(
	ctx context.Context, req *serverpb.ProblemRangesRequest,
) (*serverpb.ProblemRangesResponse, error) {
	__antithesis_instrumentation__.Notify(195358)
	ctx = s.AnnotateCtx(ctx)

	response := &serverpb.ProblemRangesResponse{
		NodeID:           s.gossip.NodeID.Get(),
		ProblemsByNodeID: make(map[roachpb.NodeID]serverpb.ProblemRangesResponse_NodeProblems),
	}

	isLiveMap := s.nodeLiveness.GetIsLiveMap()

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(195362)
		requestedNodeID, _, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(195364)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(195365)
		}
		__antithesis_instrumentation__.Notify(195363)
		isLiveMap = liveness.IsLiveMap{
			requestedNodeID: liveness.IsLiveMapEntry{IsLive: true},
		}
	} else {
		__antithesis_instrumentation__.Notify(195366)
	}
	__antithesis_instrumentation__.Notify(195359)

	type nodeResponse struct {
		nodeID roachpb.NodeID
		resp   *serverpb.RangesResponse
		err    error
	}

	responses := make(chan nodeResponse)

	for nodeID := range isLiveMap {
		__antithesis_instrumentation__.Notify(195367)
		nodeID := nodeID
		if err := s.stopper.RunAsyncTask(
			ctx, "server.statusServer: requesting remote ranges",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(195368)
				status, err := s.dialNode(ctx, nodeID)
				var rangesResponse *serverpb.RangesResponse
				if err == nil {
					__antithesis_instrumentation__.Notify(195370)
					req := &serverpb.RangesRequest{}
					rangesResponse, err = status.Ranges(ctx, req)
				} else {
					__antithesis_instrumentation__.Notify(195371)
				}
				__antithesis_instrumentation__.Notify(195369)
				response := nodeResponse{
					nodeID: nodeID,
					resp:   rangesResponse,
					err:    err,
				}

				select {
				case responses <- response:
					__antithesis_instrumentation__.Notify(195372)

				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(195373)

				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(195374)
			return nil, status.Errorf(codes.Internal, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(195375)
		}
	}
	__antithesis_instrumentation__.Notify(195360)

	for remainingResponses := len(isLiveMap); remainingResponses > 0; remainingResponses-- {
		__antithesis_instrumentation__.Notify(195376)
		select {
		case resp := <-responses:
			__antithesis_instrumentation__.Notify(195377)
			if resp.err != nil {
				__antithesis_instrumentation__.Notify(195381)
				response.ProblemsByNodeID[resp.nodeID] = serverpb.ProblemRangesResponse_NodeProblems{
					ErrorMessage: resp.err.Error(),
				}
				continue
			} else {
				__antithesis_instrumentation__.Notify(195382)
			}
			__antithesis_instrumentation__.Notify(195378)
			var problems serverpb.ProblemRangesResponse_NodeProblems
			for _, info := range resp.resp.Ranges {
				__antithesis_instrumentation__.Notify(195383)
				if len(info.ErrorMessage) != 0 {
					__antithesis_instrumentation__.Notify(195393)
					response.ProblemsByNodeID[resp.nodeID] = serverpb.ProblemRangesResponse_NodeProblems{
						ErrorMessage: info.ErrorMessage,
					}
					continue
				} else {
					__antithesis_instrumentation__.Notify(195394)
				}
				__antithesis_instrumentation__.Notify(195384)
				if info.Problems.Unavailable {
					__antithesis_instrumentation__.Notify(195395)
					problems.UnavailableRangeIDs =
						append(problems.UnavailableRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195396)
				}
				__antithesis_instrumentation__.Notify(195385)
				if info.Problems.LeaderNotLeaseHolder {
					__antithesis_instrumentation__.Notify(195397)
					problems.RaftLeaderNotLeaseHolderRangeIDs =
						append(problems.RaftLeaderNotLeaseHolderRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195398)
				}
				__antithesis_instrumentation__.Notify(195386)
				if info.Problems.NoRaftLeader {
					__antithesis_instrumentation__.Notify(195399)
					problems.NoRaftLeaderRangeIDs =
						append(problems.NoRaftLeaderRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195400)
				}
				__antithesis_instrumentation__.Notify(195387)
				if info.Problems.Underreplicated {
					__antithesis_instrumentation__.Notify(195401)
					problems.UnderreplicatedRangeIDs =
						append(problems.UnderreplicatedRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195402)
				}
				__antithesis_instrumentation__.Notify(195388)
				if info.Problems.Overreplicated {
					__antithesis_instrumentation__.Notify(195403)
					problems.OverreplicatedRangeIDs =
						append(problems.OverreplicatedRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195404)
				}
				__antithesis_instrumentation__.Notify(195389)
				if info.Problems.NoLease {
					__antithesis_instrumentation__.Notify(195405)
					problems.NoLeaseRangeIDs =
						append(problems.NoLeaseRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195406)
				}
				__antithesis_instrumentation__.Notify(195390)
				if info.Problems.QuiescentEqualsTicking {
					__antithesis_instrumentation__.Notify(195407)
					problems.QuiescentEqualsTickingRangeIDs =
						append(problems.QuiescentEqualsTickingRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195408)
				}
				__antithesis_instrumentation__.Notify(195391)
				if info.Problems.RaftLogTooLarge {
					__antithesis_instrumentation__.Notify(195409)
					problems.RaftLogTooLargeRangeIDs =
						append(problems.RaftLogTooLargeRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195410)
				}
				__antithesis_instrumentation__.Notify(195392)
				if info.Problems.CircuitBreakerError {
					__antithesis_instrumentation__.Notify(195411)
					problems.CircuitBreakerErrorRangeIDs =
						append(problems.CircuitBreakerErrorRangeIDs, info.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(195412)
				}
			}
			__antithesis_instrumentation__.Notify(195379)
			sort.Sort(roachpb.RangeIDSlice(problems.UnavailableRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.RaftLeaderNotLeaseHolderRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.NoRaftLeaderRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.NoLeaseRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.UnderreplicatedRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.OverreplicatedRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.QuiescentEqualsTickingRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.RaftLogTooLargeRangeIDs))
			sort.Sort(roachpb.RangeIDSlice(problems.CircuitBreakerErrorRangeIDs))
			response.ProblemsByNodeID[resp.nodeID] = problems
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(195380)
			return nil, status.Errorf(codes.DeadlineExceeded, ctx.Err().Error())
		}
	}
	__antithesis_instrumentation__.Notify(195361)

	return response, nil
}
