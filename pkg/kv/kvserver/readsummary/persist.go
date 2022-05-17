package readsummary

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func Load(
	ctx context.Context, reader storage.Reader, rangeID roachpb.RangeID,
) (*rspb.ReadSummary, error) {
	__antithesis_instrumentation__.Notify(114263)
	var sum rspb.ReadSummary
	key := keys.RangePriorReadSummaryKey(rangeID)
	found, err := storage.MVCCGetProto(ctx, reader, key, hlc.Timestamp{}, &sum, storage.MVCCGetOptions{})
	if !found {
		__antithesis_instrumentation__.Notify(114265)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(114266)
	}
	__antithesis_instrumentation__.Notify(114264)
	return &sum, err
}

func Set(
	ctx context.Context,
	readWriter storage.ReadWriter,
	rangeID roachpb.RangeID,
	ms *enginepb.MVCCStats,
	sum *rspb.ReadSummary,
) error {
	__antithesis_instrumentation__.Notify(114267)
	key := keys.RangePriorReadSummaryKey(rangeID)
	return storage.MVCCPutProto(ctx, readWriter, ms, key, hlc.Timestamp{}, nil, sum)
}
