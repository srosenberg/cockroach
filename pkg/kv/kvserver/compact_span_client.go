package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/errors"
)

type CompactEngineSpanClient struct {
	nd *nodedialer.Dialer
}

func NewCompactEngineSpanClient(nd *nodedialer.Dialer) *CompactEngineSpanClient {
	__antithesis_instrumentation__.Notify(98941)
	return &CompactEngineSpanClient{nd: nd}
}

func (c *CompactEngineSpanClient) CompactEngineSpan(
	ctx context.Context, nodeID, storeID int32, startKey, endKey []byte,
) error {
	__antithesis_instrumentation__.Notify(98942)
	conn, err := c.nd.Dial(ctx, roachpb.NodeID(nodeID), rpc.DefaultClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(98944)
		return errors.Wrapf(err, "could not dial node ID %d", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(98945)
	}
	__antithesis_instrumentation__.Notify(98943)
	client := NewPerStoreClient(conn)
	req := &CompactEngineSpanRequest{
		StoreRequestHeader: StoreRequestHeader{
			NodeID:  roachpb.NodeID(nodeID),
			StoreID: roachpb.StoreID(storeID),
		},
		Span: roachpb.Span{Key: roachpb.Key(startKey), EndKey: roachpb.Key(endKey)},
	}
	_, err = client.CompactEngineSpan(ctx, req)
	return err
}
