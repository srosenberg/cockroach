package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadOnlyCommand(roachpb.LeaseInfo, declareKeysLeaseInfo, LeaseInfo)
}

func declareKeysLeaseInfo(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97002)
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeLeaseKey(rs.GetRangeID())})
}

func LeaseInfo(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97003)
	reply := resp.(*roachpb.LeaseInfoResponse)
	lease, nextLease := cArgs.EvalCtx.GetLease()
	if nextLease != (roachpb.Lease{}) {
		__antithesis_instrumentation__.Notify(97005)

		reply.Lease = nextLease
		reply.CurrentLease = &lease
	} else {
		__antithesis_instrumentation__.Notify(97006)
		reply.Lease = lease
	}
	__antithesis_instrumentation__.Notify(97004)
	reply.EvaluatedBy = cArgs.EvalCtx.StoreID()
	return result.Result{}, nil
}
