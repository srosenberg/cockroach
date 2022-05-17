package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

const maxRaftMsgType = raftpb.MsgPreVoteResp

func init() {
	for v := range raftpb.MessageType_name {
		typ := raftpb.MessageType(v)
		if typ > maxRaftMsgType {
			panic(fmt.Sprintf("raft.MessageType (%s) with value larger than maxRaftMsgType", typ))
		}
	}
}

func init() {
	raft.SetLogger(&raftLogger{ctx: context.Background()})
}

type raftLogger struct {
	ctx context.Context
}

func (r *raftLogger) Debug(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112606)
	if log.V(3) {
		__antithesis_instrumentation__.Notify(112607)
		log.InfofDepth(r.ctx, 1, "", v...)
	} else {
		__antithesis_instrumentation__.Notify(112608)
	}
}

func (r *raftLogger) Debugf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112609)
	if log.V(3) {
		__antithesis_instrumentation__.Notify(112610)
		log.InfofDepth(r.ctx, 1, format, v...)
	} else {
		__antithesis_instrumentation__.Notify(112611)
	}
}

func (r *raftLogger) Info(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112612)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(112613)
		log.InfofDepth(r.ctx, 1, "", v...)
	} else {
		__antithesis_instrumentation__.Notify(112614)
	}
}

func (r *raftLogger) Infof(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112615)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(112616)
		log.InfofDepth(r.ctx, 1, format, v...)
	} else {
		__antithesis_instrumentation__.Notify(112617)
	}
}

func (r *raftLogger) Warning(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112618)
	log.WarningfDepth(r.ctx, 1, "", v...)
}

func (r *raftLogger) Warningf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112619)
	log.WarningfDepth(r.ctx, 1, format, v...)
}

func (r *raftLogger) Error(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112620)
	log.ErrorfDepth(r.ctx, 1, "", v...)
}

func (r *raftLogger) Errorf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112621)
	log.ErrorfDepth(r.ctx, 1, format, v...)
}

func (r *raftLogger) Fatal(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112622)
	wrapNumbersAsSafe(v)
	log.FatalfDepth(r.ctx, 1, "", v...)
}

func (r *raftLogger) Fatalf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112623)
	wrapNumbersAsSafe(v)
	log.FatalfDepth(r.ctx, 1, format, v...)
}

func (r *raftLogger) Panic(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112624)
	wrapNumbersAsSafe(v)
	log.FatalfDepth(r.ctx, 1, "", v...)
}

func (r *raftLogger) Panicf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(112625)
	wrapNumbersAsSafe(v)
	log.FatalfDepth(r.ctx, 1, format, v...)
}

func wrapNumbersAsSafe(v ...interface{}) {
	__antithesis_instrumentation__.Notify(112626)
	for i := range v {
		__antithesis_instrumentation__.Notify(112627)
		switch v[i].(type) {
		case uint:
			__antithesis_instrumentation__.Notify(112628)
			v[i] = redact.Safe(v[i])
		case uint8:
			__antithesis_instrumentation__.Notify(112629)
			v[i] = redact.Safe(v[i])
		case uint16:
			__antithesis_instrumentation__.Notify(112630)
			v[i] = redact.Safe(v[i])
		case uint32:
			__antithesis_instrumentation__.Notify(112631)
			v[i] = redact.Safe(v[i])
		case uint64:
			__antithesis_instrumentation__.Notify(112632)
			v[i] = redact.Safe(v[i])
		case int:
			__antithesis_instrumentation__.Notify(112633)
			v[i] = redact.Safe(v[i])
		case int8:
			__antithesis_instrumentation__.Notify(112634)
			v[i] = redact.Safe(v[i])
		case int16:
			__antithesis_instrumentation__.Notify(112635)
			v[i] = redact.Safe(v[i])
		case int32:
			__antithesis_instrumentation__.Notify(112636)
			v[i] = redact.Safe(v[i])
		case int64:
			__antithesis_instrumentation__.Notify(112637)
			v[i] = redact.Safe(v[i])
		case float32:
			__antithesis_instrumentation__.Notify(112638)
			v[i] = redact.Safe(v[i])
		case float64:
			__antithesis_instrumentation__.Notify(112639)
			v[i] = redact.Safe(v[i])
		default:
			__antithesis_instrumentation__.Notify(112640)
		}
	}
}

