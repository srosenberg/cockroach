package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

func simplePaginate(input interface{}, limit, offset int) (result interface{}, next int) {
	__antithesis_instrumentation__.Notify(194919)
	val := reflect.ValueOf(input)
	if limit <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(194924)
		return val.Kind() != reflect.Slice == true
	}() == true {
		__antithesis_instrumentation__.Notify(194925)
		return input, 0
	} else {
		__antithesis_instrumentation__.Notify(194926)
		if offset < 0 {
			__antithesis_instrumentation__.Notify(194927)
			offset = 0
		} else {
			__antithesis_instrumentation__.Notify(194928)
		}
	}
	__antithesis_instrumentation__.Notify(194920)
	startIdx := offset
	endIdx := offset + limit
	if startIdx > val.Len() {
		__antithesis_instrumentation__.Notify(194929)
		startIdx = val.Len()
	} else {
		__antithesis_instrumentation__.Notify(194930)
	}
	__antithesis_instrumentation__.Notify(194921)
	if endIdx > val.Len() {
		__antithesis_instrumentation__.Notify(194931)
		endIdx = val.Len()
	} else {
		__antithesis_instrumentation__.Notify(194932)
	}
	__antithesis_instrumentation__.Notify(194922)
	next = endIdx
	if endIdx == val.Len() {
		__antithesis_instrumentation__.Notify(194933)
		next = 0
	} else {
		__antithesis_instrumentation__.Notify(194934)
	}
	__antithesis_instrumentation__.Notify(194923)
	return val.Slice(startIdx, endIdx).Interface(), next
}

type paginationState struct {
	nodesQueried    []roachpb.NodeID
	inProgress      roachpb.NodeID
	inProgressIndex int
	nodesToQuery    []roachpb.NodeID
}

func (p *paginationState) mergeNodeIDs(allNodeIDs []roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(194935)
	sortedNodeIDs := make([]roachpb.NodeID, 0, len(p.nodesQueried)+1+len(p.nodesToQuery))
	sortedNodeIDs = append(sortedNodeIDs, p.nodesQueried...)
	if p.inProgress != 0 {
		__antithesis_instrumentation__.Notify(194939)
		sortedNodeIDs = append(sortedNodeIDs, p.inProgress)
	} else {
		__antithesis_instrumentation__.Notify(194940)
	}
	__antithesis_instrumentation__.Notify(194936)
	sortedNodeIDs = append(sortedNodeIDs, p.nodesToQuery...)
	sort.Slice(sortedNodeIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(194941)
		return sortedNodeIDs[i] < sortedNodeIDs[j]
	})
	__antithesis_instrumentation__.Notify(194937)

	j := 0
	for i := range allNodeIDs {
		__antithesis_instrumentation__.Notify(194942)

		for j < len(sortedNodeIDs) && func() bool {
			__antithesis_instrumentation__.Notify(194944)
			return sortedNodeIDs[j] < allNodeIDs[i] == true
		}() == true {
			__antithesis_instrumentation__.Notify(194945)
			j++
		}
		__antithesis_instrumentation__.Notify(194943)

		if j >= len(sortedNodeIDs) || func() bool {
			__antithesis_instrumentation__.Notify(194946)
			return sortedNodeIDs[j] != allNodeIDs[i] == true
		}() == true {
			__antithesis_instrumentation__.Notify(194947)
			p.nodesToQuery = append(p.nodesToQuery, allNodeIDs[i])
		} else {
			__antithesis_instrumentation__.Notify(194948)
		}
	}
	__antithesis_instrumentation__.Notify(194938)
	if p.inProgress == 0 && func() bool {
		__antithesis_instrumentation__.Notify(194949)
		return len(p.nodesToQuery) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(194950)
		p.inProgress = p.nodesToQuery[0]
		p.inProgressIndex = 0
		p.nodesToQuery = p.nodesToQuery[1:]
	} else {
		__antithesis_instrumentation__.Notify(194951)
	}
}

