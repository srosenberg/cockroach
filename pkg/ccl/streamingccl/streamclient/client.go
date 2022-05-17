package streamclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type SubscriptionToken []byte

type CheckpointToken []byte

type Client interface {
	Create(ctx context.Context, tenantID roachpb.TenantID) (streaming.StreamID, error)

	Heartbeat(ctx context.Context, streamID streaming.StreamID, consumed hlc.Timestamp) error

	Plan(ctx context.Context, streamID streaming.StreamID) (Topology, error)

	Subscribe(
		ctx context.Context,
		streamID streaming.StreamID,
		spec SubscriptionToken,
		checkpoint hlc.Timestamp,
	) (Subscription, error)

	Close() error
}

type Topology []PartitionInfo

type PartitionInfo struct {
	ID string
	SubscriptionToken
	SrcInstanceID int
	SrcAddr       streamingccl.PartitionAddress
	SrcLocality   roachpb.Locality
}

type Subscription interface {
	Subscribe(ctx context.Context) error

	Events() <-chan streamingccl.Event

	Err() error
}

func NewStreamClient(streamAddress streamingccl.StreamAddress) (Client, error) {
	__antithesis_instrumentation__.Notify(24953)
	var streamClient Client
	streamURL, err := streamAddress.URL()
	if err != nil {
		__antithesis_instrumentation__.Notify(24956)
		return streamClient, err
	} else {
		__antithesis_instrumentation__.Notify(24957)
	}
	__antithesis_instrumentation__.Notify(24954)

	switch streamURL.Scheme {
	case "postgres", "postgresql":
		__antithesis_instrumentation__.Notify(24958)

		return newPartitionedStreamClient(streamURL)
	case RandomGenScheme:
		__antithesis_instrumentation__.Notify(24959)
		streamClient, err = newRandomStreamClient(streamURL)
		if err != nil {
			__antithesis_instrumentation__.Notify(24961)
			return streamClient, err
		} else {
			__antithesis_instrumentation__.Notify(24962)
		}
	default:
		__antithesis_instrumentation__.Notify(24960)
		return nil, errors.Newf("stream replication from scheme %q is unsupported", streamURL.Scheme)
	}
	__antithesis_instrumentation__.Notify(24955)

	return streamClient, nil
}