func verboseRaftLoggingEnabled() bool {
	__antithesis_instrumentation__.Notify(112641)
	return log.V(5)
}

func logRaftReady(ctx context.Context, ready raft.Ready) {
	__antithesis_instrumentation__.Notify(112642)
	if !verboseRaftLoggingEnabled() {
		__antithesis_instrumentation__.Notify(112650)
		return
	} else {
		__antithesis_instrumentation__.Notify(112651)
	}
	__antithesis_instrumentation__.Notify(112643)

	var buf bytes.Buffer
	if ready.SoftState != nil {
		__antithesis_instrumentation__.Notify(112652)
		fmt.Fprintf(&buf, "  SoftState updated: %+v\n", *ready.SoftState)
	} else {
		__antithesis_instrumentation__.Notify(112653)
	}
	__antithesis_instrumentation__.Notify(112644)
	if !raft.IsEmptyHardState(ready.HardState) {
		__antithesis_instrumentation__.Notify(112654)
		fmt.Fprintf(&buf, "  HardState updated: %+v\n", ready.HardState)
	} else {
		__antithesis_instrumentation__.Notify(112655)
	}
	__antithesis_instrumentation__.Notify(112645)
	for i, e := range ready.Entries {
		__antithesis_instrumentation__.Notify(112656)
		fmt.Fprintf(&buf, "  New Entry[%d]: %.200s\n",
			i, raft.DescribeEntry(e, raftEntryFormatter))
	}
	__antithesis_instrumentation__.Notify(112646)
	for i, e := range ready.CommittedEntries {
		__antithesis_instrumentation__.Notify(112657)
		fmt.Fprintf(&buf, "  Committed Entry[%d]: %.200s\n",
			i, raft.DescribeEntry(e, raftEntryFormatter))
	}
	__antithesis_instrumentation__.Notify(112647)
	if !raft.IsEmptySnap(ready.Snapshot) {
		__antithesis_instrumentation__.Notify(112658)
		snap := ready.Snapshot
		snap.Data = nil
		fmt.Fprintf(&buf, "  Snapshot updated: %v\n", snap)
	} else {
		__antithesis_instrumentation__.Notify(112659)
	}
	__antithesis_instrumentation__.Notify(112648)
	for i, m := range ready.Messages {
		__antithesis_instrumentation__.Notify(112660)
		fmt.Fprintf(&buf, "  Outgoing Message[%d]: %.200s\n",
			i, raftDescribeMessage(m, raftEntryFormatter))
	}
	__antithesis_instrumentation__.Notify(112649)
	log.Infof(ctx, "raft ready (must-sync=%t)\n%s", ready.MustSync, buf.String())
}

func raftDescribeMessage(m raftpb.Message, f raft.EntryFormatter) string {
	__antithesis_instrumentation__.Notify(112661)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%x->%x %v Term:%d Log:%d/%d", m.From, m.To, m.Type, m.Term, m.LogTerm, m.Index)
	if m.Reject {
		__antithesis_instrumentation__.Notify(112666)
		fmt.Fprintf(&buf, " Rejected (Hint: %d)", m.RejectHint)
	} else {
		__antithesis_instrumentation__.Notify(112667)
	}
	__antithesis_instrumentation__.Notify(112662)
	if m.Commit != 0 {
		__antithesis_instrumentation__.Notify(112668)
		fmt.Fprintf(&buf, " Commit:%d", m.Commit)
	} else {
		__antithesis_instrumentation__.Notify(112669)
	}
	__antithesis_instrumentation__.Notify(112663)
	if len(m.Entries) > 0 {
		__antithesis_instrumentation__.Notify(112670)
		fmt.Fprintf(&buf, " Entries:[")
		for i, e := range m.Entries {
			__antithesis_instrumentation__.Notify(112672)
			if i != 0 {
				__antithesis_instrumentation__.Notify(112674)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(112675)
			}
			__antithesis_instrumentation__.Notify(112673)
			buf.WriteString(raft.DescribeEntry(e, f))
		}
		__antithesis_instrumentation__.Notify(112671)
		fmt.Fprintf(&buf, "]")
	} else {
		__antithesis_instrumentation__.Notify(112676)
	}
	__antithesis_instrumentation__.Notify(112664)
	if !raft.IsEmptySnap(m.Snapshot) {
		__antithesis_instrumentation__.Notify(112677)
		snap := m.Snapshot
		snap.Data = nil
		fmt.Fprintf(&buf, " Snapshot:%v", snap)
	} else {
		__antithesis_instrumentation__.Notify(112678)
	}
	__antithesis_instrumentation__.Notify(112665)
	return buf.String()
}