func (p *paginationState) paginate(
	limit int, nodeID roachpb.NodeID, length int,
) (start, end, newLimit int, err error) {
	__antithesis_instrumentation__.Notify(194952)
	if limit <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(194957)
		return nodeID == 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(194958)
		return int(p.inProgress) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(194959)

		return 0, 0, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(194960)
	}
	__antithesis_instrumentation__.Notify(194953)
	if p.inProgress != nodeID {
		__antithesis_instrumentation__.Notify(194961)
		p.nodesQueried = append(p.nodesQueried, p.inProgress)
		p.inProgress = 0
		p.inProgressIndex = 0
		for i := range p.nodesToQuery {
			__antithesis_instrumentation__.Notify(194963)
			if p.nodesToQuery[i] == nodeID {
				__antithesis_instrumentation__.Notify(194964)

				p.inProgress = nodeID
				p.nodesQueried = append(p.nodesQueried, p.nodesToQuery[0:i]...)
				p.nodesToQuery = p.nodesToQuery[i+1:]
				break
			} else {
				__antithesis_instrumentation__.Notify(194965)
			}
		}
		__antithesis_instrumentation__.Notify(194962)
		if p.inProgress == 0 {
			__antithesis_instrumentation__.Notify(194966)

			return 0, 0, 0, errors.Errorf("could not find node %d in pagination state %v", nodeID, p)
		} else {
			__antithesis_instrumentation__.Notify(194967)
		}
	} else {
		__antithesis_instrumentation__.Notify(194968)
	}
	__antithesis_instrumentation__.Notify(194954)
	doneWithNode := false
	if length > 0 {
		__antithesis_instrumentation__.Notify(194969)
		start = p.inProgressIndex
		if start > length {
			__antithesis_instrumentation__.Notify(194972)
			start = length
		} else {
			__antithesis_instrumentation__.Notify(194973)
		}
		__antithesis_instrumentation__.Notify(194970)

		if start+limit >= length {
			__antithesis_instrumentation__.Notify(194974)
			end = length
			doneWithNode = true
		} else {
			__antithesis_instrumentation__.Notify(194975)
			end = start + limit
		}
		__antithesis_instrumentation__.Notify(194971)
		limit -= end - start
		p.inProgressIndex = end
	} else {
		__antithesis_instrumentation__.Notify(194976)
	}
	__antithesis_instrumentation__.Notify(194955)
	if doneWithNode {
		__antithesis_instrumentation__.Notify(194977)
		p.nodesQueried = append(p.nodesQueried, nodeID)
		p.inProgressIndex = 0
		if len(p.nodesToQuery) > 0 {
			__antithesis_instrumentation__.Notify(194978)
			p.inProgress = p.nodesToQuery[0]
			p.nodesToQuery = p.nodesToQuery[1:]
		} else {
			__antithesis_instrumentation__.Notify(194979)
			p.nodesToQuery = p.nodesToQuery[:0]
			p.inProgress = 0
		}
	} else {
		__antithesis_instrumentation__.Notify(194980)
	}
	__antithesis_instrumentation__.Notify(194956)
	return start, end, limit, nil
}

