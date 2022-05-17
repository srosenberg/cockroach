package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/gorilla/mux"
)

type nodeStatus struct {
	NodeID int32 `json:"node_id"`

	Address  util.UnresolvedAddr `json:"address"`
	Attrs    roachpb.Attributes  `json:"attrs"`
	Locality roachpb.Locality    `json:"locality"`

	ServerVersion roachpb.Version `json:"ServerVersion"`

	BuildTag string `json:"build_tag"`

	StartedAt int64 `json:"started_at"`

	ClusterName string `json:"cluster_name"`

	SQLAddress util.UnresolvedAddr `json:"sql_address"`

	Metrics map[string]float64 `json:"metrics,omitempty"`

	TotalSystemMemory int64 `json:"total_system_memory,omitempty"`

	NumCpus int32 `json:"num_cpus,omitempty"`

	UpdatedAt int64 `json:"updated_at,omitempty"`

	LivenessStatus int32 `json:"liveness_status"`
}

type nodesResponse struct {
	Nodes []nodeStatus `json:"nodes"`

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) listNodes(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189050)
	ctx := r.Context()
	limit, offset := getSimplePaginationValues(r)
	ctx = apiToOutgoingGatewayCtx(ctx, r)

	nodes, next, err := a.status.nodesHelper(ctx, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(189053)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189054)
	}
	__antithesis_instrumentation__.Notify(189051)
	var resp nodesResponse
	resp.Next = next
	for _, n := range nodes.Nodes {
		__antithesis_instrumentation__.Notify(189055)
		resp.Nodes = append(resp.Nodes, nodeStatus{
			NodeID:            int32(n.Desc.NodeID),
			Address:           n.Desc.Address,
			Attrs:             n.Desc.Attrs,
			Locality:          n.Desc.Locality,
			ServerVersion:     n.Desc.ServerVersion,
			BuildTag:          n.Desc.BuildTag,
			StartedAt:         n.Desc.StartedAt,
			ClusterName:       n.Desc.ClusterName,
			SQLAddress:        n.Desc.SQLAddress,
			Metrics:           n.Metrics,
			TotalSystemMemory: n.TotalSystemMemory,
			NumCpus:           n.NumCpus,
			UpdatedAt:         n.UpdatedAt,
			LivenessStatus:    int32(nodes.LivenessByNodeID[n.Desc.NodeID]),
		})
	}
	__antithesis_instrumentation__.Notify(189052)
	writeJSONResponse(ctx, w, 200, resp)
}

func parseRangeIDs(input string, w http.ResponseWriter) (ranges []roachpb.RangeID, ok bool) {
	__antithesis_instrumentation__.Notify(189056)
	if len(input) == 0 {
		__antithesis_instrumentation__.Notify(189059)
		return nil, true
	} else {
		__antithesis_instrumentation__.Notify(189060)
	}
	__antithesis_instrumentation__.Notify(189057)
	for _, reqRange := range strings.Split(input, ",") {
		__antithesis_instrumentation__.Notify(189061)
		rangeID, err := strconv.ParseInt(reqRange, 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(189063)
			http.Error(w, "invalid range ID", http.StatusBadRequest)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(189064)
		}
		__antithesis_instrumentation__.Notify(189062)

		ranges = append(ranges, roachpb.RangeID(rangeID))
	}
	__antithesis_instrumentation__.Notify(189058)
	return ranges, true
}

type nodeRangeResponse struct {
	RangeInfo rangeInfo `json:"range_info"`
	Error     string    `json:"error,omitempty"`
}

type rangeResponse struct {
	Responses map[string]nodeRangeResponse `json:"responses_by_node_id"`
}

