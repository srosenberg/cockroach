package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func declareKeysProbe(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	_, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97067)

}

func init() {
	RegisterReadWriteCommand(roachpb.Probe, declareKeysProbe, Probe)
}

func Probe(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97068)
	return result.Result{
		Replicated: kvserverpb.ReplicatedEvalResult{
			IsProbe: true,
		},
	}, nil
}
