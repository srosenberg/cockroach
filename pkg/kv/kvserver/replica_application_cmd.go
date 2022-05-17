package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type replicatedCmd struct {
	ent *raftpb.Entry
	decodedRaftEntry

	proposal *ProposalData

	ctx context.Context

	sp *tracing.Span

	leaseIndex    uint64
	forcedErr     *roachpb.Error
	proposalRetry proposalReevaluationReason

	splitMergeUnlock func()

	localResult *result.LocalResult
	response    proposalResult
}

type decodedRaftEntry struct {
	idKey      kvserverbase.CmdIDKey
	raftCmd    kvserverpb.RaftCommand
	confChange *decodedConfChange
}

type decodedConfChange struct {
	raftpb.ConfChangeI
	kvserverpb.ConfChangeContext
}

func (c *replicatedCmd) decode(ctx context.Context, e *raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(114944)
	c.ent = e
	return c.decodedRaftEntry.decode(ctx, e)
}

func (c *replicatedCmd) Index() uint64 {
	__antithesis_instrumentation__.Notify(114945)
	return c.ent.Index
}

func (c *replicatedCmd) IsTrivial() bool {
	__antithesis_instrumentation__.Notify(114946)
	return isTrivial(c.replicatedResult())
}

func (c *replicatedCmd) IsLocal() bool {
	__antithesis_instrumentation__.Notify(114947)
	return c.proposal != nil
}

func (c *replicatedCmd) Ctx() context.Context {
	__antithesis_instrumentation__.Notify(114948)
	return c.ctx
}

func (c *replicatedCmd) AckErrAndFinish(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(114949)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(114951)
		c.response.Err = roachpb.NewError(roachpb.NewAmbiguousResultError(err))
	} else {
		__antithesis_instrumentation__.Notify(114952)
	}
	__antithesis_instrumentation__.Notify(114950)
	return c.AckOutcomeAndFinish(ctx)
}

func (c *replicatedCmd) Rejected() bool {
	__antithesis_instrumentation__.Notify(114953)
	return c.forcedErr != nil
}

func (c *replicatedCmd) CanAckBeforeApplication() bool {
	__antithesis_instrumentation__.Notify(114954)

	req := c.proposal.Request
	return req.IsIntentWrite() && func() bool {
		__antithesis_instrumentation__.Notify(114955)
		return !req.AsyncConsensus == true
	}() == true
}

func (c *replicatedCmd) AckSuccess(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(114956)
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(114958)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114959)
	}
	__antithesis_instrumentation__.Notify(114957)

	var resp proposalResult
	reply := *c.proposal.Local.Reply
	reply.Responses = append([]roachpb.ResponseUnion(nil), reply.Responses...)
	resp.Reply = &reply
	resp.EncounteredIntents = c.proposal.Local.DetachEncounteredIntents()
	resp.EndTxns = c.proposal.Local.DetachEndTxns(false)
	log.Event(ctx, "ack-ing replication success to the client; application will continue async w.r.t. the client")
	c.proposal.signalProposalResult(resp)
	return nil
}

func (c *replicatedCmd) AckOutcomeAndFinish(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(114960)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(114962)
		c.proposal.finishApplication(ctx, c.response)
	} else {
		__antithesis_instrumentation__.Notify(114963)
	}
	__antithesis_instrumentation__.Notify(114961)
	c.finishTracingSpan()
	return nil
}

func (c *replicatedCmd) FinishNonLocal(ctx context.Context) {
	__antithesis_instrumentation__.Notify(114964)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(114966)
		log.Fatalf(ctx, "proposal unexpectedly local: %v", c.replicatedResult())
	} else {
		__antithesis_instrumentation__.Notify(114967)
	}
	__antithesis_instrumentation__.Notify(114965)
	c.finishTracingSpan()
}

func (c *replicatedCmd) finishTracingSpan() {
	__antithesis_instrumentation__.Notify(114968)
	c.sp.Finish()
	c.ctx, c.sp = nil, nil
}

