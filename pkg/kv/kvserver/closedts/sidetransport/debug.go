package sidetransport

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func (s *Receiver) HTML() string {
	__antithesis_instrumentation__.Notify(98578)
	sb := &strings.Builder{}

	header := func(s string) {
		__antithesis_instrumentation__.Notify(98586)
		fmt.Fprintf(sb, "<h4>%s</h4>", s)
	}
	__antithesis_instrumentation__.Notify(98579)

	header("Incoming streams")
	s.mu.RLock()
	conns := make([]*incomingStream, 0, len(s.mu.conns))
	for _, c := range s.mu.conns {
		__antithesis_instrumentation__.Notify(98587)
		conns = append(conns, c)
	}
	__antithesis_instrumentation__.Notify(98580)
	s.mu.RUnlock()

	sort.Slice(conns, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(98588)
		return conns[i].nodeID < conns[j].nodeID
	})
	__antithesis_instrumentation__.Notify(98581)
	for _, c := range conns {
		__antithesis_instrumentation__.Notify(98589)
		sb.WriteString(c.html() + "<br>")
	}
	__antithesis_instrumentation__.Notify(98582)

	header("Closed streams (most recent first; only one per node)")
	s.historyMu.Lock()
	closed := make([]streamCloseInfo, 0, len(s.historyMu.lastClosed))
	for _, c := range s.historyMu.lastClosed {
		__antithesis_instrumentation__.Notify(98590)
		closed = append(closed, c)
	}
	__antithesis_instrumentation__.Notify(98583)
	s.historyMu.Unlock()

	sort.Slice(closed, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(98591)
		return closed[i].closeTime.After(closed[j].closeTime)
	})
	__antithesis_instrumentation__.Notify(98584)
	now := timeutil.Now()
	for _, c := range closed {
		__antithesis_instrumentation__.Notify(98592)
		fmt.Fprintf(sb, "n%d: incoming conn closed at %s (%s ago). err: %s\n",
			c.nodeID, c.closeTime.Truncate(time.Millisecond), now.Sub(c.closeTime).Truncate(time.Second), c.closeErr)
	}
	__antithesis_instrumentation__.Notify(98585)

	return strings.ReplaceAll(sb.String(), "\n", "<br>")
}

func (r *incomingStream) html() string {
	__antithesis_instrumentation__.Notify(98593)
	r.mu.Lock()
	defer r.mu.Unlock()

	sb := &strings.Builder{}

	bold := func(s string) {
		__antithesis_instrumentation__.Notify(98596)
		fmt.Fprintf(sb, "<h4>%s</h4>", s)
	}
	__antithesis_instrumentation__.Notify(98594)
	escape := func(s string) {
		__antithesis_instrumentation__.Notify(98597)
		sb.WriteString(html.EscapeString(s))
	}
	__antithesis_instrumentation__.Notify(98595)

	now := timeutil.Now()
	bold(fmt.Sprintf("n:%d ", r.nodeID))
	fmt.Fprintf(sb, "conn open: %s (%s ago), last received: %s (%s ago), last seq num: %d, closed timestamps: ",
		r.connectedAt.Truncate(time.Second),
		now.Sub(r.connectedAt).Truncate(time.Second),
		r.mu.lastReceived.Truncate(time.Millisecond), now.Sub(r.mu.lastReceived).Truncate(time.Millisecond),
		r.mu.streamState.lastSeqNum)
	escape(r.mu.streamState.String())
	return sb.String()
}

func (s *Sender) HTML() string {
	__antithesis_instrumentation__.Notify(98598)
	sb := &strings.Builder{}

	header := func(s string) {
		__antithesis_instrumentation__.Notify(98607)
		fmt.Fprintf(sb, "<h4>%s</h4>", s)
	}
	__antithesis_instrumentation__.Notify(98599)

	escape := func(s string) string {
		__antithesis_instrumentation__.Notify(98608)
		return strings.ReplaceAll(html.EscapeString(s), "\n", "<br>\n")
	}
	__antithesis_instrumentation__.Notify(98600)

	header("Closed timestamps sender state")
	s.leaseholdersMu.Lock()
	fmt.Fprintf(sb, "leaseholders: %d\n", len(s.leaseholdersMu.leaseholders))
	s.leaseholdersMu.Unlock()

	s.trackedMu.Lock()
	lastMsgSeq := s.trackedMu.lastSeqNum
	fmt.Fprint(sb, escape(s.trackedMu.streamState.String()))

	failed := 0
	for reason := ReasonUnknown + 1; reason < MaxReason; reason++ {
		__antithesis_instrumentation__.Notify(98609)
		failed += s.trackedMu.closingFailures[reason]
	}
	__antithesis_instrumentation__.Notify(98601)
	fmt.Fprintf(sb, "Failures to close during last cycle (%d ranges total): ", failed)
	for reason := ReasonUnknown + 1; reason < MaxReason; reason++ {
		__antithesis_instrumentation__.Notify(98610)
		if reason > ReasonUnknown+1 {
			__antithesis_instrumentation__.Notify(98612)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(98613)
		}
		__antithesis_instrumentation__.Notify(98611)
		fmt.Fprintf(sb, "%s: %d", reason, s.trackedMu.closingFailures[reason])
	}
	__antithesis_instrumentation__.Notify(98602)
	s.trackedMu.Unlock()

	s.connsMu.Lock()
	header(fmt.Sprintf("Connections (%d)", len(s.connsMu.conns)))
	nids := make([]roachpb.NodeID, 0, len(s.connsMu.conns))
	for nid := range s.connsMu.conns {
		__antithesis_instrumentation__.Notify(98614)
		nids = append(nids, nid)
	}
	__antithesis_instrumentation__.Notify(98603)
	sort.Slice(nids, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(98615)
		return nids[i] < nids[j]
	})
	__antithesis_instrumentation__.Notify(98604)
	now := timeutil.Now()
	for _, nid := range nids {
		__antithesis_instrumentation__.Notify(98616)
		state := s.connsMu.conns[nid].getState()
		fmt.Fprintf(sb, "n%d: ", nid)
		if state.connected {
			__antithesis_instrumentation__.Notify(98617)
			fmt.Fprintf(sb, "connected at: %s (%s ago)\n", state.connectedTime.Truncate(time.Millisecond), now.Sub(state.connectedTime).Truncate(time.Second))
		} else {
			__antithesis_instrumentation__.Notify(98618)
			fmt.Fprintf(sb, "disconnected at: %s (%s ago, err: %s)\n", state.lastDisconnectTime.Truncate(time.Millisecond), now.Sub(state.lastDisconnectTime).Truncate(time.Second), state.lastDisconnect)
		}
	}
	__antithesis_instrumentation__.Notify(98605)
	s.connsMu.Unlock()

	header("Last message")
	lastMsg, ok := s.buf.GetBySeq(context.Background(), lastMsgSeq)
	if !ok {
		__antithesis_instrumentation__.Notify(98619)
		fmt.Fprint(sb, "Buffer has been closed.\n")
	} else {
		__antithesis_instrumentation__.Notify(98620)
		if lastMsg == nil {
			__antithesis_instrumentation__.Notify(98621)
			fmt.Fprint(sb, "Buffer no longer has the message. This is unexpected.\n")
		} else {
			__antithesis_instrumentation__.Notify(98622)
			sb.WriteString(escape(lastMsg.String()))
		}
	}
	__antithesis_instrumentation__.Notify(98606)

	return strings.ReplaceAll(sb.String(), "\n", "<br>\n")
}