func raftEntryFormatter(data []byte) string {
	__antithesis_instrumentation__.Notify(112679)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(112681)
		return "[empty]"
	} else {
		__antithesis_instrumentation__.Notify(112682)
	}
	__antithesis_instrumentation__.Notify(112680)
	commandID, _ := kvserverbase.DecodeRaftCommand(data)
	return fmt.Sprintf("[%x] [%d]", commandID, len(data))
}

var raftMessageRequestPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(112683)
		return &kvserverpb.RaftMessageRequest{}
	},
}

func newRaftMessageRequest() *kvserverpb.RaftMessageRequest {
	__antithesis_instrumentation__.Notify(112684)
	return raftMessageRequestPool.Get().(*kvserverpb.RaftMessageRequest)
}

func releaseRaftMessageRequest(m *kvserverpb.RaftMessageRequest) {
	__antithesis_instrumentation__.Notify(112685)
	*m = kvserverpb.RaftMessageRequest{}
	raftMessageRequestPool.Put(m)
}

func (r *Replica) traceEntries(ents []raftpb.Entry, event string) {
	__antithesis_instrumentation__.Notify(112686)
	if log.V(1) || func() bool {
		__antithesis_instrumentation__.Notify(112687)
		return r.store.TestingKnobs().TraceAllRaftEvents == true
	}() == true {
		__antithesis_instrumentation__.Notify(112688)
		ids := extractIDs(nil, ents)
		traceProposals(r, ids, event)
	} else {
		__antithesis_instrumentation__.Notify(112689)
	}
}

func (r *Replica) traceMessageSends(msgs []raftpb.Message, event string) {
	__antithesis_instrumentation__.Notify(112690)
	if log.V(1) || func() bool {
		__antithesis_instrumentation__.Notify(112691)
		return r.store.TestingKnobs().TraceAllRaftEvents == true
	}() == true {
		__antithesis_instrumentation__.Notify(112692)
		var ids []kvserverbase.CmdIDKey
		for _, m := range msgs {
			__antithesis_instrumentation__.Notify(112694)
			ids = extractIDs(ids, m.Entries)
		}
		__antithesis_instrumentation__.Notify(112693)
		traceProposals(r, ids, event)
	} else {
		__antithesis_instrumentation__.Notify(112695)
	}
}

func extractIDs(ids []kvserverbase.CmdIDKey, ents []raftpb.Entry) []kvserverbase.CmdIDKey {
	__antithesis_instrumentation__.Notify(112696)
	for _, e := range ents {
		__antithesis_instrumentation__.Notify(112698)
		if e.Type == raftpb.EntryNormal && func() bool {
			__antithesis_instrumentation__.Notify(112699)
			return len(e.Data) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(112700)
			id, _ := kvserverbase.DecodeRaftCommand(e.Data)
			ids = append(ids, id)
		} else {
			__antithesis_instrumentation__.Notify(112701)
		}
	}
	__antithesis_instrumentation__.Notify(112697)
	return ids
}

func traceProposals(r *Replica, ids []kvserverbase.CmdIDKey, event string) {
	__antithesis_instrumentation__.Notify(112702)
	ctxs := make([]context.Context, 0, len(ids))
	r.mu.RLock()
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(112704)
		if prop, ok := r.mu.proposals[id]; ok {
			__antithesis_instrumentation__.Notify(112705)
			ctxs = append(ctxs, prop.ctx)
		} else {
			__antithesis_instrumentation__.Notify(112706)
		}
	}
	__antithesis_instrumentation__.Notify(112703)
	r.mu.RUnlock()
	for _, ctx := range ctxs {
		__antithesis_instrumentation__.Notify(112707)
		log.Eventf(ctx, "%v", event)
	}
}