func (a *apiV2Server) listRange(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189065)
	ctx := r.Context()
	ctx = apiToOutgoingGatewayCtx(ctx, r)
	vars := mux.Vars(r)
	rangeID, err := strconv.ParseInt(vars["range_id"], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(189072)
		http.Error(w, "invalid range ID", http.StatusBadRequest)
		return
	} else {
		__antithesis_instrumentation__.Notify(189073)
	}
	__antithesis_instrumentation__.Notify(189066)

	response := &rangeResponse{
		Responses: make(map[string]nodeRangeResponse),
	}

	rangesRequest := &serverpb.RangesRequest{
		RangeIDs: []roachpb.RangeID{roachpb.RangeID(rangeID)},
	}

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(189074)
		client, err := a.status.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(189067)
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(189075)
		status := client.(serverpb.StatusClient)
		return status.Ranges(ctx, rangesRequest)
	}
	__antithesis_instrumentation__.Notify(189068)
	responseFn := func(nodeID roachpb.NodeID, resp interface{}) {
		__antithesis_instrumentation__.Notify(189076)
		rangesResp := resp.(*serverpb.RangesResponse)

		if len(rangesResp.Ranges) == 0 {
			__antithesis_instrumentation__.Notify(189078)
			return
		} else {
			__antithesis_instrumentation__.Notify(189079)
		}
		__antithesis_instrumentation__.Notify(189077)
		var ri rangeInfo
		ri.init(rangesResp.Ranges[0])
		response.Responses[nodeID.String()] = nodeRangeResponse{RangeInfo: ri}
	}
	__antithesis_instrumentation__.Notify(189069)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(189080)
		response.Responses[nodeID.String()] = nodeRangeResponse{
			Error: err.Error(),
		}
	}
	__antithesis_instrumentation__.Notify(189070)

	if err := a.status.iterateNodes(
		ctx, fmt.Sprintf("details about range %d", rangeID), dialFn, nodeFn, responseFn, errorFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(189081)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189082)
	}
	__antithesis_instrumentation__.Notify(189071)
	writeJSONResponse(ctx, w, 200, response)
}

type rangeDescriptorInfo struct {
	RangeID int64 `json:"range_id"`

	StartKey []byte `json:"start_key,omitempty"`

	EndKey []byte `json:"end_key,omitempty"`

	StoreID int32 `json:"store_id"`

	QueriesPerSecond float64 `json:"queries_per_second"`
}

func (r *rangeDescriptorInfo) init(rd *roachpb.RangeDescriptor) {
	if rd == nil {
		*r = rangeDescriptorInfo{}
		return
	}
	*r = rangeDescriptorInfo{
		RangeID:  int64(rd.RangeID),
		StartKey: rd.StartKey,
		EndKey:   rd.EndKey,
	}
}

type rangeInfo struct {
	Desc rangeDescriptorInfo `json:"desc"`

	Span serverpb.PrettySpan `json:"span"`

	SourceNodeID int32 `json:"source_node_id,omitempty"`

	SourceStoreID int32 `json:"source_store_id,omitempty"`

	ErrorMessage string `json:"error_message,omitempty"`

	LeaseHistory []roachpb.Lease `json:"lease_history"`

	Problems serverpb.RangeProblems `json:"problems"`

	Stats serverpb.RangeStatistics `json:"stats"`

	Quiescent bool `json:"quiescent,omitempty"`

	Ticking bool `json:"ticking,omitempty"`
}

func (ri *rangeInfo) init(r serverpb.RangeInfo) {
	*ri = rangeInfo{
		Span:          r.Span,
		SourceNodeID:  int32(r.SourceNodeID),
		SourceStoreID: int32(r.SourceStoreID),
		ErrorMessage:  r.ErrorMessage,
		LeaseHistory:  r.LeaseHistory,
		Problems:      r.Problems,
		Stats:         r.Stats,
		Quiescent:     r.Quiescent,
		Ticking:       r.Ticking,
	}
	ri.Desc.init(r.State.Desc)
}

type nodeRangesResponse struct {
	Ranges []rangeInfo `json:"ranges"`

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) listNodeRanges(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189083)
	ctx := r.Context()
	ctx = apiToOutgoingGatewayCtx(ctx, r)
	vars := mux.Vars(r)
	nodeIDStr := vars["node_id"]
	if nodeIDStr != "local" {
		__antithesis_instrumentation__.Notify(189088)
		nodeID, err := strconv.ParseInt(nodeIDStr, 10, 32)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(189089)
			return nodeID <= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(189090)
			http.Error(w, "invalid node ID", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(189091)
		}
	} else {
		__antithesis_instrumentation__.Notify(189092)
	}
	__antithesis_instrumentation__.Notify(189084)

	ranges, ok := parseRangeIDs(r.URL.Query().Get("ranges"), w)
	if !ok {
		__antithesis_instrumentation__.Notify(189093)
		return
	} else {
		__antithesis_instrumentation__.Notify(189094)
	}
	__antithesis_instrumentation__.Notify(189085)
	req := &serverpb.RangesRequest{
		NodeId:   nodeIDStr,
		RangeIDs: ranges,
	}
	limit, offset := getSimplePaginationValues(r)
	statusResp, next, err := a.status.rangesHelper(ctx, req, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(189095)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189096)
	}
	__antithesis_instrumentation__.Notify(189086)
	resp := nodeRangesResponse{
		Ranges: make([]rangeInfo, 0, len(statusResp.Ranges)),
		Next:   next,
	}
	for _, r := range statusResp.Ranges {
		__antithesis_instrumentation__.Notify(189097)
		var ri rangeInfo
		ri.init(r)
		resp.Ranges = append(resp.Ranges, ri)
	}
	__antithesis_instrumentation__.Notify(189087)
	writeJSONResponse(ctx, w, 200, resp)
}

