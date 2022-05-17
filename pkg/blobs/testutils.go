package blobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func TestBlobServiceClient(externalIODir string) BlobClientFactory {
	__antithesis_instrumentation__.Notify(3420)
	return func(ctx context.Context, dialing roachpb.NodeID) (BlobClient, error) {
		__antithesis_instrumentation__.Notify(3421)
		return NewLocalClient(externalIODir)
	}
}

var TestEmptyBlobClientFactory = func(
	ctx context.Context, dialing roachpb.NodeID,
) (BlobClient, error) {
	__antithesis_instrumentation__.Notify(3422)
	return nil, nil
}