func (d *decodedRaftEntry) decode(ctx context.Context, e *raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(114969)
	*d = decodedRaftEntry{}

	if len(e.Data) == 0 {
		__antithesis_instrumentation__.Notify(114971)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114972)
	}
	__antithesis_instrumentation__.Notify(114970)
	switch e.Type {
	case raftpb.EntryNormal:
		__antithesis_instrumentation__.Notify(114973)
		return d.decodeNormalEntry(e)
	case raftpb.EntryConfChange, raftpb.EntryConfChangeV2:
		__antithesis_instrumentation__.Notify(114974)
		return d.decodeConfChangeEntry(e)
	default:
		__antithesis_instrumentation__.Notify(114975)
		log.Fatalf(ctx, "unexpected Raft entry: %v", e)
		return nil
	}
}

func (d *decodedRaftEntry) decodeNormalEntry(e *raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(114976)
	var encodedCommand []byte
	d.idKey, encodedCommand = kvserverbase.DecodeRaftCommand(e.Data)

	if len(encodedCommand) == 0 {
		__antithesis_instrumentation__.Notify(114978)
		d.idKey = ""
	} else {
		__antithesis_instrumentation__.Notify(114979)
		if err := protoutil.Unmarshal(encodedCommand, &d.raftCmd); err != nil {
			__antithesis_instrumentation__.Notify(114980)
			return wrapWithNonDeterministicFailure(err, "while unmarshaling entry")
		} else {
			__antithesis_instrumentation__.Notify(114981)
		}
	}
	__antithesis_instrumentation__.Notify(114977)
	return nil
}

func (d *decodedRaftEntry) decodeConfChangeEntry(e *raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(114982)
	d.confChange = &decodedConfChange{}

	switch e.Type {
	case raftpb.EntryConfChange:
		__antithesis_instrumentation__.Notify(114986)
		var cc raftpb.ConfChange
		if err := protoutil.Unmarshal(e.Data, &cc); err != nil {
			__antithesis_instrumentation__.Notify(114991)
			return wrapWithNonDeterministicFailure(err, "while unmarshaling ConfChange")
		} else {
			__antithesis_instrumentation__.Notify(114992)
		}
		__antithesis_instrumentation__.Notify(114987)
		d.confChange.ConfChangeI = cc
	case raftpb.EntryConfChangeV2:
		__antithesis_instrumentation__.Notify(114988)
		var cc raftpb.ConfChangeV2
		if err := protoutil.Unmarshal(e.Data, &cc); err != nil {
			__antithesis_instrumentation__.Notify(114993)
			return wrapWithNonDeterministicFailure(err, "while unmarshaling ConfChangeV2")
		} else {
			__antithesis_instrumentation__.Notify(114994)
		}
		__antithesis_instrumentation__.Notify(114989)
		d.confChange.ConfChangeI = cc
	default:
		__antithesis_instrumentation__.Notify(114990)
		const msg = "unknown entry type"
		err := errors.New(msg)
		return wrapWithNonDeterministicFailure(err, msg)
	}
	__antithesis_instrumentation__.Notify(114983)
	if err := protoutil.Unmarshal(d.confChange.AsV2().Context, &d.confChange.ConfChangeContext); err != nil {
		__antithesis_instrumentation__.Notify(114995)
		return wrapWithNonDeterministicFailure(err, "while unmarshaling ConfChangeContext")
	} else {
		__antithesis_instrumentation__.Notify(114996)
	}
	__antithesis_instrumentation__.Notify(114984)
	if err := protoutil.Unmarshal(d.confChange.Payload, &d.raftCmd); err != nil {
		__antithesis_instrumentation__.Notify(114997)
		return wrapWithNonDeterministicFailure(err, "while unmarshaling RaftCommand")
	} else {
		__antithesis_instrumentation__.Notify(114998)
	}
	__antithesis_instrumentation__.Notify(114985)
	d.idKey = kvserverbase.CmdIDKey(d.confChange.CommandID)
	return nil
}

func (d *decodedRaftEntry) replicatedResult() *kvserverpb.ReplicatedEvalResult {
	__antithesis_instrumentation__.Notify(114999)
	return &d.raftCmd.ReplicatedEvalResult
}
