package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type replicaDecoder struct {
	r      *Replica
	cmdBuf replicatedCmdBuf
}

func (r *Replica) getDecoder() *replicaDecoder {
	__antithesis_instrumentation__.Notify(115049)
	d := &r.raftMu.decoder
	d.r = r
	return d
}

func (d *replicaDecoder) DecodeAndBind(ctx context.Context, ents []raftpb.Entry) (bool, error) {
	__antithesis_instrumentation__.Notify(115050)
	if err := d.decode(ctx, ents); err != nil {
		__antithesis_instrumentation__.Notify(115052)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(115053)
	}
	__antithesis_instrumentation__.Notify(115051)
	anyLocal := d.retrieveLocalProposals(ctx)
	d.createTracingSpans(ctx)
	return anyLocal, nil
}

func (d *replicaDecoder) decode(ctx context.Context, ents []raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(115054)
	for i := range ents {
		__antithesis_instrumentation__.Notify(115056)
		ent := &ents[i]
		if err := d.cmdBuf.allocate().decode(ctx, ent); err != nil {
			__antithesis_instrumentation__.Notify(115057)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115058)
		}
	}
	__antithesis_instrumentation__.Notify(115055)
	return nil
}

func (d *replicaDecoder) retrieveLocalProposals(ctx context.Context) (anyLocal bool) {
	__antithesis_instrumentation__.Notify(115059)
	d.r.mu.Lock()
	defer d.r.mu.Unlock()

	var it replicatedCmdBufSlice
	for it.init(&d.cmdBuf); it.Valid(); it.Next() {
		__antithesis_instrumentation__.Notify(115063)
		cmd := it.cur()
		cmd.proposal = d.r.mu.proposals[cmd.idKey]
		anyLocal = anyLocal || func() bool {
			__antithesis_instrumentation__.Notify(115064)
			return cmd.IsLocal() == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(115060)
	if !anyLocal && func() bool {
		__antithesis_instrumentation__.Notify(115065)
		return d.r.mu.proposalQuota == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(115066)

		return false
	} else {
		__antithesis_instrumentation__.Notify(115067)
	}
	__antithesis_instrumentation__.Notify(115061)
	for it.init(&d.cmdBuf); it.Valid(); it.Next() {
		__antithesis_instrumentation__.Notify(115068)
		cmd := it.cur()
		var toRelease *quotapool.IntAlloc
		shouldRemove := cmd.IsLocal() && func() bool {
			__antithesis_instrumentation__.Notify(115070)
			return cmd.raftCmd.MaxLeaseIndex == cmd.proposal.command.MaxLeaseIndex == true
		}() == true
		if shouldRemove {
			__antithesis_instrumentation__.Notify(115071)

			delete(d.r.mu.proposals, cmd.idKey)
			toRelease = cmd.proposal.quotaAlloc
			cmd.proposal.quotaAlloc = nil
		} else {
			__antithesis_instrumentation__.Notify(115072)
		}
		__antithesis_instrumentation__.Notify(115069)

		if d.r.mu.proposalQuota != nil {
			__antithesis_instrumentation__.Notify(115073)
			d.r.mu.quotaReleaseQueue = append(d.r.mu.quotaReleaseQueue, toRelease)
		} else {
			__antithesis_instrumentation__.Notify(115074)
		}
	}
	__antithesis_instrumentation__.Notify(115062)
	return anyLocal
}

func (d *replicaDecoder) createTracingSpans(ctx context.Context) {
	__antithesis_instrumentation__.Notify(115075)
	const opName = "raft application"
	var it replicatedCmdBufSlice
	for it.init(&d.cmdBuf); it.Valid(); it.Next() {
		__antithesis_instrumentation__.Notify(115076)
		cmd := it.cur()

		if cmd.IsLocal() {
			__antithesis_instrumentation__.Notify(115078)

			propCtx := ctx
			var propSp *tracing.Span

			if sp := tracing.SpanFromContext(cmd.proposal.ctx); sp != nil {
				__antithesis_instrumentation__.Notify(115080)
				propCtx, propSp = sp.Tracer().StartSpanCtx(
					propCtx, "local proposal", tracing.WithParent(sp),
				)
			} else {
				__antithesis_instrumentation__.Notify(115081)
			}
			__antithesis_instrumentation__.Notify(115079)
			cmd.ctx, cmd.sp = propCtx, propSp
		} else {
			__antithesis_instrumentation__.Notify(115082)
			if cmd.raftCmd.TraceData != nil {
				__antithesis_instrumentation__.Notify(115083)

				spanMeta, err := d.r.AmbientContext.Tracer.ExtractMetaFrom(tracing.MapCarrier{
					Map: cmd.raftCmd.TraceData,
				})
				if err != nil {
					__antithesis_instrumentation__.Notify(115084)
					log.Errorf(ctx, "unable to extract trace data from raft command: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(115085)
					cmd.ctx, cmd.sp = d.r.AmbientContext.Tracer.StartSpanCtx(
						ctx,
						opName,

						tracing.WithRemoteParentFromSpanMeta(spanMeta),
						tracing.WithFollowsFrom(),
					)
				}
			} else {
				__antithesis_instrumentation__.Notify(115086)
				cmd.ctx, cmd.sp = tracing.ChildSpan(ctx, opName)
			}
		}
		__antithesis_instrumentation__.Notify(115077)

		if util.RaceEnabled && func() bool {
			__antithesis_instrumentation__.Notify(115087)
			return cmd.ctx.Done() != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(115088)
			panic(fmt.Sprintf("cancelable context observed during raft application: %+v", cmd))
		} else {
			__antithesis_instrumentation__.Notify(115089)
		}
	}
}

func (d *replicaDecoder) NewCommandIter() apply.CommandIterator {
	__antithesis_instrumentation__.Notify(115090)
	it := d.cmdBuf.newIter()
	it.init(&d.cmdBuf)
	return it
}

func (d *replicaDecoder) Reset() {
	__antithesis_instrumentation__.Notify(115091)
	d.cmdBuf.clear()
}