func (p *paginationState) UnmarshalText(text []byte) error {
	__antithesis_instrumentation__.Notify(194981)
	decoder := base64.NewDecoder(base64.URLEncoding, bytes.NewReader(text))
	var decodedText []byte
	var err error
	if decodedText, err = ioutil.ReadAll(decoder); err != nil {
		__antithesis_instrumentation__.Notify(194990)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194991)
	}
	__antithesis_instrumentation__.Notify(194982)
	parts := strings.Split(string(decodedText), "|")
	if len(parts) != 4 {
		__antithesis_instrumentation__.Notify(194992)
		return errors.New("invalid pagination state")
	} else {
		__antithesis_instrumentation__.Notify(194993)
	}
	__antithesis_instrumentation__.Notify(194983)
	parseNodeIDSlice := func(str string) ([]roachpb.NodeID, error) {
		__antithesis_instrumentation__.Notify(194994)
		nodeIDs := strings.Split(str, ",")
		res := make([]roachpb.NodeID, 0, len(nodeIDs))
		for _, part := range nodeIDs {
			__antithesis_instrumentation__.Notify(194996)

			part = strings.TrimSpace(part)
			if len(part) == 0 {
				__antithesis_instrumentation__.Notify(195000)
				continue
			} else {
				__antithesis_instrumentation__.Notify(195001)
			}
			__antithesis_instrumentation__.Notify(194997)
			val, err := strconv.ParseUint(part, 10, 32)
			if err != nil {
				__antithesis_instrumentation__.Notify(195002)
				return nil, errors.Wrap(err, "invalid pagination state")
			} else {
				__antithesis_instrumentation__.Notify(195003)
			}
			__antithesis_instrumentation__.Notify(194998)
			if val <= 0 {
				__antithesis_instrumentation__.Notify(195004)
				return nil, errors.New("expected positive nodeID in pagination token")
			} else {
				__antithesis_instrumentation__.Notify(195005)
			}
			__antithesis_instrumentation__.Notify(194999)
			res = append(res, roachpb.NodeID(val))
		}
		__antithesis_instrumentation__.Notify(194995)
		return res, nil
	}
	__antithesis_instrumentation__.Notify(194984)
	p.nodesQueried, err = parseNodeIDSlice(parts[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(195006)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195007)
	}
	__antithesis_instrumentation__.Notify(194985)
	var inProgressInt int
	inProgressInt, err = strconv.Atoi(parts[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(195008)
		return errors.Wrap(err, "invalid pagination state")
	} else {
		__antithesis_instrumentation__.Notify(195009)
	}
	__antithesis_instrumentation__.Notify(194986)
	p.inProgress = roachpb.NodeID(inProgressInt)
	p.inProgressIndex, err = strconv.Atoi(parts[2])
	if err != nil {
		__antithesis_instrumentation__.Notify(195010)
		return errors.Wrap(err, "invalid pagination state")
	} else {
		__antithesis_instrumentation__.Notify(195011)
	}
	__antithesis_instrumentation__.Notify(194987)
	if p.inProgressIndex < 0 || func() bool {
		__antithesis_instrumentation__.Notify(195012)
		return (p.inProgressIndex > 0 && func() bool {
			__antithesis_instrumentation__.Notify(195013)
			return p.inProgress <= 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(195014)
		return errors.Newf("invalid pagination resumption token: (%d, %d)", p.inProgress, p.inProgressIndex)
	} else {
		__antithesis_instrumentation__.Notify(195015)
	}
	__antithesis_instrumentation__.Notify(194988)
	p.nodesToQuery, err = parseNodeIDSlice(parts[3])
	if err != nil {
		__antithesis_instrumentation__.Notify(195016)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195017)
	}
	__antithesis_instrumentation__.Notify(194989)
	return nil
}

func (p *paginationState) MarshalText() (text []byte, err error) {
	__antithesis_instrumentation__.Notify(195018)
	var builder, builder2 bytes.Buffer
	for _, nid := range p.nodesQueried {
		__antithesis_instrumentation__.Notify(195023)
		fmt.Fprintf(&builder, "%d,", nid)
	}
	__antithesis_instrumentation__.Notify(195019)
	fmt.Fprintf(&builder, "|%d|%d|", p.inProgress, p.inProgressIndex)
	for _, nid := range p.nodesToQuery {
		__antithesis_instrumentation__.Notify(195024)
		fmt.Fprintf(&builder, "%d,", nid)
	}
	__antithesis_instrumentation__.Notify(195020)
	encoder := base64.NewEncoder(base64.URLEncoding, &builder2)
	if _, err = encoder.Write(builder.Bytes()); err != nil {
		__antithesis_instrumentation__.Notify(195025)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195026)
	}
	__antithesis_instrumentation__.Notify(195021)
	if err = encoder.Close(); err != nil {
		__antithesis_instrumentation__.Notify(195027)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195028)
	}
	__antithesis_instrumentation__.Notify(195022)
	return builder2.Bytes(), nil
}

type paginatedNodeResponse struct {
	nodeID   roachpb.NodeID
	response interface{}
	value    reflect.Value
	len      int
	err      error
}