type responseError struct {
	ErrorMessage string         `json:"error_message"`
	NodeID       roachpb.NodeID `json:"node_id,omitempty"`
}

type hotRangesResponse struct {
	Ranges []hotRangeInfo  `json:"ranges"`
	Errors []responseError `json:"response_error,omitempty"`

	Next string `json:"next,omitempty"`
}

type hotRangeInfo struct {
	RangeID           roachpb.RangeID  `json:"range_id"`
	NodeID            roachpb.NodeID   `json:"node_id"`
	QPS               float64          `json:"qps"`
	LeaseholderNodeID roachpb.NodeID   `json:"leaseholder_node_id"`
	TableName         string           `json:"table_name"`
	DatabaseName      string           `json:"database_name"`
	IndexName         string           `json:"index_name"`
	SchemaName        string           `json:"schema_name"`
	ReplicaNodeIDs    []roachpb.NodeID `json:"replica_node_ids"`
	StoreID           roachpb.StoreID  `json:"store_id"`
}

func (a *apiV2Server) listHotRanges(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189098)
	ctx := r.Context()
	ctx = apiToOutgoingGatewayCtx(ctx, r)
	nodeIDStr := r.URL.Query().Get("node_id")
	limit, start := getRPCPaginationValues(r)

	response := &hotRangesResponse{}
	var requestedNodes []roachpb.NodeID
	if len(nodeIDStr) > 0 {
		__antithesis_instrumentation__.Notify(189106)
		requestedNodeID, _, err := a.status.parseNodeID(nodeIDStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(189108)
			http.Error(w, "invalid node ID", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(189109)
		}
		__antithesis_instrumentation__.Notify(189107)
		requestedNodes = []roachpb.NodeID{requestedNodeID}
	} else {
		__antithesis_instrumentation__.Notify(189110)
	}
	__antithesis_instrumentation__.Notify(189099)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(189111)
		client, err := a.status.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(189100)
	remoteRequest := serverpb.HotRangesRequest{NodeID: "local"}
	nodeFn := func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(189112)
		status := client.(serverpb.StatusClient)
		resp, err := status.HotRangesV2(ctx, &remoteRequest)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(189115)
			return resp == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(189116)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(189117)
		}
		__antithesis_instrumentation__.Notify(189113)

		var hotRangeInfos = make([]hotRangeInfo, len(resp.Ranges))
		for i, r := range resp.Ranges {
			__antithesis_instrumentation__.Notify(189118)
			hotRangeInfos[i] = hotRangeInfo{
				RangeID:           r.RangeID,
				NodeID:            r.NodeID,
				QPS:               r.QPS,
				LeaseholderNodeID: r.LeaseholderNodeID,
				TableName:         r.TableName,
				DatabaseName:      r.DatabaseName,
				IndexName:         r.IndexName,
				ReplicaNodeIDs:    r.ReplicaNodeIds,
				SchemaName:        r.SchemaName,
				StoreID:           r.StoreID,
			}
		}
		__antithesis_instrumentation__.Notify(189114)
		return hotRangeInfos, nil
	}
	__antithesis_instrumentation__.Notify(189101)
	responseFn := func(nodeID roachpb.NodeID, resp interface{}) {
		__antithesis_instrumentation__.Notify(189119)
		response.Ranges = append(response.Ranges, resp.([]hotRangeInfo)...)
	}
	__antithesis_instrumentation__.Notify(189102)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(189120)
		response.Errors = append(response.Errors, responseError{
			ErrorMessage: err.Error(),
			NodeID:       nodeID,
		})
	}
	__antithesis_instrumentation__.Notify(189103)

	next, err := a.status.paginatedIterateNodes(
		ctx, "hot ranges", limit, start, requestedNodes, dialFn,
		nodeFn, responseFn, errorFn)

	if err != nil {
		__antithesis_instrumentation__.Notify(189121)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189122)
	}
	__antithesis_instrumentation__.Notify(189104)
	var nextBytes []byte
	if nextBytes, err = next.MarshalText(); err != nil {
		__antithesis_instrumentation__.Notify(189123)
		response.Errors = append(response.Errors, responseError{ErrorMessage: err.Error()})
	} else {
		__antithesis_instrumentation__.Notify(189124)
		response.Next = string(nextBytes)
	}
	__antithesis_instrumentation__.Notify(189105)
	writeJSONResponse(ctx, w, 200, response)
}
