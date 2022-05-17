package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.Barrier, declareKeysBarrier, Barrier)
}

func declareKeysBarrier(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(96432)

	latchSpans.AddNonMVCC(spanset.SpanReadWrite, req.Header().Span())
}

func Barrier(
	_ context.Context, _ storage.ReadWriter, cArgs CommandArgs, response roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96433)
	resp := response.(*roachpb.BarrierResponse)
	resp.Timestamp = cArgs.EvalCtx.Clock().Now()

	return result.Result{}, nil
}
