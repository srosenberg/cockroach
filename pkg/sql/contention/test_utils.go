package contention

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type fakeStatusServer struct {
	data          map[uuid.UUID]roachpb.TransactionFingerprintID
	injectedError error
}

func newFakeStatusServer() *fakeStatusServer {
	__antithesis_instrumentation__.Notify(459316)
	return &fakeStatusServer{
		data:          make(map[uuid.UUID]roachpb.TransactionFingerprintID),
		injectedError: nil,
	}
}

func (f *fakeStatusServer) txnIDResolution(
	_ context.Context, req *serverpb.TxnIDResolutionRequest,
) (*serverpb.TxnIDResolutionResponse, error) {
	__antithesis_instrumentation__.Notify(459317)
	if f.injectedError != nil {
		__antithesis_instrumentation__.Notify(459320)
		return nil, f.injectedError
	} else {
		__antithesis_instrumentation__.Notify(459321)
	}
	__antithesis_instrumentation__.Notify(459318)

	resp := &serverpb.TxnIDResolutionResponse{
		ResolvedTxnIDs: make([]contentionpb.ResolvedTxnID, 0),
	}

	for _, txnID := range req.TxnIDs {
		__antithesis_instrumentation__.Notify(459322)
		if txnFingerprintID, ok := f.data[txnID]; ok {
			__antithesis_instrumentation__.Notify(459323)
			resp.ResolvedTxnIDs = append(resp.ResolvedTxnIDs, contentionpb.ResolvedTxnID{
				TxnID:            txnID,
				TxnFingerprintID: txnFingerprintID,
			})
		} else {
			__antithesis_instrumentation__.Notify(459324)
		}
	}
	__antithesis_instrumentation__.Notify(459319)

	return resp, nil
}

func (f *fakeStatusServer) setTxnIDEntry(
	txnID uuid.UUID, txnFingerprintID roachpb.TransactionFingerprintID,
) {
	__antithesis_instrumentation__.Notify(459325)
	f.data[txnID] = txnFingerprintID
}

type fakeStatusServerCluster map[string]*fakeStatusServer

func newFakeStatusServerCluster() fakeStatusServerCluster {
	__antithesis_instrumentation__.Notify(459326)
	return make(fakeStatusServerCluster)
}

func (f fakeStatusServerCluster) getStatusServer(coordinatorID string) *fakeStatusServer {
	__antithesis_instrumentation__.Notify(459327)
	statusServer, ok := f[coordinatorID]
	if !ok {
		__antithesis_instrumentation__.Notify(459329)
		statusServer = newFakeStatusServer()
		f[coordinatorID] = statusServer
	} else {
		__antithesis_instrumentation__.Notify(459330)
	}
	__antithesis_instrumentation__.Notify(459328)
	return statusServer
}

func (f fakeStatusServerCluster) txnIDResolution(
	ctx context.Context, req *serverpb.TxnIDResolutionRequest,
) (*serverpb.TxnIDResolutionResponse, error) {
	__antithesis_instrumentation__.Notify(459331)
	return f.getStatusServer(req.CoordinatorID).txnIDResolution(ctx, req)
}

func (f fakeStatusServerCluster) setTxnIDEntry(
	coordinatorNodeID string, txnID uuid.UUID, txnFingerprintID roachpb.TransactionFingerprintID,
) {
	__antithesis_instrumentation__.Notify(459332)
	f.getStatusServer(coordinatorNodeID).setTxnIDEntry(txnID, txnFingerprintID)
}

func (f fakeStatusServerCluster) setStatusServerError(coordinatorNodeID string, err error) {
	__antithesis_instrumentation__.Notify(459333)
	f.getStatusServer(coordinatorNodeID).injectedError = err
}

func (f fakeStatusServerCluster) clear() {
	__antithesis_instrumentation__.Notify(459334)
	for k := range f {
		__antithesis_instrumentation__.Notify(459335)
		delete(f, k)
	}
}

func (f fakeStatusServerCluster) clearErrors() {
	__antithesis_instrumentation__.Notify(459336)
	for _, statusServer := range f {
		__antithesis_instrumentation__.Notify(459337)
		statusServer.injectedError = nil
	}
}
