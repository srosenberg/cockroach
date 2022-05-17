package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

var todoSpanSet = &spanset.SpanSet{}

func NewReplicaEvalContext(r *Replica, ss *spanset.SpanSet) batcheval.EvalContext {
	__antithesis_instrumentation__.Notify(117185)
	if ss == nil {
		__antithesis_instrumentation__.Notify(117188)
		log.Fatalf(r.AnnotateCtx(context.Background()), "can't create a ReplicaEvalContext with assertions but no SpanSet")
	} else {
		__antithesis_instrumentation__.Notify(117189)
	}
	__antithesis_instrumentation__.Notify(117186)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(117190)
		return &SpanSetReplicaEvalContext{
			i:  r,
			ss: *ss,
		}
	} else {
		__antithesis_instrumentation__.Notify(117191)
	}
	__antithesis_instrumentation__.Notify(117187)
	return r
}
