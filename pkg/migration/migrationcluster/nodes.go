package migrationcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type Node struct {
	ID    roachpb.NodeID
	Epoch int64
}

type Nodes []Node

func NodesFromNodeLiveness(ctx context.Context, nl NodeLiveness) (Nodes, error) {
	__antithesis_instrumentation__.Notify(128195)
	var ns []Node
	ls, err := nl.GetLivenessesFromKV(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(128198)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(128199)
	}
	__antithesis_instrumentation__.Notify(128196)
	for _, l := range ls {
		__antithesis_instrumentation__.Notify(128200)
		if l.Membership.Decommissioned() {
			__antithesis_instrumentation__.Notify(128204)
			continue
		} else {
			__antithesis_instrumentation__.Notify(128205)
		}
		__antithesis_instrumentation__.Notify(128201)
		live, err := nl.IsLive(l.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(128206)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(128207)
		}
		__antithesis_instrumentation__.Notify(128202)
		if !live {
			__antithesis_instrumentation__.Notify(128208)
			return nil, errors.Newf("n%d required, but unavailable", l.NodeID)
		} else {
			__antithesis_instrumentation__.Notify(128209)
		}
		__antithesis_instrumentation__.Notify(128203)
		ns = append(ns, Node{ID: l.NodeID, Epoch: l.Epoch})
	}
	__antithesis_instrumentation__.Notify(128197)
	return ns, nil
}

func (ns Nodes) Identical(other Nodes) (ok bool, _ []redact.RedactableString) {
	__antithesis_instrumentation__.Notify(128210)
	a, b := ns, other

	type ent struct {
		node         Node
		count        int
		epochChanged bool
	}
	m := map[roachpb.NodeID]ent{}
	for _, node := range a {
		__antithesis_instrumentation__.Notify(128214)
		m[node.ID] = ent{count: 1, node: node, epochChanged: false}
	}
	__antithesis_instrumentation__.Notify(128211)
	for _, node := range b {
		__antithesis_instrumentation__.Notify(128215)
		e, ok := m[node.ID]
		e.count--
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(128217)
			return e.node.Epoch != node.Epoch == true
		}() == true {
			__antithesis_instrumentation__.Notify(128218)
			e.epochChanged = true
		} else {
			__antithesis_instrumentation__.Notify(128219)
		}
		__antithesis_instrumentation__.Notify(128216)
		m[node.ID] = e
	}
	__antithesis_instrumentation__.Notify(128212)

	var diffs []redact.RedactableString
	for id, e := range m {
		__antithesis_instrumentation__.Notify(128220)
		if e.epochChanged {
			__antithesis_instrumentation__.Notify(128223)
			diffs = append(diffs, redact.Sprintf("n%d's Epoch changed", id))
		} else {
			__antithesis_instrumentation__.Notify(128224)
		}
		__antithesis_instrumentation__.Notify(128221)
		if e.count > 0 {
			__antithesis_instrumentation__.Notify(128225)
			diffs = append(diffs, redact.Sprintf("n%d was decommissioned", id))
		} else {
			__antithesis_instrumentation__.Notify(128226)
		}
		__antithesis_instrumentation__.Notify(128222)
		if e.count < 0 {
			__antithesis_instrumentation__.Notify(128227)
			diffs = append(diffs, redact.Sprintf("n%d joined the cluster", id))
		} else {
			__antithesis_instrumentation__.Notify(128228)
		}
	}
	__antithesis_instrumentation__.Notify(128213)

	return len(diffs) == 0, diffs
}

func (ns Nodes) String() string {
	__antithesis_instrumentation__.Notify(128229)
	return redact.StringWithoutMarkers(ns)
}

func (ns Nodes) SafeFormat(s redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(128230)
	s.SafeString("n{")
	if len(ns) > 0 {
		__antithesis_instrumentation__.Notify(128232)
		s.Printf("%d", ns[0].ID)
		for _, node := range ns[1:] {
			__antithesis_instrumentation__.Notify(128233)
			s.Printf(",%d", node.ID)
		}
	} else {
		__antithesis_instrumentation__.Notify(128234)
	}
	__antithesis_instrumentation__.Notify(128231)
	s.SafeString("}")
}