type rpcNodePaginator struct {
	limit        int
	numNodes     int
	errorCtx     string
	pagState     paginationState
	responseChan chan paginatedNodeResponse
	nodeStatuses map[roachpb.NodeID]nodeStatusWithLiveness

	dialFn     func(ctx context.Context, id roachpb.NodeID) (client interface{}, err error)
	nodeFn     func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (res interface{}, err error)
	responseFn func(nodeID roachpb.NodeID, resp interface{})
	errorFn    func(nodeID roachpb.NodeID, nodeFnError error)

	mu struct {
		syncutil.Mutex

		turnCond sync.Cond

		currentIdx, currentLen int
	}

	done int32
}

func (r *rpcNodePaginator) init() {
	r.mu.turnCond.L = &r.mu
	r.responseChan = make(chan paginatedNodeResponse, r.numNodes)
}

func (r *rpcNodePaginator) queryNode(ctx context.Context, nodeID roachpb.NodeID, idx int) {
	__antithesis_instrumentation__.Notify(195029)
	if atomic.LoadInt32(&r.done) != 0 {
		__antithesis_instrumentation__.Notify(195035)

		return
	} else {
		__antithesis_instrumentation__.Notify(195036)
	}
	__antithesis_instrumentation__.Notify(195030)
	var client interface{}
	addNodeResp := func(resp paginatedNodeResponse) {
		__antithesis_instrumentation__.Notify(195037)
		r.mu.Lock()
		defer r.mu.Unlock()

		for r.mu.currentIdx < idx && func() bool {
			__antithesis_instrumentation__.Notify(195042)
			return atomic.LoadInt32(&r.done) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(195043)
			r.mu.turnCond.Wait()
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(195044)
				r.mu.turnCond.Broadcast()
				return
			default:
				__antithesis_instrumentation__.Notify(195045)
			}
		}
		__antithesis_instrumentation__.Notify(195038)
		if atomic.LoadInt32(&r.done) != 0 {
			__antithesis_instrumentation__.Notify(195046)

			r.mu.turnCond.Broadcast()
			return
		} else {
			__antithesis_instrumentation__.Notify(195047)
		}
		__antithesis_instrumentation__.Notify(195039)
		r.responseChan <- resp
		r.mu.currentLen += resp.len
		if nodeID == r.pagState.inProgress {
			__antithesis_instrumentation__.Notify(195048)

			if resp.len > r.pagState.inProgressIndex {
				__antithesis_instrumentation__.Notify(195049)
				r.mu.currentLen -= r.pagState.inProgressIndex
			} else {
				__antithesis_instrumentation__.Notify(195050)
				r.mu.currentLen -= resp.len
			}
		} else {
			__antithesis_instrumentation__.Notify(195051)
		}
		__antithesis_instrumentation__.Notify(195040)
		if r.mu.currentLen >= r.limit {
			__antithesis_instrumentation__.Notify(195052)
			atomic.StoreInt32(&r.done, 1)
			close(r.responseChan)
		} else {
			__antithesis_instrumentation__.Notify(195053)
		}
		__antithesis_instrumentation__.Notify(195041)
		r.mu.currentIdx++
		r.mu.turnCond.Broadcast()
	}
	__antithesis_instrumentation__.Notify(195031)
	if err := contextutil.RunWithTimeout(ctx, "dial node", base.NetworkTimeout, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(195054)
		var err error
		client, err = r.dialFn(ctx, nodeID)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(195055)
		err = errors.Wrapf(err, "failed to dial into node %d (%s)",
			nodeID, r.nodeStatuses[nodeID].livenessStatus)
		addNodeResp(paginatedNodeResponse{nodeID: nodeID, err: err})
		return
	} else {
		__antithesis_instrumentation__.Notify(195056)
	}
	__antithesis_instrumentation__.Notify(195032)

	res, err := r.nodeFn(ctx, client, nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(195057)
		err = errors.Wrapf(err, "error requesting %s from node %d (%s)",
			r.errorCtx, nodeID, r.nodeStatuses[nodeID].livenessStatus)
	} else {
		__antithesis_instrumentation__.Notify(195058)
	}
	__antithesis_instrumentation__.Notify(195033)
	length := 0
	value := reflect.ValueOf(res)
	if res != nil && func() bool {
		__antithesis_instrumentation__.Notify(195059)
		return !value.IsNil() == true
	}() == true {
		__antithesis_instrumentation__.Notify(195060)
		length = 1
		if value.Kind() == reflect.Slice {
			__antithesis_instrumentation__.Notify(195061)
			length = value.Len()
		} else {
			__antithesis_instrumentation__.Notify(195062)
		}
	} else {
		__antithesis_instrumentation__.Notify(195063)
	}
	__antithesis_instrumentation__.Notify(195034)
	addNodeResp(paginatedNodeResponse{nodeID: nodeID, response: res, len: length, value: value, err: err})
}

