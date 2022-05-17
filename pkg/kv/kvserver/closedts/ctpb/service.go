package ctpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type SeqNum int64

func (SeqNum) SafeValue() { __antithesis_instrumentation__.Notify(97932) }

type LAI int64

func (LAI) SafeValue() { __antithesis_instrumentation__.Notify(97933) }

func (m *Update) String() string {
	__antithesis_instrumentation__.Notify(97934)
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Seq num: %d, sending node: n%d, snapshot: %t, size: %d bytes",
		m.SeqNum, m.NodeID, m.Snapshot, m.Size())
	sb.WriteString(", closed timestamps: ")
	now := timeutil.Now()
	for i, upd := range m.ClosedTimestamps {
		__antithesis_instrumentation__.Notify(97940)
		if i != 0 {
			__antithesis_instrumentation__.Notify(97943)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(97944)
		}
		__antithesis_instrumentation__.Notify(97941)
		ago := now.Sub(upd.ClosedTimestamp.GoTime()).Truncate(time.Millisecond)
		var agoMsg string
		if ago >= 0 {
			__antithesis_instrumentation__.Notify(97945)
			agoMsg = fmt.Sprintf("%s ago", ago)
		} else {
			__antithesis_instrumentation__.Notify(97946)
			agoMsg = fmt.Sprintf("%s in the future", -ago)
		}
		__antithesis_instrumentation__.Notify(97942)
		fmt.Fprintf(sb, "%s:%s (%s)", upd.Policy, upd.ClosedTimestamp, agoMsg)
	}
	__antithesis_instrumentation__.Notify(97935)
	sb.WriteRune('\n')

	fmt.Fprintf(sb, "Added or updated (%d ranges): (<range>:<LAI>) ", len(m.AddedOrUpdated))
	added := make([]Update_RangeUpdate, len(m.AddedOrUpdated))
	copy(added, m.AddedOrUpdated)
	sort.Slice(added, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(97947)
		return added[i].RangeID < added[j].RangeID
	})
	__antithesis_instrumentation__.Notify(97936)
	for i, upd := range m.AddedOrUpdated {
		__antithesis_instrumentation__.Notify(97948)
		if i > 0 {
			__antithesis_instrumentation__.Notify(97950)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(97951)
		}
		__antithesis_instrumentation__.Notify(97949)
		fmt.Fprintf(sb, "%d:%d", upd.RangeID, upd.LAI)
	}
	__antithesis_instrumentation__.Notify(97937)
	sb.WriteRune('\n')

	fmt.Fprintf(sb, "Removed (%d ranges): ", len(m.Removed))
	removed := make([]roachpb.RangeID, len(m.Removed))
	copy(removed, m.Removed)
	sort.Slice(removed, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(97952)
		return removed[i] < removed[j]
	})
	__antithesis_instrumentation__.Notify(97938)
	for i, rid := range removed {
		__antithesis_instrumentation__.Notify(97953)
		if i > 0 {
			__antithesis_instrumentation__.Notify(97955)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(97956)
		}
		__antithesis_instrumentation__.Notify(97954)
		fmt.Fprintf(sb, "r%d", rid)
	}
	__antithesis_instrumentation__.Notify(97939)
	sb.WriteRune('\n')
	return sb.String()
}
