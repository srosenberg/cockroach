// Package nodelivenesstest provides a mock implementation of NodeLiveness
// to facilitate testing of migration infrastructure.
package nodelivenesstest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type NodeLiveness struct {
	ls   []livenesspb.Liveness
	dead map[roachpb.NodeID]struct{}
}

func New(numNodes int) *NodeLiveness {
	__antithesis_instrumentation__.Notify(128734)
	nl := &NodeLiveness{
		ls:   make([]livenesspb.Liveness, numNodes),
		dead: make(map[roachpb.NodeID]struct{}),
	}
	for i := 0; i < numNodes; i++ {
		__antithesis_instrumentation__.Notify(128736)
		nl.ls[i] = livenesspb.Liveness{
			NodeID: roachpb.NodeID(i + 1), Epoch: 1,
			Membership: livenesspb.MembershipStatus_ACTIVE,
		}
	}
	__antithesis_instrumentation__.Notify(128735)
	return nl
}

func (t *NodeLiveness) GetLivenessesFromKV(context.Context) ([]livenesspb.Liveness, error) {
	__antithesis_instrumentation__.Notify(128737)
	return t.ls, nil
}

func (t *NodeLiveness) IsLive(id roachpb.NodeID) (bool, error) {
	__antithesis_instrumentation__.Notify(128738)
	_, dead := t.dead[id]
	return !dead, nil
}

func (t *NodeLiveness) Decommission(id roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(128739)
	for i := range t.ls {
		__antithesis_instrumentation__.Notify(128740)
		if t.ls[i].NodeID == id {
			__antithesis_instrumentation__.Notify(128741)
			t.ls[i].Membership = livenesspb.MembershipStatus_DECOMMISSIONED
			break
		} else {
			__antithesis_instrumentation__.Notify(128742)
		}
	}
}

func (t *NodeLiveness) AddNewNode() {
	__antithesis_instrumentation__.Notify(128743)
	t.AddNode(roachpb.NodeID(len(t.ls) + 1))
}

func (t *NodeLiveness) AddNode(id roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(128744)
	t.ls = append(t.ls, livenesspb.Liveness{
		NodeID:     id,
		Epoch:      1,
		Membership: livenesspb.MembershipStatus_ACTIVE,
	})
}

func (t *NodeLiveness) DownNode(id roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(128745)
	t.dead[id] = struct{}{}
}

func (t *NodeLiveness) RestartNode(id roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(128746)
	for i := range t.ls {
		__antithesis_instrumentation__.Notify(128748)
		if t.ls[i].NodeID == id {
			__antithesis_instrumentation__.Notify(128749)
			t.ls[i].Epoch++
			break
		} else {
			__antithesis_instrumentation__.Notify(128750)
		}
	}
	__antithesis_instrumentation__.Notify(128747)

	delete(t.dead, id)
}