func (r *rpcNodePaginator) processResponses(ctx context.Context) (next paginationState, err error) {
	__antithesis_instrumentation__.Notify(195064)

	next = r.pagState
	limit := r.limit
	numNodes := r.numNodes
	for numNodes > 0 {
		__antithesis_instrumentation__.Notify(195066)
		select {
		case res, ok := <-r.responseChan:
			__antithesis_instrumentation__.Notify(195068)
			if res.err != nil {
				__antithesis_instrumentation__.Notify(195071)
				r.errorFn(res.nodeID, res.err)
			} else {
				__antithesis_instrumentation__.Notify(195072)
				start, end, newLimit, err2 := next.paginate(limit, res.nodeID, res.len)
				if err2 != nil {
					__antithesis_instrumentation__.Notify(195075)
					r.errorFn(res.nodeID, err2)

					break
				} else {
					__antithesis_instrumentation__.Notify(195076)
				}
				__antithesis_instrumentation__.Notify(195073)
				var response interface{}
				if res.value.Kind() == reflect.Slice {
					__antithesis_instrumentation__.Notify(195077)
					response = res.value.Slice(start, end).Interface()
				} else {
					__antithesis_instrumentation__.Notify(195078)
					if end > start {
						__antithesis_instrumentation__.Notify(195079)

						response = res.value.Interface()
					} else {
						__antithesis_instrumentation__.Notify(195080)
					}
				}
				__antithesis_instrumentation__.Notify(195074)
				r.responseFn(res.nodeID, response)
				limit = newLimit
			}
			__antithesis_instrumentation__.Notify(195069)
			if !ok {
				__antithesis_instrumentation__.Notify(195081)
				return next, err
			} else {
				__antithesis_instrumentation__.Notify(195082)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(195070)
			err = errors.Errorf("request of %s canceled before completion", r.errorCtx)
			return next, err
		}
		__antithesis_instrumentation__.Notify(195067)
		numNodes--
	}
	__antithesis_instrumentation__.Notify(195065)
	return next, err
}

func getRPCPaginationValues(r *http.Request) (limit int, start paginationState) {
	__antithesis_instrumentation__.Notify(195083)
	var err error
	if limit, err = strconv.Atoi(r.URL.Query().Get("limit")); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(195086)
		return limit <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(195087)
		return 0, paginationState{}
	} else {
		__antithesis_instrumentation__.Notify(195088)
	}
	__antithesis_instrumentation__.Notify(195084)
	if err = start.UnmarshalText([]byte(r.URL.Query().Get("start"))); err != nil {
		__antithesis_instrumentation__.Notify(195089)
		return limit, paginationState{}
	} else {
		__antithesis_instrumentation__.Notify(195090)
	}
	__antithesis_instrumentation__.Notify(195085)
	return limit, start
}

func getSimplePaginationValues(r *http.Request) (limit, offset int) {
	__antithesis_instrumentation__.Notify(195091)
	var err error
	if limit, err = strconv.Atoi(r.URL.Query().Get("limit")); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(195094)
		return limit <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(195095)
		return 0, 0
	} else {
		__antithesis_instrumentation__.Notify(195096)
	}
	__antithesis_instrumentation__.Notify(195092)
	if offset, err = strconv.Atoi(r.URL.Query().Get("offset")); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(195097)
		return offset < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(195098)
		return limit, 0
	} else {
		__antithesis_instrumentation__.Notify(195099)
	}
	__antithesis_instrumentation__.Notify(195093)
	return limit, offset
}
